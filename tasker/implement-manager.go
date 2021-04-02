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

type taskerQueueManager struct {
	size                     int64
	timeout                  time.Duration
	tasksWithErrors          []itask.ITask
	completedTasks           []itask.ITask
	queueTasks               []itask.ITask
	channelForUpdateTasks    chan updateDataTuple
	channelForCompletedTasks chan itask.ITask
	mx                       *sync.Mutex
	eventManageTasks         itask.EventManageTasks
}

func newQueueManager(size int64, timeout time.Duration, eventManageTasks itask.EventManageTasks) *taskerQueueManager {
	manager := new(taskerQueueManager)
	manager.size = size
	manager.timeout = timeout
	manager.eventManageTasks = eventManageTasks
	manager.queueTasks = make([]itask.ITask, 0)
	manager.completedTasks = make([]itask.ITask, 0)
	manager.tasksWithErrors = make([]itask.ITask, 0)
	manager.channelForCompletedTasks = make(chan itask.ITask, size)
	manager.channelForUpdateTasks = make(chan updateDataTuple, size)
	manager.mx = new(sync.Mutex)
	go manager.ManageCompletedTasks()
	go manager.RunDeferByTimer()
	go manager.ManageUpdateTasks()
	return manager
}

func (manager *taskerQueueManager) UseTimeoutForRunning(timeout time.Duration) {
	manager.timeout = timeout
}

func (manager *taskerQueueManager) SetEventManageTasks(event itask.EventManageTasks) {
	manager.eventManageTasks = event
}

func (manager *taskerQueueManager) GetEventManageTasks() (event itask.EventManageTasks) {
	return manager.eventManageTasks
}

func (manager *taskerQueueManager) SetSizeQueue(size int64) {
	manager.size = size
}

func (manager *taskerQueueManager) GetSizeQueue() (size int64) {
	return manager.size
}

func (manager *taskerQueueManager) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) (task itask.ITask, err error) {
	if int64(len(manager.queueTasks)) > manager.size {
		return nil, gotasker.ErrorQueueIsFilled
	}
	state := &taskerState{
		send:                 send,
		update:               update,
		fields:               fields,
		eventUpdateTaskState: eventUpdateState,
	}
	return &taskerTask{
		t:            t,
		key:          key,
		eventRunTask: eventRunTask,
		state:        state,
	}, nil
}

