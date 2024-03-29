package err2

import (
	"errors"
	"fmt"

	"github.com/lainio/err2/internal/handler"
)

type (
	// Handler is a function type used to process error values in [Handle] and
	// [Catch]. We currently have a few build-ins of the Handler: [Noop],
	// [Reset], etc.
	Handler = handler.ErrorFn
)

// Sentinel error value helpers. They are convenient thanks to
// [github.com/lainio/err2/try.IsNotFound] and similar functions.
//
// [ErrNotFound] ... [ErrNotEnabled] are similar no-error like [io.EOF] for
// those who really want to use error return values to transport non errors.
// It's far better to have discriminated unions as errors for function calls.
// But if you insist the related helpers are in they
// [github.com/lainio/err2/try] package:
// [github.com/lainio/err2/try.IsNotFound], ...
//
// [ErrRecoverable] and [ErrNotRecoverable] since Go 1.20 wraps multiple errors
// same time, i.e. wrapped errors aren't list anymore but tree. This allows mark
// multiple semantics to same error. These error are mainly for that purpose.
var (
	ErrNotFound       = errors.New("not found")
	ErrNotExist       = errors.New("not exist")
	ErrAlreadyExist   = errors.New("already exist")
	ErrNotAccess      = errors.New("permission denied")
	ErrNotEnabled     = errors.New("not enabled")
	ErrNotRecoverable = errors.New("cannot recover")
	ErrRecoverable    = errors.New("recoverable")
)

// Stdnull implements [io.Writer] that writes nothing, e.g.,
// [SetLogTracer] in cases you don't want to use automatic log writer (=nil),
// i.e., [LogTracer] == /dev/null. It can be used to change how the [Catch]
// works, e.g., in CLI apps.
var Stdnull = &nullDev{}

// Handle is the general purpose error handling function. What makes it so
// convenient is its ability to handle all error handling cases:
//   - just return the error value to caller
//   - annotate the error value
//   - execute real error handling like cleanup and releasing resources.
//
// There's no performance penalty. The handler is called only when err != nil.
// There's no limit how many Handle functions can be added to defer stack. They
// all are called if an error has occurred.
//
// The function has an automatic mode where errors are annotated by function
// name if no annotation arguments or handler function is given:
//
//	func SaveData(...) (err error) {
//	     defer err2.Handle(&err) // if err != nil: annotation is "save data:"
//
// Note. If you are still using sentinel errors you must be careful with the
// automatic error annotation because it uses wrapping. If you must keep the
// error value got from error checks: [github.com/lainio/err2/try.To], you must
// disable automatic error annotation (%w), or set the returned error values in
// the handler function. Disabling can be done by setting second argument nil:
//
//	func SaveData(...) (err error) {
//	     defer err2.Handle(&err, nil) // nil arg disable automatic annotation.
//
// In case of the actual error handling, the handler function should be given as
// a second argument:
//
//	defer err2.Handle(&err, func(err error) error {
//	     if rmErr := os.Remove(dst); rmErr != nil {
//	          return fmt.Errorf("%w: cleanup error: %w", err, rmErr)
//	     }
//	     return err
//	})
//
// You can have unlimited amount of error handlers. They are called if error
// happens and they are called in the same order as they are given or until one
// of them resets the error like [Reset] (notice the other predefined error
// handlers) in the next samples:
//
//	defer err2.Handle(&err, err2.Reset, err2.Log) // Log not called
//	defer err2.Handle(&err, err2.Noop, err2.Log) // handlers > 1: err annotated
//	defer err2.Handle(&err, nil, err2.Log) // nil disables auto-annotation
//
// If you need to stop general panics in a handler, you can do that by declaring
// a panic handler. See the second handler below:
//
//	defer err2.Handle(&err,
//	     err2.Err( func(error) { os.Remove(dst) }), // err2.Err() keeps it short
//	     // below handler catches panics, but you can re-throw if needed
//	     func(p any) {}
//	)
func Handle(err *error, a ...any) {
	// This and others are similar but we need to call `recover` here because
	// how how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) && !handler.NoerrCallToDo(a) {
		return
	}

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	*err = handler.PreProcess(err, &handler.Info{
		CallerName: "Handle",
		Any:        r,
	}, a)
}

// Catch is a convenient helper to those functions that doesn't return errors.
// Note that Catch always catch the panics. If you don't want to stop them
// (i.e., use of [recover]) you should add panic handler and continue panicking
// there. There can be only one deferred Catch function per non error returning
// functions, i.e. goroutine functions like main(). There is several ways to use
// the Catch function. And always remember the [defer].
//
// The deferred Catch is very convenient, because it makes your current
// goroutine panic and error-safe. You can fine tune its 'global' behavior with
// functions like [SetErrorTracer], [SetPanicTracer], and [SetLogTracer]. Its
// 'local' behavior depends the arguments you give it. Let's start with the
// defaults and simplest version of Catch:
//
//	defer err2.Catch()
//
// In default the above writes errors to logs and panic traces to stderr.
// Naturally, you can annotate logging:
//
//	defer err2.Catch("WARNING: caught errors: %s", name)
//
// The preceding line catches the errors and panics and prints an annotated
// error message about the error source (from where the error was thrown) to the
// currently set log. Note, when log stream isn't set, the standard log is used.
// It can be bound to, e.g., glog. And if you want to suppress automatic logging
// entirely use the following setup:
//
//	err2.SetLogTracer(err2.Stdnull)
//
// The next one stops errors and panics, but allows you handle errors, like
// cleanups, etc. The error handler function has same signature as Handle's
// error handling function [Handler]. By returning nil resets the
// error, which allows e.g. prevent automatic error logs to happening.
// Otherwise, the output results depends on the current trace and assert
// settings. The default trace setting prints call stacks for panics but not for
// errors:
//
//	defer err2.Catch(func(err error) error { return err} )
//
// or if you you prefer to use dedicated helpers:
//
//	defer err2.Catch(err2.Noop)
//
// You can give unlimited amount of error handlers. They are called if error
// happens and they are called in the same order as they are given or until one
// of them resets the error like [Reset] in the next sample:
//
//	defer err2.Catch(err2.Noop, err2.Reset, err2.Log) // err2.Log not called!
//
// The next sample calls your error handler, and you have an explicit panic
// handler as well, where you can e.g. continue panicking to propagate it for
// above callers or stop it like below:
//
//	defer err2.Catch(func(err error) error { return err }, func(p any) {})
func Catch(a ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, nil) {
		return
	}

	var err error
	err = handler.PreProcess(&err, &handler.Info{
		CallerName: "Catch",
		Any:        r,
	}, a)
	doTrace(err)
}

// Throwf builds and throws an error (panic). For creation it's similar to
// [fmt.Errorf]. Because panic is used to transport the error instead of error
// return value, it's called only if you want to non-local control structure for
// error handling, i.e. your current function doesn't have error return value.
//
//   - Throwf is rarely needed. We suggest to use error return values instead.
//
// Throwf is offered for deep recursive algorithms to help readability and
// performance (see bechmarks) in those cases.
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

type nullDev struct{}

func (nullDev) Write([]byte) (int, error) { return 0, nil }

func doTrace(err error) {
	if err == nil || err.Error() == "" {
		return
	}
	if ErrorTracer() != nil {
		fmt.Fprintln(ErrorTracer(), err.Error())
	} else if PanicTracer() != nil {
		fmt.Fprintln(PanicTracer(), err.Error())
	}
}
