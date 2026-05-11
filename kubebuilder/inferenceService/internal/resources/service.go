package resources

import (
	aiv1 "inference-gateway/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewStableService creates the main service
func NewStableService(isvc *aiv1.InferenceService) *corev1.Service {
	svcType := corev1.ServiceTypeClusterIP
	if isvc.Spec.Service != nil && isvc.Spec.Service.Type != "" {
		svcType = isvc.Spec.Service.Type
	}

	labels := map[string]string{
		LabelKeyApp:    isvc.Name,
		LabelKeyRole:   "stable",
		LabelManagedBy: "kubeai-platform",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isvc.Name + "-stable",
			Namespace: isvc.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: svcType,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       getServicePort(isvc),
					TargetPort: intstr.FromInt(int(getContainerPort(isvc))),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				LabelKeyApp:  isvc.Name,
				LabelKeyRole: "stable",
			},
		},
	}
}

// NewCanaryService creates the canary service
func NewCanaryService(isvc *aiv1.InferenceService) *corev1.Service {
	if isvc.Spec.Canary == nil {
		return nil
	}
	svcType := corev1.ServiceTypeClusterIP
	if isvc.Spec.Service != nil && isvc.Spec.Service.Type != "" {
		svcType = isvc.Spec.Service.Type
	}
	labels := map[string]string{
		LabelKeyApp:    isvc.Name,
		LabelKeyRole:   "canary",
		LabelManagedBy: "kubeai-platform",
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isvc.Name + "-canary",
			Namespace: isvc.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: svcType,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       getServicePort(isvc),
					TargetPort: intstr.FromInt(int(getContainerPort(isvc))),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				LabelKeyApp:  isvc.Name,
				LabelKeyRole: "canary",
			},
		},
	}
}
