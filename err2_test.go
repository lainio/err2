package err2_test

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/require"
	"github.com/lainio/err2/try"
)

const errStringInThrow = "this is an ERROR"

func throw() (string, error) {
	return "", fmt.Errorf(errStringInThrow)
}

func twoStrNoThrow() (string, string, error)        { return "test", "test", nil }
func intStrNoThrow() (int, string, error)           { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }
func noThrow() (string, error)                      { return "test", nil }

func noErr() error {
	return nil
}

func TestTry_noError(t *testing.T) {
	t.Parallel()
	try.To1(noThrow())
	try.To2(twoStrNoThrow())
	try.To2(intStrNoThrow())
	try.To3(boolIntStrNoThrow())
}

func TestDefault_Error(t *testing.T) {
	t.Parallel()
	var err error
	defer err2.Handle(&err)

	try.To1(throw())

	t.Fail() // If everything works we are never here
}

func TestTry_Error(t *testing.T) {
	t.Parallel()
	var err error
	defer err2.Handle(&err, func(err error) error { return err })

	try.To1(throw())

	t.Fail() // If everything works we are never here
}

func TestHandle_noerrHandler(t *testing.T) {
	t.Parallel()
	t.Run("noerr handler in ONLY one and NO error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.That(t, handlerCalled)
		}()
		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			handlerCalled = noerr
		})

		try.To(noErr())
	})

	t.Run("noerr handler is the last and NO error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.That(t, handlerCalled)
		}()
		defer err2.Handle(&err, func(err error) error {
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false
			require.That(t, false)
			return err
		})

		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			handlerCalled = noerr
		})

		try.To(noErr())
	})

	t.Run("noerr handler is the last and error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.ThatNot(t, handlerCalled)
		}()

		// This is the handler we are testing!
		defer err2.Handle(&err, func(err error) error {
			require.ThatNot(t, handlerCalled)
			handlerCalled = false
			require.That(t, true, "error should be handled")
			return err
		})

		// This is the handler we are testing! AND it's not called in error.
		defer err2.Handle(&err, func(bool) {
			require.That(t, false, "when error this is not called")
		})

		try.To1(throw())
	})

	t.Run("noerr is first and error happens with many handlers", func(t *testing.T) {
		t.Parallel()
		var (
			err               error
			finalAnnotatedErr = fmt.Errorf("err: %v", errStringInThrow)
			handlerCalled     bool
			callCount         int
		)
		defer func() {
			require.ThatNot(t, handlerCalled)
			require.Equal(t, callCount, 2)
			require.Equal(t, err.Error(), finalAnnotatedErr.Error())
		}()

		// This is the handler we are testing! AND it's not called in error.
		defer err2.Handle(&err, func(noerr bool) {
			require.That(t, false, "if error occurs/reset, this cannot happen")
			handlerCalled = noerr
		})

		// important! test that our handler doesn't change the current error
		// and it's not nil
		defer err2.Handle(&err, func(er error) error {
			require.That(t, er != nil, "er val: ", er, err)
			require.Equal(t, callCount, 1, "this is called in sencond")
			callCount++
			return er
		})

		defer err2.Handle(&err, func(err error) error {
			// this should not be called, so lets try to fuckup things...
			require.Equal(t, callCount, 0, "this is called in first")
			callCount++
			handlerCalled = false
			require.That(t, err != nil)
			return finalAnnotatedErr
		})
		try.To1(throw())
	})

	t.Run("noerr handler is first and NO error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.That(t, handlerCalled)
		}()

		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			require.That(t, noerr)
			handlerCalled = noerr
		})

		defer err2.Handle(&err, func(err error) error {
			require.That(t, false, "no error to handle!")
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false // see first deferred function
			return err
		})
		try.To(noErr())
	})

	t.Run("noerr handler is first of MANY and NO error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.That(t, handlerCalled)
		}()

		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			require.That(t, true)
			require.That(t, noerr)
			handlerCalled = noerr
		})

		defer err2.Handle(&err)

		defer err2.Handle(&err, func(err error) error {
			require.That(t, false, "no error to handle!")
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false // see first deferred function
			return err
		})

		defer err2.Handle(&err, func(err error) error {
			require.That(t, false, "no error to handle!")
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false // see first deferred function
			return err
		})
		try.To(noErr())
	})

	t.Run("noerr handler is first of MANY and error happens UNTIL RESET", func(t *testing.T) {
		t.Parallel()
		var err error
		var noerrHandlerCalled, errHandlerCalled bool
		defer func() {
			require.That(t, noerrHandlerCalled)
			require.That(t, errHandlerCalled)
		}()

		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			require.That(t, true) // we are here, for debugging
			require.That(t, noerr)
			noerrHandlerCalled = noerr
		})

		// this is the err handler that -- RESETS -- the error to nil
		defer err2.Handle(&err, func(err error) error {
			require.That(t, err != nil) // helps fast debugging

			// this should not be called, so lets try to fuckup things...
			noerrHandlerCalled = false // see first deferred function
			// keep the track that we have been here
			errHandlerCalled = true // see first deferred function
			return nil
		})

		defer err2.Handle(&err, func(err error) error {
			require.That(t, err != nil) // helps fast debugging
			// this should not be called, so lets try to fuckup things...
			noerrHandlerCalled = false // see first deferred function

			errHandlerCalled = true // see first deferred function
			return err
		})
		try.To1(throw())
	})

	t.Run("noerr handler is middle of MANY and NO error happens", func(t *testing.T) {
		t.Parallel()
		var err error
		var handlerCalled bool
		defer func() {
			require.That(t, handlerCalled)
		}()

		defer err2.Handle(&err)
		defer err2.Handle(&err)

		defer err2.Handle(&err, func(err error) error {
			require.That(t, false, "no error to handle!")
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false // see first deferred function
			return err
		})

		// This is the handler we are testing!
		defer err2.Handle(&err, func(noerr bool) {
			require.That(t, true, "this must be called")
			require.That(t, noerr)
			handlerCalled = noerr
		})

		defer err2.Handle(&err, func(err error) error {
			require.That(t, false, "no error to handle!")
			// this should not be called, so lets try to fuckup things...
			handlerCalled = false // see first deferred function
			return err
		})
		try.To(noErr())
	})
}

