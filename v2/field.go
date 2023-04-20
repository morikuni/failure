package failure

import (
	"fmt"
	"strings"
)

type internalKey int

const (
	KeyCode internalKey = iota + 1
	KeyContext
	KeyMessage
	KeyCallStack
)

type codeField struct {
	code any
}

func (c codeField) SetErrorField(setter FieldSetter) {
	setter.Set(KeyCode, c.code)
}

type Context map[string]string

func (c Context) SetErrorField(setter FieldSetter) {
	setter.Set(KeyContext, c)
}

func (c Context) FormatError() string {
	var b strings.Builder
	first := true
	b.WriteRune('{')
	for k, v := range c {
		if !first {
			first = false
			b.WriteString(",")
		}
		first = false
		fmt.Fprintf(&b, "%s=%s", k, v)
	}
	b.WriteRune('}')
	return b.String()
}

type Message string

func (m Message) SetErrorField(setter FieldSetter) {
	setter.Set(KeyMessage, m)
}

func Messagef(format string, a ...any) Message {
	return Message(fmt.Sprintf(format, a...))
}
