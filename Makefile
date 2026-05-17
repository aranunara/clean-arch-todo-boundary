MIGRATE_DATABASE_URL := postgres://todo:todo@db:5432/todo?sslmode=disable
GO_CACHE ?= /tmp/clean-arch-todo-boundary-go-build
GOLANGCI_LINT_CACHE_DIR ?= /tmp/clean-arch-todo-boundary-golangci-lint

.PHONY: up up-all up-d down down-all down-v logs ports cli try migrate-up migrate-down migrate-create test test-pg-integration lint ci run

up:
	$(MAKE) up-all

up-all:
	docker compose up -d --build
	$(MAKE) ports

up-d:
	$(MAKE) up-all

down:
	$(MAKE) down-all

down-all:
	docker compose down

down-v:
	docker compose down -v

logs:
	docker compose logs -f api db migrate

ports:
	@api_port="$$(docker compose port api 8080 | sed 's/.*://')"; \
	db_port="$$(docker compose port db 5432 | sed 's/.*://')"; \
	if [ -z "$$api_port" ]; then echo "api is not running. run: make up-all"; exit 1; fi; \
	if [ -z "$$db_port" ]; then echo "db is not running. run: make up-all"; exit 1; fi; \
	echo "API: http://127.0.0.1:$$api_port"; \
	echo "DB:  127.0.0.1:$$db_port"

try:
	go run ./cmd/todoctl create "層の違いをメモする"
	go run ./cmd/todoctl list

cli:
	go run ./cmd/todoctl $(args)

migrate-up:
	docker compose run --rm migrate -path /migrations -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	docker compose run --rm migrate -path /migrations -database "$(MIGRATE_DATABASE_URL)" down 1

migrate-create:
	docker compose run --rm migrate create -ext sql -dir /migrations -seq $(name)

test:
	GOCACHE=$(GO_CACHE) go test ./...

test-pg-integration:
	GOCACHE=$(GO_CACHE) go test -tags=pg_integration ./...

lint:
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE_DIR) GOCACHE=$(GO_CACHE) golangci-lint run

ci: test lint

run:
	go run ./cmd/server
