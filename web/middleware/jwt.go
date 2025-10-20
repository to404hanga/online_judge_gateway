package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	ojmodel "github.com/to404hanga/online_judge_common/model"
	constants "github.com/to404hanga/online_judge_gateway/constant"
	ojjwt "github.com/to404hanga/online_judge_gateway/web/jwt"
	"github.com/to404hanga/pkg404/cachex/lru"
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
	cache               *lru.Cache
}

func NewJWTMiddlewareBuilder(handler ojjwt.Handler, db *gorm.DB, cache *lru.Cache, loginCheckPassPairs, adminCheckPairs []PathMethodPair, log loggerv2.Logger) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{
		Handler:             handler,
		db:                  db,
		cache:               cache,
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
				m.log.ErrorContext(ctx, "CheckAdmin GetUserClaims failed", logger.Error(err))
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}
			cacheKey := fmt.Sprintf(constants.CacheUserKey, uc.UserId)
			val, ok := m.cache.Get(cacheKey)
			if ok {
				user, ok := val.(constants.CacheUser)
				if ok {
					if user.Role != int8(ojmodel.UserRoleAdmin) {
						m.log.ErrorContext(ctx, "CheckAdmin failed", logger.Int8("actual_role", user.Role))
						ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
							"error": "权限不足",
						})
						return
					}
				} else {
					m.log.ErrorContext(ctx, "CheckAdmin assert failed", logger.Any("value", val))
				}
			}
			var user ojmodel.User
			if err = m.db.WithContext(ctx).Where("id = ?", uc.UserId).Select("username", "realname", "realname", "role").First(&user).Error; err != nil {
				m.log.ErrorContext(ctx, "CheckAdmin get db failed", logger.Error(err))
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			m.cache.Add(cacheKey, constants.CacheUser{
				Username: user.Username,
				Realname: user.Realname,
				Role:     user.Role.Int8(),
			})
		}

		ctx.Next()
	}
}
