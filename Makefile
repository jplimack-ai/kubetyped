.PHONY: test lint tidy build cover verify-tidy

test:
	go test -race -v -count=1 ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

build:
	go build ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

verify-tidy:
	go mod tidy
	git diff --exit-code go.mod go.sum
