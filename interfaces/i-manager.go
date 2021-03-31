package interfaces

type EventManageTasks func(task ITask) (deleteTasks []string)

type IManager interface {
	SetSizeQueue(size int64)
	GetSizeQueue() (size int64)
	//
	CreateTask(t Type, key string, result, fields interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateState) (task ITask, err error)
	GetTask(key string) (task ITask, err error)
	ModifyTaskAsTrigger(trigger ITask, dependents ...ITask) (task ITask, err error)
	//
	Add(task ITask) (err error)
	RunDeferTasks()
	RunDependentTasks(trigger ITask) (err error)
	ScanChannelCompletedTasks()
	//
	SetEventManageTasks(event EventManageTasks)
	GetEventManageTasks() (event EventManageTasks)
	UpdateTaskStateExecute(key string, result interface{}) (err error)
}
