package contracts

import (
	"errors"
	"reflect"
)

// Condition represents a pre or postcondition. Must be a function with the specified signature.
type Condition interface{}

func invokeCondition(c Condition, args ...interface{}) error {
	v := reflect.ValueOf(c)
	t := v.Type()
	if t.NumIn() != len(args) {
		return errors.New("wrong number of arguments for given condition")
	}
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

func isError(t reflect.Type) bool {
	errorInterface := reflect.TypeOf(new(error)).Elem()
	return t.Implements(errorInterface)
}

// Precondition function signature is `func(req *Request) error`.
func validatePreCondition(c Condition) error {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Func {
		return errors.New("PreCondition must be a function")
	}
	t := v.Type()
	if t.NumIn() != 1 {
		return errors.New("PreCondition wrong number of arguments")
	}
	if t.NumOut() != 1 {
		return errors.New("PreCondition wrong number of return values")
	}
	if !isError(t.Out(0)) {
		return errors.New("PreCondition return type mismatch")
	}
	return nil
}

// Postcondition function signature is
// `func(resp *Response, respErr error, req *Request, calls contracts.RPCCallHistory) error`.
func validatePostCondition(c Condition) error {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Func {
		return errors.New("PostCondition must be a function")
	}
	t := v.Type()
	if t.NumIn() != 4 {
		return errors.New("PostCondition wrong number of arguments")
	}
	if !isError(t.In(1)) {
		return errors.New("PostCondition input type mismatch")
	}
	if t.In(3) != reflect.TypeOf(new(RPCCallHistory)).Elem() {
		return errors.New("PostCondition input type mismatch")
	}
	if t.NumOut() != 1 {
		return errors.New("PostCondition wrong number of return values")
	}
	if !isError(t.Out(0)) {
		return errors.New("PostCondition return type mismatch")
	}
	return nil
}
