// package test should be renamed to require. Maybe we could try 1st gopls and
// then idea? TODO:
package test

import (
	"fmt"
	"testing"
)

// TODO: why we have 2 Require functions? If we'll use type assertion for the
// first argument and decide that if it's a string and maybe search % marks we
// could have only one function? Or we count v count: if 1 no format, if >1
// format.

// Require fails the test if the condition is false.
func Require(tb testing.TB, condition bool, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatal(v...)
	}
}

// Requiref fails the test if the condition is false.
func Requiref(tb testing.TB, condition bool, format string, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf(format, v...)
	}
}

// RequireEqual fails the test if the values aren't equal
func RequireEqual[T comparable](tb testing.TB, val, want T, a ...any) {
	tb.Helper()
	if want != val {
		defMsg := fmt.Sprintf("got '%v', want '%v' ", val, want)
		if len(a) == 0 {
			tb.Fatal(defMsg)
		}
		format, ok := a[0].(string)
		if ok {
			tb.Fatalf(defMsg+format, a[1:]...)
		}
	}
}
