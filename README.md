# err2

[![test](https://github.com/lainio/err2/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/lainio/err2/actions/workflows/test.yml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/lainio/err2)](https://pkg.go.dev/mod/github.com/lainio/err2)
[![Go Report Card](https://goreportcard.com/badge/github.com/lainio/err2?style=flat-square)](https://goreportcard.com/report/github.com/lainio/err2)

<img src="https://github.com/lainio/err2/raw/master/logo/logo.png" width="100">

----

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

----

`go get github.com/lainio/err2`

- [Structure](#structure)
- [Performance](#performance)
- [Automatic Error Propagation](#automatic-error-propagation)
- [Error handling](#error-handling)
  - [Error Stack Tracing](#error-stack-tracing)
- [Error Checks](#error-checks)
  - [Filters for non-errors like io.EOF](#filters-for-non-errors-like-ioeof)
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
- [History](#history)


## Structure

`err2` has the following package structure:
- The `err2` (main) package includes declarative error handling functions.
- The `try` package offers error checking functions.
- The `assert` package implements assertion helpers for **both** unit-testing
  and *design-by-contract* with the *same API and cross-usage*.

## Performance

All of the listed above **without any performance penalty**! You are welcome to
run `benchmarks` in the project repo and see yourself.

<details>
<summary><b>It's too fast!</b></summary>
<br/>

> Most of the benchmarks run 'too fast' according to the common Go
> benchmarking rules, i.e., compiler optimizations
> ([inlining](https://en.wikipedia.org/wiki/Inline_expansion)) are working so
> well that there are no meaningful results. But for this type of package, where
> **we compete with if-statements, that's precisely what we hope to achieve.**
> The whole package is written toward that goal. Especially with parametric
> polymorphism, it's been quite the effort.

</details>

## Automatic Error Propagation

Automatic error propagation is crucial because it makes your *code change
tolerant*. And, of course, it helps to make your code error-safe.

![Never send a human to do a machine's job](https://www.magicalquote.com/wp-content/uploads/2013/10/Never-send-a-human-to-do-a-machines-job.jpg)


<details>
<summary>The err2 package is your automation buddy:</summary>
<br/>

1. It helps to declare error handlers with `defer`. If you're familiar with [Zig
   language](https://ziglang.org/), you can think `defer err2.Handle(&err,...)`
   line exactly similar as
   [Zig's `errdefer`](https://ziglang.org/documentation/master/#errdefer).
2. It helps to check and transport errors to the nearest (the defer-stack) error
   handler.
3. It helps us use design-by-contract type preconditions.
4. It offers automatic stack tracing for every error, runtime error, or panic.
   If you are familiar with Zig, the `err2` error return traces are same as
   Zig's.

You can use all of them or just the other. However, if you use `try` for error
checks, you must remember to use Go's `recover()` by yourself, or your error
isn't transformed to an `error` return value at any point.

</details>

## Error Handling

The err2 relies on Go's declarative programming structure `defer`. The
err2 helps to set deferred error handlers which are only called if an error
occurs.

This is the simplest form of an automatic error handler:

```go
func doSomething() (err error) {
    defer err2.Handle(&err)
```

<details>
<summary>The explanation of the above code and its error handler:</summary>
<br/>

Simplest rule for err2 error handlers are:
1. Use named error return value: `(..., err error)`
1. Add at least one error handler at the beginning of your function (see the
   above code block). *Handlers are called only if error ≠ nil.*
1. Use `err2.handle` functions different calling schemes to achieve needed
   behaviour. For example, without no extra arguments `err2.Handle`
   automatically annotates your errors by building annotations string from the
   function's current name: `doSomething → "do something:"`. Default is decamel
   and add spaces. See `err2.SetFormatter` for more information.
1. Every function which uses err2 for error-checking should have at least one
   error handler. The current function panics if there are no error handlers and
   an error occurs. However, if *any* function above in the call stack has an
   err2 error handler, it will catch the error.

See more information from `err2.Handle`'s documentation. It supports several
error-handling scenarios. And remember that you can have as many error handlers
per function as you need. You can also chain error handling functions per
`err2.Handle` that allows you to build new error handling middleware for your
own purposes.

</details>

#### Error Stack Tracing

The err2 offers optional stack tracing in two different formats:
1. Optimized call stacks (`-err2-trace`)
1. Error return traces similar to Zig (`-err2-ret-trace`)

Both are *automatic* and fully *optimized*. 

<details>
<summary>The example of the optimized call stack:</summary>
<br/>

Optimized means that the call stack is processed before output. That means that
stack trace *starts from where the actual error/panic is occurred*, not where
the error or panic is caught. You don't need to search for the line where the
pointer was nil or received an error. That line is in the first one you are
seeing:

```
---
runtime error: index out of range [0] with length 0
---
goroutine 1 [running]:
main.test2({0x0, 0x0, 0x40XXXXXf00?}, 0x2?)
	/home/.../go/src/github.com/lainio/ic/main.go:43 +0x14c
main.main()
	/home/.../go/src/github.com/lainio/ic/main.go:77 +0x248
```

</details>

Just set the `err2.SetErrorTracer`, `err2.SetErrRetTracer` or
`err2.SetPanicTracer` to the stream you want traces to be written:

```go
err2.SetErrorTracer(os.Stderr) // write error stack trace to stderr
// or, for example:
err2.SetErrRetTracer(os.Stderr) // write error return trace (like Zig)
// or, for example:
err2.SetPanicTracer(log.Writer()) // stack panic trace to std logger
```

If no `Tracer` is set no stack tracing is done. This is the default because in
the most cases proper error messages are enough and panics are handled
immediately by a programmer.

> [!NOTE]
> Since v0.9.5 you can set *tracers* through Go's standard flag package just by
> adding `flag.Parse()` call to your source code. See more information from
> [Automatic Flags](#automatic-flags).

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

<details>
<summary><b>Immediate Error Handling Options</b></summary>
<br/>

In cases where you want to handle the error immediately after the function call
you can use Go's default `if` statement. However, we recommend you to use 
`defer err2.Handle(&err)` for all of your error handling, because it keeps your
code modifiable, refactorable, and skimmable.

Nevertheless, there might be cases where you might want to:
1. Suppress the error and use some default value. In next, use 100 if `Atoi`
   fails:
   ```go
   b := try.Out1(strconv.Atoi(s)).Catch(100)
   ```
1. Just write logging output and continue without breaking the execution. In
   next, add log if `Atoi` fails.
   ```go
   b := try.Out1(strconv.Atoi(s)).Logf("%s => 100", s).Catch(100)
   ```
1. Annotate the specific error value even when you have a general error handler.
   You are already familiar with `try.To` functions. There's *fast* annotation
   versions `try.T` which can be used as shown below:
   ```go
   b := try.T1(io.ReadAll(r))("cfg file read")
   // where original were, for example:
   b := try.To1(io.ReadAll(r))
   ```
1. You want to handle the specific error value at the same line or statement. In
   below, the function `doSomething` returns an error value. If it returns
   `ErrNotSoBad`, we just suppress it. All the other errors are send to the
   current error handler and will be handled there, but are also annotated with
   'fatal' prefix before that here.
   ```go
   try.Out(doSomething()).Handle(ErrNotSoBad, err2.Reset).Handle("fatal")
   ```

The `err2/try` package offers other helpers based on the error-handling
language/API. It's based on functions `try.Out`, `try.Out1`, and `try.Out2`,
which return instances of types `Result`, `Result1`, and `Result2`. The
`try.Result` is similar to other programming languages, i.e., discriminated
union. Please see more from its documentation.

It's easy to see that panicking about the errors at the start of the development
is far better than not checking errors at all. But most importantly, `err2/try`
**keeps the code readable.**

</details>

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

> [!NOTE] 
> Any other error than `plugin.ErrNotExist` is treated as an real error:
> 1. `try.Is` function first checks `if err == nil`, and if yes, it returns
>    `false`.
> 2. Then it checks if `errors.Is(err, plugin.ErrNotExist)` and if yes, it returns
>    `true`.
> 3. Finally, it calls `try.To` for the non nil error, and we already know what then
>    happens: nearest `err2.Handle` gets it first.

For more information see the examples in the documentation of both functions.

## Assertion

The `assert` package is meant to be used for *design-by-contract-* type of
development where you set pre- and post-conditions for *all* of your functions,
*including test functions*. These asserts are as fast as if-statements when not
triggered.

> [!IMPORTANT]
> It works *both runtime and for tests.* And even better, same asserts work in
> both running modes.

#### Asserters

<details>
<summary><b>Fast Clean Code with Asserters</b></summary>
<br/>

Asserts are not meant to replace the normal error checking but speed up the
incremental hacking cycle like TDD. The default mode is to return an `error`
value that includes a formatted and detailed assertion violation message. A
developer gets immediate and proper feedback independently of the running mode,
allowing very fast feedback cycles.

The assert package offers a few pre-build *asserters*, which are used to
configure *how the assert package deals with assert violations*. The line below
exemplifies how the default asserter is set in the package. (See the
documentation for more information about asserters.)

```go
assert.SetDefault(assert.Production)
```

If you want to suppress the caller info (source file name, line number, etc.)
from certain asserts, you can do that per a goroutine or a function. You should
set the asserter with the following line for the current function:

```go
defer assert.PushAsserter(assert.Plain)()
```

This is especially good if you want to use assert functions for CLI's flag
validation or you want your app behave like legacy Go programs.

</details>

> [!NOTE]
> Since v0.9.5 you can set these asserters through Go's standard flag package
> just by adding `flag.Parse()` to your program. See more information from
> [Automatic Flags](#automatic-flags).

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

The same asserts can be used **and shared** during the unit tests over module
boundaries.

<details>
<summary>The unit test code example:</summary>

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

</details>

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

You can deploy your applications and services with the simple *end-user friendly
error messages and no stack traces.*

<details>
<summary>You can switch them on whenever you need them again.</summary>
<br/>

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
</details>

#### Support for Cobra Flags

If you are using [cobra](https://github.com/spf13/cobra) you can still easily
support packages like `err2` and `glog` and their flags.

<details>
<summary>Add cobra support:</summary>
<br/>

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

</details>

## Code Snippets

<details>
<summary>Code snippets as learning helpers.</summary>
<br/>

The snippets are in `./snippets` and in VC code format, which is well supported
e.g. neovim, etc. They are proven to be useful tool especially when you are
starting to use the err2 and its sub-packages.

The snippets must be installed manually to your preferred IDE/editor. During the
installation you can modify the according your style or add new ones. We would
prefer if you could contribute some of the back to the err2 package.
</details>

## Background

<details>
<summary>Why this repo exists?</summary>
<br/>

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
</details>

## Learnings by so far

<details>
<summary>We have used the err2 and assert packages in several projects. </summary>
<br/>

The results have been so far very encouraging:

- If you forget to use handler, but you use checks from the package, you will
get panics on errors (and optimized stack traces that can be suppressed). That
is much better than getting unrelated panic somewhere else in the code later.
There have also been cases when code reports error correctly because the 'upper'
handler catches it.

- Because the use of `err2.Handle` is so easy, error messages are much better
and informative. When using `err2.Handle`'s automatic annotation your error
messages are always up-to-date. Even when you refactor your function name error
message is also updated.

- **When error handling is based on the actual error handlers, code changes have
been much easier.** There is an excellent [blog post](https://jesseduffield.com/Gos-Shortcomings-1/)
about the issues you are facing with Go's error handling without the help of
the err2 package.

- If you don't want to bubble up error from every function, we have learned that
`Try` prefix convention is pretty cool way to solve limitations of Go
programming language help to make your code more skimmable. If your internal
functions normally would be something like `func CopyFile(s, t string) (err
error)`, you can replace them with `func TryCopyFile(s, t string)`, where `Try`
prefix remind you that the function throws errors. You can decide at what level
of the call stack you will catch them with `err2.Handle` or `err2.Catch`,
depending your case and API.

</details>

## Support And Contributions

The package was in experimental mode quite long time. Since the Go generics we
did transit to official mode. Currently we offer support by GitHub Issues and
Discussions. Naturally, we appreciate all feedback and contributions are
very welcome!

## History

Please see the full version history from [CHANGELOG](./CHANGELOG.md).

### Latest Release

##### 1.2.0
- Now `-err2-ret-trace` and `err2.SetErrRetTracer` gives us *error return traces*
  which are even more readable than `-err2-trace`, `err2.SetErrorTracer` with
  long error return traces
- A new automatic error formatter/generator added for `TryCopyFile` convention
- Better documentation
