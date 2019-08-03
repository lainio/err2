/*
Package err2 provides simple helper functions for error handling.

The traditional error handling idiom in Go is roughly akin to

 if err != nil {
 	return err
 }

which applied recursively. That leads to problems like code noise, redundancy,
or even non-checks. The err2 package drives programmers more to focus on
error handling rather than checking errors. We think that checks should be as
easy as possible that we never forget them.

 err2.Try(io.Copy(w, r))

Error checks

The err2 provides convenient helpers to check the errors. For example, instead
of

 _, err := ioutil.ReadAll(r)
 if err != nil {
 	return err
 }

we can write

 err2.Try(ioutil.ReadAll(r))

but not without the handler.

Error handling

Package err2 relies on error handlers. In every function which uses err2 for
error-checking has to have at least one error handler. If there are no error
handlers and error occurs it panics. Panicking for the errors during the
development is better than not checking the error at all.

The handler for the previous sample is

 defer err2.Return(&err)

which is the helper handler for cases that don't annotate errors.
err2.Handle is a helper function to add needed error handlers to defer stack.
In most real-world cases, we have multiple error checks and only one or just a
few error handlers per function.
*/
package err2

import "errors"

type transport struct {
	error
}

type _string struct {
}

// String is a helper variable to demonstrate how we could build 'type wrappers'
// to make Try function as fast as Check.
var String _string

// Try is a helper method to call func() (string, error) functions with it and
// be as fast as Check(err).
func (s _string) Try(str string, err error) string {
	Check(err)
	return str
}

// Try is as similar as proposed Go2 Try macro, but it's a function and it
// returns slice of interfaces. It has quite big performance penalty when
// compared to Check function.
func Try(args ...interface{}) []interface{} {
	check(args)
	return args
}

// Check performs the error check for the given argument. If the err is nil,
// it does nothing. According the measurements, it's as fast as if err != nil
// {return err} on happy path.
func Check(err error) {
	if err != nil {
		panic(transport{err})
	}
}

// Checks the error status of the last argument. It panics with "wrong
// signature" if the last calling parameter is not error. In case of error it
// wraps it to transport{} and delivers it by panicking.
func check(args []interface{}) {
	argCount := len(args)
	last := argCount - 1
	if args[last] != nil {
		err, ok := args[last].(error)
		if !ok {
			panic("wrong signature")
		}
		panic(transport{err})
	}
}

// Handle is for adding an error handler to a function by defer. It's for
// functions returning errors them self. For those functions that doesn't
// return errors there is a Catch function. Note! The handler function f is
// called only when err != nil.
func Handle(err *error, f func()) {
	// This and Catch are similar but we need to call recover() here because
	// how it works with defer. We cannot refactor these to use same function.

	// We put real panic objects back and keep only those which are
	// carrying our errors. We must also call all of the handlers in defer
	// stack.
	switch r := recover(); r.(type) {
	case nil:
		// Defers are in the stack and the first from the stack gets the
		// opportunity to get panic object's error (below). We still must
		// call handler functions to the rest of the handlers if there is
		// an error.
		if *err != nil {
			f()
		}
	case transport:
		// We did transport this error thru panic.
		e := r.(transport)
		*err = e.error
		f()
	default:
		panic(r)
	}
}

// Catch is a convenient helper to those functions that doesn't return errors.
// The main function good example of that kind of function. Note! There can be
// only one deferred Catch function per non error returning function. See Handle
// for more information.
func Catch(f func(err error)) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these 2 to use same function.

	if r := recover(); r != nil {
		e, ok := r.(transport)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		f(e)
	}
}

// Return is same as Handle but it's for functions which don't wrap or annotate
// their errors. If you want to annotate errors see Returnf for more information.
func Return(err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(transport)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = e.error
	}
}

// Returnf is for annotating an error. It's similar to Errorf but it takes only
// two arguments: prefix string and a pointer to error.
func Annotate(prefix string, err *error) {
	// This and Handle are similar but we need to call recover here because how
	// it works with defer. We cannot refactor these two to use same function.

	if r := recover(); r != nil {
		e, ok := r.(transport)
		if !ok {
			panic(r) // Not ours, carry on panicking
		}
		*err = e
		format := prefix + e.Error()
		*err = errors.New(format)
	} else if *err != nil { // if other handlers call recovery() we still..
		format := prefix + (*err).Error()
		*err = errors.New(format)
	}
}
