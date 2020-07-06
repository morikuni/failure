package failure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// Unwrapper interface is used by iterator.
type Unwrapper interface {
	// UnwrapError should return nearest child error.
	// The returned error can be nil.
	UnwrapError() error
}

var _ = []Unwrapper{
	(*withMessage)(nil),
	(*withContext)(nil),
	(*withCallStack)(nil),
	(*formatter)(nil),
	(*withCode)(nil),
}

var _ = []interface{ Unwrap() error }{
	(*withMessage)(nil),
	(*withContext)(nil),
	(*withCallStack)(nil),
	(*formatter)(nil),
	(*withCode)(nil),
	(*withoutCode)(nil),
	(*withUnexpected)(nil),
}

// Wrapper interface is used by constructor functions.
type Wrapper interface {
	// WrapError should wrap given err to append some
	// capability to the error.
	WrapError(err error) error
}

var _ = []Wrapper{
	WrapperFunc(nil),
	Context{},
	Message(""),
}

// WrapperFunc is an adaptor to use function as the Wrapper interface.
type WrapperFunc func(err error) error

// WrapError implements the Wrapper interface.
func (f WrapperFunc) WrapError(err error) error {
	return f(err)
}

// Message is a wrapper which appends message to an error.
type Message string

// String returns underlying string message.
func (m Message) String() string {
	return string(m)
}

// WrapError implements Wrapper interface.
func (m Message) WrapError(err error) error {
	return &withMessage{m, err}
}

// Messagef returns Message with formatting.
func Messagef(format string, args ...interface{}) Message {
	return Message(fmt.Sprintf(format, args...))
}

type withMessage struct {
	message    Message
	underlying error
}

