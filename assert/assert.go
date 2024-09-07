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

// Asserters are the way to set what kind of messages assert package outputs if
// assertion is violated.
//
// [Plain] converts asserts just plain K&D error messages without extra
// information. That's useful for apps that want to use assert package to
// validate e.g. command fields:
//
//	assert.NotEmpty(c.PoolName, "pool name cannot be empty")
//
// Note that Plain is only asserter that override auto-generated assertion
// messages with given arguments like 'pool name cannot be empty'. Others add
// given arguments at the end of the auto-generated assert message.
//
// [Production] (pkg's default) is the best asserter for most cases. The
// assertion violations are treated as Go error values. And only a pragmatic
// caller info is included into the error values like source filename, line
// number, and caller function, all in one line:
//
//	copy file: main.go:37: CopyFile(): assertion violation: string shouldn't be empty
//
// [Development] is the best asserter for development use. The assertion
// violations are treated as Go error values. And a formatted caller info is
// included to the error message like source filename , line number, and caller
// function. Everything in a noticeable multi-line message:
//
//	--------------------------------
//	Assertion Fault at:
//	main.go:37 CopyFile():
//	assertion violation: string shouldn't be empty
//	--------------------------------
//
// [Test] minimalistic asserter for unit test use. More pragmatic is the
// [TestFull] asserter (test default).
//
// Use this asserter if your IDE/editor doesn't support full file names and it
// relies a relative path (Go standard). You can use this also if you need
// temporary problem solving for your programming environment.
//
// [TestFull] asserter (test default). The TestFull asserter includes the caller
// info and the call stack for unit testing, similarly like err2's error traces.
// Note that [PushTester] set's TestFull if it's not yet set.
//
// The call stack produced by the test asserts can be used over Go module
// boundaries. For example, if your app and it's sub packages both use
// err2/assert for unit testing and runtime checks, the runtime assertions will
// be automatically converted to test asserts. If any of the runtime asserts of
// the sub packages fails during the app tests, the app test fails as well.
//
// Note that the cross-module assertions produce long file names (path
// included), and some of the Go test result parsers cannot handle that. A
// proper test result parser like [nvim-go] (fork) works very well. Also most of
// the make result parsers can process the output properly and allow traverse of
// locations of the error trace.
//
// [Debug] asserter transforms assertion violations to panic calls where panic
// object's type is string, i.e., err2 package treats it as a normal panic, not
// an error.
//
// For example, the pattern that e.g. Go's standard library uses:
//
//	if p == nil {
//	     panic("pkg: ptr cannot be nil")
//	}
//
// is equal to:
//
//	assert.NotNil(p)
//
// [nvim-go]: https://github.com/lainio/nvim-go
const (
	Plain defInd = 0 + iota
	Production
	Development
	Test
	TestFull
	Debug
)

type flagAsserter struct{}

// Asserters
var (
	plain    = asserterToError
	prod     = asserterToError | asserterCallerInfo
	dev      = asserterToError | asserterFormattedCallerInfo
	test     = asserterUnitTesting
	testFull = asserterUnitTesting | asserterStackTrace | asserterCallerInfo
	dbg      = asserterDebug
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
	defAsserter = []asserter{plain, prod, dev, test, testFull, dbg}
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
	mapAsserter = map[int]asserter

	testersMap = map[int]testing.TB
	function   = func()
)

var (
	// testers must be set if assertion package is used for the unit testing.
	testers = x.NewRWMap[testersMap]()

	asserterMap = x.NewRWMap[mapAsserter]()
)

const (
	assertionMsg         = "assertion failure"
	assertionEqualMsg    = "assertion failure: equal"
	assertionNotEqualMsg = "assertion failure: not equal"
	assertionLenMsg      = "assertion failure: length"

	gotWantFmt        = ": got '%v', want '%v'"
	gotWantLongerFmt  = ": got '%v', should be longer than '%v'"
	gotWantShorterFmt = ": got '%v', should be shorter than '%v'"

	conCatErrStr = ": "
)

