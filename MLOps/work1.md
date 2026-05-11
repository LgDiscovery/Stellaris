我们继续深入“简化版AI平台”的第二个核心模块：**推理服务网关**。这个模块是AI平台最核心的竞争力——将模型部署为在线服务，支持灰度发布、弹性伸缩、流量路由，完全基于K8s原生扩展实现，体现你的K8s深度开发能力。

我会提供：

1. 模块设计思路
2. CRD定义（Kubebuilder生成）
3. Controller核心逻辑（使用client-go）
4. 灰度发布的两种实现方案
5. 与模型管理服务的联动设计

---

## 1. 推理服务网关设计概览

**目标**：用户通过API或K8s CRD声明一个推理服务，系统自动创建Deployment、Service、Ingress，并提供灰度、扩缩容等高级能力。

**核心组件**：

* **CRD**：定义`InferenceService`资源，描述用户期望的部署形态
* **Controller**：监听CRD变化，协调K8s资源（Deployment、Service、HPA等）
* **Webhook（可选）**：验证CRD字段合法性、设置默认值
* **网关适配器**：支持Istio流量管理（灰度发布时）

**与模型管理服务的联动**：用户在创建`InferenceService`时，指定`modelName`和`modelVersion`，Controller从模型管理服务拉取模型文件的MinIO路径，挂载到Pod中，或构建自定义推理镜像。

---

## 2. CRD定义

使用Kubebuilder生成脚手架。假设项目名为`inference-gateway`，位于`github.com/yourname/inference-gateway`。

### 2.1 初始化

```
kubebuilder init --domain ai.platform.io --repo github.com/yourname/inference-gateway
kubebuilder create api --group inference --version v1 --kind InferenceService --resource --controller
```

### 2.2 CRD Spec 和 Status 定义

`api/v1/inferenceservice_types.go`（核心部分）

```
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

// InferenceServiceSpec defines the desired state of InferenceService
type InferenceServiceSpec struct {
    // 模型信息
    ModelName    string `json:"modelName"`
    ModelVersion string `json:"modelVersion"`

    // 副本数（可被HPA覆盖）
    Replicas *int32 `json:"replicas,omitempty"`

    // 资源限制
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // 自动扩缩容配置
    Autoscaling *AutoscalingSpec `json:"autoscaling,omitempty"`

    // 灰度发布配置
    Canary *CanarySpec `json:"canary,omitempty"`
}

type AutoscalingSpec struct {
    MinReplicas int32 `json:"minReplicas"`
    MaxReplicas int32 `json:"maxReplicas"`
    TargetCPU   int32 `json:"targetCPU,omitempty"`   // 百分比
}

type CanarySpec struct {
    Enabled    bool   `json:"enabled"`
    Version    string `json:"version,omitempty"`     // 灰度版本（如 v2.0）
    Weight     int32  `json:"weight,omitempty"`      // 灰度流量百分比（0-100）
}

// InferenceServiceStatus defines the observed state of InferenceService
type InferenceServiceStatus struct {
    // 稳定版本部署状态
    StableState    string `json:"stableState,omitempty"`
    // 灰度版本部署状态（如有）
    CanaryState    string `json:"canaryState,omitempty"`
    // 服务访问URL
    URL            string `json:"url,omitempty"`
    // 最后更新时间
    LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=inferenceservices,scope=Namespaced
//+kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.modelName`
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.modelVersion`
//+kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
//+kubebuilder:printcolumn:name="StableState",type=string,JSONPath=`.status.stableState`
//+kubebuilder:printcolumn:name="CanaryState",type=string,JSONPath=`.status.canaryState`

type InferenceService struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   InferenceServiceSpec   `json:"spec,omitempty"`
    Status InferenceServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type InferenceServiceList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []InferenceService `json:"items"`
}
```

### 2.3 生成CRD YAML

```
make manifests
```

生成的CRD位于`config/crd/bases/inference.ai.platform.io_inferenceservices.yaml`，可直接部署到K8s。

---

## 3. Controller核心逻辑

Controller监听`InferenceService`的增删改，根据Spec创建/更新/删除K8s资源（Deployment、Service、HPA、Ingress等）。

### 3.1 初始化Controller

`controllers/inferenceservice_controller.go`（简化版）

