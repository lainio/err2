// Package handler implements handler for objects returned recovery() function.
package handler_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lainio/err2/internal/handler"
	"github.com/lainio/err2/internal/test"
	"github.com/lainio/err2/internal/x"
)

func TestProcess(t *testing.T) {
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
		// TODO: this test e.g. has problem because Process is not called
		// anymore if Any and Err are nil. Check is done before Process call.
		// - check with handler.WorkToDo() should we even test this.
		{"all nil and our handlers",
			args{Info: handler.Info{
				Any:          nil,
				Err:          &nilError,
				NilHandler:   nilHandler,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			}},
			want{
				errStr: "error",
			}},
		{"error is transported in panic",
			args{Info: handler.Info{
				Any:          errors.New("error"),
				Err:          &nilError,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "error",
			}},
		{"runtime.Error is transported in panic",
			args{Info: handler.Info{
				Any:          myErrRT,
				Err:          &nilError,
				NilHandler:   nilHandler,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: true,
				errStr:      "error",
			}},
		{"panic is transported in panic",
			args{Info: handler.Info{
				Any:          "panic",
				Err:          &nilError,
				NilHandler:   nilHandler,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: true,
				errStr:      "error",
			}},
		{"error in panic and default format print",
			args{Info: handler.Info{
				Any:          errors.New("error"),
				Err:          &nilError,
				Format:       "format %v",
				Args:         []any{"test"},
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: false,
				errStr:      "error",
			}},
		{"error transported in panic and our OWN handler",
			args{Info: handler.Info{
				Any:          errors.New("error"),
				Err:          &nilError,
				Format:       "format %v",
				Args:         []any{"test"},
				ErrorHandler: errorHandlerForAnnotate,
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "annotate: error",
			}},
		{"error is transported in error val",
			args{Info: handler.Info{
				Any:          nil,
				Err:          &myErrVal,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			}},
			want{
				panicCalled: false,
				errorCalled: true,
				errStr:      "error",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if handler.WorkToDo(tt.args.Any, tt.args.Err) {
				handler.Process(&tt.args.Info)

				test.RequireEqual(t, panicHandlerCalled, tt.want.panicCalled)
				test.RequireEqual(t, errorHandlerCalled, tt.want.errorCalled)
				test.RequireEqual(t, nilHandlerCalled, tt.want.nilCalled)

				test.RequireEqual(t, myErrVal.Error(), tt.want.errStr)
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
	Info.Err = &myErrVal // TODO: middle of the perf-refactoring
	myErrVal = handler.PreProcess(&myErrVal, &Info, a...)
}

func TestPreProcess_debug(t *testing.T) {
	// in real case PreProcess is called from Handle function. So, we make our
	// own Handle here. Now our test function name will be the Handle caller
	// and that's what error stack tracing is all about
	Handle()

	test.RequireEqual(t, panicHandlerCalled, false)
	test.RequireEqual(t, errorHandlerCalled, false)
	test.RequireEqual(t, nilHandlerCalled, false)

	// See the name of this test function. Decamel it + error
	const want = "testing t runner: error"
	test.RequireEqual(t, myErrVal.Error(), want)

	resetCalled()
}

func TestPreProcess(t *testing.T) {
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
					Any:          nil,
					Err:          &nilError,
					NilHandler:   nilHandler,
					ErrorHandler: errorHandlerForAnnotate,
					PanicHandler: panicHandler,
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if handler.WorkToDo(tt.args.Any, tt.args.Err) &&
				len(tt.args.a) > 0 {

				var err = x.Whom(tt.args.Info.Err != nil,
					*tt.args.Info.Err, nil)

				// TODO: we should assign it to myErrVal
				err = handler.PreProcess(&err, &tt.args.Info, tt.args.a...)

				test.RequireEqual(t, panicHandlerCalled, tt.want.panicCalled)
				test.RequireEqual(t, errorHandlerCalled, tt.want.errorCalled)
				test.RequireEqual(t, nilHandlerCalled, tt.want.nilCalled)

				test.RequireEqual(t, err.Error(), tt.want.errStr)
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

func nilHandlerForAnnotate() error {
	nilHandlerCalled = true
	// in real case this is closure and it has access to err val
	myErrVal = fmt.Errorf("nil annotate: %v", "error")
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

func nilHandler() error {
	nilHandlerCalled = true
	return nil
}
