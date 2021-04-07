package tasker

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/itask"
	"log"
	"sync"
	"time"
)

type Option func(manager *iManager)

type iManager struct {
	sizeOfQueue                    int64
	timeOut                        time.Duration
	sliceTask                      []itask.ITask
	channelForManageCompletedTasks chan itask.ITask
	channelForSendErrors           chan itask.IError
	channelForManageUpdateTasks    chan itask.ITupleUpdate
	eventManageTasks               itask.EventManageTasks
	mx                             *sync.Mutex
}

func SetBaseOptions(sizeOfQueue int64, eventManageTasks itask.EventManageTasks) Option {
	return func(manager *iManager) {
		manager.sizeOfQueue = sizeOfQueue
		manager.eventManageTasks = eventManageTasks
		manager.sliceTask = make([]itask.ITask, 0)
		manager.channelForManageCompletedTasks = make(chan itask.ITask, sizeOfQueue)
		manager.channelForSendErrors = make(chan itask.IError, sizeOfQueue)
		manager.channelForManageUpdateTasks = make(chan itask.ITupleUpdate, sizeOfQueue)
		manager.mx = new(sync.Mutex)
		go manager.ManageUpdates()
		go manager.ManageCompleted()
	}
}

func SetRunByTimer(timeOut time.Duration) Option {
	return func(manager *iManager) {
		manager.timeOut = timeOut
		go manager.RunDeferByTimer()
	}
}

func NewManager(options ...Option) itask.IManager {
	manager := new(iManager)
	for _, option := range options {
		option(manager)
	}
	return manager
}

//
//
//
//---IMPLEMENT CREATE TASK----------------------------------------------------------------------------------------------
//
//
//

func (manager *iManager) CreateTask(t itask.Type, key string, send, update, fields interface{}, eventRunTask itask.EventRunTask, eventUpdateState itask.EventUpdateTaskState) (task itask.ITask, err error) {
	isFilled := manager.QueueIsFilled(1)
	if isFilled {
		return nil, gotasker.ErrorQueueIsFilled
	}
	state := &iState{
		sendContext:          send,
		updateContext:        update,
		customFieldsContext:  fields,
		eventUpdateTaskState: eventUpdateState,
	}
	return &iTask{
		t:            t,
		key:          key,
		eventRunTask: eventRunTask,
		state:        state,
	}, nil
}

