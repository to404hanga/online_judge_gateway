# 在线判题系统网关 (Online Judge Gateway)

这是一个基于 Go 语言开发的在线判题系统网关服务，使用 Gin 框架构建，提供用户认证、请求代理等功能。

## 项目架构

- **框架**: Gin + GORM + Redis
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **依赖注入**: Google Wire
- **容器化**: Docker + Docker Compose

## 功能特性

- 用户认证与授权 (JWT)
- 请求代理转发
- CORS 跨域支持
- 日志记录
- 健康检查
- 容器化部署

## 快速开始

### 环境要求

- Go 1.23.4+
- Docker & Docker Compose
- MySQL 8.0 (如果本地运行)
- Redis 7+ (如果本地运行)

### 方式一：Docker Compose 部署 (推荐)

1. **克隆项目**

   ```bash
   git clone <repository-url>
   cd online_judge_gateway
   ```

2. **配置文件准备**

   ```bash
   # Windows
   copy config\config.template.yaml config\config.yaml
   
   # Linux/Mac
   cp config/config.template.yaml config/config.yaml
   ```

3. **修改配置文件**
   编辑 `config/config.yaml`，根据需要修改以下配置：

   ```yaml
   db:
     host: "mysql"  # Docker 环境使用服务名
     port: 3306
     username: "oj"
     password: "123456"
     dbname: "online_judge"
   
   redis:
     host: "redis"  # Docker 环境使用服务名
     port: 6379
     password: ""
     db: 0
   
   jwt:
     jwt_key: "a7f3e9d2c8b4f1a6e5d8c3b7f2a9e6d1c4b8f5a2e7d3c9b6f1a4e8d2c5b9f3a6"
     refresh_key: "e1d4c7b2f8a5e9d3c6b1f4a7e2d5c8b3f6a9e4d7c1b5f2a8e6d9c3b7f1a4e8d2"
   ```

4. **启动服务**

   ```bash
   docker-compose up -d
   ```

5. **查看服务状态**

   ```bash
   docker-compose ps
   docker-compose logs gateway
   ```

6. **访问服务**
   - 网关服务: <http://localhost:8080>
   - 健康检查: <http://localhost:8080/health>

### 方式二：本地开发运行

1. **安装依赖**

   ```bash
   go mod download
   ```

