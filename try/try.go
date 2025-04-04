/*
Package try is a package for error checking with the functions: [To], [To1], and
[To2]. These functions check the last given error argument isn't nil, i.e., 'if
err != nil' and if it is, it throws the err to the error handlers. More
information about err2 and try packages roles can be seen in the FileCopy
example:

	...
	r := try.To1(os.Open(src))
	defer r.Close()

	w := try.To1(os.Create(dst))
	defer err2.Handle(&err, func(error) error {
	     try.Out(os.Remove(dst)).Logf()
	     return nil
	})
	defer w.Close()
	try.To1(io.Copy(w, r))
	return nil
	...

# try.To — Fast Checking

All of the [To] functions are as fast as the simple 'if err != nil {'
statement, thanks to the compiler inlining and optimization.

We have three error check functions: [To], [To1], and [To2] because:

	"No variadic type parameters. There is no support for variadic type parameters,
	which would permit writing a single generic function that takes different
	numbers of both type parameters and regular parameters." - Go Generics

For example, the leading number at the end of the [To2] tells that [To2] takes
two different non-error arguments, and the third one must be an error value.

Looking at the [CopyFile] example again, you see that all the functions
are directed to [To1] are returning (type1, error) tuples. All of these
tuples are the correct input to [To1]. However, if you have a function that
returns (type1, type2, error), you must use [To2] function to check the error.
Currently the [To3] takes (3 + 1) return values which is the greatest amount.
If more is needed, let us know.

# try.Out — Error Handling Language

The try package offers an error handling DSL that's based on [Out], [Out1], and
[Out2] functions and their corresponding return values [Result], [Result1], and
[Result2]. DSL is for the cases where you want to do something specific after
error returning function call. Those cases are rare. But you might want, for
example, to ignore the specific error and use a default value without any
special error handling. That's possible with the following code:

	number := try.Out1(strconv.Atoi(str)).Catch(100)

Or you might want to ignore an error but write a log if something happens:

	try.Out(os.Remove(dst)).Logf("file cleanup fail")

Or you might just want to change it later to error return:

	try.Out(os.Remove(dst)).Handle("file cleanup fail")

Please see the documentation and examples of [Result], [Result1], and [Result2]
types and their methods.

# try.T — Checking and Annotation

The try package offers functions [T], [T1], [T2], and [T3] to allow fast
incremental code refactoring. For example, if you want to add an error check
specific annotation to the error check already done with [To1]:

	try.To1(io.Copy(w, r))

you can easily change it to [T1] and give extra message added to the error:

	try.T1(io.Copy(w, r))("error during stream copy")

The T functions are offered mainly to allow faste feedback loop to play with the
error messages and see what works the best.
*/
package try

import (
	"errors"
	"fmt"
	"io"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/handler"
)

// To is a helper function to call functions which returns an error value and
// check the value. If an error occurs, it panics the error so that err2
// handlers can catch it if needed. Note! If no err2.Handle or err2.Catch exist
// in the call stack and To panics an error, the error is not handled, and the
// app will crash. When using To function you should always have proper
// err2.Handle or err2.Catch statements in the call stack.
//
//	defer err2.Handle(&err)
//	...
//	try.To(w.Close())
func To(err error) {
	if err != nil {
		panic(err)
	}
}

// To1 is a helper function to call functions which returns values (T, error)
// and check the error value. If an error occurs, it panics the error so that
// err2 handlers can catch it if needed. Note! If no err2.Handle or err2.Catch
// exist in the call stack and To1 panics an error, the error is not handled,
// and the app will crash. When using To1 function you should always have
// proper err2.Handle or err2.Catch statements in the call stack.
//
//	defer err2.Handle(&err)
//	...
//	r := try.To1(os.Open(src))
func To1[T any](v T, err error) T {
	To(err)
	return v
}

// To2 is a helper function to call functions which returns values (T, U, error)
// and check the error value. If an error occurs, it panics the error so that
// err2 handlers can catch it if needed. Note! If no err2.Handle or err2.Catch
// exist in the call stack and To2 panics an error, the error is not handled,
// and the app will crash. When using To2 function you should always have
// proper err2.Handle or err2.Catch statements in the call stack.
//
//	defer err2.Handle(&err)
//	...
//	kid, pk := try.To2(keys.CreateAndExportPubKeyBytes(kms.ED25519))
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	To(err)
	return v1, v2
}

// To3 is a helper function to call functions which returns values (T, U, V,
// error) and check the error value. If an error occurs, it panics the error so
// that err2 handlers can catch it if needed. Note! If no err2.Handle or
// err2.Catch exist in the call stack and To3 panics an error, the error is
// not handled, and the app will crash. When using To3 function you should
// always have proper err2.Handle or err2.Catch statements in the call stack.
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	To(err)
	return v1, v2, v3
}

// Is function performs a filtered error check for the given argument. It's the
// same as [To] function, but it checks if the error matches the filter before
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

// IsEOF1 function performs a filtered error check for the given argument. It's
// the same as [To] function, but it checks if the error matches the [io.EOF]
// before throwing an error. The false return value tells that there are no
// errors and the true value that the error is the [io.EOF].
func IsEOF1[T any](v T, err error) (bool, T) {
	isFilter := Is(err, io.EOF)
	return isFilter, v
}

