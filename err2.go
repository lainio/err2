package err2

import (
	"errors"
	"fmt"

	"github.com/lainio/err2/internal/handler"
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
//	defer err2.Handle(&err, func() {
//		os.Remove(dst)
//	})
//
// If you need to stop general panics in handler, you can do that by giving a
// panic handler function:
//
//	defer err2.Handle(&err,
//	   func() {
//	      os.Remove(dst)
//	   },
//	   func(p any) {} // panic handler, it's stops panics, you can re-throw
//	)
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
// Note, that Catch always catch the panics. If you don't want to stop the (aka
// recover) you should add panic handler and countinue panicing there. There can
// be only one deferred Catch function per non error returning function like
// main(). There is several ways to make deferred calls to Catch.
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
