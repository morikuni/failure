package failure_test

import (
	"io"
	"testing"

	"github.com/morikuni/failure"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
		assert.IsType(t, wantTypes[c], err)
		c++
	}
}

func TestCauseOf(t *testing.T) {
	f := failure.Wrap(io.EOF)
	assert.Equal(t, io.EOF, failure.CauseOf(f))

	base := failure.Wrap(io.EOF)
	pkgErr := errors.Wrap(base, "aaa")
	assert.Equal(t, io.EOF, failure.CauseOf(failure.Wrap(pkgErr)))

	assert.Nil(t, failure.CauseOf(nil))
}
