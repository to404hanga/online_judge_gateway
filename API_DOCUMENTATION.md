# 在线判题系统网关 API 文档

## 文档信息

- **版本**: v1.0.1
- **最后更新**: 2025-10-19
- **维护者**: to404hanga

## 项目概述

这是一个基于 Gin + GORM + Redis 的微服务网关，为在线判题系统提供统一的入口点，包含用户认证、服务代理和负载均衡功能。

## 技术栈

- **框架**: Gin + GORM + Redis
- **认证**: JWT (JSON Web Token)
- **负载均衡**: 支持轮询、随机、加权随机、加权轮询
- **健康检查**: 自动服务实例健康监控
- **依赖注入**: Wire

## 基础信息

- **Base URL**: `http://localhost:8080` (默认)
- **Content-Type**: `application/json`
- **认证方式**: JWT Token (通过 Cookie)
- **API 版本**: v1

## 目录

- [认证相关 API](#认证相关-api)
- [健康检查 API](#健康检查-api)
- [服务代理 API](#服务代理-api)
- [服务管理 API](#服务管理-api-管理员接口)
- [数据模型](#数据模型)
- [错误码说明](#错误码说明)
- [认证机制](#认证机制)
- [负载均衡策略](#负载均衡策略)
- [健康检查](#健康检查)
- [中间件](#中间件)
- [使用示例](#使用示例)
- [部署说明](#部署说明)
- [版本更新记录](#版本更新记录)

## 认证相关 API

### 1. 用户登录

**接口地址**: `POST /auth/login`

**描述**: 用户登录接口，验证用户名和密码，成功后返回 JWT Token

**请求头**:

```
Content-Type: application/json
```

**请求参数**:

```json
{
  "username": "string", // 必填，用户名，长度1-50字符
  "password": "string" // 必填，密码，长度6-100字符
}
```

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "login success"
}

// 失败响应 (400)
{
  "error": "用户名或密码错误"
}

// 参数错误 (400)
{
  "error": "Key: 'LoginRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag"
}
```

**说明**:

- 登录成功后，JWT Token 会自动设置到 Cookie 中 (access_token)
- Token 包含用户 ID、会话 ID 和用户代理信息
- Token 有效期可配置，默认 24 小时

### 2. 用户登出

**接口地址**: `POST /auth/logout`

**描述**: 用户登出接口，清除 JWT Token

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**: 无

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "logout success"
}

// 失败响应 (400)
{
  "error": "清除token失败"
}

// 未认证 (401)
{
  "error": "token不存在"
}
```

### 3. 获取用户信息

**接口地址**: `GET /auth/info`

**描述**: 获取当前登录用户的详细信息

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**: 无 (需要 JWT 认证)

**响应示例**:

```json
// 成功响应 (200)
{
  "username": "admin",         // 用户名
  "realname": "管理员",        // 真实姓名
  "role": 1,                   // 用户角色 (1=管理员, 0=普通用户)
  "status": 1                  // 用户状态 (1=正常, 0=禁用)
}

// 失败响应 (500)
{
  "error": "获取用户信息失败"
}

// 未认证 (401)
{
  "error": "token不存在"
}
```

## 健康检查 API

### 4. 健康检查

**接口地址**: `GET /health`

**描述**: 检查网关服务健康状态

**请求参数**: 无

**响应示例**:

```
// 成功响应 (200)
// 无响应体，仅返回状态码
```

**说明**:

- 用于负载均衡器或监控系统检查服务状态
- 返回 200 状态码表示服务正常

## 服务代理 API

### 5. 服务转发

**接口地址**: `ANY /api/*path`

**描述**: 将请求转发到后端微服务，支持负载均衡

**请求头**:

```
Cookie: access_token=your_jwt_token
Content-Type: application/json (POST/PUT请求)
```

**请求参数**:

- **路径参数**: `path` - 服务路径，用于匹配对应的后端服务
- **查询参数**: `cmd` - 必填，目标服务的具体路径
- **请求体**: 根据目标服务要求

**响应**: 直接返回后端服务的响应

**请求示例**:

```bash
# 转发到用户服务
GET /api/user-service?cmd=users/profile

# 转发到题目服务
POST /api/problem-service?cmd=problems
```

**响应示例**:

```json
// 成功响应 - 返回后端服务响应
{
  // 后端服务的实际响应内容
}

// 服务不存在 (404)
{
  "error": "service not found"
}

// 无健康实例 (503)
{
  "error": "no healthy instance found"
}

// 后端服务错误 (502)
{
  "error": "backend service error"
}

// 未认证 (401)
{
  "error": "token不存在"
}
```

**说明**:

- 需要 JWT 认证
- 自动添加用户信息到请求头 (X-User-ID, X-Request-ID, X-Forwarded-By)
- 支持多种负载均衡策略
- 自动健康检查和故障转移
- 支持所有 HTTP 方法 (GET, POST, PUT, DELETE 等)

## 服务管理 API (管理员接口, 已弃用)

### 6. 获取所有服务

**接口地址**: `GET /admin/proxy/services`

**描述**: 获取当前网关管理的所有服务配置

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**: 无

**响应示例**:

```json
{
  "user-service": {
    "service_name": "user-service",
    "instances": [
      {
        "url": "http://localhost:8081",
        "weight": 1,
        "healthy": true,
        "last_check": "2024-12-20T10:30:00Z"
      }
    ],
    "health_check": "/health",
    "load_balancer": "round_robin"
  },
  "problem-service": {
    "service_name": "problem-service",
    "instances": [
      {
        "url": "http://localhost:8082",
        "weight": 2,
        "healthy": true,
        "last_check": "2024-12-20T10:30:00Z"
      }
    ],
    "health_check": "/health",
    "load_balancer": "weighted_random"
  }
}
```

**权限要求**: 管理员权限

### 7. 添加服务

**接口地址**: `POST /admin/proxy/services`

**描述**: 添加新的后端服务配置

**请求头**:

```
Cookie: access_token=your_jwt_token
Content-Type: application/json
```

**请求参数**:

```json
[
  {
    "service_name": "string", // 必填，服务名称，唯一标识
    "instances": [
      {
        "url": "string", // 必填，服务实例URL，格式: http://host:port
        "weight": 1, // 可选，权重，默认1，范围1-100
        "healthy": true // 可选，健康状态，默认true
      }
    ],
    "health_check": "string", // 必填，健康检查路径，如: /health
    "load_balancer": "string" // 必填，负载均衡策略
  }
]
```

**负载均衡策略**:

- `round_robin`: 轮询
- `random`: 随机
- `weighted_random`: 加权随机
- `weighted_round_robin`: 加权轮询

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "service added"
}

// 参数错误 (400)
{
  "error": "Key: 'ServiceConfig.ServiceName' Error:Field validation for 'ServiceName' failed on the 'required' tag"
}

// 权限不足 (403)
{
  "error": "权限不足"
}
```

**权限要求**: 管理员权限

### 8. 删除服务

**接口地址**: `DELETE /admin/proxy/services?service=服务名`

**描述**: 删除指定的服务配置

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**:

- **Query 参数**: `service` (必填，服务名称)

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "service removed"
}

// 服务不存在 (404)
{
  "error": "service not found"
}

// 权限不足 (403)
{
  "error": "权限不足"
}
```

**权限要求**: 管理员权限

### 9. 获取服务实例

**接口地址**: `GET /admin/proxy/services/{service}/instances`

**描述**: 获取指定服务的所有实例

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**:

- **路径参数**: `service` (必填，服务名称)

**响应示例**:

```json
[
  {
    "url": "http://localhost:8081",
    "weight": 1,
    "healthy": true,
    "last_check": "2024-12-20T10:30:00Z"
  },
  {
    "url": "http://localhost:8082",
    "weight": 2,
    "healthy": false,
    "last_check": "2024-12-20T10:29:45Z"
  }
]
```

**权限要求**: 管理员权限

### 10. 添加服务实例

**接口地址**: `POST /admin/proxy/services/{service}/instances`

**描述**: 为指定服务添加新的实例

**请求头**:

```
Cookie: access_token=your_jwt_token
Content-Type: application/json
```

**请求参数**:

```json
[
  {
    "url": "string", // 必填，实例URL，格式: http://host:port
    "weight": 1, // 可选，权重，默认1，范围1-100
    "healthy": true // 可选，健康状态，默认true
  }
]
```

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "instance added"
}

// 服务不存在 (404)
{
  "error": "service not found"
}

// 参数错误 (400)
{
  "error": "Key: 'ServiceInstance.URL' Error:Field validation for 'URL' failed on the 'required' tag"
}

// 权限不足 (403)
{
  "error": "权限不足"
}
```

**权限要求**: 管理员权限

### 11. 删除服务实例

**接口地址**: `DELETE /admin/proxy/services/{service}/instance?instance=实例URL`

**描述**: 删除指定服务的指定实例

**请求头**:

```
Cookie: access_token=your_jwt_token
```

**请求参数**:

- **路径参数**: `service` (必填，服务名称)
- **Query 参数**: `instance` (必填，实例 URL)

**响应示例**:

```json
// 成功响应 (200)
{
  "message": "instance removed"
}

// 实例不存在 (404)
{
  "error": "instance not found"
}

// 服务不存在 (404)
{
  "error": "service not found"
}

// 权限不足 (403)
{
  "error": "权限不足"
}
```

**权限要求**: 管理员权限

## 数据模型

### LoginRequest (登录请求)

```json
{
  "username": "string", // 用户名，必填，长度1-50字符
  "password": "string" // 密码，必填，长度6-100字符
}
```

### InfoResponse (用户信息响应)

```json
{
  "username": "string", // 用户名
  "realname": "string", // 真实姓名
  "role": 1, // 用户角色 (1=管理员, 0=普通用户)
  "status": 1 // 用户状态 (1=正常, 0=禁用)
}
```

### ServiceInstance (服务实例)

```json
{
  "url": "string", // 实例URL，必填，格式: http://host:port
  "weight": 1, // 权重，用于加权负载均衡，范围1-100
  "healthy": true, // 健康状态
  "last_check": "string" // 最后检查时间 (ISO 8601格式)
}
```

### ServiceConfig (服务配置)

```json
{
  "service_name": "string", // 服务名称，必填，唯一标识
  "instances": [], // 服务实例列表
  "health_check": "string", // 健康检查路径，必填，如: /health
  "load_balancer": "string" // 负载均衡策略，必填
}
```

### UserClaims (JWT 用户声明)

```json
{
  "user_id": 123, // 用户ID
  "ssid": "string", // 会话ID
  "user_agent": "string" // 用户代理
}
```

## 错误码说明

| HTTP 状态码 | 错误类型   | 说明             | 示例                         |
| ----------- | ---------- | ---------------- | ---------------------------- |
| 200         | 成功       | 请求成功         | 正常响应                     |
| 400         | 客户端错误 | 请求参数错误     | 参数验证失败、JSON 格式错误  |
| 401         | 认证错误   | 未认证或认证失败 | Token 不存在、Token 过期     |
| 403         | 权限错误   | 权限不足         | 非管理员访问管理接口         |
| 404         | 资源错误   | 资源不存在       | 服务不存在、实例不存在       |
| 500         | 服务器错误 | 服务器内部错误   | 数据库连接失败、业务逻辑错误 |
| 502         | 网关错误   | 后端服务错误     | 后端服务不可达、响应异常     |
| 503         | 服务不可用 | 服务不可用       | 无健康实例可用               |

## 认证机制

### JWT Token

- **传递方式**: 通过 Cookie 传递 (access_token)
- **Token 内容**: 包含用户 ID、会话 ID 和用户代理信息
- **有效期**: 可配置，默认 24 小时
- **刷新机制**: 支持 Token 自动刷新
- **会话管理**: 基于 Redis 的会话存储

### 权限控制

- **登录检查**: 大部分 API 需要用户登录
- **管理员检查**: 服务管理 API 需要管理员权限 (role=1)
- **路径白名单**: 支持配置无需认证的路径
  - `/auth/login` - 登录接口
  - `/health` - 健康检查接口

### 安全特性

- **密码加密**: 使用 bcrypt 加密存储
- **会话管理**: Redis 存储会话信息，支持单点登录控制
- **请求追踪**: 自动生成请求 ID，便于日志追踪

## 负载均衡策略

### 1. 轮询 (round_robin)

- **算法**: 按顺序轮流选择健康的服务实例
- **适用场景**: 实例性能相近的场景
- **特点**: 简单公平，分布均匀

### 2. 随机 (random)

- **算法**: 随机选择健康的服务实例
- **适用场景**: 实例性能相近，对顺序无要求
- **特点**: 实现简单，分布相对均匀

### 3. 加权随机 (weighted_random)

- **算法**: 根据实例权重进行加权随机选择
- **适用场景**: 实例性能差异较大的场景
- **特点**: 高权重实例获得更多请求

### 4. 加权轮询 (weighted_round_robin)

- **算法**: 根据实例权重进行加权轮询选择
- **适用场景**: 实例性能差异较大，需要精确控制分配比例
- **特点**: 严格按权重比例分配请求

## 健康检查

### 检查机制

- **检查间隔**: 可配置，默认 30 秒
- **检查超时**: 可配置，默认 5 秒
- **检查方式**: HTTP GET 请求到健康检查路径
- **并发检查**: 支持多实例并发健康检查

### 故障处理

- **自动故障转移**: 不健康的实例自动从负载均衡中移除
- **自动恢复**: 实例恢复健康后自动重新加入负载均衡
- **状态记录**: 记录每个实例的最后检查时间和状态

### 配置示例

```yaml
health_check:
  interval: 30s # 检查间隔
  timeout: 5s # 检查超时
  path: "/health" # 检查路径
```

## 中间件

### 1. CORS 中间件

- **功能**: 支持跨域请求配置
- **配置项**:
  - 允许的来源 (Origins)
  - 允许的方法 (Methods)
  - 允许的请求头 (Headers)
  - 是否允许凭证 (Credentials)

### 2. JWT 中间件

- **功能**: 自动验证 JWT Token
- **特性**:
  - 提取用户信息到请求上下文
  - 支持路径白名单配置
  - 自动处理 Token 过期和刷新

### 3. 日志中间件

- **功能**: 记录请求响应日志
- **特性**:
  - 支持结构化日志输出
  - 包含请求 ID、用户 ID 等上下文信息
  - 记录请求耗时和响应状态

## 使用示例

### 1. 用户登录

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}' \
  -c cookies.txt
```

### 2. 获取用户信息

```bash
curl -X GET http://localhost:8080/auth/info \
  -b cookies.txt
```

### 3. 添加服务 (管理员)

```bash
curl -X POST http://localhost:8080/admin/proxy/services \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '[{
    "service_name": "user-service",
    "instances": [{"url": "http://localhost:8081", "weight": 1}],
    "health_check": "/health",
    "load_balancer": "round_robin"
  }]'
```

### 4. 服务转发

```bash
curl -X GET "http://localhost:8080/api/user-service?cmd=users/profile" \
  -b cookies.txt
```

### 5. 获取所有服务 (管理员)

```bash
curl -X GET http://localhost:8080/admin/proxy/services \
  -b cookies.txt
```

### 6. 健康检查

```bash
curl -X GET http://localhost:8080/health
```

## 部署说明

### 环境变量

```bash
# Gin运行模式
GIN_MODE=release

# 数据库连接
DB_DSN=user:password@tcp(localhost:3306)/online_judge?charset=utf8mb4&parseTime=True&loc=Local

# Redis配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=your_jwt_secret_key

# 服务配置
SERVER_PORT=8080
LOG_LEVEL=info
```

### Docker 部署

```bash
# 构建镜像
docker build -t online-judge-gateway .

# 运行容器
docker run -d \
  --name gateway \
  -p 8080:8080 \
  -e DB_DSN="user:password@tcp(db:3306)/online_judge" \
  -e REDIS_ADDR="redis:6379" \
  online-judge-gateway
```

### 配置文件

参考 `config/config.template.yaml` 进行配置：

```yaml
server:
  port: 8080
  mode: release

database:
  dsn: "user:password@tcp(localhost:3306)/online_judge"

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

jwt:
  secret: "your_jwt_secret"
  expire: 24h

proxy:
  health_check_interval: 30s
  health_check_timeout: 5s
```

### 计划中的功能

- 🔄 服务发现集成 (Consul/Etcd)
- 🔄 API 限流和熔断机制
- 🔄 监控和指标收集
- 🔄 配置热重载
- 🔄 gRPC 服务代理支持

---

**注意**:

1. 本文档基于当前代码结构生成，如有 API 变更请及时更新文档
2. 生产环境部署前请仔细配置安全相关参数
3. 建议定期备份配置和日志文件
