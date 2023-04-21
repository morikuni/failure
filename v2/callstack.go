package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// CallStack represents a stack of program counters.
type CallStack []uintptr

// NewCallStack creates a new CallStack from the provided program counters.
func NewCallStack(pcs []uintptr) CallStack {
	return CallStack(pcs)
}

// Callers returns a CallStack of the caller's goroutine stack. The skip
// parameter determines the number of stack frames to skip before capturing the
// CallStack.
func Callers(skip int) CallStack {
	var pcs [32]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	return NewCallStack(pcs[:n])
}

// SetErrorField implements the Field interface.
func (cs CallStack) SetErrorField(setter FieldSetter) {
	setter.Set(KeyCallStack, cs)
}

// HeadFrame is a method of CallStack that returns the first frame in the
// CallStack. If the CallStack is empty, it returns an empty frame.
func (cs CallStack) HeadFrame() Frame {
	if len(cs) == 0 {
		return emptyFrame
	}

	rfs := runtime.CallersFrames(cs[:1])
	f, _ := rfs.Next()
	return Frame{f.File, f.Line, f.Function, f.PC}
}

// Frames is a method of CallStack that returns a slice of Frame objects
// representing the CallStack's frames.
func (cs CallStack) Frames() []Frame {
	if len(cs) == 0 {
		return nil
	}

	rfs := runtime.CallersFrames(cs)

	fs := make([]Frame, 0, len(cs))
	for {
		f, more := rfs.Next()

		fs = append(fs, Frame{f.File, f.Line, f.Function, f.PC})

		if !more {
			break
		}
	}
	return fs
}

// Format implements the fmt.Formatter interface.
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

var emptyFrame = Frame{"???", 0, "???", uintptr(0)}

// Frame represents a single frame in a CallStack.
type Frame struct {
	file     string
	line     int
	function string
	pc       uintptr
}

// Path returns the full path of the file associated with the Frame.
func (f Frame) Path() string {
	return f.file
}

// File returns the base file name associated with the Frame.
func (f Frame) File() string {
	return filepath.Base(f.file)
}

// Line returns the line number of the file.
func (f Frame) Line() int {
	return f.line
}

// Func returns the function name associated with the Frame.
func (f Frame) Func() string {
	fs := strings.Split(path.Base(f.function), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

// PkgPath returns the package path associated with the Frame.
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

// PC returns the program counter associated with the Frame
func (f Frame) PC() uintptr {
	return f.pc
}

// Pkg returns the package name associated with the Frame.
// It is the last element of the PkgPath.
func (f Frame) Pkg() string {
	return path.Base(f.PkgPath())
}

// Format implements the fmt.Formatter interface.
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
