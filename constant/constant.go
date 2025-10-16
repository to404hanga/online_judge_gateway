package constants

const GatewayServiceName = "OnlineJudge-Gateway"

const ProxyKey = "cmd" // 代理时需要转发的路径的查询参数键

const (
	HeaderForwardedByKey  = "X-Forwarded-By"
	HeaderUserID          = "X-User-ID"
	HeaderRequestIDKey    = "X-Request-ID"
	HeaderProxyByKey      = "X-Proxy-By"
	HeaderLoginTokenKey   = "X-JWT-Token"
	HeaderRefreshTokenKey = "X-Refresh-Token"
)

const (
	ContextUserClaimsKey = "X-User-Claims"
)
