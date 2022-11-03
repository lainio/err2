// Package handler implements handler for objects returned recovery() function.
package handler_test

import (
	"errors"
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
		{"all nil", args{Info: handler.Info{
			Any:          nil,
			Err:          &myErrNil,
			NilHandler:   nilHandler,
			ErrorHandler: errorHandler,
			PanicHandler: panicHandler,
		},
			panicCalled: false,
			errorCalled: false,
			nilCalled:   false,
			errStr:      "error",
		}},
		{"error in panic", args{Info: handler.Info{
			Any:          errors.New("error"),
			Err:          &myErrNil,
			NilHandler:   nilHandler, // can be here?
			ErrorHandler: errorHandler,
			PanicHandler: panicHandler,
		},
			panicCalled: false,
			errorCalled: true,
			nilCalled:   false,
			errStr:      "error",
		}},
		{"error in panic and default format print", args{Info: handler.Info{
			Any:          errors.New("error"),
			Err:          &myErrNil,
			Format:       "format %v",
			Args:         []any{"test"},
			NilHandler:   nilHandler, // can be here?
			//ErrorHandler: errorHandler,
			PanicHandler: panicHandler,
		},
			panicCalled: false,
			//errorCalled: true,
			nilCalled:   false,
			errStr:      "error",
		}},
		{"error in error", args{Info: handler.Info{
			Any: nil,
			Err: &myErrVal,
			//NilHandler:   nilHandler, // cannot be here
			ErrorHandler: errorHandler,
			PanicHandler: panicHandler,
		},
			panicCalled: false,
			errorCalled: true,
			nilCalled:   false,
			errStr:      "error",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Process(tt.args.Info)
			helper.Require(t, tt.args.panicCalled == panicHandlerCalled, "panicHandler")
			helper.Requiref(t, tt.args.errorCalled == errorHandlerCalled, "errHandled = %v", errorHandlerCalled)
			helper.Require(t, tt.args.nilCalled == nilHandlerCalled, "nilHandler")

			helper.Requiref(t, tt.args.errStr == myErrVal.Error(), "got: %v", myErrVal.Error())
			resetCalleds()
		})
	}
}

var (
	ERR_VAL = errors.New("error")
)

var (
	myErrNil error = nil
	myErrVal       = ERR_VAL

	panicHandlerCalled = false
	errorHandlerCalled = false
	nilHandlerCalled   = false
)

func resetCalleds() {
	myErrNil = nil
	myErrVal = ERR_VAL
	panicHandlerCalled = false
	errorHandlerCalled = false
	nilHandlerCalled = false
}

func panicHandler(_ any) {
	panicHandlerCalled = true
}

func errorHandler(_ error) {
	errorHandlerCalled = true
}

func nilHandler() {
	nilHandlerCalled = true
}
