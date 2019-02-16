package failure

import "strconv"

// Is checks whether err represents any of given code.
func Is(err error, codes ...Code) bool {
	if len(codes) == 0 {
		return false
	}

	c, ok := CodeOf(err)
	if !ok {
		// continue process (don't return) to accept the case Is(err, nil).
		c = nil
	}

	for i := range codes {
		if c == codes[i] {
			return true
		}
	}
	return false
}

// Code represents an error code of the error.
// The code should be able to be compared by == operator.
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
