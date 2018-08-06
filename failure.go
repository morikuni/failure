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
	Unknown Code = StringCode("unknown")
)

// Failure represents an application error with error code.
// The package failure provides some constructor functions, but you can create
// your customized constructor functions, e.g. make sure to fill in code and message.
type Failure struct {
	// Code is an error code represents what happened in application.
	// Define error code when you want to distinguish errors. It is when you
	// write if statement.
	Code Code
	// Message is an error message for the application users.
	// So the message should be human-readable and be helpful.
	// Do not put a system error message here.
	Message string
	// CallStack is a call stack when the error occurred.
	// You can get where the error occurred, e.g. file name, function name etc,
	// from head frame of the call stack.
	CallStack CallStack
	// Info is optional information of the error.
	// Put a system error message and debug information here, then write them
	// to logs.
	Info Info
	// Underlying is an underlying error of the failure.
	// The failure is chained by this field.
	Underlying error
}

// Error implements the interface error.
// This returns colon-separated errors.
// The failure is represented as `function_name(error_code)`.
func (f Failure) Error() string {
	buf := &bytes.Buffer{}

	if f.CallStack != nil {
		buf.WriteString(f.CallStack.HeadFrame().Func())
	}

	if f.Code != nil {
		if buf.Len() != 0 {
			buf.WriteRune('(')
			buf.WriteString(string(f.Code.ErrorCode()))
			buf.WriteRune(')')
		} else {
			buf.WriteString(string(f.Code.ErrorCode()))
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

// Format implements the interface fmt.Formatter.
// %s, %v: same as Error().
// %+v: %v + list of entire info + most underlying error's stack trace.
// %#v: Go's representation of the failure struct.
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
			if cs := CallStackOf(f); cs != nil {
				for _, f := range cs.Frames() {
					fmt.Fprintf(s, "    %+v\n", f)
				}
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
		io.WriteString(s, f.Error())
	}
}

// Cause returns the underlying error.
// If you want a most underlying error, using a function CauseOf
// is recommended.
func (f Failure) Cause() error {
	return f.Underlying
}

// New creates a failure from error code.
func New(code Code, opts ...Option) error {
	return newFailure(nil, code, opts)
}

// Translate translates the error to an application error indicated
// by error code.
func Translate(err error, code Code, opts ...Option) error {
	return newFailure(err, code, opts)
}

// Wrap wraps the err without error code.
func Wrap(err error, opts ...Option) error {
	if err == nil {
		return nil
	}
	return newFailure(err, nil, opts)
}

func newFailure(err error, code Code, opts []Option) Failure {
	f := Failure{
		code,
		"",
		Callers(2),
		nil,
		err,
	}
	for _, o := range opts {
		o.ApplyTo(&f)
	}
	return f
}

// CodeOf extracts an error code from the error.
// If the error does not include any error codes, the value of
// variable Unknown is returned.
func CodeOf(err error) Code {
	if err == nil {
		return nil
	}

	if f, ok := err.(Failure); ok {
		if f.Code != nil {
			return f.Code
		}
		if f.Underlying != nil {
			return CodeOf(f.Underlying)
		}
	}
	return Unknown
}

// MessageOf extracts the message from the error.
// If the error does not include any messages, the value of
// variable DefaultMessage is returned.
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
// Returned call stack is for the most underlying error.
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
		return callStackFromPkgErrors(e.StackTrace())
	}

	return nil
}

// InfoListOf extracts list of information from the error.
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

// CauseOf returns the most underlying error from the error.
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
