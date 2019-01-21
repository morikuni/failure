package failure_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestCodeA failure.StringCode = "code_a"
	TestCodeB failure.IntCode    = 1
)

func TestFailure(t *testing.T) {
	base := failure.New(TestCodeA, failure.Message("xxx"), failure.Debug{"zzz": "true"})
	pkgErr := errors.New("yyy")
	tests := map[string]struct {
		err error

		shouldNil     bool
		wantCode      failure.Code
		wantMessage   string
		wantDebugs    []failure.Debug
		wantStackLine int
		wantError     string
	}{
		"new": {
			err: failure.New(TestCodeA, failure.Debug{"aaa": "1"}),

			shouldNil:     false,
			wantCode:      TestCodeA,
			wantMessage:   "",
			wantDebugs:    []failure.Debug{{"aaa": "1"}},
			wantStackLine: 33,
			wantError:     "failure_test.TestFailure: code(code_a)",
		},
		"translate": {
			err: failure.Translate(base, TestCodeB),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "xxx",
			wantDebugs:    []failure.Debug{{"zzz": "true"}},
			wantStackLine: 20,
			wantError:     "failure_test.TestFailure: code(1): failure_test.TestFailure: code(code_a)",
		},
		"overwrite": {
			err: failure.Translate(base, TestCodeB, failure.Message("aaa"), failure.Debug{"bbb": "1"}),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "aaa",
			wantDebugs:    []failure.Debug{{"bbb": "1"}, {"zzz": "true"}},
			wantStackLine: 20,
			wantError:     "failure_test.TestFailure: code(1): failure_test.TestFailure: code(code_a)",
		},
		"wrap": {
			err: failure.Wrap(io.EOF),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantDebugs:    nil,
			wantStackLine: 63,
			wantError:     "failure_test.TestFailure: " + io.EOF.Error(),
		},
		"wrap nil": {
			err: failure.Wrap(nil),

			shouldNil:     true,
			wantCode:      nil,
			wantMessage:   "",
			wantDebugs:    nil,
			wantStackLine: 0,
			wantError:     "",
		},
		"pkg/errors": {
			err: failure.Translate(pkgErr, TestCodeB, failure.Message("aaa")),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "aaa",
			wantDebugs:    nil,
			wantStackLine: 21,
			wantError:     "failure_test.TestFailure: code(1): yyy",
		},
		"nil": {
			err: nil,

			shouldNil:     true,
			wantCode:      nil,
			wantMessage:   "",
			wantDebugs:    nil,
			wantStackLine: 0,
			wantError:     "",
		},
		"custom": {
			err: failure.Custom(io.EOF),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantDebugs:    nil,
			wantStackLine: 0,
			wantError:     io.EOF.Error(),
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			if test.shouldNil {
				assert.NoError(t, test.err)
			} else {
				assert.Error(t, test.err)
			}

			code, ok := failure.CodeOf(test.err)
			assert.Equal(t, test.wantCode != nil, ok)
			assert.Equal(t, test.wantCode, code)

			msg, ok := failure.MessageOf(test.err)
			assert.Equal(t, test.wantMessage != "", ok)
			assert.Equal(t, test.wantMessage, msg)

			assert.Equal(t, test.wantDebugs, failure.DebugsOf(test.err))

			if test.wantError != "" {
				assert.EqualError(t, test.err, test.wantError)
			} else {
				assert.Nil(t, test.err)
			}

			cs, ok := failure.CallStackOf(test.err)
			if test.wantStackLine != 0 {
				assert.True(t, ok)
				fs := cs.Frames()
				require.NotEmpty(t, fs)
				if !assert.Equal(t, test.wantStackLine, fs[0].Line()) {
					t.Log(fs[0])
				}
			} else {
				assert.False(t, ok)
				assert.Nil(t, cs)
			}
		})
	}
}

func TestFailure_Format(t *testing.T) {
	e1 := fmt.Errorf("yyy")
	e2 := failure.Translate(e1, TestCodeA, failure.Message("xxx"), failure.Debug{"zzz": "true"})
	err := failure.Wrap(e2)

	want := "failure_test.TestFailure_Format: failure_test.TestFailure_Format: code(code_a): yyy"
	assert.Equal(t, want, fmt.Sprintf("%s", err))
	assert.Equal(t, want, fmt.Sprintf("%v", err))

	exp := `failure.formatter{error:failure.withCallStack{.*`
	assert.Regexp(t, exp, fmt.Sprintf("%#v", err))

	exp = `\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:157
\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:156
    zzz = true
    message\("xxx"\)
    code\(code_a\)
    \*errors.errorString\("yyy"\)
\[CallStack\]
    \[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:156
    \[.*`
	assert.Regexp(t, exp, fmt.Sprintf("%+v", err))
}

func BenchmarkFailure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		failure.Wrap(failure.Translate(failure.New(failure.StringCode("error")), failure.StringCode("failure")))
	}
}
