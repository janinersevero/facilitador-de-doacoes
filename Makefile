run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

db/up:
	docker compose up -d

db/down:
	docker compose down

tidy:
	go mod tidy
