### 一、项目定位与目标

**项目名称**：KubeAI-Platform

定位：为算法工程师提供模型管理、在线推理、离线训练的一站式平台，所有能力通过web控制台和API开发

**核心目标**：

- 不依赖MLFLow等现成工具，自己实现模型管理模块（注册、版本、元数据）
- 基于K8s开发推理服务网关（支持多版本、灰度、弹性）
- 基于K8s Job开发离线任务调度（支持PyTorch/TensorFlow训练任务）
- 构建统一的可观测性看板（调用链、指标、日志）

**技术栈选择**：

- 后端语言：Golang
- Web框架：Gin / Go-Zero
- 数据库：PostgreSql (存储模型元数据、任务记录)
- 缓存：Redis(任务队列、分布式锁)
- 容器编排 Kubernetes
- k8s扩展：kubebuilder (开发CRD + Controller)
- 服务通信：GRPC (内部服务间调用)
- 模型存储: MinIO

### 二、 系统架构图

```plaintext
┌─────────────────────────────────────────────────────────────┐
│                      Web UI / API Gateway                    │
│                     (Gin + Swagger UI)                       │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐   ┌─────────────────┐   ┌─────────────────┐
│  模型管理服务  │   │  推理服务网关   │   │  任务调度服务   │
│ (Go + PG)     │   │ (Go + K8s)     │   │ (Go + K8s)     │
│ - 模型注册    │   │ - 部署推理Pod  │   │ - 创建训练Job  │
│ - 版本管理    │   │ - 灰度路由     │   │ - 状态跟踪     │
│ - 元数据存储  │   │ - 弹性伸缩     │   │ - 日志收集     │
└───────────────┘   └─────────────────┘   └─────────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
                    ┌─────────────────┐
                    │   Kubernetes    │
                    │ - Deployment    │
                    │ - Service       │
                    │ - Ingress       │
                    │ - Job/CronJob   │
                    │ - CRD           │
                    └─────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐   ┌─────────────────┐   ┌─────────────────┐
│   PostgreSQL  │   │      Redis      │   │ 对象存储(MinIO) │
│ (元数据)      │   │   (任务队列)    │   │  (模型文件)     │
└───────────────┘   └─────────────────┘   └─────────────────┘
```

### 三、 核心模块详细设计

#### 1. 模型管理服务

**目标**：替代MLflow核心功能，提供模型注册、版本管理、元数据检索能力。

**数据表设计**：

```sql
-- 模型注册表
CREATE TABLE models (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,          -- 模型名称，唯一标识
    description TEXT,                                   -- 模型描述
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 模型版本表
CREATE TABLE model_versions (
    id          BIGSERIAL PRIMARY KEY,
    model_id    BIGINT NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    version     VARCHAR(32) NOT NULL,                   -- 语义化版本，如 v1.0.0
    storage_path VARCHAR(512) NOT NULL,                 -- 在MinIO中的存储路径
    framework   VARCHAR(32),                            -- pytorch / tensorflow / sklearn
    metrics     JSONB,                                   -- 存储性能指标，如 {"accuracy": 0.95, "f1": 0.93}
    status      VARCHAR(16) NOT NULL DEFAULT 'active',  -- active / archived
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(model_id, version)
);

-- 可选：模型标签表（便于分类检索）
CREATE TABLE model_tags (
    id         BIGSERIAL PRIMARY KEY,
    model_id   BIGINT NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    tag        VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_model_tags_tag ON model_tags(tag);
```

**说明**：

* `storage_path` 是MinIO中的完整路径，例如 `models/bert-classifier/v1.0/model.bin`
* `metrics` 使用JSONB类型，存储任意结构，便于后续扩展
* `status` 用于软删除或版本归档，只有 `active` 的版本可用于部署

**API设计**：

所有接口挂载在 `/api/v1` 下，返回JSON。

**2.1 模型管理**


| 方法   | 路径           | 描述                         | 请求体示例                                                   | 响应示例                                    |
| ------ | -------------- | ---------------------------- | ------------------------------------------------------------ | ------------------------------------------- |
| POST   | `/models`      | 创建模型                     | `{"name": "bert-classifier", "description": "中文情感分类"}` | `{"id": 1, "name": "bert-classifier", ...}` |
| GET    | `/models`      | 列表模型                     | `?page=1&size=20`                                            | `{"total": 10, "items": [...]}`             |
| GET    | `/models/{id}` | 获取模型详情                 | -                                                            | 同上                                        |
| PUT    | `/models/{id}` | 更新模型描述                 | `{"description": "new desc"}`                                | 更新后模型对象                              |
| DELETE | `/models/{id}` | 删除模型（级联删除所有版本） | -                                                            | `204`                                       |

**2.2 模型版本管理**


