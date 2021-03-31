package example1

import (
	"github.com/RobertGumpert/gotasker/example1/creator"
	"github.com/RobertGumpert/gotasker/example1/executor"
	"testing"
	"time"
)

var (
	SendChan   = make(chan *creator.Send, 10)
	UpdateChan = make(chan *creator.Update, 10)
	DenialChan = make(chan bool, 10)
)

func Test(t *testing.T) {
	creatorService := creator.NewTaskCreatorService(
		SendChan,
		UpdateChan,
		DenialChan,
	)
	_ = executor.NewTaskExecutorService(
		SendChan,
		UpdateChan,
		DenialChan,
	)
	for i := 0; i < 10; i++ {
		creatorService.TaskAddNumbers(i*10, i*20)
	}
	time.Sleep(1*time.Hour)
}
