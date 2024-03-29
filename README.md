[![test](https://github.com/lainio/err2/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/lainio/err2/actions/workflows/test.yml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/lainio/err2)](https://pkg.go.dev/mod/github.com/lainio/err2)
[![Go Report Card](https://goreportcard.com/badge/github.com/lainio/err2?style=flat-square)](https://goreportcard.com/report/github.com/lainio/err2)

# err2

The package extends Go's error handling with **fully automatic error checking
and propagation** like other modern programming languages: **Zig**, Rust, Swift,
etc. `err2` isn't an exception handling library, but an entirely orthogonal
package with Go's existing error handling mechanism.

```go
func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err)

	r := try.To1(os.Open(src))
	defer r.Close()

	w := try.To1(os.Create(dst))
	defer err2.Handle(&err, err2.Err(func(error) {
		try.Out(os.Remove(dst)).Logf("cleaning error")
	}))
	defer w.Close()

	try.To1(io.Copy(w, r))
	return nil
}
```

`go get github.com/lainio/err2`

- [Structure](#structure)
- [Performance](#performance)
- [Automatic Error Propagation](#automatic-error-propagation)
- [Error handling](#error-handling)
  - [Error Stack Tracing](#error-stack-tracing)
- [Error Checks](#error-checks)
  - [Filters for non-errors like io.EOF](#filters-for-non-errors-like-ioeof)
- [Backwards Compatibility Promise for the API](#backwards-compatibility-promise-for-the-api)
- [Assertion](#assertion)
  - [Asserters](#asserters)
  - [Assertion Package for Runtime Use](#assertion-package-for-runtime-use)
  - [Assertion Package for Unit Testing](#assertion-package-for-unit-testing)
- [Automatic Flags](#automatic-flags)
  - [Support for Cobra Flags](#support-for-cobra-flags)
- [Code Snippets](#code-snippets)
- [Background](#background)
- [Learnings by so far](#learnings-by-so-far)
- [Support And Contributions](#support-and-contributions)
- [Roadmap](#roadmap)


## Structure

`err2` has the following package structure:
- The `err2` (main) package includes declarative error handling functions.
- The `try` package offers error checking functions.
- The `assert` package implements assertion helpers for **both** unit-testing
  and *design-by-contract*.

## Performance

All of the listed above **without any performance penalty**! You are welcome to
run `benchmarks` in the project repo and see yourself.

Please note that many benchmarks run 'too fast' according to the common Go
benchmarking rules, i.e., compiler optimizations
([inlining](https://en.wikipedia.org/wiki/Inline_expansion)) are working so well
that there are no meaningful results. But for this type of package, where **we
compete with if-statements, that's precisely what we hope to achieve.** The
whole package is written toward that goal. Especially with parametric
polymorphism, it's been quite the effort.

## Automatic Error Propagation

The current version of Go tends to produce too much error checking and too
little error handling. But most importantly, it doesn't help developers with
**automatic** error propagation, which would have the same benefits as, e.g.,
**automated** garbage collection or automatic testing:

> Automation is not just about efficiency but primarily about repeatability and
> resilience. -- Gregor Hohpe

Automatic error propagation is crucial because it makes your code change
tolerant. And, of course, it helps to make your code error-safe:

![Never send a human to do a machine's job](https://www.magicalquote.com/wp-content/uploads/2013/10/Never-send-a-human-to-do-a-machines-job.jpg)

The err2 package is your automation buddy:

1. It helps to declare error handlers with `defer`. If you're familiar with [Zig
   language](https://ziglang.org/), you can think `defer err2.Handle(&err,...)`
   line exactly similar as
   [Zig's `errdefer`](https://ziglang.org/documentation/master/#errdefer).
2. It helps to check and transport errors to the nearest (the defer-stack) error
   handler.
3. It helps us use design-by-contract type preconditions.
4. It offers automatic stack tracing for every error, runtime error, or panic.
   If you are familiar with Zig, the `err2` error traces are same as Zig's.

You can use all of them or just the other. However, if you use `try` for error
checks, you must remember to use Go's `recover()` by yourself, or your error
isn't transformed to an `error` return value at any point.

## Error handling

The `err2` relies on Go's declarative programming structure `defer`. The
`err2` helps to set deferred functions (error handlers) which are only called if
`err != nil`.

Every function which uses err2 for error-checking should have at least one error
handler. The current function panics if there are no error handlers and an error
occurs. However, if *any* function above in the call stack has an err2 error
handler, it will catch the error.

This is the simplest form of `err2` automatic error handler:

```go
func doSomething() (err error) {
    // below: if err != nil { return ftm.Errorf("%s: %w", CUR_FUNC_NAME, err) }
    defer err2.Handle(&err)
```

See more information from `err2.Handle`'s documentation. It supports several
error-handling scenarios. And remember that you can have as many error handlers
per function as you need, as well as you can chain error handling functions per
`err2.Handle` that allows you to build new error handling middleware for your
own purposes.

#### Error Stack Tracing

The err2 offers optional stack tracing. It's automatic and optimized. Optimized
means that the call stack is processed before output. That means that stack
trace starts from where the actual error/panic is occurred, not where the error
is caught. You don't need to search for the line where the pointer was nil or
received an error. That line is in the first one you are seeing:

```console
---
runtime error: index out of range [0] with length 0
---
goroutine 1 [running]:
main.test2({0x0, 0x0, 0x40XXXXXf00?}, 0x2?)
	/home/.../go/src/github.com/lainio/ic/main.go:43 +0x14c
main.main()
	/home/.../go/src/github.com/lainio/ic/main.go:77 +0x248
```

Just set the `err2.SetErrorTracer` or `err2.SetPanicTracer` to the stream you
want traces to be written:

```go
err2.SetErrorTracer(os.Stderr) // write error stack trace to stderr
// or, for example:
err2.SetPanicTracer(log.Writer()) // stack panic trace to std logger
```

If no `Tracer` is set no stack tracing is done. This is the default because in
the most cases proper error messages are enough and panics are handled
immediately by a programmer.

> Note. Since v0.9.5 you can set these tracers through Go's standard flag
> package just by adding `flag.Parse()` to your program. See more information
> from [Automatic Flags](#automatic-flags).

[Read the package documentation for more
information](https://pkg.go.dev/github.com/lainio/err2).

## Error Checks

The `try` package provides convenient helpers to check the errors. Since the Go
1.18 we have been using generics to have fast and convenient error checking.

For example, instead of

```go
b, err := io.ReadAll(r)
if err != nil {
        return err
}
...
```
we can call
```go
b := try.To1(io.ReadAll(r))
...
```

but not without an error handler (`err2.Handle`). However, you can put your
error handlers where ever you want in your call stack. That can be handy in the
internal packages and certain types of algorithms.

In cases where you want to handle the error immediately after the function call
return you can use Go's default `if` statement. However, we encourage you to use
the `errdefer` concept, `defer err2.Handle(&err)` for all of your error
handling.

Nevertheless, there might be cases where you might want to:
1. Suppress the error and use some default value.
1. Just write a logline and continue without a break.
1. Annotate the specific error value even when you have a general error handler.
1. You want to handle the specific error value, let's say, at the same line
   or statement.

The `err2/try` package offers other helpers based on the DSL concept
where the DSL's domain is error-handling. It's based on functions `try.Out`,
`try.Out1`, and `try.Out2`, which return instances of types `Result`, `Result1`,
and `Result2`. The `try.Result` is similar to other programming languages,
i.e., discriminated union. Please see more from its documentation.

Now we could have the following:

```go
b := try.Out1(strconv.Atoi(s)).Logf("%s => 100", s).Catch(100)
```

The previous statement tries to convert incoming string value `s`, but if it
doesn't succeed, it writes a warning to logs and uses the default value (100).
The logging result includes the original error message as well.

It's easy to see that panicking about the errors at the start of the development
is far better than not checking errors at all. But most importantly, `err2/try`
**keeps the code readable.**

#### Filters for non-errors like io.EOF

When error values are used to transport some other information instead of
actual errors we have functions like `try.Is` and even `try.IsEOF` for
convenience.

With these you can write code where error is translated to boolean value:
```go
notExist := try.Is(r2.err, plugin.ErrNotExist)

// real errors are cought and the returned boolean tells if value
// dosen't exist returned as `plugin.ErrNotExist`
```
**Note.** Any other error than `plugin.ErrNotExist` is treated as an real error:
1. `try.Is` function first checks `if err == nil`, and if yes, it returns
   `false`.
2. Then it checks if `errors.Is(err, plugin.ErrNotExist)` and if yes, it returns
   `true`.
3. Finally, it calls `try.To` for the non nil error, and we already know what then
   happens: nearest `err2.Handle` gets it first.

These `try.Is` functions help cleanup mesh idiomatic Go, i.e. mixing happy and
error path, leads to.

For more information see the examples in the documentation of both functions.

## Backwards Compatibility Promise for the API

The `err2` package's API will be **backward compatible**. Before version
1.0.0 is released, the API changes occasionally, but **we promise to offer
automatic conversion scripts for your repos to update them for the latest API.**
We also mark functions deprecated before they become obsolete. Usually, one
released version before. We have tested this with a large code base in our
systems, and it works wonderfully.

More information can be found in the `scripts/` directory [readme
file](./scripts/README.md).

## Assertion

The `assert` package is meant to be used for *design-by-contract-* type of
development where you set pre- and post-conditions for your functions. It's not
meant to replace the normal error checking but speed up the incremental hacking
cycle. The default mode is to return an `error` value that includes a formatted
and detailed assertion violation message. A developer gets immediate and proper
feedback, allowing cleanup of the code and APIs before the release.

#### Asserters

The assert package offers a few pre-build *asserters*, which are used to
configure how the assert package deals with assert violations. The line below
exemplifies how the default asserter is set in the package.

```go
assert.SetDefault(assert.Production)
```

If you want to suppress the caller info (source file name, line number, etc.)
and get just the plain panics from the asserts, you should set the
default asserter with the following line:

```go
assert.SetDefault(assert.Debug)
```

For certain type of programs this is the best way. It allows us to keep all the
error messages as simple as possible. And by offering option to turn additional
information on, which allows super users and developers get more technical
information when needed.

> Note. Since v0.9.5 you can set these asserters through Go's standard flag
> package just by adding `flag.Parse()` to your program. See more information
> from [Automatic Flags](#automatic-flags).

#### Assertion Package for Runtime Use

Following is example of use of the assert package:

```go
func marshalAttestedCredentialData(json []byte, data *protocol.AuthenticatorData) []byte {
	assert.SLen(data.AttData.AAGUID, 16, "wrong AAGUID length")
	assert.NotEmpty(data.AttData.CredentialID, "empty credential id")
	assert.SNotEmpty(data.AttData.CredentialPublicKey, "empty credential public key")
	...
```

We have now described design-by-contract for development and runtime use. What
makes err2's assertion packages unique, and extremely powerful, is its use for
automatic testing as well.

#### Assertion Package for Unit Testing

The same asserts can be used **and shared** during the unit tests:

```go
func TestWebOfTrustInfo(t *testing.T) {
	defer assert.PushTester(t)()

	common := dave.CommonChains(eve.Node)
	assert.SLen(common, 2)

	wot := dave.WebOfTrustInfo(eve.Node) //<- this includes asserts as well!!
	// And if there's violations during the test run they are reported as
	// test failures for this TestWebOfTrustInfo -test.

	assert.Equal(0, wot.CommonInvider)
	assert.Equal(1, wot.Hops)

	wot = NewWebOfTrust(bob.Node, carol.Node)
	assert.Equal(-1, wot.CommonInvider)
	assert.Equal(-1, wot.Hops)
	...
```

A compelling feature is that even if some assertion violation happens during the
execution of called functions like the above `NewWebOfTrust()` function instead
of the actual Test function, **it's reported as a standard test failure.** That
means we don't need to open our internal pre- and post-conditions just for
testing.

**We can share the same assertions between runtime and test execution.**

The err2 `assert` package integration to the Go `testing` package is completed at
the cross-module level. Suppose package A uses package B. If package B includes
runtime asserts in any function that A calls during testing and some of B's
asserts fail, A's current test also fails. There is no loss of information, and
even the stack trace is parsed to test logs for easy traversal. Packages A and B
can be the same or different modules.

**This means that where ever assertion violation happens during the test
execution, we will find it and can even move thru every step in the call
stack.**

## Automatic Flags

When you are using `err2` or `assert` packages, i.e., just importing them, you
have an option to automatically support for err2 configuration flags through
Go's standard `flag` package. See more information about err2 settings from
[Error Stack Tracing](#error-stack-tracing) and [Asserters](#asserters).

Now you can always deploy your applications and services with the simple
end-user friendly error messages and no stack traces, **but you can switch them
on when ever you need**.

Let's say you have build CLI (`your-app`) tool with the support for Go's flag
package, and the app returns an error. Let's assume you're a developer. You can
run it again with:

```
your-app -err2-trace stderr
```

Now you get full error trace addition to the error message. Naturally, this
also works with assertions. You can configure their output with the flag
`asserter`:

```
your-app -asserter Debug
```

That adds more information to the assertion statement, which in default is in
production (`Prod`) mode, i.e., outputs a single-line assertion message.

All you need to do is to add `flag.Parse` to your `main` function.

#### Support for Cobra Flags

If you are using [cobra](https://github.com/spf13/cobra) you can still easily
support packages like `err2` and `glog` and their flags.

1. Add std flag package to imports in `cmd/root.go`:

   ```go
   import (
       goflag "flag"
       ...
   )
   ```

1. Add the following to (usually) `cmd/root.go`'s `init` function's end:

   ```go
   func init() {
       ...
       // NOTE! Very important. Adds support for std flag pkg users: glog, err2
       pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
   }
   ```

1. And finally modify your `PersistentPreRunE` in `cmd/root.go` to something
   like:

   ```go
   PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
       defer err2.Handle(&err)

       // NOTE! Very important. Adds support for std flag pkg users: glog, err2
       goflag.Parse()

       try.To(goflag.Set("logtostderr", "true"))
       handleViperFlags(cmd) // local helper with envs
       glog.CopyStandardLogTo("ERROR") // for err2
       return nil
   },
   ```

As a result you can have bunch of usable flags added to your CLI:

```
Flags:
      --asserter asserter                 asserter: Plain, Prod, Dev, Debug (default Prod)
      --err2-log stream                   stream for logging: nil -> log pkg (default nil)
      --err2-panic-trace stream           stream for panic tracing (default stderr)
      --err2-trace stream                 stream for error tracing: stderr, stdout (default nil)
      ...
```

## Code Snippets

Most of the repetitive code blocks are offered as code snippets. They are in
`./snippets` in VC code format, which is well supported e.g. neovim, etc.

The snippets must be installed manually to your preferred IDE/editor. During the
installation you can modify the according your style or add new ones. We would
prefer if you could contribute some of the back to the err2 package.

## Background

`err2` implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).
The package does it by using internally `panic/recovery`, which some might think
isn't perfect.

We have run many benchmarks try to minimise the performance penalty this kind of
mechanism might bring. We have focused on the _happy path_ analyses. If the
performance of the *error path* is essential, don't use this mechanism presented
here. **But be aware that something is wrong if your code uses the error path as
part of the algorithm itself.**

**For happy paths** by using `try.To*` or `assert.That` error check functions
**there are no performance penalty at all**. However, the mandatory use of the
`defer` might prevent some code optimisations like function inlining. And still,
we have cases where using the `err2` and `try` package simplify the algorithm so
that it's faster than the return value if err != nil version. (**See the
benchmarks for `io.Copy` in the repo.**)

If you have a performance-critical use case, we always recommend you to write
performance tests to measure the effect. As a general guideline for maximum
performance we recommend to put error handlers as high in the call stack as
possible, and use only error checking (`try.To()` calls) in the inner loops. And
yes, that leads to non-local control structures, but it's the most performant
solution of all. (The repo has benchmarks for that as well.)

The original goal was to make it possible to write similar code that the
proposed Go2 error handling would allow and do it right now (summer 2019). The
goal was well aligned with the Go2 proposal, where it would bring a `try` macro
and let the error handling be implemented in defer blocks. The try-proposal was
canceled at its latest form. Nevertheless, we have learned that **using panics**
for early-stage **error transport isn't bad but the opposite**. It seems to
help:
- to draft algorithms much faster,
- huge improvements for the readability,
- helps to bring a new blood (developers with different programming language
  background) to projects,
- and most importantly, **it keeps your code more refactorable** because you
  don't have to repeat yourself.

## Learnings by so far

We have used the `err2` and `assert` packages in several projects. The results
have been so far very encouraging:

- If you forget to use handler, but you use checks from the package, you will
get panics (and optionally stack traces) if an error occurs. That is much better
than getting unrelated panic somewhere else in the code later. There have also
been cases when code reports error correctly because the 'upper' handler catches
it.

- Because the use of `err2.Handle` is so easy, error messages are much better
and informative. When using `err2.Handle`'s automatic annotation your error
messages are always up-to-date. Even when you refactor your function name error
message is also updated.

- **When error handling is based on the actual error handlers, code changes have
been much easier.** There is an excellent [blog post](https://jesseduffield.com/Gos-Shortcomings-1/)
about the issues you are facing with Go's error handling without the help of
the err2 package.

## Support And Contributions

The package has been in experimental mode quite long time. Since the Go generics
we are transiting towards more official mode. Currently we offer support by
GitHub Discussions. Naturally, any issues and contributions are welcome as well!

## Roadmap

Please see the full version history from [CHANGELOG](./CHANGELOG.md).

### Latest Release

##### 1.0.0
- **Finally! We are very happy, and thanks to all who have helped!**
- Lots of documentation updates and cleanups for version 1.0.0
- `Catch/Handle` take unlimited amount error handler functions
  - allows building e.g. error handling middlewares
  - this is major feature because it allows building helpers/add-ons
- automatic outputs aren't overwritten by given args, only with `assert.Plain`
- Minor API fixes to still simplify it:
  - remove exported vars, obsolete types and funcs from `assert` pkg
  - `Result2.Def2()` sets only `Val2`
- technical refactorings: variadic function calls only in API level
