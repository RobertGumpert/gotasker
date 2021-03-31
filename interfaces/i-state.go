package interfaces

type EventUpdateState func(task ITask, result interface{}) (err error)

type IState interface {
	SetRunnable(flag bool)
	IsRunnable() (flag bool)
	//
	SetExecute(flag bool)
	IsExecute() (flag bool)
	//
	SetDefer(flag bool)
	IsDefer() (flag bool)
	//
	SetResult(result interface{})
	GetResult() (result interface{})
	//
	SetCustomFields(fields interface{})
	GetCustomFields() (fields interface{})
	//
	SetError(err error)
	GetError() (err error)
	//
	SetEventUpdateState(event EventUpdateState)
	GetEventUpdateState() (event EventUpdateState)
}
