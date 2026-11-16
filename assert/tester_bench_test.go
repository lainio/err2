package assert

import (
	"testing"

	"github.com/lainio/err2/internal/x"
)

func BenchmarkPushTesterRestore(b *testing.B) {
	// cleanup the noise from other tests to keep runs same wo/ other run vars
	testers = x.NewRWMap[testersMap]()

	old := SetDefault(TestFull)
	defer SetDefault(old)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		restore := PushTester(b)
		restore()
	}
}

func BenchmarkCurrentAsserter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = current()
	}
}
