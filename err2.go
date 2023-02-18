package err2

import (
	"errors"
	"fmt"
	"os"

	"github.com/lainio/err2/internal/handler"
)

// nolint
var (
	// NotFound is similar *no-error* like io.EOF for those who really want to
	// use error return values to transport non errors. It's far better to have
	// discriminated unions as errors for function calls. But if you insist the
	// related helpers are in they try package: try.IsNotFound(), ... These
	// 'global' errors and their helper functions in try package are for
	// experimenting now.
	NotFound  = errors.New("not found")
	NotExist  = errors.New("not exist")
	Exist     = errors.New("already exist")
	NotAccess = errors.New("permission denied")

	NotRecoverable = errors.New("cannot recover")
	Recoverable    = errors.New("recoverable")
)

// Handle is the general purpose error handling helper. What makes it so
// convenient is its ability to handle all error handling cases: a) just
// return the error value to caller, b) annotate the error value, or c) execute
// real error handling like cleanup and releasing resources. There is no
// performance penalty. The handler is called only when err != nil. There is no
// limit how many Handle functions can be added to defer stack. They all are
// called if an error has occurred and they are in deferred.
//
// The function has an automatic mode where errors are annotated by function
// name if no annotation arguments or handler function is given:
//
//	func SaveData(...) (err error) {
//	     defer err2.Handle(&err) // if err != nil: annotation is "save data:"
//
// Note. If you are still using sentinel errors you must be careful with the
// automatic error annotation because it uses wrapping. If you must keep the
// error value got from error checks: 'try.To(..)', you must disable automatic
// error annotation (%w), or set the returned error values in the handler
// function. Disabling can be done by setting second argument nil:
//
//	func SaveData(...) (err error) {
//	     defer err2.Handle(&err, nil) // nil arg disable automatic annotation.
//
// In case of the actual error handling, the handler function should be given as
// an second argument:
//
//	defer err2.Handle(&err, func() {
//		os.Remove(dst)
//	})
func Handle(err *error, a ...any) {
	// This and others are similar but we need to call `recover` here because
	// how how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) {
		return
	}

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	handler.PreProcess(&handler.Info{
		CallerName: "Handle",
		Any:        r,
		Err:        err,
	}, a...)
}

// Catch is a convenient helper to those functions that doesn't return errors.
// There can be only one deferred Catch function per non error returning
// function like main(). There is several ways to make deferred calls to Catch.
//
//	defer err2.Catch()
//
// This stops errors and panics, and output depends on the current Tracer
// settings.
//
//	defer err2.Catch(func(err error) {})
//
// This one calls your error handler. You could have only panic handler, but
// that's unusual. Only if you are sure that errors are handled you should do
// that. In most cases if you need to stop panics you should have both:
//
//	defer err2.Catch(func(err error) {}, func(p any) {})
func Catch(a ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, nil) {
		return
	}

	var err error
	handler.PreProcess(&handler.Info{
		CallerName: "Catch",
		Any:        r,
		NilHandler: handler.NilNoop,
		Err:        &err,
	}, a...)
	doTrace(err)
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine. It is preferred helper for go
// workers on long running servers, because they stop panics as well.
//
// Note, if any Tracer is set stack traces are printed automatically. If you
// want to do it in the handlers by yourself, auto tracers should be nil.
// Deprecated: use Catch for everything
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, nil) {
		return
	}

	handler.Process(&handler.Info{
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: panicHandler,
		NilHandler:   handler.NilNoop,
	})
}

// CatchTrace is a helper function to catch and handle all errors. It also
// recovers a panic and prints its call stack. CatchTrace and CatchAll are
// preferred helpers for go-workers on long-running servers because they stop
// panics as well.
//
// CatchTrace prints only panic and runtime.Error stack trace if ErrorTracer
// isn't set. If it's set it prints both. The panic trace is printed to stderr.
// If you need panic trace to be printed to some other io.Writer than os.Stderr,
// you should use CatchAll or Catch with tracers.
//
// Deprecated: Use err2.Catch() and err2.SetPanicTracer() together instead.
func CatchTrace(errorHandler func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, nil) {
		return
	}

	handler.Process(&handler.Info{
		PanicTracer:  os.Stderr,
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: handler.PanicNoop, // no rethrow
		NilHandler:   handler.NilNoop,
	})
}

// Throwf builds and throws (panics) an error. For creation it's similar to
// fmt.Errorf. Because panic is used to transport the error instead of error
// return value, it's called only if you want to non-local control structure for
// error handling, i.e. your current function doesn't have error return value.
// NOTE, Throwf is rarely needed. We suggest to use error return values instead.
// Throwf is offered for deep recursive algorithms to help readability.
//
//	func yourFn() (res any) {
//	     ...
//	     if badHappens {
//	          err2.Throwf("we cannot do that for %v", subject)
//	     }
//	     ...
//	}
func Throwf(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	panic(err)
}

// Return is the same as Handle but it's for functions that don't wrap or
// annotate their errors. It's still needed to break panicking which is used for
// error transport in err2. If you want to annotate errors see Returnf and
// Returnw functions for more information.
//
// Deprecated: use err2.Handle instead.
func Return(err *error) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) {
		return
	}

	info := &handler.Info{
		Any:          r,
		Err:          err,
		ErrorHandler: func(e error) { *err = e },
	}
	handler.Process(info)
}

// Returnw wraps an error with '%w'. It's similar to fmt.Errorf, but it's called
// only if error != nil. If you don't want to wrap the error use Handle
// instead.
//
// Deprecated: use err2.Handle instead.
func Returnw(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) {
		return
	}

	handler.Process(&handler.Info{
		Any:    r,
		Err:    err,
		Format: format,
		Args:   args,
	})
}

// Returnf builds an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. It uses '%v' to wrap the error not '%w'. Use Returnw for that.
//
// Deprecated: use err2.Handle instead.
func Returnf(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) {
		return
	}

	handler.Process(&handler.Info{
		Any:    r,
		Err:    err,
		Format: format,
		Args:   args,
	})
}

func doTrace(err error) {
	if err == nil || err.Error() == "" {
		return
	}
	if ErrorTracer() != nil {
		fmt.Fprint(ErrorTracer(), err.Error())
	} else if PanicTracer() != nil {
		fmt.Fprint(PanicTracer(), err.Error())
	}
}
