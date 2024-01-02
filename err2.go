package err2

import (
	"errors"
	"fmt"
	"os"

	"github.com/lainio/err2/internal/handler"
)

type (
	// Handler is a function type used to process error values in Handle and
	// Catch. We currently have a few build-ins of the Handler: err2.Noop,
	// err2.Reset, etc.
	Handler func(error) error
)

var (
	// ErrNotFound is similar *no-error* like io.EOF for those who really want to
	// use error return values to transport non errors. It's far better to have
	// discriminated unions as errors for function calls. But if you insist the
	// related helpers are in they try package: try.IsNotFound(), ... These
	// 'global' errors and their helper functions in try package are for
	// experimenting now.
	ErrNotFound     = errors.New("not found")
	ErrNotExist     = errors.New("not exist")
	ErrAlreadyExist = errors.New("already exist")
	ErrNotAccess    = errors.New("permission denied")
	ErrNotEnabled   = errors.New("not enabled")

	// Since Go 1.20 wraps multiple errors same time, i.e. wrapped errors
	// aren't list anymore but tree. This allows mark multiple semantics to
	// same error. These error are mainly for that purpose.
	ErrNotRecoverable = errors.New("cannot recover")
	ErrRecoverable    = errors.New("recoverable")
)

// Handle is the general purpose error handling function. What makes it so
// convenient is its ability to handle all error handling cases:
//   - just return the error value to caller
//   - annotate the error value
//   - execute real error handling like cleanup and releasing resources.
//
// There is no performance penalty. The handler is called only when err != nil.
// There is no limit how many Handle functions can be added to defer stack. They
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
//	defer err2.Handle(&err, func(err error) error {
//		if rmErr := os.Remove(dst); rmErr != nil {
//			return fmt.Errorf("%w: cleanup error: %w", err, rmErr)
//		}
//		return err
//	})
//
// If you need to stop general panics in handler, you can do that by giving a
// panic handler function:
//
//	defer err2.Handle(&err,
//	   err2.Err( func(error) { os.Remove(dst) }), // err2.Err() keeps it short
//	   func(p any) {} // <- handler stops panics, re-throw or not
//	)
func Handle(err *error, a ...any) {
	// This and others are similar but we need to call `recover` here because
	// how how it works with defer.
	r := recover()

	if !handler.WorkToDo(r, err) && !handler.NoerrCallToDo(a...) {
		return
	}

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	*err = handler.PreProcess(err, &handler.Info{
		CallerName: "Handle",
		Any:        r,
	}, a...)
}

// Catch is a convenient helper to those functions that doesn't return errors.
// Note, that Catch always catch the panics. If you don't want to stop them
// (recover) you should add panic handler and continue panicking there. There
// can be only one deferred Catch function per non error returning function like
// main(). There is several ways to use the Catch function. And always remember
// the defer.
//
// The deferred Catch is very convenient, because it makes your current
// goroutine panic and error-safe, one line only! You can fine tune its
// behavior with functions like err2.SetErrorTrace, assert.SetDefault, and
// logging settings. Start with the defaults and simplest version of Catch:
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
// currently set log.
//
// The next one stops errors and panics, but allows you handle errors, like
// cleanups, etc. The error handler function has same signature as Handle's
// error handling function, i.e., err2.Handler. By returning nil resets the
// error, which allows e.g. prevent automatic error logs to happening.
// Otherwise, the output results depends on the current Tracer and assert
// settings. Default setting print call stacks for panics but not for errors:
//
//	defer err2.Catch(func(err error) error { return err} )
//
// or if you you prefer to use dedicated helpers:
//
//	defer err2.Catch(err2.Noop)
//
// The last one calls your error handler, and you have an explicit panic
// handler too, where you can e.g. continue panicking to propagate it for above
// callers or stop it like below:
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
	}, a...)
	doTrace(err)
}

// Throwf builds and throws (panics) an error. For creation it's similar to
// fmt.Errorf. Because panic is used to transport the error instead of error
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

// ToStderr is a built-in helper to use with Handle and Catch. It prints the
// error to stderr and it resets the current error value. It's a handy Catch
// handler in main function.
//
// You can use it like this:
//
//	func main() {
//		defer err2.Catch(err2.ToStderr)
func ToStderr(err error) error {
	fmt.Fprintln(os.Stderr, err.Error())
	return nil
}

// ToStdout is a built-in helper to use with Handle and Catch. It prints the
// error to stdout and it resets the current error value. It's a handy Catch
// handler in main function.
//
// You can use it like this:
//
//	func main() {
//		defer err2.Catch(err2.ToStdout)
func ToStdout(err error) error {
	fmt.Fprintln(os.Stdout, err.Error())
	return nil
}

// Noop is a built-in helper to use with Handle and Catch. It keeps the current
// error value the same. You can use it like this:
//
//	defer err2.Handle(&err, err2.Noop)
func Noop(err error) error { return err }

// Reset is a built-in helper to use with Handle and Catch. It sets the current
// error value to nil. You can use it like this to reset the error:
//
//	defer err2.Handle(&err, err2.Reset)
func Reset(error) error { return nil }

// Err is a built-in helper to use with Handle and Catch. It offers simplifier
// for error handling function for cases where you don't need to change the
// current error value. For instance, if you want to just write error to stdout,
// and don't want to use SetLogTracer and keep it to write to your logs.
//
//	defer err2.Catch(err2.Err(func(err error) {
//		fmt.Println("ERROR:", err)
//	}))
func Err(f func(err error)) func(error) error {
	return func(err error) error {
		f(err)
		return err
	}
}

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
