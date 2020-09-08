package contracts

import (
	"context"
	"sync"

	"google.golang.org/grpc"
)

// UnaryRPCContract represents a contract for a unary RPC
type UnaryRPCContract struct {
	Method         Method
	PreConditions  []Condition
	PostConditions []Condition
}

func (u *UnaryRPCContract) validate() error {
	// TODO
	return nil
}

// Logger represents the logger interface
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// UnaryRPCCall represents an RPC call and its details
type UnaryRPCCall struct {
	MethodName string
	Request    interface{}
	Response   interface{}
	Error      error
}

// ServerContract is the contract which defined for a gRPC server
type ServerContract struct {
	logger Logger

	callsLock     sync.RWMutex
	unaryRPCCalls map[string]map[string][]*UnaryRPCCall

	contractsLock     sync.Mutex
	unaryRPCContracts map[string]*UnaryRPCContract
	serve             bool
}

// RPCCallHistory lets you to have access to the RPC calls which made during an RPC lifetime
type RPCCallHistory struct {
	requestID string
	sc        *ServerContract
}

// All returns all stored RPCs
func (h *RPCCallHistory) All() []*UnaryRPCCall {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	var res []*UnaryRPCCall
	for _, calls := range h.sc.unaryRPCCalls[h.requestID] {
		res = append(res, calls...)
	}
	return res
}

// Filter returns RPC calls to the given method
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

// NewServerContract create a ServerContract which has no RPC contracts registered
func NewServerContract(logger Logger) *ServerContract {
	return &ServerContract{
		logger:            logger,
		unaryRPCCalls:     make(map[string]map[string][]*UnaryRPCCall),
		unaryRPCContracts: make(map[string]*UnaryRPCContract),
	}
}

// RegisterUnaryRPCContract registers an RPC contract to the server contract
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

// UnaryServerInterceptor returns a new unary server interceptor for monitoring server contracts
func (sc *ServerContract) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	sc.contractsLock.Lock()
	defer sc.contractsLock.Unlock()
	sc.serve = true

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// sc.logger.Info("server info full method = ", info.FullMethod)

		var requestID string
		ctx, requestID = sc.generateRequestID(ctx)

		var c *UnaryRPCContract
		for _, contract := range sc.unaryRPCContracts {
			if eq := sameMethods(contract.Method, info.FullMethod); eq {
				// sc.logger.Info("contract method = ", getMethodName(contract.Method))
				c = contract
				break
			}
		}
		if c != nil {
			// sc.logger.Info("pre")
			for _, preCondition := range c.PreConditions {
				err := invokePreCondition(preCondition, req)
				if err != nil {
					sc.logger.Error(err)
				}
			}
		}

		resp, err := handler(ctx, req)

		if c != nil {
			// sc.logger.Info("post")
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

// UnaryClientInterceptor returns a new unary client interceptor for monitoring of RPC calls made by the client
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
