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
	base := failure.New(TestCodeA, failure.Message("xxx"), failure.MessageKV{"zzz": "true"})
	pkgErr := errors.New("yyy")
	tests := map[string]struct {
		err error

		shouldNil     bool
		wantCode      failure.Code
		wantMessage   string
		wantStackLine int
		wantError     string
	}{
		"new": {
			err: failure.New(TestCodeA, failure.MessageKV{"aaa": "1"}),

			shouldNil:     false,
			wantCode:      TestCodeA,
			wantMessage:   "",
			wantStackLine: 32,
			wantError:     "failure_test.TestFailure: aaa=1: code(code_a)",
		},
		"translate": {
			err: failure.Translate(base, TestCodeB),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "xxx",
			wantStackLine: 20,
			wantError:     "failure_test.TestFailure: code(1): failure_test.TestFailure: xxx: zzz=true: code(code_a)",
		},
		"overwrite": {
			err: failure.Translate(base, TestCodeB, failure.Messagef("aaa: %s", "bbb"), failure.MessageKV{"ccc": "1", "ddd": "2"}),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "aaa: bbb",
			wantStackLine: 20,
			wantError:     "failure_test.TestFailure: aaa: bbb: ccc=1 ddd=2: code(1): failure_test.TestFailure: xxx: zzz=true: code(code_a)",
		},
		"wrap": {
			err: failure.Wrap(io.EOF),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "",
			wantStackLine: 59,
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
		"pkg/errors": {
			err: failure.Translate(pkgErr, TestCodeB, failure.Message("aaa")),

			shouldNil:     false,
			wantCode:      TestCodeB,
			wantMessage:   "aaa",
			wantStackLine: 21,
			wantError:     "failure_test.TestFailure: aaa: code(1): yyy",
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
		"fundamental": {
			err: failure.Fundamental("fundamental error", failure.MessageKV{"aaa": "1"}),

			shouldNil:     false,
			wantCode:      nil,
			wantMessage:   "fundamental error",
			wantStackLine: 104,
			wantError:     "failure_test.TestFailure: aaa=1: fundamental error",
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
	e2 := failure.Translate(e1, TestCodeA, failure.Message("xxx"), failure.MessageKV{"zzz": "true"})
	err := failure.Wrap(e2)

	want := "failure_test.TestFailure_Format: failure_test.TestFailure_Format: xxx: zzz=true: code(code_a): yyy"
	assert.Equal(t, want, fmt.Sprintf("%s", err))
	assert.Equal(t, want, fmt.Sprintf("%v", err))

	exp := `failure.formatter{error:failure.withCallStack{.*`
	assert.Regexp(t, exp, fmt.Sprintf("%#v", err))

	exp = `\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:155
\[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:154
    message\("xxx"\)
    zzz = true
    code\(code_a\)
    \*errors.errorString\("yyy"\)
\[CallStack\]
    \[failure_test.TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:154
    \[.*`
	assert.Regexp(t, exp, fmt.Sprintf("%+v", err))
}

func BenchmarkFailure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		failure.Wrap(failure.Translate(failure.New(failure.StringCode("error")), failure.StringCode("failure")))
	}
}
