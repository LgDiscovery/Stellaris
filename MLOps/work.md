## “简化版AI平台”项目细化方案

你原有的学习计划中，第7-8周是一个“云原生MLOps平台”项目。现在我们把定位从“跑通MLOps流程”升级为**“开发一个面向算法团队的AI平台（极简版）”**——这是一个可以写入简历的完整系统，重点体现你的**平台开发能力**而非工具使用能力。

---

### 一、项目定位与目标

**项目名称**：**KubeAI Platform**（或自定义名称）

**一句话定位**：为算法工程师提供模型管理、在线推理、离线训练的一站式平台，所有能力通过Web控制台和API开放。

**核心目标**：

* 不依赖MLflow等现成工具，自己实现模型管理模块（注册、版本、元数据）
* 基于K8s开发推理服务网关（支持多版本、灰度、弹性）
* 基于K8s Job开发离线任务调度（支持PyTorch/TensorFlow训练任务）
* 构建统一的可观测性看板（调用链、指标、日志）

**技术栈选择**：

* **后端语言**：Golang（主流云原生平台开发语言，符合你转型方向）
* **Web框架**：Gin / Go-Zero
* **数据库**：PostgreSQL（存储模型元数据、任务记录）
* **缓存**：Redis（任务队列、分布式锁）
* **容器编排**：Kubernetes（核心底座）
* **K8s扩展**：Kubebuilder（开发CRD + Controller）
* **服务通信**：GRPC（内部服务间调用）
* **前端（可选）**：简单Vue/React管理后台（可不做，用Swagger UI代替）

---

### 二、系统架构图（文字版）

**text**

复制

下载

```
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

---

### 三、核心模块详细设计

#### 1. 模型管理服务

**目标**：替代MLflow核心功能，提供模型注册、版本管理、元数据检索能力。

**数据表设计（PostgreSQL）**：

```
-- 模型注册表
CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 模型版本表
CREATE TABLE model_versions (
    id SERIAL PRIMARY KEY,
    model_id INT REFERENCES models(id),
    version VARCHAR(32) NOT NULL,
    storage_path VARCHAR(512),        -- 在MinIO中的路径
    framework VARCHAR(32),             -- pytorch/tensorflow/sklearn
    metrics JSONB,                     -- 性能指标
    status VARCHAR(16),                -- pending/active/archived
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(model_id, version)
);
```

**API设计**：

* `POST /api/v1/models` - 创建模型
* `POST /api/v1/models/{model}/versions` - 上传新版本（接收文件，存储到MinIO）
* `GET /api/v1/models/{model}/versions/{version}` - 获取版本元数据
* `GET /api/v1/models/{model}/versions` - 列出所有版本
* `PUT /api/v1/models/{model}/versions/{version}/status` - 切换活跃版本

**核心实现**：

* 使用MinIO SDK上传/下载模型文件
* 元数据存储在PostgreSQL，版本切换通过更新状态字段实现

#### 2. 推理服务网关

**目标**：用户通过API部署模型为在线推理服务，支持多版本、灰度发布、弹性伸缩。

**K8s CRD设计（使用Kubebuilder）**：

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
    cpu: 2
    memory: 4Gi
    gpu: 1
  autoscaling:
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilization: 70
  canary:
    enabled: true
    version: v1.1
    weight: 10
```

**Controller逻辑**：

* 监听`InferenceService`资源变化
* 根据spec创建Deployment、Service、Ingress
* 实现灰度发布：创建两个Deployment（稳定版和灰度版），通过Service的selector权重或Istio流量管理
* 集成HPA（HorizontalPodAutoscaler）实现弹性伸缩

**API设计**（供算法团队调用）：

* `POST /api/v1/deployments` - 创建推理服务（接收模型版本ID，返回service URL）
* `GET /api/v1/deployments/{id}` - 查询部署状态
* `PUT /api/v1/deployments/{id}/canary` - 调整灰度比例
* `DELETE /api/v1/deployments/{id}` - 下线服务

**关键技术点**：

