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
	// Output: testing: run example: assert_test.go:16: ExampleThat.func1(): assertion violation: optional message
}

func ExampleNotNil() {
	sample := func(b *byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.NotNil(b)
		return err
	}
	var b *byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:28: ExampleNotNil.func1(): assertion violation: pointer shouldn't be nil
}

func ExampleMNotNil() {
	sample := func(b map[string]byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.MNotNil(b)
		return err
	}
	var b map[string]byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:41: ExampleMNotNil.func1(): assertion violation: map shouldn't be nil
}

func ExampleCNotNil() {
	sample := func(c chan byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.CNotNil(c)
		return err
	}
	var c chan byte
	err := sample(c)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:54: ExampleCNotNil.func1(): assertion violation: channel shouldn't be nil
}

func ExampleSNotNil() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SNotNil(b)
		return err
	}
	var b []byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:67: ExampleSNotNil.func1(): assertion violation: slice shouldn't be nil
}

func ExampleEqual() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Equal(len(b), 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:80: ExampleEqual.func1(): assertion violation: got '2', want '3'
}

func ExampleSLen() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SLen(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:92: ExampleSLen.func1(): assertion violation: got '2', want '3'
}

func ExampleSNotEmpty() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SNotEmpty(b)
		return err
	}
	err := sample([]byte{})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:104: ExampleSNotEmpty.func1(): assertion violation: slice shouldn't be empty
}

func ExampleNotEmpty() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		assert.Empty(b)
		assert.NotEmpty(b)
		return err
	}
	err := sample("")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:117: ExampleNotEmpty.func1(): assertion violation: string shouldn't be empty
}

func ExampleMKeyExists() {
	sample := func(b string) (err error) {
		defer err2.Handle(&err, "sample")

		m := map[string]string{
			"1": "one",
		}
		v := assert.MKeyExists(m, "1")
		assert.Equal(v, "one")
		_ = assert.MKeyExists(m, b)
		return err
	}
	err := sample("2")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:134: ExampleMKeyExists.func1(): assertion violation: key '2' doesn't exist
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
	// Output: sample: assert_test.go:146: ExampleZero.func1(): assertion violation: got '1', want (== '0')
}

func ExampleSLonger() {
	sample := func(b []byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.SLonger(b, 0) // ok
		assert.SLonger(b, 1) // not ok
		return err
	}
	err := sample([]byte{01}) // len = 1
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:160: ExampleSLonger.func1(): assertion violation: got '1', should be longer than '1'
}

func ExampleMShorter() {
	sample := func(b map[byte]byte) (err error) {
		defer err2.Handle(&err, "sample")

		assert.MShorter(b, 1) // ok
		assert.MShorter(b, 0) // not ok
		return err
	}
	err := sample(map[byte]byte{01: 01}) // len = 1
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:172: ExampleMShorter.func1(): assertion violation: got '1', should be shorter than '1'
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
	// Output: sample: assert_test.go:186: ExampleSShorter.func1(): assertion violation: got '1', should be shorter than '0': optional message (test_str)
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
	// Output: sample: assert_test.go:200: ExampleLess.func1(): assertion violation: got '1', want >= '1'
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
	// Output: sample: assert_test.go:215: ExampleGreater.func1(): assertion violation: got '2', want <= '2'
}

func assertZero(i int) {
	assert.Zero(i)
}

func assertZeroGen(i int) {
	assert.Equal(i, 0)
}

func assertMLen(b map[byte]byte, l int) {
	assert.MLen(b, l)
}

func assertEqualInt2(b int) {
	assert.Equal(b, 2)
}

func BenchmarkSNotNil(b *testing.B) {
	bs := []byte{0}
	for n := 0; n < b.N; n++ {
		assert.SNotNil(bs)
	}
}

func BenchmarkNotNil(b *testing.B) {
	bs := new(int)
	for n := 0; n < b.N; n++ {
		assert.NotNil(bs)
	}
}

func BenchmarkThat(b *testing.B) {
	const four = 4
	for n := 0; n < b.N; n++ {
		assert.That(four == 2+2)
	}
}

func BenchmarkNotEmpty(b *testing.B) {
	str := "test"
	for n := 0; n < b.N; n++ {
		assert.NotEmpty(str)
	}
}

func BenchmarkZero(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assertZero(0)
	}
}

func BenchmarkEqual(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assertZeroGen(0)
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
		assertMLen(d, 2)
	}
}

func BenchmarkSLen(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assert.SLen(d, 2)
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

func BenchmarkEqualInt(b *testing.B) {
	const d = 2
	for n := 0; n < b.N; n++ {
		assertEqualInt2(d)
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
