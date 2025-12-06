APP_NAME=podvibe
CMD=./cmd/api

.PHONY: build test fmt run

build:
	go build -o bin/$(APP_NAME) $(CMD)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

run:
	go run $(CMD)
