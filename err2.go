package err2

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/lainio/err2/internal/handler"
)

// StackTraceWriter allows to set automatic stack tracing.
//  err2.StackTraceWriter = os.Stderr // write stack trace to stderr
//   or
//  err2.StackTraceWriter = log.Writer() // stack trace to std logger
var StackTraceWriter io.Writer

// Try is deprecated. Use try.To functions from try package instead.
// Try is as similar as proposed Go2 Try macro, but it's a function and it
// returns slice of interfaces. It has quite big performance penalty when
// compared to Check function.
func Try(args ...any) []any {
	check(args)
	return args
}

// Check is deprecated. Use try.To function instead. Check performs error check
// for the given argument. If the err is nil, it does nothing. According the
// measurements, it's as fast as
//  if err != nil {
//      return err
//  }
// on happy path.
func Check(err error) {
	if err != nil {
		panic(err)
	}
}

// FilterTry is deprecated. Use try.Is function instead.
// FilterTry performs filtered error check for the given argument. It's same
// as Check but before throwing an error it checks if error matches the filter.
// The return value false tells that there are no errors and true that filter is
// matched.
func FilterTry(filter, err error) bool {
	if err != nil {
		if errors.Is(err, filter) {
			return true
		}
		panic(err)
	}
	return false
}

// TryEOF is deprecated. Use try.IsEOF function instead.
// TryEOF checks errors but filters io.EOF from the exception handling and
// returns boolean which tells if io.EOF is present. See more info from
// FilterCheck.
func TryEOF(err error) bool {
	return FilterTry(io.EOF, err)
}

// check is deprecated. Used by Try. Checks the error status of the last
// argument. It panics with "wrong signature" if the last calling parameter is
// not error. In case of error it delivers it by panicking.
func check(args []any) {
	argCount := len(args)
	last := argCount - 1
	if args[last] != nil {
		err, ok := args[last].(error)
		if !ok {
			panic("wrong signature")
		}
		panic(err)
	}
}

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
	handler.Process(handler.Info{
		Trace: StackTraceWriter,
		Any:   r,
		NilHandler: func() {
			// Defers are in the stack and the first from the stack gets the
			// opportunity to get panic object's error (below). We still must
			// call handler functions to the rest of the handlers if there is
			// an error.
			if *err != nil {
				handlerFn()
			}
		},
		ErrorHandler: func(e error) {
			// We or someone did transport this error thru panic.
			*err = e
			handlerFn()
		},
	})
}

// Catch is a convenient helper to those functions that doesn't return errors.
// There can be only one deferred Catch function per non error returning
// function like main(). It doesn't stop panics and runtime errors. If that's
// important use CatchAll or CatchTrace instead. See Handle for more
// information.
func Catch(f func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        StackTraceWriter,
		Any:          r,
		ErrorHandler: f,
	})
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine. It and CatchTrace are preferred
// helperr for go workers on long running servers, because they stop panics as
// well.
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        StackTraceWriter,
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: panicHandler,
	})
}

// CatchTrace is a helper function to catch and handle all errors. It also
// recovers a panic and prints its call stack. It and CatchAll are preferred
// helpers for go-workers on long-running servers because they stop panics as
// well.
func CatchTrace(errorHandler func(err error)) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        os.Stderr,
		Any:          r,
		ErrorHandler: errorHandler,
		PanicHandler: func(v any) {}, // suppress panicking
	})
}

// Return is the same as Handle but it's for functions that don't wrap or
// annotate their errors. It's still needed to break panicking which is used for
// error transport in err2. If you want to annotate errors see other Annotate
// and Return functions for more information.
func Return(err *error) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace:        StackTraceWriter,
		Any:          r,
		ErrorHandler: func(e error) { *err = e },
	})
}

// Returnw wraps an error with '%w'. It's similar to fmt.Errorf, but it's called
// only if error != nil. If you don't want to wrap the error use Returnf
// instead.
func Returnw(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace: StackTraceWriter,
		Any:   r,
		NilHandler: func() {
			if *err != nil { // if other handlers call recovery() we still..
				*err = fmt.Errorf(format+": %w", append(args, *err)...)
			}
		},
		ErrorHandler: func(e error) {
			*err = fmt.Errorf(format+": %w", append(args, e)...)
		},
	})
}

// Annotatew is for annotating an error. It's similar to Returnf but it takes only
// two arguments: a prefix string and a pointer to error. It adds ": " between
// the prefix and the error text automatically.
func Annotatew(prefix string, err *error) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace: StackTraceWriter,
		Any:   r,
		NilHandler: func() {
			if *err != nil { // if other handlers call recovery() we still..
				format := prefix + ": %w"
				*err = fmt.Errorf(format, (*err))
			}
		},
		ErrorHandler: func(e error) {
			format := prefix + ": %w"
			*err = fmt.Errorf(format, e)
		},
	})
}

// Returnf builds an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. It uses '%v' to wrap the error not '%w'. Use Returnw for that.
func Returnf(err *error, format string, args ...any) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace: StackTraceWriter,
		Any:   r,
		NilHandler: func() {
			if *err != nil { // if other handlers call recovery() we still..
				*err = fmt.Errorf(format+": %v", append(args, *err)...)
			}
		},
		ErrorHandler: func(e error) {
			*err = fmt.Errorf(format+": %v", append(args, e)...)
		},
	})
}

// Annotate is for annotating an error. It's similar to Returnf but it takes
// only two arguments: a prefix string and a pointer to error. It adds ": "
// between the prefix and the error text automatically.
func Annotate(prefix string, err *error) {
	// This and others are similar but we need to call `recover` here because
	// how it works with defer.
	r := recover()

	handler.Process(handler.Info{
		Trace: StackTraceWriter,
		Any:   r,
		NilHandler: func() {
			if *err != nil { // if other handlers call recovery() we still..
				format := prefix + ": %v"
				*err = fmt.Errorf(format, (*err))
			}
		},
		ErrorHandler: func(e error) {
			format := prefix + ": %v"
			*err = fmt.Errorf(format, e)
		},
	})
}

type _empty struct{}

// Empty is deprecated. Use try.To functions instead.
// Empty is a helper variable to demonstrate how we could build 'type wrappers'
// to make Try function as fast as Check.
// Note! Deprecated, use try package.
var Empty _empty

// Try is deprecated. Use try.To functions instead.
// Try is a helper method to call func() (string, error) functions with it and
// be as fast as Check(err).
func (s _empty) Try(_ any, err error) {
	Check(err)
}
