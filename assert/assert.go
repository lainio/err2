package assert

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/lainio/err2/internal/x"
	"golang.org/x/exp/constraints"
)

// TODO: how about combine these to static (private) array and use only index to
// that for defaultAsserter. maybe get rid of performance penalty?
var (
	// P is a production Asserter that sets panic objects to errors which
	// allows err2 handlers to catch them.
	P = AsserterToError

	// D is a development Asserter that sets panic objects to strings that
	// doesn't by caught by err2 handlers.
	D = AsserterDebug

	T = AsserterUnitTesting | AsserterStackTrace | AsserterCallerInfo
)

var (
	defaultAsserter = atomic.Pointer[Asserter]{}
)

func init() {
	SetDefaultAsserter(AsserterToError | AsserterFormattedCallerInfo)
}

type (
	testersMap = x.TMap[int, testing.TB]
	function = func()
)

var (
	// testers is must be set if assertion package is used for the unit testing.
	testers = x.NewRWMap[int, testing.TB]()
)

const (
	assertionMsg = "assertion violation"
)

// PushTester sets the current testing context for default asserter. This must
// be called at the beginning of every test.
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.PushTester(t) // <- IMPORTANT!
//			defer assert.PopTester()
//			...
//			assert.That(something, "test won't work")
//		})
//	}
//
func PushTester(t testing.TB) function { // TODO: add argument (def asserter for the test)
	if DefaultAsserter()&AsserterUnitTesting == 0 {
		// if this is forgotten or tests don't have proper place to set it
		// it's good to keep the API as simple as possible
		SetDefaultAsserter(AsserterUnitTesting)
	}
	x.Set(testers, goid(), t)
	return PopTester
}

// PopTester pops the testing context reference from the memory. This isn't
// totally necessary, but if you want play by book, please do it. Usually done
// by defer after PushTester.
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			assert.PushTester(t) // <- important!
//			defer assert.PopTester() // <- for good girls and not so bad boys
//			...
//			assert.That(something, "test won't work")
//		})
//	}
func PopTester() {

	x.Tx(testers, func(m testersMap) {
		delete(m, goid())
	})
}

func tester() (t testing.TB) {
	return x.Get(testers, goid())
}

// NotImplemented always panics with 'not implemented' assertion message.
func NotImplemented(a ...any) {
	if DefaultAsserter().isUnitTesting() {
		tester().Helper()
	}
	DefaultAsserter().reportAssertionFault("not implemented", a...)
}

