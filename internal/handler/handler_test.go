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
		panicCalled bool
		errorCalled bool
		nilCalled   bool

		errStr string
	}
	tests := []struct {
		name string
		args args
	}{
		{"all nil and our handlers",
			args{Info: handler.Info{
				Any:          nil,
				Err:          &nilError,
				NilHandler:   nilHandler,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			},
				errStr: "error",
			}},
		{"error is transported in panic",
			args{Info: handler.Info{
				Any:          errors.New("error"),
				Err:          &nilError,
				ErrorHandler: errorHandler,
				PanicHandler: panicHandler,
			},
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
			},
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
			},
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
			},
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
			},
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
			},
				panicCalled: false,
				errorCalled: true,
				errStr:      "error",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Process(&tt.args.Info)

			helper.Requiref(t, tt.args.panicCalled == panicHandlerCalled, "panicHandler: got = %v, want = %v", tt.args.panicCalled, panicHandlerCalled)
			helper.Requiref(t, tt.args.errorCalled == errorHandlerCalled, "errorHandler: got = %v, want = %v", tt.args.errorCalled, errorHandlerCalled)
			helper.Requiref(t, tt.args.nilCalled == nilHandlerCalled, "nilHandler: got = %v, want = %v", tt.args.nilCalled, nilHandlerCalled)

			helper.Requiref(t, tt.args.errStr == myErrVal.Error(),
				"got: %v, want: %v", myErrVal.Error(), tt.args.errStr)

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