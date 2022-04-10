/*
Package err2 provides three main functionality:
  1. err2 package includes helper functions for error handling.
  2. try package is for error checking
  3. assert package is for design-by-contract and preconditions

The traditional error handling idiom in Go is roughly akin to

 if err != nil { return err }

which applied recursively.

The err2 package drives programmers to focus more on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them. The CopyFile example shows how it works:

 func CopyFile(src, dst string) (err error) {
      defer err2.Returnf(&err, "copy %s %s", src, dst)

      // These try package helpers are as fast as Check() calls which is as
      // fast as `if err != nil {}`

      r := try.To1(os.Open(src))
      defer r.Close()

      w := try.To1(os.Create(dst))
      defer err2.Handle(&err, func() {
      	os.Remove(dst)
      })
      defer w.Close()
      try.To1(io.Copy(w, r))
      return nil
 }

Error checks

The `try` package provides convenient helpers to check the errors. For example,
instead of

 b, err := ioutil.ReadAll(r)
 if err != nil {
    return err
 }

we can write

 b := try.To1(ioutil.ReadAll(r))

Note that try.ToX functions are as fast as if err != nil statements. Please see
the try package documentation for more information about the error checks.

Error handling

Package err2 relies on declarative control structures to achieve error and panic
safety. In every function which uses err2 or try package for error-checking has
to have at least one declarative error handler if it returns error value. If
there are no error handlers and error occurs it panics. We think that panicking
for the errors is much better than not checking errors at all. Nevertheless, if
the call stack includes any err2 error handlers like err2.Handle the error is
handled where the handler is saved to defer-stack. (defer is not lexically
scoped)

err2 includes many examples to play with like previous CopyFile. Please see them
for more information.
*/
package err2
