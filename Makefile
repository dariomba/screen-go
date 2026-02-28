.PHONY: generate build-migrate migrate migrate-down docker-db-start docker-db-stop docker-db-logs run test lint

# ── Code generation ───────────────────────────────────────────────────────────
generate:
	go generate ./...

# ── Database ────────────────────────────────────────────────────────────────
DB_NAME ?= screengodb
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_CONTAINER_NAME ?= screen-go-db
DB_VOLUME ?= screen-go-db-data
POSTGRES_VERSION ?= 18
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ── Migrate tool ────────────────────────────────────────────────────────────
MIGRATE_VERSION := v4.19.1
MIGRATE_PKG := github.com/golang-migrate/migrate/v4/cmd/migrate
MIGRATE_BIN := ./bin/migrate

# Build migrate if missing
$(MIGRATE_BIN):
	@mkdir -p bin
	GOBIN=$(PWD)/bin \
	go install -tags "postgres" $(MIGRATE_PKG)@$(MIGRATE_VERSION)

build-migrate: $(MIGRATE_BIN)

migrate: build-migrate
	$(MIGRATE_BIN) -path tools/migrate -database "$(DATABASE_URL)" up

migrate-down: build-migrate
	$(MIGRATE_BIN) -path tools/migrate -database "$(DATABASE_URL)" down 1

# ── Docker DB ───────────────────────────────────────────────────────────────
docker-db-start:
	@docker volume create $(DB_VOLUME) 2>/dev/null || true
	@docker rm -f $(DB_CONTAINER_NAME) 2>/dev/null || true
	docker run -d \
		--name $(DB_CONTAINER_NAME) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-v $(DB_CONTAINER_NAME)-data:/var/lib/postgresql \
		postgres:$(POSTGRES_VERSION)

docker-db-stop:
	-docker stop $(DB_CONTAINER_NAME) 2>/dev/null || true
	-docker rm $(DB_CONTAINER_NAME) 2>/dev/null || true

docker-db-logs:
	docker logs -f $(DB_CONTAINER_NAME)

# ── Development ─────────────────────────────────────────────────────────────
run: docker-db-start
	go run main.go serve

test:
	go test ./... -race -count=1 -timeout=60s

lint:
	golangci-lint run ./...