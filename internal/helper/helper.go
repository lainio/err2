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
