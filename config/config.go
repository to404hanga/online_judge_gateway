package config

import (
	"github.com/to404hanga/online_judge_gateway/web/middleware"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

type LoggerConfig struct {
	Development    bool                `yaml:"development"`    // 是否为开发模式
	Type           loggerv2.OutputType `yaml:"type"`           // 日志输出类型
	LogFilePath    string              `yaml:"logFilePath"`    // 日志文件路径
	AutoCreateFile bool                `yaml:"autoCreateFile"` // 是否自动创建文件和目录
}

func (LoggerConfig) Key() string {
	return "log"
}

type GinConfig struct {
	AllowOrigins        []string                    `yaml:"allowOrigins"`        // 允许的来源，* 表示所有来源
	AllowMethods        []string                    `yaml:"allowMethods"`        // 允许的方法，* 表示所有方法
	AllowHeaders        []string                    `yaml:"allowHeaders"`        // 允许的请求头，* 表示所有请求头
	ExposeHeaders       []string                    `yaml:"exposeHeaders"`       // 暴露的响应头，* 表示所有响应头
	AllowCredentials    bool                        `yaml:"allowCredentials"`    // 是否允许携带凭证（如 Cookies）
	MaxAge              int64                       `yaml:"maxAge"`              // 预检请求的缓存时间（单位: 秒）
	LoginCheckPassPairs []middleware.PathMethodPair `yaml:"loginCheckPassPairs"` // 绕过登录校验路径
	AdminCheckPairs     []middleware.PathMethodPair `yaml:"adminCheckPairs"`     // 管理员校验路径
	Addr                string                      `yaml:"addr"`                // 服务地址
}

func (GinConfig) Key() string {
	return "gin"
}

type DBConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DBName      string `yaml:"dbName"`
	TablePrefix string `yaml:"tablePrefix"`
	// 连接池配置
	MaxOpenConns    int `yaml:"maxOpenConns"`    // 最大打开连接数
	MaxIdleConns    int `yaml:"maxIdleConns"`    // 最大空闲连接数
	ConnMaxLifetime int `yaml:"connMaxLifetime"` // 连接最大生存时间（分钟）
	ConnMaxIdleTime int `yaml:"connMaxIdleTime"` // 连接最大空闲时间（分钟）
}

func (DBConfig) Key() string {
	return "db"
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

func (RedisConfig) Key() string {
	return "redis"
}

type JWTConfig struct {
	JWTExpiration int    `yaml:"jwtExpiration"` // jwt token 有效期（单位: 分钟）
	JWTKey        string `yaml:"jwtKey"`        // jwt 密钥
}

func (JWTConfig) Key() string {
	return "jwt"
}

type ProxyConfig struct {
	Services []string `yaml:"services"` // 服务配置
}

func (ProxyConfig) Key() string {
	return "proxy"
}

type LRUConfig struct {
	Size int `yaml:"size"` // 缓存中可容纳的项数
}

func (LRUConfig) Key() string {
	return "lru"
}
