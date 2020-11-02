package contracts

// UnaryRPCContract represents a contract for a unary RPC.
type UnaryRPCContract struct {
	// MethodName is the method name only, without the service name or package name.
	MethodName string
	// PreConditions are conditions that must always be true just prior to the execution of the RPC.
	// Each PreCondition should be a function with the following signature:
	// `func(req *Request) error`.
	PreConditions []Condition
	// PostConditions are conditions that must always be true just after the execution of the RPC.
	// Each PostCondition should be a function with the following signature:
	// `func(resp *Response, respErr error, req *Request, calls contracts.RPCCallHistory) error`.
	PostConditions []Condition
}

func (u *UnaryRPCContract) validate() error {
	for _, c := range u.PreConditions {
		if err := validatePreCondition(c); err != nil {
			return err
		}
	}
	for _, c := range u.PostConditions {
		if err := validatePostCondition(c); err != nil {
			return err
		}
	}
	return nil
}

// ServiceContract is a contract defined for a gRPC service.
type ServiceContract struct {
	// ServiceName is name the gRPC service, i.e., package.service.
	ServiceName string
	// RPCContracts are the contracts defined for RPCs of the service.
	RPCContracts []*UnaryRPCContract
}

func getFullMethodName(serviceName string, methodName string) string {
	return "/" + serviceName + "/" + methodName
}
