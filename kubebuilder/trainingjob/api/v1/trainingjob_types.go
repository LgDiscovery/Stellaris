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

// TrainingJobSpec defines the desired state of TrainingJob
type TrainingJobSpec struct {
	// 深度学习框架 pytorch/tensorflow/onnx
	// +kubebuilder:validation:Enum=pytorch;tensorflow;onnx
	// +kubebuilder:default=pytorch
	// +kubebuilder:validation:Required // 框架为核心配置，强制必填
	Framework string `json:"framework"`

	// 容器镜像
	// +kubebuilder:validation:MinLength=1 // 镜像地址不能为空
	// +kubebuilder:validation:Required    // 镜像为核心配置，强制必填
	Image string `json:"image"`

	// 容器启动参数
	// +kubebuilder:validation:Optional
	Args []string `json:"args,omitempty"`

	// 容器启动命令
	// +kubebuilder:validation:MinLength=1 // 命令至少有一个元素
	// +kubebuilder:validation:Required    // 启动命令为核心配置，强制必填
	Command []string `json:"command"`

	// 分布式训练核心配置
	// +kubebuilder:validation:Optional
	Distributed bool `json:"distributed,omitempty"` // 是否开启分布式

	// 工作节点数
	// +kubebuilder:validation:XValidation:rule="self.spec.distributed == true ? self.spec.workerNum >= 1 : true",message="workerNum ≥1 when distributed=true"
	// +kubebuilder:validation:Minimum=0 // 节点数不能为负数
	// +kubebuilder:validation:Maximum=100 // 限制最大节点数，避免资源滥用
	// +kubebuilder:default=1
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self.spec.distributed == false ? self.spec.workerNum == 1 : true",message="workerNum=1 when distributed=false"
	WorkerNum int32 `json:"workerNum,omitempty"`

	// Master 节点数
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10 // Master节点数通常不超过10
	// +kubebuilder:default=0
	// +kubebuilder:validation:Optional
	MasterNum int32 `json:"masterNum,omitempty"`

	// 计算资源
	// +kubebuilder:validation:Required // 资源配置为核心配置，强制必填
	Resources ResourceRequirements `json:"resources"`

	// 环境变量
	// +kubebuilder:validation:Optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// 挂载卷
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// 数据集路径
	// +kubebuilder:validation:Optional
	DatasetPath string `json:"datasetPath,omitempty"`

	// 训练输出路径
	// +kubebuilder:validation:XValidation:rule="self.spec.outputPath =~ '^/[^\\0]+$'",message="outputPath must be absolute path"
	// +kubebuilder:validation:MinLength=1 // 输出路径不能为空
	// +kubebuilder:validation:Required    // 输出路径为核心配置，强制必填
	OutputPath string `json:"outputPath"`

	// 失败重试次数
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=0   // 重试次数不能为负数
	// +kubebuilder:validation:Maximum=10  // 限制最大重试次数
	// +kubebuilder:validation:Optional
	BackoffLimit int32 `json:"backoffLimit,omitempty"`

	// 完成后自动清理时间（秒）
	// +kubebuilder:default=3600
	// +kubebuilder:validation:Minimum=60  // 最小清理时间60秒
	// +kubebuilder:validation:Maximum=86400 // 最大清理时间1天
	// +kubebuilder:validation:Optional
	TTLSecondsAfterFinished int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// 任务最大运行时间（秒），超时自动终止
	// +kubebuilder:default=86400
	// +kubebuilder:validation:Minimum=60  // 最小运行时间60秒
	// +kubebuilder:validation:Maximum=604800 // 最大运行时间7天
	// +kubebuilder:validation:Optional
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds,omitempty"`

	// 关联模型ID (对接业务)
	// +kubebuilder:validation:Optional
	ModelID string `json:"modelID,omitempty"`
	// +kubebuilder:validation:Optional
	ModelName string `json:"modelName,omitempty"`

	// 监控/日志开关
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	EnableMonitor bool `json:"enableMonitor,omitempty"`
	// +kubebuilder:default=true
	// +kubebuilder:validation:Optional
	EnableLogs bool `json:"enableLogging,omitempty"`

	// 调度节点名称（由调度中心填充）
	// +kubebuilder:validation:Optional
	NodeName string `json:"nodeName,omitempty"`

	// 节点选择器
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// 容忍器
	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// ResourceRequirements 资源定义（CPU/Memory/GPU）
type ResourceRequirements struct {
	// CPU资源，格式支持纯数字（核心数）或带m（毫核），例如：1、500m
	// +kubebuilder:validation:Pattern=`^[0-9]+m?$`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	CPU string `json:"cpu"`

	// 内存资源，格式支持数字+单位（Ki/Mi/Gi），例如：1Gi、512Mi
	// +kubebuilder:validation:Pattern=`^[0-9]+[KMG]i$` // 强制单位后缀（K8s 标准）
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Memory string `json:"memory"`

	// GPU卡数，格式为非负整数（如1、2），对应nvidia.com/gpu资源
	// +kubebuilder:validation:Pattern=`^[0-9]+$`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	GPU string `json:"gpu,omitempty"`
}

// TrainingJobStatus defines the observed state of TrainingJob.
type TrainingJobStatus struct {
	// Phase is the current phase of the training job (Pending/Running/Succeeded/Failed)
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Cancelled;Unknown
	// +kubebuilder:validation:Optional
	Phase string `json:"phase,omitempty"`

	// Reason is a brief reason for the current phase
	// +kubebuilder:validation:Optional
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable message explaining the current phase
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`

	// LastTransitionTime is the last time the status transitioned
	// +kubebuilder:validation:Optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Conditions represents the latest available observations of the job's current state
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// 训练完成后注册的模型ID
	// +kubebuilder:validation:Optional
	RegisteredModelID string `json:"registeredModelID,omitempty"`

	// 模型注册重试次数
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=5
	// +kubebuilder:validation:Optional
	ModelRegisterRetryCount int32 `json:"modelRegisterRetryCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Model",type="string",JSONPath=".spec.modelName"
// +kubebuilder:printcolumn:name="Framework",type="string",JSONPath=".spec.framework"
// +kubebuilder:printcolumn:name="Distributed",type="boolean",JSONPath=".spec.distributed"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=tj
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.spec.framework) || self.spec.framework == oldSelf.spec.framework",message="framework field is immutable" // 框架字段不可修改
// TrainingJob is the Schema for the trainingjobs API
type TrainingJob struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of TrainingJob
	// +required
	Spec TrainingJobSpec `json:"spec"`

	// status defines the observed state of TrainingJob
	// +optional
	Status TrainingJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TrainingJobList contains a list of TrainingJob
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrainingJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrainingJob{}, &TrainingJobList{})
}
