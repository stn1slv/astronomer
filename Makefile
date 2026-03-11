# Makefile for Astronomer project

BINARY_NAME=astronomer
DOCKER_IMAGE=stn1slv/astronomer

.PHONY: all
all: build test lint

.PHONY: setup
setup:
	@echo "Setting up development environment..."
	go mod download
	go mod tidy

.PHONY: test
test:
	@echo "Running tests..."
	go test -race -v ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Skipping..."; \
	fi

.PHONY: format
format:
	@echo "Formatting code..."
	@if command -v gofumpt > /dev/null; then \
		gofumpt -extra -w .; \
	else \
		go fmt ./...; \
	fi

.PHONY: build
build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) main.go

.PHONY: docker
docker:
	@echo "Building docker image..."
	docker build -t $(DOCKER_IMAGE) .

.PHONY: run
run: build
	@echo "Running locally..."
	./$(BINARY_NAME) $(REPO)

.PHONY: upgrade-deps
upgrade-deps:
	@echo "Upgrading dependencies..."
	go get -u ./...
	go mod tidy
