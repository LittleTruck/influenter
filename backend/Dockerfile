# Influenter Backend - 生產環境 Dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder

# 安裝編譯工具
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 編譯 API server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 編譯 Worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker

# 編譯 Migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

# Stage 2: Runtime
FROM alpine:latest

# 安裝 CA 憑證 (用於 HTTPS 連線)
RUN apk --no-cache add ca-certificates tzdata

# 設定時區為台北
ENV TZ=Asia/Taipei

WORKDIR /root/

# 從 builder 複製編譯好的執行檔
COPY --from=builder /app/server .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrate .

# 複製 migrations 資料夾
COPY --from=builder /app/migrations ./migrations

# 暴露 API 埠
EXPOSE 8080

# 預設執行 API server
CMD ["./server"]

