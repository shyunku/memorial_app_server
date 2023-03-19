package service

func ExecuteTransaction(tx *Transaction) error {
	switch tx.Type {
	case TxInitialState:
		return InitialState(tx)
	case TxCreateTask:
		return CreateTask(tx)
	default:
		return ErrInvalidTxType
	}
}

func InitialState(tx *Transaction) error {
	panic("not implemented")
}

func CreateTask(tx *Transaction) error {
	panic("not implemented")
}
