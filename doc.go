/*
Package err2 is error handling solution including three main functionality:
 1. err2 package offers helper functions for error handling & automatic error
    stack tracing
 2. [github.com/lainio/err2/try] sub-package is for error checking
 3. [github.com/lainio/err2/assert] sub-package is for design-by-contract and
    preconditions both for normal runtime and for unit testing

The err2 package drives programmers to focus on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them. The CopyFile example shows how it works:

	// CopyFile copies source file to the given destination. If any error occurs it
	// returns error value describing the reason.
	func CopyFile(src, dst string) (err error) {
	     // Add first error handler is to catch and annotate the error properly.
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
	     // Add error handler to clean up the destination file in case of
	     // error. Handler fn is called only if there has been an error at the
	     // following try.To check. We place it here that the next deferred
	     // close is called before our Remove a file call.
	     defer err2.Handle(&err, err2.Err(func(error) {
	     	try.Out(os.Remove(dst)).Logf("cleanup failed")
	     }))
	     defer w.Close()

	     // Try to copy the file. If error occurs now, all previous error handlers
	     // will be called in the reversed order. And a final error value is
	     // properly annotated and returned in all the cases.
	     try.To1(io.Copy(w, r))

	     // All OK, just return nil.
	     return nil
	}

# Error checks and Automatic Error Propagation

The [github.com/lainio/err2/try] package provides convenient helpers to check the errors. For example,
instead of

	b, err := io.ReadAll(r)
	if err != nil {
	   return err
	}

we can write

	b := try.To1(io.ReadAll(r))

Note that try.To functions are as fast as if err != nil statements. Please see
the [github.com/lainio/err2/try] package documentation for more information
about the error checks.

# Automatic Stack Tracing

err2 offers optional stack tracing. And yes, it's fully automatic. Just call

	flag.Parse() # this is enough for err2 pkg to add its flags

at the beginning your app, e.g. main function, or set the tracers
programmatically (before [flag.Parse] if you are using that):

	err2.SetErrRetTracer(os.Stderr)   // write error return trace to stderr
	 or
	err2.SetErrorTracer(os.Stderr)    // write error stack trace to stderr
	 or
	err2.SetPanicTracer(log.Writer()) // panic stack trace to std logger

Note that since [Catch]'s default mode is to recover from panics, it's a good
practice still print their stack trace. The panic tracer's default values is
[os.Stderr]. The default error tracer is nil.

	err2.SetPanicTracer(os.Stderr) // panic stack tracer's default is stderr
	err2.SetErrRetTracer(nil)      // error return tracer's default is nil
	err2.SetErrorTracer(nil)       // error stack tracer's default is nil

Note that both panic and error traces are optimized by err2 package. That means
that the head of the stack trace isn't the panic function, but an actual line
that caused it. It works for all three categories:
  - normal error values
  - [runtime.Error] values
  - any types of the panics

The last two types are handled as panics in the error handling functions given
to [Handle] and [Catch].

# Automatic Logging

Same err2 capablities support automatic logging like the [Catch] and
[try.Result.Logf] functions. To be able to tune up how logging behaves we offer a
tracer API:

	err2.SetLogTracer(nil) // the default is nil where std log pkg is used.

# Flag Package Support

The err2 package supports Go's flags. All you need to do is to call [flag.Parse].
And the following flags are supported (="default-value"):

	-err2-log stream
	      stream for logging: nil -> log pkg
	-err2-panic-trace stream
	      stream for panic tracing (default stderr)
	-err2-ret-trace stream
	      stream for error return tracing: stderr, stdout
	-err2-trace stream
	      stream for error tracing: stderr, stdout

Note that you have called [SetErrorTracer] and others, before you call
[flag.Parse]. This allows you set the defaults according your app's need and allow
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
