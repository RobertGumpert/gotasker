package tasker

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"log"
	"sync"
	"time"
)

type updateDataTuple struct {
	Task          itask.ITask
	UpdateContext interface{}
}

type implementTaskManager struct {
	sizeOfQueue                    int64
	timeoutForRunWithTimer         time.Duration
	sliceTasksCompletedWithError   []itask.ITask
	sliceTasksCompleted            []itask.ITask
	sliceTasksInQueue              []itask.ITask
	channelForManageUpdateTasks    chan updateDataTuple
	channelForManageCompletedTasks chan itask.ITask
	mx                             *sync.Mutex
	eventManageTasks               itask.EventManageTasks
}

func newQueueManager(sizeOfQueue int64, timeoutForRunWithTimer time.Duration, eventManageTasks itask.EventManageTasks) *implementTaskManager {
	manager := new(implementTaskManager)
	manager.sizeOfQueue = sizeOfQueue
	manager.timeoutForRunWithTimer = timeoutForRunWithTimer
	manager.eventManageTasks = eventManageTasks
	manager.sliceTasksInQueue = make([]itask.ITask, 0)
	manager.sliceTasksCompleted = make([]itask.ITask, 0)
	manager.sliceTasksCompletedWithError = make([]itask.ITask, 0)
	manager.channelForManageCompletedTasks = make(chan itask.ITask, sizeOfQueue)
	manager.channelForManageUpdateTasks = make(chan updateDataTuple, sizeOfQueue)
	manager.mx = new(sync.Mutex)
	go manager.ManageCompletedTasks()
	go manager.RunDeferByTimer()
	go manager.ManageUpdateTasks()
	return manager
}

//
//
//
//---IMPLEMENT PROPERTIES-----------------------------------------------------------------------------------------------
//
//
//

func (manager *implementTaskManager) SetSizeQueue(size int64) {
	manager.sizeOfQueue = size
}

func (manager *implementTaskManager) GetSizeQueue() (size int64) {
	return manager.sizeOfQueue
}

func (manager *implementTaskManager) UseTimeoutForRunning(timeout time.Duration) {
	manager.timeoutForRunWithTimer = timeout
}

func (manager *implementTaskManager) SetEventManageTasks(event itask.EventManageTasks) {
	manager.eventManageTasks = event
}

func (manager *implementTaskManager) GetEventManageTasks() (event itask.EventManageTasks) {
	return manager.eventManageTasks
}

//
//
//
//---IMPLEMENT CREATE TASK----------------------------------------------------------------------------------------------
//
//
//

func (manager *implementTaskManager) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) (task itask.ITask, err error) {
	if int64(len(manager.sliceTasksInQueue)) > manager.sizeOfQueue {
		return nil, gotasker.ErrorQueueIsFilled
	}
	state := &implementStateInterface{
		sendContext:          send,
		updateContext:        update,
		customFieldsContext:  fields,
		eventUpdateTaskState: eventUpdateState,
	}
	return &implementTaskInterface{
		t:            t,
		key:          key,
		eventRunTask: eventRunTask,
		state:        state,
	}, nil
}

func (manager *implementTaskManager) ModifyTaskAsTrigger(trigger itask.ITask, dependents ...itask.ITask) (task itask.ITask, err error) {
	if int64(len(manager.sliceTasksInQueue)+len(dependents)) > manager.sizeOfQueue {
		return nil, gotasker.ErrorQueueIsFilled
	}
	for t := 0; t < len(dependents); t++ {
		dependents[t].ModifyTaskAsDependent(trigger)
		dependents[t].GetState().SetDefer(true)
		dependents[t].GetState().SetRunnable(false)
	}
	trigger.ModifyTaskAsTrigger(dependents)
	return trigger, nil
}

//
//
//
//---IMPLEMENT MANAGE TASKS---------------------------------------------------------------------------------------------
//
//
//

func (manager *implementTaskManager) UpdateTask(key string, updateContext interface{}) (err error) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	task, err := manager.FindTaskByKey(key)
	if err != nil {
		return err
	}
	eventUpdateTask := task.GetState().GetEventUpdateState()
	if eventUpdateTask == nil {
		return gotasker.ErrorTaskIsNotExist
	}
	manager.channelForManageUpdateTasks <- updateDataTuple{
		Task:          task,
		UpdateContext: updateContext,
	}
	return nil
}

func (manager *implementTaskManager) DeleteTasksByKeys(keys map[string]struct{}) {
	if len(keys) == 0 {
		return
	}
	slice := make([]itask.ITask, 0)
	for t := 0; t < len(manager.sliceTasksInQueue); t++ {
		task := manager.sliceTasksInQueue[t]
		if _, exist := keys[task.GetKey()]; exist == false {
			slice = append(slice, manager.sliceTasksInQueue[t])
		}
	}
	manager.sliceTasksInQueue = slice
	return
}