func (manager *iManager) ModifyTaskAsTrigger(trigger itask.ITask, dependents ...itask.ITask) (task itask.ITask, err error) {
	isFilled := manager.QueueIsFilled(int64(len(dependents)) + 1)
	if isFilled {
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
//---IMPLEMENT PROPERTIES-----------------------------------------------------------------------------------------------
//
//
//

func (manager *iManager) GetChannelError() (channelForSendErrors chan itask.IError) {
	return manager.channelForSendErrors
}

func (manager *iManager) GetSizeQueue() (sizeOfQueue int64) {
	return int64(len(manager.sliceTask))
}

func (manager *iManager) QueueIsFilled(countTasks int64) (isFilled bool) {
	if (int64(len(manager.sliceTask)) + countTasks) > manager.sizeOfQueue {
		return true
	}
	return false
}

func (manager *iManager) DeleteTasksByKeys(keys map[string]struct{}) {
	if len(keys) == 0 {
		return
	}
	var (
		sliceTask = make([]itask.ITask, 0)
	)
	for next := 0; next < len(manager.sliceTask); next++ {
		task := manager.sliceTask[next]
		if task == nil {
			sliceTask = append(sliceTask, task)
			continue
		}
		_, exist := keys[task.GetKey()]
		if !exist {
			sliceTask = append(sliceTask, task)
		}
	}
	manager.sliceTask = sliceTask
	return
}

func (manager *iManager) FindTaskByKey(key string) (findTask itask.ITask, err error) {
	if key == "" {
		return nil, gotasker.ErrorTaskIsNotExist
	}
	for t := 0; t < len(manager.sliceTask); t++ {
		task := manager.sliceTask[t]
		if task.GetKey() == key {
			return task, nil
		}
	}
	return nil, gotasker.ErrorTaskIsNotExist
}

func (manager *iManager) FindRunBanTriggers() (runBanTriggers []itask.ITask) {
	runBanTriggers = make([]itask.ITask, 0)
	for next := 0; next < len(manager.sliceTask); next++ {
		task := manager.sliceTask[next]
		if task == nil {
			continue
		}
		isRunBan := task.GetState().IsRunBan()
		if isRunBan {
			if isTrigger, _ := task.IsTrigger(); isTrigger {
				runBanTriggers = append(runBanTriggers, task)
			}
		}
	}
	return runBanTriggers
}

func (manager *iManager) FindRunBanSimpleTasks() (runBanTasks []itask.ITask) {
	runBanTasks = make([]itask.ITask, 0)
	for next := 0; next < len(manager.sliceTask); next++ {
		task := manager.sliceTask[next]
		if task == nil {
			continue
		}
		if isTrigger, _ := task.IsTrigger(); isTrigger {
			continue
		}
		if isDependent, _ := task.IsDependent(); isDependent {
			continue
		}
		isRunBan := task.GetState().IsRunBan()
		if isRunBan {
			runBanTasks = append(runBanTasks, task)
		}
	}
	return runBanTasks
}

func (manager *iManager) FindDependentTasksIfTriggerNotExist(triggerKey string) (dependentsTasks []itask.ITask) {
	dependentsTasks = make([]itask.ITask, 0)
	for next := 0; next < len(manager.sliceTask); next++ {
		task := manager.sliceTask[next]
		if task == nil {
			continue
		}
		isDependent, trigger := task.IsDependent()
		if !isDependent {
			continue
		}
		if trigger != nil {
			if trigger.GetKey() == triggerKey {
				dependentsTasks = append(dependentsTasks, task)
			}
		}
	}
	return dependentsTasks
}

func (manager *iManager) SetRunBanForTasks(tasks ...itask.ITask) {
	for next := 0; next < len(tasks); next++ {
		task := tasks[next]
		stateTask := task.GetState()
		stateTask.SetRunBan(true)
		if isTrigger, dependentsTasks := task.IsTrigger(); isTrigger {
			for nextDependent := 0; nextDependent < len(dependentsTasks); nextDependent++ {
				somethingTask := dependentsTasks[nextDependent]
				if somethingTask == nil {
					continue
				}
				isDependent, _ := somethingTask.IsDependent()
				if !isDependent {
					continue
				}
				somethingTask.GetState().SetRunBan(true)
			}
		}
	}
	return
}

func (manager *iManager) TriggerIsCompleted(trigger itask.ITask) (isCompleted bool, dependentTasks map[string]bool, err error) {
	dependentTasks = make(map[string]bool)
	if isTrigger, dependents := trigger.IsTrigger(); !isTrigger {
		return false, nil, gotasker.ErrorTaskIsNotExist
	} else {
		countCompletedDependentTasks := 0
		for _, dependent := range dependents {
			if isDependent, trgr := dependent.IsDependent(); !isDependent {
				return false, nil, gotasker.ErrorTaskIsNotExist
			} else {
				if trgr.GetKey() != trigger.GetKey() {
					return false, nil, gotasker.ErrorTaskIsNotExist
				}
			}
			if dependent.GetState().IsCompleted() {
				countCompletedDependentTasks++
				dependentTasks[dependent.GetKey()] = true
			} else {
				dependentTasks[dependent.GetKey()] = false
			}
		}
		if countCompletedDependentTasks == len(dependents) {
			isCompleted = true
		} else {
			isCompleted = false
		}
	}
	return isCompleted, dependentTasks, nil
}

//
//
//
//---IMPLEMENT MANAGE QUEUES--------------------------------------------------------------------------------------------
//
//
//

func (manager *iManager) SendErrorToErrorChannel(err itask.IError) {
	log.Println("->GOTASKER: TASK [", err.GetTaskKey(), "] SEND ERROR. ")
	manager.channelForSendErrors <- err
}

func (manager *iManager) SetUpdateForTask(key string, somethingUpdateContext interface{}) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	var (
		updateTuple = &iTuple{}
		iErr        = &iError{}
	)
	updateTuple.SetTaskKey(key)
	task, err := manager.FindTaskByKey(key)
	if err != nil {
		iErr.SetTaskKey(key)
		iErr.SetError(err)
		updateTuple.SetError(iErr)
	} else {
		updateTuple.SetTask(task)
		updateTuple.SetSomethingUpdateContext(somethingUpdateContext)
	}
	manager.channelForManageUpdateTasks <- updateTuple
}

func (manager *iManager) ManageUpdates() {
	for updatesTuple := range manager.channelForManageUpdateTasks {
		iErr := updatesTuple.GetError()
		if iErr != nil {
			dependentsTasks := manager.FindDependentTasksIfTriggerNotExist(iErr.GetTaskKey())
			if len(dependentsTasks) != 0 {
				manager.SetRunBanForTasks(dependentsTasks...)
			}
			manager.SendErrorToErrorChannel(iErr)
			continue
		}
		task := updatesTuple.GetTask()
		somethingUpdateContext := updatesTuple.GetSomethingUpdateContext()
		event := task.GetState().GetEventUpdateState()
		err, sendToErrorChannel := event(task, somethingUpdateContext)
		if err != nil {
			if sendToErrorChannel {
				log.Println("->GOTASKER: TASK [", task.GetKey(), "] GET BAN FOR RUNNING AFTER UPDATE. ")
				manager.SetRunBanForTasks(task)
				manager.SendErrorToErrorChannel(&iError{
					err:     err,
					taskKey: task.GetKey(),
					task:    task,
				})
				continue
			}
		}
		if task.GetState().IsCompleted() {
			manager.channelForManageCompletedTasks <- task
		}
	}
}

func (manager *iManager) ManageCompleted() {
	for task := range manager.channelForManageCompletedTasks {
		var (
			isTrigger bool
			err       error
		)
		_, err = manager.FindTaskByKey(task.GetKey())
		if err != nil {
			continue
		}
		log.Println("->GOTASKER: TASK [", task.GetKey(), "] WAS COMPLETED. ")
		tasksKeysForDelete := manager.eventManageTasks(task)
		if isTrigger, _ = task.IsTrigger(); isTrigger {
			manager.RunDependentTasks(task)
		} else {
			if len(tasksKeysForDelete) != 0 {
				manager.DeleteTasksByKeys(tasksKeysForDelete)
			}
			manager.RunDeferTasks(false)
		}
	}
}

//
//
//
//---IMPLEMENT RUN TASKS------------------------------------------------------------------------------------------------
//
//
//

func (manager *iManager) RunTask(task itask.ITask) (doTaskAsDefer, sendToErrorChannel bool) {
	event := task.GetEventRunTask()
	doTaskAsDefer, sendToErrorChannel, err := event(task)
	if err != nil {
		if sendToErrorChannel {
			log.Println("->GOTASKER: TASK [", task.GetKey(), "] GET BAN FOR RUNNING. ")
			manager.SetRunBanForTasks(task)
			manager.SendErrorToErrorChannel(&iError{
				err:     err,
				taskKey: task.GetKey(),
				task:    task,
			})
		}
	}
	return doTaskAsDefer, sendToErrorChannel
}

func (manager *iManager) AddTaskAndTask(task itask.ITask) (err error) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	var (
		dependents                                             []itask.ITask
		isTrigger, isFilled, doTaskAsDefer, sendToErrorChannel bool
		countTasks                                             = int64(1)
	)
	isTrigger, dependents = task.IsTrigger()
	if isTrigger {
		countTasks += int64(len(dependents))
	}
	isFilled = manager.QueueIsFilled(countTasks)
	if isFilled {
		return gotasker.ErrorQueueIsFilled
	}
	manager.sliceTask = append(manager.sliceTask, task)
	if isTrigger {
		manager.sliceTask = append(manager.sliceTask, dependents...)
	}
	doTaskAsDefer, sendToErrorChannel = manager.RunTask(task)
	if !sendToErrorChannel {
		if doTaskAsDefer {
			log.Println("\t\t\t:GOTASKER: TASK IS DEFER [", task.GetKey(), "]")
			task.GetState().SetRunnable(false)
			task.GetState().SetDefer(true)
		} else {
			log.Println("\t\t\t:GOTASKER: TASK IS RUN [", task.GetKey(), "]")
			task.GetState().SetRunnable(true)
			task.GetState().SetDefer(false)
		}
	}
	return nil
}

