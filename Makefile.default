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
dev: web-air wait-for-db strapi-dev

.PHONY: web-air
web-air:
	docker compose run --service-ports --rm web air $(ENTRY)

.PHONY: wait-for-db
wait-for-db:
	docker-compose run --rm web sh -c 'until nc -z db 5432; do sleep 1; done'

.PHONY: strapi-dev
strapi-dev: wait-for-db
	docker-compose run --service-ports --rm strapi yarn develop

.PHONY: deploy
deploy: build migrate seed wait-for-db 
	docker-compose run --service-ports --rm strapi yarn build
	docker-compose run --service-ports strapi yarn strapi
	docker compose run --service-ports --rm go run $(ENTRY)
