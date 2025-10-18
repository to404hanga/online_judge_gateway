package ioc

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/to404hanga/online_judge_gateway/config"
	"github.com/to404hanga/online_judge_gateway/web"
	"github.com/to404hanga/online_judge_gateway/web/jwt"
	"github.com/to404hanga/online_judge_gateway/web/middleware"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
	"gorm.io/gorm"
)

func InitGinServer(l loggerv2.Logger, jwtHandler jwt.Handler, db *gorm.DB, authHandler *web.AuthHandler, proxyHandler *web.ProxyHandler) *web.GinServer {
	var cfg config.GinConfig
	err := viper.UnmarshalKey(cfg.Key(), &cfg)
	if err != nil {
		log.Panicf("unmarshal gin config failed, err: %v", err)
	}

	corsBuilder := middleware.NewCORSMiddlewareBuilder(
		cfg.AllowOrigins,
		cfg.AllowMethods,
		cfg.AllowHeaders,
		cfg.ExposeHeaders,
		cfg.AllowCredentials,
		time.Duration(cfg.MaxAge)*time.Second)
	jwtBuilder := middleware.NewJWTMiddlewareBuilder(jwtHandler, db, cfg.LoginCheckPassPairs, cfg.AdminCheckPairs, l)

	engine := gin.Default()
	engine.Use(
		corsBuilder.Build(),
		jwtBuilder.CheckLogin(),
		jwtBuilder.CheckAdmin(),
	)

	authHandler.Register(engine)
	proxyHandler.Register(engine)

	return &web.GinServer{
		Engine: engine,
		Addr:   cfg.Addr,
	}
}