| 方法 | 路径                                            | 描述         | 请求体/参数                                                                      | 响应     |
| ---- | ----------------------------------------------- | ------------ | -------------------------------------------------------------------------------- | -------- |
| POST | `/models/{modelId}/versions`                    | 上传新版本   | `multipart/form-data`: `version`, `framework`, `metrics`(json), `file`(模型文件) | 版本对象 |
| GET  | `/models/{modelId}/versions`                    | 列出所有版本 | `?status=active`                                                                 | 版本列表 |
| GET  | `/models/{modelId}/versions/{version}`          | 获取版本详情 | -                                                                                | 版本对象 |
| PUT  | `/models/{modelId}/versions/{version}/status`   | 切换状态     | `{"status": "archived"}`                                                         | 版本对象 |
| GET  | `/models/{modelId}/versions/{version}/download` | 下载模型文件 | -                                                                                | 文件流   |

**说明**：

* 上传时，服务端自动生成 `storage_path`：`models/{modelName}/{version}/model.bin`（若上传的是zip则解压，本例简化处理）
* 下载时从MinIO读取文件流返回

**核心实现**：

- 使用MioIO SDK 上传/下载模型文件
- 元数据存储在PostgreSQL，版本切换通过更新状态在字段实现

#### 2. 推理服务网关

**目标**：用户通过API部署模型为在线推理服务，支持多版本、灰度发布、弹性伸缩

**K8s CRD设计 （使用kubebuilder）**:

```yaml
apiVersion: kubeai.platform.io/v1
kind: InferenceService
metadata:
  name: bert-classifier
spec:
  modelName: bert-classifier
  modelVersion: v1.0
  replicas: 2
  resources:
    cpu: 2
    memory: 4Gi
    gpu: 1
  autoScaling:
     minReplicas: 1
     maxReplicas: 10
     targetCPUUtilization: 70
  canary:
     enable: true
     version: v1.1.
     weight: 10

```

**Controller逻辑**：

- 监听InferenceService资源变化
- 根据spec创建Deployment、Service、Ingress
- 实现灰度发布：创建两个Deployment(稳定版和灰度版),通过Service的selector权重或Istio流量管理
- 集成HPA(HorizontalPodAutoscaler)实现弹性伸缩


**API设计（供算法团队调用）**：

- POST  /api/v1/deployments -创建推理服务（接受模型版本ID,返回service URL）
- GET  /api/v1/deployments/{id} -查询部署状态
- PUT /api/v1/deployments/{id}/canary -调整灰度比例
- DELETE /api/v1/deployments/{id} -下线服务

**关键技术点**：

- 使用client-go操作k8s资源
- 实现自定义Controller（通过Informer监听资源变化）
- 灰度发布可通过修改Service的spec.selector或使用Istio的DestinationRule

### 3. 任务调度服务

**目标**：支持算法团队提交离线训练任务（pyTorch/TensorFlow），自动调度到K8s运行，跟踪状态和日志。

**k8s CRD设计**

```yaml
apiVersion: kubeai.platform.io/v1
kind: TrainingJob
metadata:
  name: resent-training-20250426
spec:
  framework: pytorch
  command: ["python","train.py"]
  resources:
    cpu: 4
    memory: 8Gi
    gpu: 2
  volumes:
  - name: code
   source: git://github.com/... 
  - name: data
    source: s3://bucket/data
  env:
  - name: EPOCHS
    value: "100"  
```

**Controller逻辑**：

* 监听`TrainingJob`资源
* 创建K8s Job，并自动注入环境变量、挂载存储
* 通过Job的Pod日志收集训练日志（可对接ELK或直接提供API）
* 跟踪Job状态，更新CRD的status字段

**API设计**：

* `POST /api/v1/jobs` - 提交训练任务
* `GET /api/v1/jobs/{id}` - 查询任务状态
* `GET /api/v1/jobs/{id}/logs` - 获取实时日志（WebSocket或轮询）

**关键技术点**：

* 使用K8s Job API创建一次性任务
* 支持PyTorch分布式训练（需要创建`PyTorchJob`，可使用Kubeflow的`PyTorchJob` CRD，但为了简化，可先支持单机）
* 日志收集可通过K8s API获取Pod日志，或使用`kubectl logs`封装


### 四、 代码结构示例

```plaintext
kubeai-platform/
├── cmd/
│   ├── model-manager/          # 模型管理服务入口
│   ├── inference-gateway/      # 推理网关入口（包含Controller）
│   └── job-scheduler/          # 任务调度服务入口
├── api/
│   ├── proto/                  # GRPC定义（可选）
│   └── openapi/                # Swagger文档
├── internal/
│   ├── model/                  # 模型管理业务逻辑
│   │   ├── handler/            # HTTP handler
│   │   ├── service/            # 业务服务
│   │   └── repository/         # 数据库操作
│   ├── inference/              # 推理网关逻辑
│   │   ├── controller/         # K8s Controller
│   │   └── client/             # K8s client封装
│   ├── job/                    # 任务调度逻辑
│   │   ├── controller/
│   │   └── client/
│   └── pkg/                    # 公共库
│       ├── k8s/                # K8s工具
│       ├── storage/            # MinIO封装
│       └── metrics/            # Prometheus指标
├── configs/                    # 配置文件
├── deployments/                # K8s部署文件
│   ├── crds/                   # 自定义CRD定义
│   └── manifests/              # 平台自身部署
├── docs/                       # 设计文档
├── go.mod
└── Makefile
```
