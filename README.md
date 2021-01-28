# err2

The package provides simple helper functions for _automatic_ error propagation.

`go get github.com/lainio/err2`


## Error Propagation

The current version of Go tends to produce too much error checking and too little error handling. This package helps us fix that.
1. It helps to declare error handlers with `defer`.
2. It helps to check and transport errors to the nearest (the defer-stack) error handler. 

You can use both of them or just the other. However, if you use `err2` for error checks you must remember use Go's `recover()` by yourself, or your error isn't transformed to an `error`.

## Error handling

Package `err2` relies on Go's declarative programming structure `defer`. The `err2` helps to set deferred functions (error handlers) which are only called if `err != nil`.

In every function which uses err2 for error-checking should have at least one error handler. If there are no error handlers and error occurs the current function panics. However, if function above in the call stack has `err2` error handler it will catch the error. The panicking for the errors at the start of the development is far better than not checking errors at all.

This is the simplest `err2` error handler
```go
defer err2.Return(&err)
```
which is the helper handler for cases that don't need to annotate the error. If you need to annotate the error you can use either `Annotate` or `Returnf`.

#### Error Handler
The `err2.Handle` is a helper function to add actual error handlers which are called only if an error has occurred. In most real-world cases, we have multiple error checks and only one or just a few error handlers. However, you can have as many error handlers per function as you need.

[Read the package documentation for more information](https://pkg.go.dev/github.com/lainio/err2).

## Error checks

The `err2` provides convenient helpers to check the errors.

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
b := err2.Bytes.Try(ioutil.ReadAll(r))
...
```

but not without an error handler (`Return`, `Annotate`, `Handle`) or it just panics your app if you don't have a `recovery` call in the goroutines calls stack.


#### Type Helpers

The package includes performance optimised versions of `Try` function where the actual return types of the checked function are told to `err2` package by *type helper variables*. This removes the need to use dynamic type conversion. However, *when Go2 generics are out we can replace all of these with generics*.

The sample call with type helper variable
```go
data := err2.Bytes.Try(ioutil.ReadAll(r))
```
The err2 package includes a CLI tool to generate these helpers for your own types. Please see the Makefile for example.

## Assertion (design by contract)

The `assert` package has been since version 0.6. The package is meant to be used for design by contract -type of development where you set preconditions for your functions. It's not meant to replace normal error checking but speed up incremental hacking cycle. That's the reason why default mode (`var D Asserter`) is to panic. By panicking developer get immediate and proper feedback which allows cleanup the code and APIs before actual production release.

```go
func marshalAttestedCredentialData(json []byte, data *protocol.AuthenticatorData) []byte {
	assert.D.EqualInt(len(data.AttData.AAGUID), 16, "wrong AAGUID length")
	assert.D.NotEmpty(data.AttData.CredentialID, "empty credential id")
	assert.D.NotEmpty(data.AttData.CredentialPublicKey, "empty credential public key")
	...
```

Previous code block shows the use of the asserter (`D`) for developing. If any of the assertion fails, code panics. These type of assertions can be used without help of the `err2` package.

During the software development lifecycle it isn't all the time crystal clear what are preconditions for a programmer and what should be translated to end-user errors as well. That's why `assert` package uses concept called `Asserter` to have different type of asserter for different phases of a software project.

The following code block is a sample where production time asserter is used to generate proper error messages.

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

When asserts are used to generate end-user error messages instead of immediate panics, `err2` handlers are needed to translate asserts to errors in convenient way. That's the reason we decided to build `assert` as a sub package of `err2` even there are no real dependencies between them. See the `assert` packages own documentation and examples for more information.

## Background
`err2` implements similar error handling mechanism as drafted in the original [check/handle proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md). The package does it by using internally `panic/recovery`, which some might think isn't perfect. We have run many benchmarks to try to minimise the performance penalty this kind of mechanism might bring. We have focused on the _happy path_ analyses. If the performance of the error path is essential, don't use this mechanism presented here. For happy paths by using `err2.Check` type helper variables there seems to be no performance penalty.

However, the mandatory use of the `defer` might prevent some code optimisations like function inlining. If you have a performance-critical use case, we recommend you to write performance tests to measure the effect.

The original goal was to make it possible to write similar code than proposed Go2 error handling would allow and do it right now. The goal was well aligned with the latest Go2 proposal where it would bring a `try` macro and let the error handling be implemented in defer blocks. The try-proposal was put on the hold or cancelled at its latest form. However, we have learned that using panics for early-stage error transport isn't a bad thing but opposite. It seems to help to draft algorithms.

## Learnings by so far

We have used the `err2` and `assert` packages in several internal projects. The results have been so far very encouraging:

- If you forget to use handler, but you use checks from the package, you will get panics if an error occurs. That is much better than getting unrelated panic somewhere else in the code later. There have also been cases when code reports error correctly because the 'upper' handler catches it.
- Because the use of `err2.Annotate` is so relatively easy, error messages much better and informative.
- When error handling is based on the actual error handlers, code changes have been much easier.

## Roadmap

Version history:
- 0.1, first draft (Summer 2019)
- 0.2, code generation for type helpers
- 0.3, `Returnf` added, not use own transport type anymore but just `error`
- 0.4, Documentation update
- 0.5, Go modules are in use now
- 0.6.1, `assert` package added, and new type helpers (current)


We will update both packages when Go2 generics are released. There is already a working version which uses Go generics. That has been shown that the switch will be prompt. We are also monitoring what will happen for the Go-native error handling and tune the library accordingly.
