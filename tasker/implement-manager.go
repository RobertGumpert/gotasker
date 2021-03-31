package tasker

import (
	"github.com/RobertGumpert/gotasker"
	"github.com/RobertGumpert/gotasker/interfaces"
	"log"
)

type implementManager struct {
	size    int64
	tasks   []interfaces.ITask
	channel chan interfaces.ITask
	event   interfaces.EventManageTasks
}

func newImplementManager(size int64, event interfaces.EventManageTasks) *implementManager {
	manager := &implementManager{
		size:    size,
		event:   event,
		channel: make(chan interfaces.ITask),
		tasks:   make([]interfaces.ITask, 0),
	}
	go manager.ScanChannelCompletedTasks()
	return manager
}

func (manager *implementManager) SetEventManageTasks(event interfaces.EventManageTasks) {
	manager.event = event
}

func (manager *implementManager) GetEventManageTasks() (event interfaces.EventManageTasks) {
	return manager.event
}

func (manager *implementManager) SetSizeQueue(size int64) {
	manager.size = size
}

func (manager *implementManager) GetSizeQueue() (size int64) {
	return manager.size
}

func (manager *implementManager) GetTask(key string) (task interfaces.ITask, err error) {
	for t := 0; t < len(manager.tasks); t++ {
		if manager.tasks[t].GetKey() == key {
			return manager.tasks[t], nil
		}
	}
	return nil, gotasker.ErrorTaskIsNotExist
}

func (manager *implementManager) ScanChannelCompletedTasks() {
	var (
		deleteTask = func(key string) {
			i := -1
			for t := 0; t < len(manager.tasks); t++ {
				if manager.tasks[t].GetKey() == key {
					i = t
				}
			}
			if i == -1 {
				log.Println("\t\t\t->GOTASKER: SCAN CHANNEL HAVE ERROR: ", gotasker.ErrorTaskIsNotExist)
			} else {
				manager.tasks[i] = manager.tasks[len(manager.tasks)-1]
				manager.tasks[len(manager.tasks)-1] = nil
				manager.tasks = manager.tasks[:len(manager.tasks)-1]
			}
		}
	)
	for task := range manager.channel {
		log.Println("GOTASKER: SCAN CHANNEL GETTING EXECUTE TASK...")
		keysTasksForDelete := manager.event(task)
		if len(keysTasksForDelete) != 0 {
			for key := 0; key < len(keysTasksForDelete); key++ {
				deleteTask(keysTasksForDelete[key])
			}
		} else {
			if isTrigger, _ := task.IsTrigger(); isTrigger == true {
				err := manager.RunDependentTasks(task)
				if err != nil {
					log.Println("\t\t\t->GOTASKER: SCAN CHANNEL HAVE ERROR: ", err)
				}
			}
		}
		manager.RunDeferTasks()
		log.Println("GOTASKER: SCAN CHANNEL WAIT NEXT")
	}
}

func (manager *implementManager) RunDeferTasks() {
	log.Println("GOTASKER: RUN DEFER...")
	for t := 0; t < len(manager.tasks); t++ {
		task := manager.tasks[t]
		if task.GetState().IsExecute() {
			continue
		}
		if isDependent, _ := task.IsDependent(); isDependent == true {
			continue
		}
		if task.GetState().IsRunnable() == true {
			continue
		}
		if task.GetState().IsDefer() == false {
			continue
		}
		if err := task.GetEventRunTask()(task); err == nil {
			task.GetState().SetRunnable(true)
			log.Println("\t\t\t->GOTASKER: TASK KEY [", task.GetKey(), "] IS RUN")
		}
	}
	log.Println("GOTASKER: FINISH RUN DEFER")
}

