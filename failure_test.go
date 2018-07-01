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
	type Input struct {
		Err error
	}
	type Expect struct {
		Code      failure.Code
		Message   string
		Fields    []failure.Info
		StackLine int
		Error     string
	}
	type Test struct {
		Input
		Expect
	}

	base := failure.New(TestCodeA).WithMessage("xxx").WithInfo(failure.Info{"zzz": true})
	pkgErr := errors.New("yyy")
	tests := map[string]Test{
		"new": {
			Input{failure.New(TestCodeA).WithInfo(failure.Info{"aaa": 1})},
			Expect{
				TestCodeA,
				failure.DefaultMessage,
				[]failure.Info{{"aaa": 1}},
				39,
				"TestFailure(code_a)",
			},
		},
		"translate": {
			Input{failure.Translate(base, TestCodeB)},
			Expect{
				TestCodeB,
				"xxx",
				[]failure.Info{{"zzz": true}},
				35,
				"TestFailure(1): TestFailure(code_a)",
			},
		},
		"overwrite": {
			Input{failure.Translate(base, TestCodeB).WithMessage("aaa").WithInfo(failure.Info{"bbb": 1})},
			Expect{
				TestCodeB,
				"aaa",
				[]failure.Info{{"bbb": 1}, {"zzz": true}},
				35,
				"TestFailure(1): TestFailure(code_a)",
			},
		},
		"wrap": {
			Input{failure.Wrap(io.EOF)},
			Expect{
				failure.Unknown,
				failure.DefaultMessage,
				nil,
				69,
				"TestFailure: " + io.EOF.Error(),
			},
		},
		"pkg/errors": {
			Input{failure.Translate(pkgErr, TestCodeB).WithMessage("aaa")},
			Expect{
				TestCodeB,
				"aaa",
				nil,
				36,
				"TestFailure(1): yyy",
			},
		},
		"nil": {
			Input{nil},
			Expect{
				nil,
				"",
				nil,
				0,
				"",
			},
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, test.Expect.Code, failure.CodeOf(test.Input.Err))
			assert.Equal(t, test.Expect.Message, failure.MessageOf(test.Input.Err))
			assert.Equal(t, test.Expect.Fields, failure.InfoListOf(test.Input.Err))

			if test.Expect.Error != "" {
				assert.EqualError(t, test.Input.Err, test.Expect.Error)
			} else {
				assert.Nil(t, test.Input.Err)
			}

			cs := failure.CallStackOf(test.Input.Err)
			if test.Expect.StackLine != 0 {
				fs := cs.Frames()
				require.NotEmpty(t, fs)
				if !assert.Equal(t, test.Expect.StackLine, fs[0].Line()) {
					t.Log(fs[0])
				}
			} else {
				assert.Nil(t, cs)
			}
		})
	}
}

func TestFailure_Format(t *testing.T) {
	type Input struct {
		Err    error
		Format string
	}
	type Expect struct {
		OutputRegexp string
	}
	type Test struct {
		Input
		Expect
	}

	base := failure.New(TestCodeA).WithMessage("xxx").WithInfo(failure.Info{"zzz": true})
	tests := map[string]Test{
		"v": {
			Input{
				failure.Wrap(base),
				"%v",
			},
			Expect{
				`TestFailure_Format: TestFailure_Format\(code_a\)`,
			},
		},
		"+v": {
			Input{
				failure.Wrap(base),
				"%+v",
			},
			Expect{
				`TestFailure_Format: TestFailure_Format\(code_a\)
  Info:
    zzz = true
  CallStack:
    \[TestFailure_Format\] /.*/github.com/morikuni/failure/failure_test.go:139
    \[.*`,
			},
		},
		"#v": {
			Input{
				failure.Wrap(base).WithMessage("hello"),
				"%#v",
			},
			Expect{
				`failure.Failure{code:failure.Code\(nil\), message:"hello", callStack:\(\*failure.callStack\)\(0x.*\), info:failure.Info\(nil\), underlying:\(\*failure.Failure\)\(0x.*\)}`,
			},
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			assert.Regexp(t, test.Expect.OutputRegexp, fmt.Sprintf(test.Input.Format, test.Input.Err))
		})
	}
}

func TestCauseOf(t *testing.T) {
	f := failure.New(TestCodeB)
	assert.Equal(t, f, failure.CauseOf(f))

	base := failure.New(TestCodeA)
	pkgErr := errors.Wrap(base, "aaa")
	assert.Equal(t, base, failure.CauseOf(failure.Wrap(pkgErr)))
}

func BenchmarkFailure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		failure.Wrap(failure.Translate(failure.New(failure.StringCode("error")), failure.StringCode("failure")))
	}
}
