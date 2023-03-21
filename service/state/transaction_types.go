package state

const (
	TxInitialize = iota
	TxCreateTask
)

type CreateTaskTx Task
