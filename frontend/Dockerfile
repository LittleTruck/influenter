# Influenter Frontend - 生產環境 Dockerfile
# Stage 1: Build
# 使用 Node.js 20+ 以支援 Nuxt 3 的最新版本
FROM node:20 AS builder

WORKDIR /app

# 複製 package files
COPY package*.json ./

# 安裝依賴
RUN npm ci

# 複製原始碼
COPY . .

# 建置 Nuxt 應用
RUN npm run build

# Stage 2: Runtime
FROM node:20-slim

WORKDIR /app

# 複製建置產物
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package*.json ./

# 只安裝生產依賴
RUN npm ci --only=production

# 暴露埠
EXPOSE 3000

# 設定環境變數
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

# 啟動應用
CMD ["node", ".output/server/index.mjs"]

