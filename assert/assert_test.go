package assert_test // Note!! Some tests here are related to line # of the file

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

func ExampleThat() {
	sample := func() (err error) {
		defer err2.Handle(&err)

		assert.That(false, "optional message")
		return err
	}
	err := sample()
	fmt.Printf("%v", err)
	// Output: testing: run example: assert_test.go:17: ExampleThat.func1(): assertion failure: optional message
}

func ExampleNotNil() {
	sample := func(b *byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Nil(b)    // OK
		assert.NotNil(b) // Not OK
		return err
	}
	var b *byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:30: ExampleNotNil.func1(): assertion failure: pointer should not be nil
}

func ExampleMNotNil() {
	sample := func(b map[string]byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.MEmpty(b)  // OK
		assert.MNil(b)    // OK
		assert.MNotNil(b) // Not OK
		return err
	}
	var b map[string]byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:45: ExampleMNotNil.func1(): assertion failure: map should not be nil
}

func ExampleCNotNil() {
	sample := func(c chan byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.CNil(c)    // OK
		assert.CNotNil(c) // Not OK
		return err
	}
	var c chan byte
	err := sample(c)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:59: ExampleCNotNil.func1(): assertion failure: channel should not be nil
}

func ExampleSNotNil() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SEmpty(b)  // OK
		assert.SNil(b)    // OK
		assert.SNotNil(b) // Not OK
		return err
	}
	var b []byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:74: ExampleSNotNil.func1(): assertion failure: slice should not be nil
}

func ExampleEqual() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.NotEqual(b[0], 3) // OK, b[0] != 3; (b[0] == 1)
		assert.Equal(b[1], 1)    // Not OK, b[1] == 2
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:88: ExampleEqual.func1(): assertion failure: equal: got '2', want '1'
}

func ExampleSLen() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SLen(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:100: ExampleSLen.func1(): assertion failure: length: got '2', want '3'
}

func ExampleSNotEmpty() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SNotEmpty(b)
		return err
	}
	err := sample([]byte{})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:112: ExampleSNotEmpty.func1(): assertion failure: slice should not be empty
}

func ExampleNotEmpty() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Empty(b)    // OK
		assert.NotEmpty(b) // not OK
		return err
	}
	err := sample("")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:125: ExampleNotEmpty.func1(): assertion failure: string should not be empty
}

func ExampleMKeyExists() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		m := map[string]string{
			"1": "one",
		}
		v := assert.MKeyExists(m, "1") // OK, 1 --> one
		assert.Equal(v, "one")         // OK
		_ = assert.MKeyExists(m, b)    // fails with b = 2
		return err
	}
	err := sample("2")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:142: ExampleMKeyExists.func1(): assertion failure: key '2' doesn't exist
}

func ExampleZero() {
	sample := func(b int8) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Zero(b)
		return err
	}
	var b int8 = 1 // we want sample to assert the violation.
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:154: ExampleZero.func1(): assertion failure: got '1', want (== '0')
}

func ExampleSLonger() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SLonger(b, 0) // OK
		assert.SLonger(b, 1) // Not OK
		return err
	}
	err := sample([]byte{01}) // len = 1
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:168: ExampleSLonger.func1(): assertion failure: got '1', should be longer than '1'
}

func ExampleMShorter() {
	sample := func(b map[byte]byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.MNotEmpty(b)   // OK
		assert.MShorter(b, 1) // OK
		assert.MShorter(b, 0) // Not OK
		return err
	}
	err := sample(map[byte]byte{01: 01}) // len = 1
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:181: ExampleMShorter.func1(): assertion failure: got '1', should be shorter than '1'
}

func ExampleSShorter() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SShorter(b, 2)                                      // ok
		assert.SShorter(b, 0, "optional message (%s)", "test_str") // not ok
		return err
	}
	err := sample([]byte{01}) // len = 1
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:195: ExampleSShorter.func1(): assertion failure: got '1', should be shorter than '0': optional message (test_str)
}

func ExampleLess() {
	sample := func(b int8) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Equal(b, 1) // ok
		assert.Less(b, 2)  // ok
		assert.Less(b, 1)  // not ok
		return err
	}
	var b int8 = 1
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:209: ExampleLess.func1(): assertion failure: got '1', want >= '1'
}

