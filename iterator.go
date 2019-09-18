package failure

// NewIterator creates an iterator for the err.
func NewIterator(err error) *Iterator {
	return &Iterator{guardianUnwapper{err}}
}

// Iterator is designed to iterate wrapped errors with for loop.
type Iterator struct {
	err error
}

// Next tries to unwrap an error and returns whether the next
// error is present. Since this method updates internal state of the
// iterator, should be called only once per iteration.
func (i *Iterator) Next() bool {
	i.err = i.unwrapError()
	if i.err == nil {
		return false
	}
	return true
}

func (i *Iterator) unwrapError() error {
	type causer interface {
		Cause() error
	}
	type go113error interface {
		Unwrap() error
	}

	switch t := i.err.(type) {
	case go113error:
		return t.Unwrap()
	case Unwrapper:
		return t.UnwrapError()
	case causer:
		return t.Cause()
	}
	return nil
}

// Error returns current error.
func (i *Iterator) Error() error {
	return i.err
}

// As tries to extract data from current error.
// It returns true if the current error implemented As and it returned true.
func (i *Iterator) As(x interface{}) bool {
	switch t := i.Error().(type) {
	case interface{ As(interface{}) bool }:
		return t.As(x)
	default:
		return false
	}
}

type guardianUnwapper struct {
	error
}

func (w guardianUnwapper) UnwrapError() error {
	return w.error
}

// CauseOf returns a most underlying error of the err.
func CauseOf(err error) error {
	if err == nil {
		return nil
	}

	var last error
	i := NewIterator(err)
	for i.Next() {
		last = i.Error()
	}

	return last
}
