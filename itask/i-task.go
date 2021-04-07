package itask


type Type int
type EventRunTask func(task ITask) (doTaskAsDefer, sendToErrorChannel bool, err error)

type ITask interface {
	SetType(t Type)
	GetType() (t Type)
	//
	SetKey(key string)
	GetKey() (key string)
	//
	SetState(state IState)
	GetState() (state IState)
	//
	ModifyTaskAsTrigger(dependent []ITask)
	IsTrigger() (flag bool, dependent []ITask)
	//
	ModifyTaskAsDependent(trigger ITask)
	IsDependent() (flag bool, trigger ITask)
	//
	SetEventRunTask(event EventRunTask)
	GetEventRunTask() (event EventRunTask)
}
