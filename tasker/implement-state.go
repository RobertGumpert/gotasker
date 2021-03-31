package tasker

import (
	"github.com/RobertGumpert/gotasker/interfaces"
)

type implementState struct {
	runnable, execute, def bool
	result, fields interface{}
	event interfaces.EventUpdateState
	err error
}

func (state *implementState) SetEventUpdateState(event interfaces.EventUpdateState) {
	state.event = event
}

func (state *implementState) GetEventUpdateState() (event interfaces.EventUpdateState) {
	return state.event
}

func (state *implementState) SetRunnable(flag bool) {
	state.runnable = flag
}

func (state *implementState) IsRunnable() (flag bool) {
	return state.runnable
}

func (state *implementState) SetExecute(flag bool) {
	state.execute = flag
}

func (state *implementState) IsExecute() (flag bool) {
	return state.execute
}

func (state *implementState) SetDefer(flag bool) {
	state.def = flag
}

func (state *implementState) IsDefer() (flag bool) {
	return state.def
}

func (state *implementState) SetResult(result interface{}) {
	state.result = result
}

func (state *implementState) GetResult() (result interface{}) {
	return state.result
}

func (state *implementState) SetCustomFields(fields interface{}) {
	state.fields = fields
}

func (state *implementState) GetCustomFields() (fields interface{}) {
	return state.fields
}

func (state *implementState) SetError(err error) {
	state.err = err
}

func (state *implementState) GetError() (err error) {
	return state.err
}
