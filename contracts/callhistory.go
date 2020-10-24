package contracts

import (
	"errors"
	"sort"
)

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

type CallSet []*UnaryRPCCall

// All returns all stored RPCs
func (h *RPCCallHistory) All() CallSet {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	var res CallSet
	for _, calls := range h.sc.unaryRPCCalls[h.requestID] {
		res = append(res, calls...)
	}
	return res
}

// Filter returns RPC calls to the given method
func (h *RPCCallHistory) Filter(serviceName, methodName string) CallSet {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	fullMethod := getFullMethodName(serviceName, methodName)
	src := h.sc.unaryRPCCalls[h.requestID][fullMethod]
	res := make([]*UnaryRPCCall, len(src))
	copy(res, src)
	return res
}

func (cs CallSet) Successful() CallSet {
	var res CallSet
	for _, call := range cs {
		if call.Error == nil {
			res = append(res, call)
		}
	}
	return res
}

func (cs CallSet) Ordered() CallSet {
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Order < cs[i].Order
	})
	return cs
}

func (cs CallSet) Empty() bool {
	return len(cs) <= 0
}

func (cs CallSet) Count() int {
	return len(cs)
}

func (cs CallSet) First() (*UnaryRPCCall, error) {
	if cs.Empty() {
		return nil, errors.New("No call exists")
	}
	return cs[0], nil
}
