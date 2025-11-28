package ioc

import (
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/to404hanga/online_judge_gateway/config"
	"github.com/to404hanga/online_judge_gateway/web"
	"github.com/to404hanga/pkg404/gotools/transform"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

func InitProxyHandler(l loggerv2.Logger) *web.ProxyHandler {
	var cfg config.ProxyConfig
	if err := viper.UnmarshalKey(cfg.Key(), &cfg); err != nil {
		log.Panicf("unmarshal proxy config failed: %v", err)
	}

	services := transform.MapFromSlice(cfg.Services, func(i int, svc string) (string, string) {
		parts := strings.Split(svc, ":")
		if len(parts) != 2 {
			log.Panicf("invalid service config: %s", svc)
		}
		return parts[0], svc
	})

	return web.NewProxyHandler(l, services)
}
