package assert

import "testing"

func BenchmarkPushTesterRestore(b *testing.B) {
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
