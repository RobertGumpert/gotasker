package example1

import (
	"errors"
	"github.com/RobertGumpert/gotasker/itask"
	"github.com/RobertGumpert/gotasker/tasker"
	"log"
	"strconv"
	"strings"
	"time"
)

type TaskCreator struct {
	executor *TaskExecutor
	steward  *tasker.Steward
	res      map[string]int
}

type TaskerUpdateContext struct {
	Operation itask.Type
	Key       string
	Result    []int
}

func NewTaskCreator(size int64, timeout time.Duration) *TaskCreator {
	creator := new(TaskCreator)
	creator.executor = &TaskExecutor{gettingResult: creator.gettingUpdates}
	creator.res = make(map[string]int)
	creator.steward = tasker.NewSteward(
		size,
		timeout,
		creator.eventManageTask,
	)
	return creator
}

func (creator *TaskCreator) CreateTaskAddNumbers(n1, n2 int) {
	key := strings.Join([]string{
		"key-add",
		strconv.Itoa(n1),
		strconv.Itoa(n2),
	}, "-")
	_, err := creator.steward.CreateTaskAndRun(
		creator.steward.CreateTask(
			AddNumbers,
			key,
			&InterProgramSendContext{
				Operation: AddNumbers,
				Key:       key,
				Number1:   n1,
				Number2:   n2,
			},
			&TaskerUpdateContext{
				Operation: AddNumbers,
				Key:       key,
				Result:    make([]int, 0),
			},
			nil,
			creator.eventRunTask,
			creator.eventUpdateTask,
		),
	)
	if err != nil {
		log.Println(err)
	}
}

func (creator *TaskCreator) CreateTaskMultiplyNumbers(n1, n2 int) {
	key := strings.Join([]string{
		"key-multiply",
		strconv.Itoa(n1),
		strconv.Itoa(n2),
	}, "-")
	_, err := creator.steward.CreateTaskAndRun(
		creator.steward.CreateTask(
			MultiplyNumbers,
			key,
			&InterProgramSendContext{
				Operation: MultiplyNumbers,
				Key:       key,
				Number1:   n1,
				Number2:   n2,
			},
			&TaskerUpdateContext{
				Operation: MultiplyNumbers,
				Key:       key,
				Result:    make([]int, 0),
			},
			nil,
			creator.eventRunTask,
			creator.eventUpdateTask,
		),
	)
	if err != nil {
		log.Println(err)
	}
}

func (creator *TaskCreator) CreateTriggerTask(n1, n2 int) {
	key := strings.Join([]string{
		"trigger-key-add",
		strconv.Itoa(n1),
		strconv.Itoa(n2),
	}, "-")
	trigger := creator.steward.CreateTask(
		AddNumbers,
		key,
		&InterProgramSendContext{
			Operation: AddNumbers,
			Key:       key,
			Number1:   n1,
			Number2:   n2,
		},
		&TaskerUpdateContext{
			Operation: AddNumbers,
			Key:       key,
			Result:    make([]int, 0),
		},
		nil,
		creator.eventRunTask,
		creator.eventUpdateTask,
	)
	dependents := make([]itask.TaskConstructor, 0)
	for i := 0; i < 5; i++ {
		key := strings.Join([]string{
			"depedent-key-add",
			strconv.Itoa(n1) + ":" + strconv.Itoa(i),
			strconv.Itoa(n2) + ":" + strconv.Itoa(i),
		}, "-")
		dependents = append(dependents,
			creator.steward.CreateTask(
				AddNumbers,
				key,
				&InterProgramSendContext{
					Operation: AddNumbers,
					Key:       key,
					Number1:   n1,
					Number2:   n2,
				},
				&TaskerUpdateContext{
					Operation: AddNumbers,
					Key:       key,
					Result:    make([]int, 0),
				},
				nil,
				creator.eventRunTask,
				creator.eventUpdateTask,
			),
		)
	}
	_, err := creator.steward.CreateTriggerAndRun(
		trigger,
		dependents...,
	)
	if err != nil {
		log.Println(err)
	}
}

//
//--------------------------------------EVENTS--------------------------------------------------------------------------
//

