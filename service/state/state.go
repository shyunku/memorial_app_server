package state

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"
)

// State represents the current state of the application.
type State struct {
	Tasks      map[string]Task
	SubTasks   map[string]Subtask
	Categories map[string]Category
}

func NewState() *State {
	return &State{
		Tasks:      make(map[string]Task),
		SubTasks:   make(map[string]Subtask),
		Categories: make(map[string]Category),
	}
}

func (s *State) FromBytes(b []byte) error {
	// unmarshal state
	if err := json.Unmarshal(b, s); err != nil {
		return err
	}
	return nil
}

func (s *State) ToBytes() ([]byte, error) {
	// marshal state
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *State) Hash() Hash {
	fieldBytes := make([]byte, 0)
	values := reflect.ValueOf(*s)

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
	copiedTasks := make(map[string]Task)
	for k, v := range s.Tasks {
		copiedTasks[k] = v
	}

	copiedSubTasks := make(map[string]Subtask)
	for k, v := range s.SubTasks {
		copiedSubTasks[k] = v
	}

	copiedCategories := make(map[string]Category)
	for k, v := range s.Categories {
		copiedCategories[k] = v
	}

	return &State{
		Tasks:      copiedTasks,
		SubTasks:   copiedSubTasks,
		Categories: copiedCategories,
	}
}
