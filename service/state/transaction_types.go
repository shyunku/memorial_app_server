package state

const (
	TxInitialize = iota
	TxCreateTask
)

type TxCreateTaskBody struct {
	Id            string              `json:"tid"`
	Title         string              `json:"title"`
	CreatedAt     int64               `json:"createdAt"`
	DoneAt        int64               `json:"doneAt"`
	Memo          string              `json:"memo"`
	Done          bool                `json:"done"`
	DueDate       int64               `json:"dueDate"`
	RepeatPeriod  string              `json:"repeatPeriod"`
	RepeatStartAt int64               `json:"repeatStartAt"`
	Categories    map[string]Category `json:"Categories"`

	PrevTaskId string `json:"prevTaskId"`
}
