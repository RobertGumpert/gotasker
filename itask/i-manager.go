package itask

import "time"

type EventManageTasks func(task ITask) (deleteTasks, saveTasks map[string]struct{})

type IManager interface {
	SetSizeQueue(size int64)
	GetSizeQueue() (size int64)
	//
	UseTimeoutForRunning(timeout time.Duration)
	RunDeferByTimer()
	//
	SetEventManageTasks(event EventManageTasks)
	GetEventManageTasks() (event EventManageTasks)
	//
	CreateTask(t Type, key string, send, update, fields interface{}, eventRunTask EventRunTask, eventUpdateState EventUpdateTaskState) (task ITask, err error)
	ModifyTaskAsTrigger(trigger ITask, dependents ...ITask) (task ITask, err error)
	//
	DeleteTasksByKeys(keys map[string]struct{})
	FindTasksByKeys(keys map[string]struct{}) (findTasks []ITask)
	FindTaskByKey(key string) (findTask ITask, err error)
	//
	RunTask(task ITask) (err error)
	UpdateTask(key string, updateContext interface{}) (err error)
	ManageCompletedTasks()
	RunDeferTasks()
	RunDependentTasks(task ITask)
}

