#!/bin/bash

# 변수 설정
APP_NAME="link-backend"
DOCKER_REGISTRY="harbor.jongjong2.site:30443"
DOCKER_IMAGE_NAME="link-backend"
ENV=${1:-"production"}  # 기본값은 production

# 태그 설정
TAG="latest"
BUILD_OPTION="--docker"
if [ "$ENV" = "development" ] || [ "$ENV" = "dev" ]; then
    TAG="dev"
    BUILD_OPTION="--docker-dev"
    ENV="development"
fi

echo "🚀 $APP_NAME 배포 시작 (환경: $ENV)"

# 빌드 및 Docker 이미지 푸시
echo "🔨 빌드 및 Docker 이미지 푸시 중..."
./build.sh $BUILD_OPTION --push
if [ $? -ne 0 ]; then
    echo "❌ 빌드 또는 푸시 실패"
    exit 1
fi

# 여기에 실제 배포 로직 추가 (예: Kubernetes에 배포)
# 예시: kubectl 명령어를 사용하여 배포
if [ -f "k8s/$ENV/deployment.yaml" ]; then
    echo "☸️ Kubernetes에 배포 중..."
    kubectl apply -f k8s/$ENV/deployment.yaml
    if [ $? -ne 0 ]; then
        echo "❌ Kubernetes 배포 실패"
        exit 1
    fi
    echo "✅ Kubernetes 배포 완료"
else
    echo "ℹ️ Kubernetes 배포 파일이 없습니다. 수동으로 배포해주세요."
fi

echo "🎉 배포 완료!" 