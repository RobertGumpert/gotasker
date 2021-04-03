package itask

type TaskConstructor func() (task ITask, err error)

type ISteward interface {
	CanAddTask(countTasks int64) (havePlaceInQueue bool)
	RunTask(task ITask) (err error)
	CreateTask(taskType Type, taskKey string, taskSendContext, taskUpdateContext, taskCustomFieldContext interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateTaskState) TaskConstructor
	CreateTaskAndRun(constructor TaskConstructor) (task ITask, err error)
	CreateTriggerAndRun(triggerConstructor TaskConstructor, dependentConstructors ...TaskConstructor) (trigger ITask, err error)
	ModifyTaskAsTrigger(triggerConstructor TaskConstructor, dependentConstructors ...TaskConstructor) (trigger ITask, err error)
	UpdateTask(key string, updateContext interface{}) (err error)
	FindTasksByKeys(keys map[string]struct{}) (findTasks []ITask)
	FindTaskByKey(key string) (findTask ITask, err error)
	GetTasksSavingOrCompleted() (tasks []ITask)
	GetTasksInQueue() (tasks []ITask)
	GetTasksCompletedWithError() (tasks []ITask)
}
