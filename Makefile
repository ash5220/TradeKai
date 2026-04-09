.PHONY: build test lint run docker-up docker-down migrate-up migrate-down sqlc-generate swagger integration-test

# Go
BINARY_NAME=tradekai
GO_CMD=go
BUILD_DIR=./bin

build:
	cd backend && $(GO_CMD) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

test:
	cd backend && $(GO_CMD) test -race -coverprofile=coverage.out ./internal/...
	cd backend && $(GO_CMD) tool cover -func=coverage.out

integration-test:
	cd backend && $(GO_CMD) test -race -tags=integration ./...

lint:
	cd backend && golangci-lint run ./...
	cd frontend && ng lint

run:
	cd backend && $(GO_CMD) run ./cmd/server

# Database
migrate-up:
	cd backend && migrate -path internal/store/migrations -database "$$DATABASE_URL" up

migrate-down:
	cd backend && migrate -path internal/store/migrations -database "$$DATABASE_URL" down

migrate-create:
	cd backend && migrate create -ext sql -dir internal/store/migrations -seq $(name)

# Code generation
sqlc: sqlc-generate

sqlc-generate:
	cd backend && sqlc generate

swagger:
	cd backend && swag init -g cmd/server/main.go -o api

# Docker
docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker-compose.yml down

docker-up-dev:
	docker-compose -f deployments/docker-compose.yml -f deployments/docker-compose.dev.yml up -d

# Angular
ng-install:
	cd frontend && npm install

ng-build:
	cd frontend && ng build --configuration=production

ng-dev:
	cd frontend && ng serve

# Setup
setup:
	chmod +x scripts/setup.sh && ./scripts/setup.sh
