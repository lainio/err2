# err2

The package provides simple helper functions for _automatic_ error propagation.

`go get github.com/lainio/err2`

## Structure

err2 has the following package structure:
- The `err2` (main) package includes declarative error handling functions.
- The `try` package offers error checking functions.
- The `assert` package implements assertion helpers for *design-by-contract*.

## Automatic Error Propagation And Stack Tracing

The current version of Go tends to produce too much error checking and too
little error handling. This package helps us fix that.

1. It helps to declare error handlers with `defer`.
2. It helps to check and transport errors to the nearest (the defer-stack) error
   handler. 
3. It helps us use design-by-contract type preconditions.
4. It offers automatic stack tracing for every error, runtime error, or panic.

You can use all of them or just the other. However, if you use `try` for error
checks you must remember use Go's `recover()` by yourself, or your error isn't
transformed to an `error` return value at any point.

## Error handling

Package `err2` relies on Go's declarative programming structure `defer`. The
`err2` helps to set deferred functions (error handlers) which are only called if
`err != nil`.

In every function which uses err2 for error-checking should have at least one
error handler. If there are no error handlers and error occurs the current
function panics. However, if *any* function above in the call stack has `err2`
error handler it will catch the error.

This is the simplest form of `err2` error handler

```go
defer err2.Return(&err)
```

which is the helper handler for cases that don't need to annotate the error. If
you need to annotate the error you can use either `Annotate` or `Returnf`. These
functions have their error wrapping versions as well: `Annotatew` and `Returnw`.
Our general guideline is:
> Do not wrap an error when doing so would expose implementation details.

#### Automatic Stack Tracing

err2 offers optional stack tracing. It's automatic. Just set the
`StackTraceWriter` to the stream you want traces to be written:

```go
  err2.StackStraceWriter = os.Stderr // write stack trace to stderr
   or
  err2.StackStraceWriter = log.Writer() // stack trace to std logger
```

If `StackTraceWriter` is not set no stack tracing is done. This is the default
because in the most cases proper error messages are enough and panics are
handled immediately anyhow.

#### Manual Stack Tracing

err2 offers two error catchers for manual stack tracing: `CatchTrace` and
`CatchAll`. The first one lets you to handle errors and it will print stack
trace to `stderr` for panic and `runtime.Error`. The second is same but you have
separated handler function for panic and `runtime.Error` so you can decide by
yourself where to print them or what to do with them.

#### Error Handler

The `err2.Handle` is a helper function to add actual error handlers which are
called only if an error has occurred. In most real-world cases, we have multiple
error checks and only one or just a few error handlers. However, you can have as
many error handlers per function as you need.