* 使用client-go操作K8s资源
* 实现自定义Controller（通过Informer监听资源变化）
* 灰度发布可通过修改Service的`spec.selector`或使用Istio的DestinationRule

#### 3. 任务调度服务

**目标**：支持算法团队提交离线训练任务（PyTorch/TensorFlow），自动调度到K8s运行，跟踪状态和日志。

**K8s CRD设计**：

```
apiVersion: ai.platform.io/v1
kind: TrainingJob
metadata:
  name: resnet-training-20250401
spec:
  framework: pytorch
  image: pytorch/pytorch:latest
  command: ["python", "train.py"]
  resources:
    cpu: 4
    memory: 8Gi
    gpu: 2
  volumes:
    - name: code
      source: git://github.com/...   # 可选：代码挂载
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

#### 4. 可观测性模块（可选）

**目标**：平台自身的监控和告警，不依赖Grafana，自己开发一个简单的看板。

**功能点**：

* 记录每个API的请求量、延迟、错误率（通过中间件收集）
* 采集推理服务的QPS、延迟、资源使用率（通过Prometheus client暴露指标）
* 采集训练任务的资源使用情况、任务成功率

**技术实现**：

* 使用Prometheus client-go暴露指标
* 提供简单的Dashboard页面（可集成Grafana，但自己做一个简陋的也行）

---

### 四、代码结构示例（Go项目）

```
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

---

### 六、项目亮点提炼（用于简历和面试）

1. **自研模型管理平台**：不使用MLflow，而是从零实现模型注册、版本控制、元数据存储，理解版本管理背后的设计（表结构、状态机、存储抽象）。
2. **K8s原生扩展**：通过Kubebuilder开发CRD + Controller，实现推理服务的声明式管理，支持灰度发布、弹性伸缩，展现了K8s深度开发能力。
3. **统一任务调度**：支持PyTorch训练任务提交，对接K8s Job，提供统一的任务生命周期管理，具备分布式任务调度经验。
4. **全链路可观测性**：平台内置指标暴露和简单看板，体现运维平台建设能力。
5. **代码质量**：项目结构清晰、有单元测试、有Makefile、有Dockerfile、有部署yaml，符合企业级规范。

**面试时如何讲述**：

> “这个项目是我在转型AI平台开发过程中独立完成的。我没有直接用MLflow、Kubeflow这些现成工具，而是从零实现了一个简化版AI平台，包含模型管理、推理服务、任务调度三大核心模块。推理服务模块基于K8s Operator开发，支持声明式部署、灰度发布、弹性伸缩。模型管理模块自己设计了数据库表和API，对接MinIO实现文件存储。整个项目让我深入理解了AI平台的底层机制，也锻炼了K8s二次开发能力。”

### 七、补充建议

* **选择Go而非Java**：虽然你Java经验丰富，但云原生平台开发主流是Go，而且你在简历中已经强调了Golang转型，用Go完成这个项目能直接证明你的Go能力。
* **代码量预估**：核心代码约2000-3000行，足以作为一个完整的GitHub项目。
* **文档要求**：README.md要包含架构图、快速启动、API文档、部署步骤，让面试官能快速看懂。
* **时间投入**：本项目约需40-60小时（分散到2个月中），完全可以在学习计划框架内完成。

## 1. 数据库表设计（PostgreSQL）

