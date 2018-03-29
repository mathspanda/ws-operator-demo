package controller

type TaskType string

const (
	TaskTypeAdd    TaskType = "ADD"
	TaskTypeUpdate TaskType = "UPDATE"
	TaskTypeDelete TaskType = "DELETE"
)

type CRDTask struct {
	CRDTaskType TaskType
	CRDObj      interface{}
}
