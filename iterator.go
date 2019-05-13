package failure

// NewIterator creates an iterator for given err.
func NewIterator(err error) *Iterator {
	return &Iterator{guardianUnwapper{err}}
}

// Iterator is designed to iterate errors by unwrapping it
// with for loop.
type Iterator struct {
	err error
}

// Next try to unwrap an error and returns whether the next
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
	switch t := i.err.(type) {
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

type guardianUnwapper struct {
	error
}

func (w guardianUnwapper) UnwrapError() error {
	return w.error
}

// CauseOf returns the most underlying error of err.
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
