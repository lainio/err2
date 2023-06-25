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

	Result struct {
		Err error
	}

	Result1[T any] struct {
		Val1 T
		Result
	}

	Result2[T any, U any] struct {
		Val2 U
		Result1[T]
	}
)

func (o *Result) Logf(f string, a ...any) *Result {
	if o.Err == nil {
		return o
	}
	w := tracer.Log.Tracer()
	if w != nil && f != "" {
		fmt.Fprintf(w, f+": %v\n", append(a, o.Err)...)
	}
	return o
}

func (o *Result1[T]) Logf(f string, a ...any) *Result1[T] {
	o.Result.Logf(f, a...)
	return o
}

func (o *Result2[T, U]) Logf(f string, a ...any) *Result2[T, U] {
	o.Result.Logf(f, a...)
	return o
}

func (o *Result) Throwf(f string, a ...any) *Result {
	if o.Err == nil {
		return o
	}
	o.Err = fmt.Errorf(f+wrapStr(), append(a, o.Err)...)
	panic(o.Err)
}

func (o *Result1[T]) Throwf(f string, a ...any) *Result1[T] {
	o.Result.Throwf(f, a...)
	panic(o.Err)
}

func (o *Result2[T, U]) Throwf(f string, a ...any) *Result2[T, U] {
	o.Result.Throwf(f, a...)
	panic(o.Err)
}

func (o *Result1[T]) Def1(v T) *Result1[T] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	return o
}

func (o *Result2[T, U]) Def2(v T, v2 U) *Result2[T, U] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	o.Val2 = v2
	return o
}

func (o *Result) Is(err error, f ErrFn) *Result {
	if o.Err == nil {
		return o
	}
	if errors.Is(o.Err, err) {
		o.Err = f(o.Err)
	}
	return o
}

func (o *Result1[T]) Is(err error, f ErrFn) *Result1[T] {
	o.Result.Is(err, f)
	return o
}

func (o *Result2[T, U]) Is(err error, f ErrFn) *Result2[T, U] {
	o.Result.Is(err, f)
	return o
}

// We could have Catch that don't panic even the err != nil still
func (o *Result) Handle(f ErrFn) *Result {
	if f != nil {
		o.Err = f(o.Err)
		if o.Err != nil {
			panic(o.Err)
		}
	}
	return o
}

// We could have Catch that don't panic even the err != nil still
func (o *Result1[T]) Handle(f ErrFn) *Result1[T] {
	o.Result.Handle(f)
	return o
}

// We could have Catch that don't panic even the err != nil still
func (o *Result2[T, U]) Handle(f ErrFn) *Result2[T, U] {
	o.Result.Handle(f)
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
	return &Result1[T]{Val1: v, Result: Result{Err: err}}
}

func Out2[T any, U any](v1 T, v2 U, err error) *Result2[T, U] {
	return &Result2[T, U]{Val2: v2, Result1: Result1[T]{Val1: v1, Result: Result{Err: err}}}
}

func wrapStr() string {
	return ": %w"
}
