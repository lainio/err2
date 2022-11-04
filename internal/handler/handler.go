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

// Info tells to Process function how to proceed.
type Info struct {
	Any    any    // panic transport object
	Err    *error // error transport pointer (i.e. in/output)
	Format string // format string
	Args   []any  // args for format string printing
	Wrap   bool   // if true error wrapping "%w" is used, default is "%v"

	// These are used if handler.Process caller sets them. If they aren't set
	// handler uses package level variables from tracer.
	ErrorTracer io.Writer // If nil tracer package's default is used.
	PanicTracer io.Writer // If nil tracer package's default is used.

	// These are called if handler.Process caller sets it. If they aren't set
	// default implementations are used.
	NilHandler   // If nil default implementation is used.
	ErrorHandler // If nil default implementation is used.
	PanicHandler // If nil panic() is called.
}

const (
	wrapAnnot = ": %v"
	wrapError = ": %w"
)

func PanicNoop(v any) {}
func NilNoop()        {}

// func ErrorNoop(err error) {}

func (i *Info) callNilHandler() {
	if !i.WorkToDo() {
		return
	}

	if i.safeErr() != nil {
		i.checkErrorTracer()
	}
	if i.NilHandler != nil {
		i.NilHandler()
	} else {
		i.nilHandler()
	}
}

func (i *Info) checkErrorTracer() {
	if !i.WorkToDo() {
		return
	}

	if i.ErrorTracer == nil {
		i.ErrorTracer = tracer.Error.Tracer()
	}
	if i.ErrorTracer != nil {
		si := stackPrologueError
		if i.Any == nil {
			i.Any = i.safeErr()
		}
		printStack(i.ErrorTracer, si, i.Any)
	}
}

func (i *Info) callErrorHandler() {
	if !i.WorkToDo() {
		return
	}

	i.checkErrorTracer()
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	} else {
		i.errorHandler()
	}
}

func (i *Info) checkPanicTracer() {
	if i.PanicTracer == nil {
		i.PanicTracer = tracer.Panic.Tracer()
	}
	if i.PanicTracer != nil {
		si := stackProloguePanic
		printStack(i.PanicTracer, si, i.Any)
	}
}

func (i *Info) callPanicHandler() {
	if !i.WorkToDo() {
		return
	}

	i.checkPanicTracer()
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func (i *Info) nilHandler() {
	err := i.safeErr()
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
			return
		}
	} else {
		// error transported thru i.Err not by panic (i.Any)
		// let's allow caller to use ErrorHandler if it's set
		if i.ErrorHandler != nil {
			i.ErrorHandler(err)
			return
		}
	}
	if err != nil {
		if i.Format != "" {
			*i.Err = fmt.Errorf(i.Format+i.wrapStr(), append(i.Args, err)...)
		} else {
			*i.Err = err
		}
	}
}

// errorHandler is default implementation of handling general errors (not
// runtime.Error which are treated as panics)
//
// Defers are in the stack and the first from the stack gets the opportunity to
// get panic object's error (below). We still must call handler functions to the
// rest of the handlers if there is an error.
func (i *Info) errorHandler() {
	err := i.safeErr()
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
			return
		}
	}
	if i.Format != "" {
		*i.Err = fmt.Errorf(i.Format+i.wrapStr(), append(i.Args, err)...)
	} else {
		*i.Err = err
	}
}

func (i *Info) WorkToDo() bool {
	return i.safeErr() != nil || i.Any != nil
}

func (i *Info) safeErr() error {
	if i.Err != nil {
		return *i.Err
	}
	return nil
}

func (i *Info) wrapStr() string {
	if i.Wrap {
		return wrapError
	}
	return wrapAnnot
}

func WorkToDo(r any, err *error) bool {
	return (err != nil && *err != nil) || r != nil
}

// Process executes error handling logic. Panics and whole defer stack is
// included.
func Process(info *Info) {
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
