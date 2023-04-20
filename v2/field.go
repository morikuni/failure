package failure

import (
	"fmt"
	"io"
)

type internalKey int

const (
	KeyCode internalKey = iota + 1
	KeyContext
	KeyMessage
	KeyCallStack
)

func WithCode[C Code](c C) Field {
	return codeField{c}
}

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

func (c Context) FormatError(w io.Writer) {
	first := true
	io.WriteString(w, "{")
	for k, v := range c {
		if !first {
			first = false
			io.WriteString(w, ",")
		}
		first = false
		fmt.Fprintf(w, "%s=%s", k, v)
	}
	io.WriteString(w, "}")
}

type Message string

func (m Message) SetErrorField(setter FieldSetter) {
	setter.Set(KeyMessage, m)
}

func Messagef(format string, a ...any) Message {
	return Message(fmt.Sprintf(format, a...))
}
