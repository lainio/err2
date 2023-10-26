package test

import (
	"fmt"
	"testing"
)

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
