ENTRY=./cmd/main.go

.PHONY: migrate
migrate:
	docker compose run --service-ports --rm web go run $(ENTRY) -migrate=true

.PHONY: dev
dev:
	docker compose run --service-ports --rm web air $(ENTRY)