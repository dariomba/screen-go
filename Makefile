.PHONY: generate run test lint migrate

# ── Code generation ───────────────────────────────────────────────────────────
generate:
	go generate ./...

# ── Development ───────────────────────────────────────────────────────────────
run:
	go run ./cmd/

test:
	go test ./... -race -count=1 -timeout=60s

lint:
	golangci-lint run ./...

# ── Database ──────────────────────────────────────────────────────────────────
migrate:
	migrate \
		-path internal/store/migrations \
		-database "$$DATABASE_URL" \
		up

migrate-down:
	migrate \
		-path internal/store/migrations \
		-database "$$DATABASE_URL" \
		down 1