package itask

type IError interface {
	SetTaskKey(taskKey string)
	SetTask(task ITask)
	SetError(err error)
	GetTaskKey() (taskKey string)
	GetTaskIfExist() (task ITask, exist bool)
	GetError() (err error)
}
