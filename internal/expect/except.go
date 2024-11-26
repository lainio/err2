package expect

import (
	"fmt"
	"testing"
)

// ThatNot fails the test if the condition is true.
func ThatNot(tb testing.TB, condition bool, v ...interface{}) {
	tb.Helper()
	if condition {
		tb.Fatal(v...)
	}
}

// That fails the test if the condition is false.
func That(tb testing.TB, condition bool, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatal(v...)
	}
}

// Thatf fails the test if the condition is false.
func Thatf(tb testing.TB, condition bool, format string, v ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf(format, v...)
	}
}

// Equal fails the test if the values aren't equal.
func Equal[T comparable](tb testing.TB, val, want T, a ...any) {
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

// NotEqual fails the test if the values aren't equal.
func NotEqual[T comparable](tb testing.TB, val, want T, a ...any) {
	tb.Helper()
	if want == val {
		defMsg := fmt.Sprintf("got '%v', want != '%v' ", val, want)
		if len(a) == 0 {
			tb.Fatal(defMsg)
		}
		format, ok := a[0].(string)
		if ok {
			tb.Fatalf(defMsg+format, a[1:]...)
		}
	}
}
