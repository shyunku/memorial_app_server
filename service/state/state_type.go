package state

type Task struct {
	Id            string `json:"tid"`
	Title         string `json:"title"`
	CreatedAt     int64  `json:"createdAt"`
	DoneAt        int64  `json:"doneAt"`
	Memo          string `json:"memo"`
	Done          bool   `json:"done"`
	DueDate       int64  `json:"dueDate"`
	Next          string `json:"next"`
	RepeatPeriod  string `json:"repeatPeriod"`
	RepeatStartAt int64  `json:"repeatStartAt"`

	Subtasks   map[string]Subtask `json:"subtasks"`
	Categories map[string]bool   `json:"categories"`
}

type Subtask struct {
	Id        string `json:"sid"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"createdAt"`
	DoneAt    int64  `json:"doneAt"`
	DueDate   int64  `json:"dueDate"`
	Done      bool   `json:"done"`
}

type Category struct {
	Id     string `json:"cid"`
	Title  string `json:"title"`
	Secret bool   `json:"secret"`
	Locked bool   `json:"locked"`
	Color  string `json:"color"`
}