// PushTester sets the current testing context for default asserter. This must
// be called at the beginning of every test. There is two way of doing it:
//
//	for _, tt := range tests {
//	     t.Run(tt.name, func(t *testing.T) { // Shorter way, litle magic
//	          defer assert.PushTester(t)() // <- IMPORTANT! NOTE! (t)()
//	          ...
//	          assert.That(something, "test won't work")
//	     })
//	     t.Run(tt.name, func(t *testing.T) { // Longer, explicit way, 2 lines
//	          assert.PushTester(t) // <- IMPORTANT!
//	          defer assert.PopTester()
//	          ...
//	          assert.That(something, "test won't work")
//	     })
//	}
//
// Because PushTester returns [PopTester] it allows us to merge these two calls
// to one line. See the first t.Run call above. See more information in
// [PopTester].
//
// PushTester allows you to change the current default asserter by accepting it
// as a second argument.
//
// Note that you MUST call PushTester for sub-goroutines:
//
//	defer assert.PushTester(t)() // does the cleanup
//	...
//	go func() {
//	     assert.PushTester(t) // left cleanup out! Leave it for the test, see ^
//	     ...
//
// Note that the second argument, if given, changes the default asserter for
// whole package. The argument is mainly for temporary development use and isn't
// not preferred API usage.
func PushTester(t testing.TB, a ...defInd) function {
	if len(a) > 0 {
		SetDefault(a[0])
	} else if current()&asserterUnitTesting == 0 {
		// if this is forgotten or tests don't have proper place to set it
		// it's good to keep the API as simple as possible
		SetDefault(TestFull)
	}
	testers.Set(goid(), t)
	return PopTester
}

// PopTester pops the testing context reference from the memory. This is for
// memory cleanup and adding similar to err2.Catch error/panic safety for tests.
// By using PopTester you get error logs tuned for unit testing.
//
// You have two ways to call [PopTester]. With defer right after [PushTester]:
//
//	for _, tt := range tests {
//	     t.Run(tt.name, func(t *testing.T) {
//	          assert.PushTester(t) // <- important!
//	          defer assert.PopTester() // <- for good girls and not so bad boys
//	          ...
//	          assert.That(something, "test won't work")
//	     })
//	}
//
// If you want, you can combine [PushTester] and PopTester to one-liner:
//
//	defer assert.PushTester(t)()
func PopTester() {
	defer testers.Del(goid())

	r := recover()
	if r == nil {
		return
	}

	var stackLvl = 5     // amount of functions before we're here
	var framesToSkip = 3 // how many fn calls there is before FuncName call

	var msg string
	switch t := r.(type) {
	case string:
		msg = t
	case runtime.Error:
		stackLvl--     // move stack trace cursor
		framesToSkip++ // see fatal(), skip 1 more when runtime panics
		msg = t.Error()
	case error:
		msg = t.Error()
	default:
		msg = fmt.Sprintf("panic: %v", t)
	}

	// First, print the call stack. Note that we aren't support full error
	// tracing with unit test logging. However, using it has proved the top
	// level error stack as more enough. Even so that we could consider using
	// it for normal error stack traces if it would be possible.
	debug.PrintStackForTest(os.Stderr, stackLvl)

	// Now that call stack errors are printed, if any. Let's print the actual
	// line that caused the error, i.e., was throwing the error. Note that we
	// are here in the 'catch-function'.
	fatal("assertion catching: "+msg, framesToSkip)
}

func tester() (t testing.TB) {
	return testers.Get(goid())
}

// NotImplemented always panics with 'not implemented' assertion message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func NotImplemented(a ...any) {
	current().reportAssertionFault(0, assertionMsg+": not implemented", a)
}

// ThatNot asserts that the term is NOT true. If is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func ThatNot(term bool, a ...any) {
	if term {
		defMsg := assertionMsg
		current().reportAssertionFault(0, defMsg, a)
	}
}

func ThatX(term bool, a ...any) {
	if !term {
		thatXDo(a)
	}
}

func ZeroX[T Number](val T, a ...any) {
	if val != 0 {
		doZeroX(val, a)
	}
}

func doZeroX[T Number](val T, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (== '0')", val)
	currentX().reportAssertionFault(1, defMsg, a)
}

func thatXDo(a []any) {
	defMsg := assertionMsg
	currentX().reportAssertionFault(1, defMsg, a)
}

func currentX() asserter {
	// we need thread local storage, maybe we'll implement that to x.package?
	// study `tester` and copy ideas from it.
	return asserterMap.Get(goid())
}

func SetDefaultX(i defInd) {
	asserterMap.Set(goid(), defAsserter[i])
}

// That asserts that the term is true. If not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func That(term bool, a ...any) {
	if !term {
		defMsg := assertionMsg
		current().reportAssertionFault(0, defMsg, a)
	}
}