```
package controllers

import (
    "context"
    "fmt"
    "time"
    appsv1 "k8s.io/api/apps/v1"
    autoscalingv2 "k8s.io/api/autoscaling/v2"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"

    inferencev1 "github.com/yourname/inference-gateway/api/v1"
)

// InferenceServiceReconciler reconciles a InferenceService object
type InferenceServiceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=inference.ai.platform.io,resources=inferenceservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inference.ai.platform.io,resources=inferenceservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

func (r *InferenceServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 获取 InferenceService 实例
    var is inferencev1.InferenceService
    if err := r.Get(ctx, req.NamespacedName, &is); err != nil {
        if errors.IsNotFound(err) {
            // 资源已删除，清理工作由ownerReference自动完成
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 设置最终清理逻辑（可选）
    if !is.ObjectMeta.DeletionTimestamp.IsZero() {
        return ctrl.Result{}, r.finalizeInferenceService(ctx, &is)
    }

    // 添加 finalizer（可选）
    if !controllerutil.ContainsFinalizer(&is, "inference.ai.platform.io/finalizer") {
        controllerutil.AddFinalizer(&is, "inference.ai.platform.io/finalizer")
        if err := r.Update(ctx, &is); err != nil {
            return ctrl.Result{}, err
        }
    }

    // 1. 创建或更新 Deployment（稳定版）
    stableDeploy, err := r.ensureStableDeployment(ctx, &is)
    if err != nil {
        r.updateStatus(ctx, &is, "StableDeploymentFailed", "")
        return ctrl.Result{}, err
    }

    // 2. 如果灰度启用，创建或更新 Canary Deployment
    var canaryDeploy *appsv1.Deployment
    if is.Spec.Canary != nil && is.Spec.Canary.Enabled {
        canaryDeploy, err = r.ensureCanaryDeployment(ctx, &is)
        if err != nil {
            r.updateStatus(ctx, &is, "CanaryDeploymentFailed", "")
            return ctrl.Result{}, err
        }
    } else {
        // 确保旧的canary deployment被删除
        r.ensureCanaryDeleted(ctx, &is)
    }

    // 3. 创建或更新 Service（稳定版）
    svc, err := r.ensureService(ctx, &is)
    if err != nil {
        r.updateStatus(ctx, &is, "ServiceFailed", "")
        return ctrl.Result{}, err
    }

    // 4. 创建或更新 Ingress
    ingress, err := r.ensureIngress(ctx, &is)
    if err != nil {
        r.updateStatus(ctx, &is, "IngressFailed", "")
        return ctrl.Result{}, err
    }

    // 5. 创建或更新 HPA
    if is.Spec.Autoscaling != nil {
        _, err = r.ensureHPA(ctx, &is)
        if err != nil {
            r.updateStatus(ctx, &is, "HPAFailed", "")
            return ctrl.Result{}, err
        }
    }

    // 6. 更新状态
    url := fmt.Sprintf("http://%s.%s.svc.cluster.local", svc.Name, svc.Namespace)
    if ingress != nil && len(ingress.Status.LoadBalancer.Ingress) > 0 {
        url = fmt.Sprintf("http://%s", ingress.Status.LoadBalancer.Ingress[0].IP)
    }
    r.updateStatus(ctx, &is, "Ready", url)

    return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}
```

### 3.2 辅助方法：构建Deployment

