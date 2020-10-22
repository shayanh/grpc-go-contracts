package contracts

func getFullMethodName(serviceName string, methodName string) string {
	return "/" + serviceName + "/" + methodName
}
