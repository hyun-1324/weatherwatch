.PHONY: install dev-backend dev-frontend lint test build fmt

install:
	git config core.hooksPath .githooks
	cd frontend && npm ci

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

lint:
	cd backend && golangci-lint run ./... && go mod tidy -diff
	cd frontend && npm run format:check && npm run lint && npm run typecheck

test:
	cd backend && go test -race ./...
	cd frontend && npm test

build:
	cd frontend && npm run build

fmt:
	cd backend && golangci-lint fmt ./...
	cd frontend && npm run format
