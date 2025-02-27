package try

import (
	"errors"
	"fmt"

	"github.com/lainio/err2/internal/handler"
)

type (
	// ErrFn is function type for [try.Out] handlers.
	ErrFn = handler.ErrorFn

	// Result is the base of our error handling language for [try.Out] functions.
	Result struct {
		// Err holds the error value returned from try.Out function result.
		Err error
	}

	// Result1 is the base of our error handling DSL for [try.Out1] functions.
	Result1[T any] struct {
		// Val1 holds the first value returned from try.Out1 function result.
		Val1 T
		Result
	}

	// Result2 is the base of our error handling DSL for [try.Out2] functions.
	Result2[T any, U any] struct {
		// Val2 holds the first value returned from try.Out2 function result.
		Val2 U
		Result1[T]
	}
)

// Logf prints a log line to pre-set logging stream [err2.SetLogWriter]
// if the current [Result.Err] != nil. Logf follows Printf formatting logic. The
// current error value will be added at the end of the logline with ": %v\n",
// err. For example, the line:
//
//	try.Out(server.Send(status)).Logf("error sending response")
//
// would print the logline:
//
//	error sending response: UDP not listening
func (o *Result) Logf(a ...any) *Result {
	if o.Err == nil {
		return o
	}
	return o.logf(logfFrameLvl, a)
}

// Logf prints a log line to pre-set logging stream [err2.SetLogWriter]
// if the current [Result.Err] != nil. Logf follows Printf formatting logic. The
// current error value will be added at the end of the logline with ": %v\n",
// err. For example, the line:
//
//	try.Out(server.Send(status)).Logf("error sending response")
//
// would print the logline:
//
//	error sending response: UDP not listening
func (o *Result1[T]) Logf(a ...any) *Result1[T] {
	if o.Err == nil {
		return o
	}
	o.Result.logf(logfFrameLvl, a)
	return o
}

// Logf prints a log line to pre-set logging stream [err2.SetLogWriter]
// if the current [Result.Err] != nil. Logf follows Printf formatting logic. The
// current error value will be added at the end of the logline with ": %v\n",
// err. For example, the line:
//
//	try.Out(server.Send(status)).Logf("error sending response")
//
// would print the logline:
//
//	error sending response: UDP not listening
func (o *Result2[T, U]) Logf(a ...any) *Result2[T, U] {
	if o.Err == nil {
		return o
	}
	o.Result.logf(logfFrameLvl, a)
	return o
}

// Handle allows you to add an error handler to [try.Out] handler chain. Handle
// is a general purpose error handling function. It can handle several error
// handling cases:
//   - if no argument is given and .Err != nil, it throws an error value immediately
//   - if two arguments (errTarget, ErrFn) and Is(.Err, errTarget) [ErrFn] is called
//   - if first argument is (string) and .Err != nil the error value is annotated and thrown
//   - if first argument is (ErrFn) and .Err != nil, it calls [ErrFn]
//
// The handler function [ErrFn] can process and annotate the incoming error how
// it wants and returning error value decides if error is thrown. Handle
// annotates and throws an error immediately i.e. terminates error handling DSL
// chain if [Result.Err] != nil. Handle supports error annotation similarly as
// [fmt.Errorf].
//
// For instance, to implement same as [try.To], you could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v)).Handle()
func (o *Result) Handle(a ...any) *Result {
	if o.Err == nil {
		return o
	}
	o.transportErr(a)
	return o
}

func (o *Result) transportErr(a []any) {
	noArguments := len(a) == 0
	if noArguments {
		panic(o.Err)
	}

	switch f := a[0].(type) {
	case string:
		o.Err = fmt.Errorf(f+wrapStr(), append(a[1:], o.Err)...)
	case ErrFn:
		o.Err = f(o.Err)
	case error:
		if len(a) == 2 {
			hfn, haveHandlerFn := a[1].(ErrFn)
			if haveHandlerFn {
				if errors.Is(o.Err, f) {
					o.Err = hfn(o.Err)
				}
			}
		}
	}

	// some of the handler functions might reset the error value.
	if o.Err != nil {
		panic(o.Err)
	}
}

// Handle allows you to add an error handler to [try.Out] handler chain. Handle
// is a general purpose error handling function. It can handle several error
// handling cases:
//   - if no argument is given and .Err != nil, it throws an error value immediately
//   - if two arguments (errTarget, [ErrFn]) and Is(.Err, errTarget) [ErrFn] is called
//   - if first argument is (string) and .Err != nil the error value is annotated and thrown
//   - if first argument is [ErrFn] and .Err != nil, it calls [ErrFn]
//
// The handler function [ErrFn] can process and annotate the incoming error how
// it wants and returning error value decides if error is thrown. Handle
// annotates and throws an error immediately i.e. terminates error handling DSL
// chain if [Result.Err] != nil. Handle supports error annotation similarly as
// [fmt.Errorf].
//
// For instance, to implement same as [try.To], you could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v)).Handle()
func (o *Result1[T]) Handle(a ...any) *Result1[T] {
	o.Result.Handle(a...)
	return o
}