func (creator *TaskCreator) eventRunTask(task itask.ITask) (err error) {
	s := task.GetState().GetSendContext().(*InterProgramSendContext)
	switch task.GetType() {
	case AddNumbers:
		isRunning := creator.executor.Add(s)
		if !isRunning {
			return errors.New("The executor refused to complete the task. ")
		}
		break
	case MultiplyNumbers:
		isRunning := creator.executor.Multiply(s)
		if !isRunning {
			return errors.New("The executor refused to complete the task. ")
		}
		break
	}
	return nil
}

func (creator *TaskCreator) eventUpdateTask(task itask.ITask, interProgramUpdateContext interface{}) (err error) {
	cast := interProgramUpdateContext.(*InterProgramUpdateContext)
	if cast.IsCompleted {
		task.GetState().SetCompleted(true)
	}
	context := task.GetState().GetUpdateContext().(*TaskerUpdateContext)
	context.Result = append(context.Result, cast.Result)
	task.GetState().SetUpdateContext(context)
	return nil
}

func (creator *TaskCreator) eventManageTask(task itask.ITask) (deleteTasks, saveTasks map[string]struct{}) {
	deleteTasks, saveTasks = make(map[string]struct{}), make(map[string]struct{})
	//
	res := make([]string, 0)
	slice := task.GetState().GetUpdateContext().(*TaskerUpdateContext).Result
	for i := 0; i < len(slice); i++ {
		res = append(res, strconv.Itoa(slice[i]))
	}
	str := strings.Join(res, ",")
	//
	var (
		isTrigger, dependents = task.IsTrigger()
		isDependent, trigger  = task.IsDependent()
	)
	if !isTrigger && !isDependent {
		switch task.GetType() {
		case AddNumbers:
			if task.GetState().IsCompleted() {
				deleteTasks[task.GetKey()] = struct{}{}
				if _, exist := creator.res[task.GetKey()]; exist {
					panic("RESULT IS EXIST: "+ task.GetKey())
				} else {
					creator.res[task.GetKey()] = 0
				}
			}
			break
		case MultiplyNumbers:
			if task.GetState().IsCompleted() {
				deleteTasks[task.GetKey()] = struct{}{}
				if _, exist := creator.res[task.GetKey()]; exist {
					panic("RESULT IS EXIST: "+ task.GetKey())
				} else {
					creator.res[task.GetKey()] = 0
				}
			}
			break
		}
	} else {
		if isDependent {
			if _, exist := creator.res[task.GetKey()]; exist {
				panic("RESULT IS EXIST: "+ task.GetKey())
			} else {
				creator.res[task.GetKey()] = 0
			}
			//
			count := 0
			_, dependents = trigger.IsTrigger()
			for _, dependent := range dependents {
				if dependent.GetState().IsCompleted() {
					deleteTasks[dependent.GetKey()] = struct{}{}
					count++
				}
			}
			if count == len(dependents) {
				deleteTasks[trigger.GetKey()] = struct{}{}
				log.Println("-> TaskCreator: TRIGGER WAS COMPLETED ", trigger.GetKey(), " . ")
				if _, exist := creator.res[trigger.GetKey()]; exist {
					panic("RESULT IS EXIST: "+ task.GetKey())
				} else {
					creator.res[trigger.GetKey()] = 0
				}
			} else {
				deleteTasks = nil
			}
		}
	}
	log.Println("-> TaskCreator: RESULT", task.GetKey(), " (len.=", len(res), ") = ", str, "")
	return deleteTasks, saveTasks
}

func (creator *TaskCreator) gettingUpdates(st *InterProgramUpdateContext) {
	err := creator.steward.UpdateTask(st.Key, st)
	if err != nil {
		log.Println("++++++++++++++++GETTING UPDATES WITH ERROR [", st.Key, "]: ", err)
	}
	//go func(creator *TaskCreator, st *InterProgramUpdateContext) {
	//	err := creator.steward.UpdateTask(st.Key, st)
	//	if err != nil {
	//		log.Println("++++++++++++++++GETTING UPDATES WITH ERROR [", st.Key, "]: ", err)
	//	}
	//}(creator, st)
}
