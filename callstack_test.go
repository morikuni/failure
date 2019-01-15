package failure_test

import (
	"fmt"
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func X() failure.CallStack {
	return failure.Callers(0)
}

func TestCallers(t *testing.T) {
	fs := X().Frames()

	assert.Contains(t, fs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[0].File(), "callstack_test.go")
	assert.Equal(t, fs[0].Func(), "X")
	assert.Equal(t, fs[0].Line(), 13)
	assert.Equal(t, fs[0].Pkg(), "failure_test")

	assert.Contains(t, fs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[1].File(), "callstack_test.go")
	assert.Equal(t, fs[1].Func(), "TestCallers")
	assert.Equal(t, fs[1].Line(), 17)
	assert.Equal(t, fs[1].Pkg(), "failure_test")
}

func Y() error {
	return errors.New("aaa")
}

func TestCallStackFromPkgErrors(t *testing.T) {
	err := Y()

	cs, ok := failure.CallStackOf(err)
	assert.True(t, ok)
	fs := cs.Frames()

	assert.Contains(t, fs[0].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[0].File(), "callstack_test.go")
	assert.Equal(t, fs[0].Func(), "Y")
	assert.Equal(t, fs[0].Line(), 33)
	assert.Equal(t, fs[0].Pkg(), "failure_test")

	assert.Contains(t, fs[1].Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Contains(t, fs[1].File(), "callstack_test.go")
	assert.Equal(t, fs[1].Func(), "TestCallStackFromPkgErrors")
	assert.Equal(t, fs[1].Line(), 37)
	assert.Equal(t, fs[1].Pkg(), "failure_test")
}

func TestCallStack_Format(t *testing.T) {
	cs := X()

	assert.Regexp(t,
		`failure_test.X: failure_test.TestCallStack_Format: .*`,
		fmt.Sprintf("%v", cs),
	)
	assert.Regexp(t,
		`failure_test.X: failure_test.TestCallStack_Format: .*`,
		fmt.Sprintf("%s", cs),
	)
	assert.Regexp(t,
		`\[\]failure.Frame{/.+/github.com/morikuni/failure/callstack_test.go:13, /.+/github.com/morikuni/failure/callstack_test.go:57, .*}`,
		fmt.Sprintf("%#v", cs),
	)
	assert.Regexp(t,
		`\[failure_test.X\] /.+/github.com/morikuni/failure/callstack_test.go:13
\[failure_test.TestCallStack_Format\] /.+/github.com/morikuni/failure/callstack_test.go:57
\[.*`,
		fmt.Sprintf("%+v", cs),
	)
}

func TestFrame_Format(t *testing.T) {
	f := X().HeadFrame()

	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:13`,
		fmt.Sprintf("%v", f),
	)
	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:13`,
		fmt.Sprintf("%s", f),
	)
	assert.Regexp(t,
		`/.+/github.com/morikuni/failure/callstack_test.go:13`,
		fmt.Sprintf("%#v", f),
	)
	assert.Regexp(t,
		`\[failure_test.X\] /.+/github.com/morikuni/failure/callstack_test.go:13`,
		fmt.Sprintf("%+v", f),
	)
}

func TestCallStack_Frames(t *testing.T) {
	cs := X()
	fs := cs.Frames()

	assert.Equal(t, cs.Frames(), fs)

	assert.Equal(t, 13, fs[0].Line())
	assert.Equal(t, "X", fs[0].Func())

	assert.Equal(t, 101, fs[1].Line())
	assert.Equal(t, "TestCallStack_Frames", fs[1].Func())
}

func TestCallStack_HeadFrame(t *testing.T) {
	cs := X()

	assert.Equal(t, cs.Frames()[0], cs.HeadFrame())
}

func TestFrame(t *testing.T) {
	f := X().HeadFrame()

	assert.Equal(t, "X", f.Func())
	assert.Equal(t, 13, f.Line())
	assert.Equal(t, "callstack_test.go", f.File())
	assert.Contains(t, f.Path(), "github.com/morikuni/failure/callstack_test.go")
	assert.Equal(t, "failure_test", f.Pkg())
}
