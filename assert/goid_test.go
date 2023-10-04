package assert

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGoid(t *testing.T) {
	t.Parallel()
	stackBytes := []byte(`goroutine 518 [running]:
`)

	id := myByteToInt(stackBytes[10:])
	if id != 518 {
		t.Fail()
	}
}

func Test_oldGoid(t *testing.T) {
	t.Parallel()
	stackBytes := []byte(`goroutine 518 [running]:
`)

	id := oldGoid(stackBytes)
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

func oldGoid(buf []byte) (id int) {
	_, err := fmt.Fscanf(bytes.NewReader(buf), "goroutine %d", &id)
	if err != nil {
		panic("cannot get goroutine id: " + err.Error())
	}
	return id
}

func BenchmarkGoid_Old(b *testing.B) {
	stackBytes := []byte(`goroutine 518 [running]:
`)

	for n := 0; n < b.N; n++ {
		_ = oldGoid(stackBytes)
	}
}
