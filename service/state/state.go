package state

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"
)

// State represents the current state of the application.
type State struct {
	tasks      map[int64]Task
	subTasks   map[int64]Subtask
	categories map[int64]Category
}

func NewState() *State {
	return &State{
		tasks:      make(map[int64]Task),
		subTasks:   make(map[int64]Subtask),
		categories: make(map[int64]Category),
	}
}

func (s *State) FromBytes(b []byte) error {
	// unmarshal state
	if err := json.Unmarshal(b, s); err != nil {
		return err
	}
	return nil
}

func (s *State) Hash() Hash {
	fieldBytes := make([]byte, 0)
	values := reflect.ValueOf(s)

	for i := 0; i < values.NumField(); i++ {
		value := values.Field(i)
		raw := value.Interface()
		bytes, _ := json.Marshal(raw)
		fieldBytes = append(fieldBytes, bytes...)
	}

	hash := sha256.Sum256(fieldBytes)
	return hash
}

func (s *State) Copy() *State {
	copiedTasks := make(map[int64]Task)
	for k, v := range s.tasks {
		copiedTasks[k] = v
	}

	copiedSubTasks := make(map[int64]Subtask)
	for k, v := range s.subTasks {
		copiedSubTasks[k] = v
	}

	copiedCategories := make(map[int64]Category)
	for k, v := range s.categories {
		copiedCategories[k] = v
	}

	return &State{
		tasks:      copiedTasks,
		subTasks:   copiedSubTasks,
		categories: copiedCategories,
	}
}
