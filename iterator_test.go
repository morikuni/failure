package failure_test

import (
	"io"
	"reflect"
	"testing"

	"github.com/morikuni/failure"
)

type a struct {
	error
}

func (a a) UnwrapError() error {
	return a.error
}

type b struct {
	error
}

// like a pkg/errors
func (b b) Cause() error {
	return b.error
}

type c struct {
	error
}

func (c c) UnwrapError() error {
	return c.error
}

func TestIterator(t *testing.T) {
	err := a{b{c{io.EOF}}}
	wantTypes := []interface{}{a{}, b{}, c{}, io.EOF}

	i := failure.NewIterator(err)
	var c int
	for i.Next() {
		err := i.Error()
		shouldEqual(t, reflect.TypeOf(err), reflect.TypeOf(wantTypes[c]))
		c++
	}
}

func TestCauseOf(t *testing.T) {
	f := failure.Wrap(io.EOF)
	shouldEqual(t, failure.CauseOf(f), io.EOF)

	base := failure.Wrap(io.EOF)
	err := a{b{c{base}}}
	shouldEqual(t, failure.CauseOf(failure.Wrap(err)), io.EOF)

	shouldEqual(t, failure.CauseOf(nil), nil)
}
