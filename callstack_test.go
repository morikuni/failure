package failure_test

import (
	"fmt"
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
	fs := X().Frames()

	assert.Contains(t, fs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[0].File(), "callstack_test.go")
	assert.Equal(t, fs[0].Func(), "X")
	assert.Equal(t, fs[0].Line(), 14)
	assert.Equal(t, fs[0].Pkg(), "failure_test")

	assert.Contains(t, fs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[1].File(), "callstack_test.go")
	assert.Equal(t, fs[1].Func(), "TestCallers")
	assert.Equal(t, fs[1].Line(), 18)
	assert.Equal(t, fs[1].Pkg(), "failure_test")
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

	fs := failure.CallStackFromPkgErrors(st.StackTrace()).Frames()

	assert.Contains(t, fs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[0].File(), "callstack_test.go")
	assert.Equal(t, fs[0].Func(), "Y")
	assert.Equal(t, fs[0].Line(), 34)
	assert.Equal(t, fs[0].Pkg(), "failure_test")

	assert.Contains(t, fs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[1].File(), "callstack_test.go")
	assert.Equal(t, fs[1].Func(), "TestCallStackFromPkgErrors")
	assert.Equal(t, fs[1].Line(), 42)
	assert.Equal(t, fs[1].Pkg(), "failure_test")
}

func TestFormat(t *testing.T) {
	cs := X()

	assert.Regexp(t,
		`X: TestFormat: .*`,
		fmt.Sprintf("%v", cs),
	)
	assert.Regexp(t,
		`X: TestFormat: .*`,
		fmt.Sprintf("%s", cs),
	)
	assert.Regexp(t,
		`\[\]failure.Frame{/.+/github.com/morikuni/failure/callstack_test.go:14, /.+/github.com/morikuni/failure/callstack_test.go:62, .*}`,
		fmt.Sprintf("%#v", cs),
	)
	assert.Regexp(t,
		`\[X\] /.+/github.com/morikuni/failure/callstack_test.go:14
\[TestFormat\] /.+/github.com/morikuni/failure/callstack_test.go:62
\[.*`,
		fmt.Sprintf("%+v", cs),
	)

	f := cs.Frames()[0]

	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%v", f),
	)
	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%s", f),
	)
	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%#v", f),
	)
	assert.Regexp(t,
		`\[X\] /.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%+v", f),
	)
}
