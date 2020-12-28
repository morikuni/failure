package failure

import "fmt"

type Tracer interface {
	Push(v interface{})
}

func Trace(err error, vs Tracer) {
	if err == nil {
		return
	}

	i := NewIterator(err)
	for i.Next() {
		i.As(vs)
	}
}

type StringTracer []string

func (st *StringTracer) Push(v interface{}) {
	switch t := v.(type) {
	case Code:
		*st = append(*st, fmt.Sprintf("code = %s", t.ErrorCode()))
		return
	case Message:
		*st = append(*st, fmt.Sprintf("message = %s", t))
		return
	case CallStack:
		head := t.HeadFrame()
		*st = append(*st, fmt.Sprintf("[%s] %s:%d", head.Func(), head.Path(), head.Line()))
		return
	case Context:
		for k, v := range t {
			*st = append(*st, fmt.Sprintf("%s = %s", k, v))
		}
		return
	case interface{ Unexpected() bool }:
		if t.Unexpected() {
			*st = append(*st, fmt.Sprintf("unexpected: %v", t))
			return
		}
	}
	*st = append(*st, fmt.Sprintf("%T(%v)", v, v))
}
