package ioc

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/to404hanga/onlinue_judge_gateway/config"
	"github.com/to404hanga/onlinue_judge_gateway/web"
	"github.com/to404hanga/onlinue_judge_gateway/web/jwt"
	"github.com/to404hanga/onlinue_judge_gateway/web/middleware"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

func InitGinServer(l loggerv2.Logger, jwtHandler jwt.Handler, authHandler *web.AuthHandler, proxyHandler *web.ProxyHandler) *web.GinServer {
	var cfg config.GinConfig
	err := viper.UnmarshalKey(cfg.Key(), &cfg)
	if err != nil {
		log.Panicf("unmarshal gin config failed, err: %v", err)
	}

	engine := gin.Default()
	engine.Use(
		middleware.NewCORSMiddlewareBuilder(
			cfg.AllowOrigins,
			cfg.AllowMethods,
			cfg.AllowHeaders,
			cfg.ExposeHeaders,
			cfg.AllowCredentials,
			time.Duration(cfg.MaxAge)*time.Second).Build(),
		middleware.NewJWTMiddlewareBuilder(jwtHandler, cfg.PassPathMethodPairs).CheckLogin(),
	)

	authHandler.Register(engine)
	proxyHandler.Register(engine)

	return &web.GinServer{
		Engine: engine,
		Addr:   cfg.Addr,
	}
}
