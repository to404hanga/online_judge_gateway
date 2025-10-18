# 在线判题系统网关 API 文档

## 项目概述

这是一个基于 Gin + GORM + Redis 的微服务网关，为在线判题系统提供统一的入口点，包含用户认证、服务代理和负载均衡功能。

## 技术栈

- **框架**: Gin + GORM + Redis
- **认证**: JWT (JSON Web Token)
- **负载均衡**: 支持轮询、随机、加权随机
- **健康检查**: 自动服务实例健康监控

## 基础信息

- **Base URL**: `http://localhost:8080` (默认)
- **Content-Type**: `application/json`
- **认证方式**: JWT Token (通过 Header)

## 认证相关 API

### 1. 用户登录

**接口地址**: `POST /auth/login`

**描述**: 用户登录接口，验证用户名和密码，成功后返回JWT Token

**请求参数**:
```json
{
  "username": "string",  // 必填，用户名
  "password": "string"   // 必填，密码
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
```

**说明**: 
- 登录成功后，JWT Token 会自动设置到 X-JWT-Token Header 中
- Token 包含用户ID、会话ID和用户代理信息

### 2. 用户登出

**接口地址**: `POST /auth/logout`

**描述**: 用户登出接口，清除JWT Token

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
```

### 3. 获取用户信息

**接口地址**: `GET /auth/info`

**描述**: 获取当前登录用户的详细信息

**请求参数**: 无 (需要JWT认证)

**响应示例**:
```json
// 成功响应 (200)
{
  "username": "string",    // 用户名
  "realname": "string",    // 真实姓名
  "role": "string",        // 用户角色 (admin/user)
  "status": "string",      // 用户状态
  "created_at": "2024-01-01T00:00:00Z",  // 创建时间
  "updated_at": "2024-01-01T00:00:00Z"   // 更新时间
}

