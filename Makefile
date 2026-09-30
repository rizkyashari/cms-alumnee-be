ENTRY=./cmd/main.go

.PHONY: build
build:
	podman compose build

.PHONY: migrate
migrate:
	podman compose run --service-ports --rm web go run $(ENTRY) -migrate=true

.PHONY: seed
seed:
	podman compose run --service-ports --rm web go run $(ENTRY) -seed=true

.PHONY: dev
dev: web-air wait-for-db strapi-dev

.PHONY: web-air
web-air:
	podman compose run --service-ports --rm web air $(ENTRY)

.PHONY: wait-for-db
wait-for-db:
	podman compose run --rm web sh -c 'until nc -z db 5432; do sleep 1; done'

.PHONY: strapi-dev
strapi-dev: wait-for-db
	podman compose run --service-ports --rm strapi yarn develop

.PHONY: deploy
deploy: build migrate seed wait-for-db 
	podman compose run --service-ports --rm strapi yarn build
	podman compose run --service-ports strapi yarn strapi
	podman compose run --service-ports --rm go run $(ENTRY)
