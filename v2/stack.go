package failure

import (
	"fmt"
	"reflect"
	"strings"
)

type Code comparable

type Field interface {
	SetErrorField(setter FieldSetter)
}

type FieldSetter interface {
	Set(key, value any)
}

type ErrorFormatter interface {
	FormatError() string
}

func New[C Code](c C, fields ...Field) error {
	return newStack(nil, c, fields)
}

func Translate[C Code](err error, c C, fields ...Field) error {
	return newStack(err, c, fields)
}

func Wrap(err error, fields ...Field) error {
	if err == nil {
		return nil
	}
	return newStack(err, nil, fields)
}

func newStack(err error, code any, fields []Field) error {
	var defaultFields []Field
	if code == nil {
		defaultFields = []Field{
			Callers(2),
		}
	} else {
		defaultFields = []Field{
			codeField{code},
			Callers(2),
		}
	}
	return NewStack(err, defaultFields, fields)
}

func NewStack(underlying error, fieldsSet ...[]Field) Stack {
	fieldCount := 0
	for _, fields := range fieldsSet {
		fieldCount += len(fields)
	}

	if underlying == nil && fieldCount == 0 {
		panic("failure: invalid stack")
	}

	s := Stack{
		underlying: underlying,
	}

	if fieldCount == 0 {
		return s
	}

	s.order = make([]any, 0, fieldCount)
	s.fields = make(map[any]any, fieldCount)

	setter := asSetter(s)
	for _, fields := range fieldsSet {
		for _, f := range fields {
			f.SetErrorField(&setter)
		}
	}

	return Stack(setter)
}

type Stack struct {
	underlying error
	fields     map[any]any
	order      []any
}

type asSetter Stack

func (s *asSetter) Set(key, value any) {
	if _, exists := s.fields[key]; exists {
		panic(fmt.Errorf("duplicate error field key: %T(%v)", key, key))
	}
	s.order = append(s.order, key)
	s.fields[key] = value
}

func (s Stack) Unwrap() error {
	if s.underlying == nil {
		return nil
	}
	return s.underlying
}

func (s Stack) Error() string {
	var b strings.Builder

	fieldsCount := len(s.fields)
	if v, ok := s.fields[KeyCallStack]; ok {
		fieldsCount--
		head := v.(CallStack).HeadFrame()
		b.WriteString(head.Pkg())
		b.WriteRune('.')
		b.WriteString(head.Func())
	}

	if v, ok := s.fields[KeyCode]; ok {
		fieldsCount--
		b.WriteString("(code=")
		fmt.Fprint(&b, v)
		b.WriteRune(')')
	}

	if fieldsCount > 0 {
		first := true
		b.WriteRune('[')
		for _, k := range s.order {
			v := s.fields[k]
			switch k {
			case KeyCode, KeyCallStack:
				continue
			}
			if !first {
				b.WriteString(", ")
			}
			first = false
			if ef, ok := v.(ErrorFormatter); ok {
				b.WriteString(ef.FormatError())
			} else {
				fmt.Fprint(&b, v)
			}
		}
		b.WriteRune(']')
	}

	if s.underlying != nil {
		fmt.Fprintf(&b, ": %s", s.underlying.Error())
	}

	return b.String()
}

func (s Stack) As(target any) bool {
	targetType := reflect.TypeOf(target)
	for _, f := range s.fields {
		fType := reflect.TypeOf(f)
		if targetType.Kind() == reflect.Ptr {
			targetElemType := targetType.Elem()
			// Set the value if:
			// 1. target is the same type.
			// 2. target is interface and field is assignable it.
			// Check whether assignable only if target is interface, to prevent unexpected assigning like failure.Context to map[string]string.
			if fType == targetElemType || (targetElemType.Kind() == reflect.Interface && fType.AssignableTo(targetElemType)) {
				targetVal := reflect.ValueOf(target)
				if targetVal.IsNil() {
					panic("failure: target cannot be nil")
				}
				targetVal.Elem().Set(reflect.ValueOf(f))
				return true
			}
		}

		if as, ok := f.(interface{ As(any) bool }); ok {
			if as.As(target) {
				return true
			}
		}
	}
	return false
}
