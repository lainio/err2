package handler_test

import (
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/handler"
	"github.com/lainio/err2/internal/test"
)

func TestHandlers(t *testing.T) {
	t.Parallel()
	type args struct {
		f []any // we use any because it's same as real-world case at start
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"one", args{f: []any{err2.Noop}}, err2.ErrNotFound},
		{"two", args{f: []any{err2.Noop, err2.Noop}}, err2.ErrNotFound},
		{"three", args{f: []any{err2.Noop, err2.Noop, err2.Noop}}, err2.ErrNotFound},
		{"reset", args{f: []any{err2.Noop, err2.Noop, err2.Reset}}, nil},
		{"reset first", args{f: []any{err2.Reset, err2.Noop, err2.Noop}}, nil},
		{"reset second", args{f: []any{err2.Noop, err2.Reset, err2.Noop}}, nil},
		{"set new first", args{f: []any{
			func(error) error { return err2.ErrAlreadyExist }, err2.Noop}}, err2.ErrAlreadyExist},
		{"set new second", args{f: []any{err2.Noop,
			func(error) error { return err2.ErrAlreadyExist }, err2.Noop}}, err2.ErrAlreadyExist},
		{"set new first and reset", args{f: []any{
			func(error) error { return err2.ErrAlreadyExist }, err2.Reset}}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			anys := tt.args.f

			test.Require(t, anys != nil, "cannot be nil")
			fns := handler.AssertErrHandlers(anys)
			test.Require(t, fns != nil, "cannot be nil")

			errHandler := handler.Pipeline(fns)
			err := errHandler(err2.ErrNotFound)
			if err == nil {
				test.Require(t, tt.want == nil)
			} else {
				test.RequireEqual(t, err.Error(), tt.want.Error())
			}
		})
	}
}
