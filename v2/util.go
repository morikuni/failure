package failure

import (
	"errors"
	"fmt"
)

type valuer interface {
	Value(any) any
}

// Value retrieves the value associated with the specified key from the given
// error. It unwraps the error until it finds a matching key or reaches the end
// of the error chain.
func Value[K comparable](err error, key K) any {
	for {
		if err == nil {
			return nil
		}
		if valuer, ok := err.(valuer); ok {
			v := valuer.Value(key)
			if v != nil {
				return v
			}
		}
		err = errors.Unwrap(err)
	}
}

// ValueAs is similar to Value, but also asserts that the value has the specified
// type V. If the value is not of the expected type, it panics with an error
// message.
func ValueAs[V any, K comparable](err error, key K) (zero V, _ bool) {
	v := Value(err, key)
	if v == nil {
		return zero, false
	}
	t, ok := v.(V)
	if !ok {
		panic(fmt.Sprintf("failure: value for key=%T(%v) is not type=%T", key, key, zero))
	}
	return t, true
}

// OriginValue retrieves the first value set for the specified key within the
// given error. It unwraps the error, tracking the earliest encountered value
// with the given key until reaching the end of the error chain.
func OriginValue[K comparable](err error, key K) any {
	var origin any
	for {
		if err == nil {
			return origin
		}
		if valuer, ok := err.(valuer); ok {
			v := valuer.Value(key)
			if v != nil {
				origin = v
			}
		}
		err = errors.Unwrap(err)
	}
}

// OriginValueAs is similar to OriginValue, but also asserts that the value has
// the specified type V. If the value is not of the expected type, it panics with
// an error message.
func OriginValueAs[V any, K comparable](err error, key K) (zero V, _ bool) {
	v := OriginValue(err, key)
	if v == nil {
		return zero, false
	}
	t, ok := v.(V)
	if !ok {
		panic(fmt.Sprintf("failure: value for key=%T(%v) is not type=%T", key, key, zero))
	}
	return t, true
}

// Is checks if the error has any of the specified codes. It returns true if a
// matching code is found.
func Is[C Code](err error, code ...C) bool {
	c := Value(err, KeyCode)
	for _, cc := range code {
		if c == cc {
			return true
		}
	}
	return false
}

// CodeOf retrieves an error code associated with the given error.
func CodeOf(err error) any {
	return Value(err, KeyCode)
}

// MessageOf retrieves a Message associated with the given error.
func MessageOf(err error) Message {
	v, _ := ValueAs[Message](err, KeyMessage)
	return v
}

// CallStackOf retrieves a CallStack associated with the given error.
func CallStackOf(err error) CallStack {
	v, _ := OriginValueAs[CallStack](err, KeyCallStack)
	return v
}

// PopStack unwraps the error, returning the first Stack found in the error chain
// and the remaining tail of the error.
func PopStack(err error) (_ Stack, tail error) {
	for {
		if err == nil {
			return nil, nil
		}
		if st, ok := err.(Stack); ok {
			return st, st.Unwrap()
		}
		err = errors.Unwrap(err)
	}
}
