package itask

type EventUpdateTaskState func(task ITask, interProgramUpdateContext interface{}) (err error)

type IState interface {
	SetRunnable(flag bool)
	IsRunnable() (flag bool)
	//
	SetCompleted(flag bool)
	IsCompleted() (flag bool)
	//
	SetDefer(flag bool)
	IsDefer() (flag bool)
	//
	SetSendContext(send interface{})
	GetSendContext() (send interface{})
	//
	SetUpdateContext(update interface{})
	GetUpdateContext() (update interface{})
	//
	SetCustomFields(fields interface{})
	GetCustomFields() (fields interface{})
	//
	SetError(err error)
	GetError() (err error)
	//
	SetEventUpdateState(event EventUpdateTaskState)
	GetEventUpdateState() (event EventUpdateTaskState)
}
