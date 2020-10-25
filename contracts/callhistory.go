package contracts

import (
	"errors"
	"sort"
)

// UnaryRPCCall represents an RPC call and its details.
type UnaryRPCCall struct {
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
	// Request is the body of the RPC request.
	Request interface{}
	// Response is the body of the RPC response.
	Response interface{}
	// Error is the error that returned with the RPC response.
	Error error
	// Order represents the invocation time of RPCs in ascending order.
	Order int
}

// RPCCallHistory lets you have access to the RPC calls made during an RPC lifetime.
type RPCCallHistory struct {
	requestID string
	sc        *ServerContract
}

// CallSet is a set of UnaryRPCCalls that provides APIs for simpler usage.
type CallSet []*UnaryRPCCall

// All returns all invoked RPCs.
func (h *RPCCallHistory) All() CallSet {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	var res CallSet
	for _, calls := range h.sc.unaryRPCCalls[h.requestID] {
		res = append(res, calls...)
	}
	return res
}

// Filter returns RPC calls to the given method.
// serviceName is name the gRPC service, i.e., package.service.
// methodName is the method name only, without the service name or package name.
func (h *RPCCallHistory) Filter(serviceName, methodName string) CallSet {
	h.sc.callsLock.RLock()
	defer h.sc.callsLock.RUnlock()

	fullMethod := getFullMethodName(serviceName, methodName)
	src := h.sc.unaryRPCCalls[h.requestID][fullMethod]
	res := make([]*UnaryRPCCall, len(src))
	copy(res, src)
	return res
}

// Successful filters successful RPC calls and returns them.
func (cs CallSet) Successful() CallSet {
	var res CallSet
	for _, call := range cs {
		if call.Error == nil {
			res = append(res, call)
		}
	}
	return res
}

// Ordered sorts RPC calls in the call set by their invocation time.
func (cs CallSet) Ordered() CallSet {
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Order < cs[j].Order
	})
	return cs
}

// Empty returns true if the call set is empty.
func (cs CallSet) Empty() bool {
	return len(cs) <= 0
}

// Count returns the number of RPC calls in the call set.
func (cs CallSet) Count() int {
	return len(cs)
}

// First returns the first RPC call in the call set (if exists).
func (cs CallSet) First() (*UnaryRPCCall, error) {
	if cs.Empty() {
		return nil, errors.New("No call exists")
	}
	return cs[0], nil
}
