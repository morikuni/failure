package failure

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