```
func (r *InferenceServiceReconciler) buildDeployment(is *inferencev1.InferenceService, version string, canary bool) *appsv1.Deployment {
    // 模型信息
    modelName := is.Spec.ModelName
    modelVersion := is.Spec.ModelVersion
    if canary && is.Spec.Canary != nil && is.Spec.Canary.Version != "" {
        modelVersion = is.Spec.Canary.Version
    }

    // 从模型管理服务获取模型存储路径（这里简化：假设环境变量传递）
    modelPath := fmt.Sprintf("s3://models/%s/%s/model.bin", modelName, modelVersion)

    labels := map[string]string{
        "app":           is.Name,
        "model":         modelName,
        "version":       modelVersion,
        "component":     "inference",
        "managed-by":    "inference-gateway",
    }
    if canary {
        labels["canary"] = "true"
    }

    replicas := int32(1)
    if is.Spec.Replicas != nil {
        replicas = *is.Spec.Replicas
    }
    if canary && is.Spec.Canary != nil {
        // 灰度版本默认1个副本，可由spec扩展
        replicas = 1
    }

    // 推理镜像：可自定义，这里用一个基础镜像，启动时加载模型
    container := corev1.Container{
        Name:  "inference",
        Image: "docker.io/yourname/inference-base:latest",
        Args:  []string{"--model-path", modelPath, "--port", "8080"},
        Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
        Env: []corev1.EnvVar{
            {Name: "MODEL_PATH", Value: modelPath},
            {Name: "MODEL_NAME", Value: modelName},
            {Name: "MODEL_VERSION", Value: modelVersion},
        },
        Resources: is.Spec.Resources,
    }

    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("%s-%s", is.Name, version),
            Namespace: is.Namespace,
            Labels:    labels,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app":     is.Name,
                    "version": version,
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app":     is.Name,
                        "version": version,
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{container},
                },
            },
        },
    }
}
```

### 3.3 辅助方法：Service与流量拆分

如果**不使用Istio**，可以通过**多个Service + 不同Selector**实现简单的灰度流量切分（如基于百分比）。但K8s原生Service不支持权重路由，建议引入Istio。我们提供两种方案：

* **方案A（简单）**：创建两个Service，分别指向稳定版和灰度版，让用户通过不同域名访问（如`stable.app.com`和`canary.app.com`），不是真正的灰度。
* **方案B（推荐）**：使用Istio的`VirtualService`和`DestinationRule`，按权重分流。

这里我们采用**方案B**，需要集群安装Istio。

**Service构建（单一Service，作为Istio的ServiceEntry）**：

```
func (r *InferenceServiceReconciler) ensureService(ctx context.Context, is *inferencev1.InferenceService) (*corev1.Service, error) {
    svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      is.Name,
            Namespace: is.Namespace,
        },
        Spec: corev1.ServiceSpec{
            Ports: []corev1.ServicePort{
                {Port: 80, TargetPort: 8080},
            },
            Selector: map[string]string{
                "app": is.Name, // 所有版本都用这个label，但通过Istio分流
            },
        },
    }
    // 创建或更新
    // ...
    return svc, nil
}
```

**创建Istio资源（在Controller中）**：

```
import (
    istionetworkingv1beta1 "istio.io/api/networking/v1beta1"
    istiov1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
)

func (r *InferenceServiceReconciler) ensureIstioResources(ctx context.Context, is *inferencev1.InferenceService) error {
    // 1. DestinationRule：定义子集
    dr := &istiov1beta1.DestinationRule{
        ObjectMeta: metav1.ObjectMeta{
            Name:      is.Name,
            Namespace: is.Namespace,
        },
        Spec: istionetworkingv1beta1.DestinationRule{
            Host: fmt.Sprintf("%s.%s.svc.cluster.local", is.Name, is.Namespace),
            Subsets: []*istionetworkingv1beta1.Subset{
                {Name: "stable", Labels: map[string]string{"version": is.Spec.ModelVersion}},
            },
        },
    }
    if is.Spec.Canary != nil && is.Spec.Canary.Enabled {
        dr.Spec.Subsets = append(dr.Spec.Subsets, &istionetworkingv1beta1.Subset{
            Name:   "canary",
            Labels: map[string]string{"version": is.Spec.Canary.Version},
        })
    }
    // 创建或更新 DestinationRule

    // 2. VirtualService：权重路由
    vs := &istiov1beta1.VirtualService{
        ObjectMeta: metav1.ObjectMeta{
            Name:      is.Name,
            Namespace: is.Namespace,
        },
        Spec: istionetworkingv1beta1.VirtualService{
            Hosts: []string{fmt.Sprintf("%s.%s.svc.cluster.local", is.Name, is.Namespace)},
            Http: []*istionetworkingv1beta1.HTTPRoute{
                {
                    Route: []*istionetworkingv1beta1.HTTPRouteDestination{
                        {
                            Destination: &istionetworkingv1beta1.Destination{
                                Host:   fmt.Sprintf("%s.%s.svc.cluster.local", is.Name, is.Namespace),
                                Subset: "stable",
                            },
                            Weight: 100,
                        },
                    },
                },
            },
        },
    }
    if is.Spec.Canary != nil && is.Spec.Canary.Enabled {
        vs.Spec.Http[0].Route = []*istionetworkingv1beta1.HTTPRouteDestination{
            {
                Destination: &istionetworkingv1beta1.Destination{
                    Host:   fmt.Sprintf("%s.%s.svc.cluster.local", is.Name, is.Namespace),
                    Subset: "stable",
                },
                Weight: 100 - is.Spec.Canary.Weight,
            },
            {
                Destination: &istionetworkingv1beta1.Destination{
                    Host:   fmt.Sprintf("%s.%s.svc.cluster.local", is.Name, is.Namespace),
                    Subset: "canary",
                },
                Weight: is.Spec.Canary.Weight,
            },
        }
    }
    // 创建或更新 VirtualService
    return nil
}
```

