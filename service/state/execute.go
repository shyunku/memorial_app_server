package state

import "memorial_app_server/util"

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
	var body TxCreateTaskBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	state.Tasks[body.Id] = Task{
		Id:            body.Id,
		Title:         body.Title,
		CreatedAt:     body.CreatedAt,
		DoneAt:        body.DoneAt,
		Memo:          body.Memo,
		Done:          body.Done,
		DueDate:       body.DueDate,
		RepeatPeriod:  body.RepeatPeriod,
		RepeatStartAt: body.RepeatStartAt,
	}

	// update next of previous
	if body.PrevTaskId != "" {
		prevTask := state.Tasks[body.PrevTaskId]
		prevTask.Next = body.Id
		state.Tasks[body.PrevTaskId] = prevTask
	}

	return state, nil
}
