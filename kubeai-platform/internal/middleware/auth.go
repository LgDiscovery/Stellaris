package middleware

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
	"kubeai/internal/config"
	"net/http"
	"strings"
)

// 定义自定义 Context Key 类型
type ctxKey string

const (
	CtxUserID ctxKey = "user_id"
	CtxRole   ctxKey = "role"
)

type JwtClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware(accessSecret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 提取Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				logx.Errorf("auth failed: empty/invalid header, ip: %s", r.RemoteAddr)
				httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			// 验证Token
			claims := &JwtClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				// 验证签名算法，防止算法伪造
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("invalid signing method")
				}
				return []byte(accessSecret), nil
			})

			if err != nil || !token.Valid {
				logx.Errorf("auth failed: invalid token, ip: %s, err: %v", r.RemoteAddr, err)
				httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
					"error": "invalid token",
				})
				return
			}

			// 写入上下文
			ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxRole, claims.Role)
			// 调用next，传入带上下文的request
			next(w, r.WithContext(ctx))
		}
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limitConfig config.RateLimit, redisConf redis.RedisConf) func(http.HandlerFunc) http.HandlerFunc {
	// 1. 未开启限流：返回透传中间件
	if !limitConfig.Enabled {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return next // 直接透传，不做任何处理
		}
	}

	// 2. 初始化Redis客户端 分布式限流
	redisClient, err := redis.NewRedis(redisConf)
	if err != nil {
		logx.Errorf("init redis client for rate limit failed: %v", err)
	}

	// 3. 初始化限流器（分布式限流）
	limiter := limit.NewPeriodLimit(
		int(limitConfig.Interval), // period：时间窗口（秒）
		limitConfig.Rate,          // quota：时间窗口内最大请求数
		redisClient,               // limitStore：Redis客户端
		limitConfig.KeyPrefix,     // keyPrefix：Redis key前缀
	)

	// 4. 返回中间件
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 以客户端IP作为限流key
			key := r.RemoteAddr
			code, err := limiter.Take(key)
			if err != nil {
				logx.Errorf("rate limit error: %v, ip: %s", err, key)
				httpx.WriteJson(w, http.StatusInternalServerError, map[string]string{
					"error": "rate limit internal error",
				})
				return
			}

			// 超过限流：返回429
			if code == limit.OverQuota {
				logx.Errorf("rate limit exceeded: ip=%s, path=%s", key, r.URL.Path)
				httpx.WriteJson(w, http.StatusTooManyRequests, map[string]string{
					"error": "too many requests",
				})
				return
			}

			// 限流通过：调用next HandlerFunc
			next(w, r)
		}
	}
}

// CorsMiddleware 跨域中间件
func CorsMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			// 生产环境建议替换为具体域名，如 "https://your-frontend.com"
			allowOrigin := r.Header.Get("Origin")
			if allowOrigin == "" {
				allowOrigin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-type,AccessToken,X-CSRF-Token,Authorization,Token,X-Token,X-User")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 预检请求缓存时间，减少OPTIONS请求

			if method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}
}
