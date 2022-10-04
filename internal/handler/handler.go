// Package handler implements handler for objects returned recovery() function.
package handler

import (
	"fmt"
	"io"
	"runtime"

	"github.com/lainio/err2/internal/debug"
	"github.com/lainio/err2/internal/tracer"
)

type (
	PanicHandler func(p any)
	ErrorHandler func(err error)
	NilHandler   func()
)

type Info struct {
	Any         any
	ErrorTracer io.Writer
	PanicTracer io.Writer

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
	if i.ErrorTracer == nil {
		i.ErrorTracer = tracer.Error.Tracer()
	}
	if i.ErrorTracer != nil {
		si := stackPrologueError
		printStack(i.ErrorTracer, si, i.Any)
	}
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	}
}

func (i Info) callPanicHandler() {
	if i.PanicTracer == nil {
		i.PanicTracer = tracer.Panic.Tracer()
	}
	if i.PanicTracer != nil {
		si := stackProloguePanic
		printStack(i.PanicTracer, si, i.Any)
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
	// stackPrologueRuntime = newSI("", "panic(", 1)
	stackPrologueError = newErrSI()
	stackProloguePanic = newSI("", "panic(", 1)
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
