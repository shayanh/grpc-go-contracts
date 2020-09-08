package contracts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

type ctxKey int

const (
	RequestIDKey ctxKey = iota + 1
)

type Condition interface{}

func invokeCondition(c Condition, args ...interface{}) error {
	v := reflect.ValueOf(c)
	t := v.Type()
	argv := make([]reflect.Value, t.NumIn())
	for i, arg := range args {
		expectedType := t.In(i)
		if arg == nil {
			argv[i] = reflect.New(expectedType).Elem()
		} else {
			argv[i] = reflect.ValueOf(arg)
		}
	}
	res := v.Call(argv)
	err, _ := res[0].Interface().(error)
	return err
}

func invokePreCondition(c Condition, req interface{}) error {
	return invokeCondition(c, req)
}

func invokePostCondition(c Condition, resp interface{}, respErr error, req interface{}, callHistory RPCCallHistory) error {
	return invokeCondition(c, resp, respErr, req, callHistory)
}

type Method interface{}

func getMethodName(method Method) string {
	return runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
}

func sameMethods(method Method, fullMethodName string) bool {
	tmp1 := strings.Split(getMethodName(method), ".")
	tmp2 := strings.Split(fullMethodName, "/")
	m1 := tmp1[len(tmp1)-1]
	m2 := tmp2[len(tmp2)-1]
	return m1 == m2
}

type UnaryRPCContract struct {
	Method         Method
	PreConditions  []Condition
	PostConditions []Condition
}

func (u *UnaryRPCContract) validate() error {
	// TODO
	return nil
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

type UnaryRPCCall struct {
	MethodName string
	Request    interface{}
	Response   interface{}
	Error      error
}

type ServerContract struct {
	logger Logger

	callsLock     sync.RWMutex
	unaryRPCCalls map[string]map[string][]*UnaryRPCCall

	contractsLock     sync.Mutex
	unaryRPCContracts map[string]*UnaryRPCContract
	serve             bool
}

type RPCCallHistory struct {
	requestID string
	sc        *ServerContract
}

func (h *RPCCallHistory) All() []*UnaryRPCCall {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	var res []*UnaryRPCCall
	for _, calls := range h.sc.unaryRPCCalls[h.requestID] {
		res = append(res, calls...)
	}
	return res
}

func (h *RPCCallHistory) Filter(method Method) []*UnaryRPCCall {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	var res []*UnaryRPCCall
	for methodName, calls := range h.sc.unaryRPCCalls[h.requestID] {
		if sameMethods(method, methodName) {
			res = append(res, calls...)
		}
	}
	return res
}

func NewServerContract(logger Logger) *ServerContract {
	return &ServerContract{
		logger:            logger,
		unaryRPCCalls:     make(map[string]map[string][]*UnaryRPCCall),
		unaryRPCContracts: make(map[string]*UnaryRPCContract),
	}
}

func (sc *ServerContract) RegisterUnaryRPCContract(rpcContract *UnaryRPCContract) {
	if err := rpcContract.validate(); err != nil {
		sc.logger.Fatal(err)
	}
	sc.register(rpcContract)
}

func (sc *ServerContract) register(rpcContract *UnaryRPCContract) {
	sc.contractsLock.Lock()
	defer sc.contractsLock.Unlock()

	if sc.serve {
		sc.logger.Fatal("ServerContract.RegisterUnaryRPCContract must called before ServerContract.UnaryServerInterceptor")
	}
	methodName := getMethodName(rpcContract.Method) // TODO: this is not the best key
	if _, ok := sc.unaryRPCContracts[methodName]; ok {
		sc.logger.Fatal("ServerContract.RegisterUnaryRPCContract found duplicate contract registration")
	}
	sc.unaryRPCContracts[methodName] = rpcContract
}

func shortID() string {
	b := make([]byte, 10)
	io.ReadFull(rand.Reader, b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (sc *ServerContract) generateRequestID(ctx context.Context) (context.Context, string) {
	sc.callsLock.RLock()
	defer sc.callsLock.RUnlock()

	var requestID string
	for {
		requestID = shortID()
		if _, ok := sc.unaryRPCCalls[requestID]; !ok {
			break
		}
	}
	return context.WithValue(ctx, RequestIDKey, requestID), requestID
}

func (sc *ServerContract) cleanup(requestID string) {
	sc.callsLock.Lock()
	defer sc.callsLock.Unlock()

	if _, ok := sc.unaryRPCCalls[requestID]; ok {
		delete(sc.unaryRPCCalls, requestID)
	}
}

func (sc *ServerContract) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	sc.contractsLock.Lock()
	defer sc.contractsLock.Unlock()
	sc.serve = true

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sc.logger.Info("server info full method = ", info.FullMethod)

		var requestID string
		ctx, requestID = sc.generateRequestID(ctx)

		var c *UnaryRPCContract
		for _, contract := range sc.unaryRPCContracts {
			if eq := sameMethods(contract.Method, info.FullMethod); eq {
				sc.logger.Info("contract method = ", getMethodName(contract.Method))
				c = contract
				break
			}
		}
		if c != nil {
			sc.logger.Info("pre")
			for _, preCondition := range c.PreConditions {
				err := invokePreCondition(preCondition, req)
				if err != nil {
					sc.logger.Error(err)
				}
			}
		}

		resp, err := handler(ctx, req)

		if c != nil {
			sc.logger.Info("post")
			for _, postCondition := range c.PostConditions {
				err := invokePostCondition(postCondition, resp, err, req,
					RPCCallHistory{requestID: requestID, sc: sc})
				if err != nil {
					sc.logger.Error(err)
				}
			}
			sc.cleanup(requestID)
		}
		return resp, err
	}
}

func (sc *ServerContract) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		requestID, ok := ctx.Value(RequestIDKey).(string)
		if ok {
			sc.callsLock.Lock()
			defer sc.callsLock.Unlock()

			call := &UnaryRPCCall{
				MethodName: method,
				Request:    req,
				Response:   reply,
				Error:      err,
			}
			if _, ok := sc.unaryRPCCalls[requestID]; !ok {
				sc.unaryRPCCalls[requestID] = make(map[string][]*UnaryRPCCall)
			}
			sc.unaryRPCCalls[requestID][method] = append(sc.unaryRPCCalls[requestID][method], call)
		}
		return err
	}
}