func (manager *iManager) RunDependentTasks(task itask.ITask) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	isTrigger, dependentsTasks := task.IsTrigger()
	if !isTrigger {
		return
	}
	if !task.GetState().IsCompleted() {
		return
	}
	log.Println("->GOTASKER: RUN DEPENDENT TASKS. ")
	for next := 0; next < len(dependentsTasks); next++ {
		somethingTask := dependentsTasks[next]
		if somethingTask == nil {
			continue
		}
		isDependent, _ := somethingTask.IsDependent()
		if !isDependent {
			continue
		}
		stateTask := somethingTask.GetState()
		if stateTask.IsRunBan() {
			continue
		}
		if stateTask.IsRunnable() {
			continue
		}
		if stateTask.IsCompleted() {
			continue
		}
		doTaskAsDefer, sendToErrorChannel := manager.RunTask(somethingTask)
		if !sendToErrorChannel {
			if !doTaskAsDefer {
				log.Println("->\t\t\t[", task.GetKey(), "] IS RUN AS DEPENDENT")
				somethingTask.GetState().SetRunnable(true)
			}
		}
	}
	log.Println("->GOTASKER: RUN DEPENDENT TASKS COMPLETED. ")
	return
}

func (manager *iManager) RunDeferTasks(runDependentTasks bool) {
	manager.mx.Lock()
	defer manager.mx.Unlock()
	if len(manager.sliceTask) == 0 {
		log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		return
	}
	log.Println("->GOTASKER: RUN DEFER TASKS. ")
	for next := 0; next < len(manager.sliceTask); next++ {
		task := manager.sliceTask[next]
		if task == nil {
			continue
		}
		stateTask := task.GetState()
		if isDependent, trigger := task.IsDependent(); isDependent {
			if !runDependentTasks {
				continue
			} else {
				if !trigger.GetState().IsCompleted() {
					continue
				}
			}
		}
		if stateTask.IsRunBan() {
			continue
		}
		if stateTask.IsRunnable() {
			continue
		}
		if stateTask.IsCompleted() {
			continue
		}
		doTaskAsDefer, sendToErrorChannel := manager.RunTask(task)
		if !sendToErrorChannel {
			if !doTaskAsDefer {
				log.Println("->\t\t\t[", task.GetKey(), "] IS RUN AS DEFER")
				task.GetState().SetRunnable(true)
			}
		}
	}
	log.Println("->GOTASKER: RUN DEFER TASKS COMPLETED. ")
}

func (manager *iManager) RunDeferByTimer() {
	for {
		if len(manager.sliceTask) == 0 {
			log.Println("->+++++++++++GOTASKER: QUEUE IS EMPTY. READY GETTING NEXT TASKS+++++++++++")
		} else {
			manager.RunDeferTasks(true)
		}
		time.Sleep(manager.timeOut)
	}
}
