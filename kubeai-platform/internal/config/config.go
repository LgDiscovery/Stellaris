package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	RedisConf redis.RedisConf `mapstructure:"redisConf" json:"redisConf"`
	RateLimit RateLimit       `mapstructure:"rateLimit" json:"rateLimit"`
	Auth      Auth            `mapstructure:"auth" json:"auth"`
	Upstreams Upstreams       `mapstructure:"upstreams" json:"upstreams"`
}

type Auth struct {
	AccessSecret string `mapstructure:"accessSecret" json:"accessSecret"`
}

type Upstreams struct {
	ModelManager         Upstream `mapstructure:"modelManager" json:"modelManager"`
	JobScheduler         Upstream `mapstructure:"jobScheduler" json:"jobScheduler"`
	InferenceGateway     Upstream `mapstructure:"inferenceGateway" json:"inferenceGateway"`
	ObservabilityGateway Upstream `mapstructure:"observabilityGateway" json:"observabilityGateway"`
}

type Upstream struct {
	URL string `mapstructure:"url" json:"url"`
}

type RateLimit struct {
	Enabled   bool   `mapstructure:"enabled" json:"enabled" json:",default=true"`                          // 是否开启限流，默认开启
	Rate      int    `mapstructure:"rate" json:"rate" json:",default=100"`                                 // 时间窗口内最大请求数
	Interval  int    `mapstructure:"interval" json:"interval" json:",default=60s"`                         // 时间窗口，支持 60s/1m/1h 等格式
	KeyPrefix string `mapstructure:"keyPrefix" json:"keyPrefix" json:",default=kubeai:gateway:ratelimit:"` // Redis key前缀
}