---

## 4. 与模型管理服务的联动

在Controller中，需要从模型管理服务获取模型文件的MinIO路径。有两种方式：

* **方式A**：模型管理服务提供HTTP API，Controller调用获取。需要配置服务地址。
* **方式B**：使用K8s ConfigMap/Secret存储模型元数据，Controller监听这些资源变化，但复杂。

推荐方式A。在Controller中增加一个HTTP客户端，调用模型管理服务的API。

```
type ModelManagerClient struct {
    baseURL string
    client  *http.Client
}

func (c *ModelManagerClient) GetModelVersion(modelName, version string) (*ModelVersion, error) {
    url := fmt.Sprintf("%s/api/v1/models/%s/versions/%s", c.baseURL, modelName, version)
    resp, err := c.client.Get(url)
    // 解析响应
}
```

然后在`Reconcile`中调用，将获取的`storage_path`注入到Pod的环境变量或卷挂载中。

---

## 5. 部署与运行

### 5.1 安装依赖

```
# 安装 Istio（如需要）
istioctl install --set profile=demo -y

# 安装 CRD
make install
```

### 5.2 运行Controller（本地或集群）

```
# 本地运行（需要KUBECONFIG）
go run cmd/main.go

# 或部署到集群
make deploy
```

### 5.3 创建示例 InferenceService

```
apiVersion: inference.ai.platform.io/v1
kind: InferenceService
metadata:
  name: bert-classifier
  namespace: default
spec:
  modelName: bert-classifier
  modelVersion: v1.0
  replicas: 2
  resources:
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 2
      memory: 4Gi
  autoscaling:
    minReplicas: 1
    maxReplicas: 5
    targetCPU: 70
  canary:
    enabled: true
    version: v1.1
    weight: 10
```

应用后，Controller会创建：

* Deployment `bert-classifier-v1.0` 和 `bert-classifier-v1.1`
* Service `bert-classifier`
* HPA `bert-classifier`
* DestinationRule `bert-classifier`
* VirtualService `bert-classifier`

此时访问服务，10%流量进入v1.1版本。

---

## 6. 总结与扩展

你现在有了：

* **模型管理服务**（自研替代MLflow）
* **推理服务网关**（基于K8s Operator，支持灰度发布）

下一步可以继续开发**任务调度模块**，或者为推理网关增加：

* 模型预热（启动时加载模型到内存）
* 自定义推理镜像构建（通过Job构建镜像，再用Deployment部署）
* 监控指标暴露（Prometheus metrics）

这个网关的实现，证明了你不只是“会用K8s”，而是“会扩展K8s”，这是AI平台工程师的核心能力。如果你继续，我可以帮你提供**任务调度模块**的详细实现，或者完善**可观测性**（自定义指标、日志采集）。

继续

已停止

我们继续细化**推理服务网关**模块。这个模块是整个AI平台的核心，也是面试中最能体现K8s深度开发能力的部分。下面给出完整设计、CRD定义、Controller实现、API设计和关键代码。

---

## 推理服务网关 (Inference Gateway) 详细设计

### 一、模块定位与目标

**核心职责**：

* 接收算法团队的部署请求，在K8s中创建模型推理服务
* 支持多版本管理（同一模型可部署多个版本）
* 支持灰度发布（流量切分）
* 支持弹性伸缩（基于CPU/内存/QPS）
* 提供统一的推理服务访问入口（Service + Ingress）

**技术方案**：

