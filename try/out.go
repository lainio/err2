package try

import (
	"errors"
	"fmt"

	"github.com/lainio/err2/internal/tracer"
)

type (
	// ErrFn is function type for try.OutX handlers.
	ErrFn = func(err error) error

	// Result is the base of our error handling DSL for try.Out functions.
	Result struct {
		Err error
	}

	// Result1 is the base of our error handling DSL for try.Out1 functions.
	Result1[T any] struct {
		Val1 T
		Result
	}

	// Result2 is the base of our error handling DSL for try.Out2 functions.
	Result2[T any, U any] struct {
		Val2 U
		Result1[T]
	}
)

// Logf prints a log line to pre-set logging stream (if set err2.SetLogWriter)
// if the current Result.Err != nil.
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

// Logf prints a log line to pre-set logging stream (if set err2.SetLogWriter)
// if the current Result.Err != nil.
func (o *Result1[T]) Logf(a ...any) *Result1[T] {
	o.Result.Logf(a...)
	return o
}

// Logf prints a log line to pre-set logging stream (if set err2.SetLogWriter)
// if the current Result.Err != nil.
func (o *Result2[T, U]) Logf(a ...any) *Result2[T, U] {
	o.Result.Logf(a...)
	return o
}

// Handle annotates and throws an error immediately i.e. terminates error handling
// DSL chain if Result.Err != nil. Handle supports error annotation similarly as
// fmt.Errorf.
func (o *Result) Handle(a ...any) *Result {
	if o.Err == nil {
		return o
	}
	if len(a) == 0 {
		panic(o.Err)
	}

	switch f := a[0].(type) {
	case string:
		o.Err = fmt.Errorf(f+wrapStr(), append(a[1:], o.Err)...)
	case ErrFn:
		o.Err = f(o.Err)
		if o.Err != nil {
			panic(o.Err)
		}
	case error:
		if len(a) == 2 {
			a1 := a[1]
			hfn, haveHandlerFn := a1.(ErrFn)
			if haveHandlerFn {
				if errors.Is(o.Err, f) {
					o.Err = hfn(o.Err)
				}
			}
		}
	}
	if o.Err != nil {
		panic(o.Err)
	}
	return o
}

// Handle annotates and throws an error immediately i.e. terminates error handling
// DSL chain if Result.Err != nil. Handle supports error annotation similarly as
// fmt.Errorf.
func (o *Result1[T]) Handle(a ...any) *Result1[T] {
	o.Result.Handle(a...)
	return o
}

// Handle annotates and throws an error immediately i.e. terminates error handling
// DSL chain if Result.Err != nil. Handle supports error annotation similarly as
// fmt.Errorf.
func (o *Result2[T, U]) Handle(a ...any) *Result2[T, U] {
	o.Result.Handle(a...)
	return o
}

// Def1 sets default value for Result.Val1. The value is returned in case of
// Result.Err != nil.
func (o *Result1[T]) Def1(v T) *Result1[T] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	return o
}

// Def2 sets default value for Result.Val2. The value is returned in case of
// Result.Err != nil.
func (o *Result2[T, U]) Def2(v T, v2 U) *Result2[T, U] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	o.Val2 = v2
	return o
}

// Is allows you to add error.Is handler to try.OutX handler chain. Internally
// Is calls errors.Is and if that returns true, it calls f. If f returns nil error
// value, error handling will terminate and no error is thrown. The handler
// function f can process the incoming error how it wants and returning error
// value is used after the Is call.
func (o *Result) IsX(err error, f ErrFn) *Result {
	if o.Err == nil {
		return o
	}
	if errors.Is(o.Err, err) {
		o.Err = f(o.Err)
	}
	return o
}

