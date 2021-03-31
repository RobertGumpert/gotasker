package creator

import (
	"errors"
	"github.com/RobertGumpert/gotasker/interfaces"
	"github.com/RobertGumpert/gotasker/tasker"
	"log"
	"strconv"
	"strings"
)

type resultState struct {
	operation interfaces.Type
	result    []int
}

type customFields struct {
	number1, number2 int
}

type TaskCreatorService struct {
	taskSteward    *tasker.Steward
	chanForSend    chan *Send
	chanForGetting chan *Update
	chanForDenial  chan bool
}

func NewTaskCreatorService(chanForSend chan *Send, chanForGetting chan *Update, chanForDenial chan bool) *TaskCreatorService {
	service := &TaskCreatorService{chanForSend: chanForSend, chanForGetting: chanForGetting, chanForDenial: chanForDenial}
	service.taskSteward = tasker.NewSteward(10, service.EventManageTasks)
	go service.getMessage()
	return service
}

func (s *TaskCreatorService) TaskAddNumbers(number1, number2 int) {
	task, err := s.taskSteward.CreateTask(
		AddNumbers,
		strings.Join([]string{
			"key-add",
			strconv.Itoa(number1),
			strconv.Itoa(number2),
		}, "-"),
		resultState{
			operation: AddNumbers,
			result:    make([]int, 0),
		},
		customFields{
			number1: number1,
			number2: number2,
		},
		s.EventRunTaskAddNumbers,
		s.EventUpdateStateAddNumbers,
	)
	if err != nil {
		log.Println(err)
		return
	}
	err = s.taskSteward.ServeTask(task)
	if err != nil {
		log.Println(err)
	}
}

func (s *TaskCreatorService) TaskMultiplyNumbers(number1, number2 int) {
	task, err := s.taskSteward.CreateTask(
		MultiplyNumbers,
		strings.Join([]string{
			"key-multiply",
			strconv.Itoa(number1),
			strconv.Itoa(number2),
		}, "-"),
		resultState{
			operation: MultiplyNumbers,
			result:    make([]int, 0),
		},
		customFields{
			number1: number1,
			number2: number2,
		},
		s.EventRunTaskMultiplyNumbers,
		s.EventUpdateStateMultiplyNumbers,
	)
	if err != nil {
		log.Println(err)
		return
	}
	err = s.taskSteward.ServeTask(task)
	if err != nil {
		log.Println(err)
	}
}

func (s *TaskCreatorService) EventRunTaskAddNumbers(task interfaces.ITask) error {
	s.chanForSend <- &Send{
		Operation: AddNumbers,
		Key:       task.GetKey(),
		Number1:   task.GetState().GetCustomFields().(customFields).number1,
		Number2:   task.GetState().GetCustomFields().(customFields).number2,
	}
	denial := <-s.chanForDenial
	if denial {
		return nil
	}
	return errors.New("IS DENIAL. ")
}

func (s *TaskCreatorService) EventRunTaskMultiplyNumbers(task interfaces.ITask) error {
	s.chanForSend <- &Send{
		Operation: MultiplyNumbers,
		Key:       task.GetKey(),
		Number1:   task.GetState().GetCustomFields().(customFields).number1,
		Number2:   task.GetState().GetCustomFields().(customFields).number2,
	}
	denial := <-s.chanForDenial
	if denial {
		return errors.New("IS DENIAL. ")
	}
	return nil
}

func (s *TaskCreatorService) EventUpdateStateAddNumbers(task interfaces.ITask, result interface{}) (err error) {
	current := task.GetState().GetResult().(resultState)
	update := result.(*Update)
	if update.ExecutionStatus {
		task.GetState().SetExecute(true)
	}
	current.result = append(current.result, update.Result)
	task.GetState().SetResult(current)
	return nil
}

func (s *TaskCreatorService) EventUpdateStateMultiplyNumbers(task interfaces.ITask, result interface{}) (err error) {
	current := task.GetState().GetResult().(resultState)
	update := result.(*Update)
	if update.ExecutionStatus {
		task.GetState().SetExecute(true)
	}
	current.result = append(current.result, update.Result)
	task.GetState().SetResult(current)
	return nil
}

func (s *TaskCreatorService) EventManageTasks(task interfaces.ITask) (deleteTasks []string) {
	deleteTasks = make([]string, 0)
	//
	if isTrigger, _ := task.IsTrigger(); isTrigger {
		return deleteTasks
	}
	if isDependent, trigger := task.IsDependent(); isDependent {
		isTrigger, tasks := trigger.IsTrigger()
		deleteTasks = append(deleteTasks, trigger.GetKey())
		if isTrigger {
			count := 0
			for t := 0; t < len(tasks); t++ {
				if tasks[t].GetState().IsExecute() {
					count++
					deleteTasks = append(deleteTasks, task.GetKey())
				}
			}
			if count == len(tasks) {
				return deleteTasks
			}
		}
	}
	switch task.GetType() {
	case AddNumbers:
		res := task.GetState().GetResult().(resultState)
		log.Println("EXECUTE TASK [", task.GetKey(), "] with result [", res.result, "]")
	case MultiplyNumbers:
		res := task.GetState().GetResult().(resultState)
		log.Println("EXECUTE TASK [", task.GetKey(), "] with result [", res.result, "]")
	}
	return nil
}


func (s *TaskCreatorService) getMessage() {
	for mes := range s.chanForGetting {
		log.Println("GETTING MESSAGE FOR EXECUTOR...")
		err := s.taskSteward.UpdateTaskStateExecute(mes.Key, mes)
		if err != nil {
			log.Println(err)
		}
		log.Println("GETTING MESSAGE FOR EXECUTOR.")
	}
}