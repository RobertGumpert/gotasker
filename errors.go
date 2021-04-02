package gotasker

import "errors"

var (
	ErrorTaskIsNotExist                = errors.New("Task Is Not Exist. ")
	ErrorTaskListIsEmpty               = errors.New("Task List Is Empty. ")
	ErrorQueueIsFilled                 = errors.New("Queue Is Filled. ")
	ErrorRunEventIsNil                 = errors.New("Run Event Is Nil. ")
	ErrorTriggerNotExecute             = errors.New("Trigger Not Execute. ")
	ErrorDependentTaskRunBeforeTrigger = errors.New("Dependent Task Run Before Trigger. ")
)
