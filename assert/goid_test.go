package assert

import (
	"testing"
)

func TestGoid(t *testing.T) {
	stackBytes := []byte(`goroutine 518 [running]:
`)

	id := myByteToInt(stackBytes[10:])
	if id != 518 {
		t.Fail()
	}
}

func BenchmarkGoid(b *testing.B) {
	_ = []byte(`goroutine 518 [running]:
`)

	for n := 0; n < b.N; n++ {
		_ = goid()
	}
}

func BenchmarkGoid_MyByteToInt(b *testing.B) {
	stackBytes := []byte(`goroutine 518 [running]:
`)

	for n := 0; n < b.N; n++ {
		_ = myByteToInt(stackBytes[10:])
	}
}
