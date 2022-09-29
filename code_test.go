package failure_test

import (
	"errors"
	"io"
	"testing"

	"github.com/morikuni/failure"
)

type CustomCode string

func (c CustomCode) ErrorCode() string {
	return string(c)
}

func TestCode(t *testing.T) {
	const (
		s failure.StringCode = "123"
		c CustomCode         = "123"

		s2 failure.StringCode = "123"
		c2 CustomCode         = "123"
	)

	shouldEqual(t, s.ErrorCode(), "123")
	shouldEqual(t, c.ErrorCode(), "123")

	shouldEqual(t, s, s2)
	shouldEqual(t, c, c2)

	shouldDiffer(t, s, c)
}

func TestIs(t *testing.T) {
	const (
		A failure.StringCode = "A"
		B failure.StringCode = "B"
	)

	errA := failure.New(A)
	errB := failure.Translate(errA, B)
	errC := failure.Wrap(errB)

	shouldEqual(t, failure.Is(errA, A), true)
	shouldEqual(t, failure.Is(errB, B), true)
	shouldEqual(t, failure.Is(errC, B), true)

	shouldEqual(t, failure.Is(errA, A, B), true)
	shouldEqual(t, failure.Is(errB, A, B), true)
	shouldEqual(t, failure.Is(errC, A, B), true)

	shouldEqual(t, failure.Is(errA, B), false)
	shouldEqual(t, failure.Is(errB, A), false)
	shouldEqual(t, failure.Is(errC, A), false)

	shouldEqual(t, failure.Is(nil, A, B), false)
	shouldEqual(t, failure.Is(io.EOF, A, B), false)
	shouldEqual(t, failure.Is(errA), false)

	shouldEqual(t, failure.Is(nil, nil), true)
	shouldEqual(t, failure.Is(errors.New("error"), nil), true)
}

func TestErrorsAs(t *testing.T) {
	const (
		A failure.StringCode = "A"
	)
	err := failure.New(A, failure.Message("foo"))
	var cs failure.CallStack
	wantCS, _ := failure.CallStackOf(err)
	shouldEqual(t, errors.As(err, &cs), true)
	shouldEqual(t, cs, wantCS)
	var code failure.Code
	shouldEqual(t, errors.As(err, &code), true)
	shouldEqual(t, code, A)
	var msg failure.Messenger
	shouldEqual(t, errors.As(err, &msg), true)
	shouldEqual(t, msg, failure.Message("foo"))
}
