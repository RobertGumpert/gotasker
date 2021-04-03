package tasker

import "github.com/RobertGumpert/gotasker/itask"

type implementTaskInterface struct {
	t                      itask.Type
	key                    string
	isTrigger, isDependent bool
	eventRunTask itask.EventRunTask
	state        itask.IState
	trigger      itask.ITask
	dependents   []itask.ITask
}

func (task *implementTaskInterface) SetEventRunTask(event itask.EventRunTask) {
	task.eventRunTask = event
}

func (task *implementTaskInterface) GetEventRunTask() (event itask.EventRunTask) {
	return task.eventRunTask
}

func (task *implementTaskInterface) SetState(state itask.IState) {
	task.state = state
}

func (task *implementTaskInterface) GetState() (state itask.IState) {
	return task.state
}

func (task *implementTaskInterface) SetType(t itask.Type) {
	task.t = t
}

func (task *implementTaskInterface) GetType() (t itask.Type) {
	return task.t
}

func (task *implementTaskInterface) SetKey(key string) {
	task.key = key
}

func (task *implementTaskInterface) GetKey() (key string) {
	return task.key
}

func (task *implementTaskInterface) ModifyTaskAsTrigger(dependent []itask.ITask) {
	for t := 0; t < len(dependent); t++ {
		dependent[t].ModifyTaskAsDependent(task)
	}
	task.isTrigger = true
	task.dependents = dependent
}

func (task *implementTaskInterface) IsTrigger() (flag bool, dependent []itask.ITask) {
	return task.isTrigger, task.dependents
}

func (task *implementTaskInterface) ModifyTaskAsDependent(trigger itask.ITask) {
	task.isDependent = true
	task.trigger = trigger
}

func (task *implementTaskInterface) IsDependent() (flag bool, trigger itask.ITask) {
	return task.isDependent, task.trigger
}
