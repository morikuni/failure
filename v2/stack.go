package failure

import (
	"fmt"
	"reflect"
	"strings"
)

// NewStack creates a new Stack from an underlying error and a variadic set of
// Field slices. It returns a Stack with the specified fields and the underlying
// error. Panics if both the underlying error and fields are empty.
func NewStack(underlying error, fieldsSet ...[]Field) Stack {
	fieldCount := 0
	for _, fields := range fieldsSet {
		fieldCount += len(fields)
	}

	if underlying == nil && fieldCount == 0 {
		panic("failure: invalid Stack")
	}

	s := &stack{
		underlying: underlying,
	}

	if fieldCount == 0 {
		return s
	}

	s.order = make([]any, 0, fieldCount)
	s.fields = make(map[any]any, fieldCount)

	setter := asSetter(*s)
	for _, fields := range fieldsSet {
		for _, f := range fields {
			f.SetErrorField(&setter)
		}
	}

	st := stack(setter)
	return &st
}

// Stack represents a stack of errors accumulated through wrapping, along with
// additional information. Similar to CallStack represents the program's call
// history, Stack represents the error handling history. By storing key-value
// data, Stack extends errors with arbitrary information. Stack is also designed
// to allow embedding within custom structs, enabling the implementation of
// additional interfaces, such as gRPC Error (GRPCStatus method).
type Stack interface {
	error
	Unwrap() error
	Value(key any) any
	As(key any) bool
}

var _ Stack = (*stack)(nil)

type stack struct {
	underlying error
	fields     map[any]any
	order      []any
}

type asSetter stack

func (s *asSetter) Set(key, value any) {
	if _, exists := s.fields[key]; exists {
		panic(fmt.Sprintf("failure: duplicate error field key: %T(%v)", key, key))
	}
	s.order = append(s.order, key)
	s.fields[key] = value
}

func (s *stack) Unwrap() error {
	if s.underlying == nil {
		return nil
	}
	return s.underlying
}

func (s *stack) Error() string {
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
				ef.FormatError(&b)
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

func (s *stack) As(target any) bool {
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

func (s *stack) Value(key any) any {
	return s.fields[key]
}
