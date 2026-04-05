.PHONY: test test-short vet build tidy coverage lint

# Default target
all: tidy build vet test-short

## Fetch / tidy dependencies
tidy:
	go mod tidy

## Build all packages (catches compile errors)
build:
	go build ./...

## Run all tests with race detector
test:
	go test ./... -v -race -timeout 120s

## Run fast subset (no I/O-heavy tests)
test-short:
	go test ./... -short -timeout 60s

## Run go vet across all packages
vet:
	go vet ./...

## Generate HTML coverage report
coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run golangci-lint (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

## Clean build artefacts
clean:
	rm -f coverage.out coverage.html
