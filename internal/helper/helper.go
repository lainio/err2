package helper

import "testing"

// require fails the test if the condition is false.
func Require(tb testing.TB, condition bool, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatal(v...)
	}
}

// require fails the test if the condition is false.
func Requiref(tb testing.TB, condition bool, format string, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf(format, v...)
	}
}

// Whom is exactly same as C/C++ ternary operator. In Go it's implemented with
// generics.
func Whom[T any](b bool, yes, no T) T {
	if b {
		return yes
	}
	return no
}
