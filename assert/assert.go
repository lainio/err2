package assert

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/lainio/err2/internal/debug"
	"github.com/lainio/err2/internal/x"
	"golang.org/x/exp/constraints"
)

type defInd = uint32

const (
	// Plain converts asserts just plain K&D error messages without extra
	// information.
	Plain defInd = 0 + iota

	// Production (pkg default) is the best asserter for most uses. The
	// assertion violations are treated as Go error values. And only a
	// pragmatic caller info is automatically included into the error message
	// like file name, line number, and caller function, all in one line:
	//
	//  copy file: main.go:37: CopyFile(): assertion violation: string shouldn't be empty
	Production

	// Development is the best asserter for most development uses. The
	// assertion violations are treated as Go error values. And a formatted
	// caller info is automatically included to the error message like file
	// name, line number, and caller function. Everything in a beautiful
	// multi-line message:
	//
	//  --------------------------------
	//  Assertion Fault at:
	//  main.go:37 CopyFile():
	//  assertion violation: string shouldn't be empty
	//  --------------------------------
	Development

	// Test minimalistic asserter for unit test use. More pragmatic is the
	// TestFull asserter (test default).
	//
	// Use this asserter if your IDE/editor doesn't support full file names and
	// it relies a relative path (Go standard). You can use this also if you
	// need temporary problem solving for your programming environment.
	Test

	// TestFull asserter (test default). The TestFull asserter includes the
	// caller info and the call stack for unit testing, similarly like err2's
	// error traces.
	//
	// The call stack produced by the test asserts can be used over Go module
	// boundaries. For example, if your app and it's sub packages both use
	// err2/assert for unit testing and runtime checks, the runtime assertions
	// will be automatically converted to test asserts. If any of the runtime
	// asserts of the sub packages fails during the app tests, the app test
	// fails as well.
	//
	// Note, that the cross-module assertions produce long file names (path
	// included), and some of the Go test result parsers cannot handle that.
	// A proper test result parser like 'github.com/lainio/nvim-go' (fork)
	// works very well. Also most of the make result parsers can process the
	// output properly and allow traverse of locations of the error trace.
	TestFull

	// Debug asserter transforms assertion violations to panic calls where
	// panic object's type is string, i.e., err2 package treats it as a normal
	// panic, not an error.
	//
	// The pattern that e.g. Go's standard
	// library uses:
	//
	//   if p == nil {
	//        panic("pkg: ptr cannot be nil")
	//   }
	//
	// is equal to:
	//
	//   assert.NotNil(p)
	Debug
)

type flagAsserter struct{}

// Deprecated: use e.g. assert.That(), only default asserter is used.
var (
	PL = AsserterToError

	// P is a production Asserter that sets panic objects to errors which
	// allows err2 handlers to catch them.
	P = AsserterToError | AsserterCallerInfo

	B = AsserterToError | AsserterFormattedCallerInfo

	T  = AsserterUnitTesting
	TF = AsserterUnitTesting | AsserterStackTrace | AsserterCallerInfo

	// D is a development Asserter that sets panic objects to strings that
	// doesn't by caught by err2 handlers.
	// Deprecated: use e.g. assert.That(), only default asserter is used.
	D = AsserterDebug
)

var (
	// These two are our indexing system for default asserter. Note, also the
	// mutex below. All of this is done to keep client package race detector
	// cool.
	//
	//  Plain
	//  Production
	//  Development
	//  Test
	//  TestFull
	//  Debug
	defAsserter = []Asserter{PL, P, B, T, TF, D}
	def         defInd

	// mu is package lvl Mutex that is used to cool down race detector of
	// client pkgs, i.e. packages that use us can use -race flag in their test
	// runs where they change asserter. With the mutex we can at least allow
	// the setters run at one of the time. AND when that's combined with the
	// indexing system we are using for default asserter (above) we are pretty
	// much theard safe.
	mu sync.Mutex

	asserterFlag flagAsserter
)

func init() {
	SetDefault(Production)
	flag.Var(&asserterFlag, "asserter", "`asserter`: Plain, Prod, Dev, Debug")
}

type (
	testersMap = map[int]testing.TB
	function   = func()
)

var (
	// testers must be set if assertion package is used for the unit testing.
	testers = x.NewRWMap[testersMap]()
)

const (
	assertionMsg      = "assertion violation"
	gotWantFmt        = ": got '%v', want '%v'"
	gotWantLongerFmt  = ": got '%v', should be longer than '%v'"
	gotWantShorterFmt = ": got '%v', should be shorter than '%v'"

	conCatErrStr = ": "
)

