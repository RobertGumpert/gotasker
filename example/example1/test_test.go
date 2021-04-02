package example1

import (
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	count := 5
	size := 5
	timeout := 15 * time.Second
	creator := NewTaskCreator(int64(size), timeout)
	for i := 0; i < count; i++ {
		creator.CreateTaskAddNumbers(i, i)
	}
	time.Sleep(5 *time.Second)
	for i := 0; i < count; i++ {
		creator.CreateTaskAddNumbers((i+1)*10, (i+1)*10)
	}
	time.Sleep(1 *time.Minute)
}

func TestSimpleWithFewUpdates(t *testing.T) {
	count := 10
	size := 100
	timeout := 15 * time.Second
	creator := NewTaskCreator(int64(size), timeout)
	for i := 0; i < count; i++ {
		creator.CreateTaskMultiplyNumbers(i, i)
	}
	time.Sleep(1 *time.Hour)
}

func TestSimpleTrigger(t *testing.T) {
	count := 5
	size := 100
	timeout := 15 * time.Second
	creator := NewTaskCreator(int64(size), timeout)
	for i := 0; i < count; i++ {
		creator.CreateTriggerTask(i, i)
	}
	time.Sleep(1 *time.Hour)
}
