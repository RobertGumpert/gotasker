package tasker

import "github.com/RobertGumpert/gotasker/itask"

type iError struct {
	err     error
	taskKey string
	task    itask.ITask
}

func (e *iError) SetTaskKey(taskKey string) {
	e.taskKey = taskKey
}

func (e *iError) SetTask(task itask.ITask) {
	e.task = task
}

func (e *iError) SetError(err error) {
	e.err = err
}

func (e *iError) GetTaskKey() (taskKey string) {
	return e.taskKey
}

func (e *iError) GetTaskIfExist() (task itask.ITask, exist bool) {
	if e.task == nil {
		return e.task, false
	}
	return e.task, true
}

func (e *iError) GetError() (err error) {
	return e.err
}