func TestPanickingCatchAll(t *testing.T) {
	t.Parallel()
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer err2.Catch(
						err2.Noop,
						func(any) {},
					)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.Catch(
						err2.Err(func(error) {}), // Using simplifier
						func(any) {},
					)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
		{"stop panic with empty catch",
			args{
				func() {
					defer err2.Catch()
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
		{"stop panic with error handler in catch",
			args{
				func() {
					defer err2.Catch(err2.Noop)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				require.That(t, recover() == nil, "panics should NOT carry on")
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCarryOn_Handle(t *testing.T) {
	t.Parallel()
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					var err error
					defer err2.Handle(&err, err2.Noop)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					var err error
					defer err2.Handle(&err, err2.Noop)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				require.That(t, recover() != nil, "panics should went thru when not our errors")
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Handle(t *testing.T) {
	t.Parallel()
	type args struct {
		f func() (err error)
	}
	myErr := fmt.Errorf("my error")
	annErr := fmt.Errorf("annotated: %w", myErr)

	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general error thru panic with annotion handler",
			args{
				func() (err error) {
					// If we want keep same error value second argument
					// must be nil
					defer err2.Handle(&err, func(err error) error {
						return fmt.Errorf("annotated: %w", err)
					})

					try.To(myErr)
					return nil
				},
			},
			annErr,
		},
		{"general error thru panic: handle nil: no automatic",
			args{
				func() (err error) {
					// If we want keep same error value second argument
					// must be nil
					defer err2.Handle(&err, nil)

					try.To(myErr)
					return nil
				},
			},
			myErr,
		},
		{"general panic",
			args{
				func() (err error) {
					defer err2.Handle(&err)
					panic("panic")
				},
			},
			nil,
		},
		{"general panic plus err handler",
			args{
				func() (err error) {
					defer err2.Handle(&err, err2.Noop)
					panic("panic")
				},
			},
			nil,
		},
		{"general panic stoped with handler plus err handler",
			args{
				func() (err error) {
					defer err2.Handle(&err,
						func(err error) error { return err },
						func(any) {},
					)
					panic("panic")
				},
			},
			myErr,
		},
		{"general panic stoped with handler",
			args{
				func() (err error) {
					defer err2.Handle(&err, func(any) {})
					panic("panic")
				},
			},
			myErr,
		},
		{"general panic stoped with handler plus fmt string",
			args{
				func() (err error) {
					defer err2.Handle(&err, func(any) {}, "string")
					panic("panic")
				},
			},
			myErr,
		},
		{"runtime.error panic",
			args{
				func() (err error) {
					defer err2.Handle(&err)
					var b []byte
					b[0] = 0
					return nil
				},
			},
			nil,
		},
		{"runtime.error panic stopped with handler",
			args{
				func() (err error) {
					defer err2.Handle(&err, func(any) {})
					var b []byte
					b[0] = 0
					return nil
				},
			},
			myErr,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				r := recover()
				if tt.wants == nil {
					require.That(t, r != nil, "wants err, then panic")
				}
			}()
			err := tt.args.f()
			if err != nil {
				require.Equal(t, err.Error(), tt.wants.Error())
			}
		})
	}
}

func TestPanicking_Catch(t *testing.T) {
	t.Parallel()
	type args struct {
		f func()
	}
	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general panic",
			args{
				func() {
					defer err2.Catch(err2.Noop)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.Catch(err2.Noop)
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				require.That(t, recover() == nil, "panics should NOT carry on")
			}()
			tt.args.f()
		})
	}
}

func TestCatch_Error(t *testing.T) {
	t.Parallel()
	defer err2.Catch()

	try.To1(throw())

	t.Fail() // If everything works we are never here
}

func Test_TryOutError(t *testing.T) {
	t.Parallel()
	defer err2.Catch(func(err error) error {
		require.Equal(t, err.Error(), "fails: test: this is an ERROR",
			"=> we should catch right error str here")
		return err
	})

	var retVal string

	// let's test try.Out1() and it's throw capabilities here, even try.To1()
	// is the preferred way.
	retVal = try.Out1(noThrow()).Handle().Val1
	require.Equal(t, retVal, "test", "if no error happens, we get value")

	_ = try.Out1(throw()).Handle("fails: %v", retVal).Val1
	t.Fail() // If everything works in Handle we are never here.
}

func TestCatch_Panic(t *testing.T) {
	t.Parallel()
	panicHandled := false
	defer func() {
		// when err2.Catch's panic handler works fine, panic is handled
		if !panicHandled {
			t.Fail()
		}
	}()

	defer err2.Catch(
		func(error) error {
			t.Log("it was panic, not an error")
			t.Fail() // we should not be here
			return nil
		},
		func(any) {
			panicHandled = true
		})

	panic("test panic")
}

func TestSetErrorTracer(t *testing.T) {
	t.Parallel()
	w := err2.ErrorTracer()
	require.That(t, w == nil, "error tracer should be nil")
	var w1 io.Writer
	err2.SetErrorTracer(w1)
	w = err2.ErrorTracer()
	require.That(t, w == nil, "error tracer should be nil")
}

func ExampleCatch_withFmt() {
	// Set default logger to stdout for this example
	oldLogW := err2.LogTracer()
	err2.SetLogTracer(os.Stdout)
	defer err2.SetLogTracer(oldLogW)

	transport := func() {
		// See how Catch follows given format string similarly as Handle
		defer err2.Catch("catch")
		err2.Throwf("our error")
	}
	transport()
	// Output: catch: our error
}

func ExampleHandle() {
	var err error
	defer err2.Handle(&err)
	try.To1(noThrow())
	// Output:
}

func ExampleHandle_errThrow() {
	transport := func() (err error) {
		defer err2.Handle(&err)
		err2.Throwf("our error")
		return nil
	}
	err := transport()
	fmt.Printf("%v", err)
	// Output: testing: run example: our error
}

func ExampleHandle_annotatedErrReturn() {
	normalReturn := func() (err error) {
		defer err2.Handle(&err) // automatic annotation
		return fmt.Errorf("our error")
	}
	err := normalReturn()
	fmt.Printf("%v", err)

	// ------- func name comes from Go example/test harness
	// ------- v ------------------ v --------
	// Output: testing: run example: our error
}

func ExampleHandle_errReturn() {
	normalReturn := func() (err error) {
		defer err2.Handle(&err, nil) // nil disables automatic annotation
		return fmt.Errorf("our error")
	}
	err := normalReturn()
	fmt.Printf("%v", err)
	// Output: our error
}

func ExampleHandle_empty() {
	annotated := func() (err error) {
		defer err2.Handle(&err, "annotated")
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: this is an ERROR
}

func ExampleHandle_annotate() {
	annotated := func() (err error) {
		defer err2.Handle(&err, "annotated: %s", "err2")
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: err2: this is an ERROR
}

func ExampleThrowf() {
	type fn func(v int) int
	var recursion fn
	const recursionLimit = 77 // 12+11+10+9+8+7+6+5+4+3+2+1 = 78

	recursion = func(i int) int {
		if i > recursionLimit { // simulated error case
			err2.Throwf("helper failed at: %d", i)
		} else if i == 0 {
			return 0 // recursion without error ends here
		}
		return i + recursion(i-1)
	}

	annotated := func() (err error) {
		defer err2.Handle(&err, "annotated: %s", "err2")

		r := recursion(12) // call recursive algorithm successfully
		recursion(r)       // call recursive algorithm unsuccessfully
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: err2: helper failed at: 78
}

func ExampleHandle_deferStack() {
	annotated := func() (err error) {
		defer err2.Handle(&err, "annotated 2nd")
		defer err2.Handle(&err, "annotated 1st")
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated 2nd: annotated 1st: this is an ERROR
}

func ExampleHandle_handlerFn() {
	doSomething := func(a, b int) (err error) {
		defer err2.Handle(&err, func(err error) error {
			// Example for just annotating current err. Normally Handle is
			// used for e.g. cleanup, not annotation that can be left for
			// err2 automatic annotation. See CopyFile example for more
			// information.
			return fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		try.To1(throw())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: error with (1, 2): this is an ERROR
}

func ExampleHandle_multipleHandlerFns() {
	doSomething := func(a, b int) (err error) {
		defer err2.Handle(&err,
			// cause automatic annotation <== 2 error handlers do the trick
			err2.Noop,
			func(err error) error {
				// Example for just annotating current err. Normally Handle
				// is used for e.g. cleanup, not annotation that can be left
				// for err2 automatic annotation. See CopyFile example for
				// more information.
				return fmt.Errorf("%w error with (%d, %d)", err, a, b)
			})
		try.To1(throw())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: testing: run example: this is an ERROR error with (1, 2)
}

func ExampleHandle_noThrow() {
	doSomething := func(a, b int) (err error) {
		defer err2.Handle(&err, func(err error) error {
			return fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		try.To1(noThrow())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: <nil>
}

func BenchmarkOldErrorCheckingWithIfClause(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		if err != nil {
			return
		}
	}
}

func BenchmarkTry_ErrVar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		try.To(err)
	}
}

func BenchmarkTryOut_ErrVar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		try.Out(err).Handle()
	}
}

func BenchmarkTry_StringGenerics(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = try.To1(noThrow())
	}
}

func BenchmarkTryOut_StringGenerics(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = try.Out1(noThrow()).Handle()
	}
}

func BenchmarkTry_StrStrGenerics(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = try.To2(twoStrNoThrow())
	}
}

func BenchmarkTryInsideCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		try.To(noErr())
	}
}

func BenchmarkTryVarCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := noErr()
		try.To(err)
	}
}

func BenchmarkRecursionWithOldErrorCheck(b *testing.B) {
	var recursionWithErrorCheck func(a int) (int, error)
	recursionWithErrorCheck = func(a int) (int, error) {
		if a == 0 {
			return 0, nil
		}
		s, err := noThrow()
		if err != nil {
			return 0, err
		}
		_ = s
		v, err := recursionWithErrorCheck(a - 1)
		if err != nil {
			return 0, err
		}
		return a + v, nil
	}

	for n := 0; n < b.N; n++ {
		_, err := recursionWithErrorCheck(100)
		if err != nil {
			return
		}
	}
}

func BenchmarkRecursionWithOldErrorIfCheckAnd_Defer(b *testing.B) {
	var recursionWithErrorCheckAndDefer func(a int) (_ int, err error)
	recursionWithErrorCheckAndDefer = func(a int) (_ int, err error) {
		defer err2.Handle(&err)

		if a == 0 {
			return 0, nil
		}
		s, err := noThrow()
		if err != nil {
			return 0, err
		}
		_ = s
		v, err := recursionWithErrorCheckAndDefer(a - 1)
		if err != nil {
			return 0, err
		}
		return a + v, nil
	}

	for n := 0; n < b.N; n++ {
		_, err := recursionWithErrorCheckAndDefer(100)
		if err != nil {
			return
		}
	}
}

func BenchmarkRecursionWithTryCall(b *testing.B) {
	var cleanRecursion func(a int) int
	cleanRecursion = func(a int) int {
		if a == 0 {
			return 0
		}
		s := try.To1(noThrow())
		_ = s
		return a + cleanRecursion(a-1)
	}

	for n := 0; n < b.N; n++ {
		_ = cleanRecursion(100)
	}
}

func BenchmarkRecursionWithTryAnd_Empty_Defer(b *testing.B) {
	var recursion func(a int) (r int, err error)
	recursion = func(a int) (r int, err error) {
		defer func(e error) { // try to be as close to our case, but simple!
			err = e
		}(err)

		if a == 0 {
			return 0, nil
		}
		s := try.To1(noThrow())
		_ = s
		r = try.To1(recursion(a - 1))
		r += a
		return r, nil
	}

	for n := 0; n < b.N; n++ {
		_, _ = recursion(100)
	}
}

func doWork(ePtr *error, r any) {
	switch v := r.(type) {
	case nil:
		return
	case runtime.Error:
		*ePtr = fmt.Errorf("%v: %w", *ePtr, v)
	case error:
		*ePtr = fmt.Errorf("%v: %w", *ePtr, v)
	default:
		// panicing
	}
}

// Next benchmark is only for internal test for trying to reproduce Go compilers
// missing optimization behavior.

func BenchmarkRecursionWithTryAnd_HeavyPtrPtr_Defer(b *testing.B) {
	var recursion func(a int) (r int, err error)
	recursion = func(a int) (r int, err error) {
		defer func(ePtr *error) {
			r := recover()
			nothingToDo := r == nil && (ePtr == nil || *ePtr == nil)
			if nothingToDo {
				return
			}
			doWork(ePtr, r)
		}(&err)

		if a == 0 {
			return 0, nil
		}
		s := try.To1(noThrow())
		_ = s
		r = try.To1(recursion(a - 1))
		r += a
		return r, nil
	}

	for n := 0; n < b.N; n++ {
		_, _ = recursion(100)
	}
}

func BenchmarkRecursionWithTryAndDefer(b *testing.B) {
	var recursion func(a int) (r int, err error)
	recursion = func(a int) (r int, err error) {
		defer err2.Handle(&err)

		if a == 0 {
			return 0, nil
		}
		s := try.To1(noThrow())
		_ = s
		r = try.To1(recursion(a - 1))
		r += a
		return r, nil
	}

	for n := 0; n < b.N; n++ {
		_, _ = recursion(100)
	}
}

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	err2.SetPanicTracer(nil)
}

func tearDown() {}
