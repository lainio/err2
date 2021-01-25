package assert_test

import (
	"fmt"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

func ExampleTrue() {
	// we want errors instead of immediate panics for examples
	assert.ProductionMode = true

	sample := func() (err error) {
		defer err2.Annotate("sample", &err)

		assert.True(false, "assertion test")
		return err
	}
	err := sample()
	fmt.Printf("%v", err)
	// Output: sample: assertion test
}

func ExampleLen() {
	// we want errors instead of immediate panics for examples
	assert.ProductionMode = true

	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.Len(b, 3)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: got 2, want 3
}

func ExampleEmpty() {
	// we want errors instead of immediate panics for examples
	assert.ProductionMode = true

	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.Empty(b)
		return err
	}
	err := sample([]byte{1, 2})
	fmt.Printf("%v", err)
	// Output: sample: got 2, want == 0
}

func ExampleNotNil() {
	// we want errors instead of immediate panics for examples
	assert.ProductionMode = true

	sample := func(b []byte) (err error) {
		defer err2.Annotate("sample", &err)

		assert.NotNil(b)
		return err
	}
	err := sample(nil)
	fmt.Printf("%v", err)
	// Output: sample: nil detected
}

func ExampleNoImplementation() {
	// we want errors instead of immediate panics for examples
	assert.ProductionMode = true

	sample := func(m int) (err error) {
		defer err2.Annotate("sample", &err)

		switch m {
		case 1:
			return nil
		default:
			assert.NoImplementation()
		}
		return err
	}
	err := sample(0)
	fmt.Printf("%v", err)
	// Output: sample: not implemented
}
