package tasker

import (
	"github.com/RobertGumpert/gotasker/itask"
	"time"
)

type Steward struct {
	manager *implementTaskManager
}

func NewSteward(sizeOfQueue int64, timeoutForRunWithTimer time.Duration, eventManageTasks itask.EventManageTasks) *Steward {
	steward := new(Steward)
	steward.manager = newQueueManager(
		sizeOfQueue,
		timeoutForRunWithTimer,
		eventManageTasks,
	)
	return steward
}

func (steward *Steward) CanAddTask(countTasks int64) (havePlaceInQueue bool) {
	if (int64(len(steward.manager.sliceTasksInQueue)) + countTasks) < steward.manager.GetSizeQueue() {
		return true
	} else {
		return false
	}
}

func (steward *Steward) CreateTask(taskType itask.Type, taskKey string, taskSendContext, taskUpdateContext, taskCustomFieldContext interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) itask.TaskConstructor {
	return func() (task itask.ITask, err error) {
		return steward.manager.CreateTask(
			taskType,
			taskKey,
			taskSendContext,
			taskUpdateContext,
			taskCustomFieldContext,
			eventRunTask,
			eventUpdateState,
		)
	}
}

func (steward *Steward) RunTask(task itask.ITask) (err error) {
	err = steward.manager.RunTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (steward *Steward) CreateTaskAndRun(constructor itask.TaskConstructor) (task itask.ITask, err error) {
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

func (steward *Steward) ModifyTaskAsTrigger(triggerConstructor itask.TaskConstructor, dependentConstructors ...itask.TaskConstructor) (trigger itask.ITask, err error) {
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
	return trigger, nil
}

func (steward *Steward) CreateTriggerAndRun(triggerConstructor itask.TaskConstructor, dependentConstructors ...itask.TaskConstructor) (trigger itask.ITask, err error) {
	trigger, err = steward.ModifyTaskAsTrigger(
		triggerConstructor,
		dependentConstructors...,
	)
	err = steward.manager.RunTask(trigger)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (steward *Steward) UpdateTask(key string, updateContext interface{}) (err error) {
	return steward.manager.UpdateTask(
		key,
		updateContext,
	)
}

func (steward *Steward) FindTasksByKeys(keys map[string]struct{}) (findTasks []itask.ITask) {
	return steward.manager.FindTasksByKeys(keys)
}

func (steward *Steward) FindTaskByKey(key string) (findTask itask.ITask, err error) {
	return steward.manager.FindTaskByKey(key)
}

func (steward *Steward) GetTasksSavingOrCompleted() (tasks []itask.ITask) {
	return steward.manager.GetTasksSavingOrCompleted()
}

func (steward *Steward) GetTasksInQueue() (tasks []itask.ITask) {
	return steward.manager.GetTasksInQueue()
}

func (steward *Steward) GetTasksCompletedWithError() (tasks []itask.ITask) {
	return steward.manager.GetTasksCompletedWithError()
}
