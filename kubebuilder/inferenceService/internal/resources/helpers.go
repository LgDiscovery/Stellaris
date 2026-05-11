package resources

import aiv1 "inference-gateway/api/v1"

// 默认配置常量
const (
	DefaultInferenceImage = "tensorflow/serving:latest"
	DefaultContainerPort  = 8501
	DefaultServicePort    = 80
)

// getImage 获取稳定版镜像
func getImage(isvc *aiv1.InferenceService) string {
	if isvc.Spec.Image != "" {
		return isvc.Spec.Image
	}
	return DefaultInferenceImage
}

// getCanaryImage 获取灰度版镜像
func getCanaryImage(isvc *aiv1.InferenceService) string {
	if isvc.Spec.Canary != nil && isvc.Spec.Canary.Image != "" {
		return isvc.Spec.Canary.Image
	}
	// 如果灰度没指定镜像，尝试用主配置镜像加版本逻辑，这里简化直接用默认或主镜像
	return getImage(isvc)
}

// getContainerPort 获取容器端口
func getContainerPort(isvc *aiv1.InferenceService) int32 {
	if isvc.Spec.Port != 0 {
		return isvc.Spec.Port
	}
	return DefaultContainerPort
}

// getServicePort 获取服务端口
func getServicePort(isvc *aiv1.InferenceService) int32 {
	if isvc.Spec.Service != nil && isvc.Spec.Service.Port != 0 {
		return isvc.Spec.Service.Port
	}
	return DefaultServicePort
}

// getReplicas 获取副本数指针
func getReplicas(isvc *aiv1.InferenceService) *int32 {
	if isvc.Spec.Replicas != nil {
		return isvc.Spec.Replicas
	}
	// 默认返回 1 的指针
	i := int32(1)
	return &i
}