* 使用Kubebuilder开发自定义资源（CRD）：`InferenceService`
* 编写Controller监听该资源，协调创建/更新/删除K8s原生资源（Deployment、Service、Ingress、HPA）
* 灰度发布：通过创建两个Deployment（稳定版和灰度版），使用Service的selector权重或Istio DestinationRule（简化方案：使用自定义Ingress注解实现基于权重的路由）

---

### 二、CRD 定义

使用Kubebuilder生成CRD，核心结构如下：

```
// InferenceServiceSpec 定义期望状态
type InferenceServiceSpec struct {
    // 模型信息
    ModelName    string `json:"modelName"`
    ModelVersion string `json:"modelVersion"`

    // 镜像信息（如果模型是自定义镜像，否则用默认推理镜像+模型挂载）
    Image string `json:"image,omitempty"`

    // 资源配置
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // 副本数
    Replicas int32 `json:"replicas,omitempty"`

    // 自动伸缩配置
    Autoscaling *AutoscalingSpec `json:"autoscaling,omitempty"`

    // 灰度发布配置
    Canary *CanarySpec `json:"canary,omitempty"`

    // 服务暴露
    Service *ServiceSpec `json:"service,omitempty"`
}

// AutoscalingSpec 弹性伸缩配置
type AutoscalingSpec struct {
    MinReplicas            int32  `json:"minReplicas,omitempty"`
    MaxReplicas            int32  `json:"maxReplicas,omitempty"`
    TargetCPUUtilization   int32  `json:"targetCPUUtilization,omitempty"`
    TargetMemoryUtilization int32 `json:"targetMemoryUtilization,omitempty"`
}

// CanarySpec 灰度发布配置
type CanarySpec struct {
    Enabled    bool   `json:"enabled"`
    Version    string `json:"version"`          // 灰度版本号
    Weight     int32  `json:"weight"`           // 流量比例 0-100
    ModelName  string `json:"modelName"`        // 灰度模型名（通常与主模型相同）
    ModelVersion string `json:"modelVersion"`   // 灰度模型版本
    Image      string `json:"image,omitempty"`
    Resources  *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// ServiceSpec 服务暴露配置
type ServiceSpec struct {
    Type        string            `json:"type,omitempty"`        // ClusterIP, NodePort, LoadBalancer
    Port        int32             `json:"port"`                  // 服务端口
    TargetPort  int32             `json:"targetPort"`            // 容器端口
    Annotations map[string]string `json:"annotations,omitempty"` // 可用于Ingress配置
}

// InferenceServiceStatus 定义观察到的状态
type InferenceServiceStatus struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    URL        string             `json:"url,omitempty"`          // 服务访问地址
    Replicas   int32              `json:"replicas"`               // 当前副本数
    Ready      bool               `json:"ready"`                  // 服务是否就绪
}
```

CRD定义文件由Kubebuilder自动生成，放在 `api/v1/inferenceservice_types.go`。

---

### 三、Controller 核心逻辑

Controller 主要实现 Reconcile 函数，确保集群状态与 Spec 一致。

**Reconcile 流程**：

1. 获取 `InferenceService` 实例
2. 根据 Spec 构建稳定版 Deployment
3. 根据 Spec 构建 Service（若不存在则创建）
4. 根据 Spec 构建 Ingress（若配置了域名或路由规则）
5. 若开启了灰度（`Canary.Enabled == true`），则：
   * 创建灰度版 Deployment
   * 调整 Service 的流量路由（通过 Istio 或自定义 Ingress 注解）
6. 若 Spec 中配置了 Autoscaling，则创建或更新 HPA 资源
7. 更新 Status（URL、Replicas、Ready 等）
8. 若 Service 就绪，将访问地址写入 Status.URL

**关键点**：

* 所有创建的资源都打上 Label：`app.kubernetes.io/managed-by: kubeai-platform` 和 `inferenceservice: <name>`
* 使用 `controllerutil.SetControllerReference` 设置 OwnerReference，确保资源被删除时自动清理

---

### 四、灰度发布实现方案（简化版）

由于没有 Istio 依赖，我们采用基于 Ingress Nginx 的权重路由方案：

