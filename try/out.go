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
		Throwf(f string, a ...any) handler // op
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

func (o *Result) Logf(a ...any) *Result {
	if o.Err == nil || len(a) == 0 {
		return o
	}
	w := tracer.Log.Tracer()
	if w != nil {
		f, isFormat := a[0].(string)
		if isFormat {
			fmt.Fprintf(w, f+": %v\n", append(a[1:], o.Err)...)
		}
	}
	return o
}

func (o *Result1[T]) Logf(a ...any) *Result1[T] {
	o.Result.Logf(a...)
	return o
}

func (o *Result2[T, U]) Logf(a ...any) *Result2[T, U] {
	o.Result.Logf(a...)
	return o
}

func (o *Result) Throwf(a ...any) *Result {
	if o.Err == nil {
		return o
	}
	if len(a) == 0 {
		panic(o.Err)
	}
	f, isFormat := a[0].(string)
	if isFormat {
		o.Err = fmt.Errorf(f+wrapStr(), append(a[1:], o.Err)...)
	}
	panic(o.Err)
}

func (o *Result1[T]) Throwf(a ...any) *Result1[T] {
	o.Result.Throwf(a...)
	return o
}

func (o *Result2[T, U]) Throwf(a ...any) *Result2[T, U] {
	o.Result.Throwf(a...)
	return o
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

// Out is a helper function to call functions which returns Result and start
// error handlin with DSL. For instance, to implement same as try.To(), you
// could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v).Throwf()
func Out(err error) *Result {
	return &Result{Err: err}
}

// Out1 is a helper function to call functions which returns (T, error). That
// allows you to use Result1, which makes possible to
// start error handling with DSL. For instance, instead of try.To1() you could
// do the following:
//
//	d := try.Out1(os.ReadFile(filename).Throwf().Val1
//
// or in some other cases some of these would be desired action:
//
//		number := try.Out1(strconv.Atoi(str)).Def1(100).Val1
//		try.Out(os.Remove(dst)).Logf("remove")
//	  try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Def2(1, 2)
//	  try.Out1(strconv.Atoi(s)).Logf("not number").Def1(100).Val1
func Out1[T any](v T, err error) *Result1[T] {
	return &Result1[T]{Val1: v, Result: Result{Err: err}}
}

// Out2 is a helper function to call functions which returns (T, error). That
// allows you to use Result2, which makes possible to
// start error handling with DSL. For instance, instead of try.To2() you could
// do the following:
//
//	d := try.Out2(os.ReadFile(filename).Throwf().Val2
//
// or in some other cases some of these would be desired action:
//
//	try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Def2(1, 2)
//	try.Out2(convTwoStr(s1, s2)).Throwf().Val2
func Out2[T any, U any](v1 T, v2 U, err error) *Result2[T, U] {
	return &Result2[T, U]{Val2: v2, Result1: Result1[T]{Val1: v1, Result: Result{Err: err}}}
}

func wrapStr() string {
	return ": %w"
}
