package tasker

import (
	"github.com/RobertGumpert/gotasker/itask"
	"time"
)

type TaskConstructor func() (task itask.ITask, err error)

type TaskerQueueSteward struct {
	manager *TaskerQueueManager
}

func NewSteward(size int64, timeout time.Duration, eventManageTasks itask.EventManageTasks) *TaskerQueueSteward {
	steward := new(TaskerQueueSteward)
	steward.manager = newQueueManager(
		size,
		timeout,
		eventManageTasks,
	)
	return steward
}

func (steward *TaskerQueueSteward) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) TaskConstructor {
	return func() (task itask.ITask, err error) {
		return steward.manager.CreateTask(
			t,
			key,
			send,
			update,
			fields,
			eventRunTask,
			eventUpdateState,
		)
	}
}

func (steward *TaskerQueueSteward) CreateTaskAndRun(constructor TaskConstructor) (task itask.ITask, err error) {
	task, err = constructor()
	if err != nil {
		return nil, err
	}
	err = steward.manager.RunTask(task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (steward *TaskerQueueSteward) CreateTriggerAndRun(triggerConstructor TaskConstructor, dependentConstructors ...TaskConstructor) (trigger itask.ITask, err error) {
	var (
		dependents = make([]itask.ITask, 0)
	)
	trigger, err = triggerConstructor()
	if err != nil {
		return nil, err
	}
	for _, dependentConstructor := range dependentConstructors {
		dependent, err := dependentConstructor()
		if err != nil {
			return nil, err
		}
		dependents = append(dependents, dependent)
	}
	trigger, err = steward.manager.ModifyTaskAsTrigger(
		trigger,
		dependents...,
	)
	if err != nil {
		return nil, err
	}
	err = steward.manager.RunTask(trigger)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (steward *TaskerQueueSteward) UpdateTask(key string, updateContext interface{}) (err error) {
	return steward.manager.UpdateTask(
		key,
		updateContext,
	)
}
