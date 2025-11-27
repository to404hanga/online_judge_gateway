package ioc

import (
	"log"

	"github.com/spf13/viper"
	"github.com/to404hanga/online_judge_gateway/config"
	"github.com/to404hanga/online_judge_gateway/web"
	"github.com/to404hanga/pkg404/gotools/transform"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func InitProxyHandler(l loggerv2.Logger, etcdCli *clientv3.Client) *web.ProxyHandler {
	var cfg config.ProxyConfig
	if err := viper.UnmarshalKey(cfg.Key(), &cfg); err != nil {
		log.Panicf("unmarshal proxy config failed: %v", err)
	}

	svcCfgs := transform.SliceFromSlice(cfg.Services, func(i int, cfgSvc config.ServiceConfig) web.SetServiceConfig {
		return web.SetServiceConfig{
			Prefix:       cfgSvc.Prefix,
			LoadBalancer: web.LoadBalancerType(cfgSvc.LoadBalancer),
		}
	})
	return web.NewProxyHandler(l, etcdCli, svcCfgs)
}
