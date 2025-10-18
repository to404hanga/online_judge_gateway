package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/to404hanga/pkg404/logger"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

func main() {
	cport := pflag.Int("port", 8080, "port to listen on")
	pflag.Parse()

	port := ":" + strconv.Itoa(*cport)

	r := gin.Default()

	l := loggerv2.MustNewFileLogger(fmt.Sprintf("./log/%d.log", *cport), true, true)

	r.GET("/health", func(ctx *gin.Context) {
		l.Info("health check")
		ctx.Status(http.StatusOK)
	})

	r.GET("/test", func(ctx *gin.Context) {
		l.Info("test check")
		// 打印所有查询字符串
		for k, v := range ctx.Request.URL.Query() {
			l.Info("query param", logger.String("key", k), logger.Slice("value", v))
		}
		// 打印所有请求头
		for k, v := range ctx.Request.Header {
			l.Info("header", logger.String("key", k), logger.Slice("value", v))
		}
		// 打印请求体
		body, _ := ctx.GetRawData()
		l.Info("request body", logger.String("body", string(body)))
		ctx.JSON(http.StatusOK, gin.H{
			"message": "test check",
		})
	})

	r.Run(port)
}
