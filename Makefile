COVERAGEDIR = coverage
ifdef CIRCLE_ARTIFACTS
  COVERAGEDIR = $(CIRCLE_ARTIFACTS)
endif

ifdef VERBOSE
V = -v
else
.SILENT:
endif

all: build test cover

install-deps:
	glide install
build:
	mkdir -p bin
	go build -v -o bin/ddbsync
fmt:
	go fmt ./...
test:
	mkdir -p coverage
	go test $(V) ./ -race -cover -coverprofile=$(COVERAGEDIR)/ddbsync.coverprofile
cover:
	go tool cover -html=$(COVERAGEDIR)/ddbsync.coverprofile -o $(COVERAGEDIR)/ddbsync.html
coveralls:
	gover $(COVERAGEDIR) $(COVERAGEDIR)/coveralls.coverprofile
	goveralls -coverprofile=$(COVERAGEDIR)/coveralls.coverprofile -service=circle-ci -repotoken=$(COVERALLS_TOKEN)
clean:
	go clean
	rm -f bin/ddbsync
	rm -rf coverage/