// NotNil asserts that the pointer IS NOT nil. If it is it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func NotNil[P ~*T, T any](p P, a ...any) {
	if p == nil {
		defMsg := assertionMsg + ": pointer shouldn't be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// Nil asserts that the pointer IS nil. If it is not it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Nil[T any](p *T, a ...any) {
	if p != nil {
		defMsg := assertionMsg + ": pointer should be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// INil asserts that the interface value IS nil. If it is it panics/errors
// (default Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note, use this only for real interface types. Go's interface's has two values
// so this won't work e.g. slices!
// Read more information about [the interface type].
//
// [the interface type]: https://go.dev/doc/faq#nil_error
func INil(i any, a ...any) {
	if i != nil {
		defMsg := assertionMsg + ": interface should be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// INotNil asserts that the interface value is NOT nil. If it is it
// panics/errors (default Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note, use this only for real interface types. Go's interface's has two values
// so this won't work e.g. slices!
// Read more information about [the interface type].
//
// [the interface type]: https://go.dev/doc/faq#nil_error
func INotNil(i any, a ...any) {
	if i == nil {
		defMsg := assertionMsg + ": interface shouldn't be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// SNil asserts that the slice IS nil. If it is it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func SNil[S ~[]T, T any](s S, a ...any) {
	if s != nil {
		defMsg := assertionMsg + ": slice should be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// CNil asserts that the channel is nil. If it is not it panics/errors
// (default Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func CNil[C ~chan T, T any](c C, a ...any) {
	if c != nil {
		defMsg := assertionMsg + ": channel should be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// MNil asserts that the map is nil. If it is not it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func MNil[M ~map[T]U, T comparable, U any](m M, a ...any) {
	if m != nil {
		defMsg := assertionMsg + ": map should be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// SNotNil asserts that the slice is not nil. If it is it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func SNotNil[S ~[]T, T any](s S, a ...any) {
	if s == nil {
		defMsg := assertionMsg + ": slice shouldn't be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// CNotNil asserts that the channel is not nil. If it is it panics/errors
// (default Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func CNotNil[C ~chan T, T any](c C, a ...any) {
	if c == nil {
		defMsg := assertionMsg + ": channel shouldn't be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// MNotNil asserts that the map is not nil. If it is it panics/errors (default
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func MNotNil[M ~map[T]U, T comparable, U any](m M, a ...any) {
	if m == nil {
		defMsg := assertionMsg + ": map shouldn't be nil"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// NotEqual asserts that the values aren't equal. If they are it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note, when asserter is [Plain], optional arguments are used to build a new
// assert violation message.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		doShouldNotBeEqual(assertionNotEqualMsg, val, want, a)
	}
}

// Equal asserts that the values are equal. If not it panics/errors (according
// the current Asserter) with the auto-generated message. You can append the
// generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		doShouldBeEqual(assertionEqualMsg, val, want, a)
	}
}

func doShouldBeEqual[T comparable](aname string, val, want T, a []any) {
	defMsg := fmt.Sprintf(aname+gotWantFmt, val, want)
	current().reportAssertionFault(1, defMsg, a)
}

func doShouldNotBeEqual[T comparable](aname string, val, want T, a []any) {
	defMsg := fmt.Sprintf(aname+": got '%v' want (!= '%v')", val, want)
	current().reportAssertionFault(1, defMsg, a)
}

// DeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func DeepEqual(val, want any, a ...any) {
	if !reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(assertionMsg+gotWantFmt, val, want)
		current().reportAssertionFault(0, defMsg, a)
	}
}

// NotDeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note, it uses reflect.DeepEqual which means that also the types must be the
// same:
//
//	assert.DeepEqual(pubKey, ed25519.PublicKey(pubKeyBytes))
func NotDeepEqual(val, want any, a ...any) {
	if reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(
			assertionMsg+": got '%v', want (!= '%v')",
			val,
			want,
		)
		current().reportAssertionFault(0, defMsg, a)
	}
}

// Len asserts that the length of the string is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Len(obj string, length int, a ...any) {
	l := len(obj)

	if l != length {
		doShouldBeEqual(assertionLenMsg, l, length, a)
	}
}

// Longer asserts that the length of the string is longer to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Longer(s string, length int, a ...any) {
	l := len(s)

	if l <= length {
		doLonger(l, length, a)
	}
}

func doLonger(l int, length int, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+gotWantLongerFmt, l, length)
	current().reportAssertionFault(1, defMsg, a)
}

// Shorter asserts that the length of the string is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func Shorter(str string, length int, a ...any) {
	l := len(str)

	if l >= length {
		doShorter(l, length, a)
	}
}

func doShorter(l int, length int, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+gotWantShorterFmt, l, length)
	current().reportAssertionFault(1, defMsg, a)
}

// SLen asserts that the length of the slice is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SLen[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l != length {
		doShouldBeEqual(assertionLenMsg, l, length, a)
	}
}

