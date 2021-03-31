package tasker

import (
	"github.com/RobertGumpert/gotasker/interfaces"
)

type implementTask struct {
	t                      interfaces.Type
	ownKeyTask             string
	isTrigger, isDependent bool
	//
	event      interfaces.EventRunTask
	state      interfaces.IState
	trigger    interfaces.ITask
	dependents []interfaces.ITask
}

func (task *implementTask) SetEventRunTask(event interfaces.EventRunTask) {
	task.event = event
}

func (task *implementTask) GetEventRunTask() (event interfaces.EventRunTask) {
	return task.event
}

func (task *implementTask) SetState(state interfaces.IState) {
	task.state = state
}

func (task *implementTask) GetState() (state interfaces.IState) {
	return task.state
}

func (task *implementTask) SetType(t interfaces.Type) {
	task.t = t
}

func (task *implementTask) GetType() (t interfaces.Type) {
	return task.t
}

func (task *implementTask) SetKey(key string) {
	task.ownKeyTask = key
}

func (task *implementTask) GetKey() (key string) {
	return task.ownKeyTask
}

func (task *implementTask) DoAsTrigger(dependent []interfaces.ITask) {
	for t := 0; t < len(dependent); t++ {
		dependent[t].DoAsDependent(task)
	}
	task.isTrigger = true
}

func (task *implementTask) IsTrigger() (flag bool, dependent []interfaces.ITask) {
	return task.isTrigger, task.dependents
}

func (task *implementTask) DoAsDependent(trigger interfaces.ITask) {
	task.trigger = trigger
	task.isDependent = true
}

func (task *implementTask) IsDependent() (flag bool, trigger interfaces.ITask) {
	return task.isDependent, trigger
}
