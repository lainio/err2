package assert

import (
	"errors"
	"fmt"
	"os"

	"github.com/lainio/err2/internal/debug"
	"github.com/lainio/err2/internal/str"
	"github.com/lainio/err2/internal/x"
)

// asserter is type for asserter object guided by its flags.
type asserter uint32

const (
	// asserterDebug is the default mode where all asserts are treaded as
	// panics
	asserterDebug asserter = 0

	// asserterToError is Asserter flag to guide asserter to use Go's error
	// type for panics.
	asserterToError asserter = 1 << iota

	// asserterStackTrace is Asserter flag to print call stack to stdout OR if
	// in AsserterUnitTesting mode the call stack is printed to test result
	// output if there is any assertion failures.
	asserterStackTrace

	// asserterCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name.
	// This is especially powerful with AsserterUnitTesting where it allows get
	// information where the assertion violation happens even over modules!
	asserterCallerInfo

	// asserterFormattedCallerInfo is an asserter flag to add info of the function
	// asserting. It includes filename, line number and function name in
	// multi-line formatted string output.
	asserterFormattedCallerInfo

	// asserterUnitTesting is an asserter only for unit testing. It can be
	// compined with AsserterCallerInfo and/or AsserterStackTrace. There is
	// variable T which have all of these three asserters.
	asserterUnitTesting
)

// every test log or result output has 4 spaces in them
const officialTestOutputPrefix = "    "

// reportAssertionFault reports assertion fault according the current asserter
// and its config. If extra argumnets are given (a ...any) and the first is
// string, it's treated as format string and following args as its parameters.
//
// Note. We use the pattern where we build defaultMsg argument reaady in cases
// like 'got: X, want: Y'. This hits two birds with one stone: we have automatic
// and correct assert messages, and we can add information to it if we want to.
// If asserter is Plain (isErrorOnly()) user wants to override automatic assert
// messgages with our given, usually simple message.
func (asserter asserter) reportAssertionFault(
	extraInd int,
	defaultMsg string,
	a []any,
) {
	if asserter.hasStackTrace() {
		if asserter.isUnitTesting() {
			// Note. that the assert in the test function is printed in
			// reportPanic below
			const StackLvl = 5 // amount of functions before we're here
			stackLvl := StackLvl + extraInd
			debug.PrintStackForTest(os.Stderr, stackLvl)
		} else {
			// amount of functions before we're here, which is different
			// between runtime (this) and test-run (above)
			const StackLvl = 2
			stackLvl := StackLvl + extraInd
			debug.PrintStack(stackLvl)
		}
	}
	if asserter.hasCallerInfo() {
		defaultMsg = asserter.callerInfo(defaultMsg, extraInd)
	}
	if len(a) > 0 {
		if format, ok := a[0].(string); ok {
			allowDefMsg := !asserter.isErrorOnly() && defaultMsg != ""
			f := x.Whom(allowDefMsg, defaultMsg+conCatErrStr+format, format)
			asserter.reportPanic(fmt.Sprintf(f, a[1:]...))
		} else {
			asserter.reportPanic(fmt.Sprintln(append([]any{defaultMsg}, a...)))
		}
	} else {
		asserter.reportPanic(defaultMsg)
	}
}

func (asserter asserter) reportPanic(s string) {
	if asserter.isUnitTesting() && asserter.hasCallerInfo() {
		fmt.Fprintln(os.Stderr, officialTestOutputPrefix+s)
		tester().FailNow()
	} else if asserter.isUnitTesting() {
		const framesToSkip = 4 // how many fn calls there is before FuncName call
		fatal(s, framesToSkip)
	}
	if asserter.hasToError() {
		panic(errors.New(s))
	}
	panic(s)
}

// fatal calls tester().FailNow() after printing test failing msg (arg s) that's
// build according the correct test result output. framesToSkip tells how many
// functions (stack call frames) to skip until get the function name to print.
func fatal(s string, framesToSkip int) {
	const shortFmtStr = `%s:%d: %s`
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

func (asserter asserter) callerInfo(msg string, extraInd int) (info string) {
	ourFmtStr := shortFmtStr
	if asserter.hasFormattedCallerInfo() {
		ourFmtStr = longFmtStr
	}

	const ToSkip = 3
	framesToSkip := ToSkip + extraInd // how many fn calls there is before FuncName call
	includePath := asserter.isUnitTesting()
	funcName, filename, line, ok := str.FuncName(framesToSkip, includePath)
	if ok {
		info = fmt.Sprintf(ourFmtStr,
			filename, line,
			funcName, msg)
	}

	return
}

func (asserter asserter) isErrorOnly() bool {
	return asserter == asserterToError
}

func (asserter asserter) hasToError() bool {
	return asserter&asserterToError != 0
}

func (asserter asserter) hasStackTrace() bool {
	return asserter&asserterStackTrace != 0
}

func (asserter asserter) hasCallerInfo() bool {
	return asserter&asserterCallerInfo != 0 ||
		asserter.hasFormattedCallerInfo()
}

func (asserter asserter) hasFormattedCallerInfo() bool {
	return asserter&asserterFormattedCallerInfo != 0
}

// isUnitTesting is expensive because it calls tester(). think carefully where
// to use it
func (asserter asserter) isUnitTesting() bool {
	return asserter&asserterUnitTesting != 0 && tester() != nil
}
