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

// WithCode creates a new Field with the provided code.
// Generally, you don't need to use this function directly unless you're using NewStack,
// as Code will be automatically assigned when using New.
func WithCode[C Code](c C) Field {
	return codeField{c}
}

type codeField struct {
	code any
}

func (c codeField) SetErrorField(setter FieldSetter) {
	setter.Set(KeyCode, c.code)
}

// Context can be used to store additional information related to an error.
type Context map[string]string

// SetErrorField implements the Field interface.
func (c Context) SetErrorField(setter FieldSetter) {
	setter.Set(KeyContext, c)
}

// FormatError implements the ErrorFormatter interface.
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

// Message represents an error message displayed for human.
type Message string

// SetErrorField implements the Field interface.
func (m Message) SetErrorField(setter FieldSetter) {
	setter.Set(KeyMessage, m)
}

// Messagef creates a new Message with the provided format and arguments.
func Messagef(format string, a ...any) Message {
	return Message(fmt.Sprintf(format, a...))
}
