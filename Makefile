.PHONY: all build-core build-agent clean test

all: build-core build-agent

build-core:
	cd core && CGO_ENABLED=1 go build -o ../bin/core ./cmd/core

build-agent:
	cd agent && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent ./cmd/agent

build-agent-linux:
	cd agent && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent-linux-amd64 ./cmd/agent

build-agent-windows:
	cd agent && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent-windows-amd64.exe ./cmd/agent

clean:
	rm -rf bin/

test:
	cd core && go test -v ./...
	cd agent && go test -v ./...

test-core:
	cd core && go test -v ./...

test-agent:
	cd agent && go test -v ./...
