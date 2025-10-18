package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	constants "github.com/to404hanga/online_judge_gateway/constant"
)

var ssidKey = "users:ssid:%s"

type RedisJWTHandler struct {
	client            redis.Cmdable
	signingMethod     jwt.SigningMethod
	jwtExpiration     time.Duration
	refreshExpiration time.Duration
	jwtKey            []byte
	refreshKey        []byte
}

func NewRedisJWTHandler(client redis.Cmdable, jwtKey []byte, refreshKey []byte, jwtExpiration, refreshExpiration time.Duration) Handler {
	return &RedisJWTHandler{
		client:            client,
		signingMethod:     jwt.SigningMethodHS512,
		jwtExpiration:     jwtExpiration,
		refreshExpiration: refreshExpiration,
		jwtKey:            jwtKey,
		refreshKey:        refreshKey,
	}
}

var _ Handler = &RedisJWTHandler{}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := h.client.Exists(ctx, fmt.Sprintf(ssidKey, ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token invalid")
	}
	return nil
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header(constants.HeaderLoginTokenKey, "")
	ctx.Header(constants.HeaderLoginTokenKey, "")
	uc := ctx.MustGet(constants.ContextUserClaimsKey).(UserClaims)
	return h.client.Set(ctx, fmt.Sprintf(ssidKey, uc.Ssid), "", h.refreshExpiration).Err()
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, UserId uint64) error {
	ssid := uuid.New().String()
	if err := h.SetRefreshToken(ctx, UserId, ssid); err != nil {
		return err
	}
	return h.SetJWTToken(ctx, UserId, ssid)
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	// 优先从Authorization Header提取token
	authCode := ctx.GetHeader("Authorization")
	if authCode != "" {
		segs := strings.Split(authCode, " ")
		if len(segs) == 2 && segs[0] == "Bearer" {
			return segs[1]
		}
	}

	// 如果Header中没有，尝试从Cookie中提取
	tokenFromCookie, err := ctx.Cookie(constants.HeaderLoginTokenKey)
	if err != nil || tokenFromCookie == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return ""
	}

	return tokenFromCookie
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, UserId uint64, ssid string) error {
	uc := UserClaims{
		UserId:    UserId,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.jwtExpiration)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(h.jwtKey)
	if err != nil {
		return err
	}

	// 设置响应头
	ctx.Header(constants.HeaderLoginTokenKey, tokenStr)

	// 同时设置Cookie，支持浏览器自动携带
	ctx.SetCookie(
		constants.HeaderLoginTokenKey,  // cookie名称
		tokenStr,                       // cookie 值
		int(h.jwtExpiration.Seconds()), // 过期时间（秒）
		"/",                            // 路径
		"",                             // 域名
		false,                          // secure (HTTPS)
		true,                           // httpOnly
	)

	return nil
}

func (h *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, UserId uint64, ssid string) error {
	rc := RefreshClaims{
		UserId: UserId,
		Ssid:   ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.refreshExpiration)), // 7 天过期
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, rc)
	tokenStr, err := token.SignedString(h.refreshKey)
	if err != nil {
		return err
	}
	ctx.Header(constants.HeaderLoginTokenKey, tokenStr)
	return nil
}

func (h *RedisJWTHandler) JwtKey() []byte {
	return h.jwtKey
}

func (h *RedisJWTHandler) RefreshKey() []byte {
	return h.refreshKey
}

func (h *RedisJWTHandler) GetUserClaims(ctx *gin.Context) (*UserClaims, error) {
	ucAny, exists := ctx.Get(constants.ContextUserClaimsKey)
	if !exists {
		return nil, fmt.Errorf("user claims not found in context")
	}
	uc, ok := ucAny.(UserClaims)
	if !ok {
		return nil, fmt.Errorf("user claims type assertion error")
	}
	return &uc, nil
}
