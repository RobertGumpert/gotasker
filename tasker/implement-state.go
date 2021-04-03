package tasker

import "github.com/RobertGumpert/gotasker/itask"

type implementStateInterface struct {
	isRunnable, isCompleted, isDefer                bool
	customFieldsContext, sendContext, updateContext interface{}
	eventUpdateTaskState                            itask.EventUpdateTaskState
	err                                             error
}

func (state *implementStateInterface) SetSendContext(send interface{}) {
	state.sendContext = send
}

func (state *implementStateInterface) GetSendContext() (send interface{}) {
	return state.sendContext
}

func (state *implementStateInterface) SetUpdateContext(update interface{}) {
	state.updateContext = update
}

func (state *implementStateInterface) GetUpdateContext() (update interface{}) {
	return state.updateContext
}

func (state *implementStateInterface) SetEventUpdateState(event itask.EventUpdateTaskState) {
	state.eventUpdateTaskState = event
}

func (state *implementStateInterface) GetEventUpdateState() (event itask.EventUpdateTaskState) {
	return state.eventUpdateTaskState
}

func (state *implementStateInterface) SetRunnable(flag bool) {
	state.isRunnable = flag
}

func (state *implementStateInterface) IsRunnable() (flag bool) {
	return state.isRunnable
}

func (state *implementStateInterface) SetCompleted(flag bool) {
	state.isCompleted = flag
}

func (state *implementStateInterface) IsCompleted() (flag bool) {
	return state.isCompleted
}

func (state *implementStateInterface) SetDefer(flag bool) {
	state.isDefer = flag
}

func (state *implementStateInterface) IsDefer() (flag bool) {
	return state.isDefer
}

func (state *implementStateInterface) SetCustomFields(fields interface{}) {
	state.customFieldsContext = fields
}

func (state *implementStateInterface) GetCustomFields() (fields interface{}) {
	return state.customFieldsContext
}

func (state *implementStateInterface) SetError(err error) {
	state.err = err
}

func (state *implementStateInterface) GetError() (err error) {
	return state.err
}
