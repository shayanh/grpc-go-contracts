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

// ServerContract is the contract which defined for a gRPC server
type ServerContract struct {
	logger Logger

	callsLock     sync.RWMutex
	unaryRPCCalls map[string]map[string][]*UnaryRPCCall
	callCnt       map[string]int

	contractsLock     sync.Mutex
	unaryRPCContracts map[string]*UnaryRPCContract
	serve             bool
}

// NewServerContract create a ServerContract which has no RPC contracts registered
func NewServerContract(logger Logger) *ServerContract {
	return &ServerContract{
		logger:            logger,
		unaryRPCCalls:     make(map[string]map[string][]*UnaryRPCCall),
		callCnt:           make(map[string]int),
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
		delete(sc.callCnt, requestID)
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for monitoring server contracts
func (sc *ServerContract) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	sc.contractsLock.Lock()
	defer sc.contractsLock.Unlock()
	sc.serve = true

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var requestID string
		ctx, requestID = sc.generateRequestID(ctx)

		var c *UnaryRPCContract
		for _, contract := range sc.unaryRPCContracts {
			if eq := sameMethods(contract.Method, info.FullMethod); eq {
				c = contract
				break
			}
		}
		if c != nil {
			for _, preCondition := range c.PreConditions {
				err := invokePreCondition(preCondition, req)
				if err != nil {
					sc.logger.Error(err)
				}
			}
		}

		resp, err := handler(ctx, req)

		if c != nil {
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

			if _, ok := sc.unaryRPCCalls[requestID]; !ok {
				sc.unaryRPCCalls[requestID] = make(map[string][]*UnaryRPCCall)
				sc.callCnt[requestID] = 0
			}
			call := &UnaryRPCCall{
				MethodName: method,
				Request:    req,
				Response:   reply,
				Error:      err,
				Order:      sc.callCnt[requestID],
			}
			sc.unaryRPCCalls[requestID][method] = append(sc.unaryRPCCalls[requestID][method], call)
			sc.callCnt[requestID]++
		}
		return err
	}
}
