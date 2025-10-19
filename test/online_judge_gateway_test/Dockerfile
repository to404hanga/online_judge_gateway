# 使用官方 Go 镜像作为构建环境
FROM golang:1.23.4-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum 文件（如果存在）
# 由于这是一个简单的测试程序，我们先复制 main.go
COPY main.go .

# 初始化 Go 模块（如果不存在 go.mod）
RUN go mod init online_judge_gateway_test || true

# 下载依赖
RUN go mod tidy

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test_server main.go

# 使用轻量级的 alpine 镜像作为运行环境
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/test_server .

# 创建日志目录
RUN mkdir -p log

# 更改文件所有者
RUN chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口（默认端口，可以通过环境变量覆盖）
EXPOSE ${TEST_PORT:-8080}

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${TEST_PORT:-8080}/health || exit 1

# 运行应用程序，使用环境变量传递端口
CMD ["sh", "-c", "./test_server --port=${TEST_PORT:-8080}"]