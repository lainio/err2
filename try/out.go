package try

import (
	"errors"
	"fmt"

	"github.com/lainio/err2/internal/tracer"
)

type (
	ErrFn func(err error) error

	// TODO: this is not needed this for my notes at the moment.
	handler interface {
		Logf(f string, a ...any) handler
		IsDo(err error, f ErrFn) handler
		Is(err error) bool
		Handle()
		Catch()
		Throwf(f string, a ...any) handler              // op
		IsThrowf(err error, f string, a ...any) handler // op
	}

	Result1[T any] struct {
		Val1 T
		err  error
	}

	Result2[T any, U any] struct {
		Val1 T
		Val2 U
		err  error
	}
)

func (o *Result1[T]) Logf(f string, a ...any) *Result1[T] {
	if o.err == nil {
		return o
	}
	w := tracer.Error.Tracer()
	if w != nil {
		fmt.Fprintf(w, f+": %v\n", append(a, o.err)...)
	}
	return o
}

func (o *Result1[T]) Throwf(f string, a ...any) *Result1[T] {
	if o.err == nil {
		return o
	}
	o.err = fmt.Errorf(f+wrapStr(), append(a, o.err)...)
	panic(o.err)
}

// TODO: is this really nessessary? should we rename isDo and offer only it?
func (o *Result1[T]) Is(err error) (T, bool) {
	if err == nil || o.err == nil {
		return o.Val1, err == nil && o.err == nil
	}
	return o.Val1, Is(o.err, err)
}

func (o *Result1[T]) Def1(v T) *Result1[T] {
	if o.err == nil {
		return o
	}
	o.Val1 = v
	return o
}

func (o *Result1[T]) IsDo(err error, f ErrFn) *Result1[T] {
	if o.err == nil {
		return o
	}
	if errors.Is(o.err, err) {
		o.err = f(o.err)
	}
	return o
}

// We could have Catch that don't panic even the err != nil still
func (o *Result1[T]) Handle(f ErrFn) *Result1[T] {
	if f != nil {
		o.err = f(o.err)
		if o.err != nil {
			panic(o.err)
		}
	}
	return o
}

// Out1 is a helper function to call functions which returns (T, error) and
// start error handlint with DSL. For instance, to implement try.To1() you could
// do the following:
//
//	d := try.Out1(os.ReadFile(filname).Throwf().Val1
//
// or in some other cases this would be desired action:
//
//	number := try.Out1(strconv.Atoi(str)).Def1(100).Val1
func Out1[T any](v T, err error) *Result1[T] {
	return &Result1[T]{Val1: v, err: err}
}

func Out2[T any, U any](v1 T, v2 U, err error) *Result2[T, U] {
	return &Result2[T, U]{Val1: v1, Val2: v2, err: err}
}

func wrapStr() string {
	return ": %w"
}
