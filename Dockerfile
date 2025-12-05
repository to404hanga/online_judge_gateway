# 使用官方 Go 镜像作为构建环境
FROM golang:1.23.4-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 设置Go代理（使用国内镜像源）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 生成 wire 依赖注入代码
RUN go generate ./...

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 使用轻量级的 alpine 镜像作为运行环境
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 创建配置文件目录
RUN mkdir -p config

# 复制配置文件模板（如果需要的话）
COPY --from=builder /app/config/config.template.yaml ./config/

# 更改文件所有者
RUN chown -R appuser:appgroup /root/

# 切换到非 root 用户
USER appuser

# 暴露端口（根据你的 gin 服务器配置调整）
EXPOSE 8080

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行应用程序
CMD ["./main", "--config", "/root/config/config.yaml"]