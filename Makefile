.PHONY: all build-core build-agent build-frontend clean test dev

all: build-frontend build-core build-agent

build-core:
	cd core && CGO_ENABLED=1 go build -o ../bin/core ./cmd/core

build-agent:
	cd agent && CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent ./cmd/agent

build-agent-linux:
	cd agent && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent-linux-amd64 ./cmd/agent

build-agent-linux-arm64:
	cd agent && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent-linux-arm64 ./cmd/agent

build-agent-windows:
	cd agent && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/agent-windows-amd64.exe ./cmd/agent

build-frontend:
	cd core/web/frontend && npm install && npm run build

clean:
	rm -rf bin/
	rm -rf core/web/dist/

test:
	cd core && go test -v ./...
	cd agent && go test -v ./...

test-core:
	cd core && go test -v ./...

test-agent:
	cd agent && go test -v ./...

dev-frontend:
	cd core/web/frontend && npm run dev

tidy:
	cd core && go mod tidy
	cd agent && go mod tidy

docker-build:
	docker build -t probe-core ./core
	docker build -t probe-agent ./agent

docker-build-core:
	docker build -t probe-core ./core

docker-build-agent:
	docker build -t probe-agent ./agent

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Remove .kiro from git cache (run this if .kiro was already committed)
git-clean-kiro:
	git rm -r --cached .kiro/
	git commit -m "Remove .kiro from tracking"
