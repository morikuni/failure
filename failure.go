// Package failure is a utility package for handling application errors.
// Inspired by https://middlemost.com/failure-is-your-domain and github.com/pkg/errors.
package failure

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

var (
	// DefaultMessage is a default message for an error.
	DefaultMessage = "An internal error has occurred. Please try it later or contact the developer."

	// Unknown represents unknown error code.
	Unknown Code = "unknown"
)

// Info is key-value data.
type Info map[string]interface{}

// Failure is an error representing failure of something.
type Failure struct {
	// Code is an error code to handle the error in your source code.
	Code Code
	// Message is an error message for the application user.
	// So the message should be human-readable and be helpful.
	Message string
	// CallStack is a call stack at the time of the error occurred.
	CallStack CallStack
	// Info is optional information on why the error occurred.
	Info Info
	// Underlying is an underlying error of the failure.
	Underlying error
}

// WithMessage attaches a message to the failure.
func (f Failure) WithMessage(message string) Failure {
	f.Message = message
	return f
}

// WithInfo attaches information to the failure.
func (f Failure) WithInfo(info Info) Failure {
	f.Info = info
	return f
}

// Failure implements error.
func (f Failure) Error() string {
	buf := &bytes.Buffer{}

	if len(f.CallStack) != 0 {
		buf.WriteString(f.CallStack[0].Func())
	}

	if f.Code != "" {
		if buf.Len() != 0 {
			buf.WriteRune('(')
			buf.WriteString(string(f.Code))
			buf.WriteRune(')')
		} else {
			buf.WriteString(string(f.Code))
		}
	}

	if f.Underlying != nil {
		if buf.Len() != 0 {
			buf.WriteString(": ")
		}
		buf.WriteString(f.Underlying.Error())
	}

	return buf.String()
}

// Format implements fmt.Formatter.
func (f Failure) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			fmt.Fprintf(s, "%v\n  Info:\n", f)
			for _, info := range InfoListOf(f) {
				for k, v := range info {
					fmt.Fprintf(s, "    %s = %v\n", k, v)
				}
			}
			fmt.Fprint(s, "  CallStack:\n")
			for _, pc := range CallStackOf(f) {
				fmt.Fprintf(s, "    %+v\n", pc)
			}
		case s.Flag('#'):
			// Re-define struct to remove Format method.
			// Prevent recursive call.
			type Failure struct {
				Code       Code
				Message    string
				CallStack  CallStack
				Info       Info
				Underlying error
			}
			fmt.Fprintf(s, "%#v", Failure(f))
		default:
			io.WriteString(s, f.Error())
		}
	case 's':
		fmt.Fprintf(s, "%v", f)
	}
}

// Cause returns the underlying error.
// Use the Cause function instead of calling this method directly.
func (f Failure) Cause() error {
	return f.Underlying
}

// New returns an application error.
func New(code Code) Failure {
	return Failure{
		code,
		"",
		Callers(1),
		nil,
		nil,
	}
}

// Translate translates the error to an application error.
func Translate(err error, code Code) Failure {
	return Failure{
		code,
		"",
		Callers(1),
		nil,
		err,
	}
}

// Wrap wraps the error.
func Wrap(err error) Failure {
	return Failure{
		"",
		"",
		Callers(1),
		nil,
		err,
	}
}

// Code represents an error code.
// Define your application errors with this type.
type Code string

// CodeOf extracts Code from the error.
// If the error does not contain any code, Unknown is returned.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}

	if f, ok := err.(Failure); ok {
		if f.Code != "" {
			return f.Code
		}
		if f.Underlying != nil {
			return CodeOf(f.Underlying)
		}
	}
	return Unknown
}

// MessageOf extracts message from the error.
// If the error does not contain any messages, DefaultMessage is returned.
func MessageOf(err error) string {
	if err == nil {
		return ""
	}

	if f, ok := err.(Failure); ok {
		if f.Message != "" {
			return f.Message
		}
		if f.Underlying != nil {
			return MessageOf(f.Underlying)
		}
	}
	return DefaultMessage
}

// CallStackOf extracts call stack from the error.
// Returned call stack is for the deepest error in underlying errors.
func CallStackOf(err error) CallStack {
	if err == nil {
		return nil
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	switch e := err.(type) {
	case Failure:
		if e.Underlying != nil {
			if cs := CallStackOf(e.Underlying); cs != nil {
				return cs
			}
		}
		return e.CallStack
	case stackTracer:
		return CallStackFromPkgErrors(e.StackTrace())
	}

	return nil
}

// InfoListOf extracts infos from the error.
func InfoListOf(err error) []Info {
	if err == nil {
		return nil
	}

	if f, ok := err.(Failure); ok {
		if f.Info != nil {
			return append([]Info{f.Info}, InfoListOf(f.Underlying)...)
		}
		if f.Underlying != nil {
			return InfoListOf(f.Underlying)
		}
	}
	return nil
}

// CauseOf returns an underlying error of the error.
func CauseOf(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		c, ok := err.(causer)
		if !ok {
			break
		}
		e := c.Cause()
		if e == nil {
			break
		}
		err = e
	}

	return err
}
