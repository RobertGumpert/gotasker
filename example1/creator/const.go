package creator

import (
	"github.com/RobertGumpert/gotasker/interfaces"
)

const (
	AddNumbers      interfaces.Type = 0
	MultiplyNumbers interfaces.Type = 1
)

type Send struct {
	Operation        interfaces.Type
	Key              string
	Number1, Number2 int
}

type Update struct {
	Key             string
	IsDenial        bool
	ExecutionStatus bool
	Result          int
}
