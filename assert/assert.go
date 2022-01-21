package assert

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
)

// Asserter is type for asserter object guided by its flags.
type Asserter uint32

const (
	// AsserterToError is Asserter flag to guide asserter to use Go's error
	// type for panics.
	AsserterToError Asserter = 1 << iota

	// AsserterStackTrace is Asserter flag to print call stack to stdout.
	AsserterStackTrace
)

var (
	// P is a production Asserter that types panic objects to errors which
	// allows err2 handlers to catch them.
	P = AsserterToError

	// D is a development Asserter that types panic objects to strings that
	// doesn't by caught by err2 handlers.
	D Asserter = 0
)

// NoImplementation always fails with no implementation.
func (asserter Asserter) NoImplementation(a ...any) {
	asserter.reportAssertionFault("not implemented", a...)
}

// True asserts that term is true. If not it panics with the given formatting
// string. Note! This and Truef are the most performant of all the assertion
// functions.
func (asserter Asserter) True(term bool, a ...any) {
	if !term {
		asserter.reportAssertionFault("assertion fault", a...)
	}
}

// Truef asserts that term is true. If not it panics with the given formatting
// string.
func (asserter Asserter) Truef(term bool, format string, a ...any) {
	if !term {
		if asserter.HasStackTrace() {
			debug.PrintStack()
		}
		asserter.reportPanic(fmt.Sprintf(format, a...))
	}
}

// Len asserts that length of the object is equal to given. If not it
// panics/errors (current Asserter) with the given msg. Note! This is very slow
// (before we have generics). If you need performance use EqualInt. It's not so
// convenient, though.
func (asserter Asserter) Len(obj any, length int, a ...any) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l != length {
		defMsg := fmt.Sprintf("got %d, want %d", l, length)
		asserter.reportAssertionFault(defMsg, a...)
	}
}

// EqualInt asserts that integers are equal. If not it panics/errors (current
// Asserter) with the given msg.
func (asserter Asserter) EqualInt(val, want int, a ...any) {
	if want != val {
		defMsg := fmt.Sprintf("got %d, want %d", val, want)
		asserter.reportAssertionFault(defMsg, a...)
	}
}

// Lenf asserts that length of the object is equal to given. If not it
// panics/errors (current Asserter) with the given msg. Note! This is very slow
// (before we have generics). If you need performance use EqualInt. It's not so
// convenient, though.
func (asserter Asserter) Lenf(obj any, length int, format string, a ...any) {
	args := combineArgs(format, a)
	asserter.Len(obj, length, args...)
}

// Empty asserts that length of the object is zero. If not it panics with the
// given formatting string. Note! This is slow.
func (asserter Asserter) Empty(obj any, msg ...any) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l != 0 {
		defMsg := fmt.Sprintf("got %d, want == 0", l)
		asserter.reportAssertionFault(defMsg, msg...)
	}
}

// NotEmptyf asserts that length of the object greater than zero. If not it
// panics with the given formatting string. Note! This is slow.
func (asserter Asserter) NotEmptyf(obj any, format string, msg ...any) {
	args := combineArgs(format, msg)
	asserter.Empty(obj, args...)
}

// NotEmpty asserts that length of the object greater than zero. If not it
// panics with the given formatting string. Note! This is slow.
func (asserter Asserter) NotEmpty(obj any, msg ...any) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l == 0 {
		defMsg := fmt.Sprintf("got %d, want > 0", l)
		asserter.reportAssertionFault(defMsg, msg...)
	}
}

func (asserter Asserter) reportAssertionFault(defaultMsg string, a ...any) {
	if asserter.HasStackTrace() {
		debug.PrintStack()
	}
	if len(a) > 0 {
		if format, ok := a[0].(string); ok {
			asserter.reportPanic(fmt.Sprintf(format, a[1:]...))
		} else {
			asserter.reportPanic(fmt.Sprintln(a...))
		}
	} else {
		asserter.reportPanic(defaultMsg)
	}
}

func getLen(x any) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}

func (asserter Asserter) reportPanic(s string) {
	if asserter.HasToError() {
		panic(errors.New(s))
	}
	panic(s)
}

func (asserter Asserter) HasToError() bool {
	return asserter&AsserterToError != 0
}

func (asserter Asserter) HasStackTrace() bool {
	return asserter&AsserterStackTrace != 0
}

func combineArgs(format string, a []any) []any {
	args := make([]any, 1, len(a)+1)
	args[0] = format
	args = append(args, a...)
	return args
}
