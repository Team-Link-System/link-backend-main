# Link 백엔드 서비스 실행 가이드

![Link Backend](https://img.shields.io/badge/Link-Backend-blue)
![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)

<div align="center">
  <img src="https://go.dev/images/gophers/ladder.svg" width="200" alt="Gopher">
</div>

## 📋 목차

- [소개](#-소개)
- [시스템 요구사항](#-시스템-요구사항)
- [환경 설정](#-환경-설정)
- [로컬 개발 환경 설정](#-로컬-개발-환경-설정)
- [Docker를 사용한 실행](#-docker를-사용한-실행)
- [배포](#-배포)
- [트러블슈팅](#-트러블슈팅)

## 🚀 소개

Link 백엔드 서비스는 Go 언어로 작성된 백엔드 API 및 웹소켓 서버입니다. 이 서비스는 사용자 관리, 채팅, 알림 등의 기능을 제공합니다.

## 💻 시스템 요구사항

- Go 1.23 이상
- Docker 및 Docker Compose (선택 사항)
- Git
- PostgreSQL
- Redis
- MongoDB
- NATS 메시징 서버

## 🔧 환경 설정

프로젝트는 두 가지 환경 설정 방식을 지원합니다:
1. **로컬 개발 환경**: 로컬 머신에 필요한 서비스를 직접 설치하여 실행
2. **컨테이너 환경**: Docker를 사용하여 모든 서비스를 컨테이너로 실행

### 환경 변수 설정

프로젝트 루트 디렉토리에 `.env` 파일을 생성하고 필요한 환경 변수를 설정합니다. 환경에 따라 적절하게 주석을 해제하여 사용하세요.

#### 로컬 개발 환경용 `.env` 필수 변수

```
# 프론트엔드 도메인
LINK_UI_URL=

# PostgreSQL 설정
POSTGRES_DSN=

# Redis 설정
REDIS_ADDR=
REDIS_PASSWORD=
REDIS_DB=

# MongoDB 설정
MONGO_DSN=

# Go 서버 설정
GO_ENV=
HTTP_PORT=
WS_PORT=
WS_PATH=
ACCESS_TOKEN_SECRET=
REFRESH_TOKEN_SECRET=

# 시스템 관리자 계정
SYSTEM_ADMIN_EMAIL=
SYSTEM_ADMIN_PASSWORD=

# NATS 설정
NATS_URL=
NATS_WS_URL=
NATS_JETSTREAM_URL=
```

#### 컨테이너 환경용 `.env` 필수 변수

```
# 프론트엔드 도메인
LINK_UI_URL=

# PostgreSQL 설정
POSTGRES_DSN=
POSTGRES_PORT=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=

# Redis 설정
REDIS_ADDR=
REDIS_PORT=
REDIS_PASSWORD=
REDIS_DB=

# MongoDB 설정
MONGO_DSN=

# Go 서버 설정
GO_ENV=
HTTP_PORT=
WS_PORT=
WS_PATH=
ACCESS_TOKEN_SECRET=
REFRESH_TOKEN_SECRET=

# 시스템 관리자 계정
SYSTEM_ADMIN_EMAIL=
SYSTEM_ADMIN_PASSWORD=

# NATS 설정
NATS_URL=
NATS_WS_URL=
NATS_JETSTREAM_URL=
```

## 📦 로컬 개발 환경 설정

### 1. 저장소 복제하기

```bash
git clone https://github.com/your-username/link-backend.git
cd link-backend
```

### 2. 필요한 서비스 설치

#### PostgreSQL 설치 및 실행

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib

# macOS (Homebrew)
brew install postgresql
brew services start postgresql
```

#### Redis 설치 및 실행

```bash
# Ubuntu/Debian
sudo apt-get install redis-server
sudo systemctl start redis-server

# macOS (Homebrew)
brew install redis
brew services start redis
```

#### MongoDB 설치 및 실행

```bash
# Ubuntu/Debian
sudo apt-get install mongodb
sudo systemctl start mongodb

# macOS (Homebrew)
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
```

#### NATS 설치 및 실행

```bash
# Go로 설치
go install github.com/nats-io/nats-server/v2@latest

# 실행
nats-server
```

### 3. 의존성 설치

```bash
go mod download
```

### 4. 데이터베이스 초기화 (필요한 경우)

```bash
# PostgreSQL 데이터베이스 생성
psql -U postgres -c "CREATE DATABASE test_db;"

# 마이그레이션 실행 (필요한 경우)
go run cmd/migration/main.go
```

### 5. 개발 서버 실행

#### Air를 사용한 실행 (핫 리로드)

```bash
# Air 설치 (처음 한 번만)
go install github.com/air-verse/air@latest

# Air로 실행
air
```

#### 일반 실행

```bash
go run main.go
```

## 🐳 Docker를 사용한 실행

### 1. Docker 이미지 빌드 및 푸시

#### 개발 환경용

```bash
# 빌드 및 푸시
./build.sh --docker-dev --push

# 또는 Makefile 사용
make docker-dev push-dev
```

#### 프로덕션 환경용

```bash
# 빌드 및 푸시
./build.sh --docker --push

# 또는 Makefile 사용
make docker push
```

### 2. Docker Compose로 전체 스택 실행

```bash
# 환경 변수 파일 복사 (.env.dev를 .env로)
cp .env.dev .env

# Docker Compose 실행
docker-compose up -d
```

### 3. 실행 확인

```bash
# 컨테이너 상태 확인
docker-compose ps

# 로그 확인
docker-compose logs -f link-backend
```

## 🚢 배포

### 개발 환경 배포

```bash
# 개발 환경에 배포
./deploy.sh development

# 또는 Makefile 사용
make deploy-dev
```

### 프로덕션 환경 배포

```bash
# 프로덕션 환경에 배포
./deploy.sh production

# 또는 Makefile 사용
make deploy
```

## 🔄 CI/CD 파이프라인

이 프로젝트는 GitOps 방식의 CI/CD 파이프라인을 사용합니다:

1. 코드 변경 사항을 Git 저장소에 푸시합니다.
2. CI 시스템이 테스트를 실행하고 Docker 이미지를 빌드합니다.
3. CD 시스템이 새 버전을 Kubernetes 클러스터에 배포합니다.

## 🛠️ 트러블슈팅

### 웹소켓 연결 문제

웹소켓 연결 문제가 발생하면 다음을 확인하세요:
- CORS 설정이 올바른지 확인 (`LINK_UI_URL` 환경 변수 확인)
- 클라이언트가 올바른 URL과 포트로 연결 시도하는지 확인 (`WS_PORT` 및 `WS_PATH` 확인)
- 방화벽이 웹소켓 연결을 차단하지 않는지 확인

### 데이터베이스 연결 문제

데이터베이스 연결 문제가 발생하면 다음을 확인하세요:
- 환경 변수가 올바르게 설정되었는지 확인 (`POSTGRES_DSN`, `REDIS_ADDR`, `MONGO_DSN`)
- 데이터베이스 서버가 실행 중인지 확인
- 네트워크 연결 및 방화벽 설정 확인

### NATS 연결 문제

NATS 연결 문제가 발생하면 다음을 확인하세요:
- NATS 서버가 실행 중인지 확인
- 환경 변수가 올바르게 설정되었는지 확인 (`NATS_URL`, `NATS_WS_URL`, `NATS_JETSTREAM_URL`)
- 로그에서 연결 오류 메시지 확인

---

<div align="center">
  <p>❤️ Link 팀에서 제작하였습니다 ❤️</p>
</div>
