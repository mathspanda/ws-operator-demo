package controller

type TaskType string

const (
	TaskTypeAdd          TaskType = "ADD"
	TaskTypeUpdate       TaskType = "UPDATE"
	TaskTypeDelete       TaskType = "DELETE"
	TaskTypeUpdateStatus TaskType = "UPDATE_STATUS"
)

type CRDTask struct {
	CRDTaskType   TaskType
	CRDObj        interface{}
	CRDFObjStatus interface{}
}
