build:
	docker compose up -d
	go mod tidy
	go build -o bin/ cmd/main.go

run:
	docker compose up -d
	./bin/main