package x

import "testing"

var (
	lengths = []int{2, 16, 128, 1024, 8192, 65536, 524288, 4194304, 16777216, 134217728}
)

func BenchmarkSSReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SReverse(lengths)
	}
}

func BenchmarkSReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SReverse(lengths)
	}
}

func BenchmarkSReverseClone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SReverseClone(lengths)
	}
}

func BenchmarkXSReverseClone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SReverseClone(lengths)
	}
}
