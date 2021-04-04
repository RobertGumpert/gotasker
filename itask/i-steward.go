package itask

type TaskConstructor func() (task ITask, err error)

type ISteward interface {
	//
	// PROPERTIES
	//
	FindTasksByKeys(keys map[string]struct{}) (findTasks []ITask)
	FindTaskByKey(key string) (findTask ITask, err error)
	GetTasksSavingOrCompleted() (tasks []ITask)
	GetTasksInQueue() (tasks []ITask)
	GetTasksCompletedWithError() (tasks []ITask)
	TriggerIsCompleted(trigger ITask) (isCompleted bool, dependentTasksCompletedFlags map[string]bool, err error)
	//
	// CREATE TASKS
	//
	CreateTask(taskType Type, taskKey string, taskSendContext, taskUpdateContext, taskCustomFieldContext interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateTaskState) TaskConstructor
	ModifyTaskAsTrigger(triggerTask ITask, dependentTasks ...ITask) (trigger ITask, err error)
	//
	// CREATE AND RUN TASKS
	//
	CreateTaskAndRun(constructor TaskConstructor) (task ITask, err error)
	CreateTriggerAndRun(triggerConstructor TaskConstructor, dependentConstructors ...TaskConstructor) (trigger ITask, err error)
	//
	// RUN TASK
	//
	RunTask(task ITask) (err error)
	CanAddTask(countTasks int64) (havePlaceInQueue bool)
	UpdateTask(key string, updateContext interface{}) (err error)
}
