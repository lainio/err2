/*
Package err2 provides three main functionality:
 1. err2 package includes helper functions for error handling & automatic error
    stack tracing
 2. try package is for error checking
 3. assert package is for design-by-contract and preconditions both for normal
    runtime and for testing

The traditional error handling idiom in Go is roughly akin to

	if err != nil { return err }

which applied recursively.

The err2 package drives programmers to focus on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them. The CopyFile example shows how it works:

	// CopyFile copies source file to the given destination. If any error occurs it
	// returns error value describing the reason.
	func CopyFile(src, dst string) (err error) {
	     // Add first error handler just to annotate the error properly.
	     defer err2.Handle(&err)

	     // Try to open the file. If error occurs now, err will be
	     // automatically annotated ('copy file:' prefix calculated from the
	     // function name, no performance penalty) and returned properly thanks
	     // to above err2.Handle.
	     r := try.To1(os.Open(src))
	     defer r.Close()

	     // Try to create a file. If error occurs now, err will be annotated and
	     // returned properly.
	     w := try.To1(os.Create(dst))
	     // Add error handler to clean up the destination file. Place it here that
	     // the next deferred close is called before our Remove call.
	     defer err2.Handle(&err, err2.Err(func(error) {
	     	os.Remove(dst)
	     }))
	     defer w.Close()

	     // Try to copy the file. If error occurs now, all previous error handlers
	     // will be called in the reversed order. And final return error is
	     // properly annotated in all the cases.
	     try.To1(io.Copy(w, r))

	     // All OK, just return nil.
	     return nil
	}

# Error checks and Automatic Error Propagation

The try package provides convenient helpers to check the errors. For example,
instead of

	b, err := io.ReadAll(r)
	if err != nil {
	   return err
	}

we can write

	b := try.To1(io.ReadAll(r))

Note that [try.To] functions are as fast as if err != nil statements. Please see
the try package documentation for more information about the error checks.

# Automatic Stack Tracing

err2 offers optional stack tracing. And yes, it's fully automatic. Just call

	flag.Parse() # this is enough for err2 pkg to add its flags

at the beginning your app, e.g. main function, or set the tracers
programmatically (before [flag.Parse] if you are using that):

	err2.SetErrorTracer(os.Stderr) // write error stack trace to stderr
	 or
	err2.SetPanicTracer(log.Writer()) // panic stack trace to std logger

Note. Since [Catch]'s default mode is to catch panics, the panic tracer's
default values is os.Stderr. The default error tracer is nil.

	err2.SetPanicTracer(os.Stderr) // panic stack tracer's default is stderr
	err2.SetErrorTracer(nil) // error stack tracer's default is nil

# Automatic Logging

Same err2 capablities support automatic logging like the [Catch] and
[try.Result.Logf] functions. To be able to tune up how logging behaves we offer a
tracer API:

	err2.SetLogTracer(nil) // the default is nil where std log pkg is used.

# Flag Package Support

The err2 package supports Go's flags. All you need to do is to call flag.Parse.
And the following flags are supported (="default-value"):

	-err2-log="nil"
	    A name of the stream currently supported stderr, stdout or nil
	-err2-panic-trace="stderr"
	    A name of the stream currently supported stderr, stdout or nil
	-err2-trace="nil"
	    A name of the stream currently supported stderr, stdout or nil

Note, that you have called [SetErrorTracer] and others, before you call
flag.Parse. This allows you set the defaults according your app's need and allow
end-user change them during the runtime.

# Error handling

Package err2 relies on declarative control structures to achieve error and panic
safety. In every function which uses err2 or try package for error-checking has
to have at least one declarative error handler if it returns error value. If
there are no error handlers and error occurs it panics. We think that panicking
for the errors is much better than not checking errors at all. Nevertheless, if
the call stack includes any err2 error handlers like [Handle] the error is
handled where the handler is saved to defer-stack. (defer is not lexically
scoped)

err2 includes many examples to play with like previous CopyFile. Please see them
for more information.
*/
package err2