func ExampleGreater() {
	sample := func(b int8) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Equal(b, 2)   // ok
		assert.Greater(b, 1) // ok
		assert.Greater(b, 2) // not ok
		return err
	}
	var b int8 = 2
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:224: ExampleGreater.func1(): assertion failure: got '2', want <= '2'
}

func ExampleNotZero() {
	sample := func(b int8) (err error) {
		defer err2.Handle(&err, "sample")

		assert.NotZero(b)
		return err
	}
	var b int8
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:237: ExampleNotZero.func1(): assertion failure: got '0', want (!= 0)
}

func ExampleMLen() {
	sample := func(b map[int]byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.MLonger(b, 1)  // OK
		assert.MShorter(b, 3) // OK
		assert.MLen(b, 3)     // Not OK
		return err
	}
	err := sample(map[int]byte{1: 1, 2: 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:252: ExampleMLen.func1(): assertion failure: length: got '2', want '3'
}

func ExampleCLen() {
	sample := func(b chan int) (err error) {
		defer err2.Handle(&err, "sample")

		assert.CLonger(b, 1)  // OK
		assert.CShorter(b, 3) // OK
		assert.CLen(b, 3)     // Not OK
		return err
	}
	d := make(chan int, 2)
	d <- int(1)
	d <- int(1)
	err := sample(d)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:266: ExampleCLen.func1(): assertion failure: length: got '2', want '3'
}

func ExampleThatNot() {
	sample := func() (err error) {
		defer err2.Handle(&err)

		assert.ThatNot(true, "overrides if Plain asserter")
		return err
	}

	// Set a specific asserter for this goroutine only, we want plain errors
	defer assert.PushAsserter(assert.Plain)()

	err := sample()
	fmt.Printf("%v", err)
	// Output: testing: run example: overrides if Plain asserter
}

func ExampleINotNil() {
	sample := func(b error) (err error) {
		defer err2.Handle(&err, "sample")

		assert.INotNil(b) // OK
		assert.INil(b)    // Not OK
		return err
	}
	var b = fmt.Errorf("test")
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:298: ExampleINotNil.func1(): assertion failure: interface should be nil
}

func ExampleLen() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Shorter(b, 3) // OK
		assert.Longer(b, 1)  // OK
		assert.Len(b, 3)     // Not OK
		return err
	}
	err := sample("12")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:313: ExampleLen.func1(): assertion failure: length: got '2', want '3'
}

func ExampleDeepEqual() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.NoError(err)
		assert.NotDeepEqual(len(b), 3) // OK, correct size is 2
		assert.DeepEqual(len(b), 3)    // Not OK, size is still 2
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:327: ExampleDeepEqual.func1(): assertion failure: got '2', want '3'
}

func ExampleError() {
	sample := func(b error) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Error(b)   // OK
		assert.NoError(b) // Not OK
		return err
	}
	var b = fmt.Errorf("test")
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:340: ExampleError.func1(): assertion failure: test
}

func ExampleNotImplemented() {
	sample := func(_ error) (err error) {
		defer err2.Handle(&err, "sample")

		assert.NotImplemented() // Not OK
		return err
	}
	var b = fmt.Errorf("test")
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:353: ExampleNotImplemented.func1(): assertion failure: not implemented
}

func ExampleMKeyNotExists() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		m := map[string]string{
			"1": "one",
		}
		assert.MKeyNotExists(m, b)   // Doesn't fail: ∄ b ∈ m | b=2
		assert.MKeyNotExists(m, "1") // Fails: ∃ 1 ∈ m
		return err
	}
	err := sample("2")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:370: ExampleMKeyNotExists.func1(): assertion failure: key '1' shouldn't exist
}

func ExampleThat_sentinelErrorWithIs() {
	var ErrArgumentsNeeded = errors.New("arguments missing")

	sample := func(a ...int) (err error) {
		// [err2.Handle] default is automatic error annotation, caller needs
		// to use [errors.Is] to check error values returned.
		defer err2.Handle(&err)

		// we can use asserts to return actual error values
		assert.That(len(a) != 0, ErrArgumentsNeeded)
		return err
	}
	err := sample()
	// because sample function annotates error values we must use [errors.Is]
	if errors.Is(err, ErrArgumentsNeeded) {
		// err is wrapped (annotated) by [err2.Handle]
		fmt.Printf("ERR: %v", err)
	} else {
		fmt.Print("never here!", err)
	}
	// Output: ERR: testing: run example: assert_test.go:387: ExampleThat_sentinelErrorWithIs.func1(): assertion failure: arguments missing
}

