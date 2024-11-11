package handler_test

import (
	"testing"

	"github.com/lainio/err2"
	"github.com/lainio/err2/internal/except"
	"github.com/lainio/err2/internal/handler"
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
		dis  bool
	}{
		{"one", args{f: []any{err2.Noop}}, err2.ErrNotFound, false},
		{
			"one disabled NOT real case",
			args{f: []any{nil}},
			err2.ErrNotFound,
			true,
		},
		{
			"two",
			args{f: []any{err2.Noop, err2.Noop}},
			err2.ErrNotFound,
			false,
		},
		{
			"three",
			args{f: []any{err2.Noop, err2.Noop, err2.Noop}},
			err2.ErrNotFound,
			false,
		},
		{
			"three last disabled",
			args{f: []any{err2.Noop, err2.Noop, nil}},
			err2.ErrNotFound,
			true,
		},
		{
			"three 2nd disabled",
			args{f: []any{err2.Noop, nil, err2.Noop}},
			err2.ErrNotFound,
			true,
		},
		{
			"three all disabled",
			args{f: []any{nil, nil, nil}},
			err2.ErrNotFound,
			true,
		},
		{
			"reset",
			args{f: []any{err2.Noop, err2.Noop, err2.Reset}},
			nil,
			false,
		},
		{
			"reset and disabled",
			args{f: []any{nil, err2.Noop, err2.Reset}},
			nil,
			true,
		},
		{
			"reset first",
			args{f: []any{err2.Reset, err2.Noop, err2.Noop}},
			nil,
			false,
		},
		{
			"reset second",
			args{f: []any{err2.Noop, err2.Reset, err2.Noop}},
			nil,
			false,
		},
		{"set new first", args{f: []any{
			func(error) error { return err2.ErrAlreadyExist }, err2.Noop}}, err2.ErrAlreadyExist, false},
		{"set new second", args{f: []any{err2.Noop,
			func(error) error { return err2.ErrAlreadyExist }, err2.Noop}}, err2.ErrAlreadyExist, false},
		{"set new first and reset", args{f: []any{
			func(error) error { return err2.ErrAlreadyExist }, err2.Reset}}, nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			anys := tt.args.f

			except.That(t, anys != nil, "cannot be nil")
			fns, dis := handler.ToErrorFns(anys)
			except.That(t, fns != nil, "cannot be nil")
			except.Equal(t, dis, tt.dis, "disabled wanted")

			errHandler := handler.Pipeline(fns)
			err := errHandler(err2.ErrNotFound)
			if err == nil {
				except.That(t, tt.want == nil)
			} else {
				except.Equal(t, err.Error(), tt.want.Error())
			}
		})
	}
}
