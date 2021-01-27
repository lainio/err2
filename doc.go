/*
Package err2 provides simple helper functions for Go's error handling.

The traditional error handling idiom in Go is roughly akin to

 if err != nil {
 	return err
 }

which applied recursively. That leads to code smells: redundancy, noise&verbose,
or suppressed checks. The err2 package drives programmers more to focus on
error handling rather than checking errors. We think that checks should be so
easy (help of the declarative control structures) that we never forget them.

 err2.Try(io.Copy(w, r))

Error checks

The err2 provides convenient helpers to check the errors. For example, instead
of

 b, err := ioutil.ReadAll(r)
 if err != nil {
 	return err
 }

we can write

 b := err2.Bytes.Try(ioutil.ReadAll(r))

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
