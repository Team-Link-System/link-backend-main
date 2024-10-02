# Go가 포함된 알파인 이미지 사용
FROM golang:1.23-alpine

# 필요한 도구 설치
RUN apk add --no-cache git bash curl

# 작업 디렉토리 설정
WORKDIR /app

# Air 설치 (코드 변경 감지 도구)
RUN go install github.com/air-verse/air@latest

# 의존성 설치를 위한 go.mod와 go.sum 복사
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 포트 노출
EXPOSE 8080

# Air로 애플리케이션 실행
CMD ["air", "-c", "/app/.air.toml"]
