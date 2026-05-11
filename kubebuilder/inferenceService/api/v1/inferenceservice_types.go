/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceServiceSpec defines the desired state of InferenceService
type InferenceServiceSpec struct {
	// 模型信息
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ModelName string `json:"modelName"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ModelVersion string `json:"modelVersion"`

	// 镜像信息(如果模型是自定义镜像,否则用默认推理镜像+模型挂载)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*(:[a-zA-Z0-9._-]+)?|([a-zA-Z0-9-]+\/)?[a-zA-Z0-9._-]+(:[a-zA-Z0-9._-]+)?|@sha256:[0-9a-f]{64})$`
	Image string `json:"image,omitempty"`

	// 容器端口 (默认 8501)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=8501
	Port int32 `json:"port"`

	// 副本数（可被HPA覆盖）
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// 节点名称（可被HPA覆盖）
	// +kubebuilder:validation:Optional
	NodeName string `json:"nodeName,omitempty"`

	// 资源限制
	// +kubebuilder:validation:Required
	Resources corev1.ResourceRequirements `json:"resources"`

	// 自动扩缩容配置
	// +kubebuilder:validation:Optional
	Autoscaling *AutoscalingSpec `json:"autoscaling,omitempty"`

	// 灰度发布配置
	// +kubebuilder:validation:Optional
	Canary *CanarySpec `json:"canary,omitempty"`

	// 服务暴露
	// +kubebuilder:validation:Required
	Service *ServiceSpec `json:"service"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^\/?[a-zA-Z0-9_\-\/]+$`
	ModelPath string `json:"modelPath,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	ActiveDeadline int64 `json:"activeDeadline,omitempty"`
}

// AutoscalingSpec 自动扩缩容配置
type AutoscalingSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	MinReplicas *int32 `json:"minReplicas,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// 目标 CPU 使用率 (0-100)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	TargetCPUUtilization *int32 `json:"targetCPUUtilization"`

	// 目标内存使用率 (0-100)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	TargetMemoryUtilization *int32 `json:"targetMemoryUtilization"`
}

// CanarySpec 灰度发布配置
type CanarySpec struct {
	// +kubebuilder:validation:Required
	Enabled bool `json:"enabled"`

	// 灰度版本号
	// +kubebuilder:validation:RequiredIf=Enabled=true
	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`

	// 流量比例 0-100
	// +kubebuilder:validation:RequiredIf=Enabled=true
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Weight int32 `json:"weight"`

	// 灰度模型名（通常与主模型相同）
	// +kubebuilder:validation:RequiredIf=Enabled=true
	// +kubebuilder:validation:MinLength=1
	ModelName string `json:"modelName"`

	// 灰度模型版本
	// +kubebuilder:validation:RequiredIf=Enabled=true
	// +kubebuilder:validation:MinLength=1
	ModelVersion string `json:"modelVersion"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*(:[a-zA-Z0-9._-]+)?|([a-zA-Z0-9-]+\/)?[a-zA-Z0-9._-]+(:[a-zA-Z0-9._-]+)?|@sha256:[0-9a-f]{64})$`
	Image string `json:"image,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources"`
}

// ServiceSpec 服务暴露配置
type ServiceSpec struct {
	// ClusterIp, NodePort, LoadBalancer
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	Type corev1.ServiceType `json:"type"`

	// Ingress 域名 (如果不填，则不创建 Ingress)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	Host string `json:"host,omitempty"`

	// 服务端口
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// 容器端口
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	TargetPort int32 `json:"targetPort"`

	// 可用于Ingress配置
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// InferenceServiceStatus defines the observed state of InferenceService.
type InferenceServiceStatus struct {
	// 稳定版本部署状态
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Running;Failed;Succeeded
	StableState string `json:"stableState,omitempty"`

	// 灰度版本部署状态(如有)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Running;Failed;Succeeded
	CanaryState string `json:"canaryState,omitempty"`

	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// 服务访问地址
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// 当前副本数
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// 服务是否就绪
	// +kubebuilder:validation:Optional
	Ready bool `json:"ready,omitempty"`

	// 最后更新时间
	// +kubebuilder:validation:Optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// +kubebuilder:validation:Optional
	RegisteredModel string `json:"registeredModel,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Model",type="string",JSONPath=".spec.modelName"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.modelVersion"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:resource:shortName=isvc
// +kubebuilder:validation:XValidation:rule="!has(self.spec.autoscaling) || (self.spec.autoscaling.minReplicas == nil || self.spec.autoscaling.maxReplicas >= self.spec.autoscaling.minReplicas)",message="maxReplicas must be greater than or equal to minReplicas in autoscaling"
// +kubebuilder:validation:XValidation:rule="!has(self.spec.canary) || !self.spec.canary.enabled || (self.spec.canary.weight >=0 && self.spec.canary.weight <=100)",message="canary weight must be between 0 and 100 when canary is enabled"
// InferenceService is the Schema for the inferenceservices API
type InferenceService struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of InferenceService
	// +kubebuilder:validation:Required
	Spec InferenceServiceSpec `json:"spec,omitempty"`

	// status defines the observed state of InferenceService
	// +optional
	Status InferenceServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// InferenceServiceList contains a list of InferenceService
type InferenceServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []InferenceService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceService{}, &InferenceServiceList{})
}
