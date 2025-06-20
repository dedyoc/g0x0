.PHONY: build run dev docker-up docker-down clean

build:
	go build -o bin/g0x0 ./cmd/server

run: build
	./bin/g0x0

dev:
	go run ./cmd/server

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/ uploads/

test:
	go test -v ./...

deps:
	go mod download
	go mod tidy
