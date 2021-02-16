// Package failure provides an error represented as error code and
// extensible error interface with wrappers.
package failure

import (
	"fmt"
)

// Failure represents an error with error code.
// Deprecated: This interface will be deleted in v1.0.0 release.
type Failure interface {
	error
	GetCode() Code
}

// CodeOf extracts an error code from the err.
// The first error code found in the err is returned, but if Unexpected
// interface is detected before the code is found, it behaves as if
// there is no code found.
func CodeOf(err error) (Code, bool) {
	if err == nil {
		return nil, false
	}

	i := NewIterator(err)
	for i.Next() {
		if v, ok := i.Error().(interface{ Unexpected() bool }); ok && v.Unexpected() {
			return nil, false
		}

		var c Code
		if i.As(&c) {
			return c, true
		}
	}

	return nil, false
}

// New creates an error from error code.
func New(code Code, wrappers ...Wrapper) error {
	return Custom(Custom(&withCode{code: code}, wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Translate translates the err to an error with given code.
// It wraps the error with given wrappers, and automatically
// add call stack and formatter.
func Translate(err error, code Code, wrappers ...Wrapper) error {
	return Custom(Custom(Custom(err, WithCode(code)), wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Wrap wraps err with given wrappers, and automatically add
// call stack and formatter.
func Wrap(err error, wrappers ...Wrapper) error {
	return Custom(Custom(err, wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// MarkUnexpected wraps err and preventing propagation of error code from underlying error.
// It is used where an error can be returned but expecting it does not happen.
// The returned error does not return error code from function CodeOf.
func MarkUnexpected(err error, wrappers ...Wrapper) error {
	return Custom(Custom(Custom(err, WithUnexpected()), wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// Custom is the general error wrapping constructor.
// It just wraps err with given wrappers.
func Custom(err error, wrappers ...Wrapper) error {
	if err == nil {
		return nil
	}
	// To process from left to right, iterate from the last one.
	// Custom(errors.New("foo"), Message("aaa"), Message("bbb")) should be "aaa: bbb: foo".
	for i := len(wrappers) - 1; i >= 0; i-- {
		err = wrappers[i].WrapError(err)
	}
	return err
}

type unexpected string

func (e unexpected) Error() string {
	return string(e)
}

func (e unexpected) Unexpected() bool {
	return true
}

func (e unexpected) As(x interface{}) bool {
	switch t := x.(type) {
	case Tracer:
		t.Push(e)
		return true
	default:
		return false
	}
}

// Unexpected creates an error from message without error code.
// The returned error should be kind of internal or unknown error.
func Unexpected(msg string, wrappers ...Wrapper) error {
	return Custom(Custom(unexpected(msg), wrappers...), WithFormatter(), WithCallStackSkip(1))
}

// NewFailure returns Failure without any wrappers.
// You don't have to use this directly, unless using function Custom.
// Basically, you can use function New instead of this.
// Deprecated: This function will be deleted in v1.0.0 release. Please use New.
func NewFailure(code Code) Failure {
	return &withCode{code: code}
}

// WithCode appends code to an error.
// You don't have to use this directly, unless using function Custom.
// Basically, you can use function Translate instead of this.
func WithCode(code Code) Wrapper {
	return WrapperFunc(func(err error) error {
		return &withCode{code, err}
	})
}

type withCode struct {
	code       Code
	underlying error
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use Unwrap.
func (w *withCode) UnwrapError() error {
	return w.Unwrap()
}

func (w *withCode) Unwrap() error {
	return w.underlying
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use As method on Iterator.
func (w *withCode) GetCode() Code {
	return w.code
}

func (w *withCode) As(x interface{}) bool {
	switch t := x.(type) {
	case *Code:
		*t = w.code
		return true
	case Tracer:
		t.Push(w.code)
		return true
	default:
		return false
	}
}

func (w *withCode) Error() string {
	if w.underlying == nil {
		return fmt.Sprintf("code(%s)", w.code.ErrorCode())
	}
	return fmt.Sprintf("code(%s): %s", w.code.ErrorCode(), w.underlying)
}

// WithoutCode prevents propagation of error code from underlying error
// You don't have to use this directly, unless using function Custom.
// Basically, you can use function MarkUnexpected instead of this.
// Deprecated: This function will be deleted in v1.0.0 release. Please use WithUnexpected.
func WithoutCode() Wrapper {
	return WrapperFunc(func(err error) error {
		return &withoutCode{err}
	})
}

type withoutCode struct {
	underlying error
}

func (w *withoutCode) Unwrap() error {
	return w.underlying
}

func (w *withoutCode) Error() string {
	return fmt.Sprintf("code_eliminated: %s", w.underlying)
}

func (*withoutCode) NoCode() bool {
	return true
}
