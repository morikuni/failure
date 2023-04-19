package failure_test

import (
	"fmt"
	"testing"

	"github.com/morikuni/failure"
)

func X() failure.CallStack {
	return failure.Callers(0)
}

func TestCallers(t *testing.T) {
	fs := X().Frames()

	shouldContain(t, fs[0].Path(), "/failure/callstack_test.go")
	shouldContain(t, fs[0].File(), "callstack_test.go")
	shouldEqual(t, fs[0].Func(), "X")
	shouldEqual(t, fs[0].Line(), 11)
	shouldEqual(t, fs[0].Pkg(), "failure_test")
	shouldEqual(t, fs[0].PkgPath(), "github.com/morikuni/failure_test")

	shouldContain(t, fs[1].Path(), "/failure/callstack_test.go")
	shouldContain(t, fs[1].File(), "callstack_test.go")
	shouldEqual(t, fs[1].Func(), "TestCallers")
	shouldEqual(t, fs[1].Line(), 15)
	shouldEqual(t, fs[1].Pkg(), "failure_test")
	shouldEqual(t, fs[1].PkgPath(), "github.com/morikuni/failure_test")
}

func TestCallStack_Format(t *testing.T) {
	cs := X()

	shouldMatch(t,
		fmt.Sprintf("%v", cs),
		`failure_test.X: failure_test.TestCallStack_Format: .*`,
	)
	shouldMatch(t,
		fmt.Sprintf("%s", cs),
		`failure_test.X: failure_test.TestCallStack_Format: .*`,
	)
	shouldMatch(t,
		fmt.Sprintf("%#v", cs),
		`\[\]failure.Frame{/.+/failure/callstack_test.go:11, /.+/failure/callstack_test.go:33, .*}`,
	)
	shouldMatch(t,
		fmt.Sprintf("%+v", cs),
		`\[failure_test.X\] /.+/failure/callstack_test.go:11
\[failure_test.TestCallStack_Format\] /.+/failure/callstack_test.go:33
\[.*`,
	)
}

func TestFrame_Format(t *testing.T) {
	f := X().HeadFrame()

	shouldMatch(t,
		fmt.Sprintf("%v", f),
		`/.+/failure/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%s", f),
		`/.+/failure/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%#v", f),
		`/.+/failure/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%+v", f),
		`\[failure_test.X\] /.+/failure/callstack_test.go:11`,
	)
}

func TestCallStack_Frames(t *testing.T) {
	cs := X()
	fs := cs.Frames()

	shouldEqual(t, cs.Frames(), fs)

	shouldEqual(t, fs[0].Line(), 11)
	shouldEqual(t, fs[0].Func(), "X")

	shouldEqual(t, fs[1].Line(), 77)
	shouldEqual(t, fs[1].Func(), "TestCallStack_Frames")
}

func TestCallStack_HeadFrame(t *testing.T) {
	cs := X()

	shouldEqual(t, cs.Frames()[0], cs.HeadFrame())
}

func TestFrame(t *testing.T) {
	f := func() failure.Frame {
		return func() failure.Frame {
			return failure.Callers(0).HeadFrame()
		}()
	}()

	shouldEqual(t, f.Func(), "TestFrame.func1.1")
	shouldEqual(t, f.Line(), 98)
	shouldEqual(t, f.File(), "callstack_test.go")
	shouldContain(t, f.Path(), "/failure/callstack_test.go")
	shouldEqual(t, f.Pkg(), "failure_test")
	shouldEqual(t, f.PkgPath(), "github.com/morikuni/failure_test")
}
