package itask

import "time"

type EventManageTasks func(task ITask) (deleteTasks, saveTasks map[string]struct{})

type IManager interface {
	//
	// PROPERTIES
	//
	SetSizeQueue(size int64)
	GetSizeQueue() (size int64)
	UseTimeoutForRunning(timeout time.Duration)
	SetEventManageTasks(event EventManageTasks)
	GetEventManageTasks() (event EventManageTasks)
	//
	// CREATE TASK
	//
	CreateTask(t Type, key string, send, update, fields interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateTaskState) (task ITask, err error)
	ModifyTaskAsTrigger(trigger ITask, dependents ...ITask) (task ITask, err error)
	//
	// MANAGE TASKS
	//
	UpdateTask(key string, updateContext interface{}) (err error)
	DeleteTasksByKeys(keys map[string]struct{})
	FindTasksByKeys(keys map[string]struct{}) (findTasks []ITask)
	FindTaskByKey(key string) (findTask ITask, err error)
	GetTasksSavingOrCompleted() (tasks []ITask)
	GetTasksInQueue() (tasks []ITask)
	GetTasksCompletedWithError() (tasks []ITask)
	//
	// RUN TASKS
	//
	RunTask(task ITask) (err error)
	RunDeferTasks()
	RunDependentTasks(task ITask)
	RunDeferByTimer()
	//
	// MANAGE QUEUES
	//
	ManageUpdateTasks()
	ManageCompletedTasks()
}

