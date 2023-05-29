package state

import (
	"fmt"
	"memorial_app_server/log"
	"memorial_app_server/util"
	"time"
)

const (
	TxUnknown = iota
	TxInitialize

	TxCreateTask
	TxDeleteTask
	TxUpdateTaskOrder
	TxUpdateTaskTitle
	TxUpdateTaskDueDate
	TxUpdateTaskMemo
	TxUpdateTaskDone
	TxUpdateTaskRepeatPeriod

	TxAddTaskCategory
	TxDeleteTaskCategory

	TxCreateSubtask
	TxDeleteSubtask
	TxUpdateSubtaskTitle
	TxUpdateSubtaskDueDate
	TxUpdateSubtaskDone

	TxCreateCategory
	TxDeleteCategory
)

func ExecuteTransaction(prevState *State, tx *Transaction, newBlockNumber int64) (*State, error) {
	state := prevState.Copy()

	switch tx.Type {
	case TxInitialize:
		return InitializeState(state, tx, newBlockNumber)
	case TxCreateTask:
		return CreateTask(state, tx)
	case TxDeleteTask:
		return DeleteTask(state, tx)
	case TxUpdateTaskOrder:
		return UpdateTaskOrder(state, tx)
	case TxUpdateTaskTitle:
		return UpdateTaskTitle(state, tx)
	case TxUpdateTaskDueDate:
		return UpdateTaskDueDate(state, tx)
	case TxUpdateTaskMemo:
		return UpdateTaskMemo(state, tx)
	case TxUpdateTaskDone:
		return UpdateTaskDone(state, tx)
	case TxUpdateTaskRepeatPeriod:
		return UpdateTaskRepeatPeriod(state, tx)
	case TxAddTaskCategory:
		return AddTaskCategory(state, tx)
	case TxDeleteTaskCategory:
		return DeleteTaskCategory(state, tx)
	case TxCreateSubtask:
		return CreateSubtask(state, tx)
	case TxDeleteSubtask:
		return DeleteSubtask(state, tx)
	case TxUpdateSubtaskTitle:
		return UpdateSubtaskTitle(state, tx)
	case TxUpdateSubtaskDueDate:
		return UpdateSubtaskDueDate(state, tx)
	case TxUpdateSubtaskDone:
		return UpdateSubtaskDone(state, tx)
	case TxCreateCategory:
		return CreateCategory(state, tx)
	case TxDeleteCategory:
		return DeleteCategory(state, tx)
	default:
		return nil, ErrInvalidTxType
	}
}

func InitializeState(prevState *State, tx *Transaction, newBlockNumber int64) (*State, error) {
	if newBlockNumber != 1 {
		return nil, fmt.Errorf("invalid block number for initialize state: %d, expected: 1", newBlockNumber)
	}

	var body TxInitializeBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	state := &State{
		Tasks:      body.Tasks,
		Categories: body.Categories,
	}

	return state, nil
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
		Subtasks:      map[string]Subtask{},
		Categories:    map[string]bool{},
	}

	// update next of previous
	if body.PrevTaskId != "" {
		prevTask := state.Tasks[body.PrevTaskId]
		prevTask.Next = body.Id
		state.Tasks[body.PrevTaskId] = prevTask
	}

	return state, nil
}

func DeleteTask(state *State, tx *Transaction) (*State, error) {
	var body TxDeleteTaskBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	// update next of previous
	if body.PrevTaskId != "" {
		prevTask := state.Tasks[body.PrevTaskId]

		if prevTask.Next != body.Id {
			log.Warnf("prevTask.Next(%s) != body.Id(%s)", prevTask.Next, body.Id)
			return nil, ErrStateMismatch
		}

		prevTask.Next = state.Tasks[body.Id].Next
		state.Tasks[body.PrevTaskId] = prevTask
	}

	delete(state.Tasks, body.Id)
	return state, nil
}

