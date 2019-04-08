package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// CallStack represents a call stack.
// The call stack includes information where the occurred and
// how the function was called.
type CallStack interface {
	// HeadFrame returns a frame of where call stack is created.
	// This is same as Frames()[0], but uses memory more efficiently.
	HeadFrame() Frame
	// Frames returns entire frames of the call stack.
	Frames() []Frame
}

type callStack struct {
	pcs []uintptr
}

func (cs callStack) HeadFrame() Frame {
	if len(cs.pcs) == 0 {
		return emptyFrame
	}

	rfs := runtime.CallersFrames(cs.pcs[:1])
	f, _ := rfs.Next()
	return frame{f.File, f.Line, f.Function, f.PC}
}

func (cs callStack) Frames() []Frame {
	if len(cs.pcs) == 0 {
		return nil
	}

	rfs := runtime.CallersFrames(cs.pcs)

	fs := make([]Frame, 0, len(cs.pcs))
	for {
		f, more := rfs.Next()

		fs = append(fs, frame{f.File, f.Line, f.Function, f.PC})

		if !more {
			break
		}
	}
	return fs
}

func (cs callStack) Format(s fmt.State, verb rune) {
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
	return callStack{pcs}
}

// Callers returns a call stack for the current state.
func Callers(skip int) CallStack {
	var pcs [32]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return nil
	}

	return NewCallStack(pcs[:n])
}

// Frame represents a stack frame.
type Frame interface {
	// Path returns a full path to the file.
	Path() string
	// File returns a file name.
	File() string
	// Line returns a line number in the file.
	Line() int
	// Func returns a function name.
	Func() string
	// Pkg returns a package name of the function.
	Pkg() string
	// PC returns a program counter of this frame.
	PC() uintptr
}

var emptyFrame = frame{"???", 0, "???", uintptr(0)}

type frame struct {
	file     string
	line     int
	function string
	pc       uintptr
}

func (f frame) Path() string {
	return f.file
}

func (f frame) File() string {
	return filepath.Base(f.file)
}

func (f frame) Line() int {
	return f.line
}

func (f frame) Func() string {
	fs := strings.Split(path.Base(f.function), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

func (f frame) PC() uintptr {
	return f.pc
}

func (f frame) Pkg() string {
	fs := strings.Split(path.Base(f.function), ".")
	return fs[0]
}

func (f frame) Format(s fmt.State, verb rune) {
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
