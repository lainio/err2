package err2_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func throw() (string, error) {
	return "", fmt.Errorf("this is an ERROR")
}

func twoStrNoThrow() (string, string, error)        { return "test", "test", nil }
func intStrNoThrow() (int, string, error)           { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }
func noThrow() (string, error)                      { return "test", nil }
func wrongSignature() (int, int)                    { return 0, 0 }

func recursion(a int) int {
	if a == 0 {
		return 0
	}
	s, err := noThrow()
	err2.Check(err)
	_ = s
	return a + recursion(a-1)
}

func recursionWithErrorCheck(a int) (int, error) {
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
	defer err2.Return(&err)

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
					defer err2.CatchAll(func(err error) {}, func(v any) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.CatchAll(func(err error) {}, func(v any) {})
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
				if recover() != nil {
					t.Error("panics should not fly thru")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanickingCatchTrace(t *testing.T) {
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
					defer err2.CatchTrace(func(err error) {})
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					defer err2.CatchTrace(func(err error) {})
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
				if recover() != nil {
					t.Error("panics should NOT carry on when tracing")
				}
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
				if recover() == nil {
					t.Error("panics should went thru when not our errors")
				}
			}()
			tt.args.f()
		})
	}
}

func TestPanicking_Return(t *testing.T) {
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
					defer err2.Return(&err)
					panic("panic")
				},
			},
			nil,
		},
		{"runtime.error panic",
			args{
				func() {
					var err error
					defer err2.Return(&err)
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
				if recover() == nil {
					t.Error("panics should carry on")
				}
			}()
			tt.args.f()
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
				if recover() == nil {
					t.Error("panics should carry on")
				}
			}()
			tt.args.f()
		})
	}
}

func TestCatch_Error(t *testing.T) {
	defer err2.Catch(func(err error) {
		//fmt.Printf("error and defer handling:%s\n", err)
	})

	try.To1(throw())

	t.Fail() // If everything works we are newer here
}

func ExampleFilterTry() {
	copyStream := func(src string) (s string, err error) {
		defer err2.Returnf(&err, "copy stream %s", src)

		in := bytes.NewBufferString(src)
		tmp := make([]byte, 4)
		var out bytes.Buffer
		for n, err := in.Read(tmp); !err2.FilterTry(io.EOF, err); n, err = in.Read(tmp) {
			out.Write(tmp[:n])
		}

		return out.String(), nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: testing string
}

func ExampleTryEOF() {
	copyStream := func(src string) (s string, err error) {
		defer err2.Returnf(&err, "copy stream %s", src)

		in := bytes.NewBufferString(src)
		tmp := make([]byte, 4)
		var out bytes.Buffer
		for n, err := in.Read(tmp); !err2.TryEOF(err); n, err = in.Read(tmp) {
			out.Write(tmp[:n])
		}

		return out.String(), nil
	}

	str, err := copyStream("testing string")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(str)
	// Output: testing string
}

func Example_copyFile() {
	copyFile := func(src, dst string) (err error) {
		defer err2.Returnf(&err, "copy %s %s", src, dst)

		// These try package helpers are as fast as Check() calls which is as
		// fast as `if err != nil {}`

		r := try.To1(os.Open(src))
		defer r.Close()

		w := try.To1(os.Create(dst))
		defer err2.Handle(&err, func() {
			os.Remove(dst)
		})
		defer w.Close()
		try.To1(io.Copy(w, r))
		return nil
	}

	err := copyFile("/notfound/path/file.go", "/notfound/path/file.bak")
	if err != nil {
		fmt.Println(err)
	}
	// Output: copy /notfound/path/file.go /notfound/path/file.bak: open /notfound/path/file.go: no such file or directory
}

func ExampleReturn() {
	var err error
	defer err2.Return(&err)
	try.To1(noThrow())
	// Output:
}

func ExampleAnnotate() {
	annotated := func() (err error) {
		defer err2.Annotate("annotated", &err)
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: this is an ERROR
}

func ExampleReturnf() {
	annotated := func() (err error) {
		defer err2.Returnf(&err, "annotated: %s", "err2")
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: err2: this is an ERROR
}

func ExampleAnnotate_deferStack() {
	annotated := func() (err error) {
		defer err2.Annotate("annotated 2nd", &err)
		defer err2.Annotate("annotated 1st", &err)
		try.To1(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated 2nd: annotated 1st: this is an ERROR
}

func ExampleHandle() {
	doSomething := func(a, b int) (err error) {
		defer err2.Handle(&err, func() {
			err = fmt.Errorf("error with (%d, %d): %v", a, b, err)
		})
		try.To1(throw())
		return err
	}
	err := doSomething(1, 2)
	fmt.Printf("%v", err)
	// Output: error with (1, 2): this is an ERROR
}

func BenchmarkOldErrorCheckingWithIfClause(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		if err != nil {
			return
		}
	}
}

func BenchmarkTry(b *testing.B) {
	for n := 0; n < b.N; n++ {
		try.To1(noThrow())
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

func BenchmarkCheckInsideCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err2.Check(noErr())
	}
}

func BenchmarkCheckVarCall(b *testing.B) {
	for n := 0; n < b.N; n++ {
		err := noErr()
		err2.Check(err)
	}
}

func BenchmarkCheck_ErrVar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		err2.Check(err)
	}
}

func BenchmarkRecursionNoCheck_NotRelated(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = recursion(100)
	}
}

func BenchmarkRecursionWithErrorCheck_NotRelated(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := recursionWithErrorCheck(100)
		if err != nil {
			return
		}
	}
}
