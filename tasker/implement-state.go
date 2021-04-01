package tasker

import "github.com/RobertGumpert/gotasker/itask"

type taskerState struct {
	runnable, execute, def       bool
	result, fields, send, update interface{}
	eventUpdateTaskState         itask.EventUpdateTaskState
	err                          error
}

func (state *taskerState) SetSendContext(send interface{}) {
	state.send = send
}

func (state *taskerState) GetSendContext() (send interface{}) {
	return state.send
}

func (state *taskerState) SetUpdateContext(update interface{}) {
	state.update = update
}

func (state *taskerState) GetUpdateContext() (update interface{}) {
	return state.update
}

func (state *taskerState) SetEventUpdateState(event itask.EventUpdateTaskState) {
	state.eventUpdateTaskState = event
}

func (state *taskerState) GetEventUpdateState() (event itask.EventUpdateTaskState) {
	return state.eventUpdateTaskState
}

func (state *taskerState) SetRunnable(flag bool) {
	state.runnable = flag
}

func (state *taskerState) IsRunnable() (flag bool) {
	return state.runnable
}

func (state *taskerState) SetExecute(flag bool) {
	state.execute = flag
}

func (state *taskerState) IsExecute() (flag bool) {
	return state.execute
}

func (state *taskerState) SetDefer(flag bool) {
	state.def = flag
}

func (state *taskerState) IsDefer() (flag bool) {
	return state.def
}

func (state *taskerState) SetCustomFields(fields interface{}) {
	state.fields = fields
}

func (state *taskerState) GetCustomFields() (fields interface{}) {
	return state.fields
}

func (state *taskerState) SetError(err error) {
	state.err = err
}

func (state *taskerState) GetError() (err error) {
	return state.err
}
