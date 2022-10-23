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
	Any    any    // panic transport object
	Err    *error // error transport pointer (i.e. in/output)
	Format string // format string
	Args   []any  // ags for format string printing
	Wrap   bool   // if true error wrapping "%w" is used, default is "%v"

	ErrorTracer io.Writer
	PanicTracer io.Writer

	NilHandler
	ErrorHandler
	PanicHandler
}

const (
	wrapAnnot = ": %v"
	wrapError = ": %w"
)

func PanicNoop(v any)     {}
func ErrorNoop(err error) {}
func NilNoop()            {}

func (i Info) callNilHandler() {
	if i.Err != nil {
		i.checkErrorTracer()
	}
	if i.NilHandler != nil {
		i.NilHandler()
	} else {
		i.nilHandler()
	}
}

func (i Info) checkErrorTracer() {
	if i.ErrorTracer == nil {
		i.ErrorTracer = tracer.Error.Tracer()
	}
	if i.ErrorTracer != nil {
		si := stackPrologueError
		if i.Any == nil {
			i.Any = *i.Err
		}
		printStack(i.ErrorTracer, si, i.Any)
	}
}

func (i Info) callErrorHandler() {
	i.checkErrorTracer()
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	} else {
		i.errorHandler()
	}
}

func (i Info) checkPanicTracer() {
	if i.PanicTracer == nil {
		i.PanicTracer = tracer.Panic.Tracer()
	}
	if i.PanicTracer != nil {
		si := stackProloguePanic
		printStack(i.PanicTracer, si, i.Any)
	}
}

func (i Info) callPanicHandler() {
	i.checkPanicTracer()
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func (i Info) nilHandler() {
	err := *i.Err
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
			return
		}
	} else {
		// error transported thru i.Err not by panic (i.Any)
		// let's give caller to use ErrorHandler if it's set
		if i.ErrorHandler != nil {
			i.ErrorHandler(err)
			return
		}
	}
	if err != nil {
		if i.Format != "" {
			*i.Err = fmt.Errorf(i.Format+i.WrapStr(), append(i.Args, err)...)
		} else {
			*i.Err = err
		}
	}
}

func (i Info) errorHandler() {
	err := *i.Err
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
			return
		}
	}
	if i.Format != "" {
		*i.Err = fmt.Errorf(i.Format+i.WrapStr(), append(i.Args, err)...)
	} else {
		*i.Err = err
	}
}

func (i Info) WrapStr() string {
	if i.Wrap {
		return wrapError
	}
	return wrapAnnot
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