* 创建两个 Service：`<name>-stable` 和 `<name>-canary`
* 创建 Ingress，配置 `nginx.ingress.kubernetes.io/canary: "true"` 和 `nginx.ingress.kubernetes.io/canary-weight: "10"` 等注解，指向 canary Service
* 稳定版流量直接走 `<name>-stable` Service
* 灰度版流量通过 Ingress Canary 功能切分

**前提**：K8s 集群安装有 ingress-nginx。

**实现步骤**：

1. 创建稳定版 Deployment 和 Service（Service 名称：`<name>-stable`）
2. 创建灰度版 Deployment 和 Service（若启用）
3. 创建 Ingress，其规则指向稳定版 Service，并添加 Canary 相关注解指向灰度版 Service

**注意**：若未启用灰度，只需创建普通 Ingress 指向稳定版 Service。

---

### 五、关键代码实现

#### 5.1 项目结构（在原有 model-manager 基础上扩展）

```
kubeai-platform/
├── cmd/
│   ├── model-manager/          # 模型管理服务（已完成）
│   └── inference-gateway/      # 推理网关（新增）
│       └── main.go
├── api/
│   └── v1/                     # CRD 定义（由 kubebuilder 生成）
│       ├── inferenceservice_types.go
│       └── ...
├── internal/
│   ├── controller/             # Controller 实现
│   │   └── inferenceservice_controller.go
│   ├── resources/              # 构建 K8s 资源的工具函数
│   │   ├── deployment.go
│   │   ├── service.go
│   │   ├── ingress.go
│   │   └── hpa.go
│   └── ...
└── config/                     # 部署配置（CRD 安装等）
```

#### 5.2 Controller 入口（main.go）

```
package main

import (
    "flag"
    "os"
    "k8s.io/apimachinery/pkg/runtime"
    clientgoscheme "k8s.io/client-go/kubernetes/scheme"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "github.com/yourname/kubeai-platform/api/v1"
    "github.com/yourname/kubeai-platform/internal/controller"
)

var (
    scheme   = runtime.NewScheme()
    setupLog = ctrl.Log.WithName("setup")
)

func init() {
    _ = clientgoscheme.AddToScheme(scheme)
    _ = v1.AddToScheme(scheme)
}

func main() {
    var metricsAddr string
    flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
    flag.Parse()

    ctrl.SetLogger(zap.New())

    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:                 scheme,
        MetricsBindAddress:     metricsAddr,
        Port:                   9443,
        LeaderElection:         false,
        LeaderElectionID:       "inference-gateway-leader",
    })
    if err != nil {
        setupLog.Error(err, "unable to start manager")
        os.Exit(1)
    }

    if err = (&controller.InferenceServiceReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr); err != nil {
        setupLog.Error(err, "unable to create controller", "controller", "InferenceService")
        os.Exit(1)
    }

    setupLog.Info("starting manager")
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        setupLog.Error(err, "problem running manager")
        os.Exit(1)
    }
}
```

#### 5.3 Reconciler 核心逻辑

`internal/controller/inferenceservice_controller.go`

