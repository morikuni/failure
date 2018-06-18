package failure_test

import (
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func X() failure.CallStack {
	return failure.Callers(0)
}

func TestCallers(t *testing.T) {
	cs := X()

	assert.Contains(t, cs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[0].Func(), "X")
	assert.Equal(t, cs[0].Line(), 13)
	assert.Equal(t, cs[0].Pkg(), "failure_test")

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "TestCallers")
	assert.Equal(t, cs[1].Line(), 17)
	assert.Equal(t, cs[1].Pkg(), "failure_test")
}

func Y() error {
	return errors.New("aaa")
}

func TestCallStackFromPkgErrors(t *testing.T) {
	type StackTracer interface {
		StackTrace() errors.StackTrace
	}

	err := Y()
	st, ok := err.(StackTracer)
	require.True(t, ok)

	cs := failure.CallStackFromPkgErrors(st.StackTrace())

	assert.Contains(t, cs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[0].Func(), "Y")
	assert.Equal(t, cs[0].Line(), 31)
	assert.Equal(t, cs[0].Pkg(), "failure_test")

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "TestCallStackFromPkgErrors")
	assert.Equal(t, cs[1].Line(), 39)
	assert.Equal(t, cs[1].Pkg(), "failure_test")
}
