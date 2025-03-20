#!/bin/bash

# 변수 설정
APP_NAME="link-backend"
VERSION=$(git describe --tags --always --dirty || echo "1.0.0")
BUILD_DIR="./build"
PLATFORMS=("linux:amd64" "darwin:amd64" "windows:amd64")
DOCKER_REGISTRY="harbor.jongjong2.site:30443"
DOCKER_IMAGE_NAME="link-backend"

# 빌드 디렉토리 생성
mkdir -p $BUILD_DIR

echo "🔨 Go 애플리케이션 빌드 시작: $APP_NAME v$VERSION"

echo "🧪 테스트 실행 중..."
go test ./...
if [ $? -ne 0 ]; then
    echo "❌ 테스트 실패"
    exit 1
fi
echo "✅ 테스트 성공"

# 각 플랫폼별 빌드
echo "🏗️ 크로스 플랫폼 빌드 시작..."
for platform in "${PLATFORMS[@]}"; do
    IFS=':' read -r -a parts <<< "$platform"
    OS="${parts[0]}"
    ARCH="${parts[1]}"
    
    output="${BUILD_DIR}/${APP_NAME}_${OS}_${ARCH}"
    if [ "$OS" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo "Building for $OS/$ARCH..."
    GOOS=$OS GOARCH=$ARCH go build -ldflags="-X main.Version=$VERSION" -o $output .
    if [ $? -ne 0 ]; then
        echo "❌ $OS/$ARCH 빌드 실패"
    else
        echo "✅ $OS/$ARCH 빌드 성공: $output"
    fi
done

echo "🎉 빌드 완료!"

# Docker 이미지 빌드 (환경에 따라 다른 Dockerfile 사용)
if [ "$1" = "--docker" ] || [ "$1" = "--docker-dev" ]; then
    ENV="production"
    TAG="latest"
    DOCKERFILE="Dockerfile"
    
    if [ "$1" = "--docker-dev" ]; then
        ENV="development"
        TAG="dev"
        DOCKERFILE="Dockerfile.dev"
    fi
    
    echo "🐳 Docker 이미지 빌드 중... (환경: $ENV)"
    docker build -f $DOCKERFILE -t "$DOCKER_REGISTRY/$DOCKER_IMAGE_NAME:$TAG" .
    if [ $? -ne 0 ]; then
        echo "❌ Docker 이미지 빌드 실패"
        exit 1
    fi
    echo "✅ Docker 이미지 빌드 완료"
    
    # 도커 레지스트리 푸시 (선택적)
    if [ "$2" = "--push" ]; then
        echo "🚀 Docker 이미지 푸시 중..."
        docker push "$DOCKER_REGISTRY/$DOCKER_IMAGE_NAME:$TAG"
        if [ $? -ne 0 ]; then
            echo "❌ Docker 이미지 푸시 실패"
            exit 1
        fi
        echo "✅ Docker 이미지 푸시 완료"
    fi
fi 