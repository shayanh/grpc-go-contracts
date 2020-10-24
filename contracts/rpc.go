package contracts

// UnaryRPCContract represents a contract for a unary RPC.
type UnaryRPCContract struct {
	MethodName     string
	PreConditions  []Condition
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
	ServiceName  string
	RPCContracts []*UnaryRPCContract
}

func getFullMethodName(serviceName string, methodName string) string {
	return "/" + serviceName + "/" + methodName
}
