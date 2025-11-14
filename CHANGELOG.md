## Changelog

### Version history

##### 1.2.2
- Bug Fix (issue-27): automatic error annotation works now for try.T functions
- Updated documentation

##### 1.2.1
- Optimization and Refactoring 
- Updated documentation

##### 1.2.0
- Now `-err2-ret-trace` and `err2.SetErrRetTracer` gives us *error return traces*
  which are even more readable than `-err2-trace`, `err2.SetErrorTracer` with
  long error return traces
- A new automatic error formatter/generator added for `TryCopyFile` convention
- New features for `sample/` to demonstrate latest features
- Extended documentation

##### 1.1.0
- `assert` package:
    - bug fix: call stack traversal during unit testing in some situations
    - **all generics-based functions are inline expansed**
    - *performance* is now *same as if-statements for all functions*
    - new assert functions: `MNil`, `CNil`, `Less`, `Greater`, etc.
    - all assert messages follow Go idiom: `got, want`
    - `Asserter` can be set per goroutine: `PushAsserter`
- `try` package:
    - new check functions: `T`, `T1`, `T2`, `T3`, for quick refactoring from `To` functions to annotate an error locally
    - **all functions are inline expansed**: if-statement equal performance

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

##### 0.9.52
- `err2.Stderr` helpers for `Catch/Handle` to direct auto-logging + snippets
- `assert` package `Shorter` `Longer` helpers for automatic messages
- `asserter` package remove deprecated slow reflection based funcs
- cleanup and refactoring for sample apps

##### 0.9.51
- `flag` package support to set `err2` and `assert` package configuration
- `err2.Catch` default mode is to log error
- cleanup and refactoring, new tests and benchmarks

##### 0.9.5 **mistake in build number: 5 < 41**
- `flag` package support to set `err2` and `assert` package configuration
- `err2.Catch` default mode is to log error
- cleanup and refactoring, new tests and benchmarks

##### 0.9.41
- Issue #18: **bug fixed**: noerr-handler had to be the last one of the err2
  handlers

##### 0.9.40
- Significant performance boost for: `defer err2.Handle/Catch()`
  - **3x faster happy path than the previous version, which is now equal to
    simplest `defer` function in the `err`-returning function** . (Please see
    the `defer` benchmarks in the `err2_test.go` and run `make bench_reca`)
  - the solution caused a change to API, where the core reason is Go's
    optimization "bug". (We don't have confirmation yet.)
- Changed API for deferred error handling: `defer err2.Handle/Catch()`
  - *Obsolete*:
    ```go
    defer err2.Handle(&err, func() {}) // <- relaying closure to access err val
    ```
  - Current version:
    ```go
    defer err2.Handle(&err, func(err error) error { return err }) // not a closure
    ```
    Because handler function is not relaying closures any more, it opens a new
    opportunity to use and build general helper functions: `err2.Noop`, etc.
  - Use auto-migration scripts especially for large code-bases. More information
    can be found in the `scripts/` directory's [readme file](./scripts/README.md).
  - Added a new (*experimental*) API:
    ```go
    defer err2.Handle(&err, func(noerr bool) {
            assert.That(noerr) // noerr is always true!!
            doSomething()
    })
    ```
    This is experimental because we aren't sure if this is something we want to
    have in the `err2` package.
- Bug fixes: `ResultX.Logf()` now works as it should
- More documentation

##### 0.9.29
- New API for immediate error handling: `try out handle/catch err`
  `val := try.Out1strconv.Atois.Catch(10)`
- New err2.Catch API for automatic logging
- Performance boost for assert pkg: `defer assert.PushTester(t)()`
- Our API has now *all the same features Zig's error handling has*

##### 0.9.1
- **Performance boost for assert pkg**: `assert.That(boolVal)` == `if boolVal`
- Go version 1.18 is a new minimum (was 1.19, use of `atomic.Pointer`)
- Generic functions support type aliases
- More support for `assert` package for tests: support for cross module asserts
  during the tests
- Using `assert` pkg for tests allow us to have **traversable call stack
  during unit tests** -- cross module boundaries solved
- Implementation: simplified `assert` pkg to `testing` pkg integration, and
  especially performance

##### 0.9.0
- **Clean and simple API**
- Removing deprecated functions:
    - Only `err2.Handle` for error returning functions
    - Only `err2.Catch` for function that doesn't return error
    - Please see `scripts/README.md` for *Auto-migration for your repos*
- Default `err2.SetPanicTracer(os.Stderr)` allows `defer err2.Catch()`

##### 0.8.14
- `err2.Handle` supports sentinel errors, can now stop panics
- `err2.Catch` has one generic API and it stops panics as default
- Deprecated `CatchTrace` and `CatchAll` which merged with `Catch`
- Auto-migration offered (similar to `go fix`)
- **Code snippets** added
- New assertion functions
- No direct variables in APIs (race), etc.

##### 0.8.13
- **Bug-fix:** automatic error strings for methods
- Added API to set preferred error string *Formatter* or implement own

##### 0.8.12
- New super **Handle** for most of the use cases to simplify the API
- **Deferred error handlers are 2x faster now**
- Restructuring internal pkgs
- New documentation and tests, etc.

##### 0.8.11
- remove deprecations
- New *global* err values and `try.IsXX` functions
- More documentation

##### 0.8.10
- New assertion functions and helpers for tests

##### 0.8.9
- bug fixes
- Deprecations
- New Tracer API
- Preparing `err2` API for 1.0

##### 0.8.8
- **Assertion package integrates with Go's testing system**
- Type variables removed

##### 0.8.7
- **Auto-migration tool** to convert deprecated API usage for your repos
- `err2.Throwf` added

##### 0.8.6
- Stack Tracing bug fixed
- URL helper restored until migration tool

##### 0.8.5
- Typo in `StackTraceWriter` fixed

##### 0.8.4
- **Optimized** Stack Tracing
- Documentation
- Benchmarks, other tests

##### 0.8.3
- `try.IsXX()` bug fix
- Lots of new docs
- **Automatic Stack Tracing!**

##### 0.8.1
- **bug-fix**: `runtime.Error` types are treated as `panics` now (Issue #1)

##### 0.8.0
- `try.To()`, **Start to use Go generics**
- `assert.That()` and other assert functions with the help of the generics

##### 0.7.0
- Filter functions for non-errors like `io.EOF`

##### 0.6.1
- `assert` package added, and new type helpers

##### 0.5
- Go modules are in use

##### 0.4
- Documentation update

##### 0.3
- `Returnf` added, not use own transport type anymore but just `error`

##### 0.2
- Code generation for type helpers

##### 0.1
- First draft (Summer 2019)

