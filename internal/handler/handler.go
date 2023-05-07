// Package handler implements handler for objects returned recovery() function.
package handler

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/lainio/err2/internal/debug"
	fmtstore "github.com/lainio/err2/internal/formatter"
	"github.com/lainio/err2/internal/str"
	"github.com/lainio/err2/internal/tracer"
)

type (
	// we want these to be type aliases, so they are much nicer to use
	PanicHandler = func(p any)
	ErrorHandler = func(err error)
	NilHandler   = func()
)

// Info tells to Process function how to proceed.
type Info struct {
	Any    any    // panic transport object
	Err    *error // error transport pointer (i.e. in/output)
	Format string // format string
	Args   []any  // args for format string printing

	// These are used if handler.Process caller sets them. If they aren't set
	// handler uses package level variables from tracer.
	ErrorTracer io.Writer // If nil tracer package's default is used.
	PanicTracer io.Writer // If nil tracer package's default is used.

	// These are called if handler.Process caller sets it. If they aren't set
	// default implementations are used. NOTE. We have to use both which means
	// that we get nilHandler call if recovery() is called by any other
	// handler then we call still ErrorHandler and get the error from Any. It
	// goes for other way around: we get error but nilHandler is only one to
	// set, we use that for the error (which is accessed from the closure).
	NilHandler   // If nil default implementation is used.
	ErrorHandler // If nil default implementation is used.

	PanicHandler // If nil panic() is called.

	CallerName string
}

const (
	wrapError = ": %w"
)

func PanicNoop(_ any) {}
func NilNoop()        {}

// func ErrorNoop(err error) {}

func (i *Info) callNilHandler() {
	if !i.workToDo() {
		return
	}

	if i.safeErr() != nil {
		i.checkErrorTracer()
	}
	if i.NilHandler != nil {
		i.NilHandler()
	} else {
		i.defaultNilHandler()
	}
}

func (i *Info) checkErrorTracer() {
	if !i.workToDo() {
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
	if !i.workToDo() {
		return
	}

	i.checkErrorTracer()
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	} else {
		i.defaultErrorHandler()
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
	if !i.workToDo() {
		return
	}

	i.checkPanicTracer()
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func (i *Info) defaultNilHandler() {
	err := i.safeErr()
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
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
	if i.workToDo() {
		// error transported thru i.Err not by panic (i.Any)
		// let's allow caller to use ErrorHandler if it's set
		if i.ErrorHandler != nil {
			i.ErrorHandler(err)
			return
		}
	}
}

// defaultErrorHandler is default implementation of handling general errors (not
// runtime.Error which are treated as panics)
//
// Defers are in the stack and the first from the stack gets the opportunity to
// get panic object's error (below). We still must call handler functions to the
// rest of the handlers if there is an error.
func (i *Info) defaultErrorHandler() {
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
	if i.workToDo() {
		// error transported thru i.Err not by panic (i.Any)
		// let's allow caller to use NilHandler if it's set
		if i.NilHandler != nil {
			i.NilHandler()
			return
		}
	}
}

func (i *Info) workToDo() bool {
	return i.safeErr() != nil || i.Any != nil
}

func (i *Info) safeErr() error {
	if i.Err != nil {
		return *i.Err
	}
	return nil
}

// wrapStr returns always wrap string that means we are using "%w" to chain
// errors to be able to use errors.Is and errors.As functions form Go stl.
func (i *Info) wrapStr() string {
	return wrapError
}

// WorkToDo returns if there is something to process. This is offered for
// optimizations. Starting and executing full error handler stack with the
// tracers and other stuff is heavy. This function offers us a API to make the
// decision to abort the processing ASAP.
func WorkToDo(r any, err *error) bool {
	return (err != nil && *err != nil) || r != nil
}

// Process executes error handling logic. Panics and whole defer stack is
// included.
//
// NOTE! That there is an error or a panic to handle i.e. that's taken care.
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

// PreProcess is currently used for err2.Handle.
//
// NOTE! That there is an error or a panic to handle i.e. that's taken care.
//
//nolint:nestif
func PreProcess(info *Info, a ...any) {
	if len(a) > 0 {
		subProcess(info, a...)
	} else {
		// We want the function who sets the handler, i.e. calls the
		// err2.Handle function via defer. Because call stack is in reverse
		// order we need negative, and because the Handle caller is just
		// previous AND funcName can search! This is enough:
		const lvl = -1

		fnName := "Handle"
		if info.CallerName != "" {
			fnName = info.CallerName
		}
		funcName, _, ok := debug.FuncName(debug.StackInfo{
			PackageName: "",
			FuncName:    fnName,
			Level:       lvl,
		})
		if ok {
			fmter := fmtstore.Formatter()
			if fmter != nil { // TODO: check the init order!
				info.Format = fmter.Format(funcName)
			} else {
				info.Format = str.Decamel(funcName)
			}
		}
	}
	if info.PanicHandler == nil && info.CallerName == "Catch" {
		info.PanicHandler = PanicNoop
	}

	Process(info)
}

func subProcess(info *Info, a ...any) {
	switch len(a) {
	case 2:
		processArg(info, 0, a...)
		if _, ok := a[1].(PanicHandler); ok {
			processArg(info, 1, a...)
		}
	default: // more than 2
		processArg(info, 0, a...)
	}
}

func processArg(info *Info, i int, a ...any) {
	switch first := a[i].(type) {
	case string:
		info.Format = first
		info.Args = a[i+1:]
	case ErrorHandler: // err2.Catch uses this
		info.ErrorHandler = first
	case PanicHandler: // err2.Catch uses this
		info.PanicHandler = first
	case NilHandler:
		info.NilHandler = first
	case nil:
		info.NilHandler = NilNoop
	default:
		// we don't panic because we can already be in recovery, but lets
		// try to show an error message at least.
		fmt.Fprintln(os.Stderr, "fatal error: err2.Handle: unsupported type")
	}
}

func printStack(w io.Writer, si debug.StackInfo, msg any) {
	// TODO: if we wanted to use this for unit test time error & panic tracing
	// we should be able to dedect if we are in unit test mode.
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
