# Detect OS for platform-specific variables.
ifeq ($(OS),Windows_NT)
  BIN      := Mock-API-Database-Sandbox-Collaborative.exe
  RM       := del /f /q
  NULLDEV  := nul
  ECHO     := echo
else
  BIN      := Mock-API-Database-Sandbox-Collaborative
  RM       := rm -f
  NULLDEV  := /dev/null
  ECHO     := echo
endif

PKG    := ./...
PORT   := 8080
CRED   := Service_Account_Key.json

# Go build flags for production
GOFLAGS  := -trimpath
LDFLAGS  := -s -w

.PHONY: help tidy build build-linux run test vet fmt clean smoke

help:
	@echo "Available targets:"
	@echo "  make tidy        - go mod tidy"
	@echo "  make build       - build binary ($(BIN))"
	@echo "  make build-linux - cross-compile for linux/amd64"
	@echo "  make run         - run from source"
	@echo "  make test        - run unit tests"
	@echo "  make vet         - go vet"
	@echo "  make fmt         - gofmt -w"
	@echo "  make smoke       - run binary, hit /hello, then stop"
	@echo "  make clean       - remove binary and logs"

tidy:
	go mod tidy

build: tidy
	CGO_ENABLED=0 APP_ENV=production go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .

build-linux: tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 APP_ENV=production go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o server-linux-amd64 .

run: tidy
	go run . -credentials $(CRED)

test:
	go test -race -timeout 60s $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

clean:
	$(RM) $(BIN) server-linux-amd64 server.out server.err 2>$(NULLDEV)
	@$(ECHO) "cleaned"

smoke: build
	@echo "Starting $(BIN)..."
	@./$(BIN) -credentials $(CRED) & \
	  PID=$$!; \
	  sleep 2; \
	  STATUS=$$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$(PORT)/hello 2>/dev/null || echo "000"); \
	  kill $$PID 2>/dev/null; \
	  if [ "$$STATUS" = "200" ]; then \
	    echo "SMOKE: PASS (status=$$STATUS)"; \
	  else \
	    echo "SMOKE: FAILED (status=$$STATUS)"; \
	    exit 1; \
	  fi
