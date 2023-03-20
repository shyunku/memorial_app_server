package state

func ExecuteTransaction(prevState *State, tx *Transaction) (*State, error) {
	switch tx.Type {
	case TxInitialize:
		return InitialState(prevState, tx)
	case TxCreateTask:
		return CreateTask(prevState, tx)
	default:
		return nil, ErrInvalidTxType
	}
}

func InitialState(prevState *State, tx *Transaction) (*State, error) {
	panic("not implemented")
}

func CreateTask(prevState *State, tx *Transaction) (*State, error) {
	panic("not implemented")
}
