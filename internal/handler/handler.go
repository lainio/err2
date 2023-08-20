// Package handler implements handler for objects returned recovery() function.
package handler

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/lainio/err2/internal/color"
	"github.com/lainio/err2/internal/debug"
	fmtstore "github.com/lainio/err2/internal/formatter"
	"github.com/lainio/err2/internal/str"
	"github.com/lainio/err2/internal/tracer"
	"github.com/lainio/err2/internal/x"
)

type (
	// we want these to be type aliases, so they are much nicer to use
	PanicHandler = func(p any)
	ErrorHandler = func(err error) error // this is only proper type that work
	NilHandler   = func(err error) error // these two are the same

	//CheckHandler = func(noerr bool, err error) error
	CheckHandler = func(noerr bool)
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
	ErrorHandler // If nil default implementation is used.
	NilHandler   // If nil default (pre-defined here) implementation is used.

	PanicHandler // If nil panic() is called.

	CheckHandler // this would be for cases where there isn't any error, but
	// this should be the last defer.

	CallerName string

	werr error
}

const (
	wrapError = ": %w"
)

func PanicNoop(any)           {}
func NilNoop(err error) error { return err }

// func ErrorNoop(err error) {}

func (i *Info) callNilHandler() {
	if i.CheckHandler != nil && i.safeErr() == nil {
		i.CheckHandler(true)
		// there is no err and user wants to handle OK with our pkg:
		// nothing more to do here after callNilHandler call
		return
	}

	if i.safeErr() != nil {
		i.checkErrorTracer()
	}
	if i.NilHandler != nil {
		*i.Err = i.NilHandler(i.werr)
		i.werr = *i.Err // remember change both our errors!
	} else {
		i.defaultNilHandler()
	}
}

func (i *Info) checkErrorTracer() {
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
	i.checkErrorTracer()
	if i.ErrorHandler != nil {
		*i.Err = i.ErrorHandler(i.Any.(error))
		i.werr = *i.Err // remember change both our errors!
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
	i.checkPanicTracer()
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func (i *Info) workError() (err error) {
	err = i.safeErr()
	if err == nil {
		var ok bool
		err, ok = i.Any.(error)
		if !ok {
			return nil
		}
	}
	return err
}

func (i *Info) fmtErr() {
	*i.Err = fmt.Errorf(i.Format+i.wrapStr(), append(i.Args, i.werr)...)
	i.werr = *i.Err // remember change both our errors!
}

func (i *Info) buildFmtErr() {
	if i.Format != "" {
		i.fmtErr()
		return
	}
	*i.Err = i.werr
}

func (i *Info) safeCallErrorHandler() {
	if i.ErrorHandler != nil {
		*i.Err = i.ErrorHandler(i.werr)
	}
}

func (i *Info) defaultNilHandler() {
	i.werr = i.workError()
	if i.werr == nil {
		return
	}
	i.buildFmtErr()
	if i.workToDo() {
		// error transported thru i.Err not by panic (i.Any)
		// let's allow caller to use ErrorHandler if it's set
		i.safeCallErrorHandler()
	}
}

func (i *Info) safeCallNilHandler() {
	if i.NilHandler != nil {
		*i.Err = i.NilHandler(i.werr)
	}
}

// defaultErrorHandler is default implementation of handling general errors (not
// runtime.Error which are treated as panics)
//
// Defers are in the stack and the first from the stack gets the opportunity to
// get panic object's error (below). We still must call handler functions to the
// rest of the handlers if there is an error.
func (i *Info) defaultErrorHandler() {
	i.werr = i.workError()
	if i.werr == nil {
		return
	}
	i.buildFmtErr()
	if i.workToDo() {
		// error transported thru i.Err not by panic (i.Any)
		// let's allow caller to use NilHandler if it's set
		i.safeCallNilHandler()
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

func NoerrCallToDo(a ...any) (yes bool) {
	//var yes bool
	if len(a) != 0 {
		_, yes = a[0].(CheckHandler)
	}
	return yes
}

// Process executes error handling logic. Panics and whole defer stack is
// included.
//
//   - err2.API functions call PreProcess and Process is exported for tests.
//   - That there is an error or a panic to handle i.e. that's taken care.
func Process(info *Info) {
	switch info.Any.(type) {
	case nil:
		info.callNilHandler()
	case runtime.Error: // need own or handled like errors
		info.callPanicHandler()
	case error:
		info.callErrorHandler()
	default:
		info.callPanicHandler()
	}
}

// PreProcess is currently used for err2 API like err2.Handle and .Catch.
//   - replaces the Process
func PreProcess(errPtr *error, info *Info, a ...any) error {
	// Bug in Go?
	// start to use local error ptr only for optimization reasons.
	// We get 3x faster defer handlers without unsing ptr to original err
	// named return val. Reason is unknown.
	err := x.Whom(errPtr != nil, *errPtr, nil)
	info.Err = &err
	info.werr = *info.Err // remember change both our errors!

	// We want the function who sets the handler, i.e. calls the
	// err2.Handle function via defer. Because call stack is in reverse
	// order we need negative, and because the Handle caller is just
	// previous AND funcName can search! This is enough:
	const lvl = -1

	if len(a) > 0 {
		subProcess(info, a...)
	} else {
		fnName := "Handle"
		if info.CallerName != "" {
			fnName = info.CallerName
		}
		funcName, _, _, ok := debug.FuncName(debug.StackInfo{
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
	defCatchCallMode := info.PanicHandler == nil && info.CallerName == "Catch"
	if defCatchCallMode {
		info.PanicHandler = PanicNoop
	}

	Process(info)

	logCatchCallMode := defCatchCallMode && firstArgIsString(a...)
	if curErr := info.safeErr(); logCatchCallMode && curErr != nil {
		_, _, frame, ok := debug.FuncName(debug.StackInfo{
			PackageName: "",
			FuncName:    "Catch",
			Level:       lvl,
		})
		const framesToSkip = 6
		frame = x.Whom(ok, frame, framesToSkip)
		if LogOutput(frame, curErr.Error()) == nil {
			*info.Err = nil // prevent dublicate "logging"
		}
	}
	return err
}

func firstArgIsString(a ...any) bool {
	if len(a) > 0 {
		_, isStr := a[0].(string)
		return isStr
	}
	return false
}

func subProcess(info *Info, a ...any) {
	switch len(a) {
	case 2: // currently we support only this order of 2 handlers in Catch
		processArg(info, 0, a...)
		if _, ok := a[1].(PanicHandler); ok {
			processArg(info, 1, a...)
		}
	default:
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
		info.NilHandler = first
	case PanicHandler: // err2.Catch uses this
		info.PanicHandler = first
	case CheckHandler:
		info.CheckHandler = first
	case nil:
		info.NilHandler = NilNoop
	default:
		// we don't panic here because we can already be in recovery, but lets
		// try to show an RED error message at least.
		const msg = `err2 fatal error:  
---
unsupported handler function type: err2.Handle/Catch:
see 'err2/scripts/README.md' and run auto-migration scripts for your repo
---`
		fmt.Fprintln(os.Stderr, color.Red()+msg+color.Reset())
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

func LogOutput(lvl int, s string) (err error) {
	w := tracer.Log.Tracer()
	if w == nil {
		return log.Output(lvl, s)
	}
	fmt.Fprintln(w, s)
	return nil
}
