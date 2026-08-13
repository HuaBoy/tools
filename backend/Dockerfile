# 多阶段构建
# ========================================
# 阶段 1: 构建
# ========================================
FROM golang:1.21-alpine AS builder

# 设置 Go 代理（中国大陆加速）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /app

# 复制源代码
COPY . .

# 下载依赖并生成 go.sum
RUN go mod tidy

# 编译
RUN go build -ldflags="-s -w" -o /app/bin/tester-platform ./cmd/server

# ========================================
# 阶段 2: 运行
# ========================================
FROM alpine:latest

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 复制编译产物
COPY --from=builder /app/bin/tester-platform /app/tester-platform

# 创建日志目录和知识库存储目录
RUN mkdir -p /app/logs /app/data/knowledge && chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 启动
ENTRYPOINT ["/app/tester-platform"]