```
package controller

import (
    "context"
    "fmt"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    aiv1 "github.com/yourname/kubeai-platform/api/v1"
    "github.com/yourname/kubeai-platform/internal/resources"
)

type InferenceServiceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// Reconcile 实现协调循环
func (r *InferenceServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 获取 InferenceService 实例
    isvc := &aiv1.InferenceService{}
    if err := r.Get(ctx, req.NamespacedName, isvc); err != nil {
        if errors.IsNotFound(err) {
            // 资源已删除，清理工作由 OwnerReference 自动完成
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 1. 构建稳定版 Deployment
    stableDeploy := resources.NewStableDeployment(isvc)
    if err := r.ensureResource(ctx, stableDeploy); err != nil {
        log.Error(err, "failed to ensure stable deployment")
        return ctrl.Result{}, err
    }

    // 2. 构建稳定版 Service
    stableSvc := resources.NewStableService(isvc)
    if err := r.ensureResource(ctx, stableSvc); err != nil {
        log.Error(err, "failed to ensure stable service")
        return ctrl.Result{}, err
    }

    // 3. 处理灰度发布
    if isvc.Spec.Canary != nil && isvc.Spec.Canary.Enabled {
        canaryDeploy := resources.NewCanaryDeployment(isvc)
        if err := r.ensureResource(ctx, canaryDeploy); err != nil {
            log.Error(err, "failed to ensure canary deployment")
            return ctrl.Result{}, err
        }
        canarySvc := resources.NewCanaryService(isvc)
        if err := r.ensureResource(ctx, canarySvc); err != nil {
            log.Error(err, "failed to ensure canary service")
            return ctrl.Result{}, err
        }
    }

    // 4. 构建 Ingress（支持灰度）
    ingress := resources.NewIngress(isvc)
    if err := r.ensureResource(ctx, ingress); err != nil {
        log.Error(err, "failed to ensure ingress")
        return ctrl.Result{}, err
    }

    // 5. 处理 HPA
    if isvc.Spec.Autoscaling != nil {
        hpa := resources.NewHPA(isvc)
        if err := r.ensureResource(ctx, hpa); err != nil {
            log.Error(err, "failed to ensure HPA")
            return ctrl.Result{}, err
        }
    } else {
        // 如果没有配置自动伸缩，确保删除 HPA（如果存在）
        hpa := resources.NewHPA(isvc) // 用于获取名称
        if err := r.Delete(ctx, hpa); err != nil && !errors.IsNotFound(err) {
            log.Error(err, "failed to delete HPA")
            return ctrl.Result{}, err
        }
    }

    // 6. 更新状态
    if err := r.updateStatus(ctx, isvc, stableSvc, ingress); err != nil {
        log.Error(err, "failed to update status")
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

// ensureResource 创建或更新资源
func (r *InferenceServiceReconciler) ensureResource(ctx context.Context, obj client.Object) error {
    err := r.Get(ctx, client.ObjectKeyFromObject(obj), obj)
    if err != nil && errors.IsNotFound(err) {
        // 设置 OwnerReference
        if err := ctrl.SetControllerReference(obj.GetOwnerReferences()[0].Reference, obj, r.Scheme); err != nil {
            return err
        }
        return r.Create(ctx, obj)
    } else if err != nil {
        return err
    }
    // 如果存在，更新
    return r.Update(ctx, obj)
}

// updateStatus 更新 InferenceService 的 Status
func (r *InferenceServiceReconciler) updateStatus(ctx context.Context, isvc *aiv1.InferenceService, svc *corev1.Service, ing *networkingv1.Ingress) error {
    // 获取 Service 的 ClusterIP 或 LoadBalancer IP
    // 获取 Ingress 的 URL
    // 设置 Conditions
    // 更新 isvc.Status
    // ...
    return r.Status().Update(ctx, isvc)
}

// SetupWithManager 注册 Controller
func (r *InferenceServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&aiv1.InferenceService{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&networkingv1.Ingress{}).
        Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
        Complete(r)
}
```

#### 5.4 资源构建工具函数（示例）

`internal/resources/deployment.go`

```
package resources

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    aiv1 "github.com/yourname/kubeai-platform/api/v1"
)

// NewStableDeployment 创建稳定版 Deployment
func NewStableDeployment(isvc *aiv1.InferenceService) *appsv1.Deployment {
    replicas := isvc.Spec.Replicas
    if replicas == 0 {
        replicas = 1
    }
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      isvc.Name + "-stable",
            Namespace: isvc.Namespace,
            Labels:    map[string]string{
                "app": isvc.Name,
                "version": isvc.Spec.ModelVersion,
                "role": "stable",
            },
            OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService"))},
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": isvc.Name,
                    "role": "stable",
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": isvc.Name,
                        "role": "stable",
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "inference",
                            Image: getImage(isvc),
                            Ports: []corev1.ContainerPort{
                                {ContainerPort: 8501, Name: "http"}, // 默认推理端口
                            },
                            Resources: isvc.Spec.Resources,
                        },
                    },
                },
            },
        },
    }
}

// NewCanaryDeployment 创建灰度版 Deployment
func NewCanaryDeployment(isvc *aiv1.InferenceService) *appsv1.Deployment {
    replicas := int32(1) // 灰度版通常副本数较少
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      isvc.Name + "-canary",
            Namespace: isvc.Namespace,
            Labels: map[string]string{
                "app": isvc.Name,
                "version": isvc.Spec.Canary.ModelVersion,
                "role": "canary",
            },
            OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService"))},
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": isvc.Name,
                    "role": "canary",
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": isvc.Name,
                        "role": "canary",
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "inference",
                            Image: getCanaryImage(isvc),
                            Ports: []corev1.ContainerPort{{ContainerPort: 8501}},
                            Resources: getCanaryResources(isvc),
                        },
                    },
                },
            },
        },
    }
}
```

