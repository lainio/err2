package assert

import (
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
	Production defInd = 0 + iota
	Development
	Test
	TestFull
	Debug
)

// Deprecated: use e.g. assert.That(), only default asserter is used.
var (
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
	// These two are our indexing system for default asserter. Note also the
	// mutex blew. All of this is done to keep client package race detector
	// cool.
	defAsserter = []Asserter{P, B, T, TF, D}
	def         defInd

	// mu is package lvl Mutex that is used to cool down race detector of
	// client pkgs, i.e. packages that use us can use -race flag in their test
	// runs where they change asserter. With the mutex we can at least allow
	// the setters run at one of the time. AND when that's combined with the
	// indexing system we are using for default asserter (above) we are pretty
	// much theard safe.
	mu sync.Mutex
)

func init() {
	SetDefault(Production)
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
	assertionMsg = "assertion violation"
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
// one line. See the first t.Run call. NOTE. More information in PopTester.
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

	// First, print the call stack. Note. that we aren't support full error
	// tracing with unit test logging. However, using it has proved the top
	// level error stack as more enough. Even so that we could consider using
	// it for normal error stack straces if it would be possible.
	const stackLvl = 6 // amount of functions before we're here
	debug.PrintStackForTest(os.Stderr, stackLvl)

	// Now that call stack errors are printed, if any. Let's print the actual
	// line that caused the error, i.e., was throwing the error. Note that we
	// are here in the 'catch-function'.
	const framesToSkip = 4 // how many fn calls there is before FuncName call
	fatal("assertion catching: "+msg, framesToSkip)
}

func tester() (t testing.TB) {
	return testers.Get(goid())
}

// NotImplemented always panics with 'not implemented' assertion message.
func NotImplemented(a ...any) {
	Default().reportAssertionFault("not implemented", a...)
}

// ThatNot asserts that the term is NOT true. If is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func ThatNot(term bool, a ...any) {
	if term {
		defMsg := "ThatNot: " + assertionMsg
		Default().reportAssertionFault(defMsg, a...)
	}
}

// That asserts that the term is true. If not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func That(term bool, a ...any) {
	if !term {
		defMsg := "That: " + assertionMsg
		Default().reportAssertionFault(defMsg, a...)
	}
}

