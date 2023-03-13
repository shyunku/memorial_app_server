package service

func ExecuteTransaction(tx *Transaction) error {
	switch tx.Type {
	case TxCreateTask:
		return ExecuteCreateTask(tx)
	default:
		return ErrInvalidTxType
	}
}

func ExecuteCreateTask(tx *Transaction) error {
	panic("not implemented")
}