// 失败响应 (500)
{
  "error": "获取用户信息失败"
}
```

## 服务代理 API

### 4. 服务转发

**接口地址**: `ANY /api/*path`

**描述**: 将请求转发到后端微服务，支持负载均衡

**请求参数**: 
- 路径参数中的服务路径会被用于匹配对应的后端服务
- 支持所有HTTP方法 (GET, POST, PUT, DELETE等)

**响应**: 直接返回后端服务的响应

**说明**:
- 需要JWT认证
- 自动添加用户信息到请求头
- 支持多种负载均衡策略
- 自动健康检查和故障转移

## 服务管理 API (管理员接口)

### 5. 获取所有服务

**接口地址**: `GET /admin/proxy/services`

**描述**: 获取当前网关管理的所有服务配置

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
        "last_check": "2024-01-01T00:00:00Z"
      }
    ],
    "health_check": "/health",
    "load_balancer": "round_robin"
  }
}
```

### 6. 添加服务

**接口地址**: `POST /admin/proxy/services`

**描述**: 添加新的后端服务配置

**请求参数**:
```json
[
  {
    "service_name": "string",        // 必填，服务名称
    "instances": [
      {
        "url": "string",             // 必填，服务实例URL
        "weight": 1,                 // 权重，默认1
        "healthy": true              // 健康状态，默认true
      }
    ],
    "health_check": "string",        // 必填，健康检查路径
    "load_balancer": "string"        // 必填，负载均衡策略
  }
]
```

**负载均衡策略**:
- `round_robin`: 轮询
- `random`: 随机
- `weighted_random`: 加权随机

**响应示例**:
```json
// 成功响应 (200)
{
  "message": "service added"
}
```

### 7. 删除服务

**接口地址**: `DELETE /admin/proxy/services?service=服务名`

**描述**: 删除指定的服务配置

**请求参数**: 
- Query参数: `service` (必填，服务名称)

**响应示例**:
```json
// 成功响应 (200)
{
  "message": "service removed"
}

// 失败响应 (404)
{
  "error": "service not found"
}
```

### 8. 获取服务实例

**接口地址**: `GET /admin/proxy/services/{service}/instances`

**描述**: 获取指定服务的所有实例

**请求参数**: 
- 路径参数: `service` (必填，服务名称)

**响应示例**:
```json
[
  {
    "url": "http://localhost:8081",
    "weight": 1,
    "healthy": true,
    "last_check": "2024-01-01T00:00:00Z"
  },
  {
    "url": "http://localhost:8082",
    "weight": 2,
    "healthy": false,
    "last_check": "2024-01-01T00:00:00Z"
  }
]
```

### 9. 添加服务实例

**接口地址**: `POST /admin/proxy/services/{service}/instances`

**描述**: 为指定服务添加新的实例

**请求参数**:
```json
[
  {
    "url": "string",      // 必填，实例URL
    "weight": 1,          // 权重，默认1
    "healthy": true       // 健康状态，默认true
  }
]
```

**响应示例**:
```json
// 成功响应 (200)
{
  "message": "instance added"
}

// 失败响应 (404)
{
  "error": "service not found"
}
```

### 10. 删除服务实例

**接口地址**: `DELETE /admin/proxy/services/{service}/instance?instance=实例URL`

**描述**: 删除指定服务的指定实例

**请求参数**: 
- 路径参数: `service` (必填，服务名称)
- Query参数: `instance` (必填，实例URL)

**响应示例**:
```json
// 成功响应 (200)
{
  "message": "instance removed"
}

// 失败响应 (404)
{
  "error": "instance not found"
}
```

## 数据模型

### LoginRequest (登录请求)
```json
{
  "username": "string",  // 用户名，必填
  "password": "string"   // 密码，必填
}
```

### InfoResponse (用户信息响应)
```json
{
  "username": "string",    // 用户名
  "realname": "string",    // 真实姓名
  "role": "string",        // 用户角色
  "status": "string",      // 用户状态
  "created_at": "string",  // 创建时间 (ISO 8601)
  "updated_at": "string"   // 更新时间 (ISO 8601)
}
```

### ServiceInstance (服务实例)
```json
{
  "url": "string",         // 实例URL，必填
  "weight": 1,             // 权重，用于加权负载均衡
  "healthy": true,         // 健康状态
  "last_check": "string"   // 最后检查时间 (ISO 8601)
}
```

### ServiceConfig (服务配置)
```json
{
  "service_name": "string",      // 服务名称，必填
  "instances": [],               // 服务实例列表
  "health_check": "string",      // 健康检查路径，必填
  "load_balancer": "string"      // 负载均衡策略，必填
}
```

### UserClaims (JWT用户声明)
```json
{
  "user_id": 123,          // 用户ID
  "ssid": "string",        // 会话ID
  "user_agent": "string"   // 用户代理
}
```

## 错误码说明

| HTTP状态码 | 说明 |
|-----------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或认证失败 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 (无健康实例) |

## 认证机制

### JWT Token
- Token 通过 Cookie 传递
- Token 包含用户ID、会话ID和用户代理信息
- 支持 Token 刷新机制
- 会话管理基于 Redis

### 权限控制
- **登录检查**: 大部分API需要用户登录
- **管理员检查**: 服务管理API需要管理员权限
- **路径白名单**: 支持配置无需认证的路径

## 负载均衡策略

### 1. 轮询 (round_robin)
按顺序轮流选择健康的服务实例

### 2. 随机 (random)
随机选择健康的服务实例

### 3. 加权随机 (weighted_random)
根据实例权重进行加权随机选择

## 健康检查

- **检查间隔**: 可配置 (默认30秒)
- **检查超时**: 可配置 (默认5秒)
- **检查方式**: HTTP GET 请求到健康检查路径
- **自动故障转移**: 不健康的实例自动从负载均衡中移除

## 中间件

### 1. CORS 中间件
- 支持跨域请求配置
- 可配置允许的来源、方法、请求头等

### 2. JWT 中间件
- 自动验证JWT Token
- 提取用户信息到请求上下文
- 支持路径白名单配置

### 3. 日志中间件
- 记录请求响应日志
- 支持结构化日志输出
- 包含请求ID、用户ID等上下文信息

## 使用示例

### 1. 用户登录
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

### 2. 获取用户信息
```bash
curl -X GET http://localhost:8080/auth/info \
  -H "Cookie: access_token=your_jwt_token"
```

### 3. 添加服务
```bash
curl -X POST http://localhost:8080/admin/proxy/services \
  -H "Content-Type: application/json" \
  -H "Cookie: access_token=your_jwt_token" \
  -d '[{
    "service_name": "user-service",
    "instances": [{"url": "http://localhost:8081"}],
    "health_check": "/health",
    "load_balancer": "round_robin"
  }]'
```

### 4. 服务转发
```bash
curl -X GET http://localhost:8080/api/user-service/users \
  -H "Cookie: access_token=your_jwt_token"
```

## 部署说明

### 环境变量
- `GIN_MODE`: Gin运行模式 (debug/release)
- `DB_DSN`: 数据库连接字符串
- `REDIS_ADDR`: Redis地址
- `JWT_SECRET`: JWT签名密钥

### Docker 部署
项目包含 Dockerfile，支持容器化部署。

### 配置文件
参考 `config/config.template.yaml` 进行配置。

---

**注意**: 本文档基于当前代码结构生成，如有API变更请及时更新文档。