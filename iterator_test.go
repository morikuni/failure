package failure_test

import (
	"io"
	"reflect"
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
)

type a struct {
	error
}

func (a a) UnwrapError() error {
	return a.error
}

type b struct {
	a
}

type c struct {
	a
}

func TestIterator(t *testing.T) {
	err := a{b{a{c{a{io.EOF}}}}}
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
	pkgErr := errors.Wrap(base, "aaa")
	shouldEqual(t, failure.CauseOf(failure.Wrap(pkgErr)), io.EOF)

	shouldEqual(t, failure.CauseOf(nil), nil)
}
