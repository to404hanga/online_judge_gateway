package ioc

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/to404hanga/onlinue_judge_gateway/config"
	"github.com/to404hanga/onlinue_judge_gateway/web"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

func InitProxyHandler(l loggerv2.Logger) *web.ProxyHandler {
	var cfg config.ProxyConfig
	if err := viper.UnmarshalKey(cfg.Key(), &cfg); err != nil {
		log.Panicf("unmarshal proxy config failed: %v", err)
	}

	return web.NewProxyHandler(
		time.Duration(cfg.HealthCheckInterval)*time.Second,
		time.Duration(cfg.HealthCheckTimeout)*time.Second,
		l)
}
