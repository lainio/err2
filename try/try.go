/*
Package try is a package for try.ToX functions that implement the error
checking. try.ToX functions check 'if err != nil' and if it throws the err to the
error handlers, which are implemented by the err2 package. More information
about err2 and try packager roles can be seen in the FileCopy example:

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

# try.To — Fast Checking

All of the try.To functions are as fast as the simple 'if err != nil {'
statement, thanks to the compiler inlining and optimization.

Note that try.ToX function names end to a number (x) because:

	"No variadic type parameters. There is no support for variadic type parameters,
	which would permit writing a single generic function that takes different
	numbers of both type parameters and regular parameters." - Go Generics

The leading number at the end of the To2 tells that To2 takes two different
non-error arguments, and the third one must be an error value.

Looking at the FileCopy example again, you see that all the functions
are directed to try.To1 are returning (type1, error) tuples. All of these
tuples are the correct input to try.To1. However, if you have a function that
returns (type1, type2, error), you must use try.To2 function to check the error.
Currently the try.To3 takes (3 + 1) return values which is the greatest amount.
If more is needed, let us know.

# try.Out — Error Handling Language

The try package offers an error handling DSL. It's for cases where you want to
do something specific after error returing function call. For example, you might
want to ignore the specific error and use a default value. That's possible with
the following code:

	number := try.Out1(strconv.Atoi(str)).Def1(100).Val1

Or you might want to ignore an error but write a log if something happens:

	try.Out(os.Remove(dst)).Logf("file cleanup fail")

Or you might just want to change it later to error return:

	try.Out(os.Remove(dst)).Handle("file cleanup fail")

Please see the documentation and examples of ResultX types and their methods.
*/
package try

import (
	"errors"
	"io"

	"github.com/lainio/err2"
)

// To is a helper function to call functions which returns (error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To(err error) {
	if err != nil {
		panic(err)
	}
}

// To1 is a helper function to call functions which returns (T, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To1[T any](v T, err error) T {
	To(err)
	return v
}

// To2 is a helper function to call functions which returns (T, U, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	To(err)
	return v1, v2
}

// To3 is a helper function to call functions which returns (T, U, V, error)
// and check the error value. If an error occurs, it panics the error where err2
// handlers can catch it if needed.
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	To(err)
	return v1, v2, v3
}

// Is function performs a filtered error check for the given argument. It's the
// same as To function, but it checks if the error matches the filter before
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

// IsEOF1 function performs a filtered error check for the given argument. It's the
// same as To function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the 'io.EOF'.
func IsEOF1[T any](v T, err error) (bool, T) {
	isFilter := Is(err, io.EOF)
	return isFilter, v
}

// IsEOF2 function performs a filtered error check for the given argument. It's the
// same as To function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the 'io.EOF'.
func IsEOF2[T, U any](v1 T, v2 U, err error) (bool, T, U) {
	isFilter := Is(err, io.EOF)
	return isFilter, v1, v2
}

// IsEOF function performs a filtered error check for the given argument. It's the
// same as To function, but it checks if the error matches the 'io.EOF' before
// throwing an error. The false return value tells that there are no errors.
// The true tells that the err's chain includes 'io.EOF'.
func IsEOF(err error) bool {
	return Is(err, io.EOF)
}

// IsNotFound function performs a filtered error check for the given argument.
// It's the same as To function, but it checks if the error matches the
// 'err2.NotFound' before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// 'err2.NotFound'.
func IsNotFound(err error) bool {
	return Is(err, err2.ErrNotFound)
}

// IsNotFound1 function performs a filtered error check for the given argument.
// It's the same as To function, but it checks if the error matches the
// 'err2.NotFound' before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// 'err2.NotFound'.
func IsNotFound1[T any](v T, err error) (bool, T) {
	isFilter := Is(err, err2.ErrNotFound)
	return isFilter, v
}

// IsNotExist function performs a filtered error check for the given argument.
// It's the same as To function, but it checks if the error matches the
// 'err2.NotExist' before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// 'err2.NotExist'.
func IsNotExist(err error) bool {
	return Is(err, err2.ErrNotExist)
}

// IsExist function performs a filtered error check for the given argument. It's
// the same as To function, but it checks if the error matches the 'err2.Exist'
// before throwing an error. The false return value tells that there are no
// errors. The true tells that the err's chain includes 'err2.Exist'.
func IsAlreadyExist(err error) bool {
	return Is(err, err2.ErrAlreadyExist)
}

// IsNotAccess function performs a filtered error check for the given argument.
// It's the same as To function, but it checks if the error matches the
// 'err2.NotAccess' before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// 'err2.NotAccess'.
func IsNotAccess(err error) bool {
	return Is(err, err2.ErrNotAccess)
}

// IsRecoverable function performs a filtered error check for the given
// argument. It's the same as To function, but it checks if the error matches
// the 'err2.ErrRecoverable' before throwing an error. The false return value
// tells that there are no errors. The true tells that the err's chain includes
// 'err2.ErrRecoverable'.
func IsRecoverable(err error) bool {
	return Is(err, err2.ErrRecoverable)
}

// IsNotRecoverable function performs a filtered error check for the given
// argument. It's the same as To function, but it checks if the error matches
// the 'err2.ErrNotRecoverable' before throwing an error. The false return value
// tells that there are no errors. The true tells that the err's chain includes
// 'err2.ErrNotRecoverable'.
func IsNotRecoverable(err error) bool {
	return Is(err, err2.ErrNotRecoverable)
}

// InNotEnabled function performs a filtered error check for the given argument.
// It's the same as To function, but it checks if the error matches the
// 'err2.ErrNotEnabled' before throwing an error. The false return value tells
// that there are no errors. The true tells that the err's chain includes
// 'err2.ErrNotEnabled'.
func InNotEnabled(err error) bool {
	return Is(err, err2.ErrNotEnabled)
}
