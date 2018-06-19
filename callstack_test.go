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
	cs := X()

	assert.Contains(t, cs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[0].Func(), "X")
	assert.Equal(t, cs[0].Line(), 14)
	assert.Equal(t, cs[0].Pkg(), "failure_test")

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "TestCallers")
	assert.Equal(t, cs[1].Line(), 18)
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
	assert.Equal(t, cs[0].Line(), 32)
	assert.Equal(t, cs[0].Pkg(), "failure_test")

	assert.Contains(t, cs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, cs[1].Func(), "TestCallStackFromPkgErrors")
	assert.Equal(t, cs[1].Line(), 40)
	assert.Equal(t, cs[1].Pkg(), "failure_test")
}

func TestFormat(t *testing.T) {
	cs := X()[:2]

	assert.Equal(t,
		`X: TestFormat`,
		fmt.Sprintf("%v", cs),
	)
	assert.Equal(t,
		`X: TestFormat`,
		fmt.Sprintf("%s", cs),
	)
	assert.Regexp(t,
		`\[\]failure.PC{/.+/github.com/morikuni/failure/callstack_test.go:14, /.+/github.com/morikuni/failure/callstack_test.go:58}`,
		fmt.Sprintf("%#v", cs),
	)
	assert.Regexp(t,
		`\[X\] /.+/github.com/morikuni/failure/callstack_test.go:14
\[TestFormat\] /.+/github.com/morikuni/failure/callstack_test.go:58`,
		fmt.Sprintf("%+v", cs),
	)

	pc := cs[0]

	assert.Equal(t,
		`X`,
		fmt.Sprintf("%v", pc),
	)
	assert.Equal(t,
		`X`,
		fmt.Sprintf("%s", pc),
	)
	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%#v", pc),
	)
	assert.Regexp(t,
		`\[X\] /.+/github.com/morikuni/failure/callstack_test.go:14`,
		fmt.Sprintf("%+v", pc),
	)
}
