# Original from github.com/pkg/errors

PKGS := github.com/lainio/err2
SRCDIRS := $(shell go list -f '{{.Dir}}' $(PKGS))
GO := go

check: test vet gofmt misspell unconvert staticcheck ineffassign unparam

test:
	$(GO) test $(PKGS)

bench:
	$(GO) test -bench=. $(PKGS)

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

gen:
	go run cmd/main.go -name=Int -type=int > int.go
	go run cmd/main.go -name=String -type=string > string.go
	go run cmd/main.go -name=File -type=*os.File > file.go
	go run cmd/main.go -name=Bytes -type=[]byte > bytes.go
	go run cmd/main.go -name=Byte -type=byte > byte.go
	go run cmd/main.go -name=Strings -type=[]string > strings.go
	go run cmd/main.go -name=Ints -type=[]int > ints.go
	go run cmd/main.go -name=Bool -type=bool > bool.go
	go run cmd/main.go -name=Bools -type=[]bool > bools.go
	goimports -l -w .