[Read the package documentation for more
information](https://pkg.go.dev/github.com/lainio/err2).

## Error checks

The `try` package provides convenient helpers to check the errors. Since the Go
1.18 we have been using generics to have fast and convenient error checking.

For example, instead of

```go
b, err := ioutil.ReadAll(r)
if err != nil {
        return err
}
...
```
we can call
```go
b := try.To1(ioutil.ReadAll(r))
...
```

but not without an error handler (`Return`, `Annotate`, `Handle`) or it just
panics your app if you don't have a `recovery` call in the current call stack.
However, you can put your error handlers where ever you want in your call stack.
That can be handy in the internal packages and certain types of algorithms.

We think that panicking for the errors at the start of the development is far
better than not checking errors at all.


#### Filters for non-errors like io.EOF

When error values are used to transport some other information instead of
actual errors we have functions like `try.Is` and even `try.IsEOF` for
convenience.

With these you can write code where error is translated to boolean value:
```go
	notExist := try.Is(r2.err, plugin.ErrNotExist)

	// real errors are cought and the returned boolean tells if value
	// dosen't exist returnend as `plugin.ErrNotExist`
```

For more information see the examples of both functions.

## Assertion (design by contract)

The `assert` package is meant to be used for design-by-contract-type of
development where you set preconditions for your functions. It's not meant to
replace normal error checking but speed up incremental hacking cycle. That's the
reason why default mode (`var D Asserter`) is to panic. By panicking developer
get immediate and proper feedback which allows cleanup the code and APIs before
actual production release.

```go
func marshalAttestedCredentialData(json []byte, data *protocol.AuthenticatorData) []byte {
	assert.SLen(data.AttData.AAGUID, 16, "wrong AAGUID length")
	assert.NotEmpty(data.AttData.CredentialID, "empty credential id")
	assert.SNotEmpty(data.AttData.CredentialPublicKey, "empty credential public key")
	...
```

Previous code block shows the use of the default asserter for developing.

```go
assert.DefaultAsserter = 0
```

If any of the assertion fails, code panics. These type of assertions can be used
without help of the `err2` package.

During the software development lifecycle it isn't all the time crystal clear
what are preconditions for a programmer and what should be translated to
end-user errors as well. That's why `assert` package uses concept called
`Asserter` to have different type of asserter for different phases of a software
project.

The following code block is a sample where production time asserter is used to
generate proper error messages.

```go
func (ac *Cmd) Validate() (err error) {
	defer err2.Return(&err)

	assert.P.NotEmpty(ac.SubCmd, "sub command needed")
	assert.P.Truef(ac.SubCmd == "register" || ac.SubCmd == "login",
		"wrong sub command: %s: want: register|login", ac.SubCmd)
	assert.P.NotEmpty(ac.UserName, "user name needed")
	assert.P.NotEmpty(ac.Url, "connection URL cannot be empty")
	assert.P.NotEmpty(ac.AAGUID, "authenticator ID needed")
	assert.P.NotEmpty(ac.Key, "master key needed")

	return nil
}
```

When asserts are used to generate end-user error messages instead of immediate
panics, `err2` handlers are needed to translate asserts to errors in convenient
way. That's the reason we decided to build `assert` as a sub package of `err2`
even there are no real dependencies between them. See the `assert` packages own
documentation and examples for more information.


## Background

`err2` implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).
The package does it by using internally `panic/recovery`, which some might think
isn't perfect. We have run many benchmarks to try to minimise the performance
penalty this kind of mechanism might bring. We have focused on the _happy path_
analyses. If the performance of the *error path* is essential, don't use this
mechanism presented here. But be aware that if your code uses the error path as 
a part of algorithm itself something is wrong.

**For happy paths** by using `try.ToX` error check functions **there are no
performance penalty at all**. However, the mandatory use of the `defer` might
prevent some code optimisations like function inlining. If you have a
performance-critical use case, we recommend you to write performance tests to
measure the effect. As a general guideline for maximum performance we recommend
to put error handlers as high in the call stack as possible.

The original goal was to make it possible to write similar code than proposed
Go2 error handling would allow and do it right now (summer 2019). The goal was
well aligned with the latest Go2 proposal where it would bring a `try` macro and
let the error handling be implemented in defer blocks. The try-proposal was
cancelled at its latest form. Nevertheless, we have learned that using panics
for early-stage error transport isn't a bad thing but opposite. It seems to help
to draft algorithms much faster, and still maintains the readability.

## Learnings by so far

We have used the `err2` and `assert` packages in several internal projects. The
results have been so far very encouraging:

- If you forget to use handler, but you use checks from the package, you will
get panics if an error occurs. That is much better than getting unrelated panic
somewhere else in the code later. There have also been cases when code reports
error correctly because the 'upper' handler catches it.

- Because the use of `err2.Annotate` is so relatively easy, error messages much
better and informative.

- **When error handling is based on the actual error handlers, code changes have
been much easier.**

- You don't seem to need '%w' wrapping. See the Go's official blog post what are
[cons](https://go.dev/blog/go1.13-errors) for that.
  > Do not wrap an error when doing so would expose implementation details.

## Support

The package has been in experimental mode quite long time. Since the Go generics
we are transiting towards more official mode. Currently we offer support by
author's email. Before sending questions about the package we suggest you will
read all the documentation and examples thru. They are pretty comprehensive.

- harlain at gmail.com

## Roadmap

Version history:
- 0.1, first draft (Summer 2019)
- 0.2, code generation for type helpers
- 0.3, `Returnf` added, not use own transport type anymore but just `error`
- 0.4, Documentation update
- 0.5, Go modules are in use now
- 0.6.1, `assert` package added, and new type helpers (current)
- 0.7.0 filter functions for the cases where errors aren't real errors like
  io.EOF
- 0.8.0 `try.To()` & `assert.That()`, etc. functions with the help of the generics
- 0.8.1 **bug-fix**: `runtime.Error` types are treated as `panics` now (Issue #1)
- 0.8.3 `try.IsXX()` bug fix, lots of new docs, and **automatic stack tracing!**