func ExampleThat_sentinelErrorWithValueComparison() {
	var ErrArgumentsNeeded = errors.New("arguments missing")

	sample := func(a ...int) (err error) {
		// prevent assert pkg's to add extra info to err values
		defer assert.PushAsserter(assert.Plain)()
		// also remove automatic error annotation with nil argument
		defer err2.Handle(&err, nil)

		// we can use asserts to return actual error values
		assert.That(len(a) != 0, ErrArgumentsNeeded)
		return err
	}

	// sample function is documented to return just plain error values
	err := sample()

	if err == ErrArgumentsNeeded {
		fmt.Printf("ERR: %v", err)
	} else {
		fmt.Print("never here!")
	}
	// Output: ERR: arguments missing
}

func ExampleEqual_sentinel() {
	var ErrNotEqual = errors.New("different values")

	sample := func(b []byte) (err error) {
		defer assert.PushAsserter(assert.Plain)()

		defer err2.Handle(&err, "sample")

		assert.NotEqual(b[0], 3) // OK, b[0] != 3; (b[0] == 1)

		// Note that only ErrNotEqual value is used, rest of the args are
		// ignored.
		assert.Equal(b[1], 1, ErrNotEqual, "%d", 1) // Not OK, b[1] == 2
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: different values
}

func BenchmarkMKeyExists(b *testing.B) {
	bs := map[int]int{0: 0, 1: 1}
	for n := 0; n < b.N; n++ {
		assert.MKeyExists(bs, 1)
	}
}

func BenchmarkMKeyExistsOKIdiom(b *testing.B) {
	bs := map[int]int{0: 0, 1: 1}
	found := false
	for n := 0; n < b.N; n++ {
		_, ok := bs[1]
		if ok {
			found = ok
		}
	}
	_ = found
}

func BenchmarkMNotEmpty(b *testing.B) {
	bs := map[int]int{0: 0, 1: 1}
	for n := 0; n < b.N; n++ {
		assert.MNotEmpty(bs)
	}
}

func BenchmarkMEmpty(b *testing.B) {
	bs := map[int]int{}
	for n := 0; n < b.N; n++ {
		assert.MEmpty(bs)
	}
}

func BenchmarkNotEmpty(b *testing.B) {
	bs := "not empty"
	for n := 0; n < b.N; n++ {
		assert.NotEmpty(bs)
	}
}

func BenchmarkEmpty(b *testing.B) {
	bs := ""
	for n := 0; n < b.N; n++ {
		assert.Empty(bs)
	}
}

func BenchmarkLonger(b *testing.B) {
	bs := "tst"
	for n := 0; n < b.N; n++ {
		assert.Longer(bs, 2)
	}
}

func BenchmarkShorter(b *testing.B) {
	bs := "1"
	for n := 0; n < b.N; n++ {
		assert.Shorter(bs, 2)
	}
}

func BenchmarkSEmpty(b *testing.B) {
	bs := []int{}
	for n := 0; n < b.N; n++ {
		assert.SEmpty(bs)
	}
}

func BenchmarkSNotEmpty(b *testing.B) {
	bs := []byte{0}
	for n := 0; n < b.N; n++ {
		assert.SNotEmpty(bs)
	}
}

func BenchmarkSNotNil(b *testing.B) {
	bs := []byte{0}
	for n := 0; n < b.N; n++ {
		assert.SNotNil(bs)
	}
}

func BenchmarkMNotNil(b *testing.B) {
	var bs = map[int]int{0: 0}
	for n := 0; n < b.N; n++ {
		assert.MNotNil(bs)
	}
}

func BenchmarkCNotNil(b *testing.B) {
	var bs = make(chan int)
	for n := 0; n < b.N; n++ {
		assert.CNotNil(bs)
	}
}

func BenchmarkINotNil(b *testing.B) {
	var bs any = err2.ErrNotAccess
	for n := 0; n < b.N; n++ {
		assert.INotNil(bs)
	}
}

func BenchmarkINil(b *testing.B) {
	var bs any
	for n := 0; n < b.N; n++ {
		assert.INil(bs)
	}
}

func BenchmarkNil(b *testing.B) {
	var bs *int
	for n := 0; n < b.N; n++ {
		assert.Nil(bs)
	}
}

func BenchmarkNotNil(b *testing.B) {
	bs := new(int)
	for n := 0; n < b.N; n++ {
		assert.NotNil(bs)
	}
}

func BenchmarkSNil(b *testing.B) {
	var bs []int
	for n := 0; n < b.N; n++ {
		assert.SNil(bs)
	}
}

func BenchmarkMNil(b *testing.B) {
	var bs map[int]int
	for n := 0; n < b.N; n++ {
		assert.MNil(bs)
	}
}

func BenchmarkCNil(b *testing.B) {
	var bs chan int
	for n := 0; n < b.N; n++ {
		assert.CNil(bs)
	}
}

func BenchmarkThat(b *testing.B) {
	const four = 4
	for n := 0; n < b.N; n++ {
		assert.That(four == 2+2)
	}
}

func BenchmarkZero(b *testing.B) {
	const zero = 0
	for n := 0; n < b.N; n++ {
		assert.Zero(zero)
	}
}

func BenchmarkGreater(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assert.Greater(1, 0)
	}
}

