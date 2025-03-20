# 빌드 스테이지
FROM golang:1.23-alpine AS builder

# 필요한 빌드 도구 설치
RUN apk add --no-cache git

# 작업 디렉토리 설정
WORKDIR /build

# 의존성 설치를 위한 go.mod와 go.sum 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드 (정적 바이너리)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o link-backend .

# 실행 스테이지 - 최소한의 이미지 사용
FROM alpine:latest

# 필요한 런타임 패키지만 설치
RUN apk --no-cache add ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /app

# 빌드 스테이지에서 바이너리만 복사
COPY --from=builder /build/link-backend .

# 필요한 포트 노출 (HTTP 및 WS)
EXPOSE 8080 1884

# 비루트 사용자로 실행 (보안 강화)
RUN adduser -D -u 1000 appuser
USER appuser

# 빌드된 바이너리 직접 실행
CMD ["./link-backend"]
