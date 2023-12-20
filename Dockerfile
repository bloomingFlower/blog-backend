# Go 애플리케이션 빌드를 위한 베이스 이미지
FROM golang:1.21.3 AS builder

# 작업 디렉토리 설정
WORKDIR /app

# 의존성 파일 복사 및 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# Go 애플리케이션 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 최종 실행 이미지
FROM alpine:3.19.0
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY .env /root/.env

# 실행 명령
CMD ["./main"]