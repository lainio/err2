package assert_test

import (
	"fmt"
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

func ifPanicZero(i int) {
	if i == 0 {
		panic("i == 0")
	}
}

func assertZero(i int) {
	assert.D.True(i != 0)
}

func assertLen(b []byte) {
	assert.D.Len(b, 2)
}

func assertEqualInt(b []byte) {
	assert.D.EqualInt(len(b), 2)
}

func BenchmarkAsserter_True(b *testing.B) {
	for n := 0; n < b.N; n++ {
		assertZero(4)
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

func BenchmarkAsserter_EqualInt(b *testing.B) {
	d := []byte{1, 2}
	for n := 0; n < b.N; n++ {
		assertEqualInt(d)
	}
}
