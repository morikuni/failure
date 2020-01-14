package failure_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/morikuni/failure"
)

const (
	TestCodeA failure.StringCode = "code_a"
	TestCodeB failure.StringCode = "1"
)

func TestFailure(t *testing.T) {
	base := failure.New(TestCodeA, failure.Message("xxx"), failure.Context{"zzz": "true"})
	tests := map[string]struct {
		err error

		shouldNil     bool
		wantCode      failure.Code
		wantMessage   string
		wantStackLine int
		wantError     string
	}{
		"new": {
			err: failure.New(TestCodeA, failure.Context{"aaa": "1"}),

			shouldNil:     false,
			wantCode:      TestCodeA,
			wantMessage:   "",
			wantStackLine: 28,
			wantError:     "failure_test.TestFailure: aaa=1: code(code_a)",
		},
		"translate": {
			err: failure.Translate(base, TestCodeB),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "xxx",
			wantStackLine: 17,
			wantError:     "failure_test.TestFailure: code(1): failure_test.TestFailure: xxx: zzz=true: code(code_a)",
		},
		"overwrite": {
			err: failure.Translate(base, TestCodeB, failure.Messagef("aaa: %s", "bbb"), failure.Context{"ccc": "1", "ddd": "2"}),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "aaa: bbb",
			wantStackLine: 17,
			wantError:     "failure_test.TestFailure: aaa: bbb: ccc=1 ddd=2: code(1): failure_test.TestFailure: xxx: zzz=true: code(code_a)",
		},
		"wrap": {
			err: failure.Wrap(io.EOF),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 55,
			wantError:     "failure_test.TestFailure: " + io.EOF.Error(),
		},
		"wrap nil": {
			err: failure.Wrap(nil),

			shouldNil:     true,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 0,
			wantError:     "",
		},
		"nil": {
			err: nil,

			shouldNil:     true,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 0,
			wantError:     "",
		},
		"custom": {
			err: failure.Custom(io.EOF),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 0,
			wantError:     io.EOF.Error(),
		},
		"unexpected": {
			err: failure.Unexpected("unexpected error", failure.Context{"aaa": "1"}),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 91,
			wantError:     "failure_test.TestFailure: aaa=1: unexpected error",
		},
		"mark unexpected": {
			err: failure.MarkUnexpected(base),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "xxx",
			wantStackLine: 17,
			wantError:     "failure_test.TestFailure: code_eliminated: failure_test.TestFailure: xxx: zzz=true: code(code_a)",
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			if test.shouldNil {
				shouldEqual(t, test.err, nil)
			} else {
				shouldDiffer(t, test.err, nil)
			}

			code, ok := failure.CodeOf(test.err)
			shouldEqual(t, ok, test.wantCode != nil)
			shouldEqual(t, code, test.wantCode)

			msg, ok := failure.MessageOf(test.err)
			shouldEqual(t, ok, test.wantMessage != "")
			shouldEqual(t, msg, test.wantMessage)

			if test.wantError != "" {
				shouldEqual(t, test.err.Error(), test.wantError)
			} else {
				shouldEqual(t, test.err, nil)
			}

			cs, ok := failure.CallStackOf(test.err)
			if test.wantStackLine != 0 {
				shouldEqual(t, ok, true)
				fs := cs.Frames()
				shouldDiffer(t, len(fs), 0)
				shouldEqual(t, fs[0].Line(), test.wantStackLine)
			} else {
				shouldEqual(t, ok, false)
				shouldEqual(t, cs, nil)
			}
		})
	}
}

func TestFailure_Format(t *testing.T) {
	e1 := fmt.Errorf("yyy")
	e2 := failure.Translate(e1, TestCodeA, failure.Message("xxx"), failure.Context{"zzz": "true"})
	err := failure.Wrap(e2)

	want := "failure_test.TestFailure_Format: failure_test.TestFailure_Format: xxx: zzz=true: code(code_a): yyy"
	shouldEqual(t, fmt.Sprintf("%s", err), want)
	shouldEqual(t, fmt.Sprintf("%v", err), want)

	exp := `&failure.formatter{error:\(\*failure.withCallStack\)\(.*`
	shouldMatch(t, fmt.Sprintf("%#v", err), exp)

	exp = `\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:149
\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:148
    message\("xxx"\)
    zzz = true
    code\(code_a\)
    \*errors.errorString\("yyy"\)
\[CallStack\]
    \[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:148
    \[.*`
	shouldMatch(t, fmt.Sprintf("%+v", err), exp)
}

func BenchmarkFailure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		failure.Wrap(failure.Translate(failure.New(failure.StringCode("error")), failure.StringCode("failure")))
	}
}
