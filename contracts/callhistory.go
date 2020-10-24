package contracts

// UnaryRPCCall represents an RPC call and its details
type UnaryRPCCall struct {
	FullMethod string
	Request    interface{}
	Response   interface{}
	Error      error
	Order      int
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
func (h *RPCCallHistory) Filter(serviceName, methodName string) []*UnaryRPCCall {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	fullMethod := getFullMethodName(serviceName, methodName)
	src := h.sc.unaryRPCCalls[h.requestID][fullMethod]
	res := make([]*UnaryRPCCall, len(src))
	copy(res, src)
	return res
}
