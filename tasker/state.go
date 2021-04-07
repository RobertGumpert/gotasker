package tasker

import "github.com/RobertGumpert/gotasker/itask"

type iState struct {
	isRunnable, isCompleted, isDefer, runBan                bool
	customFieldsContext, sendContext, updateContext interface{}
	eventUpdateTaskState                            itask.EventUpdateTaskState
	err                                             error
}

func (s *iState) SetRunBan(flag bool) {
	s.runBan = flag
}

func (s *iState) IsRunBan() (flag bool) {
	return s.runBan
}

func (s *iState) SetSendContext(send interface{}) {
	s.sendContext = send
}


func (s *iState) GetSendContext() (send interface{}) {
	return s.sendContext
}

func (s *iState) SetUpdateContext(update interface{}) {
	s.updateContext = update
}

func (s *iState) GetUpdateContext() (update interface{}) {
	return s.updateContext
}

func (s *iState) SetEventUpdateState(event itask.EventUpdateTaskState) {
	s.eventUpdateTaskState = event
}

func (s *iState) GetEventUpdateState() (event itask.EventUpdateTaskState) {
	return s.eventUpdateTaskState
}

func (s *iState) SetRunnable(flag bool) {
	s.isRunnable = flag
}

func (s *iState) IsRunnable() (flag bool) {
	return s.isRunnable
}

func (s *iState) SetCompleted(flag bool) {
	s.isCompleted = flag
}

func (s *iState) IsCompleted() (flag bool) {
	return s.isCompleted
}

func (s *iState) SetDefer(flag bool) {
	s.isDefer = flag
}

func (s *iState) IsDefer() (flag bool) {
	return s.isDefer
}

func (s *iState) SetCustomFields(fields interface{}) {
	s.customFieldsContext = fields
}

func (s *iState) GetCustomFields() (fields interface{}) {
	return s.customFieldsContext
}

func (s *iState) SetError(err error) {
	s.err = err
}

func (s *iState) GetError() (err error) {
	return s.err
}
