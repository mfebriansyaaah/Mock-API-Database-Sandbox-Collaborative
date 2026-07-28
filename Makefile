SHELL  := powershell.exe
.SHELLFLAGS := -NoProfile -NonInteractive -Command
BIN    := Mock-API-Database-Sandbox-Collaborative.exe
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
	@echo "  make build       - build binary to $(BIN)"
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
	$$env:CGO_ENABLED='0'; $$env:APP_ENV='production'; go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .

build-linux: tidy
	$$env:CGO_ENABLED='0'; $$env:APP_ENV='production'; $$env:GOOS='linux'; $$env:GOARCH='amd64'; go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o server-linux-amd64 .

run: tidy
	go run . -credentials $(CRED)

test:
	go test -race -timeout 60s $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

clean:
	Remove-Item -Force -ErrorAction SilentlyContinue $(BIN); Remove-Item -Force -ErrorAction SilentlyContinue server-linux-amd64; Remove-Item -Force -ErrorAction SilentlyContinue server.out; Remove-Item -Force -ErrorAction SilentlyContinue server.err; echo "cleaned"

smoke: build
	$$proc = Start-Process -NoNewWindow -PassThru -FilePath ".\$(BIN)" -ArgumentList "-credentials","$(CRED)"; Start-Sleep -Seconds 2; \
	try { $$code = (Invoke-WebRequest -Uri "http://localhost:$(PORT)/hello" -UseBasicParsing -TimeoutSec 5).StatusCode; echo "SMOKE: PASS (status=$$code)" } catch { echo "SMOKE: FAILED — see output"; $$proc.Kill(); exit 1 }; \
	$$proc.Kill()
