package failure

// Is checks whether an error code from the err is any of given code.
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

// Code represents an error code of an error.
// The code should be able to be compared by == operator.
// Basically, it should to be defined as constants.
//
// You can also define your own code type instead of using StringCode type,
// when you want to distinguish errors by type for some purpose (e.g. define
// code type for each package like user, item, auth etc).
type Code interface {
	// ErrorCode returns an error code in string representation.
	ErrorCode() string
}

// StringCode represents an error Code in string.
type StringCode string

// ErrorCode implements the Code interface.
func (c StringCode) ErrorCode() string {
	return string(c)
}