// ThatNot asserts that the term is NOT true. If is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func ThatNot(term bool, a ...any) {
	if term {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := "ThatNot: " + assertionMsg
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// That asserts that the term is true. If not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func That(term bool, a ...any) {
	if !term {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := "That: " + assertionMsg
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// NotNil asserts that the pointer IS NOT nil. If it is it panics/errors (default
// Asserter) with the given message.
func NotNil[T any](p *T, a ...any) {
	if p == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": pointer shouldn't be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Nil asserts that the pointer IS nil. If it is not it panics/errors (default
// Asserter) with the given message.
func Nil[T any](p *T, a ...any) {
	if p != nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": pointer should be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// INil asserts that the interface value IS nil. If it is it panics/errors
// (default Asserter) with the given message.
func INil(i any, a ...any) {
	if i != nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": interface should be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// INotNil asserts that the interface value is NOT nil. If it is it
// panics/errors (default Asserter) with the given message.
func INotNil(i any, a ...any) {
	if i == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": interface shouldn't be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// SNil asserts that the slice IS nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNil[T any](s []T, a ...any) {
	if s != nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": slice should be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// SNotNil asserts that the slice is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func SNotNil[T any](s []T, a ...any) {
	if s == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": slice shouldn't be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// CNotNil asserts that the channel is not nil. If it is it panics/errors
// (default Asserter) with the given message.
func CNotNil[T any](c chan T, a ...any) {
	if c == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": channel shouldn't be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// MNotNil asserts that the map is not nil. If it is it panics/errors (default
// Asserter) with the given message.
func MNotNil[T comparable, U any](m map[T]U, a ...any) {
	if m == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": map shouldn't be nil"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// NotEqual asserts that the values aren't equal. If they are it panics/errors
// (current Asserter) with the given message.
func NotEqual[T comparable](val, want T, a ...any) {
	if want == val {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v want (!= %v)", val, want)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Equal asserts that the values are equal. If not it panics/errors (current
// Asserter) with the given message.
func Equal[T comparable](val, want T, a ...any) {
	if want != val {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v, want %v", val, want)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// DeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (current Asserter) with the given message.
func DeepEqual(val, want any, a ...any) {
	if !reflect.DeepEqual(val, want) {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v, want %v", val, want)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// NotDeepEqual asserts that the (whatever) values are equal. If not it
// panics/errors (current Asserter) with the given message. NOTE, it uses
// reflect.DeepEqual which means that also the types must be the same:
//
//	assert.DeepEqual(pubKey, ed25519.PublicKey(pubKeyBytes))
func NotDeepEqual(val, want any, a ...any) {
	if reflect.DeepEqual(val, want) {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v, want (!= %v)", val, want)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Len asserts that the length of the string is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func Len(obj string, length int, a ...any) {
	l := len(obj)

	if l != length {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %d, want %d", l, length)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// SLen asserts that the length of the slice is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func SLen[T any](obj []T, length int, a ...any) {
	l := len(obj)

	if l != length {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %d, want %d", l, length)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// MLen asserts that the length of the map is equal to the given. If not it
// panics/errors (current Asserter) with the given message. Note! This is
// reasonably fast but not as fast as 'That' because of lacking inlining for the
// current implementation of Go's type parametric functions.
func MLen[T comparable, U any](obj map[T]U, length int, a ...any) {
	l := len(obj)

	if l != length {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %d, want %d", l, length)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// MKeyExists asserts that the map key exists. If not it panics/errors (current
// Asserter) with the given message.
func MKeyExists[T comparable, U any](obj map[T]U, key T, a ...any) (val U) {
	var ok bool
	val, ok = obj[key]

	if !ok {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": key '%v' doesn't exist", key)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
	return val
}

// NotEmpty asserts that the string is not empty. If it is, it panics/errors
// (current Asserter) with the given message.
func NotEmpty(obj string, a ...any) {
	if obj == "" {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": string shouldn't be empty"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Empty asserts that the string is empty. If it is NOT, it panics/errors
// (current Asserter) with the given message.
func Empty(obj string, a ...any) {
	if obj != "" {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": string should be empty"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// SNotEmpty asserts that the slice is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func SNotEmpty[T any](obj []T, a ...any) {
	l := len(obj)

	if l == 0 {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": slice shouldn't be empty"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// MNotEmpty asserts that the map is not empty. If it is, it panics/errors
// (current Asserter) with the given message. Note! This is reasonably fast but
// not as fast as 'That' because of lacking inlining for the current
// implementation of Go's type parametric functions.
func MNotEmpty[T comparable, U any](obj map[T]U, a ...any) {
	l := len(obj)

	if l == 0 {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := assertionMsg + ": map shouldn't be empty"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// NoError asserts that the error is nil. If is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NoError(err error, a ...any) {
	if err != nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := "NoError:" + assertionMsg + ": " + err.Error()
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Error asserts that the err is not nil. If it is it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Error(err error, a ...any) {
	if err == nil {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := "Error:" + assertionMsg + ": missing error"
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// Zero asserts that the value is 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func Zero[T Number](val T, a ...any) {
	if val != 0 {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v, want (== 0)", val)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// NotZero asserts that the value != 0. If it is not it panics with the given
// formatting string. Thanks to inlining, the performance penalty is equal to a
// single 'if-statement' that is almost nothing.
func NotZero[T Number](val T, a ...any) {
	if val == 0 {
		if DefaultAsserter().isUnitTesting() {
			tester().Helper()
		}
		defMsg := fmt.Sprintf(assertionMsg+": got %v, want (!= 0)", val)
		DefaultAsserter().reportAssertionFault(defMsg, a...)
	}
}

// DefaultAsserter returns a current default asserter used for package-level
// functions like assert.That(). The package sets the default asserter as
// follows:
//
//	SetDefaultAsserter(AsserterToError | AsserterFormattedCallerInfo)
//
// Which means that it is treats assert failures as Go errors, but in addition
// to that, it formats the assertion message properly. Naturally, only if err2
// handlers are found in the call stack, these errors are caught.

// You are free to set it according to your current preferences with the
// SetDefaultAsserter function.
func DefaultAsserter() Asserter {
	return *defaultAsserter.Load()
}

// SetDefaultAsserter set the current default asserter for the package. For
// example, you might set it to panic about every assertion fault, and in other
// cases, throw an error, and print the call stack immediately when assert
// occurs. Note, that if you are using tracers you might get two call stacks, so
// test what's best for your case.
//
//	SetDefaultAsserter(AsserterDebug | AsserterStackTrace)
func SetDefaultAsserter(a Asserter) Asserter {
	c := defaultAsserter.Load()
	defaultAsserter.Store(&a)
	if c == nil {
		return a
	}
	return *c
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
	var id int
	_, err := fmt.Fscanf(bytes.NewReader(buf[:]), "goroutine %d", &id)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

type Number interface {
	constraints.Float | constraints.Integer
}
