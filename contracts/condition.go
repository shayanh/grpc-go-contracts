package contracts

import "reflect"

// Condition represents a pre or post condition. Must be a function with specified signature
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
