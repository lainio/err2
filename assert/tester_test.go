package assert

import (
	"fmt"
	"sync"
	"testing"

	"github.com/lainio/err2/internal/expect"
)

// Embedding TB retains the testing package's private methods while allowing
// separate contexts on the same test goroutine to have observable identities.
type namedTester struct {
	testing.TB
	name string
}

func (t *namedTester) Name() string { return t.name }

func testerDefault(t *testing.T) {
	t.Helper()
	old := SetDefault(TestFull)
	t.Cleanup(func() { SetDefault(old) })
}

func wantTester(t *testing.T, want testing.TB) {
	t.Helper()
	got := tester()
	expect.Thatf(t, got == want, "tester=%v, want %v", got, want)
	if want != nil {
		expect.Equal(t, got.Name(), want.Name())
		That(true)
	}
}

func TestPushTesterScopes(t *testing.T) {
	testerDefault(t)
	for _, depth := range []int{1, 2, 4} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			wantTester(t, nil)
			var scope func(int)
			scope = func(i int) {
				ctx := &namedTester{TB: t, name: fmt.Sprint(i)}
				defer PushTester(ctx)()
				wantTester(t, ctx)
				if i+1 < depth {
					scope(i + 1)
					wantTester(t, ctx)
				}
			}
			scope(0)
			wantTester(t, nil)
		})
	}
}

func TestPushTesterLegacyAndMixedScopes(t *testing.T) {
	testerDefault(t)
	outer := &namedTester{TB: t, name: "outer"}
	inner := &namedTester{TB: t, name: "inner"}
	wantTester(t, nil)
	// PopTester removes the registration; nested scopes use their restorers.
	func() {
		PushTester(outer)
		defer PopTester()
		wantTester(t, outer)
		func() {
			defer PushTester(inner)()
			wantTester(t, inner)
		}()
		wantTester(t, outer)
		restoreInner := PushTester(inner)
		wantTester(t, inner)
		restoreInner()
		wantTester(t, outer)
	}()
	wantTester(t, nil)
}

func TestPopTesterWithoutRegistration(t *testing.T) {
	wantTester(t, nil)
	PopTester()
	wantTester(t, nil)
}

func TestPushTesterConcurrentScopes(t *testing.T) {
	testerDefault(t)
	defer PushTester(t)()
	const workers = 24
	var ready, done sync.WaitGroup
	ready.Add(workers)
	done.Add(workers)
	release := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer done.Done()
			outer := &namedTester{TB: t, name: fmt.Sprintf("outer-%d", i)}
			inner := &namedTester{TB: t, name: fmt.Sprintf("inner-%d", i)}
			if tester() != nil {
				t.Error("new goroutine inherited a tester")
			}
			func() {
				defer PushTester(outer)()
				ready.Done()
				<-release
				for j := 0; j < 100; j++ {
					undo := PushTester(inner)
					if got := tester(); got != inner || got.Name() != inner.name {
						t.Error("wrong inner testing context")
					}
					undo()
					if got := tester(); got != outer || got.Name() != outer.name {
						t.Error("outer testing context was not restored")
					}
				}
			}()
			if tester() != nil {
				t.Error("completed scope retained a tester")
			}
		}(i)
	}
	ready.Wait()
	if tester() != t {
		t.Error("workers changed the parent testing context")
	}
	close(release)
	done.Wait()
	wantTester(t, t)
}

func TestPushTesterSubtests(t *testing.T) {
	testerDefault(t)
	defer PushTester(t)()
	t.Run("child", func(t *testing.T) {
		wantTester(t, nil)
		defer PushTester(t)()
		t.Run("grandchild", func(t *testing.T) {
			wantTester(t, nil)
			defer PushTester(t)()
			wantTester(t, t)
		})
		wantTester(t, t)
	})
	wantTester(t, t)
}

func TestPushTesterParallelSubtests(t *testing.T) {
	testerDefault(t)
	defer PushTester(t)()
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 8; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				wantTester(t, nil)
				func() {
					defer PushTester(t)()
					inner := &namedTester{TB: t, name: t.Name() + "-inner"}
					func() {
						defer PushTester(inner)()
						wantTester(t, inner)
					}()
					wantTester(t, t)
				}()
				wantTester(t, nil)
			})
		}
	})
	wantTester(t, t)
}

func TestPushTesterConcurrentInitialization(t *testing.T) {
	old := SetDefault(Production)
	defer SetDefault(old)
	var done sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 32; i++ {
		done.Add(1)
		go func() {
			defer done.Done()
			<-start
			restore := PushTester(t)
			defer restore()
			if !current().isUnitTesting() {
				t.Error("first registration did not initialize test assertions")
			}
		}()
	}
	close(start)
	done.Wait()
	expect.Equal(t, current(), testFull)
}

func TestTesterConcurrentDefaultConfiguration(t *testing.T) {
	testerDefault(t)
	var done sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 12; i++ {
		done.Add(1)
		go func(i int) {
			defer done.Done()
			<-start
			for j := 0; j < 200; j++ {
				switch i % 3 {
				case 0:
					SetDefault(Asserter(j % len(defAsserter)))
				case 1:
					_ = current()
					_ = asserterFlag.String()
					_ = asserterFlag.Get()
					restore := PushAsserter(Plain)
					if current() != plain {
						t.Error("global configuration replaced a local asserter")
					}
					restore()
				case 2:
					restore := PushTester(t)
					if tester() != t {
						t.Error("configuration changed the active tester")
					}
					restore()
				}
			}
		}(i)
	}
	close(start)
	done.Wait()
}

func TestPushTesterDefaultSemantics(t *testing.T) {
	testerDefault(t)
	expect.Equal(t, SetDefault(Production), TestFull)
	func() {
		defer PushTester(t, Test)()
		expect.Equal(t, current(), test)
	}()
	expect.Equal(t, current(), test) // Restoration does not undo global settings.
	global := make(chan asserter, 1)
	go func() { global <- current() }()
	expect.Equal(t, <-global, test)

	func() {
		defer PushAsserter(Plain)()
		defer PushTester(t)()
		expect.Equal(t, current(), plain)
		expect.Equal(t, defaultAsserterString(), "TestFull")
	}()
	expect.Equal(t, current(), testFull)
	wantTester(t, nil)
}