func UpdateTaskOrder(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskOrderBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	currentTask, ok := state.Tasks[body.Id]
	if !ok {
		log.Warnf("updating order task(%s) not found", body.Id)
		return nil, ErrStateMismatch
	}

	// get next Task
	nextTaskId := currentTask.Next

	// update previous task's next
	if body.PrevTaskId != "" {
		prevTask := state.Tasks[body.PrevTaskId]

		if prevTask.Next != body.Id {
			log.Warnf("prevTask.Next(%s) != body.Id(%s)", prevTask.Next, body.Id)
			return nil, ErrStateMismatch
		}

		prevTask.Next = nextTaskId
		state.Tasks[body.PrevTaskId] = prevTask
	}

	targetTask, targetTaskExists := state.Tasks[body.TargetTaskId]

	if body.AfterTarget {
		var targetTaskNextId string
		// update target task's next
		if targetTaskExists {
			targetTaskNextId = targetTask.Next
			targetTask.Next = body.Id
			state.Tasks[targetTask.Id] = targetTask
		} else {
			targetTaskNextId = ""
		}
		// update task's next
		currentTask.Next = targetTaskNextId
		state.Tasks[currentTask.Id] = currentTask
	} else {
		// update target prev task's next
		if body.TargetPrevTaskId != "" {
			targetPrevTask, ok := state.Tasks[body.TargetPrevTaskId]
			if !ok {
				log.Warnf("updating order targetPrevTask(%s) not found", body.TargetPrevTaskId)
				return nil, ErrStateMismatch
			}
			if targetPrevTask.Next != body.TargetTaskId {
				log.Warnf("targetPrevTask.Next(%s) != body.TargetTaskId(%s)", targetPrevTask.Next, body.TargetTaskId)
				return nil, ErrStateMismatch
			}
			targetPrevTask.Next = body.Id
			state.Tasks[targetPrevTask.Id] = targetPrevTask
		}
		// update task's next
		currentTask.Next = body.TargetTaskId
		state.Tasks[currentTask.Id] = currentTask
	}

	return state, nil
}

func UpdateTaskTitle(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskTitleBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.Id]
	if !ok {
		log.Warnf("updating title task(%s) not found", body.Id)
		return nil, ErrStateMismatch
	}

	task.Title = body.Title
	state.Tasks[body.Id] = task

	return state, nil
}

func UpdateTaskDueDate(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskDueDateBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.Id]
	if !ok {
		log.Warnf("updating dueDate task(%s) not found", body.Id)
		return nil, ErrStateMismatch
	}

	task.DueDate = body.DueDate
	state.Tasks[body.Id] = task

	return state, nil
}

func UpdateTaskMemo(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskMemoBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.Id]
	if !ok {
		log.Warnf("updating memo task(%s) not found", body.Id)
		return nil, ErrStateMismatch
	}

	task.Memo = body.Memo
	state.Tasks[body.Id] = task

	return state, nil
}

func UpdateTaskDone(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskDoneBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("updating done task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	if task.RepeatPeriod != "" {
		repeatStartAt := task.RepeatStartAt
		if repeatStartAt == 0 {
			repeatStartAt = task.DueDate
		}
		repeatPeriod := task.RepeatPeriod
		// date milli (int64) -> time
		nextDueDate := time.Unix(0, repeatStartAt*int64(time.Millisecond))
		// compare with milliseconds
		for nextDueDate.Before(time.Now()) || nextDueDate.UnixNano() / int64(time.Millisecond) <= task.DueDate {
			switch repeatPeriod {
			case "day":
				nextDueDate = nextDueDate.AddDate(0, 0, 1)
			case "week":
				nextDueDate = nextDueDate.AddDate(0, 0, 7)
			case "month":
				nextDueDate = nextDueDate.AddDate(0, 1, 0)
			case "year":
				nextDueDate = nextDueDate.AddDate(1, 0, 0)
			}
		}

		task.Done = false
		task.DoneAt = body.DoneAt
		task.DueDate = nextDueDate.UnixNano() / int64(time.Millisecond)
	} else {
		task.Done = body.Done
		task.DoneAt = body.DoneAt
	}

	state.Tasks[body.TaskId] = task
	return state, nil
}

