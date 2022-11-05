// Package handler implements handler for objects returned recovery() function.
package handler_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lainio/err2/internal/handler"
	"github.com/lainio/err2/internal/helper"
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
			handler.Process(&tt.args.Info)

			helper.Requiref(t, tt.want.panicCalled == panicHandlerCalled,
				"panicHandler: got = %v, want = %v",
				panicHandlerCalled, tt.want.panicCalled)
			helper.Requiref(t, tt.want.errorCalled == errorHandlerCalled,
				"errorHandler: got = %v, want = %v",
				errorHandlerCalled, tt.want.errorCalled)
			helper.Requiref(t, tt.want.nilCalled == nilHandlerCalled,
				"nilHandler: got = %v, want = %v",
				nilHandlerCalled, tt.want.nilCalled)

			helper.Requiref(t, tt.want.errStr == myErrVal.Error(),
				"got: %v, want: %v", myErrVal.Error(), tt.want.errStr)

			resetCalled()
		})
	}
}

func TestPreProcess_debug(t *testing.T) {
	// this is easier to debug even the same test(s) are in table
	Info := handler.Info{
		Any: errors.New("error"),
		Err: &nilError,
	}
	a := []any{nilHandlerForAnnotate}

	handler.PreProcess(&Info, a...)
	helper.Requiref(t, false == panicHandlerCalled,
		"panicHandler: got = %v, want = %v",
		panicHandlerCalled, false)
	helper.Requiref(t, false == errorHandlerCalled,
		"errorHandler: got = %v, want = %v",
		errorHandlerCalled, false)
	helper.Requiref(t, true == nilHandlerCalled,
		"nilHandler: got = %v, want = %v",
		nilHandlerCalled, true)

	const want = "nil annotate: error"
	helper.Requiref(t, want == myErrVal.Error(),
		"got: %v, want: %v", myErrVal.Error(), want)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args.a) > 0 {
				handler.PreProcess(&tt.args.Info, tt.args.a...)

				helper.Requiref(t, tt.want.panicCalled == panicHandlerCalled,
					"panicHandler: got = %v, want = %v",
					panicHandlerCalled, tt.want.panicCalled)
				helper.Requiref(t, tt.want.errorCalled == errorHandlerCalled,
					"errorHandler: got = %v, want = %v",
					errorHandlerCalled, tt.want.errorCalled)
				helper.Requiref(t, tt.want.nilCalled == nilHandlerCalled,
					"nilHandler: got = %v, want = %v",
					nilHandlerCalled, tt.want.nilCalled)

				helper.Requiref(t, tt.want.errStr == myErrVal.Error(),
					"got: %v, want: %v", myErrVal.Error(), tt.want.errStr)

				resetCalled()
			}
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
	nilError error = nil

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

func nilHandlerForAnnotate() {
	nilHandlerCalled = true
	// in real case this is closure and it has access to err val
	myErrVal = fmt.Errorf("nil annotate: %v", "error")
}

func errorHandlerForAnnotate(err error) {
	errorHandlerCalled = true
	myErrVal = fmt.Errorf("annotate: %v", err)
}

func errorHandler(_ error) {
	errorHandlerCalled = true
}

func nilHandler() {
	nilHandlerCalled = true
}
