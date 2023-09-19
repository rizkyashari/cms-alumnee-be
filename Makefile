ENTRY=./cmd/main.go

.PHONY: build
build:
	docker compose build

.PHONY: migrate
migrate:
	docker compose run --service-ports --rm web go run $(ENTRY) -migrate=true

.PHONY: seed
seed:
	docker compose run --service-ports --rm web go run $(ENTRY) -seed=true

.PHONY: dev
dev:
	docker compose run --service-ports --rm web air $(ENTRY)