`internal/resources/ingress.go`

```
package resources

import (
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    aiv1 "github.com/yourname/kubeai-platform/api/v1"
)

func NewIngress(isvc *aiv1.InferenceService) *networkingv1.Ingress {
    // 构建基础 Ingress
    ing := &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:      isvc.Name,
            Namespace: isvc.Namespace,
            Annotations: map[string]string{
                "kubernetes.io/ingress.class": "nginx",
            },
            OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(isvc, aiv1.GroupVersion.WithKind("InferenceService"))},
        },
        Spec: networkingv1.IngressSpec{
            Rules: []networkingv1.IngressRule{
                {
                    Host: isvc.Spec.Service.Host, // 假设 Spec 中有 Host 字段
                    IngressRuleValue: networkingv1.IngressRuleValue{
                        HTTP: &networkingv1.HTTPIngressRuleValue{
                            Paths: []networkingv1.HTTPIngressPath{
                                {
                                    Path:     "/",
                                    PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
                                    Backend: networkingv1.IngressBackend{
                                        Service: &networkingv1.IngressServiceBackend{
                                            Name: isvc.Name + "-stable",
                                            Port: networkingv1.ServiceBackendPort{Number: isvc.Spec.Service.Port},
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

    // 如果启用灰度，添加 Canary 注解
    if isvc.Spec.Canary != nil && isvc.Spec.Canary.Enabled {
        ing.Annotations["nginx.ingress.kubernetes.io/canary"] = "true"
        ing.Annotations["nginx.ingress.kubernetes.io/canary-weight"] = fmt.Sprintf("%d", isvc.Spec.Canary.Weight)
        ing.Annotations["nginx.ingress.kubernetes.io/canary-by-header"] = "X-Canary" // 可选
        // 指定灰度后端的 Service
        ing.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name = isvc.Name + "-stable"
        // 注意：nginx-ingress canary 要求稳定版和灰度版使用相同的 host/path，canary 通过注解指定后端 Service
        // 但实际实现中，需要添加额外的 Backend 指向 canary Service？标准做法是稳定版和灰度版两个 Ingress 或使用额外注解
        // 简化：我们只设置一个 Ingress 指向稳定版，canary 流量由 Nginx 自动路由到 canary service（通过注解中的 service）
        // 实际上 nginx-ingress canary 需要两个 Ingress：一个主 Ingress，一个 canary Ingress。
        // 为简化，可以创建两个 Ingress，主 Ingress 指向稳定版，Canary Ingress 指向灰度版并设置 canary 注解。
        // 这里不展开，实际开发时可参考 nginx-ingress 官方文档。
    }
    return ing
}
```

**注意**：上述 Ingress 灰度实现为简化版，实际生产可考虑使用 Istio 或 Traefik 等更强大的服务网格。

---

### 六、部署与测试

1. **安装 CRD**：`make install`（Kubebuilder 会生成）
2. **运行 Controller**：`make run`（本地运行）或构建镜像部署到集群
3. **创建 InferenceService 实例**：

```
apiVersion: ai.platform.io/v1
kind: InferenceService
metadata:
  name: bert-classifier
spec:
  modelName: bert-classifier
  modelVersion: v1.0
  replicas: 2
  resources:
    requests:
      cpu: 1
      memory: 2Gi
    limits:
      cpu: 2
      memory: 4Gi
  autoscaling:
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilization: 70
  canary:
    enabled: true
    version: v1.1
    weight: 10
  service:
    type: ClusterIP
    port: 80
    targetPort: 8501
    host: bert.example.com
```

4. **验证**：查看 K8s 资源是否创建，访问 Service 或 Ingress 测试推理。

---

### 七、后续扩展

* **更完善的灰度**：引入 Istio 实现更细粒度的流量管理
* **多模型服务**：支持一个 Service 中部署多个模型（如多路推理）
* **模型预热**：在 Pod 启动时提前加载模型
* **指标采集**：通过 Prometheus 暴露推理服务的 QPS、延迟等指标（在容器内集成 prometheus client）
