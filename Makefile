SHELL := powershell.exe
.SHELLFLAGS := -NoProfile -Command
BIN  := server.exe
PKG  := ./...

.PHONY: help tidy build run test vet fmt clean smoke

help:
	@echo "Available targets:"
	@echo "  make tidy   - go mod tidy"
	@echo "  make build  - build binary to $(BIN)"
	@echo "  make run    - run from source (loads .env automatically)"
	@echo "  make test   - run unit tests"
	@echo "  make vet    - go vet"
	@echo "  make fmt    - gofmt -w"
	@echo "  make smoke  - run binary, hit /hello, then stop"
	@echo "  make clean  - remove binary and logs"

tidy:
	go mod tidy

build: tidy
	go build -o $(BIN) .

run: tidy
	go run .

test:
	go test -race -timeout 60s $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -w .

clean:
	- if (Test-Path $(BIN)) { Remove-Item -Force $(BIN) }
	- if (Test-Path server.out) { Remove-Item -Force server.out }
	- if (Test-Path server.err) { Remove-Item -Force server.err }
	@echo "cleaned"

smoke: build
	@if (Get-Process server -ErrorAction SilentlyContinue) { Get-Process server | Stop-Process -Force }
	Start-Process -FilePath .\$(BIN) -RedirectStandardOutput server.out -RedirectStandardError server.err
	Start-Sleep -Seconds 2
	try { $r = Invoke-WebRequest -Uri "http://localhost:8080/hello" -UseBasicParsing; Write-Host "SMOKE: status=$($r.StatusCode) body=$($r.Content)" } catch { Write-Host "SMOKE: FAILED - $_" }
	Get-Process server | Stop-Process -Force
