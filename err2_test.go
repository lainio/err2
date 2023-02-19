package err2_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/helper"
	"github.com/lainio/err2/try"
)

func throw() (string, error) {
	return "", fmt.Errorf("this is an ERROR")
}

func twoStrNoThrow() (string, string, error)        { return "test", "test", nil }
func intStrNoThrow() (int, string, error)           { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }
func noThrow() (string, error)                      { return "test", nil }

func noErr() error {
	return nil
}

func TestTry_noError(t *testing.T) {
	try.To1(noThrow())
	try.To2(twoStrNoThrow())
	try.To2(intStrNoThrow())
	try.To3(boolIntStrNoThrow())
}

func TestDefault_Error(t *testing.T) {
	var err error
	defer err2.Handle(&err)

	try.To1(throw())

	t.Fail() // If everything works we are newer here
}

func TestTry_Error(t *testing.T) {
	var err error
	defer err2.Handle(&err, func() {})

	try.To1(throw())

	t.Fail() // If everything works we are newer here
}

func TestPanickingCatchAll(t *testing.T) {
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
					defer err2.Catch(func(err error) {}, func(v any) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.Catch(func(err error) {}, func(v any) {})
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
					defer err2.Catch(func(err error) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				helper.Require(t, recover() == nil, "panics should NOT carry on")
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCarryOn_Handle(t *testing.T) {
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
					defer err2.Handle(&err, func() {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					var err error
					defer err2.Handle(&err, func() {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				helper.Require(t, recover() != nil, "panics should went thru when not our errors")
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Handle(t *testing.T) {
	type args struct {
		f func() (err error)
	}
	myErr := fmt.Errorf("my error")

	tests := []struct {
		name  string
		args  args
		wants error
	}{
		{"general error thru panic",
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
					defer err2.Handle(&err, func() {})
					panic("panic")
				},
			},
			nil,
		},
		{"general panic stoped with handler plus err handler",
			args{
				func() (err error) {
					defer err2.Handle(&err, func() {}, func(p any) {})
					panic("panic")
				},
			},
			myErr,
		},
		{"general panic stoped with handler",
			args{
				func() (err error) {
					defer err2.Handle(&err, func(p any) {})
					panic("panic")
				},
			},
			myErr,
		},
		{"general panic stoped with handler plus fmt string",
			args{
				func() (err error) {
					defer err2.Handle(&err, func(p any) {}, "string")
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
					defer err2.Handle(&err, func(p any) {})
					var b []byte
					b[0] = 0
					return nil
				},
			},
			myErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wants == nil {
					helper.Require(t, r != nil, "wants err, then panic")
				}
			}()
			err := tt.args.f()
			if err != nil {
				helper.Requiref(t, err == myErr, "got %p, want %p", err, myErr)
			}
		})
	}
}

func TestPanicking_Catch(t *testing.T) {
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
					defer err2.Catch(func(err error) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.Catch(func(err error) {})
					var b []byte
					b[0] = 0
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				helper.Require(t, recover() == nil, "panics should NOT carry on")
			}()
			tt.args.f()
		})
	}
}

func TestCatch_Error(t *testing.T) {
	defer err2.Catch()

	try.To1(throw())

	t.Fail() // If everything works we are newer here
}

func TestCatch_Panic(t *testing.T) {
	panicHandled := false
	defer func() {
		// when err2.Catch's panic handler works fine, panic is handled
		if !panicHandled {
			t.Fail()
		}
	}()

	defer err2.Catch(
		func(err error) {
			t.Log("it was panic, not an error")
			t.Fail() // we should not be here
		},
		func(v any) {
			panicHandled = true
		})

	panic("test panic")
}

func TestSetErrorTracer(t *testing.T) {
	w := err2.ErrorTracer()
	helper.Require(t, w == nil, "error tracer should be nil")
	var w1 io.Writer
	err2.SetErrorTracer(w1)
	w = err2.ErrorTracer()
	helper.Require(t, w == nil, "error tracer should be nil")
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
	// Output: testing run example: our error
}

func ExampleHandle_errReturn() {
	normalReturn := func() (err error) {
		defer err2.Handle(&err, "")
		return fmt.Errorf("our error")
	}
	err := normalReturn()
	fmt.Printf("%v", err)
	// Output: our error
}

func ExampleReturnf_empty() {
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

func ExampleReturnf_deferStack() {
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
		defer err2.Handle(&err, func() {
			// Example for just annotating current err. Normally Handle is
			// used for cleanup. See CopyFile example for more information.
			err = fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		try.To1(throw())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: error with (1, 2): this is an ERROR
}

func ExampleHandle_noThrow() {
	doSomething := func(a, b int) (err error) {
		defer err2.Handle(&err, func() {
			err = fmt.Errorf("error with (%d, %d): %v", a, b, err)
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

func BenchmarkTry_StringGenerics(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = try.To1(noThrow())
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
