package assert

import (
	"errors"
	"fmt"
	"os"

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

// reportAssertionFault reports assertion fault according the current asserter
// and its config. If extra argumnets are given (a ...any) and the first is
// string, it's treated as format string and following args as its parameters.
//
// Note. We use the pattern where we build defaultMsg argument reaady in cases
// like 'got: X, want: Y'. This hits two birds with one stone: we have automatic
// and correct assert messages, and we can add information to it if we want to.
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
		// we concat given format string to the in cases we
		// have got: want: pattern in use, i.e. best from both worlds
		if format, ok := a[0].(string); ok {
			asserter.reportPanic(fmt.Sprintf(format, a[1:]...))
		} else {
			asserter.reportPanic(fmt.Sprintln(a...))
		}
	} else {
		asserter.reportPanic(defaultMsg)
	}
}

func (asserter Asserter) reportPanic(s string) {
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
