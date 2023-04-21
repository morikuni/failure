package failure_test

import (
	"errors"
	"testing"

	"github.com/morikuni/failure/v2"
)

func TestStack_Error(t *testing.T) {
	st1 := failure.NewStack(nil, []failure.Field{failure.Callers(0), failure.Message("test1")})
	st2 := failure.NewStack(st1, []failure.Field{failure.Message("test2")}, []failure.Field{failure.WithCode("code"), failure.Context{"a": "b"}})

	equal(t, st1.Error(), "v2_test.TestStack_Error[test1]")
	equal(t, st2.Error(), "(code=code)[test2, {a=b}]: v2_test.TestStack_Error[test1]")
}

func TestStack_Unwrap(t *testing.T) {
	err := errors.New("test")
	wrap := failure.NewStack(err)
	noWrap := failure.NewStack(nil, []failure.Field{failure.Message("aaa")})

	equal(t, wrap.Unwrap(), err)
	equal(t, noWrap.Unwrap(), nil)
}

func TestStack_As(t *testing.T) {
	st := failure.NewStack(nil, []failure.Field{failure.Message("aaa"), failure.Context{"a": "b"}})

	var msg failure.Message
	equal(t, st.As(&msg), true)
	equal(t, msg, failure.Message("aaa"))

	var ctx failure.Context
	equal(t, st.As(&ctx), true)
	equal(t, ctx, failure.Context{"a": "b"})

	var cs failure.CallStack
	equal(t, st.As(&cs), false)
	equal(t, cs, failure.CallStack(nil))

	var f failure.Field
	equal(t, st.As(&f), true)
	equal(t, f, failure.Messagef("aaa"))

	var err error
	equal(t, st.As(&err), false)
	equal(t, err, nil)
}

func TestStack_Value(t *testing.T) {
	st := failure.NewStack(nil, []failure.Field{failure.Message("aaa"), failure.Context{"a": "b"}})

	equal(t, st.Value(failure.KeyMessage), failure.Messagef("aaa"))
	equal(t, st.Value(failure.KeyContext), failure.Context{"a": "b"})
	equal(t, st.Value(failure.KeyCode), nil)
}
