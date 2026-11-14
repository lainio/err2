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

func caughtPanic(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func TestPushTesterScopes(t *testing.T) {
	testerDefault(t)
	for _, depth := range []int{1, 2, 4} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			wantTester(t, nil)
			contexts := make([]testing.TB, 0, depth)
			restores := make([]Function, 0, depth)
			for i := 0; i < depth; i++ {
				ctx := &namedTester{TB: t, name: fmt.Sprint(i)}
				restore := PushTester(ctx)
				defer restore()
				contexts = append(contexts, ctx)
				restores = append(restores, restore)
				wantTester(t, ctx)
			}
			for i := depth - 1; i >= 0; i-- {
				restores[i]()
				if i > 0 {
					wantTester(t, contexts[i-1])
				} else {
					wantTester(t, nil)
				}
			}
		})
	}
}

func TestPushTesterLegacyAndMixedScopes(t *testing.T) {
	testerDefault(t)
	outer := &namedTester{TB: t, name: "outer"}
	inner := &namedTester{TB: t, name: "inner"}
	restoreOuter := PushTester(outer)
	defer restoreOuter()
	restoreInner := PushTester(inner)
	PopTester()
	wantTester(t, outer)
	restoreInner() // The legacy pop already consumed this registration.
	wantTester(t, outer)

	func() {
		PushTester(inner)
		defer PopTester()
		wantTester(t, inner)
	}()
	wantTester(t, outer)
	restoreInner = PushTester(inner)
	restoreInner()
	wantTester(t, outer)
	PopTester()
	wantTester(t, nil)
	restoreOuter()
	PopTester() // Popping an empty context remains harmless.
	wantTester(t, nil)

	restoreNew := PushTester(inner)
	defer restoreNew()
	restoreOuter() // A retained closure must not remove a new registration.
	wantTester(t, inner)
}

func TestPushTesterOutOfOrder(t *testing.T) {
	testerDefault(t)
	outer := &namedTester{TB: t, name: "outer"}
	inner := &namedTester{TB: t, name: "inner"}
	restoreOuter := PushTester(outer)
	defer restoreOuter()
	restoreInner := PushTester(inner)
	defer restoreInner()
	r := caughtPanic(restoreOuter)
	expect.Thatf(t, r == "assert: tester scopes must be restored in LIFO order", "unexpected panic: %v", r)
	wantTester(t, inner)
	r = caughtPanic(func() {
		defer restoreOuter()
		panic("out-of-order unwinding")
	})
	expect.Thatf(t, r == "assert: tester scopes must be restored in LIFO order", "unexpected panic: %v", r)
	wantTester(t, inner)
	restoreInner()
	wantTester(t, outer)
	restoreOuter()
	wantTester(t, nil)
	restoreInner()
	wantTester(t, nil)
}

func TestPushTesterWrongGoroutine(t *testing.T) {
	testerDefault(t)
	restore := PushTester(t)
	defer restore()
	result := make(chan any, 1)
	go func() { result <- caughtPanic(restore) }()
	r := <-result
	expect.Thatf(t, r == "assert: tester scope restored by a different goroutine", "unexpected panic: %v", r)
	wantTester(t, t)
	restore()
	wantTester(t, nil)
}

func TestPushTesterConsumedClosureDoesNotRecover(t *testing.T) {
	testerDefault(t)
	restore := PushTester(t)
	restore()
	r := caughtPanic(func() {
		defer restore()
		panic("still propagates")
	})
	expect.Thatf(t, r == "still propagates", "unexpected panic: %v", r)
	wantTester(t, nil)
	undo := PushTester(t)
	defer undo()
	r = caughtPanic(func() {
		defer restore()
		panic("still propagates")
	})
	expect.Thatf(t, r == "still propagates", "unexpected panic: %v", r)
	wantTester(t, t)
}

func TestPushTesterReleasesReferences(t *testing.T) {
	testerDefault(t)
	restoreOuter := PushTester(t)
	defer restoreOuter()
	outer := testers.Get(goid())
	restoreInner := PushTester(&namedTester{TB: t, name: "inner"})
	defer restoreInner()
	inner := testers.Get(goid())
	restoreInner()
	wantTester(t, t)
	expect.That(t, inner.closed && inner.t == nil && inner.previous == nil)
	PopTester()
	wantTester(t, nil)
	expect.That(t, outer.closed && outer.t == nil && outer.previous == nil)
	expect.That(t, testers.Get(goid()) == nil)
	// Keep the closures alive through verification: they must not retain TBs.
	restoreInner()
	restoreOuter()
	wantTester(t, nil)
}

func TestPopTesterWithoutRegistration(t *testing.T) {
	wantTester(t, nil)
	r := caughtPanic(func() {
		defer PopTester()
		panic("no registered tester")
	})
	expect.Thatf(t, r == "no registered tester", "unexpected panic: %v", r)
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
			restore := PushTester(outer)
			defer restore()
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
			restore()
			if tester() != nil || testers.Get(goid()) != nil {
				t.Error("completed scope retained a tester")
			}
		}(i)
	}
	ready.Wait()
	wantTester(t, t)
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
				restore := PushTester(t)
				defer restore()
				inner := &namedTester{TB: t, name: t.Name() + "-inner"}
				undo := PushTester(inner)
				wantTester(t, inner)
				undo()
				wantTester(t, t)
				restore()
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
	restore := PushTester(t, Test)
	expect.Equal(t, current(), test)
	restore()
	expect.Equal(t, current(), test) // Restoration does not undo global settings.
	global := make(chan asserter, 1)
	go func() { global <- current() }()
	expect.Equal(t, <-global, test)

	undo := PushAsserter(Plain)
	defer undo()
	restore = PushTester(t)
	defer restore()
	expect.Equal(t, current(), plain)
	expect.Equal(t, defaultAsserterString(), "TestFull")
	undo()
	expect.Equal(t, current(), testFull)
}
