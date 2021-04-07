package tasker

import "github.com/RobertGumpert/gotasker/itask"

type iTuple struct {
	taskKey                string
	task                   itask.ITask
	somethingUpdateContext interface{}
	err                    itask.IError
}

func (t iTuple) SetError(err itask.IError) {
	t.err = err
}

func (t iTuple) SetTaskKey(key string) {
	t.taskKey = key
}

func (t iTuple) SetSomethingUpdateContext(somethingUpdateContext interface{}) {
	t.somethingUpdateContext = somethingUpdateContext
}

func (t iTuple) SetTask(task itask.ITask) {
	t.task = task
}

func (t iTuple) GetError() (err itask.IError) {
	return t.err
}

func (t iTuple) GetTAskKey() (taskKey string) {
	return t.taskKey
}

func (t iTuple) GetSomethingUpdateContext() (somethingUpdateContext interface{}) {
	return t.somethingUpdateContext
}

func (t iTuple) GetTask() (task itask.ITask) {
	return t.task
}