// NotNil asserts that the pointer IS NOT nil. If it is it panics/errors (default
// Asserter) with the given message.
func NotNil[P ~*T, T any](p P, a ...any) {
	if p == nil {
		defMsg := assertionMsg + ": pointer shouldn't be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Nil asserts that the pointer IS nil. If it is not it panics/errors (default
// Asserter) with the given message.
func Nil[T any](p *T, a ...any) {
	if p != nil {
		defMsg := assertionMsg + ": pointer should be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// INil asserts that the interface value IS nil. If it is it panics/errors
// (default Asserter) with the given message.
func INil(i any, a ...any) {
	if i != nil {
		defMsg := assertionMsg + ": interface should be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// INotNil asserts that the interface value is NOT nil. If it is it
// panics/errors (default Asserter) with the given message.
func INotNil(i any, a ...any) {
	if i == nil {
		defMsg := assertionMsg + ": interface shouldn't be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// SNil asserts that the slice IS nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNil[S ~[]T, T any](s S, a ...any) {
	if s != nil {
		defMsg := assertionMsg + ": slice should be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// SNotNil asserts that the slice is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNotNil[S ~[]T, T any](s S, a ...any) {
	if s == nil {
		defMsg := assertionMsg + ": slice shouldn't be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// CNotNil asserts that the channel is not nil. If it is it panics/errors
// (default Asserter) with the given message.
func CNotNil[C ~chan T, T any](c C, a ...any) {
	if c == nil {
		defMsg := assertionMsg + ": channel shouldn't be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// MNotNil asserts that the map is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func MNotNil[M ~map[T]U, T comparable, U any](m M, a ...any) {
	if m == nil {
		defMsg := assertionMsg + ": map shouldn't be nil"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// NotEqual asserts that the values aren't equal. If they are it panics/errors
// (current Asserter) with the given message.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v' want (!= '%v')", val, want)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Equal asserts that the values are equal. If not it panics/errors (current
// Asserter) with the given message.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want '%v'", val, want)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// DeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (current Asserter) with the given message.
func DeepEqual(val, want any, a ...any) {
	if !reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want '%v'", val, want)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// NotDeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (current Asserter) with the given message. NOTE, it uses
// reflect.DeepEqual which means that also the types must be the same:
//
//	assert.DeepEqual(pubKey, ed25519.PublicKey(pubKeyBytes))
func NotDeepEqual(val, want any, a ...any) {
	if reflect.DeepEqual(val, want) {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (!= '%v')", val, want)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Len asserts that the length of the string is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func Len(obj string, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+": got '%d', want '%d'", l, length)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// SLen asserts that the length of the slice is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func SLen[S ~[]T, T any](obj S, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+": got '%d', want '%d'", l, length)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// MLen asserts that the length of the map is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func MLen[M ~map[T]U, T comparable, U any](obj M, length int, a ...any) {
	l := len(obj)

	if l != length {
		defMsg := fmt.Sprintf(assertionMsg+": got '%d', want '%d'", l, length)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// MKeyExists asserts that the map key exists. If not it panics/errors (current
// Asserter) with the given message.
func MKeyExists[M ~map[T]U, T comparable, U any](obj M, key T, a ...any) (val U) {
	var ok bool
	val, ok = obj[key]

	if !ok {
		defMsg := fmt.Sprintf(assertionMsg+": key '%v' doesn't exist", key)
		Default().reportAssertionFault(defMsg, a...)
	}
	return val
}

// NotEmpty asserts that the string is not empty. If it is, it panics/errors
// (current Asserter) with the given message.
func NotEmpty(obj string, a ...any) {
	if obj == "" {
		defMsg := assertionMsg + ": string shouldn't be empty"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Empty asserts that the string is empty. If it is NOT, it panics/errors
// (current Asserter) with the given message.
func Empty(obj string, a ...any) {
	if obj != "" {
		defMsg := assertionMsg + ": string should be empty"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// SNotEmpty asserts that the slice is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func SNotEmpty[S ~[]T, T any](obj S, a ...any) {
	l := len(obj)

	if l == 0 {
		defMsg := assertionMsg + ": slice shouldn't be empty"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// MNotEmpty asserts that the map is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func MNotEmpty[M ~map[T]U, T comparable, U any](obj M, a ...any) {
	l := len(obj)

	if l == 0 {
		defMsg := assertionMsg + ": map shouldn't be empty"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// NoError asserts that the error is nil. If is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NoError(err error, a ...any) {
	if err != nil {
		defMsg := "NoError:" + assertionMsg + ": " + err.Error()
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Error asserts that the err is not nil. If it is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Error(err error, a ...any) {
	if err == nil {
		defMsg := "Error:" + assertionMsg + ": missing error"
		Default().reportAssertionFault(defMsg, a...)
	}
}

// Zero asserts that the value is 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Zero[T Number](val T, a ...any) {
	if val != 0 {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (== '0')", val)
		Default().reportAssertionFault(defMsg, a...)
	}
}

// NotZero asserts that the value != 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NotZero[T Number](val T, a ...any) {
	if val == 0 {
		defMsg := fmt.Sprintf(assertionMsg+": got '%v', want (!= 0)", val)
		Default().reportAssertionFault(defMsg, a...)
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

// SetDefault set the current default asserter for the package. For example, you
// might set it to panic about every assertion fault, and in other cases, throw
// an error, and print the call stack immediately when assert occurs. Note, that
// if you are using tracers you might get two call stacks, so test what's best
// for your case. Tip. If our own packages (client packages for assert) have
// lots of parallel testing and race detection, please try to use same asserter
// for allo foot hem and do it only one in TestMain, or in init.
//
//	SetDefault(assert.TestFull)
func SetDefault(i defInd) Asserter {
	// pkg lvl lock to allow only one pkg client call this at the time
	mu.Lock()
	defer mu.Unlock()

	// to make this fully thread safe the def var should be atomic, BUT it
	// would be owerkill. We need only defaults to be set at once.
	def = i
	return defAsserter[i]
}

func combineArgs(format string, a []any) []any {
	args := make([]any, 1, len(a)+1)
	args[0] = format
	args = append(args, a...)
	return args
}

func goid() int {
	var buf [64]byte
	runtime.Stack(buf[:], false)
	return myByteToInt(buf[10:])
}

func myByteToInt(b []byte) int {
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