// PushTester sets the current testing context for default asserter. This must
// be called at the beginning of every test. There is two way of doing it:
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) { // Shorter way, litle magic
//			defer assert.PushTester(t)() // <- IMPORTANT! NOTE! (t)()
//			...
//			assert.That(something, "test won't work")
//		})
//		t.Run(tt.name, func(t *testing.T) { // Longer, explicit way, 2 lines
//			assert.PushTester(t) // <- IMPORTANT!
//			defer assert.PopTester()
//			...
//			assert.That(something, "test won't work")
//		})
//	}
//
// Because PushTester returns PopTester it allows us to merge these two calls to
// one line. See the first t.Run call above. See more information in PopTester.
//
// PushTester allows you to change the current default asserter by accepting it
// as a second argument.
//
// Note, that the second argument, if given, changes the default asserter for
// whole package. The argument is mainly for temporary development use and isn't
// not preferred API usage.
func PushTester(t testing.TB, a ...defInd) function {
	if len(a) > 0 {
		SetDefault(a[0])
	} else if Default()&AsserterUnitTesting == 0 {
		// if this is forgotten or tests don't have proper place to set it
		// it's good to keep the API as simple as possible
		SetDefault(TestFull)
	}
	testers.Tx(func(m testersMap) {
		rid := goid()
		if _, ok := m[rid]; ok {
			panic("PushTester is already called")
		}
		m[rid] = t
	})
	return PopTester
}

// PopTester pops the testing context reference from the memory. This is for
// memory cleanup and adding similar to err2.Catch error/panic safety for tests.
// By using PopTester you get error logs tuned for unit testing.
//
// You have two ways to call PopTester. With defer right after PushTester:
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.PushTester(t) // <- important!
//			defer assert.PopTester() // <- for good girls and not so bad boys
//			...
//			assert.That(something, "test won't work")
//		})
//	}
//
// If you want to have one liner to combine Push/PopTester:
//
//	defer assert.PushTester(t)()
func PopTester() {
	defer testers.Tx(func(m testersMap) {
		goid := goid()
		delete(m, goid)
	})

	r := recover()
	if r == nil {
		return
	}

	var msg string
	switch t := r.(type) {
	case string:
		msg = t
	case runtime.Error:
		msg = t.Error()
	case error:
		msg = t.Error()
	default:
		msg = "test panic catch"
	}

	// First, print the call stack. Note, that we aren't support full error
	// tracing with unit test logging. However, using it has proved the top
	// level error stack as more enough. Even so that we could consider using
	// it for normal error stack straces if it would be possible.
	const stackLvl = 6 // amount of functions before we're here
	debug.PrintStackForTest(os.Stderr, stackLvl)

	// Now that call stack errors are printed, if any. Let's print the actual
	// line that caused the error, i.e., was throwing the error. Note, that we
	// are here in the 'catch-function'.
	const framesToSkip = 4 // how many fn calls there is before FuncName call
	fatal("assertion catching: "+msg, framesToSkip)
}

func tester() (t testing.TB) {
	return testers.Get(goid())
}

// NotImplemented always panics with 'not implemented' assertion message.
func NotImplemented(a ...any) {
	Default().reportAssertionFault("not implemented", a)
}

// ThatNot asserts that the term is NOT true. If is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func ThatNot(term bool, a ...any) {
	if term {
		defMsg := assertionMsg
		Default().reportAssertionFault(defMsg, a)
	}
}

// That asserts that the term is true. If not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func That(term bool, a ...any) {
	if !term {
		defMsg := assertionMsg
		Default().reportAssertionFault(defMsg, a)
	}
}

