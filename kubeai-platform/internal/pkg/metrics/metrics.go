package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// 全局指标定义
var (
	// API网关指标
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubeai_api_request_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "path", "status", "service"},
	)
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubeai_api_request_duration_seconds",
			Help:    "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "service"},
	)

	// 模型管理指标
	ModelTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "kubeai_model_total",
			Help: "Total number of models",
		},
	)
	ModelVersionTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubeai_model_version_total",
			Help: "Total number of model versions by status",
		},
		[]string{"status"},
	)
	ModelHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubeai_model_health_status",
			Help: "Model health status (1=healthy, 0=unhealthy)",
		},
		[]string{"model_name", "version"},
	)

	// 推理服务指标
	InferenceTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubeai_inference_total",
			Help: "Total number of inference requests",
		},
		[]string{"model_name", "version", "status"},
	)
	InferenceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kubeai_inference_duration_seconds",
			Help:    "Inference execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 10, 10),
		},
		[]string{"model_name", "version"},
	)
	InferenceReplicas = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kubeai_inference_replicas",
			Help: "Number of inference replicas per model",
		},
		[]string{"model_name", "version"},
	)

	// 训练任务指标
	TrainingJobTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubeai_training_job_total",
			Help: "Total number of training jobs by status",
		},
		[]string{"status", "framework"},
	)
)

// Handler 暴露Prometheus指标接口
func Handler() http.Handler {
	return promhttp.Handler()
}
