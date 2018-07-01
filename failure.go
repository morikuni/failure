// Package failure is a utility package for handling application errors.
// Inspired by https://middlemost.com/failure-is-your-domain and github.com/pkg/errors.
package failure

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

var (
	// DefaultMessage is a default message for an error.
	DefaultMessage = "An internal error has occurred. Please try it later or contact the developer."

	// Unknown represents unknown error code.
	Unknown Code = StringCode("unknown")
)

// Info is key-value data.
type Info map[string]interface{}

// Failure is an error representing failure of something.
type Failure struct {
	// code is an error code to handle the error in your source code.
	code Code
	// message is an error message for the application user.
	// So the message should be human-readable and be helpful.
	message string
	// callStack is a call stack at the time of the error occurred.
	callStack CallStack
	// Info is optional information on why the error occurred.
	info Info
	// Underlying is an underlying error of the failure.
	underlying error
}

// WithMessage attaches a message to the failure.
func (f *Failure) WithMessage(message string) *Failure {
	f.message = message
	return f
}

// WithInfo attaches information to the failure.
func (f *Failure) WithInfo(info Info) *Failure {
	f.info = info
	return f
}

// Failure implements error.
func (f *Failure) Error() string {
	buf := &bytes.Buffer{}

	if f.callStack != nil {
		buf.WriteString(f.callStack.HeadFrame().Func())
	}

	if f.code != nil {
		if buf.Len() != 0 {
			buf.WriteRune('(')
			buf.WriteString(f.code.ErrorCode())
			buf.WriteRune(')')
		} else {
			buf.WriteString(f.code.ErrorCode())
		}
	}

	if f.underlying != nil {
		if buf.Len() != 0 {
			buf.WriteString(": ")
		}
		buf.WriteString(f.underlying.Error())
	}

	return buf.String()
}

// Format implements fmt.Formatter.
func (f *Failure) Format(s fmt.State, verb rune) {
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
			for _, f := range CallStackOf(f).Frames() {
				fmt.Fprintf(s, "    %+v\n", f)
			}
		case s.Flag('#'):
			// Re-define struct to remove Format method.
			// Prevent recursive call.
			type Failure struct {
				code       Code
				message    string
				callStack  CallStack
				info       Info
				underlying error
			}
			fmt.Fprintf(s, "%#v", Failure(*f))
		default:
			s.Write([]byte(f.Error()))
		}
	case 's':
		fmt.Fprintf(s, "%v", f)
	}
}

// Cause returns the underlying error.
// Use the Cause function instead of calling this method directly.
func (f *Failure) Cause() error {
	return f.underlying
}

// New returns an application error.
func New(code Code) *Failure {
	return &Failure{
		code:       code,
		message:    "",
		callStack:  Callers(1),
		info:       nil,
		underlying: nil,
	}
}

// Translate translates the error to an application error.
func Translate(err error, code Code) *Failure {
	return &Failure{
		code:       code,
		message:    "",
		callStack:  Callers(1),
		info:       nil,
		underlying: err,
	}
}

// Wrap wraps the error.
func Wrap(err error) *Failure {
	return &Failure{
		code:       nil,
		message:    "",
		callStack:  Callers(1),
		info:       nil,
		underlying: err,
	}
}

// CodeOf extracts Code from the error.
// If the error does not contain any code, Unknown is returned.
func CodeOf(err error) Code {
	if err == nil {
		return nil
	}

	if f, ok := err.(*Failure); ok {
		if f.code != nil {
			return f.code
		}
		if f.underlying != nil {
			return CodeOf(f.underlying)
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

	if f, ok := err.(*Failure); ok {
		if f.message != "" {
			return f.message
		}
		if f.underlying != nil {
			return MessageOf(f.underlying)
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
	case *Failure:
		if e.underlying != nil {
			if cs := CallStackOf(e.underlying); cs != nil {
				return cs
			}
		}
		return e.callStack
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

	if f, ok := err.(*Failure); ok {
		if f.info != nil {
			return append([]Info{f.info}, InfoListOf(f.underlying)...)
		}
		if f.underlying != nil {
			return InfoListOf(f.underlying)
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
