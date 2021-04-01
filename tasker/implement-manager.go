package tasker

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"log"
	"sync"
	"time"
)

type TaskerQueueManager struct {
	size                     int64
	timeout                  time.Duration
	completedTasks           []itask.ITask
	queueTasks               []itask.ITask
	channelForCompletedTasks chan itask.ITask
	mx                       *sync.Mutex
	eventManageTasks         itask.EventManageTasks
}

func newQueueManager(size int64, timeout time.Duration, eventManageTasks itask.EventManageTasks) *TaskerQueueManager {
	manager := new(TaskerQueueManager)
	manager.size = size
	manager.timeout = timeout
	manager.eventManageTasks = eventManageTasks
	manager.queueTasks = make([]itask.ITask, 0)
	manager.completedTasks = make([]itask.ITask, 0)
	manager.channelForCompletedTasks = make(chan itask.ITask)
	manager.mx = new(sync.Mutex)
	go manager.ManageCompletedTasks()
	go manager.RunDeferByTimer()
	return manager
}

func (manager *TaskerQueueManager) UseTimeoutForRunning(timeout time.Duration) {
	manager.timeout = timeout
}

func (manager *TaskerQueueManager) SetEventManageTasks(event itask.EventManageTasks) {
	manager.eventManageTasks = event
}

func (manager *TaskerQueueManager) GetEventManageTasks() (event itask.EventManageTasks) {
	return manager.eventManageTasks
}

func (manager *TaskerQueueManager) SetSizeQueue(size int64) {
	manager.size = size
}

func (manager *TaskerQueueManager) GetSizeQueue() (size int64) {
	return manager.size
}

func (manager *TaskerQueueManager) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) (task itask.ITask, err error) {
	if int64(len(manager.queueTasks)) == manager.size {
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

func (manager *TaskerQueueManager) ModifyTaskAsTrigger(trigger itask.ITask, dependents ...itask.ITask) (task itask.ITask, err error) {
	if int64(len(manager.queueTasks)+len(dependents)) == manager.size {
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

func (manager *TaskerQueueManager) DeleteTasksByKeys(keys map[string]struct{}) {
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

func (manager *TaskerQueueManager) FindTasksByKeys(keys map[string]struct{}) (findTasks []itask.ITask) {
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

func (manager *TaskerQueueManager) FindTaskByKey(key string) (findTask itask.ITask, err error) {
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

func (manager *TaskerQueueManager) UpdateTask(key string, updateContext interface{}) (err error) {
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
	err = eventUpdateTask(task, updateContext)
	if err != nil {
		return err
	}
	if task.GetState().IsExecute() {
		log.Println("-> GOTASKER: TASK COMPLETED [", task.GetKey(), "] START WRITE TO CHANNEL...")
		manager.channelForCompletedTasks <- task
		log.Println("<- GOTASKER: TASK COMPLETED [", task.GetKey(), "] FINISH WRITE TO CHANNEL...")
	}
	return nil
}

func (manager *TaskerQueueManager) ManageCompletedTasks() {
	for task := range manager.channelForCompletedTasks {
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

func (manager *TaskerQueueManager) RunTask(task itask.ITask) (err error) {
	flag, dependents := task.IsTrigger()
	if !flag {
		if int64(len(manager.queueTasks)) == manager.size {
			return gotasker.ErrorQueueIsFilled
		}
	} else {
		if int64(len(manager.queueTasks)+len(dependents)) == manager.size {
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

func (manager *TaskerQueueManager) RunDeferTasks() {
	for t := 0; t < len(manager.queueTasks); t++ {
		if manager.queueTasks[t].GetState().IsExecute() {
			continue
		}
		if isDependent, trigger := manager.queueTasks[t].IsDependent(); isDependent == true {
			if trigger == nil {
				log.Println()
				continue
			}
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

func (manager *TaskerQueueManager) RunDependentTasks(task itask.ITask) {
	if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
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

func (manager *TaskerQueueManager) RunDeferByTimer() {
	for {
		if len(manager.queueTasks) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks()
		}
		time.Sleep(manager.timeout)
	}
}