// Is allows you to add error.Is handler to try.OutX handler chain. Internally
// Is calls errors.Is and if that returns true, it calls f. If f returns nil error
// value, error handling will terminate and no error is thrown. The handler
// function f can process the incoming error how it wants and returning error
// value is used after the Is call.
func (o *Result1[T]) IsX(err error, f ErrFn) *Result1[T] {
	o.Result.IsX(err, f)
	return o
}

// Is allows you to add error.Is handler to try.OutX handler chain. Internally
// Is calls errors.Is and if that returns true, it calls f. If f returns nil error
// value, error handling will terminate and no error is thrown. The handler
// function f can process the incoming error how it wants and returning error
// value is used after the Is call.
func (o *Result2[T, U]) IsX(err error, f ErrFn) *Result2[T, U] {
	o.Result.IsX(err, f)
	return o
}

// Handle allows you to add an error handler to try.OutX handler chain.
// Internally Handle calls f if Result.Err != nil. If a handler f returns nil
// error value, error handling will terminate and no error is thrown. The
// handler function f can process and annotate the incoming error how it wants
// and returning error value decides if error is thrown immediately.
func (o *Result) HandleX(f ErrFn) *Result {
	if o.Err == nil {
		return o
	}
	if f != nil {
		o.Err = f(o.Err)
		if o.Err != nil {
			panic(o.Err)
		}
	}
	return o
}

// Handle allows you to add an error handler to try.OutX handler chain.
// Internally Handle calls f if Result.Err != nil. If a handler f returns nil
// error value, error handling will terminate and no error is thrown. The
// handler function f can process and annotate the incoming error how it wants
// and returning error value decides if error is thrown immediately.
func (o *Result1[T]) HandleX(f ErrFn) *Result1[T] {
	o.Result.HandleX(f)
	return o
}

// Handle allows you to add an error handler to try.OutX handler chain.
// Internally Handle calls f if Result.Err != nil. If a handler f returns nil
// error value, error handling will terminate and no error is thrown. The
// handler function f can process and annotate the incoming error how it wants
// and returning error value decides if error is thrown immediately.
func (o *Result2[T, U]) HandleX(f ErrFn) *Result2[T, U] {
	o.Result.HandleX(f)
	return o
}

// Out is a helper function to call functions which returns (error) and start
// error handling with DSL. For instance, to implement same as try.To(), you
// could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v).Handle()
//
// or in some other cases some of these would be desired action:
//
//	try.Out(os.Remove(dst)).Logf("file cleanup fail")
func Out(err error) *Result {
	return &Result{Err: err}
}

// Out1 is a helper function to call functions which returns (T, error). That
// allows you to use Result1, which makes possible to
// start error handling with DSL. For instance, instead of try.To1() you could
// do the following:
//
//	d := try.Out1(os.ReadFile(filename).Handle().Val1
//
// or in some other cases, some of these would be desired action:
//
//	number := try.Out1(strconv.Atoi(str)).Def1(100).Val1
//	try.Out(os.Remove(dst)).Logf("remove")
//	try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Def2(1, 2)
//	try.Out1(strconv.Atoi(s)).Logf("not number").Def1(100).Val1
func Out1[T any](v T, err error) *Result1[T] {
	return &Result1[T]{Val1: v, Result: Result{Err: err}}
}

// Out2 is a helper function to call functions which returns (T, error). That
// allows you to use Result2, which makes possible to
// start error handling with DSL. For instance, instead of try.To2() you could
// do the following:
//
//	token := try.Out2(p.ParseUnverified(tokenStr, &customClaims{})).Handle().Val1
//
// or in some other cases, some of these would be desired action:
//
//	try.Out2(convTwoStr(s1, s2)).Logf("wrong number").Def2(1, 2)
//	try.Out2(convTwoStr(s1, s2)).Handle().Val2
func Out2[T any, U any](v1 T, v2 U, err error) *Result2[T, U] {
	return &Result2[T, U]{Val2: v2, Result1: Result1[T]{Val1: v1, Result: Result{Err: err}}}
}

func wrapStr() string {
	return ": %w"
}
