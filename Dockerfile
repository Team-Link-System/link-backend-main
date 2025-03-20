FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY build/link-backend_linux_amd64 ./link-backend

EXPOSE 8080 1884

# 비루트 사용자로 실행 (보안 강화)
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

RUN chmod +x ./link-backend

CMD ["./link-backend"]