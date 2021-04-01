package example1

import (
	"github.com/RobertGumpert/gotasker/itask"
)

const (
	AddNumbers      itask.Type = 0
	MultiplyNumbers itask.Type = 1
)

type InterProgramSendContext struct {
	Operation        itask.Type
	Key              string
	Number1, Number2 int
}

type InterProgramUpdateContext struct {
	Operation   itask.Type
	IsCompleted bool
	Key         string
	Result      int
}
