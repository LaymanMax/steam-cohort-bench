.PHONY: all build install ingest dev test test-go test-web clean docker-build docker-up

all: build

install:
	go mod download
	cd web && pnpm install

ingest:
	go run ./cmd/ingest --input ./data/raw/steam-games.parquet

ingest-download:
	go run ./cmd/ingest --download

build-web:
	cd web && pnpm run build

build-server:
	go build -o bin/server ./cmd/server
	go build -o bin/ingest ./cmd/ingest

build: build-web build-server

dev-server:
	go run ./cmd/server

dev-web:
	cd web && pnpm run dev

dev:
	@echo "To run local development with Vite live reloading:"
	@echo "Terminal 1: make dev-server"
	@echo "Terminal 2: make dev-web"

test-go:
	go test -v ./...

test-web:
	cd web && pnpm test

test: test-go test-web

clean:
	rm -rf bin/
	rm -rf web/dist/

docker-build:
	docker build -t steam-cohort-bench:latest .

docker-up:
	docker compose up -d
