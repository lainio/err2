// Package try is a new sub package for `try to` functions which replace all of
// the error checking. The error handlers still stay in the main package err2.
// For more information see FileCopy example in err2:
//  ...
//  r := try.To1(os.Open(src))
//  defer r.Close()
//
//  w := try.To1(os.Create(dst))
//  defer err2.Handle(&err, func() {
//  	os.Remove(dst)
//  })
//  defer w.Close()
//  try.To1(io.Copy(w, r))
//  return nil
//  ...
package try

import (
	"errors"
)

// To is a helper function to call functions which returns (error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To(err error) {
	if err != nil {
		panic(err)
	}
}

// To1 is a helper function to call functions which returns (any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To1[T any](v T, err error) T {
	To(err)
	return v
}

// To2 is a helper function to call functions which returns (any, any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	To(err)
	return v1, v2
}

// To3 is a helper function to call functions which returns (any, any, any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	To(err)
	return v1, v2, v3
}

// Is performs filtered error check for the given argument. It's same
// as To but before throwing an error it checks if error matches the filter.
// The return value false tells that there are no errors and true that filter is
// matched.
func Is(filter, err error) bool {
	if err != nil {
		if errors.Is(filter, err) {
			return true
		}
		panic(err)
	}
	return false
}
// TODO: add ToIsX() & ToAsX() funcs to support errors.Is & errors.As IFF we
// will support wrapping at all.
