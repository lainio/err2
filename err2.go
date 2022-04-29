package err2

import (
	"errors"
	"fmt"
	"io"
	"os"

	"runtime"

	"github.com/lainio/err2/internal/debug"
)

// StackStraceWriter allows to set automatic stack tracing.
//  err2.StackStraceWriter = os.Stderr // write stack trace to stderr
//   or
//  err2.StackStraceWriter = log.Writer() // stack trace to std logger
var StackStraceWriter io.Writer

// Try is deprecated. Use try.To functions from try package instead.
// Try is as similar as proposed Go2 Try macro, but it's a function and it
// returns slice of interfaces. It has quite big performance penalty when
// compared to Check function.
func Try(args ...any) []any {
	check(args)
	return args
}

// Check is deprecated. Use try.To function instead.
// Check performs error check for the given argument. If the err is nil, it does
// nothing. According the measurements, it's as fast as
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

// Handle is for adding an error handler to a function by defer. It's for
// functions returning errors them self. For those functions that doesn't return
// errors there is a Catch function. Note! The handler is called only when err
// != nil.
func Handle(err *error, handler func()) {
	// This and Catch are similar but we need to call recover() here because
	// how it works with defer. We cannot refactor these to use same function.

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		// Defers are in the stack and the first from the stack gets the
		// opportunity to get panic object's error (below). We still must
		// call handler functions to the rest of the handlers if there is
		// an error.
		if *err != nil {
			handler()
		}
	case runtime.Error:
		panic(r)
	case error:
		// We or someone did transport this error thru panic.
		*err = r.(error)
		handler()
	default:
		panic(r)
	}
}

// Catch is a convenient helper to those functions that doesn't return errors.
// Go's main function is a good example. Note! There can be only one deferred
// Catch function per non error returning function. See Handle for more
// information.
func Catch(f func(err error)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.
	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		break
	case runtime.Error:
		panic(r)
	case error:
		f(r.(error))
	default:
		panic(r)
	}
}

// CatchAll is a helper function to catch and write handlers for all errors and
// all panics thrown in the current go routine.
func CatchAll(errorHandler func(err error), panicHandler func(v any)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		break
	case runtime.Error:
		panicHandler(r)
	case error:
		errorHandler(r.(error))
	default:
		panicHandler(r)
	}
}

// CatchTrace is a helper function to catch and handle all errors. It recovers a
// panic as well and prints its call stack. This is preferred helper for go
// workers on long running servers.
func CatchTrace(errorHandler func(err error)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	r := recover()
	printStackTrace(os.Stderr, r)

	switch r.(type) {
	case nil:
		break
	case error:
		errorHandler(r.(error))
	}
}

// Return is same as Handle but it's for functions which don't wrap or annotate
// their errors. If you want to annotate errors see Annotate for more
// information.
func Return(err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		break
	case runtime.Error:
		panic(r)
	case error:
		*err = r.(error)
	default:
		panic(r)
	}
}

// Returnw wraps an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. Note! If you don't want to wrap the error use Returnf instead.
func Returnw(err *error, format string, args ...any) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		if *err != nil { // if other handlers call recovery() we still..
			*err = fmt.Errorf(format+": %w", append(args, *err)...)
		}
	case runtime.Error:
		panic(r)
	case error:
		e := r.(error)
		*err = fmt.Errorf(format, e)
	default:
		panic(r)
	}
}

// Annotatew is for annotating an error. It's similar to Returnf but it takes only
// two arguments: a prefix string and a pointer to error. It adds ": " between
// the prefix and the error text automatically.
func Annotatew(prefix string, err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		if *err != nil { // if other handlers call recovery() we still..
			format := prefix + ": %w"
			*err = fmt.Errorf(format, (*err))
		}
	case runtime.Error:
		panic(r)
	case error:
		e := r.(error)
		format := prefix + ": %w"
		*err = fmt.Errorf(format, e)
	default:
		panic(r)
	}
}

// Returnf builds an error. It's similar to fmt.Errorf, but it's called only if
// error != nil. Note! It doesn't use %w to wrap the error. Use Returnw for
// that.
func Returnf(err *error, format string, args ...any) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		if *err != nil { // if other handlers call recovery() we still..
			*err = fmt.Errorf(format+": %v", append(args, *err)...)
		}
	case runtime.Error:
		panic(r)
	case error:
		e := r.(error)
		*err = fmt.Errorf(format+": %v", append(args, e)...)
	default:
		panic(r)
	}
}

// Annotate is for annotating an error. It's similar to Returnf but it takes only
// two arguments: a prefix string and a pointer to error. It adds ": " between
// the prefix and the error text automatically.
func Annotate(prefix string, err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	r := recover()
	checkStackTracePrinting(r)

	switch r.(type) {
	case nil:
		if *err != nil { // if other handlers call recovery() we still..
			format := prefix + ": %v"
			*err = fmt.Errorf(format, (*err))
		}
	case runtime.Error:
		panic(r)
	case error:
		e := r.(error)
		format := prefix + ": %v"
		*err = fmt.Errorf(format, e)
	default:
		panic(r)
	}
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

func printStack(w io.Writer, si debug.StackInfo, msg any) {
	fmt.Fprintf(w, "---\n%v\n---\n", msg)
	debug.FprintStack(w, si)
}

func printStackIf(si debug.StackInfo, msg any) {
	if StackStraceWriter != nil {
		printStack(StackStraceWriter, si, msg)
	}
}

func checkStackTracePrinting(r any) {
	switch r.(type) {
	case nil:
		break
	case runtime.Error:
		si := stackPrologRuntime
		printStackIf(si, r)
	case error:
		si := stackPrologError
		printStackIf(si, r)
	default:
		si := stackPrologPanic
		printStackIf(si, r)
	}
}

func printStackTrace(w io.Writer, r any) {
	switch r.(type) {
	case nil:
		break
	case runtime.Error:
		si := stackPrologRuntime
		printStack(w, si, r)
	case error:
		si := stackPrologError
		printStack(w, si, r)
	default:
		si := stackPrologPanic
		printStack(w, si, r)
	}
}

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

var (
	stackPrologRuntime = newSI("", "panic(", 1)
	stackPrologError   = newErrSI()
	stackPrologPanic   = newSI("", "panic(", 1)
)
