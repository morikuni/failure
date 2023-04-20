package failure

import (
	"errors"
)

func Value(err error, key any) any {
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

func OriginValue(err error, key any) any {
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

func Is[C Code](err error, code ...C) bool {
	c := Value(err, KeyCode)
	for _, cc := range code {
		if c == cc {
			return true
		}
	}
	return false
}
