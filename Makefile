COVERAGEDIR = coverage
ifdef CIRCLE_ARTIFACTS
  COVERAGEDIR = $(CIRCLE_ARTIFACTS)
endif

ifdef VERBOSE
V = -v
X = -x
else
.SILENT:
endif

all: build test cover

build:
	GOBIN=$(shell pwd)/bin go install ./...

fmt:
	go fmt $(X) ./...
	go mod tidy $(V)

test:
	mkdir -p coverage
	go test $(V) -race -cover -coverprofile=$(COVERAGEDIR)/ddbsync.coverprofile ./...

cover:
	go tool cover -html=$(COVERAGEDIR)/ddbsync.coverprofile -o $(COVERAGEDIR)/ddbsync.html

clean:
	rm -rf bin/ coverage/
	go clean -i $(X) -cache -testcache
