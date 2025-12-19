package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler interface {
	ClearToken(ctx *gin.Context) error
	ExtractToken(ctx *gin.Context) string
	SetLoginToken(ctx *gin.Context, uid uint64) error
	SetJWTToken(ctx *gin.Context, uid uint64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	GetUserTokenVersion(ctx *gin.Context, uid uint64) (int64, error)

	JwtKey() []byte
	GetUserClaims(ctx *gin.Context) (*UserClaims, error)
}

type UserClaims struct {
	jwt.RegisteredClaims
	UserId       uint64
	Ssid         string
	UserAgent    string
	TokenVersion int64
}
