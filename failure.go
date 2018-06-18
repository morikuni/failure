// Package failure is a utility package for handling an application error.
// Inspired by https://middlemost.com/failure-is-your-domain and github.com/pkg/errors.
package failure

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	// DefaultMessage is a default message for an error.
	DefaultMessage = "An internal error has occurred. Please contact the developer."

	// Unknown represents unknown error code.
	Unknown Code = "unknown"
)

// Info is information on why the error occurred.
type Info map[string]interface{}

// Failure is an error representing failure of something.
type Failure struct {
	Code       Code
	Message    string
	CallStack  CallStack
	Info       Info
	Underlying error
}

// Failure implements error.
func (e Failure) Error() string {
	buf := strings.Builder{}

	if len(e.CallStack) != 0 {
		buf.WriteString(e.CallStack[0].Func())
	}

	if e.Code != "" {
		if buf.Len() != 0 {
			buf.WriteRune('(')
			buf.WriteString(string(e.Code))
			buf.WriteRune(')')
		} else {
			buf.WriteString(string(e.Code))
		}
	}

	if e.Underlying != nil {
		if buf.Len() != 0 {
			buf.WriteString(": ")
		}
		buf.WriteString(e.Underlying.Error())
	}

	return buf.String()
}

// New returns application error.
func New(code Code, message string, info Info) error {
	return Failure{
		code,
		message,
		Callers(1),
		info,
		nil,
	}
}

// Translate translates an error to application error.
func Translate(err error, code Code, message string, info Info) error {
	return Failure{
		code,
		message,
		Callers(1),
		info,
		err,
	}
}

// WithFields adds fields to the error.
func WithFields(err error, info Info) error {
	return Failure{
		"",
		"",
		Callers(1),
		info,
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

// InfosOf extracts infos from the error.
func InfosOf(err error) []Info {
	if err == nil {
		return nil
	}

	if f, ok := err.(Failure); ok {
		if f.Info != nil {
			return append([]Info{f.Info}, InfosOf(f.Underlying)...)
		}
		if f.Underlying != nil {
			return InfosOf(f.Underlying)
		}
	}
	return nil
}
