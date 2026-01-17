package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yuhai94/anywhere_agent/internal/config"
	"github.com/yuhai94/anywhere_agent/internal/logger"
	"github.com/yuhai94/anywhere_agent/internal/v2ray"
	"go.uber.org/zap"
)

// APIServer API服务器
type APIServer struct {
	config     *config.Config
	address    string
	port       int
	jwtSecret  string
	v2rayStats *v2ray.TrafficMonitor
	deployChan chan *v2ray.DeployStatus
	server     *http.Server // 保存HTTP服务器实例
}

// NewAPIServer 创建新的API服务器
func NewAPIServer(cfg *config.Config, deployChan chan *v2ray.DeployStatus, v2rayStats *v2ray.TrafficMonitor) *APIServer {
	return &APIServer{
		config:     cfg,
		address:    cfg.API.Address,
		port:       cfg.API.Port,
		jwtSecret:  cfg.API.JWTSecret,
		v2rayStats: v2rayStats,
		deployChan: deployChan,
	}
}

// Start 启动API服务器
func (s *APIServer) Start() error {
	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode) // 生产模式
	r := gin.Default()

	// API路由组
	api := r.Group("/api")

	// 应用JWT认证中间件
	api.Use(s.jwtAuthMiddleware)

	// 简化后的API端点：同时返回状态和配置
	api.GET("/status", s.handleStatusAndConfig)

	// 健康检查端点（无需认证）
	r.GET("/health", s.handleHealth)

	// 启动HTTP服务器
	addr := fmt.Sprintf("%s:%d", s.address, s.port)
	logger.Info("API server starting",
		zap.String("address", addr),
		zap.String("protocol", "HTTP"))

	// 创建HTTP服务器实例
	s.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 使用ListenAndServe启动服务器
	return s.server.ListenAndServe()
}

// Stop 停止API服务器
func (s *APIServer) Stop() error {
	if s.server == nil {
		return nil
	}

	logger.Info("Stopping API server...")
	// 优雅关闭服务器，允许10秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// jwtAuthMiddleware JWT认证中间件
func (s *APIServer) jwtAuthMiddleware(c *gin.Context) {
	// 从Authorization头获取JWT令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}

	// 检查Authorization头格式
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
		c.Abort()
		return
	}

	tokenString := parts[1]

	// 验证JWT令牌
	_, err := ValidateJWT(tokenString, s.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		c.Abort()
		return
	}

	// 令牌有效，继续处理请求
	c.Next()
}

// handleStatusAndConfig 同时处理状态和配置查询请求
func (s *APIServer) handleStatusAndConfig(c *gin.Context) {
	// 获取V2Ray状态
	status, err := v2ray.GetV2RayStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get v2ray status: %v", err)})
		return
	}

	// 返回合并的响应
	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"config": map[string]interface{}{
			"port":       s.config.V2Ray.Port,
			"uuid":       s.config.V2Ray.UUID,
			"access_log": s.config.V2Ray.AccessLog,
		},
	})
}

// handleHealth 处理健康检查请求
func (s *APIServer) handleHealth(c *gin.Context) {
	// 返回健康状态
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "anywhere-agent",
	})
}
