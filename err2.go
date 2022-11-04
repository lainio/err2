package err2

import (
	"errors"
	"fmt"
	"os"

	"github.com/lainio/err2/internal/handler"
)

//nolint:stylecheck
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
)

// Handle is for adding an error handler to a function by deferring. It's for
// functions returning errors themself. For those functions that don't return
// errors, there is a CatchXxxx functions. The handler is called only when err
// != nil. There is no limit how many Handle functions can be added to defer
// stack. They all are called if an error has occurred and they are in deferred.
func Handle(err *error, handlerFn func()) {
	// This and others are similar but we need to call `recover` here because
	// how how it works with defer.
	r := recover()

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	handler.Process(&handler.Info{
		Any:        r,
		Err:        err,
		NilHandler: handlerFn,
		ErrorHandler: func(e error) {
			// We or someone did transport this error thru panic.
			*err = e
			handlerFn()
		},
	})
}

// Catch is a convenient helper to those functions that doesn't return errors.
// There can be only one deferred Catch function per non error returning
// function like main(). It doesn't catch panics and runtime errors. If that's
// important use CatchAll or CatchTrace instead. See Handle for more
// information.
func Catch(f func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(&handler.Info{
		Any:          r,
		ErrorHandler: f,
		NilHandler:   handler.NilNoop,
	})
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine. It and CatchTrace are preferred
// helpers for go workers on long running servers, because they stop panics as
// well.
//
// Note, if any Tracer is set stack traces are printed automatically. If you
// want to do it in the handlers by yourself, auto tracers should be nil.
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

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
func CatchTrace(errorHandler func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

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
// only if error != nil. If you don't want to wrap the error use Returnf
// instead.
func Returnw(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(&handler.Info{
		Any:    r,
		Err:    err,
		Format: format,
		Args:   args,
		Wrap:   true,
	})
}

// Returnf builds an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. It uses '%v' to wrap the error not '%w'. Use Returnw for that.
func Returnf(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(&handler.Info{
		Any:    r,
		Err:    err,
		Format: format,
		Args:   args,
	})
}