// NotNil asserts that the pointer IS NOT nil. If it is it panics/errors (default
// Asserter) with the given message.
func NotNil[P ~*T, T any](p P, a ...any) {
	if p == nil {
		defMsg := assertionMsg + ": pointer shouldn't be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// Nil asserts that the pointer IS nil. If it is not it panics/errors (default
// Asserter) with the given message.
func Nil[T any](p *T, a ...any) {
	if p != nil {
		defMsg := assertionMsg + ": pointer should be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// INil asserts that the interface value IS nil. If it is it panics/errors
// (default Asserter) with the given message.
//
// Note, use this only for real interface types. Go's interface's has two values
// so this won't work e.g. slices! Read more information about interface type.
//
//	https://go.dev/doc/faq#nil_error
func INil(i any, a ...any) {
	if i != nil {
		defMsg := assertionMsg + ": interface should be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// INotNil asserts that the interface value is NOT nil. If it is it
// panics/errors (default Asserter) with the given message.
//
// Note, use this only for real interface types. Go's interface's has two values
// so this won't work e.g. slices! Read more information about interface type.
//
//	https://go.dev/doc/faq#nil_error
func INotNil(i any, a ...any) {
	if i == nil {
		defMsg := assertionMsg + ": interface shouldn't be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// SNil asserts that the slice IS nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNil[S ~[]T, T any](s S, a ...any) {
	if s != nil {
		defMsg := assertionMsg + ": slice should be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// SNotNil asserts that the slice is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNotNil[S ~[]T, T any](s S, a ...any) {
	if s == nil {
		defMsg := assertionMsg + ": slice shouldn't be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// CNotNil asserts that the channel is not nil. If it is it panics/errors
// (default Asserter) with the given message.
func CNotNil[C ~chan T, T any](c C, a ...any) {
	if c == nil {
		defMsg := assertionMsg + ": channel shouldn't be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// MNotNil asserts that the map is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func MNotNil[M ~map[T]U, T comparable, U any](m M, a ...any) {
	if m == nil {
		defMsg := assertionMsg + ": map shouldn't be nil"
		Default().reportAssertionFault(defMsg, a)
	}
}

// NotEqual asserts that the values aren't equal. If they are it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v' want (!= '%v')", val, want)
		Default().reportAssertionFault(defMsg, a)
	}
}

// Equal asserts that the values are equal. If not it panics/errors (according
// the current Asserter) with the auto-generated message. You can append the
// generated got-want message by using optional message arguments.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, val, want)
		Default().reportAssertionFault(defMsg, a)
	}
}

// DeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
func DeepEqual(val, want any, a ...any) {
	if !reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, val, want)
		Default().reportAssertionFault(defMsg, a)
	}
}

// NotDeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note, it uses reflect.DeepEqual which means that also the types must be the
// same:
//
//	assert.DeepEqual(pubKey, ed25519.PublicKey(pubKeyBytes))
func NotDeepEqual(val, want any, a ...any) {
	if reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (!= '%v')", val, want)
		Default().reportAssertionFault(defMsg, a)
	}
}

// Len asserts that the length of the string is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Len(obj string, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// Longer asserts that the length of the string is longer to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Longer(s string, length int, a ...any) {
	l := len(s)

	if l <= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantLongerFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// Shorter asserts that the length of the string is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Shorter(str string, length int, a ...any) {
	l := len(str)

	if l >= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantShorterFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// SLen asserts that the length of the slice is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SLen[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// SLonger asserts that the length of the slice is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SLonger[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l <= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantLongerFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// SShorter asserts that the length of the slice is equal to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SShorter[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l >= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantShorterFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// MLen asserts that the length of the map is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MLen[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// MLonger asserts that the length of the map is longer to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MLonger[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l <= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantLongerFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// MShorter asserts that the length of the map is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MShorter[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l >= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantShorterFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// CLen asserts that the length of the chan is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CLen[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// CLonger asserts that the length of the chan is longer to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CLonger[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l <= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantLongerFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// CShorter asserts that the length of the chan is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CShorter[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l >= length {
		defMsg := fmt.Sprintf(assertionMsg+gotWantShorterFmt, l, length)
		Default().reportAssertionFault(defMsg, a)
	}
}

// MKeyExists asserts that the map key exists. If not it panics/errors (current
// Asserter) with the given message.
func MKeyExists[M ~map[T]U, T comparable, U any](obj M, key T, a ...any) (val U) {
	var ok bool
	val, ok = obj[key]

	if !ok {
		defMsg := fmt.Sprintf(assertionMsg+": key '%v' doesn't exist", key)
		Default().reportAssertionFault(defMsg, a)
	}
	return val
}

// NotEmpty asserts that the string is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
func NotEmpty(obj string, a ...any) {
	if obj == "" {
		defMsg := assertionMsg + ": string shouldn't be empty"
		Default().reportAssertionFault(defMsg, a)
	}
}

// Empty asserts that the string is empty. If it is NOT, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
func Empty(obj string, a ...any) {
	if obj != "" {
		defMsg := assertionMsg + ": string should be empty"
		Default().reportAssertionFault(defMsg, a)
	}
}

// SNotEmpty asserts that the slice is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SNotEmpty[S ~[]T, T any](obj S, a ...any) {
	l := len(obj)

	if l == 0 {
		defMsg := assertionMsg + ": slice shouldn't be empty"
		Default().reportAssertionFault(defMsg, a)
	}
}

// MNotEmpty asserts that the map is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
// You can append the generated got-want message by using optional message
// arguments.
//
// Note! This is reasonably fast but not as fast as 'That' because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MNotEmpty[M ~map[T]U, T comparable, U any](obj M, a ...any) {
	l := len(obj)

	if l == 0 {
		defMsg := assertionMsg + ": map shouldn't be empty"
		Default().reportAssertionFault(defMsg, a)
	}
}

