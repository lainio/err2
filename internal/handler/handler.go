package handler

import "runtime"

type (
	PanicHandler func(p any)
	ErrorHandler func(err error)
	NilHandler   func()
)

type Info struct {
	Any any

	NilHandler
	ErrorHandler
	PanicHandler
}

func (i Info) CallNilHandler() {
	if i.NilHandler != nil {
		i.NilHandler()
	}
}

func (i Info) CallErrorHandler() {
	if i.ErrorHandler != nil {
		i.ErrorHandler(i.Any.(error))
	}
}

func (i Info) CallPanicHandler() {
	if i.PanicHandler != nil {
		i.PanicHandler(i.Any)
	} else {
		panic(i.Any)
	}
}

func Process(info Info) {
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
