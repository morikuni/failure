package failure

import (
	"bytes"
	"fmt"
	"sort"

	"io"
)

// Unwrapper interface is used by iterator.
type Unwrapper interface {
	// UnwrapError should return nearest child error.
	// The returned error can be nil.
	UnwrapError() error
}

// Wrapper interface is used by constructor functions.
type Wrapper interface {
	// WrapError should wrap given err to append some
	// capability to the error.
	WrapError(err error) error
}

// WrapperFunc is an adaptor to use function as the Wrapper interface.
type WrapperFunc func(err error) error

// WrapError implements the Wrapper interface.
func (f WrapperFunc) WrapError(err error) error {
	return f(err)
}

// Message appends an error message to the err.
func Message(msg string) Wrapper {
	return WrapperFunc(func(err error) error {
		return &withMessage{msg, err}
	})
}

// Messagef appends formatted error message to an error.
func Messagef(format string, args ...interface{}) Wrapper {
	return WrapperFunc(func(err error) error {
		return &withMessage{fmt.Sprintf(format, args...), err}
	})
}

type withMessage struct {
	message    string
	underlying error
}

func (w *withMessage) Error() string {
	return fmt.Sprintf("%s: %s", w.message, w.underlying)
}

func (w *withMessage) UnwrapError() error {
	return w.underlying
}

func (w *withMessage) GetMessage() string {
	return w.message
}

// MessageOf extracts a message from the err.
func MessageOf(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	type messageGetter interface {
		GetMessage() string
	}

	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if g, ok := err.(messageGetter); ok {
			return g.GetMessage(), true
		}
	}

	return "", false
}

// Context is a key-value data which describes the how the error occurred
// for debugging purpose. You must not use context data as a part of your
// application logic. Just print it.
// If you want to extract Context from error for printing purpose, you can
// define an interface with method `GetContext() Context` and use it with
// iterator, like other extraction functions (see: MessageOf).
type Context map[string]string

// WrapError implements the Wrapper interface.
func (m Context) WrapError(err error) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	buf := &bytes.Buffer{}
	for _, k := range keys {
		v := m[k]
		if buf.Len() != 0 {
			buf.WriteRune(' ')
		}
		fmt.Fprintf(buf, "%s=%s", k, v)
	}
	return &withContext{m, buf.String(), err}
}

type withContext struct {
	ctx        Context
	memo       string
	underlying error
}

func (m *withContext) Error() string {
	return fmt.Sprintf("%s: %s", m.memo, m.underlying)
}

func (m *withContext) UnwrapError() error {
	return m.underlying
}

func (m *withContext) GetContext() Context {
	return m.ctx
}

// WithCallStackSkip appends a call stack to the err skipping first N frames.
// You don't have to use this directly, unless using function Custom.
func WithCallStackSkip(skip int) Wrapper {
	cs := Callers(skip + 1)
	return WrapperFunc(func(err error) error {
		return &withCallStack{
			cs,
			err,
		}
	})
}

type withCallStack struct {
	callStack  CallStack
	underlying error
}

func (w *withCallStack) Error() string {
	head := w.callStack.HeadFrame()
	return fmt.Sprintf("%s.%s: %s", head.Pkg(), head.Func(), w.underlying)
}

func (w *withCallStack) UnwrapError() error {
	return w.underlying
}

func (w *withCallStack) GetCallStack() CallStack {
	return w.callStack
}

// CallStackOf extracts a call stack from the err.
// Returned call stack is for the most deepest place (appended first).
func CallStackOf(err error) (CallStack, bool) {
	if err == nil {
		return nil, false
	}

	type callStackGetter interface {
		GetCallStack() CallStack
	}

	var last CallStack
	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if g, ok := err.(callStackGetter); ok {
			last = g.GetCallStack()
		}
	}

	if last == nil {
		return nil, false
	}
	return last, true
}

// WithFormatter appends an error formatter to the err.
//
//     %v+: Print trace for each place, and deepest call stack.
//     %#v: Print raw structure of the error.
//     others (%s, %v): Same as err.Error().
//
// You don't have to use this directly, unless using function Custom.
func WithFormatter() Wrapper {
	return WrapperFunc(func(err error) error {
		return &formatter{err}
	})
}

type formatter struct {
	error
}

func (f *formatter) UnwrapError() error {
	return f.error
}

func (f *formatter) IsFormatter() {}

func (f *formatter) Format(s fmt.State, verb rune) {
	if verb != 'v' { // %s
		io.WriteString(s, f.Error())
		return
	}

	if s.Flag('#') { // %#v
		type formatter struct {
			error
		}
		fmt.Fprintf(s, "%#v", &formatter{f.error})
		return
	}

	if !s.Flag('+') { // %v
		io.WriteString(s, f.Error())
		return
	}

	// %+v
	type callStacker interface {
		GetCallStack() CallStack
	}
	type contexter interface {
		GetContext() Context
	}
	type messenger interface {
		GetMessage() string
	}
	type coder interface {
		GetCode() Code
	}
	type formatter interface {
		IsFormatter()
	}

	i := NewIterator(f.error)
	for i.Next() {
		err := i.Error()
		switch t := err.(type) {
		case callStacker:
			fmt.Fprintf(s, "%+v\n", t.GetCallStack().HeadFrame())
		case contexter:
			kv := t.GetContext()
			for k, v := range kv {
				fmt.Fprintf(s, "    %s = %s\n", k, v)
			}
		case messenger:
			fmt.Fprintf(s, "    message(%q)\n", t.GetMessage())
		case coder:
			fmt.Fprintf(s, "    code(%s)\n", t.GetCode().ErrorCode())
		case formatter:
			// do nothing
		default:
			fmt.Fprintf(s, "    %T(%q)\n", err, err.Error())
		}
	}

	fmt.Fprint(s, "[CallStack]\n")
	if cs, ok := CallStackOf(f); ok {
		for _, f := range cs.Frames() {
			fmt.Fprintf(s, "    %+v\n", f)
		}
	}
}
