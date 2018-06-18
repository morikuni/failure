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
	TestCodeA failure.Code = "a"
	TestCodeB failure.Code = "b"
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
				"aaa",
			},
		},
		"nested": {
			Input{failure.Translate(base, TestCodeB, "aaa", failure.Info{"bbb": 1})},
			Expect{
				TestCodeB,
				"aaa",
				[]failure.Info{{"bbb": 1}, {"zzz": true}},
				34,
				"aaa: xxx",
			},
		},
		"with stack": {
			Input{failure.WithFields(io.EOF, nil)},
			Expect{
				failure.Unknown,
				failure.DefaultMessage,
				nil,
				58,
				io.EOF.Error(),
			},
		},
		"pkg/errors": {
			Input{failure.Translate(pkgErr, TestCodeB, "aaa", nil)},
			Expect{
				TestCodeB,
				"aaa",
				nil,
				35,
				"aaa: yyy",
			},
		},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			assert.Equal(t, test.Expect.Code, failure.CodeOf(test.Input.Err))
			assert.Equal(t, test.Expect.Message, failure.MessageOf(test.Input.Err))
			assert.Equal(t, test.Expect.Fields, failure.InfosOf(test.Input.Err))

			st := failure.CallStackOf(test.Input.Err)
			require.NotEmpty(t, st)
			assert.Equal(t, test.Expect.StackLine, st[0].Line())
		})
	}

}
