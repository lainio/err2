package assert

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/lainio/err2/internal/debug"
	"github.com/lainio/err2/internal/str"
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

	// AsserterStackTrace is Asserter flag to print call stack to stdout OR if
	// in AsserterUnitTesting mode the call stack is printed to test result
	// output if there is any assertion failures.
	AsserterStackTrace

	// AsserterCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name.
	// This is especially powerful with AsserterUnitTesting where it allows get
	// information where the assertion violation happens even over modules!
	AsserterCallerInfo

	// AsserterFormattedCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name in
	// multi-line formatted string output.
	AsserterFormattedCallerInfo

	// AsserterUnitTesting is an asserter only for unit testing. It can be
	// compined with AsserterCallerInfo and/or AsserterStackTrace. There is
	// variable T which have all of these three asserters.
	AsserterUnitTesting
)

// every test log or result output has 4 spaces in them
const officialTestOutputPrefix = "    "

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
		if asserter.isUnitTesting() {
			// Note. that the assert in the test function is printed in
			// reportPanic below
			const stackLvl = 5 // amount of functions before we're here
			debug.PrintStackForTest(os.Stderr, stackLvl)
		} else {
			// amount of functions before we're here, which is different
			// between runtime (this) and test-run (above)
			const stackLvl = 2
			debug.PrintStack(stackLvl)
		}
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
	if asserter.isUnitTesting() && asserter.hasCallerInfo() {
		fmt.Fprintln(os.Stderr, officialTestOutputPrefix+s)
		tester().FailNow()
	} else if asserter.isUnitTesting() {
		fatal(s)
	}
	if asserter.hasToError() {
		panic(errors.New(s))
	}
	panic(s)
}

func fatal(s string) {
	const shortFmtStr = `%s:%d: %s`
	const framesToSkip = 4 // how many fn calls there is before FuncName call
	includePath := false
	_, filename, line, ok := str.FuncName(framesToSkip, includePath)
	info := s
	if ok {
		info = fmt.Sprintf(shortFmtStr, filename, line, s)
	}
	// test output goes thru stderr, no need for t.Log(), test Fail needs it.
	fmt.Fprintln(os.Stderr, officialTestOutputPrefix+info)
	tester().FailNow()
}

var longFmtStr = `
--------------------------------
Assertion Fault at:
%s:%d %s():
%s
--------------------------------
`

var shortFmtStr = `%s:%d: %s(): %s`

func (asserter Asserter) callerInfo(msg string) (info string) {
	ourFmtStr := shortFmtStr
	if asserter.hasFormattedCallerInfo() {
		ourFmtStr = longFmtStr
	}

	const framesToSkip = 3 // how many fn calls there is before FuncName call
	includePath := asserter.isUnitTesting()
	funcName, filename, line, ok := str.FuncName(framesToSkip, includePath)
	if ok {
		info = fmt.Sprintf(ourFmtStr,
			filename, line,
			funcName, msg)
	}

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

// isUnitTesting is expensive because it calls tester(). think carefully where
// to use it
func (asserter Asserter) isUnitTesting() bool {
	return asserter&AsserterUnitTesting != 0 && tester() != nil
}
