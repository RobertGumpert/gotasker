package itask

type EventManageTasks func(task ITask) (deleteTasks map[string]struct{})

type IManager interface {
	CreateTask(t Type, key string, send, update, fields interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateTaskState) (task ITask, err error)
	ModifyTaskAsTrigger(trigger ITask, dependents ...ITask) (task ITask, err error)
	//
	// RUN TASKS
	//
	RunTask(task ITask) (doTaskAsDefer, sendToErrorChannel bool)
	AddTaskAndTask(task ITask) (err error)
	RunDependentTasks(task ITask)
	RunDeferTasks(runDependentTasks bool)
	RunDeferByTimer()
	//
	// MANAGE TASKS
	//
	CreateError(e error, taskKey string, task ITask) (err IError)
	SendErrorToErrorChannel(err IError)
	SetUpdateForTask(key string, somethingUpdateContext interface{})
	ManageUpdates()
	ManageCompleted()
	//
	// PROPERTIES
	//
	GetChannelError() (channelForSendErrors chan IError)
	GetSizeQueue() (sizeOfQueue int64)
	QueueIsFilled(countTasks int64) (isFilled bool)
	DeleteTasksByKeys(keys map[string]struct{})
	FindTaskByKey(key string) (findTask ITask, err error)
	FindRunBanTriggers() (runBanTriggers []ITask)
	FindRunBanSimpleTasks() (runBanTasks []ITask)
	FindDependentTasksIfTriggerNotExist(triggerKey string) (dependentsTasks []ITask)
	SetRunBan(tasks ...ITask)
	SetRunBanInQueue(tasks ...ITask)
	TakeOffRunBanInQueue(tasks ...ITask)
	TriggerIsCompleted(trigger ITask) (isCompleted bool, dependentTasks map[string]bool, err error)
}

type ITaskEventFacade interface {
	EventManageTasks(task ITask) (deleteTasks map[string]struct{})
	EventUpdateTaskState(task ITask, somethingUpdateContext interface{}) (err error, sendToErrorChannel bool)
	EventRunTask(task ITask) (doTaskAsDefer, sendToErrorChannel bool, err error)
}
