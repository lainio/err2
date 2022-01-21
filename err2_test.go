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

func twoStrNoThrow() (string, string, error) { return "test", "test", nil }
func intStrNoThrow() (int, string, error) { return 1, "test", nil }
func boolIntStrNoThrow() (bool, int, string, error) { return true, 1, "test", nil }

func noThrow() (string, error) { return "test", nil }

func wrongSignature() (int, int) { return 0, 0 }

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

	err2.Try(throw())

	t.Fail() // If everything works we are newer here
}

func TestTry_Error(t *testing.T) {
	var err error
	defer err2.Handle(&err, func() {})

	err2.Try(throw())

	t.Fail() // If everything works we are newer here
}

func panickingHandle() {
	var err error
	defer err2.Handle(&err, func() {})

	err2.Try(wrongSignature())
}

func TestPanickingCarryOn_Handle(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panics should went thru when not our errors")
		}
	}()
	panickingHandle()
}

func panickingCatchAll() {
	defer err2.CatchAll(func(err error) {}, func(v interface{}) {})

	err2.Try(wrongSignature())
}

func TestPanickingCatchAll(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Error("panics should not fly thru")
		}
	}()
	panickingCatchAll()
}

func panickingCatchTrace() {
	defer err2.CatchTrace(func(err error) {})

	err2.Try(wrongSignature())
}

func TestPanickingCatchTrace(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Error("panics should NOT carry on when tracing")
		}
	}()
	panickingCatchTrace()
}

func panickingReturn() {
	var err error
	defer err2.Return(&err)

	err2.Try(wrongSignature())
}

func TestPanicking_Return(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panics should carry on")
		}
	}()
	panickingReturn()
}

func panickingCatch() {
	defer err2.Catch(func(err error) {})

	err2.Try(wrongSignature())
}

func TestPanicking_Catch(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("panics should carry on")
		}
	}()
	panickingCatch()
}

func TestCatch_Error(t *testing.T) {
	defer err2.Catch(func(err error) {
		//fmt.Printf("error and defer handling:%s\n", err)
	})

	err2.Try(throw())

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

		// These helpers are as fast as Check() calls
		r := err2.File.Try(os.Open(src))
		defer r.Close()

		w := err2.File.Try(os.Create(dst))
		defer err2.Handle(&err, func() {
			os.Remove(dst)
		})
		defer w.Close()
		err2.Empty.Try(io.Copy(w, r))
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
	err2.Try(noThrow())
	// Output:
}

func ExampleAnnotate() {
	annotated := func() (err error) {
		defer err2.Annotate("annotated", &err)
		err2.Try(throw())
		return err
	}
	err := annotated()
	fmt.Printf("%v", err)
	// Output: annotated: this is an ERROR
}

func ExampleReturnf() {
	annotated := func() (err error) {
		defer err2.Returnf(&err, "annotated: %s", "err2")
		err2.Try(throw())
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
		err2.Try(throw())
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
		err2.Try(throw())
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
		err2.Try(noThrow())
	}
}

func BenchmarkTry_ErrVar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, err := noThrow()
		err2.Try(err)
	}
}

func BenchmarkTry_StringHelper(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = try.To1(noThrow())
	}
}

func BenchmarkTry_StrStrHelper(b *testing.B) {
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