// NoError asserts that the error is nil. If is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note. We recommend that you prefer try.To. They work exactly the same during
// the test runs and you can use the same code for both: runtime and tests.
// However, there are cases that you want assert that there is no error in cases
// where fast fail and immediate stop of execution is wanted at runtime. With
// asserts you get the file location as well. (See the asserters).
func NoError(err error, a ...any) {
	if err != nil {
		defMsg := "NoError:" + assertionMsg + conCatErrStr + err.Error()
		Default().reportAssertionFault(defMsg, a)
	}
}

// Error asserts that the err is not nil. If it is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Error(err error, a ...any) {
	if err == nil {
		defMsg := "Error:" + assertionMsg + ": missing error"
		Default().reportAssertionFault(defMsg, a)
	}
}

// Zero asserts that the value is 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Zero[T Number](val T, a ...any) {
	if val != 0 {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (== '0')", val)
		Default().reportAssertionFault(defMsg, a)
	}
}

// NotZero asserts that the value != 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NotZero[T Number](val T, a ...any) {
	if val == 0 {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (!= 0)", val)
		Default().reportAssertionFault(defMsg, a)
	}
}

// Default returns a current default asserter used for package-level
// functions like assert.That(). The package sets the default asserter as
// follows:
//
//	SetDefaultAsserter(AsserterToError | AsserterFormattedCallerInfo)
//
// Which means that it is treats assert failures as Go errors, but in addition
// to that, it formats the assertion message properly. Naturally, only if err2
// handlers are found in the call stack, these errors are caught.

// You are free to set it according to your current preferences with the
// SetDefault function.
func Default() Asserter {
	return defAsserter[def]
}

// SetDefault set the current default asserter for assert pkg.
//
// Note, that you should use this in TestMain function, and use Flag package to
// set it for the app. For the tests you can set it to panic about every
// assertion fault, or to throw an error, or/and print the call stack
// immediately when assert occurs. The err2 package helps you to catch and
// report all types of the asserts.
//
// Note, that if you are using tracers you might get two call stacks, so test
// what's best for your case.
//
// Tip. If our own packages (client packages for assert) have lots of parallel
// testing and race detection, please try to use same asserter for all of them
// and set asserter only one in TestMain, or in init.
//
//	func TestMain(m *testing.M) {
//		SetDefault(assert.TestFull)
func SetDefault(i defInd) Asserter {
	// pkg lvl lock to allow only one pkg client call this at the time
	mu.Lock()
	defer mu.Unlock()

	// to make this fully thread safe the def var should be atomic, BUT it
	// would be owerkill. We need only defaults to be set at once.
	def = i
	return defAsserter[i]
}

// mapDefInd runtime asserters, that's why test asserts are removed for now.
var mapDefInd = map[string]defInd{
	"Plain": Plain,
	"Prod":  Production,
	"Dev":   Development,
	//"Test":        Test,
	//"TestFull":    TestFull,
	"Debug": Debug,
}

var mapDefIndToString = map[defInd]string{
	Plain:       "Plain",
	Production:  "Prod",
	Development: "Dev",
	Test:        "Test",
	TestFull:    "TestFull",
	Debug:       "Debug",
}

func defaultAsserterString() string {
	return mapDefIndToString[def]
}

func newDefInd(v string) defInd {
	ind, found := mapDefInd[v]
	if !found {
		return Plain
	}
	return ind
}

func combineArgs(format string, a []any) []any {
	args := make([]any, 1, len(a)+1)
	args[0] = format
	args = append(args, a)
	return args
}

func goid() int {
	var buf [64]byte
	runtime.Stack(buf[:], false)
	return asciiWordToInt(buf[10:])
}

func asciiWordToInt(b []byte) int {
	n := 0
	for _, ch := range b {
		if ch == ' ' {
			break
		}
		ch -= '0'
		if ch > 9 {
			panic("cannot get goroutine")
		}
		n = n*10 + int(ch)
	}
	return n
}

type Number interface {
	constraints.Float | constraints.Integer
}

// String is part of the flag interfaces
func (f *flagAsserter) String() string {
	return defaultAsserterString()
}

// Get is part of the flag interfaces, getter.
func (f *flagAsserter) Get() any {
	return mapDefIndToString[def]
}

// Set is part of the flag.Value interface.
func (*flagAsserter) Set(value string) error {
	SetDefault(newDefInd(value))
	return nil
}