func (manager *implementTaskManager) FindTasksByKeys(keys map[string]struct{}) (findTasks []itask.ITask) {
	findTasks = make([]itask.ITask, 0)
	if len(keys) == 0 {
		return findTasks
	}
	for t := 0; t < len(manager.sliceTasksInQueue); t++ {
		task := manager.sliceTasksInQueue[t]
		if _, ok := keys[task.GetKey()]; ok {
			findTasks = append(findTasks, task)
		}
	}
	return findTasks
}

func (manager *implementTaskManager) FindTaskByKey(key string) (findTask itask.ITask, err error) {
	if key == "" {
		return nil, gotasker.ErrorTaskIsNotExist
	}
	for t := 0; t < len(manager.sliceTasksInQueue); t++ {
		task := manager.sliceTasksInQueue[t]
		if task.GetKey() == key {
			return task, nil
		}
	}
	return nil, gotasker.ErrorTaskIsNotExist
}

func (manager *implementTaskManager) GetTasksSavingOrCompleted() (tasks []itask.ITask) {
	return manager.sliceTasksCompleted
}

func (manager *implementTaskManager) GetTasksInQueue() (tasks []itask.ITask) {
	return manager.sliceTasksInQueue
}

func (manager *implementTaskManager) GetTasksCompletedWithError() (tasks []itask.ITask) {
	return manager.sliceTasksCompletedWithError
}

//
//
//
//---IMPLEMENT RUN TASKS------------------------------------------------------------------------------------------------
//
//
//

func (manager *implementTaskManager) RunTask(task itask.ITask) (err error) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	isTrigger, dependents := task.IsTrigger()
	if !isTrigger {
		if int64(len(manager.sliceTasksInQueue)) > manager.sizeOfQueue {
			return gotasker.ErrorQueueIsFilled
		}
	} else {
		if int64(len(manager.sliceTasksInQueue)+len(dependents)+1) > manager.sizeOfQueue {
			return gotasker.ErrorQueueIsFilled
		}
		manager.sliceTasksInQueue = append(manager.sliceTasksInQueue, dependents...)
	}
	log.Println("-> GOTASKER: START ADD TASK [", task.GetKey(), "]-----------------------------------------------------")
	manager.sliceTasksInQueue = append(manager.sliceTasksInQueue, task)
	errRunning := task.GetEventRunTask()(task)
	if errRunning == nil {
		log.Println("\t\t\t:GOTASKER: TASK IS RUN [", task.GetKey(), "]")
		task.GetState().SetRunnable(true)
		task.GetState().SetDefer(false)
	} else {
		log.Println("\t\t\t:GOTASKER: TASK IS DEFER [", task.GetKey(), "]")
		task.GetState().SetRunnable(false)
		task.GetState().SetDefer(true)
	}
	log.Println("<- GOTASKER: FINISH ADD TASK [", task.GetKey(), "]----------------------------------------------------")
	return nil
}

func (manager *implementTaskManager) RunDeferTasks() {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	for t := 0; t < len(manager.sliceTasksInQueue); t++ {
		log.Println("\t\t\t-юGOTASKER: DEFER TASK ATTEMPT TO START [", manager.sliceTasksInQueue[t].GetKey(), "] ")
		if manager.sliceTasksInQueue[t].GetState().IsCompleted() {
			log.Println("\t\t\t\t\t\t->DEFER [", manager.sliceTasksInQueue[t].GetKey(), "] IS COMPLETED")
			continue
		}
		if isDependent, trigger := manager.sliceTasksInQueue[t].IsDependent(); isDependent == true {
			if trigger.GetState().IsCompleted() == false {
				continue
			}
		}
		if manager.sliceTasksInQueue[t].GetState().IsRunnable() == true {
			log.Println("\t\t\t\t\t\t->DEFER [", manager.sliceTasksInQueue[t].GetKey(), "] IS RUNNABLE")
			continue
		}
		if manager.sliceTasksInQueue[t].GetState().IsDefer() == false {
			log.Println("\t\t\t\t\t\t->DEFER [", manager.sliceTasksInQueue[t].GetKey(), "] ISN'T DEFER")
			continue
		}
		if err := manager.sliceTasksInQueue[t].GetEventRunTask()(manager.sliceTasksInQueue[t]); err == nil {
			manager.sliceTasksInQueue[t].GetState().SetRunnable(true)
			log.Println("\t\t\t->GOTASKER: DEFER TASK IS RUN [", manager.sliceTasksInQueue[t].GetKey(), "]")
		} else {
			log.Println("\t\t\t->GOTASKER: DEFER TASK ISN'T RUN [", manager.sliceTasksInQueue[t].GetKey(), "]")
		}
	}
}