```
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

---

## 2. API 接口定义（RESTful，Swagger 风格）

所有接口挂载在 `/api/v1` 下，返回JSON。

### 2.1 模型管理


| 方法   | 路径           | 描述                         | 请求体示例                                                   | 响应示例                                    |
| ------ | -------------- | ---------------------------- | ------------------------------------------------------------ | ------------------------------------------- |
| POST   | `/models`      | 创建模型                     | `{"name": "bert-classifier", "description": "中文情感分类"}` | `{"id": 1, "name": "bert-classifier", ...}` |
| GET    | `/models`      | 列表模型                     | `?page=1&size=20`                                            | `{"total": 10, "items": [...]}`             |
| GET    | `/models/{id}` | 获取模型详情                 | -                                                            | 同上                                        |
| PUT    | `/models/{id}` | 更新模型描述                 | `{"description": "new desc"}`                                | 更新后模型对象                              |
| DELETE | `/models/{id}` | 删除模型（级联删除所有版本） | -                                                            | `204`                                       |

### 2.2 模型版本管理


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

---

## 3. 核心代码实现（Go + Gin + GORM + MinIO）

我们搭建一个最小可运行版本，包含以上核心API。

### 3.1 项目结构

```
model-manager/
├── cmd/
│   └── main.go                     # 入口
├── internal/
│   ├── handler/                     # HTTP 处理层
│   │   ├── model.go
│   │   └── version.go
│   ├── service/                     # 业务逻辑层
│   │   ├── model.go
│   │   └── version.go
│   ├── repository/                  # 数据访问层
│   │   ├── model.go
│   │   └── version.go
│   ├── model/                       # 数据模型（GORM）
│   │   ├── model.go
│   │   └── version.go
│   └── storage/                     # MinIO 封装
│       └── minio.go
├── pkg/
│   └── utils/                       # 工具函数
├── configs/                         # 配置文件
│   └── config.yaml
├── go.mod
└── Makefile
```

### 3.2 依赖安装

```
go mod init github.com/yourname/model-manager
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/gin-gonic/gin
go get -u github.com/minio/minio-go/v7
go get -u github.com/spf13/viper
```

### 3.3 配置文件 `configs/config.yaml`

```
server:
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: ai_platform
  sslmode: disable

minio:
  endpoint: localhost:9000
  accessKey: minioadmin
  secretKey: minioadmin
  bucket: models
  useSSL: false
```

### 3.4 模型定义（GORM）

`internal/model/model.go`

```
package model

import (
    "time"
    "gorm.io/gorm"
)

type Model struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Name        string         `gorm:"uniqueIndex;size:128;not null" json:"name"`
    Description string         `json:"description"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    Versions    []ModelVersion `gorm:"foreignKey:ModelID" json:"versions,omitempty"`
}
```

`internal/model/version.go`

```
package model

import (
    "time"
    "gorm.io/datatypes"
)

type ModelVersion struct {
    ID          uint            `gorm:"primaryKey" json:"id"`
    ModelID     uint            `gorm:"not null;index" json:"model_id"`
    Version     string          `gorm:"size:32;not null" json:"version"`
    StoragePath string          `gorm:"size:512;not null" json:"storage_path"`
    Framework   string          `gorm:"size:32" json:"framework"`
    Metrics     datatypes.JSON  `json:"metrics"`
    Status      string          `gorm:"size:16;default:active" json:"status"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}
```

> 注意：需要引入 `gorm.io/datatypes`。

### 3.5 MinIO 封装

`internal/storage/minio.go`

```
package storage

import (
    "context"
    "fmt"
    "io"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
    Client *minio.Client
    Bucket string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, err
    }
    // 确保 bucket 存在
    ctx := context.Background()
    exists, err := client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, err
    }
    if !exists {
        err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
        if err != nil {
            return nil, err
        }
    }
    return &MinIOClient{Client: client, Bucket: bucket}, nil
}

// UploadFile 上传文件到指定路径
func (m *MinIOClient) UploadFile(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) error {
    _, err := m.Client.PutObject(ctx, m.Bucket, objectPath, reader, size, minio.PutObjectOptions{ContentType: contentType})
    return err
}

// DownloadFile 获取文件 reader
func (m *MinIOClient) DownloadFile(ctx context.Context, objectPath string) (io.ReadCloser, error) {
    obj, err := m.Client.GetObject(ctx, m.Bucket, objectPath, minio.GetObjectOptions{})
    return obj, err
}
```

### 3.6 Repository 层（以 Model 为例）

`internal/repository/model.go`

```
package repository

import (
    "gorm.io/gorm"
    "github.com/yourname/model-manager/internal/model"
)

type ModelRepository struct {
    db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
    return &ModelRepository{db: db}
}

