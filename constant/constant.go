package constants

const GatewayServiceName = "OnlineJudge-Gateway"

const ProxyKey = "cmd" // 代理时需要转发的路径的查询参数键

const (
	HeaderForwardedByKey = "X-Forwarded-By"
	HeaderUserIDKey      = "X-User-ID"
	HeaderRequestIDKey   = "X-Request-ID"
	HeaderProxyByKey     = "X-Proxy-By"
	HeaderLoginTokenKey  = "X-JWT-Token"
)

const (
	ContextUserClaimsKey = "X-User-Claims"
)

const (
	CacheUserKey = "user:%d" // args: user.ID
)

type CacheUser struct {
	Username string
	Realname string
	Role     int8
}
