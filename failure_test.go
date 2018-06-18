package failure_test

import (
	"io"
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestCodeA failure.Code = "code_a"
	TestCodeB failure.Code = "code_b"
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

	base := failure.New(TestCodeA, "xxx", failure.Info{"zzz": true})
	pkgErr := errors.New("yyy")
	tests := map[string]Test{
		"new": {
			Input{failure.New(TestCodeA, "aaa", failure.Info{"bbb": 1})},
			Expect{
				TestCodeA,
				"aaa",
				[]failure.Info{{"bbb": 1}},
				38,
				"TestFailure(code_a)",
			},
		},
		"nested": {
			Input{failure.Translate(base, TestCodeB, "aaa", failure.Info{"bbb": 1})},
			Expect{
				TestCodeB,
				"aaa",
				[]failure.Info{{"bbb": 1}, {"zzz": true}},
				34,
				"TestFailure(code_b): TestFailure(code_a)",
			},
		},
		"with info": {
			Input{failure.WithInfo(io.EOF, failure.Info{"bbb": 1})},
			Expect{
				failure.Unknown,
				failure.DefaultMessage,
				[]failure.Info{{"bbb": 1}},
				58,
				"TestFailure: " + io.EOF.Error(),
			},
		},
		"pkg/errors": {
			Input{failure.Translate(pkgErr, TestCodeB, "aaa", nil)},
			Expect{
				TestCodeB,
				"aaa",
				nil,
				35,
				"TestFailure(code_b): yyy",
			},
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, test.Expect.Code, failure.CodeOf(test.Input.Err))
			assert.Equal(t, test.Expect.Message, failure.MessageOf(test.Input.Err))
			assert.Equal(t, test.Expect.Fields, failure.InfosOf(test.Input.Err))
			assert.Equal(t, test.Expect.Error, test.Input.Err.Error())

			cs := failure.CallStackOf(test.Input.Err)
			require.NotEmpty(t, cs)
			if !assert.Equal(t, test.Expect.StackLine, cs[0].Line()) {
				t.Log(cs[0])
			}
		})
	}

}
