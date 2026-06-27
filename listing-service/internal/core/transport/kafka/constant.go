package core_kafka

type ContextKey string

const (
	OperationID ContextKey = "operationID"
	OpUserID    ContextKey = "opUserID"
)

var FlushTimeOut = 5000
