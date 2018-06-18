package failure

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func X() CallStack {
	return Callers(0)
}

func TestCallers(t *testing.T) {
	cs := X()

	assert.Contains(t, cs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[0].Func(), "failure.X")
	assert.Equal(t, cs[0].Line(), 12)

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "failure.TestCallers")
	assert.Equal(t, cs[1].Line(), 16)
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

	cs := CallStackFromPkgErrors(st.StackTrace())

	assert.Contains(t, cs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[0].Func(), "failure.Y")
	assert.Equal(t, cs[0].Line(), 28)

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "failure.TestCallStackFromPkgErrors")
	assert.Equal(t, cs[1].Line(), 36)
}
