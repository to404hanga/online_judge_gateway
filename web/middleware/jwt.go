package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	ojmodel "github.com/to404hanga/online_judge_common/model"
	constants "github.com/to404hanga/online_judge_gateway/constant"
	ojjwt "github.com/to404hanga/online_judge_gateway/web/jwt"
	"github.com/to404hanga/pkg404/logger"
	loggerv2 "github.com/to404hanga/pkg404/logger/v2"
	"gorm.io/gorm"
)

type PathMethodPair struct {
	Path   string   `yaml:"path"`
	Method string   `yaml:"method"`
	Cmd    []string `yaml:"cmd"` // 仅当请求转发时鉴权管理员有效
}

type JWTMiddlewareBuilder struct {
	ojjwt.Handler
	db                  *gorm.DB
	loginCheckPassPairs []PathMethodPair
	adminCheckPairs     []PathMethodPair
	log                 loggerv2.Logger
}

func NewJWTMiddlewareBuilder(handler ojjwt.Handler, db *gorm.DB, loginCheckPassPairs, adminCheckPairs []PathMethodPair, log loggerv2.Logger) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{
		Handler:             handler,
		db:                  db,
		loginCheckPassPairs: loginCheckPassPairs,
		adminCheckPairs:     adminCheckPairs,
		log:                 log,
	}
}

// CheckLogin 检查登录状态
func (m *JWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		for _, p := range m.loginCheckPassPairs {
			if strings.HasPrefix(path, p.Path) && method == p.Method {
				return
			}
		}

		var uc ojjwt.UserClaims
		token, err := jwt.ParseWithClaims(m.ExtractToken(ctx), &uc, func(t *jwt.Token) (any, error) {
			return m.JwtKey(), nil
		})
		if err != nil || token == nil || !token.Valid {
			m.log.ErrorContext(ctx, "CheckLogin failed",
				logger.Error(err),
				logger.Bool("token==nil", token == nil),
			)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err = m.CheckSession(ctx, uc.Ssid); err != nil {
			m.log.ErrorContext(ctx, "CheckLogin failed", logger.Error(err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(constants.ContextUserClaimsKey, uc)
	}
}

// CheckAdmin 检查管理员权限
func (m *JWTMiddlewareBuilder) CheckAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		shouldCheck := false
		cmd := ctx.Query(constants.ProxyKey)
		for _, p := range m.adminCheckPairs {
			if path == p.Path && method == p.Method {
				shouldCheck = true
				break
			}
			// 只有 /api/*path 才会有 cmd, 但是 /api/*path 设置的 method 为 ANY
			for _, c := range p.Cmd {
				if c == cmd {
					shouldCheck = true
					break
				}
			}
		}

		if shouldCheck {
			uc, err := m.GetUserClaims(ctx)
			if err != nil {
				m.log.ErrorContext(ctx, "CheckAdmin failed", logger.Error(err))
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			var user ojmodel.User
			if err = m.db.Where("id = ?", uc.UserId).Select("role").First(&user).Error; err != nil {
				m.log.ErrorContext(ctx, "CheckAdmin failed", logger.Error(err))
				ctx.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if *user.Role != ojmodel.UserRoleAdmin {
				m.log.ErrorContext(ctx, "CheckAdmin failed", logger.Error(err))
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		ctx.Next()
	}
}
