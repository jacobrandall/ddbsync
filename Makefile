ifdef VERBOSE
V = -v
X = -x
else
.SILENT:
endif

all: build test cover

build:
	GOBIN=$(shell pwd)/bin go install $(V) ./...

fmt:
	go fmt $(X) ./...
	go mod tidy $(V)

test:
	mkdir -p coverage
	go test $(V) -race -cover -coverprofile=coverage/ddbsync.coverprofile ./...

cover:
	go tool cover -html=coverage/ddbsync.coverprofile -o coverage/ddbsync.html

clean:
	rm -rf bin/ coverage/
	go clean -i $(X) -cache -testcache
