.PHONY: build dev migrate test vet verify

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

dev:
	bash -c 'source scripts/env.sh; load_dotenv .env; exec go run ./cmd/api'

migrate:
	bash -c 'source scripts/env.sh; load_dotenv .env; exec go run ./cmd/migrate'

test:
	go test ./...

vet:
	go vet ./...

verify:
	bash ./scripts/verify-foundation.sh
