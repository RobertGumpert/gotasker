package interfaces

type Type int
type EventRunTask func(task ITask) (err error)

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
	DoAsTrigger(dependent []ITask)
	IsTrigger() (flag bool, dependent []ITask)
	//
	DoAsDependent(trigger ITask)
	IsDependent() (flag bool, trigger ITask)
	//
	SetEventRunTask(event EventRunTask)
	GetEventRunTask() (event EventRunTask)
}