func (w *withMessage) Error() string {
	return fmt.Sprintf("%s: %s", w.message, w.underlying)
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use Unwrap.
func (w *withMessage) UnwrapError() error {
	return w.Unwrap()
}

func (w *withMessage) Unwrap() error {
	return w.underlying
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use As method on Iterator.
func (w *withMessage) GetMessage() string {
	return w.message.String()
}

func (w *withMessage) As(x interface{}) bool {
	if m, ok := x.(*Message); ok {
		*m = w.message
		return true
	}
	return false
}

// MessageOf extracts a message from the err.
func MessageOf(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	i := NewIterator(err)
	for i.Next() {
		var msg Message
		if i.As(&msg) {
			return msg.String(), true
		}
	}

	return "", false
}

// Context is a key-value data which describes the how the error occurred
// for debugging purpose. You must not use context data as a part of your
// application logic. Just print it.
// If you want to extract Context from error for printing purpose, you can
// define an interface with method `GetContext() Context` and use it with
// iterator, like other extraction functions (see: MessageOf).
type Context map[string]string

// WrapError implements the Wrapper interface.
func (c Context) WrapError(err error) error {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	buf := &bytes.Buffer{}
	for _, k := range keys {
		v := c[k]
		if buf.Len() != 0 {
			buf.WriteRune(' ')
		}
		fmt.Fprintf(buf, "%s=%s", k, v)
	}
	return &withContext{c, buf.String(), err}
}

type withContext struct {
	ctx        Context
	memo       string
	underlying error
}

func (w *withContext) Error() string {
	return fmt.Sprintf("%s: %s", w.memo, w.underlying)
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use Unwrap.
func (w *withContext) UnwrapError() error {
	return w.Unwrap()
}

func (w *withContext) Unwrap() error {
	return w.underlying
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use As method on Iterator.
func (w *withContext) GetContext() Context {
	return w.ctx
}

func (w *withContext) As(x interface{}) bool {
	if c, ok := x.(*Context); ok {
		*c = w.ctx
		return true
	}
	return false
}

// WithCallStackSkip appends a call stack to the err skipping first N frames.
// You don't have to use this directly, unless using function Custom.
func WithCallStackSkip(skip int) Wrapper {
	cs := Callers(skip + 1)
	return WrapperFunc(func(err error) error {
		return &withCallStack{
			cs,
			err,
		}
	})
}

type withCallStack struct {
	callStack  CallStack
	underlying error
}

func (w *withCallStack) Error() string {
	head := w.callStack.HeadFrame()
	return fmt.Sprintf("%s.%s: %s", head.Pkg(), head.Func(), w.underlying)
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use Unwrap.
func (w *withCallStack) UnwrapError() error {
	return w.Unwrap()
}

func (w *withCallStack) Unwrap() error {
	return w.underlying
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use As method on Iterator.
func (w *withCallStack) GetCallStack() CallStack {
	return w.callStack
}

func (w *withCallStack) As(x interface{}) bool {
	if cs, ok := x.(*CallStack); ok {
		*cs = w.callStack
		return true
	}
	return false
}

// CallStackOf extracts a call stack from the err.
// Returned call stack is for the most deepest place (appended first).
func CallStackOf(err error) (CallStack, bool) {
	if err == nil {
		return nil, false
	}

	var (
		last   CallStack
		exists bool
	)
	i := NewIterator(err)
	for i.Next() {
		exists = i.As(&last) || exists
	}

	return last, exists
}

// WithFormatter appends an error formatter to the err.
//
//     %v+: Print trace for each place, and deepest call stack.
//     %#v: Print raw structure of the error.
//     others (%s, %v): Same as err.Error().
//
// You don't have to use this directly, unless using function Custom.
func WithFormatter() Wrapper {
	return WrapperFunc(func(err error) error {
		return &formatter{error: err}
	})
}

type formatter struct {
	error     `json:"-"`
	Detail    []jsonDetail `json:"detail"`
	CallStack []jsonFrame  `json:"callStack"`
}

type jsonDetail struct {
	jsonFrame     `json:"frame"`
	Context       map[string]string `json:"context,omitempty"`
	Message       *string           `json:"message,omitempty"`
	Code          *string           `json:"code,omitempty"`
	*jsonRawError `json:"rawError,omitempty"`
}

type jsonRawError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type jsonFrame struct {
	Func   string `json:"func"`
	Source string `json:"source"`
}

// Deprecated: This function will be deleted in v1.0.0 release. Please use Unwrap.
func (f *formatter) UnwrapError() error {
	return f.Unwrap()
}

func (f *formatter) Unwrap() error {
	return f.error
}

func (*formatter) IsFormatter() {}

func (f *formatter) Format(s fmt.State, verb rune) {
	if verb != 'v' { // %s
		io.WriteString(s, f.Error())
		return
	}

	if s.Flag('#') { // %#v
		type formatter struct {
			error
		}
		fmt.Fprintf(s, "%#v", &formatter{f.error})
		return
	}

	if !s.Flag('+') { // %v
		io.WriteString(s, f.Error())
		return
	}

	// %+v
	type formatter interface {
		IsFormatter()
	}

	i := NewIterator(f.error)
	for i.Next() {
		err := i.Error()
		if _, ok := err.(formatter); ok {
			continue
		}
		var (
			cs   CallStack
			ctx  Context
			msg  Message
			code Code
		)
		switch {
		case i.As(&cs):
			fmt.Fprintf(s, "%+v\n", cs.HeadFrame())
		case i.As(&ctx):
			for k, v := range ctx {
				fmt.Fprintf(s, "    %s = %s\n", k, v)
			}
		case i.As(&msg):
			fmt.Fprintf(s, "    message(%q)\n", msg)
		case i.As(&code):
			fmt.Fprintf(s, "    code(%s)\n", code.ErrorCode())
		default:
			fmt.Fprintf(s, "    %T(%q)\n", err, err.Error())
		}
	}

	fmt.Fprint(s, "[CallStack]\n")
	if cs, ok := CallStackOf(f); ok {
		for _, f := range cs.Frames() {
			fmt.Fprintf(s, "    %+v\n", f)
		}
	}
}

func JSONFormat(err error) []byte {
	type jsonFormatter interface {
		JSON() ([]byte, error)
	}
	jf := err.(jsonFormatter)
	b, _ := jf.JSON()
	return b
}

func (f *formatter) JSON() ([]byte, error) {
	var detail *jsonDetail

	type formatter interface {
		IsFormatter()
	}

	i := NewIterator(f.error)
	for i.Next() {
		err := i.Error()
		if _, ok := err.(formatter); ok {
			continue
		}

		var (
			cs    CallStack
			ctx   Context
			msg   Message
			code  Code
			frame Frame
		)

		switch {
		case i.As(&cs):
			if detail != nil {
				f.Detail = append(f.Detail, *detail)
			}
			detail = new(jsonDetail)
			frame = cs.HeadFrame()
			detail.jsonFrame.Func = fmt.Sprintf("%s.%s", frame.Pkg(), frame.Func())
			detail.jsonFrame.Source = fmt.Sprintf(" %s:%d ", frame.Path(), frame.Line())
		case i.As(&ctx):
			detail.Context = make(map[string]string)
			for k, v := range ctx {
				detail.Context[k] = v
			}
		case i.As(&msg):
			s := msg.String()
			detail.Message = &s
		case i.As(&code):
			s := code.ErrorCode()
			detail.Code = &s
		default:
			detail.jsonRawError = &jsonRawError{
				Type:    fmt.Sprintf("%T", err),
				Message: fmt.Sprintf("%q", err.Error()),
			}

		}
	}

	if detail != nil {
		f.Detail = append(f.Detail, *detail)
	}

	// reverse order
	for i, j := 0, len(f.Detail)-1; i < j; i, j = i+1, j-1 {
		f.Detail[i], f.Detail[j] = f.Detail[j], f.Detail[i]
	}

	if cs, ok := CallStackOf(f); ok {
		for _, frame := range cs.Frames() {
			frame := jsonFrame{
				Func:   fmt.Sprintf("%s.%s", frame.Pkg(), frame.Func()),
				Source: fmt.Sprintf(" %s:%d ", frame.Path(), frame.Line()),
			}
			f.CallStack = append(f.CallStack, frame)
		}
	}

	return json.Marshal(f)
}

// WithUnexpected wraps the err to mark it is unexpected.
// You don't have to use this directly, unless using function Custom.
// Please use Unexpected or MarkUnexpected.
func WithUnexpected() Wrapper {
	return WrapperFunc(func(err error) error {
		return &withUnexpected{err}
	})
}

type withUnexpected struct {
	underlying error
}

func (w *withUnexpected) Unwrap() error {
	return w.underlying
}

func (w *withUnexpected) Error() string {
	return fmt.Sprintf("unexpected: %s", w.underlying)
}

func (*withUnexpected) Unexpected() bool {
	return true
}
