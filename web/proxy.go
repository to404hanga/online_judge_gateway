package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	constants "github.com/to404hanga/online_judge_gateway/constant"
	"github.com/to404hanga/online_judge_gateway/web/jwt"
	"github.com/to404hanga/online_judge_gateway/web/middleware"
	"github.com/to404hanga/pkg404/logger"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

type ProxyHandler struct {
	services map[string]string
	log      loggerv2.Logger
}

var _ Handler = (*ProxyHandler)(nil)

func NewProxyHandler(log loggerv2.Logger, services map[string]string) *ProxyHandler {
	return &ProxyHandler{
		services: services,
		log:      log,
	}
}

func (h *ProxyHandler) Register(r *gin.Engine) {
	r.Any("/api/*path", middleware.Logger(h.log), h.ProxyHandler) // 转发路由不使用日志中间件
}

func (h *ProxyHandler) ProxyHandler(c *gin.Context) {
	path := c.Param("path")

	ucAny, exists := c.Get(constants.ContextUserClaimsKey)
	if !exists {
		h.log.ErrorContext(c, "user claims not found in context",
			logger.String("service_path", path),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user claims not found in context"})
		return
	}
	uc, ok := ucAny.(jwt.UserClaims)
	if !ok {
		h.log.ErrorContext(c, "user claims type assertion error",
			logger.String("service_path", path),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user claims type assertion error"})
		return
	}

	path = strings.TrimPrefix(path, "/")
	target := "http://" + h.services[path]
	if len(target) == 0 {
		h.log.ErrorContext(c, "service not found",
			logger.String("service_path", path),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}
	targetURL, err := url.Parse(target)
	if err != nil {
		h.log.ErrorContext(c, "parse target url error",
			logger.String("service_path", path),
			logger.String("target", target),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse target url error"})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义请求修改
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		query := req.URL.Query()
		cmd := query.Get(constants.ProxyKey)
		if len(cmd) != 0 {
			// 重写请求路径为 cmd 值
			req.URL.Path = "/" + cmd

			// 移除 cmd 参数
			query.Del(constants.ProxyKey)
			req.URL.RawQuery = query.Encode()
		} else {
			h.log.ErrorContext(req.Context(), "request missing cmd parameter",
				logger.String("service_path", path),
			)
		}

		req.Header.Set(constants.HeaderForwardedByKey, constants.GatewayServiceName)
		req.Header.Set(constants.HeaderRequestIDKey, generateRequestID())
		req.Header.Set(constants.HeaderUserIDKey, strconv.FormatUint(uc.UserId, 10))
	}

	// 响应修改
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set(constants.HeaderProxyByKey, constants.GatewayServiceName)
		return nil
	}

	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.log.ErrorContext(r.Context(), "proxy error",
			logger.String("target", target),
			logger.Error(err),
		)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"backend service error"}`))
	}

	h.log.InfoContext(c, "proxying request",
		logger.String("method", c.Request.Method),
		logger.String("target", target),
	)

	// 执行代理
	proxy.ServeHTTP(c.Writer, c.Request)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}
