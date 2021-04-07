package itask

type ITupleUpdate interface {
	SetError(err IError)
	SetTaskKey(key string)
	SetSomethingUpdateContext(somethingUpdateContext interface{})
	SetTask(task ITask)
	//
	GetError() (err IError)
	GetTAskKey() (taskKey string)
	GetSomethingUpdateContext() (somethingUpdateContext interface{})
	GetTask() (task ITask)
}