// Handle allows you to add an error handler to [try.Out] handler chain. Handle
// is a general purpose error handling function. It can handle several error
// handling cases:
//   - if no argument is given and .Err != nil, it throws an error value immediately
//   - if two arguments (errTarget, [ErrFn]) and Is(.Err, errTarget) ErrFn is called
//   - if first argument is (string) and .Err != nil the error value is annotated and thrown
//   - if first argument is [ErrFn] and .Err != nil, it calls [ErrFn]
//
// The handler function [ErrFn] can process and annotate the incoming error how
// it wants and returning error value decides if error is thrown. Handle
// annotates and throws an error immediately i.e. terminates error handling DSL
// chain if [Result.Err] != nil. Handle supports error annotation similarly as
// [fmt.Errorf].
//
// For instance, to implement same as [try.To], you could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v)).Handle()
func (o *Result2[T, U]) Handle(a ...any) *Result2[T, U] {
	o.Result.Handle(a...)
	return o
}

// Catch catches the error and sets [Result1.Val1] if given. The value is used
// only in the case if [Result1.Err] != nil. Catch returns the Val1 in all cases.
func (o *Result1[T]) Catch(v ...T) T {
	if o.Err != nil && len(v) == 1 {
		o.Val1 = v[0]
	}
	return o.Val1
}

// Catch catches the error and sets [Result2.Val1] [Result2.Val2] if given. The
// value(s) is used in the case of [Result2.Err] != nil. Catch returns the [Val1] and
// [Val2] in all cases. In case you want to set only [Val2]'s default value, use
// [Def2] before Catch call.
func (o *Result2[T, U]) Catch(a ...any) (T, U) {
	if o.Err != nil {
		switch len(a) {
		case 2:
			o.Val2 = a[1].(U)
			fallthrough
		case 1:
			o.Val1 = a[0].(T)
		}
	}
	return o.Val1, o.Val2
}

// Def1 sets default value for [Result.Val1.] The value is returned in case of
// [Result.Err] != nil.
func (o *Result1[T]) Def1(v T) *Result1[T] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	return o
}

// Def1 sets default value for [Result.Val1]. The value is returned in case of
// [Result.Err] != nil.
func (o *Result2[T, U]) Def1(v T) *Result2[T, U] {
	if o.Err == nil {
		return o
	}
	o.Val1 = v
	return o
}

// Def2 sets default value for [Result.Val2]. The value is returned in case of
// [Result.Err] != nil.
func (o *Result2[T, U]) Def2(v2 U) *Result2[T, U] {
	if o.Err == nil {
		return o
	}
	o.Val2 = v2
	return o
}

// Out is a helper function to call functions which returns (error) and start
// error handling with DSL. For instance, to implement same as [try.To], you
// could do the following:
//
//	d := try.Out(json.Unmarshal(b, &v)).Handle()
//
// or in some other cases some of these would be desired action:
//
//	try.Out(os.Remove(dst)).Logf("file cleanup fail")
func Out(err error) *Result {
	return &Result{Err: err}
}

// Out1 is a helper function to call functions which returns (T, error). That
// allows you to use [Result1], which makes possible to
// start error handling with DSL. For instance, instead of [try.To1] you could
// do the following:
//
//	d := try.Out1(os.ReadFile(filename).Handle().Val1
//
// or in some other cases, some of these would be desired action:
//
//	number := try.Out1(strconv.Atoi(str)).Catch(100)
//	x := try.Out1(strconv.Atoi(s)).Logf("not number").Catch(100)
func Out1[T any](v T, err error) *Result1[T] {
	return &Result1[T]{Val1: v, Result: Result{Err: err}}
}

// Out2 is a helper function to call functions which returns (T, error). That
// allows you to use [Result2], which makes possible to
// start error handling with DSL. For instance, instead of [try.To2] you could
// do the following:
//
//	token := try.Out2(p.ParseUnverified(tokenStr, &customClaims{})).Handle().Val1
//
// or in some other cases, some of these would be desired action:
//
//	x, y := try.Out2(convTwoStr(s1, s2)).Logf("bad number").Catch(1, 2)
//	y := try.Out2(convTwoStr(s1, s2)).Handle().Val2
func Out2[T any, U any](v1 T, v2 U, err error) *Result2[T, U] {
	return &Result2[T, U]{
		Val2:    v2,
		Result1: Result1[T]{Val1: v1, Result: Result{Err: err}},
	}
}

func wrapStr() string {
	return handler.WrapError
}

func (o *Result) logf(lvl int, a []any) *Result {
	s := o.Err.Error()
	if len(a) != 0 {
		f, isFormat := a[0].(string)
		if isFormat {
			s = fmt.Sprintf(f+": %v", append(a[1:], o.Err)...)
		}
	}
	_ = handler.LogOutput(lvl, s)
	return o
}

const (
	logfFrameLvl = 4
)
