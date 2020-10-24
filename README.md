# err2

The package provides simple helper functions for error propagation.

`go get github.com/lainio/err2`

##Error Propagation

The current version of Go tends to produce too much error checking and too little error handling. This package helps you fix that.
1. It helps to declare error handlers with `defer`.
2. It helps to check and transport errors to the nearest error handler with `panic` 

You can use both of them or just to the other. However, if you use `err2` for error checks you must remember use `recover` by yourself, or your error isn't transformed to an `error`.

## Error handling

Package `err2` relies on the declarative programming to add error handlers. The `err2` error handlers are only called if `err != nil` which makes them convenient to use and reduce boilerplate. **To error handlers to work, you must name the error return variable.**

**In every function which uses err2 for error-checking should have at least one error handler**. If there are no error handlers and error occurs it just panics. However, if function above in the call stack has `err2` error handler it will catch the error. The panicking for the errors **at the start of the development is far better than not checking errors at all**.

This is the simplest `err2` error handler
```go
defer err2.Return(&err)
```
which is the helper handler for cases that don't need to annotate the error. If you need to annotate the error you can use either `Annotate` or `Returnf`.

The `Handle` is a helper function to add actual error handlers. These handlers are called only if an error has occurred. In most real-world cases, we have multiple error checks and only one or just a few error handlers. However, **you can have as many error handlers per function as you need**.

[Read the package documentation for more information](https://godoc.org/github.com/lainio/err2).

## Error checks

The err2 provides convenient helpers to check the errors.

For example, instead of
```go
_, err := ioutil.ReadAll(r)
if err != nil {
        return err
}
```
we can call
```go
err2.Try(ioutil.ReadAll(r))
```

**but not without an error handler (`Return`, `Annote`, `Handle`) or it just panics your app** if you don't have a `recovery` call in the goroutines calls stack.

####Type Helpers

The package includes performance optimised versions of `Try` function where the actual return types of the checked function are told to `err2` package by *type helper variables*. This removes the need to use dynamic type conversion. However, *when Go2 generics are out we can replace all of these with generics*.

The sample call with type helper variable
```go
data := err2.Bytes.Try(ioutil.ReadAll(r))
```
The err2 package includes a CLI tool to generate these helpers for your own types. Please see the Makefile for example.


## Background
err2 implements similar error handling mechanism as drafted in the original [check/handle proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md). The package does it by using internally `panic/recovery`, which some might think isn't perfect. We have run many benchmarks to try to minimise the performance penalty this kind of mechanism might bring. We have focused on the happy path analyses. If the performance of the error path is essential, don't use this mechanism presented here. For happy paths by using `err2.Check` type helper variables there seems to be no performance penalty.

However, the mandatory use of the `defer` might prevent some code optimisations like function inlining. If you have a performance-critical use case we recommend you to write performance tests to measure the effect.

The original goal was to make it possible to write similar code than proposed Go2 error handling would allow and do it right now. The goal was well aligned with the latest Go2 proposal where it would bring a `try` macro and let the error handling be implemented in defer blocks. The try-proposal was put on the hold or cancelled at its latest form. However, we have learned that using panics for early-stage error transport isn't a bad thing but opposite. It seems to help to draft algorithms.

## Learnings by so far

We have used the err2 package in several internal projects. The results have been so far very encouraging:

- If you forget to use handler, but you use checks from the package, you will get panics if an error occurs. That is much better than getting unrelated panic later. There have also been cases when code reports error correctly because the 'upper' handler catches it.
- Because the use of `err2.Annocate` is so relatively easy, developers use it always which makes error messages much better and informative. Could say that by this they include a logical call stack in the user-friendly form.
- There has been a couple of the cases when a quite complex function has needed update. When error handling is based on the actual error handlers, not just passing them up in the stack, the code changes have been much easier. More importantly, because there are no DRY violations fixing is easier.

## Roadmap

Version history:
- 0.1, first draft (Summer 2019)
- 0.2, code generation for type helpers
- 0.3, `Returnf` added, not use own transport type anymore but just `error`
- 0.4, Documentation update
- 0.5, Go modules are in use now (current)


We will update this package when Go2 generics are released. There is a working version which uses Go generics. We are also monitoring what will happen for the Go-native error handling and tune the library accordingly.
