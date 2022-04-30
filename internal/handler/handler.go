package handler

import "runtime"

type (
	PanicHandler func(p any)
	ErrorHandler func(err error, errOut *error)
	NilHandler   func(errOut *error)
)

type Info struct {
	Any  any
	Err  error
	Out  *error
	Fmt  string
	Args []any

	NilHandler
	ErrorHandler
	PanicHandler
}

func (i Info) CallNilHandler() {
	if i.NilHandler != nil {
		i.NilHandler(i.Out)
	}
}

func (i Info) CallErrorHandler() {
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error), i.Out)
	}
}

func (i Info) CallPanicHandler() {
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	}
}

func All(info Info) {
	switch info.Any.(type) {
	case nil:
		info.CallNilHandler()
	case runtime.Error:
		info.CallPanicHandler()
	case error:
		info.CallErrorHandler()
	default:
		info.CallPanicHandler()
	}

}

func Return(info Info) {
	switch info.Any.(type) {
	case nil:
		info.NilHandler(info.Out)
	case runtime.Error:
		panic(info.Any)
	case error:
		info.ErrorHandler(info.Any.(error), info.Out)
	default:
		panic(info.Any)
	}

}