func BenchmarkLess(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assert.Less(0, 1)
	}
}

func BenchmarkError(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assert.Error(err2.ErrNotAccess)
	}
}

func BenchmarkErrorIs(b *testing.B) {
	err := err2.ErrNotAccess
	for n := 0; n < b.N; n++ {
		assert.ErrorIs(err, err2.ErrNotAccess)
	}
}
func BenchmarkEqual(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assert.Equal(n, n)
	}
}

func BenchmarkAsserter_TrueIfVersion(b *testing.B) {
	ifPanicZero := func(i int) {
		if i == 0 {
			panic("i == 0")
		}
	}

	for n := 0; n < b.N; n++ {
		ifPanicZero(4)
	}
}

func BenchmarkMLen(b *testing.B) {
	d := map[byte]byte{1: 1, 2: 2}
	for n := 0; n < b.N; n++ {
		assert.MLen(d, 2)
	}
}

func BenchmarkMShorter(b *testing.B) {
	d := map[byte]byte{1: 1, 2: 2}
	for n := 0; n < b.N; n++ {
		assert.MShorter(d, 4)
	}
}

func BenchmarkMLonger(b *testing.B) {
	d := map[byte]byte{1: 1, 2: 2}
	for n := 0; n < b.N; n++ {
		assert.MLonger(d, 1)
	}
}

func BenchmarkSLen(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assert.SLen(d, 2)
	}
}

func BenchmarkSShorter(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assert.SShorter(d, 3)
	}
}

func BenchmarkSLonger(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assert.SLonger(d, 1)
	}
}

func BenchmarkCLen(b *testing.B) {
	d := make(chan int, 2)
	d <- int(1)
	d <- int(1)
	for n := 0; n < b.N; n++ {
		assert.CLen(d, 2)
	}
}

func BenchmarkCShorter(b *testing.B) {
	d := make(chan int, 2)
	d <- int(1)
	d <- int(1)
	for n := 0; n < b.N; n++ {
		assert.CShorter(d, 3)
	}
}

func BenchmarkCLonger(b *testing.B) {
	d := make(chan int, 2)
	d <- int(1)
	d <- int(1)
	for n := 0; n < b.N; n++ {
		assert.CLonger(d, 1)
	}
}

func BenchmarkLen(b *testing.B) {
	s := "len"
	for n := 0; n < b.N; n++ {
		assert.Len(s, 3)
	}
}

func BenchmarkSLen_thatVersion(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assert.That(len(d) == 2)
	}
}

func BenchmarkNotEqualInt(b *testing.B) {
	const d = 2
	for n := 0; n < b.N; n++ {
		assert.NotEqual(d, 3)
	}
}

func BenchmarkEqualInt(b *testing.B) {
	const d = 2
	for n := 0; n < b.N; n++ {
		assert.Equal(d, 2)
	}
}

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	assert.SetDefault(assert.Production)
}

func tearDown() {}

func TestPushAsserter(t *testing.T) {
	sentinel := errors.New("sentinel")

	functionReturningSentinel := func() (err error) {
		defer err2.Handle(&err, nil /* <- no error annotations */)

		assert.That(false, sentinel) // i.e., `sentinel` is returned
		return nil
	}

	defer assert.PushTester(t, assert.Production)() // Examples need Production
	defer assert.PushAsserter(assert.Plain)()       // We want sentinel

	err := functionReturningSentinel()
	assert.That(err == sentinel) // we can use ==; aren't forced to errors.Is
}
