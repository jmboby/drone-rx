.PHONY: build build-api build-frontend lint lint-go lint-frontend lint-helm test test-go test-frontend clean

REGISTRY ?= ghcr.io/jmboby
API_IMAGE ?= $(REGISTRY)/dronerx-api
FRONTEND_IMAGE ?= $(REGISTRY)/dronerx-frontend
TAG ?= $(shell git rev-parse --short HEAD)

build: build-api build-frontend
build-api:
	docker build -f Dockerfile.api -t $(API_IMAGE):$(TAG) -t $(API_IMAGE):latest .
build-frontend:
	docker build -f Dockerfile.frontend -t $(FRONTEND_IMAGE):$(TAG) -t $(FRONTEND_IMAGE):latest .

lint: lint-go lint-frontend lint-helm
lint-go:
	go vet ./...
lint-frontend:
	cd frontend && npx svelte-check
lint-helm:
	helm lint chart/

test: test-go test-frontend
test-go:
	go test ./... -v
test-frontend:
	cd frontend && npm test

clean:
	rm -rf frontend/build frontend/.svelte-kit
	docker rmi $(API_IMAGE):$(TAG) $(FRONTEND_IMAGE):$(TAG) 2>/dev/null || true
