# Build stage
FROM golang:1.21-alpine AS builder

# 필요한 패키지 설치
RUN apk add --no-cache git gcc musl-dev

# 작업 디렉토리 설정
WORKDIR /app

# 의존성 파일 복사 및 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Runtime stage
FROM alpine:latest

# 필요한 패키지 설치 (Oracle Instant Client 등)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 빌드된 바이너리 복사
COPY --from=builder /app/main .

# 설정 파일 복사 (필요한 경우)
# COPY --from=builder /app/.env .

# 포트 노출
EXPOSE 8080

# 애플리케이션 실행
CMD ["./main"]
