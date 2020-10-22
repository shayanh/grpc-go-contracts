package contracts

// UnaryRPCCall represents an RPC call and its details
type UnaryRPCCall struct {
	MethodName string
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
func (h *RPCCallHistory) Filter(method string) []*UnaryRPCCall {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	src := h.sc.unaryRPCCalls[h.requestID][method]
	res := make([]*UnaryRPCCall, len(src))
	copy(res, src)
	return res
}
