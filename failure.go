// Package failure provides an error represented as error code and
// extensible error interface with wrappers.
package failure

import (
	"fmt"
	"strings"
)

// Failure represents an error with error code.
type Failure struct {
	code       Code
	underlying error
}

// UnwrapError returns the underlying error.
// It also implements the ErrorWrapper interface.
func (f Failure) UnwrapError() error {
	return f.underlying
}

// GetCode returns the error code of the error.
func (f Failure) GetCode() Code {
	return f.code
}

// Error implements the error interface.
func (f Failure) Error() string {
	msg := fmt.Sprintf("code(%s)", f.code.ErrorCode())
	if f.underlying != nil {
		msg = strings.Join([]string{msg, f.underlying.Error()}, ": ")
	}
	return msg
}

// CodeOf extracts an error Code from the error.
func CodeOf(err error) Code {
	if err == nil {
		return nil
	}

	type codeGetter interface {
		GetCode() Code
	}

	i := NewIterator(err)
	for i.Next() {
		err := i.Error()
		if g, ok := err.(codeGetter); ok {
			return g.GetCode()
		}
	}

	return nil
}

// New creates a Failure from error Code.
func New(code Code, wrappers ...Wrapper) error {
	return newFailure(nil, code, wrappers)
}

// Translate translates err to an error with given code.
// It wraps the error with given wrappers, and automatically
// add call stack and formatter.
func Translate(err error, code Code, wrappers ...Wrapper) error {
	return newFailure(err, code, wrappers)
}

// Wrap wraps err with given wrappers, and automatically add
// call stack and formatter.
func Wrap(err error, wrappers ...Wrapper) error {
	return Custom(err, append(wrappers, WithCallStackSkip(1), WithFormatter())...)
}

func newFailure(err error, code Code, wrappers []Wrapper) error {
	f := Failure{
		code,
		err,
	}
	return Custom(f, append(wrappers, WithCallStackSkip(2), WithFormatter())...)
}

// Custom is the general error wrapping constructor.
// It just wraps err with given wrappers.
func Custom(err error, wrappers ...Wrapper) error {
	if err == nil {
		return nil
	}
	for _, w := range wrappers {
		err = w.WrapError(err)
	}
	return err
}