func (r *ModelRepository) Create(m *model.Model) error {
    return r.db.Create(m).Error
}

func (r *ModelRepository) GetByID(id uint) (*model.Model, error) {
    var m model.Model
    err := r.db.Preload("Versions").First(&m, id).Error
    return &m, err
}

func (r *ModelRepository) List(page, size int) ([]model.Model, int64, error) {
    var models []model.Model
    var total int64
    offset := (page - 1) * size
    r.db.Model(&model.Model{}).Count(&total)
    err := r.db.Offset(offset).Limit(size).Find(&models).Error
    return models, total, err
}

func (r *ModelRepository) Update(m *model.Model) error {
    return r.db.Save(m).Error
}

func (r *ModelRepository) Delete(id uint) error {
    return r.db.Delete(&model.Model{}, id).Error
}
```

类似地，创建 `VersionRepository`（含按model+version查找、更新等）。

### 3.7 Service 层

`internal/service/model.go`

```
package service

import (
    "errors"
    "github.com/yourname/model-manager/internal/model"
    "github.com/yourname/model-manager/internal/repository"
)

type ModelService struct {
    repo *repository.ModelRepository
}

func NewModelService(repo *repository.ModelRepository) *ModelService {
    return &ModelService{repo: repo}
}

func (s *ModelService) Create(name, description string) (*model.Model, error) {
    m := &model.Model{
        Name:        name,
        Description: description,
    }
    if err := s.repo.Create(m); err != nil {
        return nil, err
    }
    return m, nil
}

func (s *ModelService) GetByID(id uint) (*model.Model, error) {
    return s.repo.GetByID(id)
}

func (s *ModelService) List(page, size int) ([]model.Model, int64, error) {
    return s.repo.List(page, size)
}

func (s *ModelService) Update(id uint, description string) (*model.Model, error) {
    m, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    m.Description = description
    if err := s.repo.Update(m); err != nil {
        return nil, err
    }
    return m, nil
}

func (s *ModelService) Delete(id uint) error {
    return s.repo.Delete(id)
}
```

### 3.8 Handler 层

`internal/handler/model.go`

```
package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/yourname/model-manager/internal/service"
)

type ModelHandler struct {
    svc *service.ModelService
}

func NewModelHandler(svc *service.ModelService) *ModelHandler {
    return &ModelHandler{svc: svc}
}

func (h *ModelHandler) Create(c *gin.Context) {
    var req struct {
        Name        string `json:"name" binding:"required"`
        Description string `json:"description"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    m, err := h.svc.Create(req.Name, req.Description)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, m)
}

func (h *ModelHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    m, err := h.svc.GetByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
        return
    }
    c.JSON(http.StatusOK, m)
}

func (h *ModelHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
    models, total, err := h.svc.List(page, size)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "total": total,
        "items": models,
    })
}

func (h *ModelHandler) Update(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    var req struct {
        Description string `json:"description"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    m, err := h.svc.Update(uint(id), req.Description)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, m)
}

func (h *ModelHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    if err := h.svc.Delete(uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}
```

**版本上传 Handler** 较为复杂，涉及 multipart 解析和 MinIO 上传，下面给出关键代码：

`internal/handler/version.go` (片段)

```
func (h *VersionHandler) Upload(c *gin.Context) {
    modelID, _ := strconv.ParseUint(c.Param("modelId"), 10, 32)
    // 1. 解析表单
    version := c.PostForm("version")
    framework := c.PostForm("framework")
    metricsStr := c.PostForm("metrics") // JSON string
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
        return
    }
    defer file.Close()

    // 2. 验证 model 存在
    model, err := h.modelSvc.GetByID(uint(modelID))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
        return
    }

    // 3. 生成存储路径
    storagePath := fmt.Sprintf("models/%s/%s/model.bin", model.Name, version)

    // 4. 上传到 MinIO
    ctx := context.Background()
    if err := h.minioClient.UploadFile(ctx, storagePath, file, header.Size, header.Header.Get("Content-Type")); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
        return
    }

    // 5. 创建版本记录
    var metrics datatypes.JSON
    if metricsStr != "" {
        metrics = datatypes.JSON([]byte(metricsStr))
    }
    v := &model.ModelVersion{
        ModelID:     uint(modelID),
        Version:     version,
        StoragePath: storagePath,
        Framework:   framework,
        Metrics:     metrics,
        Status:      "active",
    }
    if err := h.versionRepo.Create(v); err != nil {
        // 回滚 MinIO 上传（可选）
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, v)
}
```

### 3.9 主程序入口 `cmd/main.go`

```
package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/spf13/viper"
    "github.com/yourname/model-manager/internal/handler"
    "github.com/yourname/model-manager/internal/repository"
    "github.com/yourname/model-manager/internal/service"
    "github.com/yourname/model-manager/internal/storage"
    "github.com/yourname/model-manager/internal/model"
)

