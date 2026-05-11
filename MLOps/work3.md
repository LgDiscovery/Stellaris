### 一、推理服务网关目标与功能

### 目标

* 提供**声明式**的模型推理服务部署能力，用户通过YAML或API即可将模型部署到K8s
* 支持**多版本管理**（如v1、v2同时运行）
* 支持**灰度发布**（基于流量比例或Header路由）
* 支持**自动弹性伸缩**（基于CPU/内存或自定义指标）
* 提供统一的**服务发现**和**路由**（如`http://inference-gateway/bert-classifier/predict`）

### 核心功能

1. **推理服务创建**：用户指定模型版本、副本数、资源、环境变量等，自动创建K8s Deployment、Service、Ingress
2. **版本升级/回滚**：通过修改CRD的`spec.canary`实现灰度发布，支持逐步切换流量
3. **弹性伸缩**：集成HorizontalPodAutoscaler（HPA）
4. **监控集成**：自动为推理服务添加Prometheus annotations，暴露`/metrics`端点
5. **生命周期管理**：暂停、恢复、删除推理服务

### 技术选型

* **Kubebuilder**：开发CRD和Controller
* **client-go**：操作K8s资源
* **Gin**：提供REST API（可选，也可以直接用CRD）
* **Istio / Ingress Nginx**：灰度流量切分（我们选择使用Service + 多Deployment + 自定义路由，或利用Istio，但为了简化，我们先用K8s原生方式：两个Service + 流量权重？原生不支持权重，所以推荐用Istio或Nginx Ingress的canary功能）
  * 为了项目轻量，我们选择**使用K8s Service + 多个Deployment + 自定义网关（简易）**，但为了展示能力，我们会提及如何用Istio实现流量分割，并给出示例。

**简化版**：先实现单版本部署，再扩展灰度。

---

## 二、CRD 定义（InferenceService）

### 2.1 设计目标

* 描述一个推理服务的完整配置
* 支持多个版本（通过`spec.versions`列表）
* 支持灰度发布配置（canary策略）

### 2.2 示例 YAML

**yaml**

复制

下载

```
apiVersion: ai.platform.io/v1
kind: InferenceService
metadata:
  name: bert-classifier
  namespace: default
spec:
  # 稳定版本配置
  stable:
    modelRef:
      name: bert-classifier
      version: v1.0
    replicas: 2
    resources:
      cpu: 2
      memory: 4Gi
    env:
      - name: LOG_LEVEL
        value: info
  # 灰度版本配置（可选）
  canary:
    enabled: true
    modelRef:
      name: bert-classifier
      version: v1.1
    replicas: 1
    resources:
      cpu: 2
      memory: 4Gi
    trafficWeight: 10   # 10% 流量到 canary
  # 自动伸缩
  autoscaling:
    enabled: true
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  # 路由配置
  routing:
    type: istio  # 或 nginx
    host: bert-classifier.example.com
```

### 2.3 定义结构（Go）

在Kubebuilder项目中创建`InferenceService`类型。

**go**

复制

下载

```
// api/v1/inferenceservice_types.go

package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelRef references a model version from the Model Manager
type ModelRef struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// ResourceRequirements defines compute resources
type ResourceRequirements struct {
    CPU    string `json:"cpu,omitempty"`
    Memory string `json:"memory,omitempty"`
    GPU    string `json:"gpu,omitempty"`
}

// EnvVar represents an environment variable
type EnvVar struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}

// ServiceVersion represents a single version of the inference service
type ServiceVersion struct {
    ModelRef  ModelRef            `json:"modelRef"`
    Replicas  int32               `json:"replicas,omitempty"`
    Resources ResourceRequirements `json:"resources,omitempty"`
    Env       []EnvVar            `json:"env,omitempty"`
}

// CanaryConfig configures canary deployment
type CanaryConfig struct {
    Enabled       bool   `json:"enabled"`
    ServiceVersion `json:",inline"`
    TrafficWeight int    `json:"trafficWeight,omitempty"` // percentage (0-100)
}

// AutoscalingConfig defines HPA configuration
type AutoscalingConfig struct {
    Enabled                        bool   `json:"enabled"`
    MinReplicas                    int32  `json:"minReplicas,omitempty"`
    MaxReplicas                    int32  `json:"maxReplicas,omitempty"`
    TargetCPUUtilizationPercentage int32  `json:"targetCPUUtilizationPercentage,omitempty"`
}

// RoutingConfig defines how traffic is routed
type RoutingConfig struct {
    Type string `json:"type,omitempty"` // istio, nginx, none
    Host string `json:"host,omitempty"`
}

// InferenceServiceSpec defines the desired state of InferenceService
type InferenceServiceSpec struct {
    Stable      ServiceVersion    `json:"stable"`
    Canary      *CanaryConfig     `json:"canary,omitempty"`
    Autoscaling AutoscalingConfig `json:"autoscaling,omitempty"`
    Routing     RoutingConfig     `json:"routing,omitempty"`
}

// InferenceServiceStatus defines the observed state
type InferenceServiceStatus struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    // StableDeploymentName is the name of the K8s deployment for stable version
    StableDeploymentName string `json:"stableDeploymentName,omitempty"`
    // CanaryDeploymentName is the name of the canary deployment (if any)
    CanaryDeploymentName string `json:"canaryDeploymentName,omitempty"`
    // ServiceName is the name of the K8s service
    ServiceName string `json:"serviceName,omitempty"`
    // URL is the externally accessible URL (if ingress/istio configured)
    URL string `json:"url,omitempty"`
    // ObservedGeneration is the last generation reconciled
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Stable",type="string",JSONPath=".spec.stable.modelRef.name"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[-1].reason"
// +kubebuilder:printcolumn:name="URL",type="string",JSONPath=".status.url"

type InferenceService struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   InferenceServiceSpec   `json:"spec,omitempty"`
    Status InferenceServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type InferenceServiceList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []InferenceService `json:"items"`
}

func init() {
    SchemeBuilder.Register(&InferenceService{}, &InferenceServiceList{})
}
```