// SLonger asserts that the length of the slice is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SLonger[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l <= length {
		doLonger(l, length, a)
	}
}

// SShorter asserts that the length of the slice is equal to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SShorter[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l >= length {
		doShorter(l, length, a)
	}
}

// MLen asserts that the length of the map is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MLen[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l != length {
		doShouldBeEqual(assertionLenMsg, l, length, a)
	}
}

// MLonger asserts that the length of the map is longer to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MLonger[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l <= length {
		doLonger(l, length, a)
	}
}

// MShorter asserts that the length of the map is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MShorter[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l >= length {
		doShorter(l, length, a)
	}
}

// CLen asserts that the length of the chan is equal to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CLen[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l != length {
		doShouldBeEqual(assertionLenMsg, l, length, a)
	}
}

// CLonger asserts that the length of the chan is longer to the given. If not it
// panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CLonger[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l <= length {
		doLonger(l, length, a)
	}
}

// CShorter asserts that the length of the chan is shorter to the given. If not
// it panics/errors (according the current Asserter) with the auto-generated
// message. You can append the generated got-want message by using optional
// message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func CShorter[C ~chan T, T any](obj C, length int, a ...any) {
	l := len(obj)

	if l >= length {
		doShorter(l, length, a)
	}
}

// MKeyExists asserts that the map key exists. If not it panics/errors (current
// Asserter) the auto-generated (args appended) message.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func MKeyExists[M ~map[T]U, T comparable, U any](
	obj M,
	key T,
	a ...any,
) (val U) {
	var ok bool
	val, ok = obj[key]

	if !ok {
		doMKeyExists(key, a)
	}
	return val
}

func doMKeyExists(key any, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": key '%v' doesn't exist", key)
	current().reportAssertionFault(1, defMsg, a)
}

// NotEmpty asserts that the string is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func NotEmpty(obj string, a ...any) {
	if obj == "" {
		defMsg := assertionMsg + ": string shouldn't be empty"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// Empty asserts that the string is empty. If it is NOT, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Empty(obj string, a ...any) {
	if obj != "" {
		defMsg := assertionMsg + ": string should be empty"
		current().reportAssertionFault(0, defMsg, a)
	}
}

// SEmpty asserts that the slice is empty. If it is NOT, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SEmpty[S ~[]T, T any](obj S, a ...any) {
	l := len(obj)

	if l != 0 {
		doEmptyNamed("", "slice", a)
	}
}

// SNotEmpty asserts that the slice is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func SNotEmpty[S ~[]T, T any](obj S, a ...any) {
	l := len(obj)

	if l == 0 {
		doEmptyNamed("not", "slice", a)
	}
}

// MEmpty asserts that the map is empty. If it is NOT, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
// You can append the generated got-want message by using optional message
// arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MEmpty[M ~map[T]U, T comparable, U any](obj M, a ...any) {
	l := len(obj)

	if l != 0 {
		doEmptyNamed("", "map", a)
	}
}

// MNotEmpty asserts that the map is not empty. If it is, it panics/errors
// (according the current Asserter) with the auto-generated message. You can
// append the generated got-want message by using optional message arguments.
// You can append the generated got-want message by using optional message
// arguments.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
//
// Note! This is reasonably fast but not as fast as [That] because of lacking
// inlining for the current implementation of Go's type parametric functions.
func MNotEmpty[M ~map[T]U, T comparable, U any](obj M, a ...any) {
	l := len(obj)

	if l == 0 {
		doEmptyNamed("not", "map", a)
	}
}

func doEmptyNamed(not, name string, a []any) {
	not = x.Whom(not == "not", " not ", "")
	defMsg := assertionMsg + ": " + name + " should" + not + "be empty"
	current().reportAssertionFault(1, defMsg, a)
}

