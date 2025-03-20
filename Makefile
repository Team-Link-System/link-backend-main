.PHONY: build test clean docker-build docker-build-dev push push-dev local-dev

APP_NAME=link-backend
DOCKER_REGISTRY=harbor.jongjong2.site:30443/link-backend
DOCKER_IMAGE_NAME=link-backend
VERSION=$(shell git describe --tags --always --dirty || echo "1.0.0")
BUILD_DIR=./build

# 기본 명령어: 빌드
build:
	@echo "Building Go application..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

# 테스트 실행
test:
	@echo "Running tests..."
	@go test ./...

# 빌드 디렉토리 정리
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# 프로덕션용 도커 이미지 빌드
docker-build:
	@echo "Building Docker image for production..."
	@docker build -f Dockerfile -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest .

# 개발용 도커 이미지 빌드
docker-build-dev:
	@echo "Building Docker image for development..."
	@docker build -f Dockerfile.dev -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev .

# 프로덕션용 이미지 Harbor 푸시
push: docker-build
	@echo "Pushing Docker image for production..."
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):latest

# 개발용 이미지 Harbor 푸시
push-dev: docker-build-dev
	@echo "Pushing Docker image for development..."
	@docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):dev

# 로컬 개발 서버 실행 (Air)
local-dev:
	@echo "Starting development server with air..."
	@air -c .air.toml

local-prod:
	@echo "Starting production server..."
	@chmod +x ./build/link-backend
	@./build/link-backend

# 빠른 빌드-푸시 (프로덕션)
build-push: build docker-build push
	@echo "Build and push completed for production"

# 빠른 빌드-푸시 (개발)
build-push-dev: build docker-build-dev push-dev
	@echo "Build and push completed for development"