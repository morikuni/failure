package failure_test

import (
	"errors"
	"io"
	"testing"

	"github.com/morikuni/failure"
)

type errorsAsDelegateError struct {
	err error
}

func (errorsAsDelegateError) Error() string {
	return "errors.As delegate error"
}

func (err errorsAsDelegateError) As(target interface{}) bool {
	return errors.As(err.err, target)
}

func TestErrorsAs(t *testing.T) {
	const (
		A failure.StringCode = "A"
	)
	err := failure.Translate(errorsAsDelegateError{io.EOF}, A, failure.Message("foo"))
	var cs failure.CallStack
	wantCS, ok := failure.CallStackOf(err)
	shouldEqual(t, ok, true)
	shouldEqual(t, errors.As(err, &cs), true)
	shouldEqual(t, cs, wantCS)
	var code failure.Code
	shouldEqual(t, errors.As(err, &code), true)
	shouldEqual(t, code, A)
	var msg failure.Messenger
	shouldEqual(t, errors.As(err, &msg), true)
	shouldEqual(t, msg, failure.Message("foo"))
	var tracer failure.StringTracer
	var tracerInterface failure.Tracer = &tracer
	shouldEqual(t, errors.As(err, &tracerInterface), true)
	shouldEqual(t, len(tracer), 1)

	code, ok = failure.CodeOf(err)
	shouldEqual(t, ok, true)
	shouldEqual(t, code, A)
	m, ok := failure.MessageOf(err)
	shouldEqual(t, ok, true)
	shouldEqual(t, m, "foo")
	tracer = nil
	failure.Trace(err, &tracer)
	shouldEqual(t, len(tracer), 3)
}
