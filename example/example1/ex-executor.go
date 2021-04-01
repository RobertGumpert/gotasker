package example1

import (
	"math/rand"
	"time"
)

type TaskExecutor struct {
	gettingResult func(st *InterProgramUpdateContext)
}

func NewTaskExecutor(gettingResult func(st *InterProgramUpdateContext)) *TaskExecutor {
	return &TaskExecutor{gettingResult: gettingResult}
}

func (executor *TaskExecutor) Add(s *InterProgramSendContext) bool {
	denial := rand.Intn(2)
	if denial == 0 {
		return false
	}
	go func(executor *TaskExecutor, s *InterProgramSendContext) {
		res := s.Number1 + s.Number2
		time.Sleep(3*time.Second)
		executor.gettingResult(&InterProgramUpdateContext{
			Operation:   s.Operation,
			Key:         s.Key,
			IsCompleted: true,
			Result:      res,
		})
		return
	}(executor, s)
	return true
}

func (executor *TaskExecutor) Multiply(s *InterProgramSendContext) bool {
	denial := rand.Intn(2)
	if denial == 0 {
		return false
	}
	go func(executor *TaskExecutor, s *InterProgramSendContext) {
		for i := 0; i<10; i++ {
			res := s.Number1 * s.Number2
			time.Sleep(3*time.Second)
			executor.gettingResult(&InterProgramUpdateContext{
				Operation:   s.Operation,
				Key:         s.Key,
				IsCompleted: false,
				Result:      res,
			})
		}
		time.Sleep(3*time.Second)
		executor.gettingResult(&InterProgramUpdateContext{
			Operation:   s.Operation,
			Key:         s.Key,
			IsCompleted: true,
			Result:      0,
		})
		return
	}(executor, s)
	return true
}