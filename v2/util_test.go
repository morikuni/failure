package failure_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/morikuni/failure/v2"
)

func TestValue(t *testing.T) {
	err := errors.New("error")
	err = failure.NewStack(err, []failure.Field{failure.Message("1"), failure.Context{"a": "b"}})
	err = fmt.Errorf("fmt: %w", err)
	err = failure.NewStack(err, []failure.Field{failure.Callers(0), failure.Message("2")})

	equal(t, failure.Value(err, failure.KeyMessage), failure.Message("2"))
	equal(t, failure.Value(err, failure.KeyContext), failure.Context{"a": "b"})
	equal(t, failure.Value(err, failure.KeyCode), nil)
}

func TestCauseValue(t *testing.T) {
	err := errors.New("error")
	err = failure.NewStack(err, []failure.Field{failure.Message("1"), failure.Context{"a": "b"}})
	err = fmt.Errorf("fmt: %w", err)
	err = failure.NewStack(err, []failure.Field{failure.Callers(0), failure.Message("2")})

	equal(t, failure.CauseValue(err, failure.KeyMessage), failure.Message("1"))
	equal(t, failure.CauseValue(err, failure.KeyContext), failure.Context{"a": "b"})
	equal(t, failure.CauseValue(err, failure.KeyCode), nil)
}

func TestIs(t *testing.T) {
	err := errors.New("error")
	err = failure.NewStack(err, []failure.Field{failure.WithCode(1)})
	err = fmt.Errorf("fmt: %w", err)
	err = failure.NewStack(err, []failure.Field{failure.WithCode(2)})

	equal(t, failure.Is(err, 1), false)
	equal(t, failure.Is(err, 2), true)
}

func TestCodeOf(t *testing.T) {
	err1 := errors.New("error")
	err2 := failure.NewStack(err1, []failure.Field{failure.WithCode(1)})
	err3 := failure.NewStack(err2, []failure.Field{failure.WithCode(2)})
	err4 := fmt.Errorf("fmt: %w", err3)

	equal(t, failure.CodeOf(err1), nil)
	equal(t, failure.CodeOf(err2), 1)
	equal(t, failure.CodeOf(err3), 2)
	equal(t, failure.CodeOf(err4), 2)
}

func TestMessageOf(t *testing.T) {
	err1 := errors.New("error")
	err2 := failure.NewStack(err1, []failure.Field{failure.Message("1")})
	err3 := failure.NewStack(err2, []failure.Field{failure.Message("2")})
	err4 := fmt.Errorf("fmt: %w", err3)

	equal(t, failure.MessageOf(err1), failure.Message(""))
	equal(t, failure.MessageOf(err2), failure.Message("1"))
	equal(t, failure.MessageOf(err3), failure.Message("2"))
	equal(t, failure.MessageOf(err4), failure.Message("2"))
}

func TestCallStackOf(t *testing.T) {
	baseLine := failure.Callers(0).HeadFrame().Line()
	err1 := errors.New("error")
	err2 := failure.NewStack(err1, []failure.Field{failure.Callers(0)})
	err3 := failure.NewStack(err2, []failure.Field{failure.Callers(0)})
	err4 := fmt.Errorf("fmt: %w", err3)

	equal(t, failure.CallStackOf(err1), failure.CallStack(nil))
	equal(t, failure.CallStackOf(err2).HeadFrame().Line(), baseLine+2)
	equal(t, failure.CallStackOf(err3).HeadFrame().Line(), baseLine+2)
	equal(t, failure.CallStackOf(err4).HeadFrame().Line(), baseLine+2)
}

func TestCauseOf(t *testing.T) {
	err1 := errors.New("error")
	err2 := failure.NewStack(err1, []failure.Field{failure.WithCode(1)})
	err3 := failure.NewStack(err2, []failure.Field{failure.WithCode(2)})
	err4 := fmt.Errorf("fmt: %w", err3)

	equal(t, failure.CauseOf(err1), err1)
	equal(t, failure.CauseOf(err2), err1)
	equal(t, failure.CauseOf(err3), err1)
	equal(t, failure.CauseOf(err4), err1)
}
