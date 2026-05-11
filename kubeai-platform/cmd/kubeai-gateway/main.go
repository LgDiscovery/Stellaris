package main

import (
	"flag"
	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"kubeai/internal/config"
	"kubeai/internal/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 初始化 go-zero HTTP 服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 全局中间件：限流、CORS
	server.Use(middleware.RateLimitMiddleware(c.RateLimit, c.RedisConf))
	server.Use(middleware.CorsMiddleware())

	// 1. 公开接口：登录
	server.AddRoute(rest.Route{
		Method:  http.MethodPost,
		Path:    "/api/v1/auth/login",
		Handler: loginHandler(c.Auth.AccessSecret),
	})

	// 2. 初始化鉴权中间件（给需要鉴权的路由单独使用）
	authMid := middleware.AuthMiddleware(c.Auth.AccessSecret)

	// 3. 模型管理路由转发
	modelProxy, err := newReverseProxy(c.Upstreams.ModelManager.URL)
	if err != nil {
		logx.Errorf("init model manager proxy failed: %v", err)
	}
	server.AddRoute(rest.Route{
		Method:  "*",
		Path:    "/api/v1/models/*",
		Handler: authMid(modelProxy.ServeHTTP),
	})

	// 4. 任务调度路由转发
	jobProxy, err := newReverseProxy(c.Upstreams.JobScheduler.URL)
	if err != nil {
		logx.Errorf("init jobscheduler scheduler proxy failed: %v", err)
	}
	server.AddRoute(rest.Route{
		Method:  "*",
		Path:    "/api/v1/jobs/*",
		Handler: authMid(jobProxy.ServeHTTP),
	})

	// 5. 推理服务路由转发
	inferenceProxy, err := newReverseProxy(c.Upstreams.InferenceGateway.URL)
	if err != nil {
		logx.Errorf("init inference gateway proxy failed: %v", err)
	}
	server.AddRoute(rest.Route{
		Method:  "*",
		Path:    "/api/v1/inference/*",
		Handler: authMid(inferenceProxy.ServeHTTP),
	})

	// 6. 可观测服务路由转发
	observabilityProxy, err := newReverseProxy(c.Upstreams.ObservabilityGateway.URL)
	if err != nil {
		logx.Errorf("init observability gateway proxy failed: %v", err)
	}
	server.AddRoute(rest.Route{
		Method:  "*",
		Path:    "/api/v1/observability/*",
		Handler: authMid(observabilityProxy.ServeHTTP),
	})

	logx.Infof("gateway started on port %d", c.Port)
	server.Start()
}

// loginHandler 登录逻辑处理
func loginHandler(accessSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		// 解析请求参数（go-zero 自带解析）
		if err := httpx.Parse(r, &loginReq); err != nil {
			logx.Errorf("parse login request failed: %v", err)
			httpx.WriteJson(w, http.StatusBadRequest, map[string]string{
				"error":   "invalid request",
				"message": "missing or invalid username/password",
			})
			return
		}

		// 简化的用户名密码验证（提示：生产环境需替换为加密验证+数据库查询）
		if loginReq.Username != "admin" || loginReq.Password != "password" {
			logx.Errorf("login failed: invalid credentials for user %s", loginReq.Username)
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]string{
				"error":   "invalid credentials",
				"message": "username or password is incorrect",
			})
			return
		}

		// 生成 JWT Token
		claims := middleware.JwtClaims{
			UserID: 1,
			Role:   "admin",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "kubeai-gateway", // 增加签发者，提升 Token 安全性
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(accessSecret))
		if err != nil {
			logx.Errorf("generate jwt token failed: %v", err)
			httpx.WriteJson(w, http.StatusInternalServerError, map[string]string{
				"error":   "failed to generate token",
				"message": "internal server error",
			})
			return
		}

		httpx.WriteJson(w, http.StatusOK, map[string]string{
			"token":  tokenString,
			"expire": time.Now().Add(24 * time.Hour).Format(time.RFC3339), // 返回过期时间，方便前端处理
		})
	}
}

// newReverseProxy 反向代理工具函数
func newReverseProxy(target string) (http.Handler, error) {
	// 处理 url.Parse 错误
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// 重写请求路径，去掉/api/v1前缀
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		// 传递原始请求头，避免上游服务丢失信息
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
	}

	// 增加反向代理错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logx.Errorf("reverse proxy error: %v, target: %s, path: %s", err, target, r.URL.Path)
		httpx.WriteJson(w, http.StatusBadGateway, map[string]string{
			"error":   "bad gateway",
			"message": "upstream service unavailable",
		})
	}

	return proxy, nil
}
