.PHONY: all build test lint docker-up docker-down migrate migrate-down clean

all: lint build test

# Build all services
build: build-shared build-api build-actuary build-crawler build-llm-gateway

build-shared:
	cd shared && go build ./...

build-api:
	cd services/api-server && go build ./cmd/main.go

build-actuary:
	cd services/actuarial-engine && go build ./cmd/main.go

build-crawler:
	cd services/policy-crawler && go build ./cmd/main.go

build-llm-gateway:
	cd services/llm-gateway && go build ./cmd/main.go

# Test all services
test: test-shared test-api test-actuary test-crawler test-llm-gateway

test-shared:
	cd shared && go test ./... -count=1

test-api:
	cd services/api-server && go test ./... -count=1 -v

test-actuary:
	cd services/actuarial-engine && go test ./... -count=1 -v

test-crawler:
	cd services/policy-crawler && go test ./... -count=1 -v

test-llm-gateway:
	cd services/llm-gateway && go test ./... -count=1 -v

# Lint
lint:
	cd services/api-server && golangci-lint run ./... 2>/dev/null || true
	cd services/actuarial-engine && golangci-lint run ./... 2>/dev/null || true
	cd services/policy-crawler && golangci-lint run ./... 2>/dev/null || true
	cd services/llm-gateway && golangci-lint run ./... 2>/dev/null || true

# Docker compose
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-ps:
	docker compose ps

# Database migrations (local direct, requires PG running)
migrate:
	cd services/api-server && goose -dir migrations postgres "$${DATABASE_URL}" up
	cd services/policy-crawler && goose -dir migrations postgres "$${DATABASE_URL}" up

migrate-down:
	cd services/api-server && goose -dir migrations postgres "$${DATABASE_URL}" down
	cd services/policy-crawler && goose -dir migrations postgres "$${DATABASE_URL}" down

# Clean
clean:
	rm -f services/api-server/api-server
	rm -f services/actuarial-engine/actuarial-engine
	rm -f services/policy-crawler/policy-crawler
	rm -f services/llm-gateway/llm-gateway