// NoError asserts that the error is nil. If is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note. We recommend that you prefer [github.com/lainio/err2/try.To]. They work
// exactly the same during the test runs and you can use the same code for both:
// runtime and tests. However, there are cases that you want assert that there
// is no error in cases where fast fail and immediate stop of execution is
// wanted at runtime. With asserts ([Production], [Development], [Debug]) you
// get the file location as well.
func NoError(err error, a ...any) {
	if err != nil {
		defMsg := assertionMsg + conCatErrStr + err.Error()
		current().reportAssertionFault(0, defMsg, a)
	}
}

// Error asserts that the err is not nil. If it is it panics and builds a
// violation message. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Error(err error, a ...any) {
	if err == nil {
		doErrorX(a)
	}
}

func ErrorX(err error, a ...any) {
	if err == nil {
		doError(a)
	}
}

func doErrorX(a []any) {
	defMsg := "Error:" + assertionMsg + ": missing error"
	currentX().reportAssertionFault(0, defMsg, a)
}

func doError(a []any) {
	defMsg := "Error:" + assertionMsg + ": missing error"
	current().reportAssertionFault(0, defMsg, a)
}

// Greater asserts that the value is greater than want. If it is not it panics
// and builds a violation message. Thanks to inlining, the performance penalty
// is equal to a single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Greater[T Number](val, want T, a ...any) {
	if val <= want {
		doGreater(val, want, a)
	}
}

func doGreater[T Number](val, want T, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": got '%v', want <= '%v'", val, want)
	current().reportAssertionFault(1, defMsg, a)
}

// Less asserts that the value is less than want. If it is not it panics and
// builds a violation message. Thanks to inlining, the performance penalty is
// equal to a single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Less[T Number](val, want T, a ...any) {
	if val >= want {
		doLess(val, want, a)
	}
}

func doLess[T Number](val, want T, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": got '%v', want >= '%v'", val, want)
	current().reportAssertionFault(1, defMsg, a)
}

// Zero asserts that the value is 0. If it is not it panics and builds a
// violation message. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func Zero[T Number](val T, a ...any) {
	if val != 0 {
		doZero(val, a)
	}
}

func doZero[T Number](val T, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (== '0')", val)
	current().reportAssertionFault(1, defMsg, a)
}

// NotZero asserts that the value != 0. If it is not it panics and builds a
// violation message. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
//
// Note that when [Plain] asserter is used ([SetDefault]), optional arguments
// are used to override the auto-generated assert violation message.
func NotZero[T Number](val T, a ...any) {
	if val == 0 {
		doNotZero(val, a)
	}
}

func doNotZero[T Number](val T, a []any) {
	defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (!= 0)", val)
	current().reportAssertionFault(1, defMsg, a)
}

// current returns a current default asserter used for package-level
// functions like assert.That().
//
// Note, this indexing stuff is done because of race detection to work on client
// packages. And, yes, we have tested it. This is fastest way to make it without
// locks HERE. Only the setting the index is secured with the mutex.
func current() asserter {
	return defAsserter[def]
}

// SetDefault sets the current default asserter for assert pkg. It also returns
// the previous asserter.
//
// Note that you should use this in TestMain function, and use [flag] package to
// set it for the app. For the tests you can set it to panic about every
// assertion fault, or to throw an error, or/and print the call stack
// immediately when assert occurs. The err2 package helps you to catch and
// report all types of the asserts.
//
// Note that if you are using tracers you might get two call stacks, so test
// what's best for your case.
//
// Tip. If our own packages (client packages for assert) have lots of parallel
// testing and race detection, please try to use same asserter for all of them
// and set asserter only one in TestMain, or in init.
//
//	func TestMain(m *testing.M) {
//	     SetDefault(assert.TestFull)
func SetDefault(i defInd) (old defInd) {
	// pkg lvl lock to allow only one pkg client call this at one of the time
	// together with the indexing, i.e we don't need to switch asserter
	// variable or pointer to it but just index to array they are stored.
	// All of this make race detector happy at the client pkgs.
	mu.Lock()
	defer mu.Unlock()

	old = def
	// theoretically, to make this fully thread safe the def var should be
	// atomic, BUT it would be overkill. We need only defaults to be set at
	// once. AND because we use indexing to actual asserter the thread-safety
	// and performance are guaranteed,
	def = i
	return
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
