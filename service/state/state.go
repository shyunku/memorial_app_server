package state

import (
	"crypto/sha256"
	"encoding/json"
	"reflect"
)

// State represents the current state of the application.
type State struct {
	categories map[int]*Category
	tasks      map[int]*Task
}

func NewState() *State {
	return &State{
		categories: make(map[int]*Category),
		tasks:      make(map[int]*Task),
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
