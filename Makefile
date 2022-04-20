# Original from github.com/pkg/errors

PKG1 := github.com/lainio/err2
PKG2 := github.com/lainio/err2/assert
PKG3 := github.com/lainio/err2/try
PKG4 := github.com/lainio/err2/internal/debug
PKGS := $(PKG1) $(PKG2) $(PKG3) $(PKG4)

SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))

GO := go
# GO := go1.18beta2

check: test vet gofmt misspell unconvert staticcheck ineffassign unparam

test1:
	$(GO) test $(PKG1)

test2:
	$(GO) test $(PKG2)

test3:
	$(GO) test $(PKG3)

test4:
	$(GO) test $(PKG4)

test:
	$(GO) test $(PKGS)

bench:
	$(GO) test -bench=. $(PKGS)

bench1:
	$(GO) test -bench=. $(PKG1)

bench2:
	$(GO) test -bench=. $(PKG2)

vet: | test
	$(GO) vet $(PKGS)

staticcheck:
	$(GO) get honnef.co/go/tools/cmd/staticcheck
	staticcheck -checks all $(PKGS)

misspell:
	$(GO) get github.com/client9/misspell/cmd/misspell
	misspell \
		-locale GB \
		-error \
		*.md *.go

unconvert:
	$(GO) get github.com/mdempsky/unconvert
	unconvert -v $(PKGS)

ineffassign:
	$(GO) get github.com/gordonklaus/ineffassign
	find $(SRCDIRS) -name '*.go' | xargs ineffassign

pedantic: check errcheck

unparam:
	$(GO) get mvdan.cc/unparam
	unparam ./...

errcheck:
	$(GO) get github.com/kisielk/errcheck
	errcheck $(PKGS)

gofmt:
	@echo Checking code is gofmted
	@test -z "$(shell gofmt -s -l -d -e $(SRCDIRS) | tee /dev/stderr)"

godoc:
	@GO111MODULE=off godoc -http=0.0.0.0:6060

