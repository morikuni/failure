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

	shouldContain(t, fs[0].Path(), "/failure/v2/callstack_test.go")
	shouldContain(t, fs[0].File(), "callstack_test.go")
	shouldEqual(t, fs[0].Func(), "X")
	shouldEqual(t, fs[0].Line(), 11)
	shouldEqual(t, fs[0].Pkg(), "v2_test") // is this bug?
	shouldEqual(t, fs[0].PkgPath(), "github.com/morikuni/failure/v2_test")

	shouldContain(t, fs[1].Path(), "/failure/v2/callstack_test.go")
	shouldContain(t, fs[1].File(), "callstack_test.go")
	shouldEqual(t, fs[1].Func(), "TestCallers")
	shouldEqual(t, fs[1].Line(), 15)
	shouldEqual(t, fs[1].Pkg(), "v2_test")
	shouldEqual(t, fs[1].PkgPath(), "github.com/morikuni/failure/v2_test")
}

func TestCallStack_Format(t *testing.T) {
	cs := X()

	shouldMatch(t,
		fmt.Sprintf("%v", cs),
		`v2_test.X: v2_test.TestCallStack_Format: .*`,
	)
	shouldMatch(t,
		fmt.Sprintf("%s", cs),
		`v2_test.X: v2_test.TestCallStack_Format: .*`,
	)
	shouldMatch(t,
		fmt.Sprintf("%#v", cs),
		`\[\]failure.Frame{/.+/failure/v2/callstack_test.go:11, /.+/failure/v2/callstack_test.go:33, .*}`,
	)
	shouldMatch(t,
		fmt.Sprintf("%+v", cs),
		`\[v2_test.X\] /.+/failure/v2/callstack_test.go:11
\[v2_test.TestCallStack_Format\] /.+/failure/v2/callstack_test.go:33
\[.*`,
	)
}

func TestFrame_Format(t *testing.T) {
	f := X().HeadFrame()

	shouldMatch(t,
		fmt.Sprintf("%v", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%s", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%#v", f),
		`/.+/failure/v2/callstack_test.go:11`,
	)
	shouldMatch(t,
		fmt.Sprintf("%+v", f),
		`\[v2_test.X\] /.+/failure/v2/callstack_test.go:11`,
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
	shouldContain(t, f.Path(), "/failure/v2/callstack_test.go")
	shouldEqual(t, f.Pkg(), "v2_test")
	shouldEqual(t, f.PkgPath(), "github.com/morikuni/failure/v2_test")
}
