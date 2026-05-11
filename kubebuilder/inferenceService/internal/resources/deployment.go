package resources

import (
	aiv1 "inference-gateway/api/v1"
	"inference-gateway/pkg/modelmanager"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelKeyApp     = "app.kubernetes.io/name"
	LabelKeyVersion = "app.kubernetes.io/version"
	LabelKeyRole    = "app.kubernetes.io/component"
	LabelManagedBy  = "app.kubernetes.io/managed-by"
)

// NewStableDeployment 创建稳定版 Deployment
func NewStableDeployment(isvc *aiv1.InferenceService, meta *modelmanager.ModelMetadata) *appsv1.Deployment {
	labels := map[string]string{
		LabelKeyApp:     isvc.Name,
		LabelKeyVersion: isvc.Spec.ModelVersion,
		LabelKeyRole:    "stable",
		LabelManagedBy:  "kubeai-platform",
	}
	// 构建环境变量列表
	var envVars []corev1.EnvVar

	// 如果获取到了模型元数据，注入环境变量
	if meta != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MODEL_STORAGE_PATH",
			Value: meta.StoragePath,
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isvc.Name + "-stable",
			Namespace: isvc.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: getReplicas(isvc),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "inference-server",
							Image: getImage(isvc),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: getContainerPort(isvc),
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								}, // 默认推理端口
							},
							Resources: isvc.Spec.Resources,
							Env:       envVars, // 【关键】注入环境变量
							// LivenessProbe & ReadinessProbe 建议加上，这里简化省略
						},
					},
				},
			},
		},
	}
}

// NewCanaryDeployment 创建灰度版 Deployment
func NewCanaryDeployment(isvc *aiv1.InferenceService, meta *modelmanager.ModelMetadata) *appsv1.Deployment {
	if isvc.Spec.Canary == nil {
		return nil
	}

	labels := map[string]string{
		LabelKeyApp:     isvc.Name,
		LabelKeyVersion: isvc.Spec.Canary.Version,
		LabelKeyRole:    "canary",
		LabelManagedBy:  "kubeai-platform",
	}

	replicas := int32(1) // 灰度默认副本数通常较少，或者直接继承 Stable 的逻辑，这里简化为 1

	// 构建环境变量列表
	var envVars []corev1.EnvVar

	// 如果获取到了模型元数据，注入环境变量
	if meta != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "MODEL_STORAGE_PATH",
			Value: meta.StoragePath,
		})
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isvc.Name + "-canary",
			Namespace: isvc.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "inference-server",
							Image: getCanaryImage(isvc),
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: getContainerPort(isvc),
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: func() corev1.ResourceRequirements {
								if isvc.Spec.Canary.Resources != nil {
									return *isvc.Spec.Canary.Resources
								}
								return isvc.Spec.Resources
							}(),
							Env: envVars, // 【关键】注入环境变量
						},
					},
				},
			},
		},
	}
}
