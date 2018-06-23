package failure_test

import (
	"testing"

	"github.com/morikuni/failure"
	"github.com/stretchr/testify/assert"
)

type CustomCode string

func (c CustomCode) ErrorCode() string {
	return string(c)
}

func TestCode(t *testing.T) {
	const (
		s failure.StringCode = "123"
		i failure.IntCode    = 123
		c CustomCode         = "123"

		s2 failure.StringCode = "123"
		i2 failure.IntCode    = 123
		c2 CustomCode         = "123"
	)

	assert.Equal(t, "123", s.ErrorCode())
	assert.Equal(t, "123", i.ErrorCode())
	assert.Equal(t, "123", c.ErrorCode())

	assert.Equal(t, s, s2)
	assert.Equal(t, i, i2)
	assert.Equal(t, c, c2)

	assert.NotEqual(t, s, i)
	assert.NotEqual(t, s, c)
	assert.NotEqual(t, i, c)
}
