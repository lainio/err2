package assert

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/lainio/err2/internal/debug"
)

// Asserter is type for asserter object guided by its flags.
type Asserter uint32

const (
	// AsserterDebug is the default mode where all asserts are treaded as
	// panics
	AsserterDebug Asserter = 0

	// AsserterToError is Asserter flag to guide asserter to use Go's error
	// type for panics.
	AsserterToError Asserter = 1 << iota

	// AsserterStackTrace is Asserter flag to print call stack to stdout.
	AsserterStackTrace

	// AsserterCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name.
	AsserterCallerInfo

	// AsserterFormattedCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name in
	// multi-line formatted string output.
	AsserterFormattedCallerInfo
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
		if asserter.hasStackTrace() {
			debug.PrintStack(1)
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
	if asserter.hasStackTrace() {
		debug.PrintStack(2)
	}
	if asserter.hasCallerInfo() {
		defaultMsg = asserter.callerInfo(defaultMsg)
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
	if asserter.hasToError() {
		panic(errors.New(s))
	}
	panic(s)
}

var longFmtStr = `
--------------------------------
Assertion Fault at:
%s:%d %s
%s
--------------------------------
`

var shortFmtStr = `%s:%d %s %s`

func (asserter Asserter) callerInfo(msg string) (info string) {
	const stackLevel = 3
	pc, file, line, ok := runtime.Caller(stackLevel)
	if !ok {
		return msg
	}

	ourFmtStr := shortFmtStr
	if asserter.hasFormattedCallerInfo() {
		ourFmtStr = longFmtStr
	}

	fn := runtime.FuncForPC(pc)
	filename := filepath.Base(file)
	ext := filepath.Ext(filename)
	trimmedFilename := strings.TrimSuffix(filename, ext) + "."
	funcName := strings.TrimPrefix(filepath.Base(fn.Name()), trimmedFilename)
	info = fmt.Sprintf(ourFmtStr,
		filename, line,
		funcName, msg)

	return
}

func (asserter Asserter) hasToError() bool {
	return asserter&AsserterToError != 0
}

func (asserter Asserter) hasStackTrace() bool {
	return asserter&AsserterStackTrace != 0
}

func (asserter Asserter) hasCallerInfo() bool {
	return asserter&AsserterCallerInfo != 0 || asserter.hasFormattedCallerInfo()
}

func (asserter Asserter) hasFormattedCallerInfo() bool {
	return asserter&AsserterFormattedCallerInfo != 0
}
