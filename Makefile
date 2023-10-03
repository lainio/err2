# Original from github.com/pkg/errors

PKG_ERR2 := github.com/lainio/err2
PKG_ASSERT := github.com/lainio/err2/assert
PKG_TRY := github.com/lainio/err2/try
PKG_DEBUG := github.com/lainio/err2/internal/debug
PKG_HANDLER := github.com/lainio/err2/internal/handler
PKG_STR := github.com/lainio/err2/internal/str
PKG_X := github.com/lainio/err2/internal/x
PKGS := $(PKG_ERR2) $(PKG_ASSERT) $(PKG_TRY) $(PKG_DEBUG) $(PKG_HANDLER) $(PKG_STR) $(PKG_X)

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

GO ?= go
TEST_ARGS ?= -benchmem

# GO ?= go1.20rc2

check: lint vet gofmt test

test_err2:
	$(GO) test $(TEST_ARGS) $(PKG_ERR2)

test_assert:
	$(GO) test $(TEST_ARGS) $(PKG_ASSERT)

test_try:
	$(GO) test $(TEST_ARGS) $(PKG_TRY)

test_debug:
	$(GO) test $(TEST_ARGS) $(PKG_DEBUG)

test_handler:
	$(GO) test $(TEST_ARGS) $(PKG_HANDLER)

test_str:
	$(GO) test $(TEST_ARGS) $(PKG_STR)

test_x:
	$(GO) test $(TEST_ARGS) $(PKG_X)

testv:
	$(GO) test -v $(TEST_ARGS) $(PKGS)

test:
	$(GO) test $(TEST_ARGS) $(PKGS)

inline_err2:
	$(GO) test -c -gcflags=-m=2 $(PKG_ERR2) 2>&1 | ag 'inlin' 

tinline_err2:
	$(GO) test -c -gcflags=-m=2 $(PKG_ERR2) 2>&1 | ag 'inlin' | ag 'err2_test'

inline_handler:
	$(GO) test -c -gcflags=-m=2 $(PKG_HANDLER) 2>&1 | ag 'inlin' 

tinline_handler:
	$(GO) test -c -gcflags=-m=2 $(PKG_HANDLER) 2>&1 | ag 'inlin'

bench:
	$(GO) test $(TEST_ARGS) -bench=. $(PKGS)

bench_goid:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkGoid' $(PKG_ASSERT)

bench_reca:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkRecursion.*' $(PKG_ERR2)

bench_out:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkTryOut_.*' $(PKG_ERR2)

bench_go:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkTry_StringGenerics' $(PKG_ERR2)

bench_that:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkThat.*' $(PKG_ASSERT)

bench_copy:
	$(GO) test $(TEST_ARGS) -bench='Benchmark_CopyBuffer' $(PKG_TRY)

bench_rech:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkRecursionWithTryAnd_HeavyPtrPtr_Defer' $(PKG_ERR2)

bench_rece:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkRecursionWithTryAnd_Empty_Defer' $(PKG_ERR2)

bench_rec:
	$(GO) test $(TEST_ARGS) -bench='BenchmarkRecursionWithOldErrorIfCheckAnd_Defer' $(PKG_ERR2)

bench_err2:
	$(GO) test $(TEST_ARGS) -bench=. $(PKG_ERR2)

bench_assert:
	$(GO) test $(TEST_ARGS) -bench=. $(PKG_ASSERT)

bench_str:
	$(GO) test $(TEST_ARGS) -bench=. $(PKG_STR)

bench_x:
	$(GO) test $(TEST_ARGS) -bench=. $(PKG_X)

vet: | test
	$(GO) vet $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

godoc:
	@GO111MODULE=off godoc -http=0.0.0.0:6060

test_cov_out:
	go test -p 1 -failfast \
		-coverpkg=$(PKG_ERR2)/... \
		-coverprofile=coverage.txt  \
		-covermode=atomic \
		./...

test_cov: test_cov_out
	go tool cover -html=coverage.txt -o=coverage.html
	firefox ./coverage.html 1>&- 2>&-  &

lint:
	@golangci-lint run

.PHONY:	check

