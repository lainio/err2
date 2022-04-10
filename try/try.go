/*
Package try is a package for `try to` functions that implement the error
checking. 'try.To' functions check if err != nil and if it throws the err to the
error handlers, which are implemented by the err2 package. More information
about err2 and try packager roles can be seen in the FileCopy example from err2:
  ...
  r := try.To1(os.Open(src))
  defer r.Close()

  w := try.To1(os.Create(dst))
  defer err2.Handle(&err, func() {
  	os.Remove(dst)
  })
  defer w.Close()
  try.To1(io.Copy(w, r))
  return nil
  ...

All of the try package functions are as fast as the simple 'if err != nil {'
statement, thanks to the compiler inlining and optimization.
*/
package try

import (
	"errors"
	"io"
)

// To is a helper function to call functions which returns (error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To(err error) {
	if err != nil {
		panic(err)
	}
}

// To1 is a helper function to call functions which returns (any, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To1[T any](v T, err error) T {
	To(err)
	return v
}

// To2 is a helper function to call functions which returns (any, any, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	To(err)
	return v1, v2
}

// To3 is a helper function to call functions which returns (any, any, any, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	To(err)
	return v1, v2, v3
}

// Is-function performs a filtered error check for the given argument. It's the
// same as To-function, but it checks if the error matches the filter before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the filter.
func Is(err, filter error) bool {
	if err != nil {
		if errors.Is(err, filter) {
			return true
		}
		panic(err)
	}
	return false
}

// IsEOF1-function performs a filtered error check for the given argument. It's the
// same as To-function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the 'io.EOF'.
func IsEOF1[T any](v T, err error) (bool, T) {
	isFilter := Is(err, io.EOF)
	return isFilter, v
}

// IsEOF2-function performs a filtered error check for the given argument. It's the
// same as To-function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the 'io.EOF'.
func IsEOF2[T, U any](v1 T, v2 U, err error) (bool, T, U) {
	isFilter := Is(err, io.EOF)
	return isFilter, v1, v2
}

// IsEOF-function performs a filtered error check for the given argument. It's the
// same as To-function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the 'io.EOF'.
func IsEOF(err error) bool {
	return Is(err, io.EOF)
}
