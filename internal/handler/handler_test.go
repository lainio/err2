// Package handler implements handler for objects returned recovery() function.
package handler_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lainio/err2/internal/except"
	"github.com/lainio/err2/internal/handler"
	"github.com/lainio/err2/internal/x"
)

func TestProcess(t *testing.T) {
	// NOTE. No Parallel, uses pkg lvl variables
	type args struct {
		handler.Info
	}
	type want struct {
		panicCalled bool
		errorCalled bool
		nilCalled   bool

		errStr string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"all nil and our handlers",
			args{Info: handler.Info{
				Any:     nil,
				Err:     &nilError,
				NilFn:   nilHandler,
				ErrorFn: errorHandler,
				PanicFn: panicHandler,
			}},
			want{
				errStr: "error",
			}},
		{"error is transported in panic",
			args{Info: handler.Info{
				Any:     errors.New("error"),
				Err:     &nilError,
				ErrorFn: errorHandler,
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "error",
			}},
		{"runtime.Error is transported in panic",
			args{Info: handler.Info{
				Any:     myErrRT,
				Err:     &nilError,
				NilFn:   nilHandler,
				ErrorFn: errorHandler,
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: true,
				errStr:      "error",
			}},
		{"panic is transported in panic",
			args{Info: handler.Info{
				Any:     "panic",
				Err:     &nilError,
				NilFn:   nilHandler,
				ErrorFn: errorHandler,
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: true,
				errStr:      "error",
			}},
		{"error in panic and default format print",
			args{Info: handler.Info{
				Any:     errors.New("error"),
				Err:     &nilError,
				Format:  "format %v",
				Args:    []any{"test"},
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: false,
				errStr:      "error",
			}},
		{"error transported in panic and our OWN handler",
			args{Info: handler.Info{
				Any:     errors.New("error"),
				Err:     &nilError,
				Format:  "format %v",
				Args:    []any{"test"},
				ErrorFn: errorHandlerForAnnotate,
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "annotate: error",
			}},
		{"error is transported in error val",
			args{Info: handler.Info{
				Any:     nil,
				Err:     &myErrVal,
				ErrorFn: errorHandler,
				PanicFn: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "error",
			}},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			// NOTE. No Parallel, uses pkg lvl variables
			if handler.WorkToDo(tt.args.Any, tt.args.Err) {
				handler.Process(&tt.args.Info)

				except.Equal(t, panicHandlerCalled, tt.want.panicCalled)
				except.Equal(t, errorHandlerCalled, tt.want.errorCalled)
				except.Equal(t, nilHandlerCalled, tt.want.nilCalled)

				except.Equal(t, myErrVal.Error(), tt.want.errStr)
			}
			resetCalled()
		})
	}
}

// this is easier to debug even the same test(s) are in table
var Info = handler.Info{
	Any: nil,
	Err: &myErrVal,
}

func Handle() {
	a := []any{}
	Info.Err = &myErrVal
	myErrVal = handler.PreProcess(&myErrVal, &Info, a)
}

func TestPreProcess_debug(t *testing.T) {
	// NOTE. No Parallel, uses pkg lvl variables

	// in real case PreProcess is called from Handle function. So, we make our
	// own Handle here. Now our test function name will be the Handle caller
	// and that's what error stack tracing is all about
	Handle()

	except.ThatNot(t, panicHandlerCalled)
	except.ThatNot(t, errorHandlerCalled)
	except.ThatNot(t, nilHandlerCalled)

	// See the name of this test function. Decamel it + error
	const want = "testing: t runner: error"
	except.Equal(t, myErrVal.Error(), want)

	resetCalled()
}

func TestPreProcess(t *testing.T) {
	// NOTE. No Parallel, uses pkg lvl variables
	type args struct {
		handler.Info
		a []any
	}
	type want struct {
		panicCalled bool
		errorCalled bool
		nilCalled   bool

		errStr string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"error is transported in panic",
			args{
				Info: handler.Info{
					Any: errors.New("error"),
					Err: &nilError,
				},
				a: []any{nilHandlerForAnnotate},
			},
			want{
				nilCalled: true,
				errStr:    "nil annotate: error",
			}},
		{"all nil and our handlers",
			args{
				Info: handler.Info{
					Any:     nil,
					Err:     &nilError,
					NilFn:   nilHandler,
					ErrorFn: errorHandlerForAnnotate,
					PanicFn: panicHandler,
				},
				a: []any{"test"}}, // no affec because
			want{
				errStr: "error",
			}},
		{"error in panic and only nilHandler is used",
			args{
				Info: handler.Info{
					Any: errors.New("error"),
					Err: &nilError,
				},
				a: []any{nilHandlerForAnnotate},
			},
			want{
				nilCalled: true,
				errStr:    "nil annotate: error",
			}},
		{"error in panic and default annotation",
			args{
				Info: handler.Info{
					Any: nil,
					Err: &myErrVal,
				},
				a: []any{},
			},
			want{
				nilCalled: false,
				errStr:    "",
			}},
	}
	for _, ttv := range tests {
		tt := ttv
		t.Run(tt.name, func(t *testing.T) {
			// NOTE. No Parallel, uses pkg lvl variables
			if handler.WorkToDo(tt.args.Any, tt.args.Err) &&
				len(tt.args.a) > 0 {

				var err = x.Whom(tt.args.Info.Err != nil,
					*tt.args.Info.Err, nil)

				err = handler.PreProcess(&err, &tt.args.Info, tt.args.a)

				except.Equal(t, panicHandlerCalled, tt.want.panicCalled)
				except.Equal(t, errorHandlerCalled, tt.want.errorCalled)
				except.Equal(t, nilHandlerCalled, tt.want.nilCalled)

				except.Equal(t, err.Error(), tt.want.errStr)
			}
			resetCalled()
		})
	}
}

type myRuntimeErr struct{}

func (rte myRuntimeErr) RuntimeError() {}
func (rte myRuntimeErr) Error() string { return "runtime error" }

var (
	errVal = errors.New("error")
	errRT  = new(myRuntimeErr)
)

var (
	// Important 'cause our errors are ptrs to error interface
	nilError error

	myErrVal = errVal
	myErrRT  = errRT

	panicHandlerCalled = false
	errorHandlerCalled = false
	nilHandlerCalled   = false
)

func resetCalled() {
	nilError = nil
	myErrVal = errVal
	myErrRT = errRT
	panicHandlerCalled = false
	errorHandlerCalled = false
	nilHandlerCalled = false
}

func panicHandler(_ any) {
	panicHandlerCalled = true
}

func nilHandlerForAnnotate(err error) error {
	nilHandlerCalled = true
	myErrVal = fmt.Errorf("nil annotate: %w", err)
	return myErrVal
}

func errorHandlerForAnnotate(err error) error {
	errorHandlerCalled = true
	myErrVal = fmt.Errorf("annotate: %v", err)
	return myErrVal
}

func errorHandler(err error) error {
	errorHandlerCalled = true
	return err
}

func nilHandler(err error) error {
	nilHandlerCalled = true
	return err
}