func UpdateTaskRepeatPeriod(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateTaskRepeatPeriodBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("updating repeatPeriod task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	task.RepeatPeriod = body.RepeatPeriod
	// update repeat period start time
	repeatStartAt := task.RepeatStartAt
	if repeatStartAt == 0 {
		repeatStartAt = task.DueDate
		if repeatStartAt != 0 {
			task.RepeatStartAt = repeatStartAt
		}
	}

	state.Tasks[body.TaskId] = task

	return state, nil
}

func AddTaskCategory(state *State, tx *Transaction) (*State, error) {
	var body TxAddTaskCategoryBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("adding category task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	_, ok = state.Categories[body.CategoryId]
	if !ok {
		log.Warnf("adding category category(%s) not found", body.CategoryId)
		return nil, ErrStateMismatch
	}

	task.Categories[body.CategoryId] = true
	state.Tasks[body.TaskId] = task

	return state, nil
}

func DeleteTaskCategory(state *State, tx *Transaction) (*State, error) {
	var body TxDeleteTaskCategoryBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("deleting category task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	_, ok = task.Categories[body.CategoryId]
	if !ok {
		log.Warnf("deleting category category(%s) not found", body.CategoryId)
		return nil, ErrStateMismatch
	}

	delete(task.Categories, body.CategoryId)
	state.Tasks[body.TaskId] = task

	return state, nil
}

func CreateSubtask(state *State, tx *Transaction) (*State, error) {
	var body TxCreateSubtaskBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("adding subtask task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	task.Subtasks[body.SubtaskId] = Subtask{
		Id:        body.SubtaskId,
		Title:     body.Title,
		CreatedAt: body.CreatedAt,
		DoneAt:    body.DoneAt,
		Done:      body.Done,
		DueDate:   body.DueDate,
	}
	state.Tasks[body.TaskId] = task

	return state, nil
}

func DeleteSubtask(state *State, tx *Transaction) (*State, error) {
	var body TxDeleteSubtaskBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("deleting subtask task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	_, ok = task.Subtasks[body.SubtaskId]
	if !ok {
		log.Warnf("deleting subtask subtask(%s) not found", body.SubtaskId)
		return nil, ErrStateMismatch
	}

	delete(task.Subtasks, body.SubtaskId)
	state.Tasks[body.TaskId] = task

	return state, nil
}

func UpdateSubtaskTitle(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateSubtaskTitleBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("updating subtask title task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	subtask, ok := task.Subtasks[body.SubtaskId]
	if !ok {
		log.Warnf("updating subtask title subtask(%s) not found", body.SubtaskId)
		return nil, ErrStateMismatch
	}

	subtask.Title = body.Title
	task.Subtasks[body.SubtaskId] = subtask
	state.Tasks[body.TaskId] = task

	return state, nil
}

func UpdateSubtaskDueDate(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateSubtaskDueDateBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("updating subtask dueDate task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	subtask, ok := task.Subtasks[body.SubtaskId]
	if !ok {
		log.Warnf("updating subtask dueDate subtask(%s) not found", body.SubtaskId)
		return nil, ErrStateMismatch
	}

	subtask.DueDate = body.DueDate
	task.Subtasks[body.SubtaskId] = subtask
	state.Tasks[body.TaskId] = task

	return state, nil
}

func UpdateSubtaskDone(state *State, tx *Transaction) (*State, error) {
	var body TxUpdateSubtaskDoneBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	task, ok := state.Tasks[body.TaskId]
	if !ok {
		log.Warnf("updating subtask done task(%s) not found", body.TaskId)
		return nil, ErrStateMismatch
	}

	subtask, ok := task.Subtasks[body.SubtaskId]
	if !ok {
		log.Warnf("updating subtask done subtask(%s) not found", body.SubtaskId)
		return nil, ErrStateMismatch
	}

	subtask.Done = body.Done
	subtask.DoneAt = body.DoneAt
	task.Subtasks[body.SubtaskId] = subtask
	state.Tasks[body.TaskId] = task

	return state, nil
}

func CreateCategory(state *State, tx *Transaction) (*State, error) {
	var body TxCreateCategoryBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	state.Categories[body.Id] = Category{
		Id:        body.Id,
		Title:     body.Title,
		Secret:    body.Secret,
		Locked:    body.Locked,
		Color:     body.Color,
		CreatedAt: body.CreatedAt,
	}

	return state, nil
}

func DeleteCategory(state *State, tx *Transaction) (*State, error) {
	var body TxDeleteCategoryBody
	if err := util.InterfaceToStruct(tx.Content, &body); err != nil {
		return nil, err
	}

	// if tasks that contains category exists, return error
	alreadyUsing := 0
	for _, task := range state.Tasks {
		for categoryId, _ := range task.Categories {
			if categoryId == body.Id {
				alreadyUsing++
				break
			}
		}
	}
	if alreadyUsing > 0 {
		return nil, fmt.Errorf("category is already used by %d tasks", alreadyUsing)
	}

	delete(state.Categories, body.Id)
	return state, nil
}
