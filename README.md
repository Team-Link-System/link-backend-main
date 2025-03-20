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
- [주요 명령어](#-주요-명령어)
- [로컬 개발 환경 설정](#-로컬-개발-환경-설정)
- [Docker 이미지 빌드 및 푸시](#-docker-이미지-빌드-및-푸시)
- [트러블슈팅](#-트러블슈팅)

## 🚀 소개

Link 백엔드 서비스는 Go 언어로 작성된 백엔드 API 및 웹소켓 서버입니다. 이 서비스는 사용자 관리, 채팅, 알림 등의 기능을 제공합니다.

## 💻 시스템 요구사항

- Go 1.23 이상
- Docker
- Git
- Air (개발용 핫 리로드)

## 🔧 환경 설정

프로젝트 루트 디렉토리에 `.env` 파일을 생성하고 필요한 환경 변수를 설정합니다.

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

## 🛠 주요 명령어

### Makefile 명령어

| 명령어 | 설명 |
|--------|------|
| `make build` | Go 애플리케이션 빌드 |
| `make test` | 테스트 실행 |
| `make clean` | 빌드 디렉토리 정리 |
| `make docker-build` | 프로덕션용 Docker 이미지 빌드 |
| `make docker-build-dev` | 개발용 Docker 이미지 빌드 |
| `make push` | 프로덕션용 이미지 Harbor에 푸시 |
| `make push-dev` | 개발용 이미지 Harbor에 푸시 |
| `make local-dev` | 로컬 개발 서버 실행 (Air) |
| `make local-prod` | 로컬 프로덕션 서버 실행 |
| `make build-push` | 빌드, Docker 이미지 생성, Harbor 푸시 (프로덕션) |
| `make build-push-dev` | 빌드, Docker 이미지 생성, Harbor 푸시 (개발) |

### build.sh 스크립트 옵션

| 옵션 | 설명 |
|------|------|
| `--skip-tests` | 테스트 실행 단계 건너뛰기 |
| `--linux-only` | Linux 플랫폼만 빌드 |
| `--darwin-only` | macOS 플랫폼만 빌드 |
| `--windows-only` | Windows 플랫폼만 빌드 |
| `--docker` | 프로덕션용 Docker 이미지 빌드 |
| `--docker-dev` | 개발용 Docker 이미지 빌드 |
| `--push` | Docker 이미지를 Harbor에 푸시 |

## 📦 로컬 개발 환경 설정

### 1. 저장소 복제하기

```bash
git clone https://github.com/your-username/link-backend.git
cd link-backend
```

### 2. 의존성 설치

```bash
go mod download
```

### 3. 로컬 개발 서버 실행 (Air)

Air를 사용하면 코드 변경 시 자동으로 서버가 재시작됩니다.

```bash
# Air 설치 (처음 한 번만)
go install github.com/air-verse/air@latest

# Air로 개발 서버 실행
make local-dev
```

### 4. 테스트 실행

```bash
# 모든 테스트 실행
make test
```

## 🐳 Docker 이미지 빌드 및 푸시

### 프로덕션 환경용

```bash
# 한 번에 빌드 및 푸시
make build-push

# 또는 단계별로 실행
make build
make docker-build
make push
```

### 개발 환경용

```bash
# 한 번에 빌드 및 푸시
make build-push-dev

# 또는 단계별로 실행
make build
make docker-build-dev
make push-dev
```

### build.sh 스크립트 사용

더 많은 옵션이 필요한 경우 build.sh 스크립트를 직접 사용할 수 있습니다.

```bash
# 테스트 건너뛰고 Linux 플랫폼만 빌드
./build.sh --skip-tests --linux-only

# 테스트 건너뛰고 프로덕션 Docker 이미지 빌드 및 푸시
./build.sh --skip-tests --docker --push

# 개발용 Docker 이미지 빌드 및 푸시
./build.sh --docker-dev --push
```

## 📄 프로젝트 구조

```
/
├── cmd/                # 메인 애플리케이션 코드
│   └── main.go         # 애플리케이션 진입점
├── internal/           # 내부 패키지
├── pkg/                # 외부에서 사용 가능한 패키지
├── build/              # 빌드 산출물
├── .air.toml           # Air 설정
├── Dockerfile          # 프로덕션용 Dockerfile
├── Dockerfile.dev      # 개발용 Dockerfile
├── build.sh            # 빌드 스크립트
├── Makefile            # 빌드 자동화
└── go.mod              # Go 모듈 정의
```

## 🔄 CI/CD 파이프라인

로컬에서 빌드하고 Harbor에 푸시한 이미지는 Kubernetes를 통해 배포될 수 있습니다:

1. `make build-push` 또는 `make build-push-dev`로 이미지 빌드 및 푸시
2. Kubernetes에서 해당 이미지를 사용하여 배포

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

### 도커 빌드 문제

도커 빌드에 문제가 있다면 다음을 확인하세요:
- `Dockerfile`과 `Dockerfile.dev`가 올바르게 설정되었는지 확인
- 빌드 전에 `make build`로 바이너리가 생성되었는지 확인
- Docker 데몬이 실행 중인지 확인

### Harbor 푸시 문제

Harbor 레지스트리에 푸시할 때 문제가 발생하면 다음을 확인하세요:
- Docker가 Harbor 레지스트리에 로그인되어 있는지 확인 (`docker login harbor.jongjong2.site:30443`)
- 적절한 네임스페이스와 태그를 사용하고 있는지 확인
- Harbor 레지스트리 연결 상태 확인

---

<div align="center">
  <p> Link 팀에서 제작하였습니다 </p>
</div>