func (manager *implementTaskManager) RunDependentTasks(task itask.ITask) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
		if task.GetState().IsCompleted() == false {
			return
		}
		for d := 0; d < len(dependentsTasks); d++ {
			if isDependent, _ := dependentsTasks[d].IsDependent(); isDependent {
				log.Println("\t\t\t->GOTASKER: DEPENDENT TASK KEY ATTEMPT TO START [", dependentsTasks[d].GetKey(), "]")
				if dependentsTasks[d].GetState().IsCompleted() {
					log.Println("\t\t\t\t\t\t->DEPENDENT [", dependentsTasks[d].GetKey(), "] ISN'T COMPLETED")
					continue
				}
				if dependentsTasks[d].GetState().IsRunnable() {
					log.Println("\t\t\t\t\t\t->DEPENDENT [", dependentsTasks[d].GetKey(), "] ISN'T RUNNABLE")
					continue
				}
				if err := dependentsTasks[d].GetEventRunTask()(dependentsTasks[d]); err == nil {
					dependentsTasks[d].GetState().SetRunnable(true)
					log.Println("\t\t\t->GOTASKER: DEPENDENT TASK IS RUN [", dependentsTasks[d].GetKey(), "] ")
				} else {
					log.Println("\t\t\t->GOTASKER: DEPENDENT TASK ISN'T RUN [", dependentsTasks[d].GetKey(), "]")
				}
			}
		}
	}
	return
}

func (manager *implementTaskManager) RunDeferByTimer() {
	for {
		if len(manager.sliceTasksInQueue) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks()
		}
		time.Sleep(manager.timeoutForRunWithTimer)
	}
}

//
//
//
//---IMPLEMENT MANAGE QUEUES--------------------------------------------------------------------------------------------
//
//
//

func (manager *implementTaskManager) ManageUpdateTasks() {
	for context := range manager.channelForManageUpdateTasks {
		if context.Task.GetState().GetError() != nil {
			continue
		}
		if isDependent, trigger := context.Task.IsDependent(); isDependent {
			if !trigger.GetState().IsCompleted() {
				log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetError(gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetUpdateContext(context.UpdateContext)
				manager.sliceTasksCompletedWithError = append(manager.sliceTasksCompletedWithError, context.Task)
				continue
			}
			if trigger.GetState().GetError() != nil {
				log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetError(gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetUpdateContext(context.UpdateContext)
				manager.sliceTasksCompletedWithError = append(manager.sliceTasksCompletedWithError, context.Task)
				continue
			}
		}
		eventUpdateTask := context.Task.GetState().GetEventUpdateState()
		err := eventUpdateTask(context.Task, context.UpdateContext)
		if err != nil {
			log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", err)
			context.Task.GetState().SetError(err)
			context.Task.GetState().SetUpdateContext(context.UpdateContext)
			manager.sliceTasksCompletedWithError = append(manager.sliceTasksCompletedWithError, context.Task)
			continue
		}
		if context.Task.GetState().IsCompleted() {
			log.Println("-> GOTASKER: TASK COMPLETED [", context.Task.GetKey(), "] START WRITE TO CHANNEL...")
			manager.channelForManageCompletedTasks <- context.Task
			log.Println("<- GOTASKER: TASK COMPLETED [", context.Task.GetKey(), "] FINISH WRITE TO CHANNEL...")
		}
	}
}

func (manager *implementTaskManager) ManageCompletedTasks() {
	for task := range manager.channelForManageCompletedTasks {
		_, err := manager.FindTaskByKey(task.GetKey())
		if err != nil {
			continue
		}
		tasksKeysForDelete, tasksKeysForSaving := manager.eventManageTasks(task)
		if isTrigger, _ := task.IsTrigger(); isTrigger {
			manager.RunDependentTasks(task)
		}
		if len(tasksKeysForSaving) != 0 {
			tasksForSaving := manager.FindTasksByKeys(tasksKeysForSaving)
			manager.sliceTasksCompleted = append(manager.sliceTasksCompleted, tasksForSaving...)
		}
		if len(tasksKeysForDelete) != 0 {
			log.Println(tasksKeysForDelete)
			manager.DeleteTasksByKeys(tasksKeysForDelete)
		}
		if len(manager.sliceTasksInQueue) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks()
		}
	}
}

