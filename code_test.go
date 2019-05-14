package failure_test

import (
	"errors"
	"testing"

	"io"

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
