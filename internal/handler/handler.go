package handler

import (
	"fmt"
	"io"
	"runtime"

	"github.com/lainio/err2/internal/debug"
)

type (
	PanicHandler func(p any)
	ErrorHandler func(err error)
	NilHandler   func()
)

type Info struct {
	Any any
	W   io.Writer

	NilHandler
	ErrorHandler
	PanicHandler
}

func (i Info) callNilHandler() {
	if i.NilHandler != nil {
		i.NilHandler()
	}
}

func (i Info) callErrorHandler() {
	if i.W != nil {
		si := stackPrologError
		printStack(i.W, si, i.Any)
	}
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	}
}

func (i Info) callPanicHandler() {
	if i.W != nil {
		si := stackPrologPanic
		printStack(i.W, si, i.Any)
	}
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func Process(info Info) {
	switch info.Any.(type) {
	case nil:
		info.callNilHandler()
	case runtime.Error:
		info.callPanicHandler()
	case error:
		info.callErrorHandler()
	default:
		info.callPanicHandler()
	}
}

func printStack(w io.Writer, si debug.StackInfo, msg any) {
	fmt.Fprintf(w, "---\n%v\n---\n", msg)
	debug.FprintStack(w, si)
}

var (
	stackPrologRuntime = newSI("", "panic(", 1)
	stackPrologError   = newErrSI()
	stackPrologPanic   = newSI("", "panic(", 1)
)

func newErrSI() debug.StackInfo {
	return debug.StackInfo{Regexp: debug.PackageRegexp, Level: 1}
}

func newSI(pn, fn string, lvl int) debug.StackInfo {
	return debug.StackInfo{
		PackageName: pn,
		FuncName:    fn,
		Level:       lvl,
		Regexp:      debug.PackageRegexp,
	}
}
