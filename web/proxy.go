package web

import (
	"context"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	constants "github.com/to404hanga/online_judge_gateway/constant"
	"github.com/to404hanga/online_judge_gateway/web/jwt"
	"github.com/to404hanga/pkg404/logger"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	URL       string    `json:"url" binding:"required"`
	Weight    int       `json:"weight"`     // 权重
	Healthy   bool      `json:"healthy"`    // 是否健康
	LastCheck time.Time `json:"last_check"` // 最后检查时间
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceName  string             `json:"service_name" binding:"required"`  // 服务名称
	Instances    []*ServiceInstance `json:"instances"`                        // 服务实例列表
	HealthCheck  string             `json:"health_check" binding:"required"`  // 健康检查路径
	LoadBalancer LoadBalancerType   `json:"load_balancer" binding:"required"` // 负载均衡策略
	mux          sync.RWMutex
	currentIndex int // 当前轮询索引
}

type LoadBalancerType string

const (
	LoadBalancerTypeRoundRobin         LoadBalancerType = "round_robin"          // 轮询负载均衡
	LoadBalancerTypeRandom             LoadBalancerType = "random"               // 随机负载均衡
	LoadBalancerTypeWeightedRandom     LoadBalancerType = "weighted_random"      // 加权随机负载均衡
	LoadBalancerTypeWeightedRoundRobin LoadBalancerType = "weighted_round_robin" // 加权轮询负载均衡
)

type ProxyHandler struct {
	services            map[string]*ServiceConfig
	mux                 sync.RWMutex
	healthCheckInterval time.Duration
	healthCheckTimeout  time.Duration
	log                 loggerv2.Logger
}

var _ Handler = (*ProxyHandler)(nil)

func NewProxyHandler(healthCheckInterval, healthCheckTimeout time.Duration, log loggerv2.Logger, services map[string]*ServiceConfig) *ProxyHandler {
	handler := &ProxyHandler{
		services:            services,
		healthCheckInterval: healthCheckInterval,
		healthCheckTimeout:  healthCheckTimeout,
		log:                 log,
	}

	// 启动健康检查
	go handler.startHealthCheck()

	return handler
}

