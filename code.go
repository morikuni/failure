package failure

import "strconv"

// Code represents an error code.
// The package provides 2 types codes, StringCode and IntCode.
type Code interface {
	// ErrorCode returns an error code in string representation.
	// Use this only for printing the code, not for comparing codes.
	ErrorCode() string
}

// StringCode represents an error code in string.
type StringCode string

// ErrorCode implements Code.
func (c StringCode) ErrorCode() string {
	return string(c)
}

// IntCode represents an error code in int64.
type IntCode int64

// ErrorCode implements Code.
func (c IntCode) ErrorCode() string {
	return strconv.FormatInt(int64(c), 10)
}