// IsEOF2 function performs a filtered error check for the given argument. It's the
// same as [To] function, but it checks if the error matches the [io.EOF] before
// throwing an error. The false return value tells that there are no errors and
// the true value that the error is the [io.EOF].
func IsEOF2[T, U any](v1 T, v2 U, err error) (bool, T, U) {
	isFilter := Is(err, io.EOF)
	return isFilter, v1, v2
}

// IsEOF function performs a filtered error check for the given argument. It's the
// same as [To] function, but it checks if the error matches the [io.EOF] before
// throwing an error. The false return value tells that there are no errors.
// The true tells that the err's chain includes [io.EOF].
func IsEOF(err error) bool {
	return Is(err, io.EOF)
}

// IsNotFound function performs a filtered error check for the given argument.
// It's the same as [To] function, but it checks if the error matches the
// [err2.NotFound] before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// [err2.NotFound].
func IsNotFound(err error) bool {
	return Is(err, err2.ErrNotFound)
}

// IsNotFound1 function performs a filtered error check for the given argument.
// It's the same as [To] function, but it checks if the error matches the
// [err2.NotFound] before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// [err2.NotFound].
func IsNotFound1[T any](v T, err error) (bool, T) {
	isFilter := Is(err, err2.ErrNotFound)
	return isFilter, v
}

// IsNotExist function performs a filtered error check for the given argument.
// It's the same as [To] function, but it checks if the error matches the
// [err2.NotExist] before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// [err2.NotExist].
func IsNotExist(err error) bool {
	return Is(err, err2.ErrNotExist)
}

// IsExist function performs a filtered error check for the given argument. It's
// the same as [To] function, but it checks if the error matches the
// [err2.AlreadyExist] before throwing an error. The false return value tells
// that there are no errors. The true tells that the err's chain includes
// [err2.AlreadyExist].
func IsAlreadyExist(err error) bool {
	return Is(err, err2.ErrAlreadyExist)
}

// IsNotAccess function performs a filtered error check for the given argument.
// It's the same as [To] function, but it checks if the error matches the
// [err2.NotAccess] before throwing an error. The false return value tells that
// there are no errors. The true tells that the err's chain includes
// [err2.NotAccess].
func IsNotAccess(err error) bool {
	return Is(err, err2.ErrNotAccess)
}

// IsRecoverable function performs a filtered error check for the given
// argument. It's the same as [To] function, but it checks if the error matches
// the [err2.ErrRecoverable] before throwing an error. The false return value
// tells that there are no errors. The true tells that the err's chain includes
// [err2.ErrRecoverable].
func IsRecoverable(err error) bool {
	return Is(err, err2.ErrRecoverable)
}

// IsNotRecoverable function performs a filtered error check for the given
// argument. It's the same as [To] function, but it checks if the error matches
// the [err2.ErrNotRecoverable] before throwing an error. The false return value
// tells that there are no errors. The true tells that the err's chain includes
// [err2.ErrNotRecoverable].
func IsNotRecoverable(err error) bool {
	return Is(err, err2.ErrNotRecoverable)
}

// IsNotEnabled function performs a filtered error check for the given argument.
// It's the same as [To] function, but it checks if the error matches the
// [err2.ErrNotEnabled] before throwing an error. The false return value tells
// that there are no errors. The true tells that the err's chain includes
// [err2.ErrNotEnabled].
func IsNotEnabled(err error) bool {
	return Is(err, err2.ErrNotEnabled)
}

// T is similar as [To] but it let's you to annotate a possible error at place.
//
//	try.T(f.Close)("annotations")
//
// Note that T is a helper, which means that you start with it randomly. You
// start with [To] and end up using T if you really need to add context
// related to a specific error check, which is a very rare case.
func T(err error) func(fs string) {
	return func(fs string) {
		if err == nil {
			return
		}
		panic(annotateErr(err, fs))
	}
}

// T1 is similar as [To1] but it let's you to annotate a possible error at place.
//
//	f := try.T1(os.Open("filename")("cannot open cfg file")
//
// Note that T1 is a helper, which means that you start with it randomly. You
// start with [To1] and end up using T1 if you really need to add context
// related to a specific error check, which is a very rare case.
func T1[T any](v T, err error) func(fs string) T {
	return func(fs string) T {
		if err == nil {
			return v
		}
		panic(annotateErr(err, fs))
	}
}

// T2 is similar as [To2] but it let's you to annotate a possible error at place.
//
// Note that T2 is a helper, which means that you start with it randomly. You
// start with [To2] and end up using T2 if you really need to add context
// related to a specific error check, which is a very rare case.
func T2[T, U any](v T, u U, err error) func(fs string) (T, U) {
	return func(fs string) (T, U) {
		if err == nil {
			return v, u
		}
		panic(annotateErr(err, fs))
	}

}

func annotateErr(err error, fs string) error {
	return fmt.Errorf(fs+handler.WrapError, err)
}

// T3 is similar as [To3] but it let's you to annotate a possible error at place.
//
// Note that T3 is a helper, which means that you start with it randomly. You
// start with [To3] and end up using T3 if you really need to add context
// related to a specific error check, which is a very rare case.
func T3[T, U, V any](v1 T, v2 U, v3 V, err error) func(fs string) (T, U, V) {
	return func(fs string) (T, U, V) {
		if err == nil {
			return v1, v2, v3
		}
		panic(annotateErr(err, fs))
	}
}
