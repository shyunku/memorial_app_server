package state

import (
	"encoding/json"
	"memorial_app_server/log"
)

func ExecuteTransaction(prevState *State, tx *Transaction) (*State, error) {
	state := prevState.Copy()
	switch tx.Type {
	case TxInitialize:
		return InitialState(state, tx)
	case TxCreateTask:
		return CreateTask(state, tx)
	default:
		return nil, ErrInvalidTxType
	}
}

func InitialState(prevState *State, tx *Transaction) (*State, error) {
	panic("not implemented")
}

func CreateTask(state *State, tx *Transaction) (*State, error) {
	var task Task
	if err := json.Unmarshal(tx.Content, &task); err != nil {
		log.Fatal("Error decoding JSON: ", err)
	}
	state.tasks[task.Id] = task
	return state, nil
}
