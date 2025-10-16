package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/to404hanga/onlinue_judge_gateway/domain"
	"github.com/to404hanga/onlinue_judge_gateway/service"
	ojjwt "github.com/to404hanga/onlinue_judge_gateway/web/jwt"
	"github.com/to404hanga/onlinue_judge_gateway/web/middleware"
	"github.com/to404hanga/pkg404/logger"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
)

type AuthHandler struct {
	authService service.AuthService
	jwtHandler  ojjwt.Handler
	log         loggerv2.Logger
}

var _ Handler = (*AuthHandler)(nil)

func NewAuthHandler(authService service.AuthService, jwtHandler ojjwt.Handler, log loggerv2.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtHandler:  jwtHandler,
		log:         log,
	}
}

func (h *AuthHandler) Register(r *gin.Engine) {
	auth := r.Group("/auth").Use(middleware.Logger(h.log))
	{
		auth.POST("/login", h.LoginHandler)
		auth.POST("/logout", h.LogoutHandler)
		auth.GET("/info", h.InfoHandler)
	}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.ErrorContext(c, "loginHandler bind json failed", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := loggerv2.ContextWithFields(c, logger.String("username", req.Username))

	userID, err := h.authService.Login(ctx, &req)
	if err != nil {
		h.log.ErrorContext(ctx, "loginHandler login failed", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.jwtHandler.SetLoginToken(c, userID)

	c.JSON(http.StatusOK, gin.H{"message": "login success"})
}

func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	if err := h.jwtHandler.ClearToken(c); err != nil {
		h.log.ErrorContext(c, "logoutHandler clear token failed", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

func (h *AuthHandler) InfoHandler(c *gin.Context) {
	uc, err := h.jwtHandler.GetUserClaims(c)
	if err != nil {
		h.log.ErrorContext(c, "infoHandler get user claims failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx := loggerv2.ContextWithFields(c, logger.Uint64("user_id", uc.UserId))

	resp, err := h.authService.Info(ctx, uc.UserId)
	if err != nil {
		h.log.ErrorContext(ctx, "infoHandler get user info failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
