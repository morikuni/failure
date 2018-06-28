package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// CallStack represents call stack.
type CallStack interface {
	HeadFrame() Frame
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
	return frame{f.File, f.Line, f.Function}
}

func (cs callStack) Frames() []Frame {
	if len(cs.pcs) == 0 {
		return nil
	}

	rfs := runtime.CallersFrames(cs.pcs)

	fs := make([]Frame, 0, len(cs.pcs))
	for {
		f, more := rfs.Next()

		fs = append(fs, frame{f.File, f.Line, f.Function})

		if !more {
			break
		}
	}
	return fs
}

// Format implements fmt.Formatter.
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
				fmt.Fprintf(s, "%s: ", f.Func())
			}
			fmt.Fprintf(s, "%v", fs[l-1].Func())
		}
	case 's':
		fmt.Fprintf(s, "%v", cs)
	}
}

// Callers returns call stack for the current state.
func Callers(skip int) CallStack {
	var pcs [32]uintptr
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return nil
	}

	return callStack{pcs[:n]}
}

// CallStackFromPkgErrors creates CallStack from errors.StackTrace.
func CallStackFromPkgErrors(st errors.StackTrace) CallStack {
	pcs := make([]uintptr, len(st))
	for i, v := range st {
		pcs[i] = uintptr(v)
	}

	return callStack{[]uintptr(pcs)}
}

type Frame interface {
	Path() string
	File() string
	Line() int
	Func() string
	Pkg() string
}

var emptyFrame = frame{"???", 0, "???"}

type frame struct {
	file     string
	line     int
	function string
}

// Path returns a full path to the file for pc.
func (f frame) Path() string {
	return f.file
}

// File returns a file name for pc.
func (f frame) File() string {
	return filepath.Base(f.file)
}

// Line returns a line number for pc.
func (f frame) Line() int {
	return f.line
}

// Func returns a function name for pc.
func (f frame) Func() string {
	fs := strings.Split(path.Base(f.function), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

// Pkg returns a package name for pc.
func (f frame) Pkg() string {
	fs := strings.Split(path.Base(f.function), ".")
	return fs[0]
}

// Format implements fmt.Formatter.
func (f frame) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "[%s] ", f.Func())
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s:%d", f.Path(), f.Line())
	}
}
