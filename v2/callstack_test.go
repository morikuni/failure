package failure_test

import (
	"fmt"
	"testing"

	"github.com/morikuni/failure/v2"
)

func X() failure.CallStack {
	return failure.Callers(0)
}

func TestCallers(t *testing.T) {
	fs := X().Frames()

	contain(t, fs[0].Path(), "/failure/v2/callstack_test.go")
	contain(t, fs[0].File(), "callstack_test.go")
	equal(t, fs[0].Func(), "X")
	equal(t, fs[0].Line(), 11)
	equal(t, fs[0].Pkg(), "v2_test") // is this bug?
	equal(t, fs[0].PkgPath(), "github.com/morikuni/failure/v2_test")

	contain(t, fs[1].Path(), "/failure/v2/callstack_test.go")
	contain(t, fs[1].File(), "callstack_test.go")
	equal(t, fs[1].Func(), "TestCallers")
	equal(t, fs[1].Line(), 15)
	equal(t, fs[1].Pkg(), "v2_test")
	equal(t, fs[1].PkgPath(), "github.com/morikuni/failure/v2_test")
}

func TestCallStack_Format(t *testing.T) {
	cs := X()

	match(t,
		fmt.Sprintf("%v", cs),
		`v2_test.X: v2_test.TestCallStack_Format: .*`,
	)
	match(t,
		fmt.Sprintf("%s", cs),
		`v2_test.X: v2_test.TestCallStack_Format: .*`,
	)
	match(t,
		fmt.Sprintf("%#v", cs),
		`\[\]failure.Frame{/.+/failure/v2/callstack_test.go:11, /.+/failure/v2/callstack_test.go:33, .*}`,
	)
	match(t,
		fmt.Sprintf("%+v", cs),
		`\[v2_test.X\] /.+/failure/v2/callstack_test.go:11
\[v2_test.TestCallStack_Format\] /.+/failure/v2/callstack_test.go:33
\[.*`,
	)
}

func TestFrame_Format(t *testing.T) {
	f := X().HeadFrame()

	match(t,
		fmt.Sprintf("%v", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	match(t,
		fmt.Sprintf("%s", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	match(t,
		fmt.Sprintf("%#v", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	match(t,
		fmt.Sprintf("%+v", f),
		`\[v2_test.X\] /.+/failure/v2/callstack_test.go:11`,
	)
}

func TestCallStack_Frames(t *testing.T) {
	cs := X()
	fs := cs.Frames()

	equal(t, cs.Frames(), fs)

	equal(t, fs[0].Line(), 11)
	equal(t, fs[0].Func(), "X")

	equal(t, fs[1].Line(), 77)
	equal(t, fs[1].Func(), "TestCallStack_Frames")
}

func TestCallStack_HeadFrame(t *testing.T) {
	cs := X()

	equal(t, cs.Frames()[0], cs.HeadFrame())
}

func TestFrame(t *testing.T) {
	f := func() failure.Frame {
		return func() failure.Frame {
			return failure.Callers(0).HeadFrame()
		}()
	}()

	equal(t, f.Func(), "TestFrame.func1.1")
	equal(t, f.Line(), 98)
	equal(t, f.File(), "callstack_test.go")
	contain(t, f.Path(), "/failure/v2/callstack_test.go")
	equal(t, f.Pkg(), "v2_test")
	equal(t, f.PkgPath(), "github.com/morikuni/failure/v2_test")
}