---

## 三、Controller 实现（核心逻辑）

Controller负责将InferenceService转换为K8s Deployment、Service、HPA等资源，并处理灰度发布。

### 3.1 Reconcile 流程

1. 获取InferenceService对象
2. 创建/更新稳定版本的Deployment
3. 如果配置了canary，创建/更新canary Deployment
4. 创建Service（统一指向stable和canary，通过selector选择？原生Service不能按权重分配，需要借助Istio或Nginx）
   * 简化方案：创建两个Service（stable-service和canary-service），然后使用Nginx Ingress的canary注解来实现流量分割。
5. 配置HPA（如果autoscaling启用）
6. 更新Status（deployment名称、service名称、访问URL等）
7. 如果资源发生变化，触发相应的更新（如replicas变化）

### 3.2 构造Deployment的函数

**go**

复制

下载

```
func (r *InferenceServiceReconciler) constructDeployment(is *crdv1.InferenceService, version ServiceVersion, nameSuffix string) (*appsv1.Deployment, error) {
    deploymentName := is.Name + "-" + nameSuffix
    modelRef := version.ModelRef

    // 构建容器
    container := corev1.Container{
        Name:  "model-server",
        Image: "your-registry/model-server:latest", // 通用推理镜像，可通过环境变量指定
        Args:  []string{"--model", modelRef.Name, "--version", modelRef.Version},
        Resources: corev1.ResourceRequirements{
            Requests: corev1.ResourceList{},
            Limits:   corev1.ResourceList{},
        },
        Env: convertEnvVars(version.Env),
    }
    // 添加资源
    if cpu := version.Resources.CPU; cpu != "" {
        container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(cpu)
        container.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(cpu)
    }
    if memory := version.Resources.Memory; memory != "" {
        container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(memory)
        container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(memory)
    }
    if gpu := version.Resources.GPU; gpu != "" {
        container.Resources.Limits["nvidia.com/gpu"] = resource.MustParse(gpu)
    }

    replicas := int32(1)
    if version.Replicas != 0 {
        replicas = version.Replicas
    }

    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      deploymentName,
            Namespace: is.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(is, crdv1.GroupVersion.WithKind("InferenceService")),
            },
            Labels: map[string]string{
                "inferenceservice": is.Name,
                "version":          modelRef.Version,
            },
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "inferenceservice": is.Name,
                    "version":          modelRef.Version,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "inferenceservice": is.Name,
                        "version":          modelRef.Version,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{container},
                },
            },
        },
    }
    return deployment, nil
}
```

### 3.3 构造Service（简化版，用于稳定版本和canary共用一个Service，通过selector指向stable？但无法灰度）

我们采用**两个Service + Ingress + 注解**的方式实现灰度。

* 创建**stable-service**，selector指向stable版本的pod
* 创建**canary-service**，selector指向canary版本的pod
* 创建**Ingress**，根据配置路由流量到两个service（需要ingress-nginx支持canary）

