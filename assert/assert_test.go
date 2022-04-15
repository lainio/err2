package assert_test // Note!! Some tests here are related to line # of the file

import (
	"fmt"
	"os"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

func ExampleAsserter_True() {
	sample := func() (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.True(false, "assertion test")
		return err
	}
	err := sample()
	fmt.Printf("%v", err)
	// Output: sample: assertion test
}

func ExampleAsserter_Truef() {
	sample := func() (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.Truef(false, "assertion test %d", 2)
		return err
	}
	err := sample()
	fmt.Printf("%v", err)
	// Output: sample: assertion test 2
}

func ExampleAsserter_Len() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.Len(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: got 2, want 3
}

func ExampleAsserter_EqualInt() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.EqualInt(len(b), 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: got 2, want 3
}

func ExampleNotNil() {
	sample := func(b *byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.NotNil(b)
		return err
	}
	var b *byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:64 ExampleNotNil.func1 pointer is nil
}

func ExampleMNotNil() {
	sample := func(b map[string]byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.MNotNil(b)
		return err
	}
	var b map[string]byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:77 ExampleMNotNil.func1 map is nil
}

func ExampleCNotNil() {
	sample := func(c chan byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.CNotNil(c)
		return err
	}
	var c chan byte
	err := sample(c)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:90 ExampleCNotNil.func1 channel is nil
}

func ExampleSNotNil() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.SNotNil(b)
		return err
	}
	var b []byte
	err := sample(b)
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:103 ExampleSNotNil.func1 slice is nil
}

func ExampleEqual() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.Equal(len(b), 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:116 ExampleEqual.func1 got 2, want 3
}

func ExampleSLen() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.SLen(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:128 ExampleSLen.func1 got 2, want 3
}

func ExampleAsserter_Lenf() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.Lenf(b, 3, "actual len = %d", len(b))
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: actual len = 2
}

func ExampleAsserter_Empty() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.P.Empty(b)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: got 2, want == 0
}

func ExampleAsserter_NoImplementation() {
	sample := func(m int) (err error) {
		defer err2.Annotate("sample", &err)

		switch m {
		case 1:
			return nil
		default:
			assert.P.NoImplementation()
		}
		return err
	}
	err := sample(0)
	fmt.Printf("%v", err)
	// Output: sample: not implemented
}

func ExampleSNotEmpty() {
	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.SNotEmpty(b)
		return err
	}
	err := sample([]byte{})
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:181 ExampleSNotEmpty.func1 slice shouldn't be empty
}

func ExampleNotEmpty() {
	sample := func(b string) (err error) {
		defer err2.Annotate("sample", &err)

		assert.NotEmpty(b)
		return err
	}
	err := sample("")
	fmt.Printf("%v", err)
	// Output: sample: assert_test.go:193 ExampleNotEmpty.func1 string shouldn't be empty
}

func ifPanicZero(i int) {
	if i == 0 {
		panic("i == 0")
	}
}

func assertThat(term bool) {
	assert.That(term)
}

func assertZero(i int) {
	assert.D.True(i != 0)
}

func assertZeroGen(i int) {
	assert.Equal(i, 0)
}

func assertLen(b []byte) {
	assert.D.Len(b, 2)
}

func assertLenf(b []byte, l int) {
	assert.D.Lenf(b, l, "")
}

func assertSLen(b []byte, l int) {
	assert.SLen(b, l)
}

func assertMLen(b map[byte]byte, l int) {
	assert.MLen(b, l)
}

func assertEqualInt(b []byte) {
	assert.D.EqualInt(len(b), 2)
}

func assertEqualInt2(b int) {
	assert.Equal(b, 2)
}

func BenchmarkThat(b *testing.B) {
	const four = 4
	for n := 0; n < b.N; n++ {
		assertThat(2+2 == four)
	}
}

func BenchmarkAsserter_True(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assertZero(4)
	}
}

func BenchmarkEqual(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assertZeroGen(0)
	}
}

func BenchmarkAsserter_TrueIfVersion(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ifPanicZero(4)
	}
}

func BenchmarkAsserter_Len(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assertLen(d)
	}
}

func BenchmarkAsserter_Lenf(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assertLenf(d, 2)
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
		assertSLen(d, 2)
	}
}

func BenchmarkAsserter_EqualInt(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assertEqualInt(d)
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
	assert.DefaultAsserter = assert.AsserterToError | assert.AsserterCallerInfo
}

func tearDown() {}
