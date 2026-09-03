.PHONY: dev dev-backend dev-frontend dev-ai dev-infra test lint build migrate-up migrate-down seed generate-types

dev-infra:
	docker compose -f infrastructure/local-testbed/docker-compose.yml up -d

dev-backend:
	cd apps/backend && go run cmd/server/main.go

dev-frontend:
	cd apps/frontend && npm run dev

dev-ai:
	cd apps/ai-service && uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload

test:
	@echo "Running tests..."
	cd apps/backend && go test ./...
	cd apps/frontend && npm run test -- --run
	cd apps/ai-service && pytest

lint:
	@echo "Running linters..."
	cd apps/backend && golangci-lint run ./...
	cd apps/frontend && npm run lint
	cd apps/ai-service && ruff check .

build:
	@echo "Building services..."
	docker build -t cifo/backend:latest apps/backend
	docker build -t cifo/frontend:latest apps/frontend
	docker build -t cifo/ai-service:latest apps/ai-service

migrate-up:
	@echo "Running migrations..."
	go run apps/backend/cmd/migrator/main.go up

migrate-down:
	@echo "Reverting migrations..."
	go run apps/backend/cmd/migrator/main.go down

seed:
	@echo "Seeding initial config..."
	psql -h localhost -p 5432 -U cifo_admin -d cifo_db -f scripts/seed-data.sql

generate-types:
	@echo "Generating API contracts..."
	./scripts/generate-api-types.sh
