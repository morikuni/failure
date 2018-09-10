package failure

import "strconv"

// Is checks whether err represents any of given code.
func Is(err error, codes ...Code) bool {
	if len(codes) == 0 {
		return false
	}

	c := CodeOf(err)
	if c == nil {
		return false
	}

	for i := range codes {
		if c == codes[i] {
			return true
		}
	}
	return false
}

// Code represents an error Code.
// The code should not have internal state, so it should be
// defined as a variable.
// StringCode or IntCode are recommended if you don't need
// custom behavior on the code.
type Code interface {
	// ErrorCode returns an error Code in string representation.
	ErrorCode() string
}

// StringCode represents an error Code in string.
type StringCode string

// ErrorCode implements the Code interface.
func (c StringCode) ErrorCode() string {
	return string(c)
}

// IntCode represents an error Code in int64.
type IntCode int64

// ErrorCode implements the Code interface.
func (c IntCode) ErrorCode() string {
	return strconv.FormatInt(int64(c), 10)
}
