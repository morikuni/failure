package failure

import (
	"fmt"
	"reflect"
	"strings"
)

type Code comparable

type Field interface {
	// ErrorFieldKey returns the value for identifying the Field type.
	// The returned value must be comparable and should not be built-in types to avoid collisions.
	ErrorFieldKey() any
}

func New[C Code](c C, fields ...Field) error {
	s := Stack{
		code: c,
	}
	s.applyFields(fields)
	return s
}

func Translate[C Code](err error, c C, fields ...Field) error {
	s := Stack{
		code:       c,
		underlying: err,
	}
	s.applyFields(fields)
	return s
}

func Wrap(err error, fields ...Field) error {
	s := Stack{
		underlying: err,
	}
	s.applyFields(fields)
	return s
}

type Stack struct {
	// Using generics here would make it difficult because it requires resolving the type
	// before comparing the error code values, forcing us to use Stack[C].
	code       any
	underlying error
	fields     map[any]any
	order      []any
}

func (s *Stack) applyFields(fields []Field) {
	if len(fields) == 0 {
		return
	}

	s.order = make([]any, len(fields))
	s.fields = make(map[any]any, len(fields))
	for i, f := range fields {
		key := f.ErrorFieldKey()
		if _, exists := s.fields[key]; exists {
			panic(fmt.Errorf("duplicate error field key: %T(%v)", key, key))
		}
		s.order[i] = key
		s.fields[key] = f
	}
}

func (s Stack) Unwrap() error {
	if s.underlying == nil {
		return nil
	}
	return s.underlying
}

func (s Stack) Error() string {
	if s.underlying == nil {
		return s.string()
	}
	return fmt.Sprintf("%s: %s", s.string(), s.underlying)
}

func (s Stack) string() string {
	var b strings.Builder

	fieldsCount := len(s.fields)
	if v, ok := s.fields[KeyCallStack]; ok {
		fieldsCount--
		head := v.(CallStack).HeadFrame()
		b.WriteString(head.Pkg())
		b.WriteRune('.')
		b.WriteString(head.Func())
	}

	if s.code != nil {
		b.WriteRune('(')
		_, err := fmt.Fprint(&b, s.code)
		if err != nil {
			panic(fmt.Errorf("%s: %T", err, s.code))
		}
		b.WriteRune(')')
	}

	if fieldsCount > 0 {
		first := true
		b.WriteRune('[')
		for k, v := range s.fields {
			if k == KeyCallStack {
				continue
			}
			if !first {
				b.WriteString(", ")
			}
			first = false
			_, err := fmt.Fprint(&b, v)
			if err != nil {
				panic(fmt.Errorf("%s: %T", err, v))
			}
		}
		b.WriteRune(']')
	}

	if s.underlying != nil {
		_, err := fmt.Fprintf(&b, ": %v", s.underlying)
		if err != nil {
			panic(fmt.Errorf("%s: %T", err, s.underlying))
		}
	}

	return b.String()
}

func (s Stack) As(target any) bool {
	if t, ok := target.(*Stack); ok {
		*t = s
		return true
	}

	targetType := reflect.TypeOf(target)
	for _, f := range s.fields {
		fType := reflect.TypeOf(f)
		if targetType.Kind() == reflect.Ptr || fType.AssignableTo(targetType.Elem()) {
			targetVal := reflect.ValueOf(target)
			if targetVal.IsNil() {
				panic("failure: target cannot be nil")
			}
			targetVal.Elem().Set(reflect.ValueOf(f))
			return true
		}

		if as, ok := f.(interface{ As(any) bool }); ok {
			if as.As(target) {
				return true
			}
		}
	}
	return false
}
