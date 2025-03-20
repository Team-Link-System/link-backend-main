#!/bin/bash

# 변수 설정
APP_NAME="link-backend"
VERSION=$(git describe --tags --always --dirty || echo "1.0.0")
BUILD_DIR="./build"
PLATFORMS=("linux:amd64" "darwin:amd64" "windows:amd64")
DOCKER_REGISTRY="harbor.jongjong2.site:30443/link-backend"
DOCKER_IMAGE_NAME="link-backend"

# 빌드 디렉토리 생성
mkdir -p $BUILD_DIR

echo "🔨 Go 애플리케이션 빌드 시작: $APP_NAME v$VERSION"

# 테스트 실행 여부 확인 (--skip-tests 옵션으로 건너뛸 수 있음)
if [[ "$*" == *"--skip-tests"* ]]; then
    echo "🔄 테스트 건너뛰기..."
else
    echo "🧪 테스트 실행 중..."
    go test ./...
    if [ $? -ne 0 ]; then
        echo "❌ 테스트 실패"
        exit 1
    fi
    echo "✅ 테스트 성공"
fi

# 특정 플랫폼만 빌드할지 확인 (--linux-only, --darwin-only, --windows-only 옵션)
if [[ "$*" == *"--linux-only"* ]]; then
    PLATFORMS=("linux:amd64")
    echo "🔧 Linux 플랫폼만 빌드합니다."
elif [[ "$*" == *"--darwin-only"* ]]; then
    PLATFORMS=("darwin:amd64")
    echo "🔧 macOS 플랫폼만 빌드합니다."
elif [[ "$*" == *"--windows-only"* ]]; then
    PLATFORMS=("windows:amd64")
    echo "🔧 Windows 플랫폼만 빌드합니다."
fi

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
    GOOS=$OS GOARCH=$ARCH go build -ldflags="-X main.Version=$VERSION" -o $output ./cmd/main.go
    if [ $? -ne 0 ]; then
        echo "❌ $OS/$ARCH 빌드 실패"
    else
        echo "✅ $OS/$ARCH 빌드 성공: $output"
    fi
done

echo "🎉 빌드 완료!"

# Docker 이미지 빌드 (환경에 따라 다른 Dockerfile 사용)
if [[ "$*" == *"--docker"* ]] || [[ "$*" == *"--docker-dev"* ]]; then
    ENV="production"
    TAG="latest"
    DOCKERFILE="Dockerfile"
    
    if [[ "$*" == *"--docker-dev"* ]]; then
        ENV="development"
        TAG="dev"
        DOCKERFILE="Dockerfile.dev"
    fi
    
    echo "🐳 Docker 이미지 빌드 중... (환경: $ENV)"
    # 로컬 빌드된 바이너리 활용
    docker build -f $DOCKERFILE -t "$DOCKER_REGISTRY:$TAG" .
    if [ $? -ne 0 ]; then
        echo "❌ Docker 이미지 빌드 실패"
        exit 1
    fi
    echo "✅ Docker 이미지 빌드 완료"
    
    # 도커 레지스트리 푸시 (선택적)
    if [[ "$*" == *"--push"* ]]; then
        echo "🚀 Docker 이미지 푸시 중..."
        docker push "$DOCKER_REGISTRY:$TAG"
        if [ $? -ne 0 ]; then
            echo "❌ Docker 이미지 푸시 실패"
            exit 1
        fi
        echo "✅ Docker 이미지 푸시 완료"
    fi
fi