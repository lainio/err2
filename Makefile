# Original from github.com/pkg/errors

PKG1 := github.com/lainio/err2
PKG2 := github.com/lainio/err2/assert
PKG3 := github.com/lainio/err2/try
PKG4 := github.com/lainio/err2/internal/debug
PKG5 := github.com/lainio/err2/internal/str
PKG6 := github.com/lainio/err2/internal/x
PKGS := $(PKG1) $(PKG2) $(PKG3) $(PKG4) $(PKG5) $(PKG6)

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

GO ?= go
# GO ?= go1.20rc2

check: lint vet gofmt test

test1:
	$(GO) test $(PKG1)

test2:
	$(GO) test $(PKG2)

test3:
	$(GO) test $(PKG3)

test4:
	$(GO) test $(PKG4)

test5:
	$(GO) test $(PKG5)

test6:
	$(GO) test $(PKG6)

test:
	$(GO) test $(PKGS)

bench:
	$(GO) test -bench=. $(PKGS)

bench_go:
	$(GO) test -bench='BenchmarkTry_StringGenerics' $(PKG1)

bench_arec:
	$(GO) test -bench='BenchmarkRecursion.*' $(PKG1)

bench_that:
	$(GO) test -bench='BenchmarkThat.*' $(PKG2)

bench_copy:
	$(GO) test -bench='Benchmark_CopyBuffer' $(PKG3)

bench_rec:
	$(GO) test -bench='BenchmarkRecursionWithOldErrorIfCheckAnd_Defer' $(PKG1)

bench1:
	$(GO) test -bench=. $(PKG1)

bench2:
	$(GO) test -bench=. $(PKG2)

bench5:
	$(GO) test -bench=. $(PKG5)

bench6:
	$(GO) test -bench=. $(PKG6)

vet: | test
	$(GO) vet $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

godoc:
	@GO111MODULE=off godoc -http=0.0.0.0:6060

test_cov_out:
	go test -p 1 -failfast \
		-coverpkg=$(PKG1)/... \
		-coverprofile=coverage.txt  \
		-covermode=atomic \
		./...

test_cov: test_cov_out
	go tool cover -html=coverage.txt -o=coverage.html
	firefox ./coverage.html 1>&- 2>&-  &

lint:
	@golangci-lint run

