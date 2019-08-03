#err2

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

**but not without the handler**.

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

The original goal was to make it possible to write similar code than Go2 and
do it right now. The goal seems to be more than valid now when the latest Go
proposal is to have a `try` macro and let the error handling be implemented in
defer blocks.

## Roadmap

Version history:
- 0.1, first draft (Current)

When Go2 `try` macro is released this package should be updated accordingly.
However, the actual effect of such an update seems to be minor, but the
`err.Try/Check` calls should be **changed** to `try` macro calls. The err2 handlers
should work as is.