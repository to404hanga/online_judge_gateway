package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	constants "github.com/to404hanga/onlinue_judge_gateway/constant"
	ojjwt "github.com/to404hanga/onlinue_judge_gateway/web/jwt"
)

type PathMethodPair struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

type JWTMiddlewareBuilder struct {
	ojjwt.Handler
	passPathMethodPairs []PathMethodPair
}

func NewJWTMiddlewareBuilder(handler ojjwt.Handler, passPathMethodPairs []PathMethodPair) *JWTMiddlewareBuilder {
	return &JWTMiddlewareBuilder{
		Handler:             handler,
		passPathMethodPairs: passPathMethodPairs,
	}
}

func (m *JWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		for _, p := range m.passPathMethodPairs {
			if path == p.Path && method == p.Method {
				return
			}
		}

		var uc ojjwt.UserClaims
		token, err := jwt.ParseWithClaims(m.ExtractToken(ctx), &uc, func(t *jwt.Token) (any, error) {
			return m.JwtKey(), nil
		})
		if err != nil || token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err = m.CheckSession(ctx, uc.Ssid); err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(constants.ContextUserClaimsKey, uc)
	}
}
