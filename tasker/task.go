package tasker

import "github.com/RobertGumpert/gotasker/itask"

type iTask struct {
	t                      itask.Type
	key                    string
	isTrigger, isDependent bool
	eventRunTask itask.EventRunTask
	state        itask.IState
	trigger      itask.ITask
	dependents   []itask.ITask
}

func (tsk *iTask) SetEventRunTask(event itask.EventRunTask) {
	tsk.eventRunTask = event
}

func (tsk *iTask) GetEventRunTask() (event itask.EventRunTask) {
	return tsk.eventRunTask
}

func (tsk *iTask) SetState(state itask.IState) {
	tsk.state = state
}

func (tsk *iTask) GetState() (state itask.IState) {
	return tsk.state
}

func (tsk *iTask) SetType(t itask.Type) {
	tsk.t = t
}

func (tsk *iTask) GetType() (t itask.Type) {
	return tsk.t
}

func (tsk *iTask) SetKey(key string) {
	tsk.key = key
}

func (tsk *iTask) GetKey() (key string) {
	return tsk.key
}

func (tsk *iTask) ModifyTaskAsTrigger(dependent []itask.ITask) {
	for t := 0; t < len(dependent); t++ {
		dependent[t].ModifyTaskAsDependent(tsk)
	}
	tsk.isTrigger = true
	tsk.dependents = dependent
}

func (tsk *iTask) IsTrigger() (flag bool, dependent []itask.ITask) {
	return tsk.isTrigger, tsk.dependents
}

func (tsk *iTask) ModifyTaskAsDependent(trigger itask.ITask) {
	tsk.isDependent = true
	tsk.trigger = trigger
}

func (tsk *iTask) IsDependent() (flag bool, trigger itask.ITask) {
	return tsk.isDependent, tsk.trigger
}
