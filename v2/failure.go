package failure

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
