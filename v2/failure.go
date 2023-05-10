package failure

import (
	"errors"
	"io"
)

// Code represents error codes. Any comparable types can be used as error codes.
type Code comparable

// Field is an interface that can be implemented by custom error fields.
type Field interface {
	SetErrorField(setter FieldSetter)
}

// FieldSetter is an interface used by Field implementations to set their error
// information.
type FieldSetter interface {
	Set(key, value any)
}

// ErrorFormatter is an interface that formats their error information to an
// io.Writer. Implement this interface to customize the error output for your
// custom error fields.
type ErrorFormatter interface {
	FormatError(io.Writer)
}

// New creates a new error with the provided error code.
func New[C Code](c C, fields ...Field) error {
	return newStack(nil, c, fields)
}

// Translate takes an existing error and returns a new error with the translated
// error code. Use this function to replace the error code of an existing error
// with a new one.
func Translate[C Code](err error, c C, fields ...Field) error {
	return newStack(err, c, fields)
}

// Convert takes an existing error and returns a new error with the provided
// error code. Use this function to replace the existing error with a new one.
// This function is similar to Translate, but doesn't unwrap the original error.
func Convert[C Code](err error, c C, fields ...Field) error {
	return newStack(opaque{err}, c, fields)
}

// Wrap takes an existing and returns a new error. Use this function to add
// context to an existing error without changing its error code.
func Wrap(err error, fields ...Field) error {
	if err == nil {
		return nil
	}
	return newStack(err, nil, fields)
}

// Error creates an error with the provided text. It is recommended to use New
// with error codes whenever possible, and reserve the use of Error for
// a unexpected situation.
func Error(text string, fields ...Field) error {
	return newStack(errors.New(text), nil, fields)
}

// Opaque creates an error that cannot be unwrapped using the standard Unwrap
// method. Use this function when you want to prevent propagating data like error
// codes or context to the caller. However, using ForceUnwrap will still allow
// retrieving the original error.
func Opaque(err error, fields ...Field) error {
	return newStack(opaque{err}, nil, fields)
}

// Unexpected is the alias of Error.
func Unexpected(text string, fields ...Field) error {
	return newStack(errors.New(text), nil, fields)
}

// MarkUnexpected is the alias of Opaque.
func MarkUnexpected(err error, fields ...Field) error {
	return newStack(opaque{err}, nil, fields)
}

type opaque struct {
	error
}

func (opaque) Unwrap() error {
	return nil
}

func (o opaque) ForceUnwrap() error {
	return o.error
}

func newStack(err error, code any, fields []Field) error {
	var defaultFields []Field
	if code == nil {
		defaultFields = []Field{
			Callers(2),
		}
	} else {
		defaultFields = []Field{
			codeField{code},
			Callers(2),
		}
	}
	return NewStack(err, defaultFields, fields)
}
