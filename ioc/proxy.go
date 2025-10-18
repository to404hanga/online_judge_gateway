package ioc

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/to404hanga/online_judge_gateway/config"
	"github.com/to404hanga/online_judge_gateway/web"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

func InitProxyHandler(l loggerv2.Logger) *web.ProxyHandler {
	var cfg config.ProxyConfig
	if err := viper.UnmarshalKey(cfg.Key(), &cfg); err != nil {
		log.Panicf("unmarshal proxy config failed: %v", err)
	}

	// 设置默认值，防止配置为0导致panic
	healthCheckInterval := cfg.HealthCheckInterval
	if healthCheckInterval <= 0 {
		healthCheckInterval = 60 // 默认60秒
		l.Warn("health_check_interval not configured or invalid, using default value: 60s")
	}

	healthCheckTimeout := cfg.HealthCheckTimeout
	if healthCheckTimeout <= 0 {
		healthCheckTimeout = 5 // 默认5秒
		l.Warn("health_check_timeout not configured or invalid, using default value: 5s")
	}

	services := make(map[string]*web.ServiceConfig, len(cfg.Services))
	for _, cfgSvc := range cfg.Services {
		svc := &web.ServiceConfig{
			ServiceName:  cfgSvc.ServiceName,
			HealthCheck:  cfgSvc.HealthCheck,
			LoadBalancer: web.LoadBalancerType(cfgSvc.LoadBalancer),
		}
		svc.Instances = make([]*web.ServiceInstance, len(cfgSvc.Instances))
		for idx, cfgInst := range cfgSvc.Instances {
			svc.Instances[idx] = &web.ServiceInstance{
				URL:     cfgInst.URL,
				Weight:  cfgInst.Weight,
				Healthy: cfgInst.Healthy,
			}
		}
		services[cfgSvc.ServiceName] = svc
	}

	return web.NewProxyHandler(
		time.Duration(healthCheckInterval)*time.Second,
		time.Duration(healthCheckTimeout)*time.Second,
		l,
		services,
	)
}
