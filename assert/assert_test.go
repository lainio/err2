package assert_test // Note!! Some tests here are related to line # of the file

import (
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
	// Output: testing: run example: assert_test.go:16: ExampleThat.func1(): assertion failure: optional message
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
	// Output: sample: assert_test.go:29: ExampleNotNil.func1(): assertion failure: pointer should not be nil
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
	// Output: sample: assert_test.go:44: ExampleMNotNil.func1(): assertion failure: map should not be nil
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
	// Output: sample: assert_test.go:58: ExampleCNotNil.func1(): assertion failure: channel should not be nil
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
	// Output: sample: assert_test.go:73: ExampleSNotNil.func1(): assertion failure: slice should not be nil
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
	// Output: sample: assert_test.go:87: ExampleEqual.func1(): assertion failure: equal: got '2', want '1'
}

func ExampleSLen() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SLen(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:99: ExampleSLen.func1(): assertion failure: length: got '2', want '3'
}

func ExampleSNotEmpty() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SNotEmpty(b)
		return err
	}
	err := sample([]byte{})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:111: ExampleSNotEmpty.func1(): assertion failure: slice should not be empty
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
	// Output: sample: assert_test.go:124: ExampleNotEmpty.func1(): assertion failure: string should not be empty
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
	// Output: sample: assert_test.go:141: ExampleMKeyExists.func1(): assertion failure: key '2' doesn't exist
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
	// Output: sample: assert_test.go:153: ExampleZero.func1(): assertion failure: got '1', want (== '0')
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
	// Output: sample: assert_test.go:167: ExampleSLonger.func1(): assertion failure: got '1', should be longer than '1'
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
	// Output: sample: assert_test.go:180: ExampleMShorter.func1(): assertion failure: got '1', should be shorter than '1'
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
	// Output: sample: assert_test.go:194: ExampleSShorter.func1(): assertion failure: got '1', should be shorter than '0': optional message (test_str)
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
	// Output: sample: assert_test.go:208: ExampleLess.func1(): assertion failure: got '1', want >= '1'
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
	// Output: sample: assert_test.go:223: ExampleGreater.func1(): assertion failure: got '2', want <= '2'
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
	// Output: sample: assert_test.go:236: ExampleNotZero.func1(): assertion failure: got '0', want (!= 0)
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
	// Output: sample: assert_test.go:251: ExampleMLen.func1(): assertion failure: length: got '2', want '3'
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
	// Output: sample: assert_test.go:265: ExampleCLen.func1(): assertion failure: length: got '2', want '3'
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
	// Output: sample: assert_test.go:297: ExampleINotNil.func1(): assertion failure: interface should be nil
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
	// Output: sample: assert_test.go:312: ExampleLen.func1(): assertion failure: length: got '2', want '3'
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
	// Output: sample: assert_test.go:326: ExampleDeepEqual.func1(): assertion failure: got '2', want '3'
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
	// Output: sample: assert_test.go:339: ExampleError.func1(): assertion failure: test
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
	// Output: sample: assert_test.go:352: ExampleNotImplemented.func1(): assertion failure: not implemented
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
