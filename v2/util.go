package failure

import (
	"errors"
	"fmt"
)

func Value[K comparable](err error, key K) any {
	for {
		if err == nil {
			return nil
		}
		if valuer, ok := err.(interface{ Value(any) any }); ok {
			v := valuer.Value(key)
			if v != nil {
				return v
			}
		}
		err = errors.Unwrap(err)
	}
}

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

func OriginValue[K comparable](err error, key K) any {
	var origin any
	for {
		if err == nil {
			return origin
		}
		if valuer, ok := err.(interface{ Value(any) any }); ok {
			v := valuer.Value(key)
			if v != nil {
				origin = v
			}
		}
		err = errors.Unwrap(err)
	}
}

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

func Is[C Code](err error, code ...C) bool {
	c := Value(err, KeyCode)
	for _, cc := range code {
		if c == cc {
			return true
		}
	}
	return false
}

func CodeOf(err error) any {
	return Value(err, KeyCode)
}

func MessageOf(err error) Message {
	v, _ := ValueAs[Message](err, KeyMessage)
	return v
}

func CallStackOf(err error) CallStack {
	v, _ := OriginValueAs[CallStack](err, KeyCallStack)
	return v
}

func PopStack(err error) (_ *Stack, tail error) {
	for {
		if err == nil {
			return nil, nil
		}
		if st, ok := err.(*Stack); ok {
			return st, st.Unwrap()
		}
		err = errors.Unwrap(err)
	}
}
