package assert

import (
	"fmt"
)

var (
	// P is a production Asserter that types panic objects to errors which
	// allows err2 handlers to catch them.
	P = AsserterToError

	// D is a development Asserter that types panic objects to strings that
	// doesn't by caught by err2 handlers.
	D Asserter = 0

	// DefaultAsserter is a default asserter used for package level functions.
	// If not changed it is the same as P the production asserter that treats
	// assert failures as Go errors, i.e. if err2 handlers are found in the
	// callstack these errors are caught.
	DefaultAsserter = AsserterToError
)

// That asserts that term is true. If not it panics with the given formatting
// string. Note! That is the most performant of all the assertion functions.
func That(term bool, a ...any) {
	if !term {
		DefaultAsserter.reportAssertionFault("assertion fault", a...)
	}
}

// NotNil asserts that value in not nil. If it is it panics/errors (default
// Asserter) with the given msg.
func NotNil[T any](p *T, a ...any) {
	if p == nil {
		defMsg := "pointer is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SNotNil asserts that value in not nil. If it is it panics/errors (default
// Asserter) with the given msg.
func SNotNil[T any](s []T, a ...any) {
	if s == nil {
		defMsg := "slice is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// CNotNil asserts that value in not nil. If it is it panics/errors (default
// Asserter) with the given msg.
func CNotNil[T any](c chan T, a ...any) {
	if c == nil {
		defMsg := "channel is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// MNotNil asserts that value in not nil. If it is it panics/errors (default
// Asserter) with the given msg.
func MNotNil[T comparable, U any](m map[T]U, a ...any) {
	if m == nil {
		defMsg := "map is nil"
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// NotEqual asserts that values are equal. If not it panics/errors (current
// Asserter) with the given msg.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		defMsg := fmt.Sprintf("got %v, want %v", val, want)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// Equal asserts that values are equal. If not it panics/errors (current
// Asserter) with the given msg.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		defMsg := fmt.Sprintf("got %v, want %v", val, want)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// SLen asserts that length of the object is equal to given. If not it
// panics/errors (current Asserter) with the given msg. Note! This is very slow
// (before we have generics). If you need performance use EqualInt. It's not so
// convenient, though.
func SLen[T any](obj []T, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf("got %d, want %d", l, length)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

// MLen asserts that length of the object is equal to given. If not it
// panics/errors (current Asserter) with the given msg. Note! This is very slow
// (before we have generics). If you need performance use EqualInt. It's not so
// convenient, though.
func MLen[T comparable, U any](obj map[T]U, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf("got %d, want %d", l, length)
		DefaultAsserter.reportAssertionFault(defMsg, a...)
	}
}

func combineArgs(format string, a []any) []any {
	args := make([]any, 1, len(a)+1)
	args[0] = format
	args = append(args, a...)
	return args
}
