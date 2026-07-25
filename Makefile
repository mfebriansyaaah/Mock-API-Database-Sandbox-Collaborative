SHELL  := /bin/bash
.SHELLFLAGS := -euo pipefail -c
BIN    := server
PKG    := ./...
PORT   := 8080

# Go build flags for production
GOFLAGS  := -trimpath
LDFLAGS  := -s -w

.PHONY: help tidy build build-linux run test vet fmt clean smoke

help:
	@echo "Available targets:"
	@echo "  make tidy        - go mod tidy"
	@echo "  make build       - build binary to $(BIN)"
	@echo "  make build-linux - cross-compile for linux/amd64"
	@echo "  make run         - run from source (loads .env automatically)"
	@echo "  make test        - run unit tests"
	@echo "  make vet         - go vet"
	@echo "  make fmt         - gofmt -w"
	@echo "  make smoke       - run binary, hit /hello, then stop"
	@echo "  make clean       - remove binary and logs"

tidy:
	go mod tidy

build: tidy
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .

build-linux: tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN)-linux-amd64 .

run: tidy
	go run .

test:
	go test -race -timeout 60s $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

clean:
	rm -f $(BIN) $(BIN)-linux-amd64 server.out server.err
	@echo "cleaned"

smoke: build
	@pkill -f './$(BIN)' 2>/dev/null || true
	./$(BIN) > server.out 2> server.err &
	@SERVER_PID=$$!; sleep 2; \
	if curl -sf -o /dev/null -w '%{http_code}' "http://localhost:$(PORT)/hello" | grep -qE '^(200|201)$$'; then \
		echo "SMOKE: PASS (status=$$(curl -s -o /dev/null -w '%{http_code}' http://localhost:$(PORT)/hello))"; \
	else \
		echo "SMOKE: FAILED — see server.err"; cat server.err; \
	fi
	@pkill -f './$(BIN)' 2>/dev/null || true
