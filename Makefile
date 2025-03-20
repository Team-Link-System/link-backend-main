.PHONY: build test clean docker docker-dev push push-dev deploy deploy-dev all

APP_NAME=link-backend
DOCKER_REGISTRY=harbor.jongjong2.site:30443
DOCKER_IMAGE_NAME=link-backend
VERSION=$(shell git describe --tags --always --dirty || echo "1.0.0")
BUILD_DIR=./build

all: clean test build

build:
	@echo "Building Go application..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) .

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Running linter..."
	@golangci-lint run

docker:
	@echo "Building Docker image for production..."
	@docker build -f Dockerfile -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest .

docker-dev:
	@echo "Building Docker image for development..."
	@docker build -f Dockerfile.dev -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev .

push: docker
	@echo "Pushing Docker image for production..."
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest

push-dev: docker-dev
	@echo "Pushing Docker image for development..."
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev

deploy: build docker push
	@echo "Deploying to production..."
	@./deploy.sh production

deploy-dev: build docker-dev push-dev
	@echo "Deploying to development..."
	@./deploy.sh development

dev:
	@echo "Starting development server with air..."
	@air -c .air.toml 