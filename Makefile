.PHONY: build build-api build-frontend build-amd64 build-api-amd64 build-frontend-amd64 \
        push push-api push-frontend lint lint-go lint-frontend lint-helm \
        test test-go test-frontend clean helm-package

# Docker image settings
REGISTRY ?= ghcr.io/jmboby
API_IMAGE ?= $(REGISTRY)/dronerx-api
FRONTEND_IMAGE ?= $(REGISTRY)/dronerx-frontend
TAG ?= $(shell git rev-parse --short HEAD)

## Build (local arch)

build: build-api build-frontend

build-api:
	docker build -f Dockerfile.api -t $(API_IMAGE):$(TAG) -t $(API_IMAGE):latest .

build-frontend:
	docker build -f Dockerfile.frontend -t $(FRONTEND_IMAGE):$(TAG) -t $(FRONTEND_IMAGE):latest .

## Build (linux/amd64 for CI/CMX)

build-amd64: build-api-amd64 build-frontend-amd64

build-api-amd64:
	docker build --platform linux/amd64 -f Dockerfile.api -t $(API_IMAGE):$(TAG) .

build-frontend-amd64:
	docker build --platform linux/amd64 -f Dockerfile.frontend -t $(FRONTEND_IMAGE):$(TAG) .

## Push

push: push-api push-frontend

push-api:
	docker push $(API_IMAGE):$(TAG)

push-frontend:
	docker push $(FRONTEND_IMAGE):$(TAG)

## Lint

lint: lint-go lint-frontend lint-helm

lint-go:
	go vet ./...

lint-frontend:
	cd frontend && npx svelte-check

lint-helm:
	helm lint chart/

## Test

test: test-go test-frontend

test-go:
	go test ./... -v

test-frontend:
	cd frontend && npm test

## Helm

helm-package:
	helm dependency build chart/
	helm package chart/ -d .

## Clean

clean:
	rm -rf frontend/build frontend/.svelte-kit drone-rx-*.tgz
	docker rmi $(API_IMAGE):$(TAG) $(FRONTEND_IMAGE):$(TAG) 2>/dev/null || true
