package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"memorial_app_server/log"
	"reflect"
)

// State represents the current state of the application.
type State struct {
	Tasks      map[string]Task     `json:"tasks"`
	Categories map[string]Category `json:"categories"`
}

func NewState() *State {
	return &State{
		Tasks:      make(map[string]Task),
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

func (s *State) Validate() error {
	log.Debug("Validating new state...")
	// validate tasks links
	type taskNode struct {
		Id     string
		NextId string
		PrevId string
	}

	taskNodes := make(map[string]taskNode)
	for _, task := range s.Tasks {
		taskNodes[task.Id] = taskNode{
			Id:     task.Id,
			NextId: task.Next,
		}
	}

	// rewrite task nodes with prev
	for _, node := range taskNodes {
		if node.NextId != "" {
			nextNode := taskNodes[node.NextId]
			nextNode.PrevId = node.Id
			taskNodes[node.NextId] = nextNode
		}
	}

	// find first task
	var firstTaskId string
	for _, node := range taskNodes {
		if node.PrevId == "" {
			firstTaskId = node.Id
			break
		}
	}

	sortedTasks := make(map[string]taskNode)
	if firstTaskId != "" {
		// first task exists, iterate through tasks
		iterator := taskNodes[firstTaskId]
		for {
			sortedTasks[iterator.Id] = iterator
			if iterator.NextId == "" {
				break
			}
			iterator = taskNodes[iterator.NextId]
		}
	}

	// check if all tasks are sorted
	if len(sortedTasks) != len(s.Tasks) {
		// collect unsorted task ids
		unsortedTaskIds := make([]string, 0)
		for _, task := range s.Tasks {
			_, exists := sortedTasks[task.Id]
			if !exists {
				unsortedTaskIds = append(unsortedTaskIds, task.Id)
			}
		}
		return fmt.Errorf("tasks are not sorted: %v", unsortedTaskIds)
	}

	// check if tasks' categories exists
	for _, task := range s.Tasks {
		for categoryId, category := range task.Categories {
			_, exists := s.Categories[categoryId]
			if !exists {
				return fmt.Errorf("task %s has non-existing category %s", task.Id, category)
			}
		}
	}

	return nil
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

	copiedCategories := make(map[string]Category)
	for k, v := range s.Categories {
		copiedCategories[k] = v
	}

	return &State{
		Tasks:      copiedTasks,
		Categories: copiedCategories,
	}
}
