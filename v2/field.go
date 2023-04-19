package failure

import (
	"fmt"
)

type internalKey int

const (
	KeyContext internalKey = iota + 1
	KeyMessage
	KeyCallStack
)

type Context map[string]string

func (c Context) ErrorFieldKey() any {
	return KeyContext
}

type Message string

func (m Message) ErrorFieldKey() any {
	return KeyMessage
}

func Messagef(format string, a ...any) Message {
	return Message(fmt.Sprintf(format, a...))
}
