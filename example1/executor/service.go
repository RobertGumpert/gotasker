package executor

import (
	"github.com/RobertGumpert/gotasker/example1/creator"
	"math/rand"
	"time"
)

type TaskExecutorService struct {
	chanForSend    chan *creator.Send
	chanForGetting chan *creator.Update
	chanForDenial  chan bool
}

func NewTaskExecutorService(chanForSend chan *creator.Send, chanForGetting chan *creator.Update, chanForDenial chan bool) *TaskExecutorService {
	service :=  &TaskExecutorService{chanForSend: chanForSend, chanForGetting: chanForGetting, chanForDenial: chanForDenial}
	go service.ScanChannel()
	return service
}

func (s *TaskExecutorService) ScanChannel() {
	for task := range s.chanForSend {
		denial := rand.Intn(2)
		if denial == 0 {
			s.chanForDenial <- true
		}
		//
		if task.Operation == creator.AddNumbers {
			go s.Add(task)
		}
		if task.Operation == creator.MultiplyNumbers {
			go s.Multiply(task)
		}
		//
		s.chanForDenial <- false
	}
}

func (s *TaskExecutorService) Add(send *creator.Send) {
	res := send.Number1 + send.Number2
	time.Sleep(3 * time.Second)
	s.chanForGetting <- &creator.Update{
		Key:             send.Key,
		IsDenial:        false,
		ExecutionStatus: true,
		Result:          res,
	}
}

func (s *TaskExecutorService) Multiply(send *creator.Send) {
	for i := 0; i < 10; i++ {
		res := send.Number1 * send.Number2
		time.Sleep(3 * time.Second)
		s.chanForGetting <- &creator.Update{
			Key:             send.Key,
			IsDenial:        false,
			ExecutionStatus: true,
			Result:          res,
		}
	}
}
