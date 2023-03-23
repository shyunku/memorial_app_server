package state

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
	PrevTaskId    string              `json:"prevTaskId"`
}

type TxDeleteTaskBody struct {
	Id         string `json:"tid"`
	PrevTaskId string `json:"prevTaskId"`
}

type TxUpdateTaskOrderBody struct {
	Id               string `json:"tid"`
	TargetTaskId     string `json:"targetTaskId"`
	AfterTarget      bool   `json:"afterTarget"`
	PrevTaskId       string `json:"prevTaskId"`
	TargetPrevTaskId string `json:"targetPrevTaskId"`
}

type TxUpdateTaskTitleBody struct {
	Id    string `json:"tid"`
	Title string `json:"title"`
}

type TxUpdateTaskDueDateBody struct {
	Id      string `json:"tid"`
	DueDate int64  `json:"dueDate"`
}

type TxUpdateTaskMemoBody struct {
	Id   string `json:"tid"`
	Memo string `json:"memo"`
}

type TxAddCategoryBody struct {
	Id     string `json:"cid"`
	Title  string `json:"title"`
	Secret bool   `json:"secret"`
	Locked bool   `json:"locked"`
	Color  string `json:"color"`
}

type TxDeleteCategoryBody struct {
	Id string `json:"cid"`
}

type TxAddTaskCategoryBody struct {
	TaskId     string `json:"tid"`
	CategoryId string `json:"cid"`
}

type TxUpdateTaskDoneBody struct {
	TaskId string `json:"tid"`
	Done   bool   `json:"done"`
	DoneAt int64  `json:"doneAt"`
}

type TxDeleteTaskCategoryBody struct {
	TaskId     string `json:"tid"`
	CategoryId string `json:"cid"`
}

type TxUpdateTaskRepeatPeriodBody struct {
	TaskId       string `json:"tid"`
	RepeatPeriod string `json:"repeatPeriod"`
}

type TxAddSubtaskBody struct {
	TaskId    string `json:"tid"`
	SubtaskId string `json:"sid"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"createdAt"`
	DueDate   int64  `json:"dueDate"`
	Done      bool   `json:"done"`
	DoneAt    int64  `json:"doneAt"`
}

type TxDeleteSubtaskBody struct {
	TaskId    string `json:"tid"`
	SubtaskId string `json:"sid"`
}

type TxUpdateSubtaskTitleBody struct {
	TaskId    string `json:"tid"`
	SubtaskId string `json:"sid"`
	Title     string `json:"title"`
}

type TxUpdateSubtaskDueDateBody struct {
	TaskId    string `json:"tid"`
	SubtaskId string `json:"sid"`
	DueDate   int64  `json:"dueDate"`
}

type TxUpdateSubtaskDoneBody struct {
	TaskId    string `json:"tid"`
	SubtaskId string `json:"sid"`
	Done      bool   `json:"done"`
	DoneAt    int64  `json:"doneAt"`
}
