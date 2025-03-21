# 빌드 스테이지 - 더 최신 버전의 Go 사용
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 소스 코드 복사
COPY . .

# 정적으로 빌드
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o link-backend ./cmd/main.go

# 실행 스테이지
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 빌드 스테이지에서 바이너리만 복사
COPY --from=builder /app/link-backend .

EXPOSE 8080 1884

# 비루트 사용자로 실행
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

RUN chmod +x ./link-backend

CMD ["./link-backend"]