package try_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/expect"
	"github.com/lainio/err2/try"
)

func TestResultHandleCombinations(t *testing.T) {
	t.Parallel()

	original := errors.New("original error")
	wrapped := fmt.Errorf("wrapped: %w", original)
	replacement := errors.New("replacement error")
	tests := []struct {
		name       string
		err        error
		args       func(try.ErrFn) []any
		handlerErr error
		wantErr    error
		wantText   string
		wantCalled bool
	}{
		{name: "no arguments", err: original, wantErr: original},
		{name: "annotation", err: original,
			args:    func(try.ErrFn) []any { return []any{"operation"} },
			wantErr: original, wantText: "operation: original error"},
		{name: "formatted annotation", err: wrapped,
			args:    func(try.ErrFn) []any { return []any{"operation %d", 42} },
			wantErr: original, wantText: "operation 42: wrapped: original error"},
		{name: "handler preserves error", err: original,
			args: handlerArgs, handlerErr: original,
			wantErr: original, wantCalled: true},
		{name: "handler replaces error", err: original,
			args: handlerArgs, handlerErr: replacement,
			wantErr: replacement, wantCalled: true},
		{name: "handler suppresses error", err: original,
			args: handlerArgs, wantCalled: true},
		{name: "matching target suppresses error", err: original,
			args: targetArgs(original), wantCalled: true},
		{name: "wrapped target suppresses error", err: wrapped,
			args: targetArgs(original), wantCalled: true},
		{name: "matching target preserves error", err: wrapped,
			args: targetArgs(original), handlerErr: wrapped,
			wantErr: wrapped, wantCalled: true},
		{name: "matching target replaces error", err: wrapped,
			args: targetArgs(original), handlerErr: replacement,
			wantErr: replacement, wantCalled: true},
		{name: "nonmatching target", err: wrapped,
			args: targetArgs(io.EOF), wantErr: wrapped},
	}

	for _, resultType := range []string{"Result", "Result1", "Result2"} {
		resultType := resultType
		t.Run(resultType, func(t *testing.T) {
			t.Parallel()
			for _, tt := range tests {
				tt := tt
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					called := 0
					handler := func(err error) error {
						called++
						expect.Thatf(t, err == tt.err,
							"handler received %v, want original %v", err, tt.err)
						return tt.handlerErr
					}
					var args []any
					if tt.args != nil {
						args = tt.args(handler)
					}
					for _, input := range []error{tt.err, nil} {
						invoke, current := resultHandle(t, resultType, input)
						called = 0
						var thrown any
						func() {
							defer func() { thrown = recover() }()
							invoke(args...)
						}()
						wantErr := tt.wantErr
						wantCalled := tt.wantCalled
						if input == nil {
							wantErr, wantCalled = nil, false
						}
						expect.Thatf(t, (called == 1) == wantCalled && called <= 1,
							"input %v: handler called %d times, want called=%v", input, called, wantCalled)
						if wantErr == nil {
							expect.Thatf(t, thrown == nil && current() == nil,
								"input %v: panic=%v, Err=%v, want both nil", input, thrown, current())
							continue
						}
						got, ok := thrown.(error)
						expect.Thatf(t, ok && errors.Is(got, wantErr),
							"panic=%v, want error wrapping %v", thrown, wantErr)
						expect.Thatf(t, current() == got,
							"Err=%v, want thrown error %v", current(), got)
						if tt.wantText != "" {
							expect.Equal(t, got.Error(), tt.wantText)
						} else {
							expect.Thatf(t, got == wantErr,
								"panic=%v, want exact error %v", got, wantErr)
						}
					}
				})
			}
		})
	}
}

func handlerArgs(fn try.ErrFn) []any { return []any{fn} }

func targetArgs(target error) func(try.ErrFn) []any {
	return func(fn try.ErrFn) []any { return []any{target, fn} }
}

// resultHandle checks that successful calls preserve fluent return types and values.
func resultHandle(t *testing.T, resultType string, err error) (func(...any), func() error) {
	t.Helper()
	switch resultType {
	case "Result1":
		r := try.Out1(42, err)
		return func(a ...any) {
			got := r.Handle(a...)
			expect.That(t, got == r && got.Val1 == 42,
				"Handle must return the same Result1 and preserve Val1")
		}, func() error { return r.Err }
	case "Result2":
		r := try.Out2(42, "value", err)
		return func(a ...any) {
			got := r.Handle(a...)
			expect.That(t, got == r && got.Val1 == 42 && got.Val2 == "value",
				"Handle must return the same Result2 and preserve both values")
		}, func() error { return r.Err }
	default:
		r := try.Out(err)
		return func(a ...any) {
			expect.Equal(t, r.Handle(a...), r, "Handle must return the same Result")
		}, func() error { return r.Err }
	}
}

func ExampleResult_Handle() {
	read := func(err error) (result error) {
		defer err2.Handle(&result, nil)

		try.Out(err).Handle(io.EOF, func(err error) error {
			fmt.Println("end of input:", err)
			return nil // Suppress only errors matching io.EOF.
		})
		return nil
	}

	fmt.Println("matching:", read(fmt.Errorf("read: %w", io.EOF)))
	fmt.Println("other:", read(io.ErrClosedPipe))
	fmt.Println("success:", read(nil))
	// Output:
	// end of input: read: EOF
	// matching: <nil>
	// other: io: read/write on closed pipe
	// success: <nil>
}
