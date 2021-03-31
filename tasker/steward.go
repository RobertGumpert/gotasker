package tasker

import "github.com/RobertGumpert/gotasker/interfaces"

type Steward struct {
	manager *implementManager
}

func NewSteward(size int64, event interfaces.EventManageTasks) *Steward {
	return &Steward{
		manager: newImplementManager(
			size,
			event,
		),
	}
}

func (steward *Steward) CreateTask(t interfaces.Type, key string, result, fields interface{}, runTask interfaces.EventRunTask, updateTask interfaces.EventUpdateState) (task interfaces.ITask, err error) {
	return steward.manager.CreateTask(
		t,
		key,
		result,
		fields,
		runTask,
		updateTask,
	)
}

func (steward *Steward) UpdateTaskStateExecute(key string, result interface{}) (err error) {
	return steward.manager.UpdateTaskStateExecute(
		key,
		result,
	)
}

func (steward *Steward) ModifyTaskAsTrigger(trigger interfaces.ITask, dependents ...interfaces.ITask) (task interfaces.ITask, err error) {
	return steward.manager.ModifyTaskAsTrigger(
		trigger,
		dependents...,
	)
}

func (steward *Steward) ServeTask(task interfaces.ITask) (err error) {
	return steward.manager.Add(
		task,
	)
}
