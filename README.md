# err2

The package provides simple helper functions for error handling.

`go get github.com/lainio/err2`

The traditional error handling idiom in Go is roughly akin to
```go
if err != nil {
        return err
}
```

which applied recursively. That leads to problems like code noise, redundancy,
or even non-checks. The err2 package drives programmers more to focus on
**error handling** rather than checking errors.

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

**but not without a handler**.

## Error handling

Package err2 relies on error handlers. **In every function which uses err2 for
error-checking has to have at least one error handler**. If there are no error
handlers and error occurs it panics. Panicking for the errors during the
development is far better than not checking errors at all.

The handler for the previous sample is
```go
defer err2.Return(&err)
```
which is the helper handler for cases that don't annotate errors.
`err2.Handle` is a helper function to add needed error handlers to defer stack.
In most real-world cases, we have multiple error checks and only one or just a
few error handlers.

[Read the package documentation for more information](https://godoc.org/github.com/lainio/err2).

## Background
err2 implements similar error handling mechanism as drafted in the original
[check/handle
proposal](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md).
The package does it using internally `panic/recovery`, which does not make it
perfect. We have run many benchmarks to try to minimise performance penalty this
kind of mechanism might bring. We have focused on the happy path analyses. If the 
performance of the error path is essential, don't use this mechanism presented
here. For happy paths by using `err2.Check` there seems to be no performance
penalty.

However, the mandatory use of the `defer` might prevent some code optimise like
function inlining. If you have a performance critical use case we recommend to
write a performance tests to measure the effect.

The original goal was to make it possible to write similar code than proposed
Go2 error handling would allow and do it right now. The goal was well 
aligned with the latest Go proposal where it would brought a `try` macro and let
the error handling be implemented in defer blocks. Unfortunately the try-
proposal was put on the hold or cancelled at its latest form. 

## Learning by so far

We have used the err2 package in several internal projects. The results have
been so far very encouraging:

- If you forget to use handler but you use checks from the package, you will get
panics if error occurs. That is much better than getting unrelated panic later.
There has been even cases when code reports error correctly because the 'upper'
handler catches it.

- Because the use of `err2.Annocate()` is so relatively easy, developers use it
always which makes error messages much better and informative. Could say, they
include a logical call stack in the user friendly form.

- There has been a couple of the cases when a quite complex function has needed
update. When error handling is based on the actual error handlers, not just
passing them up in the stack, the code changes have been much easier.

## Roadmap

Version history:
- 0.1, first draft (Summer 2019)
- 0.2, code generation for helpers (Current)

When Go2 `try` macro would be released this package should be updated
accordingly. However, the actual effect of such an update seems to be minor, but
the `err.Try/Check` calls should be **changed** to `try` macro calls. The err2
handlers should work as is.

We are monitoring what will happened for the Go-native error handling and tune
the library accordingly.