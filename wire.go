//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/to404hanga/online_judge_gateway/ioc"
	"github.com/to404hanga/online_judge_gateway/service"
	"github.com/to404hanga/online_judge_gateway/web"
)

func BuildDependency() *web.GinServer {
	wire.Build(
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitRedis,
		ioc.InitJWTHandler,
		ioc.InitProxyHandler,
		ioc.InitLRUCache,

		service.NewAuthService,

		web.NewAuthHandler,

		ioc.InitGinServer,
	)
	return &web.GinServer{}
}
