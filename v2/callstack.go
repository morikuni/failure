package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type CallStack struct {
	pcs []uintptr
}

func (cs CallStack) SetErrorField(setter FieldSetter) {
	setter.Set(KeyCallStack, cs)
}

func (cs CallStack) HeadFrame() Frame {
	if len(cs.pcs) == 0 {
		return emptyFrame
	}

	rfs := runtime.CallersFrames(cs.pcs[:1])
	f, _ := rfs.Next()
	return Frame{f.File, f.Line, f.Function, f.PC}
}

func (cs CallStack) Frames() []Frame {
	if len(cs.pcs) == 0 {
		return nil
	}

	rfs := runtime.CallersFrames(cs.pcs)

	fs := make([]Frame, 0, len(cs.pcs))
	for {
		f, more := rfs.Next()

		fs = append(fs, Frame{f.File, f.Line, f.Function, f.PC})

		if !more {
			break
		}
	}
	return fs
}

func (cs CallStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range cs.Frames() {
				fmt.Fprintf(s, "%+v\n", f)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", cs.Frames())
		default:
			fs := cs.Frames()
			l := len(fs)
			if l == 0 {
				return
			}
			for _, f := range fs[:l-1] {
				fmt.Fprintf(s, "%s.%s: ", f.Pkg(), f.Func())
			}
			fmt.Fprintf(s, "%v", fs[l-1].Func())
		}
	case 's':
		fmt.Fprintf(s, "%v", cs)
	}
}

// NewCallStack returns call stack from program counters.
// You can use Callers for usual usage.
func NewCallStack(pcs []uintptr) CallStack {
	return CallStack{pcs}
}

// Callers returns a call stack for the current state.
func Callers(skip int) CallStack {
	var pcs [32]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	return NewCallStack(pcs[:n])
}

var emptyFrame = Frame{"???", 0, "???", uintptr(0)}

type Frame struct {
	file     string
	line     int
	function string
	pc       uintptr
}

func (f Frame) Path() string {
	return f.file
}

func (f Frame) File() string {
	return filepath.Base(f.file)
}

func (f Frame) Line() int {
	return f.line
}

func (f Frame) Func() string {
	fs := strings.Split(path.Base(f.function), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

func (f Frame) PkgPath() string {
	// e.g.
	//   When f.function = github.com/morikuni/failure_test.TestFrame.func1.1
	//   f.PkgPath() = github.com/morikuni/failure_test
	lastSlash := strings.LastIndex(f.function, "/")
	if lastSlash == -1 {
		lastSlash = 0
	}
	return f.function[:strings.Index(f.function[lastSlash:], ".")+lastSlash]
}

func (f Frame) PC() uintptr {
	return f.pc
}

func (f Frame) Pkg() string {
	fs := strings.Split(path.Base(f.function), ".")
	return fs[0]
}

func (f Frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "[%s.%s] ", f.Pkg(), f.Func())
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s:%d", f.Path(), f.Line())
	}
}