2. **启动数据库服务**

   ```bash
   # 启动 MySQL 和 Redis (可使用 Docker)
   docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=rootpassword -e MYSQL_DATABASE=online_judge -e MYSQL_USER=oj -e MYSQL_PASSWORD=123456 mysql:8.0
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

3. **配置文件准备**

   ```bash
   # Windows
   copy config\config.template.yaml config\config.yaml
   
   # Linux/Mac
   cp config/config.template.yaml config/config.yaml
   ```

   修改 `config/config.yaml` 中的数据库连接信息：

   ```yaml
   db:
     host: "localhost"
     port: 3306
     username: "oj"
     password: "123456"
     dbname: "online_judge"
   
   redis:
     host: "localhost"
     port: 6379
   ```

4. **生成依赖注入代码**

   ```bash
   go generate ./...
   ```

5. **运行服务**

   ```bash
   go run main.go --config ./config/config.yaml
   ```

## 配置说明

### 数据库配置 (db)

- `host`: 数据库主机地址
- `port`: 数据库端口
- `username`: 数据库用户名
- `password`: 数据库密码
- `dbname`: 数据库名称
- `table_prefix`: 表前缀

### 日志配置 (log)

- `development`: 是否为开发模式
- `type`: 日志输出类型 (0-控制台, 1-文件, 2-控制台+文件)
- `log_file_path`: 日志文件路径
- `auto_create_file`: 是否自动创建日志文件

### Gin 配置 (gin)

- `addr`: 服务监听地址
- `allow_origins`: 允许的跨域来源
- `allow_methods`: 允许的HTTP方法
- `allow_headers`: 允许的请求头
- `expose_headers`: 暴露的响应头
- `AllowCredentials`: 是否允许携带凭证
- `max_age`: 预检请求缓存时间
- `pass_path_method_pairs`: 无需认证的路径

### Redis 配置 (redis)

- `host`: Redis 主机地址
- `port`: Redis 端口
- `password`: Redis 密码
- `db`: Redis 数据库编号

### JWT 配置 (jwt)

- `jwt_expiration`: JWT 令牌过期时间 (分钟)
- `refresh_expiration`: 刷新令牌过期时间 (分钟)
- `jwt_key`: JWT 签名密钥 (64位随机字符串)
- `refresh_key`: 刷新令牌签名密钥 (64位随机字符串)

## API 接口

### 认证相关

- `POST /auth/login` - 用户登录
- `POST /auth/refresh` - 刷新令牌
- `POST /auth/logout` - 用户登出

### 健康检查

- `GET /health` - 服务健康检查

## 开发指南

### 项目结构

```
├── config/          # 配置文件
│   ├── config.go    # 配置加载逻辑
│   ├── config.template.yaml  # 配置模板
│   └── types.go     # 配置类型定义
├── constant/        # 常量定义
├── domain/          # 领域模型
├── ioc/            # 依赖注入配置
│   ├── db.go       # 数据库配置
│   ├── gin.go      # Gin 服务器配置
│   ├── jwt.go      # JWT 配置
│   ├── logger.go   # 日志配置
│   ├── proxy.go    # 代理配置
│   └── redis.go    # Redis 配置
├── service/        # 业务逻辑层
├── web/            # Web 层
│   ├── middleware/ # 中间件
│   │   ├── cors.go # CORS 中间件
│   │   ├── jwt.go  # JWT 中间件
│   │   └── logger.go # 日志中间件
│   ├── jwt/        # JWT 相关
│   ├── auth.go     # 认证处理器
│   ├── proxy.go    # 代理处理器
│   ├── server.go   # 服务器配置
│   └── types.go    # Web 类型定义
├── main.go         # 程序入口
├── wire.go         # Wire 配置
├── wire_gen.go     # Wire 生成代码
└── docker-compose.yml
```

### 添加新的 API 端点

1. 在 `web/` 目录下创建对应的处理器
2. 在 `service/` 目录下实现业务逻辑
3. 在 `ioc/gin.go` 中注册路由
4. 更新 Wire 配置

### 代码生成

```bash
# 重新生成 Wire 依赖注入代码
go generate ./...
```

## 部署说明

### 生产环境部署

1. **修改生产配置**
   - 更新数据库连接信息
   - 设置强密码和密钥
   - 配置日志级别
   - 设置 `GIN_MODE=release`

2. **构建镜像**

   ```bash
   docker build -t online-judge-gateway:latest .
   ```

3. **使用 Docker Compose 部署**

   ```bash
   docker-compose up -d
   ```

### 环境变量

可以通过环境变量覆盖配置：

- `GIN_MODE`: Gin 运行模式 (debug/release)

### 监控和日志

- 日志文件位置: `./log/gateway.log`
- 健康检查端点: `/health`
- 容器日志查看: `docker-compose logs -f gateway`

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接配置是否正确
   - 确认网络连通性

2. **Redis 连接失败**
   - 检查 Redis 服务状态
   - 验证 Redis 配置
   - 检查防火墙设置

3. **JWT 令牌问题**
   - 确认 JWT 密钥配置正确
   - 检查令牌过期时间设置
   - 验证令牌格式

4. **端口冲突**
   - 检查端口 8080 是否被占用
   - 修改 docker-compose.yml 中的端口映射

5. **Wire 依赖注入问题**
   - 运行 `go generate ./...` 重新生成代码
   - 检查 wire.go 配置是否正确

### 日志查看

```bash
# 查看应用日志
docker-compose logs gateway

# 查看实时日志
docker-compose logs -f gateway

# 查看特定时间的日志
docker-compose logs --since="2024-01-01T00:00:00" gateway
```

### 调试模式

在开发环境中，可以设置以下配置启用调试模式：

```yaml
log:
  development: true
  type: 2  # 控制台+文件输出

gin:
  # Gin 会自动检测 GIN_MODE 环境变量
```

## 性能优化

1. **数据库连接池**
   - 根据并发需求调整连接池大小
   - 设置合适的连接超时时间

2. **Redis 缓存**
   - 合理设置缓存过期时间
   - 使用 Redis 集群提高可用性

3. **JWT 优化**
   - 设置合理的令牌过期时间
   - 使用 Redis 存储刷新令牌

## 安全建议

1. **密钥管理**
   - 使用强随机密钥
   - 定期轮换密钥
   - 不要在代码中硬编码密钥

2. **网络安全**
   - 使用 HTTPS
   - 配置防火墙规则
   - 限制数据库访问

3. **认证授权**
   - 实施最小权限原则
   - 定期审查用户权限
   - 记录安全相关日志

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
