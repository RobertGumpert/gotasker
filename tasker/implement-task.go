package tasker

import "github.com/RobertGumpert/gotasker/itask"

type taskerTask struct {
	t                      itask.Type
	key                    string
	isTrigger, isDependent bool
	//
	eventRunTask itask.EventRunTask
	state        itask.IState
	trigger      itask.ITask
	dependents   []itask.ITask
}

func (task *taskerTask) SetEventRunTask(event itask.EventRunTask) {
	task.eventRunTask = event
}

func (task *taskerTask) GetEventRunTask() (event itask.EventRunTask) {
	return task.eventRunTask
}

func (task *taskerTask) SetState(state itask.IState) {
	task.state = state
}

func (task *taskerTask) GetState() (state itask.IState) {
	return task.state
}

func (task *taskerTask) SetType(t itask.Type) {
	task.t = t
}

func (task *taskerTask) GetType() (t itask.Type) {
	return task.t
}

func (task *taskerTask) SetKey(key string) {
	task.key = key
}

func (task *taskerTask) GetKey() (key string) {
	return task.key
}

func (task *taskerTask) ModifyTaskAsTrigger(dependent []itask.ITask) {
	for t := 0; t < len(dependent); t++ {
		dependent[t].ModifyTaskAsDependent(task)
	}
	task.isTrigger = true
	task.dependents = dependent
}

func (task *taskerTask) IsTrigger() (flag bool, dependent []itask.ITask) {
	return task.isTrigger, task.dependents
}

func (task *taskerTask) ModifyTaskAsDependent(trigger itask.ITask) {
	task.isDependent = true
	task.trigger = trigger
}

func (task *taskerTask) IsDependent() (flag bool, trigger itask.ITask) {
	return task.isDependent, task.trigger
}
