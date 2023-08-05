package main

import (
	"errors"
	"os"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"golang.org/x/exp/slog"
)

var (
	opts = slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	errAddNode = errors.New("add node error")
	myErr error = nil

	logger *slog.Logger
)

func Init() {
	textHandler := opts.NewTextHandler(os.Stdout)
	logger = slog.New(textHandler)
	slog.SetDefault(logger)
	if *isErr {
		myErr = errAddNode
	}
}

func doMainAll()  {
	logger.Info("=== 1. preferred successful status output ===")
	doMain1()
	logger.Info("=== 2. NilThen and try.To successful status ===")
	doMain2()
	logger.Info("=== 3. NilThen and try.Out successful status ===")
	doMain3()

	logger.Info("=== ERROR status versions ===")
	myErr = errAddNode
	logger.Info("=== 1. preferred successful status output ===")
	doMain1()
	logger.Info("=== 2. NilThen and try.To successful status ===")
	doMain2()
	logger.Info("=== 3. NilThen and try.Out successful status ===")
	doMain3()
}
func doMain3() {
	defer err2.Catch("CATCH")
	logger.Debug("3: ADD node")
	defer NilThen(func() {
		logger.Debug("3: add node successful")
	})
	try.Out(AddNode()).Logf("3: no error handling, only logging")
}

func doMain2() {
	defer err2.Catch("CATCH")
	logger.Debug("2: ADD node")
	defer NilThen(func() {
		logger.Debug("2: add node successful")
	})

	try.To(AddNode())
}

func doMain1() {
	defer err2.Catch("CATCH")
	logger.Debug("1: ADD node")

	try.To(AddNode())
	logger.Debug("1: add node successful")
}

func AddNode() error { return myErr }

func NilThen(fn func()) {
	if r := recover(); r != nil {
		panic(r)
	}
	fn()
}

