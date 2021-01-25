package assert

import (
	"errors"
	"fmt"
	"reflect"
)

var ProductionMode = false

// NoImplementation always fails with no implementation
func NoImplementation(a ...interface{}) {
	reportAssertionFault("not implemented", a...)
}

// True asserts that term is true. If not it panics with the given formatting
// string.
func True(term bool, a ...interface{}) {
	if !term {
		reportAssertionFault("assertion fault", a...)
	}
}

// NotNil asserts that object isn't nil. If not it panics with the given
// formatting string.
func NotNil(obj interface{}, a ...interface{}) {
	if obj != nil {
		reportAssertionFault("nil detected", a...)
	}
}

// Len asserts that length of the object is equal. If not it panics with the
// given formatting string.
func Len(obj interface{}, length int, a ...interface{}) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l != length {
		defMsg := fmt.Sprintf("got %d, want %d", l, length)
		reportAssertionFault(defMsg, a...)
	}
}

// Empty asserts that length of the object is zero. If not it panics with the
// given formatting string.
func Empty(obj interface{}, msg ...interface{}) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l != 0 {
		defMsg := fmt.Sprintf("got %d, want == 0", l)
		reportAssertionFault(defMsg, msg...)
	}
}

// NotEmpty asserts that length of the object greater than zero. If not it
// panics with the given formatting string.
func NotEmpty(obj interface{}, msg ...interface{}) {
	ok, l := getLen(obj)
	if !ok {
		panic("cannot get length")
	}

	if l == 0 {
		defMsg := fmt.Sprintf("got %d, want > 0", l)
		reportAssertionFault(defMsg, msg...)
	}
}

func reportAssertionFault(defaultMsg string, a ...interface{}) {
	if len(a) > 0 {
		if format, ok := a[0].(string); ok {
			reportPanic(fmt.Sprintf(format, a[1:]...))
		} else {
			reportPanic(fmt.Sprintln(a...))
		}
	} else {
		reportPanic(defaultMsg)
	}
}

func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}

func reportPanic(s string) {
	if ProductionMode {
		panic(errors.New(s))
	}
	panic(s)
}
