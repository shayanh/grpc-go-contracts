package contracts

import (
	"reflect"
	"runtime"
	"strings"
)

type Method interface{}

func getMethodName(method Method) string {
	return runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
}

// TODO double check correctness of this function
func sameMethods(method Method, fullMethodName string) bool {
	tmp1 := strings.Split(getMethodName(method), ".")
	tmp2 := strings.Split(fullMethodName, "/")
	m1 := tmp1[len(tmp1)-1]
	m2 := tmp2[len(tmp2)-1]
	return m1 == m2
}