func (manager *implementManager) RunDependentTasks(trigger interfaces.ITask) (err error) {
	log.Println("GOTASKER: RUN DEPENDENTS FOR TRIGGER TASK KEY [", trigger.GetKey(), "]...")
	flag, dependents := trigger.IsTrigger()
	if !flag {
		return gotasker.ErrorTaskIsNotExist
	}
	if !trigger.GetState().IsExecute() {
		return gotasker.ErrorTriggerNotExecute
	}
	for t := 0; t < len(dependents); t++ {
		task := dependents[t]
		if task.GetState().IsExecute() {
			continue
		}
		if isDependent, _ := task.IsDependent(); isDependent == false {
			continue
		}
		if task.GetState().IsRunnable() == true {
			continue
		}
		if err := task.GetEventRunTask()(task); err == nil {
			task.GetState().SetRunnable(true)
			log.Println("\t\t\t->GOTASKER: TASK KEY [", task.GetKey(), "] IS RUN")
		}
	}
	log.Println("GOTASKER: FINISH DEPENDENTS FOR TRIGGER TASK KEY [", trigger.GetKey(), "]")
	return err
}

func (manager *implementManager) CreateTask(t interfaces.Type, key string, result, fields interface{}, eventRunTask interfaces.EventRunTask, eventUpdateState interfaces.EventUpdateState) (task interfaces.ITask, err error) {
	if int64(len(manager.tasks)) == manager.size {
		return nil, gotasker.ErrorQueueIsFilled
	}
	state := &implementState{
		result: result,
		fields: fields,
		event:  eventUpdateState,
	}
	return &implementTask{
		t:          t,
		ownKeyTask: key,
		event:      eventRunTask,
		state:      state,
	}, nil
}

func (manager *implementManager) ModifyTaskAsTrigger(trigger interfaces.ITask, dependents ...interfaces.ITask) (task interfaces.ITask, err error) {
	if int64(len(manager.tasks)+(len(dependents)+1)) == manager.size {
		return nil, gotasker.ErrorQueueIsFilled
	}
	for t := 0; t < len(dependents); t++ {
		dependents[t].DoAsDependent(trigger)
		dependents[t].GetState().SetDefer(true)
		dependents[t].GetState().SetRunnable(false)
	}
	trigger.DoAsTrigger(dependents)
	return trigger, nil
}

func (manager *implementManager) Add(task interfaces.ITask) (err error) {
	var count int64 = 1
	flag, dependents := task.IsTrigger()
	if !flag {
		if int64(len(manager.tasks)) == manager.size {
			return gotasker.ErrorQueueIsFilled
		}
	} else {
		count += int64(len(dependents))
		if int64(len(manager.tasks))+count == manager.size {
			return gotasker.ErrorQueueIsFilled
		}
	}
	log.Println("GOTASKER: ADD TASK [",task.GetKey(),"]...")
	manager.tasks = append(manager.tasks, task)
	errRunning := task.GetEventRunTask()(task)
	if errRunning != nil {
		log.Println("\t\t\t->GOTASKER: TASK IS RUN [",task.GetKey(),"]")
		task.GetState().SetRunnable(true)
		task.GetState().SetDefer(false)
	} else {
		log.Println("\t\t\t->GOTASKER: TASK IS DEFER [",task.GetKey(),"]")
		task.GetState().SetRunnable(false)
		task.GetState().SetDefer(true)
	}
	log.Println("GOTASKER: ADD TASK [",task.GetKey(),"].")
	return nil
}

func (manager *implementManager) UpdateTaskStateExecute(key string, result interface{}) (err error) {
	log.Println("GOTASKER: UPDATE TASK STATE [",key,"]...")
	task, err := manager.GetTask(key)
	if err != nil {
		return err
	}
	err = task.GetState().GetEventUpdateState()(task, result)
	if err != nil {
		log.Println("\t\t\t->GOTASKER: UPDATE TASK STATE [",key,"] HAVE ERR: ", err)
		return err
	}
	if task.GetState().IsExecute() {
		task.GetState().SetExecute(true)
		go func(manager *implementManager, task interfaces.ITask) {
			manager.channel <- task
		}(manager, task)
	}
	log.Println("GOTASKER: UPDATE TASK STATE [",key,"].")
	return nil
}
