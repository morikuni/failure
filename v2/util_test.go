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

func TestOriginValue(t *testing.T) {
	err := errors.New("error")
	err = failure.NewStack(err, []failure.Field{failure.Message("1"), failure.Context{"a": "b"}})
	err = fmt.Errorf("fmt: %w", err)
	err = failure.NewStack(err, []failure.Field{failure.Callers(0), failure.Message("2")})

	equal(t, failure.OriginValue(err, failure.KeyMessage), failure.Message("1"))
	equal(t, failure.OriginValue(err, failure.KeyContext), failure.Context{"a": "b"})
	equal(t, failure.OriginValue(err, failure.KeyCode), nil)
}

func TestIs(t *testing.T) {
	err := errors.New("error")
	err = failure.NewStack(err, []failure.Field{failure.WithCode(1)})
	err = fmt.Errorf("fmt: %w", err)
	err = failure.NewStack(err, []failure.Field{failure.WithCode(2)})

	equal(t, failure.Is(err, 1), false)
	equal(t, failure.Is(err, 2), true)
}
