package resources

import (
	"fmt"
	aiv1 "inference-gateway/api/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewIngress 为了实现 Nginx Ingress 的灰度，我们需要创建 两个 Ingress 资源：
// 主 Ingress：处理大部分流量，指向 Stable Service。
// 副 Ingress：带有 canary 注解，指向 Canary Service。
func NewIngress(isvc *aiv1.InferenceService) *networkingv1.Ingress {
	if isvc.Spec.Service == nil || isvc.Spec.Service.Host == "" {
		return nil // 没有配置 Host 则不创建
	}

	annotations := map[string]string{
		"kubernetes.io/ingress.class": "nginx",
	}

	// 合并用户自定义注解
	if isvc.Spec.Service.Annotations != nil {
		for k, v := range isvc.Spec.Service.Annotations {
			annotations[k] = v
		}
	}

	pathType := networkingv1.PathTypePrefix

	// 构建基础 Ingress
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            isvc.Name,
			Namespace:       isvc.Namespace,
			Annotations:     annotations,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService"))},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: isvc.Spec.Service.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: isvc.Name + "-stable",
											Port: networkingv1.ServiceBackendPort{Number: getServicePort(isvc)},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ing
}

// NewCanaryIngress creates the Canary Ingress resource
func NewCanaryIngress(isvc *aiv1.InferenceService) *networkingv1.Ingress {
	if isvc.Spec.Service == nil || isvc.Spec.Service.Host == "" {
		return nil
	}
	if isvc.Spec.Canary == nil || !isvc.Spec.Canary.Enabled {
		return nil
	}
	annotations := map[string]string{
		"kubernetes.io/ingress.class":               "nginx",
		"nginx.ingress.kubernetes.io/canary":        "true",
		"nginx.ingress.kubernetes.io/canary-weight": fmt.Sprintf("%d", isvc.Spec.Canary.Weight),
	}

	pathType := networkingv1.PathTypePrefix
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        isvc.Name,
			Namespace:   isvc.Namespace,
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService")),
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: isvc.Spec.Service.Host, // 必须和主 Ingress 一致
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: isvc.Name + "-canary",
											Port: networkingv1.ServiceBackendPort{
												Number: getServicePort(isvc),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ing
}
