package resources

import (
	aiv1 "inference-gateway/api/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewHPA creates the HorizontalPodAutoscaler targeting the Stable Deployment
func NewHPA(isvc *aiv1.InferenceService) *autoscalingv2.HorizontalPodAutoscaler {
	if isvc.Spec.Autoscaling == nil {
		return nil
	}

	metrics := []autoscalingv2.MetricSpec{}

	if isvc.Spec.Autoscaling.TargetCPUUtilization != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: isvc.Spec.Autoscaling.TargetCPUUtilization,
				},
			},
		})
	}

	if isvc.Spec.Autoscaling.TargetMemoryUtilization != nil {
		metrics = append(metrics, autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: corev1.ResourceMemory,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: isvc.Spec.Autoscaling.TargetMemoryUtilization,
				},
			},
		})
	}
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isvc.Name,
			Namespace: isvc.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.SchemeGroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       isvc.Name + "-stable",
			},
			MinReplicas: isvc.Spec.Autoscaling.MinReplicas,
			MaxReplicas: isvc.Spec.Autoscaling.MaxReplicas,
			Metrics:     metrics,
		},
	}
}
