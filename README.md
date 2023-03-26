[![test](https://github.com/lainio/err2/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/lainio/err2/actions/workflows/test.yml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.19-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/lainio/err2)](https://pkg.go.dev/mod/github.com/lainio/err2)
[![Go Report Card](https://goreportcard.com/badge/github.com/lainio/err2?style=flat-square)](https://goreportcard.com/report/github.com/lainio/err2)

# err2

The package extends Go's error handling with **fully automatic error
propagation** similar to other modern programming languages: Zig, Rust, Swift,
etc.

```go 
func CopyFile(src, dst string) (err error) {
	defer err2.Handle(&err)

	assert.NotEmpty(src)
	assert.NotEmpty(dst)

	r := try.To1(os.Open(src))
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("mixing traditional error checking: %w", err)
	}
	defer err2.Handle(&err, func() {
		os.Remove(dst)
	})
	defer w.Close()
	try.To1(io.Copy(w, r))
	return nil
}
```

`go get github.com/lainio/err2`

- [Structure](#structure)
- [Automatic Error Propagation](#automatic-error-propagation)
- [Error handling](#error-handling)
  - [Error Stack Tracing](#error-stack-tracing)
- [Error Checks](#error-checks)
  - [Filters for non-errors like io.EOF](#filters-for-non-errors-like-ioeof)
- [Backwards Compatibility Promise for the API](#backwards-compatibility-promise-for-the-api)
- [Assertion](#assertion)
  - [Assertion Package for Unit Testing](#assertion-package-for-unit-testing)
  - [Assertion Package for Runtime Use](#assertion-package-for-runtime-use)
- [Code Snippets](#code-snippets)
- [Background](#background)
- [Learnings by so far](#learnings-by-so-far)
- [Support](#support)
- [Roadmap](#roadmap)


## Structure

`err2` has the following package structure:
- The `err2` (main) package includes declarative error handling functions.
- The `try` package offers error checking functions.
- The `assert` package implements assertion helpers for **both** unit-testing
  and *design-by-contract*.

## Automatic Error Propagation

The current version of Go tends to produce too much error checking and too
little error handling. But most importantly, it doesn't help developers with
**automatic** error propagation, which would have the same benefits as, e.g.,
**automated** garbage collection or automatic testing:

> Automation is not just about efficiency but primarily about repeatability and
> resilience. -- Gregor Hohpe

Automatic error propagation is so important because it makes your code tolerant
of the change. And, of course, it helps to make your code error-safe: 

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
per function as you need.

#### Error Stack Tracing

The err2 offers optional stack tracing. It's automatic and optimized. Optimized
means that the call stack is processed before output. That means that stack
trace starts from where the actual error/panic is occurred, not where the error
is caught. You don't need to search for the line where the pointer was nil or
received an error. That line is in the first one you are see:

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
  or, for example:
err2.SetPanicTracer(log.Writer()) // stack panic trace to std logger
```

If no `Tracer` is set no stack tracing is done. This is the default because in
the most cases proper error messages are enough and panics are handled
immediately by a programmer.

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
// dosen't exist returned as `plugin.ErrNotExist`
```

For more information see the examples in the documentation of both functions.

## Backwards Compatibility Promise for the API

The `err2` package's API will be **backward compatible**. Before version
1.0.0 is released, the API changes occasionally, but **we promise to offer
automatic conversion scripts for your repos to update them for the latest API.**
We also mark functions deprecated before they become obsolete. Usually, one
released version before. We have tested this with a large code base in our
systems, and it works wonderfully.

More information can be found in the scripts' [readme file](./scripts/README.md).

## Assertion

The `assert` package is meant to be used for *design-by-contract-* type of
development where you set pre- and post-conditions for your functions. It's not
meant to replace the normal error checking but speed up the incremental hacking
cycle. The default mode is to return an `error` value that includes a formatted
and detailed assertion violation message. A developer gets immediate and proper
feedback, allowing cleanup of the code and APIs before the release.

The assert package offers a few pre-build *asserters*, which are used to
configure how the assert package deals with assert violations. The line below
exemplifies how the default asserter is set in the package.

```go
SetDefaultAsserter(AsserterToError | AsserterFormattedCallerInfo)
```

If you want to suppress the caller info (source file name, line number, etc.)
and get just the plain error messages from the asserts, you should set the
default asserter with the following line:

```go
SetDefaultAsserter(AsserterToError) // we offer separated flags for caller info
```

For certain type of programs this is the best way. It allows us to keep all the
error messages as simple as possible. And by offering option to turn additional
information on, which allows super users and developers get more technical
information when needed.

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
	assert.PushTester(t)
	defer assert.PopTester()

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
testing. **We can share the same assertions between runtime and test
execution.**

The only minus is that test coverage figures are too conservative. The code that
uses design-by-contract assertions is typically much more robust what the actual
test coverage results tell you. However, this's a well-known problem with test
coverage metric in general.

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
isn't perfect. We have run many benchmarks to try to minimise the performance
penalty this kind of mechanism might bring. We have focused on the _happy path_
analyses. If the performance of the *error path* is essential, don't use this
mechanism presented here. But be aware that if your code uses the **error path
as a part of algorithm itself something is wrong**.

**For happy paths** by using `try.ToX` error check functions **there are no
performance penalty at all**. However, the mandatory use of the `defer` might
prevent some code optimisations like function inlining. And still, we have cases
where using the `err2` and `try` package simplify the algorithm so that it's
faster than the return value if err != nil version. (See the benchmarks for
`io.Copy` in the repo)

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

- You don't seem to need '%w' wrapping. See the Go's official blog post what are
[cons](https://go.dev/blog/go1.13-errors) for that.
  > Do not wrap an error when doing so would expose implementation details.

## Support

The package has been in experimental mode quite long time. Since the Go generics
we are transiting towards more official mode. Currently we offer support by
GitHub Discussions. Naturally, any issues are welcome as well!

## Roadmap

### Version history

##### 0.1
- First draft (Summer 2019)

##### 0.2
- Code generation for type helpers

##### 0.3
- `Returnf` added, not use own transport type anymore but just `error`

##### 0.4
- Documentation update

##### 0.5
- Go modules are in use

##### 0.6.1
- `assert` package added, and new type helpers

##### 0.7.0
- Filter functions for non-errors like `io.EOF`

#### 0.8.0
- `try.To()`, **Start to use Go generics**
- `assert.That()` and other assert functions with the help of the generics

##### 0.8.1
- **bug-fix**: `runtime.Error` types are treated as `panics` now (Issue #1)

##### 0.8.3
- `try.IsXX()` bug fix
- Lots of new docs
- **Automatic Stack Tracing!**

##### 0.8.4
- **Optimized** Stack Tracing
- Documentation
- Benchmarks, other tests

##### 0.8.5
- Typo in `StackTraceWriter` fixed

##### 0.8.6
- Stack Tracing bug fixed
- URL helper restored until migration tool

##### 0.8.7
- **Auto-migration tool** to convert deprecated API usage for your repos
- `err2.Throwf` added

##### 0.8.8
- **Assertion package integrates with Go's testing system**
- Type variables removed

##### 0.8.9
- bug fixes
- Deprecations
- New Tracer API
- Preparing `err2` API for 1.0

##### 0.8.10
- New assertion functions and helpers for tests

##### 0.8.11
- remove deprecations
- New *global* err values and `try.IsXX` functions
- More documentation

##### 0.8.12
- New super **Handle** for most of the use cases to simplify the API
- **Deferred error handlers are 2x faster now**
- Restructuring internal pkgs
- New documentation and tests, etc.

##### 0.8.13
- **Bug-fix:** automatic error strings for methods
- Added API to set preferred error string *Formatter* or implement own

##### 0.8.14
- `err2.Handle` supports sentinel errors, can now stop panics
- `err2.Catch` has one generic API and it stops panics as default
- Deprecated `CatchTrace` and `CatchAll` which merged with `Catch`
- Auto-migration offered (similar to `go fix`)
- **Code snippets** added
- New assertion functions
- No direct variables in APIs (race), etc.

#### 0.9.0
- **Clean and simple API** 
- Removing deprecated functions:
    - Only `err2.Handle` for error returning functions
    - Only `err2.Catch` for function that doesn't return error
    - Please see `scripts/README.md' for *Auto-migration for your repos*
- Default `err2.SetPanicTracer(os.Stderr)` allows `defer err2.Catch()`


### Upcoming releases

##### 0.9.1
- More support for `assert` package for tests: plugins like nvim-go
- More support for wrapping multiple errors

##### 0.9.2 
- More documentation, reparing for some sort of marketing
