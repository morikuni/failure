package failure

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// some code is copied from github.com/pkg/errors.

// CallStack represents call stack.
type CallStack []PC

// Format implements fmt.Formatter.
func (cs CallStack) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, pc := range cs {
				fmt.Fprintf(s, "%+v\n", pc)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []PC(cs))
		default:
			l := len(cs)
			if l == 0 {
				return
			}
			for _, pc := range cs[:l-1] {
				fmt.Fprintf(s, "%s: ", pc.Func())
			}
			fmt.Fprintf(s, "%v", cs[l-1].Func())
		}
	case 's':
		fmt.Fprintf(s, "%v", cs)
	}
}

// Callers returns call stack for the current state.
func Callers(skip int) CallStack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+2, pcs[:])

	cs := make(CallStack, n)
	for i, pc := range pcs[:n] {
		cs[i] = PC(pc)
	}

	return cs
}

// CallStackFromPkgErrors creates CallStack from errors.StackTrace.
func CallStackFromPkgErrors(st errors.StackTrace) CallStack {
	cs := make(CallStack, len(st))
	for i, f := range st {
		cs[i] = PC(f)
	}
	return cs
}

// PC represents program counter.
type PC uintptr

func (pc PC) pc() uintptr {
	// Copied from github.com/pkg/errors.Frame
	// I don't know why add -1 here.
	// If remove this, some test fails.
	return uintptr(pc) - 1
}

// Path returns a full path to the file for pc.
func (pc PC) Path() string {
	fn := runtime.FuncForPC(pc.pc())
	if fn == nil {
		return "???"
	}
	file, _ := fn.FileLine(pc.pc())
	return file
}

// File returns a file name for pc.
func (pc PC) File() string {
	return filepath.Base(pc.Path())
}

// Line returns a line number for pc.
func (pc PC) Line() int {
	fn := runtime.FuncForPC(pc.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(pc.pc())
	return line
}

// Func returns a function name for pc.
func (pc PC) Func() string {
	fn := runtime.FuncForPC(pc.pc())
	if fn == nil {
		return ""
	}
	fs := strings.Split(path.Base(fn.Name()), ".")
	if len(fs) >= 1 {
		return strings.Join(fs[1:], ".")
	}
	return fs[0]
}

// Pkg returns a package name for pc.
func (pc PC) Pkg() string {
	fn := runtime.FuncForPC(pc.pc())
	if fn == nil {
		return ""
	}
	fs := strings.Split(path.Base(fn.Name()), ".")
	return fs[0]
}

// Format implements fmt.Formatter.
func (pc PC) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "[%s] ", pc.Func())
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s:%d", pc.Path(), pc.Line())
	}
}