func (manager *taskerQueueManager) ModifyTaskAsTrigger(trigger itask.ITask, dependents ...itask.ITask) (task itask.ITask, err error) {
	if int64(len(manager.queueTasks)+len(dependents)) > manager.size {
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

func (manager *taskerQueueManager) DeleteTasksByKeys(keys map[string]struct{}) {
	if len(keys) == 0 {
		return
	}
	slice := make([]itask.ITask, 0)
	for t := 0; t < len(manager.queueTasks); t++ {
		task := manager.queueTasks[t]
		if _, exist := keys[task.GetKey()]; exist == false {
			slice = append(slice, manager.queueTasks[t])
		}
	}
	manager.queueTasks = slice
	return
}

func (manager *taskerQueueManager) FindTasksByKeys(keys map[string]struct{}) (findTasks []itask.ITask) {
	findTasks = make([]itask.ITask, 0)
	if len(keys) == 0 {
		return findTasks
	}
	for t := 0; t < len(manager.queueTasks); t++ {
		task := manager.queueTasks[t]
		if _, ok := keys[task.GetKey()]; ok {
			findTasks = append(findTasks, task)
		}
	}
	return findTasks
}

func (manager *taskerQueueManager) FindTaskByKey(key string) (findTask itask.ITask, err error) {
	if key == "" {
		return nil, gotasker.ErrorTaskIsNotExist
	}
	for t := 0; t < len(manager.queueTasks); t++ {
		task := manager.queueTasks[t]
		if task.GetKey() == key {
			return task, nil
		}
	}
	return nil, gotasker.ErrorTaskIsNotExist
}

//func (manager *taskerQueueManager) UpdateTask(key string, updateContext interface{}) (err error) {
//	manager.mx.Lock()
//	defer manager.mx.Unlock()
//	log.Println("WRITE UPDATE ", key)
//	//
//	task, err := manager.FindTaskByKey(key)
//	if err != nil {
//		return err
//	}
//	eventUpdateTask := task.GetState().GetEventUpdateState()
//	if eventUpdateTask == nil {
//		return gotasker.ErrorTaskIsNotExist
//	}
//	err = eventUpdateTask(task, updateContext)
//	if err != nil {
//		return err
//	}
//	if task.GetState().IsExecute() {
//		log.Println("-> GOTASKER: TASK COMPLETED [", task.GetKey(), "] START WRITE TO CHANNEL...")
//		manager.channelForCompletedTasks <- task
//		log.Println("<- GOTASKER: TASK COMPLETED [", task.GetKey(), "] FINISH WRITE TO CHANNEL...")
//	}
//	return nil
//}

func (manager *taskerQueueManager) UpdateTask(key string, updateContext interface{}) (err error) {
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
	manager.channelForUpdateTasks <- updateDataTuple{
		Task:          task,
		UpdateContext: updateContext,
	}
	return nil
}

func (manager *taskerQueueManager) ManageUpdateTasks() {
	for context := range manager.channelForUpdateTasks {
		if isDependent, trigger := context.Task.IsDependent(); isDependent {
			if !trigger.GetState().IsExecute() {
				log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetError(gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetUpdateContext(context.UpdateContext)
				manager.tasksWithErrors = append(manager.tasksWithErrors, context.Task)
				continue
			}
			if trigger.GetState().GetError() != nil {
				log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetError(gotasker.ErrorDependentTaskRunBeforeTrigger)
				context.Task.GetState().SetUpdateContext(context.UpdateContext)
				manager.tasksWithErrors = append(manager.tasksWithErrors, context.Task)
				continue
			}
		}
		eventUpdateTask := context.Task.GetState().GetEventUpdateState()
		err := eventUpdateTask(context.Task, context.UpdateContext)
		if err != nil {
			log.Println("-> GOTASKER: TASK UPDATE WITH ERROR [", context.Task.GetKey(), "]: ", err)
			context.Task.GetState().SetError(err)
			context.Task.GetState().SetUpdateContext(context.UpdateContext)
			manager.tasksWithErrors = append(manager.tasksWithErrors, context.Task)
			continue
		}
		if context.Task.GetState().IsExecute() {
			log.Println("-> GOTASKER: TASK COMPLETED [", context.Task.GetKey(), "] START WRITE TO CHANNEL...")
			manager.channelForCompletedTasks <- context.Task
			log.Println("<- GOTASKER: TASK COMPLETED [", context.Task.GetKey(), "] FINISH WRITE TO CHANNEL...")
		}
	}
}

func (manager *taskerQueueManager) ManageCompletedTasks() {
	for task := range manager.channelForCompletedTasks {
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
			manager.completedTasks = append(manager.completedTasks, tasksForSaving...)
		}
		if len(tasksKeysForDelete) != 0 {
			log.Println(tasksKeysForDelete)
			manager.DeleteTasksByKeys(tasksKeysForDelete)
		}
		if len(manager.queueTasks) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks()
		}
	}
}

func (manager *taskerQueueManager) RunTask(task itask.ITask) (err error) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	isTrigger, dependents := task.IsTrigger()
	if !isTrigger {
		if int64(len(manager.queueTasks)) > manager.size {
			return gotasker.ErrorQueueIsFilled
		}
	} else {
		if int64(len(manager.queueTasks)+len(dependents)+1) > manager.size {
			return gotasker.ErrorQueueIsFilled
		}
		manager.queueTasks = append(manager.queueTasks, dependents...)
	}
	log.Println("-> GOTASKER: START ADD TASK [", task.GetKey(), "]-----------------------------------------------------")
	manager.queueTasks = append(manager.queueTasks, task)
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

func (manager *taskerQueueManager) RunDeferTasks() {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	for t := 0; t < len(manager.queueTasks); t++ {
		if manager.queueTasks[t].GetState().IsExecute() {
			continue
		}
		if isDependent, trigger := manager.queueTasks[t].IsDependent(); isDependent == true {
			if trigger.GetState().IsExecute() == false {
				continue
			}
		}
		if manager.queueTasks[t].GetState().IsRunnable() == true {
			continue
		}
		if manager.queueTasks[t].GetState().IsDefer() == false {
			continue
		}
		if err := manager.queueTasks[t].GetEventRunTask()(manager.queueTasks[t]); err == nil {
			manager.queueTasks[t].GetState().SetRunnable(true)
			log.Println("\t\t\t->GOTASKER: DEFER TASK KEY [", manager.queueTasks[t].GetKey(), "] IS RUN")
		}
	}
}

func (manager *taskerQueueManager) RunDependentTasks(task itask.ITask) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	//
	if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
		if task.GetState().IsExecute() == false {
			return
		}
		for d := 0; d < len(dependentsTasks); d++ {
			if isDependent, _ := dependentsTasks[d].IsDependent(); isDependent {
				if dependentsTasks[d].GetState().IsExecute() {
					continue
				}
				if dependentsTasks[d].GetState().IsRunnable() {
					continue
				}
				if err := dependentsTasks[d].GetEventRunTask()(dependentsTasks[d]); err == nil {
					dependentsTasks[d].GetState().SetRunnable(true)
					log.Println("\t\t\t->GOTASKER: DEPENDENT TASK KEY [", dependentsTasks[d].GetKey(), "] IS RUN")
				}
			}
		}
	}
	return
}

func (manager *taskerQueueManager) RunDeferByTimer() {
	for {
		if len(manager.queueTasks) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks()
		}
		time.Sleep(manager.timeout)
	}
}