func (h *ProxyHandler) Register(r *gin.Engine) {
	r.Any("/api/*path", h.ProxyHandler) // 转发路由不使用日志中间件

	// 管理接口, 已弃用
	// admin := r.Group("/admin/proxy").Use(middleware.Logger(h.log))
	// {
	// 	admin.GET("/services", h.GetServicesHandler)
	// 	admin.POST("/services", h.AddServiceHandler)
	// 	admin.DELETE("/services", h.RemoveServiceHandler)

	// 	admin.GET("/services/:service/instances", h.GetServiceInstancesHandler)
	// 	admin.POST("/services/:service/instances", h.AddInstancesHandler)
	// 	admin.DELETE("/services/:service/instance", h.RemoveInstanceHandler)
	// }
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

	// 获取服务配置
	serviceConfig := h.getServiceConfig(path)
	if serviceConfig == nil {
		h.log.ErrorContext(c, "service config not found",
			logger.String("service_path", path),
		)
		c.JSON(404, gin.H{"error": "service not found"})
		return
	}

	// 选择健康的实例
	instance := h.selectInstance(serviceConfig)
	if instance == nil {
		h.log.ErrorContext(c, "no healthy instance found",
			logger.String("service_path", path),
		)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no healthy instance found"})
		return
	}

	// 解析目标 URL
	target, err := url.Parse(instance.URL)
	if err != nil {
		h.log.ErrorContext(c, "parse instance url error",
			logger.String("service_path", path),
			logger.String("target", instance.URL),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse instance url error"})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

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
			logger.String("target", instance.URL),
			logger.Error(err),
		)
		// 标记实例为不健康
		instance.Healthy = false
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"backend service error"}`))
	}

	h.log.InfoContext(c, "proxying request",
		logger.String("method", c.Request.Method),
		logger.String("target", instance.URL),
	)

	// 执行代理
	proxy.ServeHTTP(c.Writer, c.Request)
}

// getServiceConfig 根据路径获取服务配置
func (h *ProxyHandler) getServiceConfig(path string) *ServiceConfig {
	h.mux.RLock()
	defer h.mux.RUnlock()

	for prefix, config := range h.services {
		if strings.Contains(path, prefix) {
			return config
		}
	}
	return nil
}

// selectInstance 根据负载均衡策略选择实例
func (h *ProxyHandler) selectInstance(config *ServiceConfig) *ServiceInstance {
	config.mux.Lock()
	defer config.mux.Unlock()

	healthyInstances := make([]*ServiceInstance, 0, len(config.Instances))
	for _, instance := range config.Instances {
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	switch config.LoadBalancer {
	case LoadBalancerTypeRoundRobin:
		instance := healthyInstances[config.currentIndex%len(healthyInstances)]
		config.currentIndex++
		return instance
	case LoadBalancerTypeRandom:
		return healthyInstances[rand.IntN(len(healthyInstances))]
	case LoadBalancerTypeWeightedRandom:
		return h.selectWeightedRandomInstance(healthyInstances)
	case LoadBalancerTypeWeightedRoundRobin:
		instance := h.selectWeightedRoundRobinInstance(healthyInstances, config.currentIndex)
		config.currentIndex++
		return instance
	default:
		return healthyInstances[0]
	}
}

// selectWeightedRandomInstance 根据加权随机负载均衡策略选择实例
func (h *ProxyHandler) selectWeightedRandomInstance(instances []*ServiceInstance) *ServiceInstance {
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}
	if totalWeight == 0 {
		return instances[0]
	}
	randWeight := rand.IntN(totalWeight)
	for _, instance := range instances {
		if randWeight < instance.Weight {
			return instance
		}
		randWeight -= instance.Weight
	}
	return instances[0]
}

func (h *ProxyHandler) selectWeightedRoundRobinInstance(instances []*ServiceInstance, currentIndex int) *ServiceInstance {
	totalWeight := 0
	for _, instance := range instances {
		totalWeight += instance.Weight
	}
	if totalWeight == 0 {
		return instances[0]
	}
	targetWeight := currentIndex % totalWeight
	for _, instance := range instances {
		if targetWeight < instance.Weight {
			return instance
		}
		targetWeight -= instance.Weight
	}
	return instances[0]
}

// startHealthCheck 启动健康检查
func (h *ProxyHandler) startHealthCheck() {
	ticker := time.NewTicker(h.healthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		h.performHealthCheck()
	}
}

// performHealthCheck 执行健康检查
func (h *ProxyHandler) performHealthCheck() {
	h.mux.RLock()
	services := make(map[string]*ServiceConfig)
	for k, v := range h.services {
		services[k] = v
	}
	h.mux.RUnlock()

	for serviceName, config := range services {
		for _, instance := range config.Instances {
			go h.checkInstanceHealth(serviceName, instance, config.HealthCheck)
		}
	}
}

// checkInstanceHealth 检查实例健康状态
func (h *ProxyHandler) checkInstanceHealth(serviceName string, instance *ServiceInstance, healthPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), h.healthCheckTimeout)
	defer cancel()

	healthURL := instance.URL + healthPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		instance.Healthy = false
		h.log.WarnContext(ctx, "create health check request error",
			logger.String("service_name", serviceName),
			logger.String("instance_url", instance.URL),
			logger.Error(err),
		)
		return
	}

	client := &http.Client{Timeout: h.healthCheckTimeout}
	resp, err := client.Do(req)
	if err != nil {
		instance.Healthy = false
		h.log.WarnContext(ctx, "health check request error",
			logger.String("service_name", serviceName),
			logger.String("instance_url", instance.URL),
			logger.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	instance.Healthy = resp.StatusCode == http.StatusOK
	instance.LastCheck = time.Now()

	if !instance.Healthy {
		h.log.WarnContext(ctx, "health unhealthy",
			logger.String("service_name", serviceName),
			logger.String("instance_url", instance.URL),
			logger.Int("status_code", resp.StatusCode),
		)
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}

func (h *ProxyHandler) GetServicesHandler(c *gin.Context) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	c.JSON(http.StatusOK, h.services)
}

func (h *ProxyHandler) AddServiceHandler(c *gin.Context) {
	var services []ServiceConfig
	if err := c.ShouldBindJSON(&services); err != nil {
		h.log.ErrorContext(c, "bind json error",
			logger.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mux.Lock()
	defer h.mux.Unlock()

	for i := range services {
		h.services[services[i].ServiceName] = &services[i]
		h.log.InfoContext(c, "add service success",
			logger.String("service", services[i].ServiceName),
			logger.String("health_check", services[i].HealthCheck),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "service added"})
}

func (h *ProxyHandler) RemoveServiceHandler(c *gin.Context) {
	serviceName := c.Query("service")

	h.mux.Lock()
	defer h.mux.Unlock()

	_, exists := h.services[serviceName]
	if !exists {
		h.log.ErrorContext(c, "service not found",
			logger.String("service", serviceName),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	delete(h.services, serviceName)
	c.JSON(http.StatusOK, gin.H{"message": "service removed"})
}

func (h *ProxyHandler) GetServiceInstancesHandler(c *gin.Context) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	serviceName := c.Param("service")
	config, exists := h.services[serviceName]
	if !exists {
		h.log.ErrorContext(c, "service not found",
			logger.String("service", serviceName),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	c.JSON(http.StatusOK, config.Instances)
}

func (h *ProxyHandler) AddInstancesHandler(c *gin.Context) {
	serviceName := c.Param("service")
	var instances []ServiceInstance
	if err := c.ShouldBindJSON(&instances); err != nil {
		h.log.ErrorContext(c, "bind json error",
			logger.Error(err),
			logger.String("service", serviceName),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.mux.Lock()
	defer h.mux.Unlock()

	config, exists := h.services[serviceName]
	if !exists {
		h.log.ErrorContext(c, "service not found",
			logger.String("service", serviceName),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	for _, instance := range instances {
		config.Instances = append(config.Instances, &instance)
		h.log.InfoContext(c, "add instance success",
			logger.String("service", serviceName),
			logger.String("instance", instance.URL),
			logger.Int("weight", instance.Weight),
			logger.Bool("healthy", instance.Healthy),
		)
	}
	c.JSON(http.StatusOK, gin.H{"message": "instance added"})
}

func (h *ProxyHandler) RemoveInstanceHandler(c *gin.Context) {
	serviceName := c.Param("service")
	instanceURL := c.Query("instance")

	h.mux.Lock()
	defer h.mux.Unlock()

	config, exists := h.services[serviceName]
	if !exists {
		h.log.ErrorContext(c, "service not found",
			logger.String("service", serviceName),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	for idx, instance := range config.Instances {
		if instance.URL == instanceURL {
			config.Instances = append(config.Instances[:idx], config.Instances[idx+1:]...)
			h.log.InfoContext(c, "remove instance success",
				logger.String("service", serviceName),
				logger.String("instance", instanceURL),
			)
			c.JSON(http.StatusOK, gin.H{"message": "instance removed"})
			return
		}
	}
	h.log.ErrorContext(c, "instance not found",
		logger.String("service", serviceName),
		logger.String("instance", instanceURL),
	)
	c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
}
