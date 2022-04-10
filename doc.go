/*
Package err2 provides three main functionality:
  1. err2 package includes helper functions for error handling.
  2. try package is for error checking
  3. assert package is for design-by-contract and preconditions

The traditional error handling idiom in Go is roughly akin to

 if err != nil { return err }

which applied recursively.

The err2 package drives programmers more to focus on error handling rather than
checking errors. We think that checks should be so easy that we never forget
them.

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

(help of the declarative control structures)

Error checks

The err2/try provides convenient helpers to check the errors. For example,
instead of

 b, err := ioutil.ReadAll(r)
 if err != nil {
 	return err
 }

we can write

 b := try.To1(ioutil.ReadAll(r))

but not without the handler.

Error handling

Package err2 relies on error handlers. In every function which uses err2 or try
package for error-checking has to have at least one error handler. If there are
no error handlers and error occurs it panics. Nevertheless, we think that
panicking for the errors during the development is much better than not checking
errors at all. However, if the call stack includes any err2 error handlers like
err2.Handle() the error is handled there where the handler is saved to defer
stack.

The handler for the previous sample is

 defer err2.Return(&err)

which is the helper handler for cases that don't annotate errors.
err2.Handle is a helper function to add needed error handlers to defer stack.
In most real-world cases, we have multiple error checks and only one or just a
few error handlers per function. And if whole control flow is thought the ratio
is even greater.
*/
package err2