Ingress示例（使用[nginx.ingress.kubernetes.io/canary](https://nginx.ingress.kubernetes.io/canary)注解）：

**yaml**

复制

下载

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: bert-classifier
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"   # 10% 流量到 canary-service
spec:
  rules:
  - host: bert-classifier.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: bert-classifier-stable   # 主服务
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: bert-classifier-canary   # canary服务
            port:
              number: 80
```

Controller需要根据spec.routing.type来创建对应的路由资源。如果type是"nginx"，则创建Ingress并设置canary注解；如果是"istio"，则创建VirtualService。

### 3.4 构建HPA

**go**

复制

下载

```
func (r *InferenceServiceReconciler) ensureHPA(is *crdv1.InferenceService) error {
    if !is.Spec.Autoscaling.Enabled {
        return r.deleteHPA(is)
    }
    hpa := &autoscalingv2.HorizontalPodAutoscaler{
        ObjectMeta: metav1.ObjectMeta{
            Name:      is.Name + "-hpa",
            Namespace: is.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(is, crdv1.GroupVersion.WithKind("InferenceService")),
            },
        },
        Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
            ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
                Kind:       "Deployment",
                Name:       is.Status.StableDeploymentName, // 注意：HPA只能针对一个Deployment，但我们可以只给stable做伸缩
                APIVersion: "apps/v1",
            },
            MinReplicas: &is.Spec.Autoscaling.MinReplicas,
            MaxReplicas: is.Spec.Autoscaling.MaxReplicas,
            Metrics: []autoscalingv2.MetricSpec{
                {
                    Type: autoscalingv2.ResourceMetricSourceType,
                    Resource: &autoscalingv2.ResourceMetricSource{
                        Name: corev1.ResourceCPU,
                        Target: autoscalingv2.MetricTarget{
                            Type:               autoscalingv2.UtilizationMetricType,
                            AverageUtilization: &is.Spec.Autoscaling.TargetCPUUtilizationPercentage,
                        },
                    },
                },
            },
        },
    }
    return r.CreateOrUpdate(hpa)
}
```

### 3.5 灰度发布的协调逻辑

当spec.canary改变时（比如启用/禁用、调整权重），Controller需要：

* 创建/更新canary Deployment
* 更新Ingress的canary-weight注解
* 如果canary被禁用，删除canary Deployment并移除Ingress的canary注解

---

## 四、REST API 接口（可选）

为方便算法团队，可以提供REST API，内部调用client-go创建InferenceService CRD。


| 方法   | 路径                                       | 描述             | 请求体                |
| ------ | ------------------------------------------ | ---------------- | --------------------- |
| POST   | `/api/v1/inference-services`               | 创建推理服务     | InferenceService spec |
| GET    | `/api/v1/inference-services`               | 列出所有推理服务 | -                     |
| GET    | `/api/v1/inference-services/{name}`        | 获取详情         | -                     |
| PUT    | `/api/v1/inference-services/{name}/canary` | 调整灰度权重     | `{"weight": 20}`      |
| DELETE | `/api/v1/inference-services/{name}`        | 删除推理服务     | -                     |

实现方式类似之前，使用controller-runtime client。

---

## 五、部署与验证

### 5.1 安装CRD和Controller

**bash**

复制

下载

```
make install
make run
```

### 5.2 创建推理服务示例

**yaml**

复制

下载

```
apiVersion: ai.platform.io/v1
kind: InferenceService
metadata:
  name: bert-classifier
spec:
  stable:
    modelRef:
      name: bert-classifier
      version: v1.0
    replicas: 2
    resources:
      cpu: 2
      memory: 4Gi
  canary:
    enabled: true
    modelRef:
      name: bert-classifier
      version: v1.1
    replicas: 1
    trafficWeight: 10
  autoscaling:
    enabled: true
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
  routing:
    type: nginx
    host: bert-classifier.example.com
```

应用：

**bash**

复制

下载

```
kubectl apply -f inference-service.yaml
```

### 5.3 验证状态

**bash**

复制

下载

```
kubectl get inferenceservice bert-classifier
kubectl describe inferenceservice bert-classifier
```

查看生成的资源：

**bash**

复制

下载

```
kubectl get deployment -l inferenceservice=bert-classifier
kubectl get service
kubectl get ingress
```

### 5.4 访问推理服务

假设Ingress配置了域名，可通过：

**bash**

复制

下载

```
curl -X POST http://bert-classifier.example.com/predict -d '{"text": "good movie"}'
```

实际请求会被Ingress根据权重路由到stable或canary版本。

---

## 六、与模型管理服务集成

推理服务需要从模型管理服务获取模型文件。有两种方式：

1. **在构建推理镜像时预置模型**（不灵活）
2. **在Pod启动时动态下载模型**：推理服务容器启动脚本调用模型管理服务的下载API，根据环境变量中的模型名称和版本从MinIO拉取模型。

我们可以在Deployment的容器中添加initContainer，从模型管理服务下载模型到共享卷，然后主容器加载。

**yaml**

复制

下载

```
initContainers:
- name: model-downloader
  image: curlimages/curl
  command:
  - /bin/sh
  - -c
  - |
    curl -o /models/model.bin http://model-manager/api/v1/models/bert-classifier/versions/v1.0/download
  volumeMounts:
  - name: model-volume
    mountPath: /models
```

这样模型管理服务和推理服务网关完全解耦。

---

## 七、后续扩展

* **支持更多路由策略**：基于Header、Cookie的路由，可以使用Istio的VirtualService
* **推理服务监控**：自动为Deployment添加Prometheus annotations，暴露模型服务的自定义指标（如推理延迟、错误率）
* **服务网格集成**：对接Istio，提供更丰富的流量管理能力
* **多框架支持**：允许用户指定不同的推理引擎（TensorFlow Serving、TorchServe、Triton等），在CRD中增加`engine`字段，Controller根据引擎类型生成对应的Pod配置

现在你已经有了**模型管理服务**、**任务调度服务**、**推理服务网关**三大模块，可以组合成一个完整的AI平台。如果需要，我可以继续提供平台的整体集成方案（如前端管理界面、认证授权等），或者帮助你针对这些模块编写详细的单元测试和部署脚本。
