package err2

import (
	"testing"

	"github.com/lainio/err2/internal/test"
)

func TestHandlers(t *testing.T) {
	t.Parallel()
	type args struct {
		f []Handler
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"one", args{f: []Handler{Noop}}, ErrNotFound},
		{"two", args{f: []Handler{Noop, Noop}}, ErrNotFound},
		{"three", args{f: []Handler{Noop, Noop, Noop}}, ErrNotFound},
		{"reset", args{f: []Handler{Noop, Noop, Reset}}, nil},
		{"reset first", args{f: []Handler{Reset, Noop, Noop}}, nil},
		{"reset second", args{f: []Handler{Noop, Reset, Noop}}, nil},
		{"set new first", args{f: []Handler{
			func(error) error { return ErrAlreadyExist }, Noop}}, ErrAlreadyExist},
		{"set new second", args{f: []Handler{Noop,
			func(error) error { return ErrAlreadyExist }, Noop}}, ErrAlreadyExist},
		{"set new first and reset", args{f: []Handler{
			func(error) error { return ErrAlreadyExist }, Reset}}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errHandler := Handlers(tt.args.f...)
			err := errHandler(ErrNotFound)
			if err == nil {
				test.Require(t, tt.want == nil)
			} else {
				test.RequireEqual(t, err.Error(), tt.want.Error())
			}
		})
	}
}
