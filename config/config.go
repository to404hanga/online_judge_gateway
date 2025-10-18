package config

import (
	"github.com/to404hanga/online_judge_gateway/web/middleware"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

type LoggerConfig struct {
	Development    bool                `yaml:"development"`      // 是否为开发模式
	Type           loggerv2.OutputType `yaml:"type"`             // 日志输出类型
	LogFilePath    string              `yaml:"log_file_path"`    // 日志文件路径
	AutoCreateFile bool                `yaml:"auto_create_file"` // 是否自动创建文件和目录
}

func (LoggerConfig) Key() string {
	return "log"
}

type GinConfig struct {
	AllowOrigins        []string                    `yaml:"allow_origins"`          // 允许的来源，* 表示所有来源
	AllowMethods        []string                    `yaml:"allow_methods"`          // 允许的方法，* 表示所有方法
	AllowHeaders        []string                    `yaml:"allow_headers"`          // 允许的请求头，* 表示所有请求头
	ExposeHeaders       []string                    `yaml:"expose_headers"`         // 暴露的响应头，* 表示所有响应头
	AllowCredentials    bool                        `yaml:"allow_credentials"`      // 是否允许携带凭证（如 Cookies）
	MaxAge              int64                       `yaml:"max_age"`                // 预检请求的缓存时间（单位: 秒）
	LoginCheckPassPairs []middleware.PathMethodPair `yaml:"login_check_pass_pairs"` // 绕过登录校验路径
	AdminCheckPairs     []middleware.PathMethodPair `yaml:"admin_check_pairs"`      // 管理员校验路径
	Addr                string                      `yaml:"addr"`                   // 服务地址
}

func (GinConfig) Key() string {
	return "gin"
}

type DBConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DBName      string `yaml:"database"`
	TablePrefix string `yaml:"table_prefix"`
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
	JWTExpiration     int    `yaml:"jwt_expiration"`     // jwt token 有效期（单位: 分钟）
	RefreshExpiration int    `yaml:"refresh_expiration"` // refresh token 有效期（单位: 分钟）
	JWTKey            string `yaml:"jwt_key"`            // jwt 密钥
	RefreshKey        string `yaml:"refresh_key"`        // refresh token 密钥
}

func (JWTConfig) Key() string {
	return "jwt"
}

type ProxyConfig struct {
	HealthCheckInterval int `yaml:"health_check_interval"` // 健康检查间隔（单位: 秒）
	HealthCheckTimeout  int `yaml:"health_check_timeout"`  // 健康检查超时时间（单位: 秒）
}

func (ProxyConfig) Key() string {
	return "proxy"
}
