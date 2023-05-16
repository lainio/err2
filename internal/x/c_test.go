package x

import (
	"reflect"
	"testing"

	"github.com/lainio/err2/internal/test"
)

var (
	original       = []int{2, 16, 128, 1024, 8192, 65536, 524288, 4194304, 16777216, 134217728}
	lengths        = []int{2, 16, 128, 1024, 8192, 65536, 524288, 4194304, 16777216, 134217728}
	reverseLengths = []int{134217728, 16777216, 4194304, 524288, 65536, 8192, 1024, 128, 16, 2}
)

func TestSReverse(t *testing.T) {
	SReverse(lengths)
	test.Require(t, reflect.DeepEqual(lengths, reverseLengths))
	SReverse(lengths) // it's reverse now turn it to original
	test.Require(t, reflect.DeepEqual(lengths, original))
}

func BenchmarkSSReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SSReverse(lengths)
	}
}

func BenchmarkSReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SReverse(lengths)
	}
}
