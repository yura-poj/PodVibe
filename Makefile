APP_NAME=podvibe
CMD=./cmd/api

.PHONY: build test fmt run docker docker-up docker-down

build:
	go build -o bin/$(APP_NAME) $(CMD)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

run:
	go run $(CMD)

docker:
	docker build -t $(APP_NAME):latest .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
