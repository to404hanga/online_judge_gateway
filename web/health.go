package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

var _ Handler = (*HealthHandler)(nil)

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}