func initConfig() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("read config failed: %v", err)
    }
}

func main() {
    initConfig()

    // 连接数据库
    dsn := viper.GetString("database.dsn") // 或者拼接
    dsn = "host=localhost user=postgres password=postgres dbname=ai_platform port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("db connect failed: %v", err)
    }
    // 自动迁移
    db.AutoMigrate(&model.Model{}, &model.ModelVersion{})

    // 初始化 MinIO
    minioClient, err := storage.NewMinIOClient(
        viper.GetString("minio.endpoint"),
        viper.GetString("minio.accessKey"),
        viper.GetString("minio.secretKey"),
        viper.GetString("minio.bucket"),
        viper.GetBool("minio.useSSL"),
    )
    if err != nil {
        log.Fatalf("minio init failed: %v", err)
    }

    // 初始化 repository
    modelRepo := repository.NewModelRepository(db)
    versionRepo := repository.NewVersionRepository(db)

    // 初始化 service
    modelSvc := service.NewModelService(modelRepo)
    versionSvc := service.NewVersionService(versionRepo, modelSvc, minioClient)

    // 初始化 handler
    modelHandler := handler.NewModelHandler(modelSvc)
    versionHandler := handler.NewVersionHandler(versionSvc)

    // 路由
    r := gin.Default()
    api := r.Group("/api/v1")
    {
        // models
        api.POST("/models", modelHandler.Create)
        api.GET("/models", modelHandler.List)
        api.GET("/models/:id", modelHandler.Get)
        api.PUT("/models/:id", modelHandler.Update)
        api.DELETE("/models/:id", modelHandler.Delete)

        // versions
        api.POST("/models/:modelId/versions", versionHandler.Upload)
        api.GET("/models/:modelId/versions", versionHandler.List)
        api.GET("/models/:modelId/versions/:version", versionHandler.Get)
        api.PUT("/models/:modelId/versions/:version/status", versionHandler.UpdateStatus)
        api.GET("/models/:modelId/versions/:version/download", versionHandler.Download)
    }

    port := viper.GetString("server.port")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}
```

---

## 4. 快速启动步骤

1. **安装依赖**：PostgreSQL、MinIO（可用Docker快速启动）

   ```
   docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres
   docker run -d --name minio -p 9000:9000 -p 9001:9001 -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin minio/minio server /data --console-address ":9001"
   ```
2. **创建数据库**：

   ```
   docker exec -it postgres psql -U postgres -c "CREATE DATABASE ai_platform;"
   ```
3. **配置**：修改 `configs/config.yaml` 中的数据库和MinIO地址。
4. **运行**：
   **bash**

   复制

   下载

   ```
   go run cmd/main.go
   ```
5. **测试API**：使用 Postman 或 curl 测试创建模型、上传文件等接口。

---

## 5. 下一步建议

有了模型管理服务的基础，你可以：

* 扩展 `metrics` 的解析，支持更丰富的指标展示
* 增加标签管理功能（增删改查）
* 为版本添加 `download_count` 统计
* 使用 JWT 增加认证（可选）

当你完成这个模块，就可以继续开发**推理服务网关**和**任务调度**模块，形成完整的AI平台。后续我可以继续为你提供那些模块的详细代码实现。
