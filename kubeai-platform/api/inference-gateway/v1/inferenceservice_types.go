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
	ModelName    string `json:"modelName"`
	ModelVersion string `json:"modelVersion"`

	// 镜像信息(如果模型是自定义镜像,否则用默认推理镜像+模型挂载)
	Image string `json:"image,omitempty"`
	// 容器端口 (默认 8501)
	Port int32 `json:"port,omitempty"`
	// 副本数（可被HPA覆盖）
	Replicas *int32 `json:"replicas,omitempty"`

	//资源限制
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// 自动扩缩容配置
	Autoscaling *AutoscalingSpec `json:"autoscaling,omitempty"`

	// 灰度发布配置
	Canary *CanarySpec `json:"canary,omitempty"`

	// 服务暴露
	Service *ServiceSpec `json:"service,omitempty"`
}

type AutoscalingSpec struct {
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	MaxReplicas int32  `json:"maxReplicas"`
	// 目标 CPU 使用率 (0-100)
	TargetCPUUtilization *int32 `json:"targetCPUUtilization,omitempty"`
	// 目标内存使用率 (0-100)
	TargetMemoryUtilization *int32 `json:"targetMemoryUtilization,omitempty"`
}

type CanarySpec struct {
	Enabled      bool                         `json:"enabled"`
	Version      string                       `json:"version"`      // 灰度版本号
	Weight       int32                        `json:"weight"`       // 流量比例 0-100
	ModelName    string                       `json:"modelName"`    // 灰度模型名（通常与主模型相同）
	ModelVersion string                       `json:"modelVersion"` // 灰度模型版本
	Image        string                       `json:"image,omitempty"`
	Resources    *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ServiceSpec 服务暴露配置
type ServiceSpec struct {
	// ClusterIp, NodePort, LoadBalancer
	Type corev1.ServiceType `json:"type,omitempty"`
	// Ingress 域名 (如果不填，则不创建 Ingress)
	Host        string            `json:"host,omitempty"`
	Port        int32             `json:"port"`                  // 服务端口
	TargetPort  int32             `json:"targetPort"`            //容器端口
	Annotations map[string]string `json:"annotations,omitempty"` // 可用于Ingress配置
}

// InferenceServiceStatus defines the observed state of InferenceService.
type InferenceServiceStatus struct {
	// 稳定版本部署状态
	StableState string `json:"stableState,omitempty"`
	//灰度版本部署状态(如有)
	CanaryState   string             `json:"canaryState,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
	URL           string             `json:"url,omitempty"`           // 服务访问地址
	ReadyReplicas int32              `json:"readyReplicas,omitempty"` // 当前副本数
	Ready         bool               `json:"ready"`                   // 服务是否就绪
	//最后更新时间
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=inferenceservices,scope=Namespaced
// +kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.modelName`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.modelVersion`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="ReadyReplicas",type=string,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="StableState",type=string,JSONPath=`.status.stableState`
// +kubebuilder:printcolumn:name="CanaryState",type=string,JSONPath="`.status.canaryState`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// InferenceService is the Schema for the inferenceservices API
type InferenceService struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of InferenceService
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
