.PHONY: run test clean

.DEFAULT_GOAL := run

build:
	go mod download && go build -o ./.bin ./cmd/main.go

run:
	go mod download && go run ./cmd/main.go

test:
	go mod download && go test ./...

clean:
	rm -f ./.bin
