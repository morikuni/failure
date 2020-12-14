package failure

import "fmt"

type VirtualStack interface {
	Push(v interface{})
}

func AsVirtualStack(err error, vs VirtualStack) {
	if err == nil {
		return
	}

	i := NewIterator(err)
	for i.Next() {
		i.As(vs)
	}
}

type SliceStack []string

func (s *SliceStack) Push(v interface{}) {
	switch t := v.(type) {
	case Code:
		*s = append(*s, fmt.Sprintf("code = %s", t.ErrorCode()))
		return
	case Message:
		*s = append(*s, fmt.Sprintf("message = %s", t))
		return
	case CallStack:
		head := t.HeadFrame()
		*s = append(*s, fmt.Sprintf("[%s] %s:%d", head.Func(), head.Path(), head.Line()))
		return
	case Context:
		for k, v := range t {
			*s = append(*s, fmt.Sprintf("%s = %s", k, v))
		}
		return
	case interface{ Unexpected() bool }:
		if t.Unexpected() {
			*s = append(*s, fmt.Sprintf("unexpected: %v", t))
			return
		}
	}
	*s = append(*s, fmt.Sprintf("%T(%v)", v, v))
}
