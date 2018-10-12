package failure

import (
	"fmt"

	"io"

	"github.com/pkg/errors"
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

// Message appends error message to an error.
func Message(msg string) Wrapper {
	return WrapperFunc(func(err error) error {
		return withMessage{err, msg}
	})
}

type withMessage struct {
	error
	message string
}

func (w withMessage) UnwrapError() error {
	return w.error
}

func (w withMessage) GetMessage() string {
	return w.message
}

// MessageOf extracts the message from err.
func MessageOf(err error) string {
	if err == nil {
		return ""
	}

	type messageGetter interface {
		GetMessage() string
	}

	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if g, ok := err.(messageGetter); ok {
			return g.GetMessage()
		}
	}

	return ""
}

// Debug is a key-value data appended to an error
// for debugging purpose.
type Debug map[string]interface{}

// WrapError implements the Wrapper interface.
func (d Debug) WrapError(err error) error {
	return withDebug{err, d}
}

type withDebug struct {
	error
	debug Debug
}

func (w withDebug) UnwrapError() error {
	return w.error
}

func (w withDebug) GetDebug() Debug {
	return w.debug
}

// DebugsOf extracts list of information from the error.
func DebugsOf(err error) []Debug {
	if err == nil {
		return nil
	}

	type debugGetter interface {
		GetDebug() Debug
	}

	var debugs []Debug
	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if g, ok := err.(debugGetter); ok {
			debugs = append(debugs, g.GetDebug())
		}
	}

	return debugs
}

// WithCallStackSkip appends call stack to an error
// skipping top N of frames.
func WithCallStackSkip(skip int) Wrapper {
	cs := Callers(skip + 1)
	return WrapperFunc(func(err error) error {
		return withCallStack{
			err,
			cs,
		}
	})
}

type withCallStack struct {
	err       error
	callStack CallStack
}

func (w withCallStack) Error() string {
	return fmt.Sprintf("%s: %s", w.callStack.HeadFrame().Func(), w.err.Error())
}

func (w withCallStack) UnwrapError() error {
	return w.err
}

func (w withCallStack) GetCallStack() CallStack {
	return w.callStack
}

// CallStackOf extracts call stack from the error.
// Returned call stack is for the most deepest place (appended first).
func CallStackOf(err error) CallStack {
	if err == nil {
		return nil
	}

	type callStackGetter interface {
		GetCallStack() CallStack
	}
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var last CallStack
	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		switch t := err.(type) {
		case callStackGetter:
			last = t.GetCallStack()
		case stackTracer:
			last = callStackFromPkgErrors(t.StackTrace())
		}
	}

	return last
}

// WithFormatter appends error formatter to an error.
//
//     %v+: Print trace for each place, and deepest call stack.
//     %#v: Print raw structure of the error.
//     others (%s, %v): Same as err.Error().
func WithFormatter() Wrapper {
	return WrapperFunc(func(err error) error {
		return formatter{err}
	})
}

type formatter struct {
	error
}

func (f formatter) UnwrapError() error {
	return f.error
}

func (f formatter) Format(s fmt.State, verb rune) {
	if verb != 'v' { // %s
		io.WriteString(s, f.Error())
		return
	}

	if s.Flag('#') { // %#v
		type formatter struct {
			error
		}
		fmt.Fprintf(s, "%#v", formatter{f.error})
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
	type debugger interface {
		GetDebug() Debug
	}
	type messenger interface {
		GetMessage() string
	}
	type coder interface {
		GetCode() Code
	}

	i := NewIterator(f.error)
	for i.Next() {
		err := i.Error()
		switch t := err.(type) {
		case callStacker:
			fmt.Fprintf(s, "%+v\n", t.GetCallStack().HeadFrame())
		case debugger:
			debug := t.GetDebug()
			for k, v := range debug {
				fmt.Fprintf(s, "    %s = %v\n", k, v)
			}
		case messenger:
			fmt.Fprintf(s, "    message(%q)\n", t.GetMessage())
		case coder:
			fmt.Fprintf(s, "    code(%s)\n", t.GetCode().ErrorCode())
		case formatter:
			// do nothing
		default:
			fmt.Fprintf(s, "    error(%q)\n", err.Error())
		}
	}

	fmt.Fprint(s, "[CallStack]\n")
	if cs := CallStackOf(f); cs != nil {
		for _, f := range cs.Frames() {
			fmt.Fprintf(s, "    %+v\n", f)
		}
	}
}
