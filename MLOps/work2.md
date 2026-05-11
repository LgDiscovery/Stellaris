## 任务调度服务详细设计

### 1. 核心功能

* 提交训练任务（指定镜像、命令、资源、环境变量）
* 查询任务状态（pending/running/succeeded/failed）
* 获取任务日志（实时或历史）
* 支持多种框架（PyTorch、TensorFlow、自定义）
* 可选：支持分布式训练（单机多卡、多机多卡）

### 2. 技术选型

* **K8s Job**：执行一次性任务
* **client-go**：与K8s API交互
* **Pod日志收集**：通过K8s API获取或对接日志系统（简化版直接通过API获取）
* **数据库**：PostgreSQL存储任务元数据（与模型管理共用）

### 3. 数据库表设计

```
-- 训练任务表
CREATE TABLE training_jobs (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,                 -- 任务名称，唯一
    model_id    BIGINT REFERENCES models(id),          -- 可选：关联的模型
    framework   VARCHAR(32),                           -- pytorch / tensorflow
    image       VARCHAR(256) NOT NULL,                 -- 训练镜像
    command     TEXT,                                   -- 启动命令，如 python train.py
    resources   JSONB NOT NULL,                        -- 资源需求: {"cpu": "2", "memory": "4Gi", "gpu": 1}
    env         JSONB,                                  -- 环境变量: [{"name":"EPOCHS","value":"100"}]
    volumes     JSONB,                                  -- 卷挂载: [{"name":"code","source":"git://..."}]
    status      VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending/running/succeeded/failed
    k8s_name    VARCHAR(256),                           -- 对应的K8s Job名称
    started_at  TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_training_jobs_status ON training_jobs(status);
CREATE INDEX idx_training_jobs_model_id ON training_jobs(model_id);
```

**说明**：

* `resources` 存储资源需求，支持CPU、内存、GPU等
* `env` 存储环境变量列表
* `volumes` 可存储数据卷挂载信息（简化版可先不支持，直接使用镜像内置代码和数据）

### 4. API接口设计


| 方法   | 路径                     | 描述                          | 请求体示例                       | 响应     |
| ------ | ------------------------ | ----------------------------- | -------------------------------- | -------- |
| POST   | `/api/v1/jobs`           | 提交训练任务                  | 见下方                           | 任务对象 |
| GET    | `/api/v1/jobs`           | 列表任务                      | `?page=1&size=20&status=running` | 分页列表 |
| GET    | `/api/v1/jobs/{id}`      | 获取任务详情                  | -                                | 任务对象 |
| GET    | `/api/v1/jobs/{id}/logs` | 获取任务日志                  | `?tail=1000`                     | 日志文本 |
| DELETE | `/api/v1/jobs/{id}`      | 删除任务（停止Job并删除记录） | -                                | 204      |

**提交任务请求体示例**：

```
{
    "name": "bert-training-20250401",
    "model_id": 1,
    "framework": "pytorch",
    "image": "pytorch/pytorch:2.0.0-cuda11.7-cudnn8-runtime",
    "command": "python /workspace/train.py --epochs 10",
    "resources": {
        "cpu": "4",
        "memory": "8Gi",
        "gpu": 1
    },
    "env": [
        {"name": "EPOCHS", "value": "100"},
        {"name": "BATCH_SIZE", "value": "32"}
    ]
}
```

### 5. 核心代码实现

#### 5.1 项目结构（新增模块）

在原有 `model-manager` 基础上增加：

```
internal/
├── job/
│   ├── handler/          # HTTP处理
│   ├── service/          # 业务逻辑
│   ├── repository/       # 数据库操作
│   ├── model/            # GORM模型
│   └── k8s/              # K8s客户端封装
```

#### 5.2 GORM模型定义

`internal/job/model/job.go`

```
package model

import (
    "time"
    "gorm.io/datatypes"
)

type TrainingJob struct {
    ID         uint           `gorm:"primaryKey" json:"id"`
    Name       string         `gorm:"size:128;not null;uniqueIndex" json:"name"`
    ModelID    *uint          `json:"model_id,omitempty"`
    Framework  string         `gorm:"size:32" json:"framework"`
    Image      string         `gorm:"size:256;not null" json:"image"`
    Command    string         `json:"command"`
    Resources  datatypes.JSON `gorm:"not null" json:"resources"`
    Env        datatypes.JSON `json:"env"`
    Status     string         `gorm:"size:32;default:pending;index" json:"status"`
    K8sName    string         `gorm:"size:256" json:"k8s_name"`
    StartedAt  *time.Time     `json:"started_at,omitempty"`
    FinishedAt *time.Time     `json:"finished_at,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
}
```

#### 5.3 K8s客户端封装

`internal/job/k8s/client.go`

```
package k8s

import (
    "context"
    "fmt"
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "os"
)

type K8sClient struct {
    clientset *kubernetes.Clientset
    namespace string
}

func NewK8sClient() (*K8sClient, error) {
    var config *rest.Config
    var err error

    // 优先使用 in-cluster 配置（如果运行在Pod内）
    config, err = rest.InClusterConfig()
    if err != nil {
        // 回退到 kubeconfig
        kubeconfig := os.Getenv("KUBECONFIG")
        if kubeconfig == "" {
            kubeconfig = clientcmd.RecommendedHomeFile
        }
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            return nil, err
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    namespace := os.Getenv("K8S_NAMESPACE")
    if namespace == "" {
        namespace = "default"
    }

    return &K8sClient{
        clientset: clientset,
        namespace: namespace,
    }, nil
}

// CreateJob 创建K8s Job，返回job名称
func (c *K8sClient) CreateJob(ctx context.Context, name, image, command string, resources map[string]interface{}, envVars []map[string]string) (string, error) {
    // 解析资源需求
    var cpu, memory string
    var gpu int64
    if resources != nil {
        if v, ok := resources["cpu"].(string); ok {
            cpu = v
        }
        if v, ok := resources["memory"].(string); ok {
            memory = v
        }
        if v, ok := resources["gpu"].(float64); ok {
            gpu = int64(v)
        }
    }

    // 构建环境变量
    var env []corev1.EnvVar
    for _, e := range envVars {
        env = append(env, corev1.EnvVar{
            Name:  e["name"],
            Value: e["value"],
        })
    }

    // 构建容器资源需求
    resourceRequirements := corev1.ResourceRequirements{
        Requests: corev1.ResourceList{},
        Limits:   corev1.ResourceList{},
    }
    if cpu != "" {
        resourceRequirements.Requests[corev1.ResourceCPU] = resource.MustParse(cpu)
        resourceRequirements.Limits[corev1.ResourceCPU] = resource.MustParse(cpu)
    }
    if memory != "" {
        resourceRequirements.Requests[corev1.ResourceMemory] = resource.MustParse(memory)
        resourceRequirements.Limits[corev1.ResourceMemory] = resource.MustParse(memory)
    }
    if gpu > 0 {
        resourceRequirements.Requests["nvidia.com/gpu"] = *resource.NewQuantity(gpu, resource.DecimalSI)
        resourceRequirements.Limits["nvidia.com/gpu"] = *resource.NewQuantity(gpu, resource.DecimalSI)
    }

    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: c.namespace,
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{
                        {
                            Name:      "trainer",
                            Image:     image,
                            Command:   []string{"/bin/sh", "-c"},
                            Args:      []string{command},
                            Env:       env,
                            Resources: resourceRequirements,
                        },
                    },
                },
            },
        },
    }

    _, err := c.clientset.BatchV1().Jobs(c.namespace).Create(ctx, job, metav1.CreateOptions{})
    if err != nil {
        return "", err
    }
    return job.Name, nil
}

// GetJobStatus 获取Job状态
func (c *K8sClient) GetJobStatus(ctx context.Context, jobName string) (string, error) {
    job, err := c.clientset.BatchV1().Jobs(c.namespace).Get(ctx, jobName, metav1.GetOptions{})
    if err != nil {
        return "", err
    }
    if job.Status.Succeeded > 0 {
        return "succeeded", nil
    } else if job.Status.Failed > 0 {
        return "failed", nil
    } else if job.Status.Active > 0 {
        return "running", nil
    } else {
        return "pending", nil
    }
}

// DeleteJob 删除Job
func (c *K8sClient) DeleteJob(ctx context.Context, jobName string) error {
    // 删除Job会级联删除关联的Pod
    return c.clientset.BatchV1().Jobs(c.namespace).Delete(ctx, jobName, metav1.DeleteOptions{})
}

// GetJobLogs 获取Job的Pod日志（取第一个Pod）
func (c *K8sClient) GetJobLogs(ctx context.Context, jobName string, tailLines int64) (string, error) {
    // 找到该Job的Pod
    pods, err := c.clientset.CoreV1().Pods(c.namespace).List(ctx, metav1.ListOptions{
        LabelSelector: fmt.Sprintf("job-name=%s", jobName),
    })
    if err != nil {
        return "", err
    }
    if len(pods.Items) == 0 {
        return "", fmt.Errorf("no pod found for job %s", jobName)
    }
    // 取第一个Pod的日志
    pod := pods.Items[0]
    opts := corev1.PodLogOptions{}
    if tailLines > 0 {
        opts.TailLines = &tailLines
    }
    req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(pod.Name, &opts)
    logs, err := req.Stream(ctx)
    if err != nil {
        return "", err
    }
    defer logs.Close()
    buf := new(bytes.Buffer)
    _, err = io.Copy(buf, logs)
    return buf.String(), err
}
```

#### 5.4 Repository层

`internal/job/repository/job.go`

```
package repository

import (
    "gorm.io/gorm"
    "github.com/yourname/model-manager/internal/job/model"
)

type JobRepository struct {
    db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
    return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *model.TrainingJob) error {
    return r.db.Create(job).Error
}

func (r *JobRepository) GetByID(id uint) (*model.TrainingJob, error) {
    var job model.TrainingJob
    err := r.db.First(&job, id).Error
    return &job, err
}

func (r *JobRepository) GetByName(name string) (*model.TrainingJob, error) {
    var job model.TrainingJob
    err := r.db.Where("name = ?", name).First(&job).Error
    return &job, err
}

func (r *JobRepository) List(page, size int, status string) ([]model.TrainingJob, int64, error) {
    var jobs []model.TrainingJob
    var total int64
    query := r.db.Model(&model.TrainingJob{})
    if status != "" {
        query = query.Where("status = ?", status)
    }
    query.Count(&total)
    offset := (page - 1) * size
    err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&jobs).Error
    return jobs, total, err
}

func (r *JobRepository) Update(job *model.TrainingJob) error {
    return r.db.Save(job).Error
}

func (r *JobRepository) Delete(id uint) error {
    return r.db.Delete(&model.TrainingJob{}, id).Error
}
```

#### 5.5 Service层

`internal/job/service/job.go`

```
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/yourname/model-manager/internal/job/k8s"
    "github.com/yourname/model-manager/internal/job/model"
    "github.com/yourname/model-manager/internal/job/repository"
    "gorm.io/datatypes"
)

type JobService struct {
    repo     *repository.JobRepository
    k8sClient *k8s.K8sClient
}

func NewJobService(repo *repository.JobRepository, k8sClient *k8s.K8sClient) *JobService {
    return &JobService{
        repo:     repo,
        k8sClient: k8sClient,
    }
}

type CreateJobRequest struct {
    Name      string                 `json:"name"`
    ModelID   *uint                  `json:"model_id,omitempty"`
    Framework string                 `json:"framework"`
    Image     string                 `json:"image"`
    Command   string                 `json:"command"`
    Resources map[string]interface{} `json:"resources"`
    Env       []map[string]string    `json:"env"`
}

func (s *JobService) CreateJob(ctx context.Context, req CreateJobRequest) (*model.TrainingJob, error) {
    // 检查名称唯一性
    existing, _ := s.repo.GetByName(req.Name)
    if existing != nil {
        return nil, fmt.Errorf("job with name %s already exists", req.Name)
    }

    // 序列化 Resources 和 Env
    resourcesJSON, err := json.Marshal(req.Resources)
    if err != nil {
        return nil, err
    }
    envJSON, err := json.Marshal(req.Env)
    if err != nil {
        return nil, err
    }

    // 生成K8s Job名称（使用任务名+随机后缀，避免冲突）
    k8sJobName := fmt.Sprintf("%s-%s", req.Name, uuid.New().String()[:8])

    job := &model.TrainingJob{
        Name:       req.Name,
        ModelID:    req.ModelID,
        Framework:  req.Framework,
        Image:      req.Image,
        Command:    req.Command,
        Resources:  datatypes.JSON(resourcesJSON),
        Env:        datatypes.JSON(envJSON),
        Status:     "pending",
        K8sName:    k8sJobName,
    }

    // 保存到数据库
    if err := s.repo.Create(job); err != nil {
        return nil, err
    }

    // 异步创建K8s Job（避免阻塞API响应）
    go func() {
        // 注意：需要传递新的context，避免原请求取消导致任务中止
        ctx := context.Background()
        _, err := s.k8sClient.CreateJob(ctx, k8sJobName, req.Image, req.Command, req.Resources, req.Env)
        if err != nil {
            // 更新状态为失败
            job.Status = "failed"
            now := time.Now()
            job.FinishedAt = &now
            s.repo.Update(job)
            return
        }
        // 更新状态为running和开始时间
        job.Status = "running"
        now := time.Now()
        job.StartedAt = &now
        s.repo.Update(job)

        // 启动一个goroutine定期同步状态（可后续实现）
        // 也可以依赖外部状态同步服务
    }()

    return job, nil
}

func (s *JobService) GetJob(id uint) (*model.TrainingJob, error) {
    return s.repo.GetByID(id)
}

func (s *JobService) ListJobs(page, size int, status string) ([]model.TrainingJob, int64, error) {
    return s.repo.List(page, size, status)
}

func (s *JobService) DeleteJob(ctx context.Context, id uint) error {
    job, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    // 如果Job还在K8s中，先删除
    if job.K8sName != "" && (job.Status == "running" || job.Status == "pending") {
        _ = s.k8sClient.DeleteJob(ctx, job.K8sName)
    }
    return s.repo.Delete(id)
}

// SyncJobStatus 同步K8s Job状态到数据库（用于定时任务）
func (s *JobService) SyncJobStatus(ctx context.Context, id uint) error {
    job, err := s.repo.GetByID(id)
    if err != nil {
        return err
    }
    if job.Status == "succeeded" || job.Status == "failed" {
        return nil // 终态不再同步
    }
    k8sStatus, err := s.k8sClient.GetJobStatus(ctx, job.K8sName)
    if err != nil {
        return err
    }
    if k8sStatus != job.Status {
        job.Status = k8sStatus
        if k8sStatus == "succeeded" || k8sStatus == "failed" {
            now := time.Now()
            job.FinishedAt = &now
        }
        return s.repo.Update(job)
    }
    return nil
}

func (s *JobService) GetJobLogs(ctx context.Context, id uint, tailLines int64) (string, error) {
    job, err := s.repo.GetByID(id)
    if err != nil {
        return "", err
    }
    if job.K8sName == "" {
        return "", fmt.Errorf("job has no k8s job")
    }
    return s.k8sClient.GetJobLogs(ctx, job.K8sName, tailLines)
}
```

#### 5.6 Handler层

`internal/job/handler/job.go`

```
package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/yourname/model-manager/internal/job/service"
)

type JobHandler struct {
    svc *service.JobService
}

func NewJobHandler(svc *service.JobService) *JobHandler {
    return &JobHandler{svc: svc}
}

func (h *JobHandler) Create(c *gin.Context) {
    var req service.CreateJobRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    job, err := h.svc.CreateJob(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, job)
}

func (h *JobHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    job, err := h.svc.GetJob(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
        return
    }
    c.JSON(http.StatusOK, job)
}

func (h *JobHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
    status := c.Query("status")
    jobs, total, err := h.svc.ListJobs(page, size, status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "total": total,
        "items": jobs,
    })
}

func (h *JobHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := h.svc.DeleteJob(c.Request.Context(), uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}

func (h *JobHandler) GetLogs(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    tail := c.Query("tail")
    var tailLines int64
    if tail != "" {
        tailLines, _ = strconv.ParseInt(tail, 10, 64)
    }
    logs, err := h.svc.GetJobLogs(c.Request.Context(), uint(id), tailLines)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.String(http.StatusOK, logs)
}
```

#### 5.7 主程序集成

在 `cmd/main.go` 中增加初始化任务调度服务的代码：

```
// 初始化 Job 相关
jobRepo := jobRepo.NewJobRepository(db)
k8sClient, err := k8s.NewK8sClient()
if err != nil {
    log.Fatalf("k8s client init failed: %v", err)
}
jobSvc := jobService.NewJobService(jobRepo, k8sClient)
jobHandler := jobHandler.NewJobHandler(jobSvc)

// 路由
api := r.Group("/api/v1")
{
    // ... 原有模型管理路由 ...
  
    // jobs
    api.POST("/jobs", jobHandler.Create)
    api.GET("/jobs", jobHandler.List)
    api.GET("/jobs/:id", jobHandler.Get)
    api.DELETE("/jobs/:id", jobHandler.Delete)
    api.GET("/jobs/:id/logs", jobHandler.GetLogs)
}
```

### 6. 状态同步机制

由于Job状态可能在K8s中变化，建议增加一个后台定时任务（使用cron）定期同步所有非终态任务的状态。可以在主程序中启动一个goroutine定期执行：

```
// 启动状态同步协程
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        ctx := context.Background()
        // 获取所有未完成的任务
        jobs, _, _ := jobSvc.ListJobs(1, 1000, "pending,running")
        for _, job := range jobs {
            jobSvc.SyncJobStatus(ctx, job.ID)
        }
    }
}()
```

### 7. 测试与验证

1. **提交任务**：使用Postman POST `/api/v1/jobs` 提交任务
2. **查看状态**：GET `/api/v1/jobs/{id}` 确认状态变为running
3. **查看日志**：GET `/api/v1/jobs/{id}/logs?tail=100`
4. **删除任务**：DELETE `/api/v1/jobs/{id}`

---

## 下一步建议

完成了**任务调度服务**后，整个AI平台的核心模块已经具备：

* 模型管理（自研替代MLflow）
* 推理服务网关（基于K8s Operator）
* 训练任务调度（基于K8s Job）

接下来你可以：

1. **实现推理服务网关**（基于K8s Operator）
2. **完善前端界面**（可选，用Vue/React搭建管理控制台）
3. **增加认证授权**（JWT）
4. **完善文档和部署脚本**

如果你需要，我可以继续提供**推理服务网关**的详细代码实现，包括CRD定义、Controller逻辑、灰度发布、弹性伸缩等。

继续

已思考（用时 4 秒）

我们正在细化AI平台项目，已完成了模型管理服务的详细设计、代码实现。现在需要继续下一个核心模块：**任务调度服务**。用户需要详细的代码实现和说明。我将提供任务调度服务的完整设计，包括CRD定义、Controller实现、API接口、与K8s交互的代码示例等。同时保持项目整体一致性，使用Go + Kubebuilder + client-go等。

我将按照以下结构组织回答：

1. 任务调度服务的目标与功能
2. CRD定义（TrainingJob）
3. Controller核心逻辑（创建Job、状态跟踪、日志收集）
4. API接口（提交任务、查询状态、获取日志）
5. 关键代码片段
6. 与模型管理服务的集成
7. 部署与测试建议

确保内容详细且可直接落地。

好的，我们继续实现第二个核心模块：**任务调度服务**。这个模块将允许算法团队提交离线训练任务（如PyTorch、TensorFlow），并在Kubernetes上自动运行，同时提供任务状态跟踪和日志查询能力。

我们将使用 **Kubebuilder** 开发自定义资源（CRD）和控制器，这是AI平台开发的典型模式。

---

## 一、任务调度服务目标与功能

### 目标

* 提供统一的离线训练任务提交与生命周期管理
* 支持多种深度学习框架（PyTorch、TensorFlow）的分布式训练
* 自动调度到K8s集群，管理任务资源（CPU/GPU）
* 提供REST API供算法团队使用

### 核心功能

1. **提交训练任务**：接收任务定义（框架、镜像、命令、资源等），创建CRD实例
2. **任务调度**：Controller监听到CRD创建，创建对应的K8s Job（或Pod）
3. **状态跟踪**：实时更新CRD的status，反映任务运行阶段（Pending、Running、Succeeded、Failed）
4. **日志获取**：提供API获取任务的实时日志
5. **任务清理**：支持手动删除任务，自动清理K8s资源

### 技术选型

* **Kubebuilder**：生成CRD和Controller脚手架
* **client-go**：操作K8s资源
* **Gin**：提供REST API
* **PostgreSQL**：存储任务元数据（可选，也可以直接用CRD作为唯一数据源，但为性能建议加数据库）
* **K8s API**：获取Pod日志

---

## 二、CRD 定义（TrainingJob）

### 2.1 设计目标

* 描述一个训练任务的完整规格
* 支持单机或分布式（如PyTorch的`elastic`或`torchrun`）
* 支持环境变量、存储挂载、镜像拉取策略等

### 2.2 示例 YAML

```
apiVersion: ai.platform.io/v1
kind: TrainingJob
metadata:
  name: resnet-training-20250401
  namespace: default
spec:
  framework: pytorch
  image: pytorch/pytorch:2.0.1-cuda11.7-cudnn8-runtime
  command: ["python", "-m", "torch.distributed.run", "--nproc_per_node=2", "train.py"]
  resources:
    cpu: 4
    memory: 8Gi
    gpu: 2
  env:
    - name: EPOCHS
      value: "100"
    - name: BATCH_SIZE
      value: "64"
  volumes:
    - name: code
      source: git://github.com/your/repo.git  # 简化处理，可以用PVC
      mountPath: /workspace
    - name: data
      source: pvc://training-data
      mountPath: /data
  backoffLimit: 3
  ttlSecondsAfterFinished: 3600
```

### 2.3 定义结构（Go）

在Kubebuilder项目中，定义API版本和Kind。通常使用`group: ai.platform.io`，`version: v1`。

```
// api/v1/trainingjob_types.go

package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrainingJobSpec defines the desired state of TrainingJob
type TrainingJobSpec struct {
    // Framework indicates the deep learning framework (pytorch, tensorflow, etc.)
    Framework string `json:"framework"`
    // Image is the container image to use for training
    Image string `json:"image"`
    // Command is the entrypoint command (overrides Docker's CMD)
    Command []string `json:"command,omitempty"`
    // Resources defines compute resources (CPU, memory, GPU)
    Resources ResourceRequirements `json:"resources,omitempty"`
    // Env defines environment variables
    Env []EnvVar `json:"env,omitempty"`
    // Volumes defines volumes to mount (code, data)
    Volumes []VolumeMount `json:"volumes,omitempty"`
    // BackoffLimit defines number of retries before marking job failed
    BackoffLimit int32 `json:"backoffLimit,omitempty"`
    // TTLSecondsAfterFinished cleans up job after finished
    TTLSecondsAfterFinished int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// ResourceRequirements defines resource requests and limits
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

// VolumeMount represents a volume mount
type VolumeMount struct {
    Name      string `json:"name"`
    Source    string `json:"source"`   // e.g., "pvc://my-pvc", "git://url"
    MountPath string `json:"mountPath"`
}

// TrainingJobStatus defines the observed state of TrainingJob
type TrainingJobStatus struct {
    // Conditions represent the latest available observations of the job's state
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    // StartTime is the time when the job started
    StartTime *metav1.Time `json:"startTime,omitempty"`
    // CompletionTime is the time when the job completed (success or failure)
    CompletionTime *metav1.Time `json:"completionTime,omitempty"`
    // JobName is the name of the underlying Kubernetes Job
    JobName string `json:"jobName,omitempty"`
    // Active pods count
    Active int32 `json:"active,omitempty"`
    // Succeeded pods count
    Succeeded int32 `json:"succeeded,omitempty"`
    // Failed pods count
    Failed int32 `json:"failed,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[-1].reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// TrainingJob is the Schema for the trainingjobs API
type TrainingJob struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   TrainingJobSpec   `json:"spec,omitempty"`
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
```

---

## 三、Controller 实现（核心逻辑）

Controller负责将TrainingJob转换为K8s Job，并持续同步状态。

### 3.1 生成脚手架

使用Kubebuilder初始化项目后，执行：

```
kubebuilder create api --group ai.platform.io --version v1 --kind TrainingJob --resource --controller
```

### 3.2 Controller 核心逻辑

在 `controllers/trainingjob_controller.go` 中实现Reconcile循环。

**Reconcile流程**：

1. 获取TrainingJob对象
2. 检查是否存在对应的K8s Job
   * 若不存在，创建Job
   * 若存在，更新状态（同步Job的状态到TrainingJob）
3. 若TrainingJob被删除，清理关联资源（如果设置了OwnerReference，Job会自动被GC）
4. 更新TrainingJob的Status

**关键代码片段**：

```
// controllers/trainingjob_controller.go (精简版)

func (r *TrainingJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("trainingjob", req.NamespacedName)

    // 1. 获取 TrainingJob
    var job crdv1.TrainingJob
    if err := r.Get(ctx, req.NamespacedName, &job); err != nil {
        // 可能已被删除
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. 处理删除（如果设置了 Finalizer）
    if !job.DeletionTimestamp.IsZero() {
        // 清理外部资源（可选）
        return ctrl.Result{}, nil
    }

    // 3. 检查底层 K8s Job 是否存在
    var k8sJob batchv1.Job
    jobName := job.Status.JobName
    if jobName == "" {
        // 创建新 Job
        k8sJob, err := r.constructJob(&job)
        if err != nil {
            r.updateStatusCondition(&job, "Failed", "JobConstructionFailed", err.Error())
            return ctrl.Result{}, err
        }
        if err := r.Create(ctx, k8sJob); err != nil {
            r.updateStatusCondition(&job, "Failed", "JobCreationFailed", err.Error())
            return ctrl.Result{}, err
        }
        // 更新 TrainingJob 状态
        job.Status.JobName = k8sJob.Name
        job.Status.StartTime = &metav1.Time{Time: time.Now()}
        r.updateStatusCondition(&job, "Running", "JobCreated", "Job created successfully")
        if err := r.Status().Update(ctx, &job); err != nil {
            return ctrl.Result{}, err
        }
        return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
    }

    // 4. 获取已存在的 Job
    if err := r.Get(ctx, types.NamespacedName{Name: jobName, Namespace: job.Namespace}, &k8sJob); err != nil {
        // Job 丢失，更新状态为失败
        r.updateStatusCondition(&job, "Failed", "JobNotFound", "Underlying job not found")
        return ctrl.Result{}, err
    }

    // 5. 同步状态
    r.syncStatusFromJob(&job, &k8sJob)

    // 6. 更新 TrainingJob 状态
    if err := r.Status().Update(ctx, &job); err != nil {
        return ctrl.Result{}, err
    }

    // 7. 如果 Job 已完成，根据 TTL 决定是否删除（可选）
    if isJobFinished(&k8sJob) {
        if job.Spec.TTLSecondsAfterFinished > 0 {
            // 可以设置删除 TrainingJob 的定时器，或者直接删除 Job 但保留 CRD
            // 这里简单起见，只记录完成
        }
        return ctrl.Result{}, nil
    }

    // 定期同步
    return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}
```

**构造Job函数**：

```
func (r *TrainingJobReconciler) constructJob(tj *crdv1.TrainingJob) (*batchv1.Job, error) {
    // 生成唯一的Job名称（可以基于TrainingJob名称加随机后缀）
    jobName := tj.Name + "-" + randString(5)

    // 构建容器
    container := corev1.Container{
        Name:    "trainer",
        Image:   tj.Spec.Image,
        Command: tj.Spec.Command,
        Env:     convertEnvVars(tj.Spec.Env),
        Resources: corev1.ResourceRequirements{
            Requests: corev1.ResourceList{},
            Limits:   corev1.ResourceList{},
        },
    }
    // 设置资源
    if cpu := tj.Spec.Resources.CPU; cpu != "" {
        container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(cpu)
        container.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(cpu)
    }
    if memory := tj.Spec.Resources.Memory; memory != "" {
        container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(memory)
        container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(memory)
    }
    if gpu := tj.Spec.Resources.GPU; gpu != "" {
        container.Resources.Limits["nvidia.com/gpu"] = resource.MustParse(gpu)
        // 通常GPU只设limits
    }

    // 构建Volume和VolumeMounts
    volumes, mounts := buildVolumes(tj.Spec.Volumes)
    container.VolumeMounts = mounts

    // 构建Pod模板
    podSpec := corev1.PodSpec{
        Containers:    []corev1.Container{container},
        RestartPolicy: corev1.RestartPolicyNever,
        Volumes:       volumes,
    }

    // 设置backoffLimit
    backoffLimit := int32(3)
    if tj.Spec.BackoffLimit != 0 {
        backoffLimit = tj.Spec.BackoffLimit
    }

    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: tj.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(tj, crdv1.GroupVersion.WithKind("TrainingJob")),
            },
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "trainingjob": tj.Name,
                    },
                },
                Spec: podSpec,
            },
            BackoffLimit: &backoffLimit,
            TTLSecondsAfterFinished: tj.Spec.TTLSecondsAfterFinished,
        },
    }
    return job, nil
}
```

**状态同步**：

```
func (r *TrainingJobReconciler) syncStatusFromJob(tj *crdv1.TrainingJob, job *batchv1.Job) {
    // 更新 Job 的 active/succeeded/failed 计数
    tj.Status.Active = job.Status.Active
    tj.Status.Succeeded = job.Status.Succeeded
    tj.Status.Failed = job.Status.Failed

    // 判断整体状态
    if job.Status.Succeeded > 0 {
        r.updateStatusCondition(tj, "Complete", "JobSucceeded", "Training completed successfully")
        tj.Status.CompletionTime = &metav1.Time{Time: time.Now()}
    } else if job.Status.Failed > 0 {
        r.updateStatusCondition(tj, "Failed", "JobFailed", "Training failed")
        tj.Status.CompletionTime = &metav1.Time{Time: time.Now()}
    } else if job.Status.Active > 0 {
        r.updateStatusCondition(tj, "Running", "JobRunning", "Training is running")
    } else {
        r.updateStatusCondition(tj, "Pending", "JobPending", "Waiting for resources")
    }
}
```

---

## 四、REST API 接口

我们提供HTTP API供算法团队使用，后端可以是一个独立的Gin服务，通过client-go与K8s交互（也可直接调用Controller的Reconcile，但通常通过创建CRD对象）。

### 4.1 API 定义


| 方法   | 路径                                | 描述         | 请求体                |
| ------ | ----------------------------------- | ------------ | --------------------- |
| POST   | `/api/v1/training-jobs`             | 提交训练任务 | TrainingJob spec JSON |
| GET    | `/api/v1/training-jobs`             | 列出所有任务 | 分页参数              |
| GET    | `/api/v1/training-jobs/{name}`      | 获取任务详情 | -                     |
| GET    | `/api/v1/training-jobs/{name}/logs` | 获取任务日志 | `?follow=true` (可选) |
| DELETE | `/api/v1/training-jobs/{name}`      | 删除任务     | -                     |

### 4.2 实现（Gin + client-go）

```
// internal/api/training.go

package api

import (
    "context"
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    crdv1 "github.com/yourname/ai-platform/api/v1"
)

type TrainingJobHandler struct {
    k8sClient client.Client
}

func NewTrainingJobHandler(k8sClient client.Client) *TrainingJobHandler {
    return &TrainingJobHandler{k8sClient: k8sClient}
}

func (h *TrainingJobHandler) Create(c *gin.Context) {
    var tj crdv1.TrainingJob
    if err := c.ShouldBindJSON(&tj); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // 设置 namespace（可以从请求头或配置默认）
    tj.Namespace = "default"
    if err := h.k8sClient.Create(context.Background(), &tj); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, tj)
}

func (h *TrainingJobHandler) Get(c *gin.Context) {
    name := c.Param("name")
    var tj crdv1.TrainingJob
    if err := h.k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: name}, &tj); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "training job not found"})
        return
    }
    c.JSON(http.StatusOK, tj)
}

func (h *TrainingJobHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
    var list crdv1.TrainingJobList
    // 使用分页，实际需要支持limit/continue
    if err := h.k8sClient.List(context.Background(), &list, client.InNamespace("default")); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    // 简单分页
    start := (page - 1) * size
    end := start + size
    if start > len(list.Items) {
        start = len(list.Items)
    }
    if end > len(list.Items) {
        end = len(list.Items)
    }
    c.JSON(http.StatusOK, gin.H{
        "total": len(list.Items),
        "items": list.Items[start:end],
    })
}

func (h *TrainingJobHandler) Delete(c *gin.Context) {
    name := c.Param("name")
    var tj crdv1.TrainingJob
    if err := h.k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: name}, &tj); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    if err := h.k8sClient.Delete(context.Background(), &tj); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.Status(http.StatusNoContent)
}
```

### 4.3 日志获取（通过K8s API）

```
func (h *TrainingJobHandler) GetLogs(c *gin.Context) {
    name := c.Param("name")
    follow := c.Query("follow") == "true"

    // 首先获取TrainingJob，找到对应的Job名称
    var tj crdv1.TrainingJob
    if err := h.k8sClient.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: name}, &tj); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "training job not found"})
        return
    }
    jobName := tj.Status.JobName
    if jobName == "" {
        c.JSON(http.StatusNotFound, gin.H{"error": "no underlying job found"})
        return
    }

    // 获取Job的Pod列表
    pods := &corev1.PodList{}
    if err := h.k8sClient.List(context.Background(), pods, client.InNamespace("default"), client.MatchingLabels{"job-name": jobName}); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if len(pods.Items) == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "no pod found for job"})
        return
    }

    // 取第一个Pod的日志（简化）
    pod := pods.Items[0]
    // 使用client-go的RESTClient获取日志流
    req := h.k8sClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: follow})
    stream, err := req.Stream(context.Background())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer stream.Close()

    // 如果需要实时流式返回，设置SSE或chunked transfer
    c.Header("Content-Type", "text/plain")
    c.Stream(func(w io.Writer) bool {
        io.Copy(w, stream)
        return false
    })
}
```

---

## 五、集成到平台

### 5.1 与模型管理服务联动

任务调度服务可以引用模型管理服务中的模型版本，在训练时自动拉取模型文件。例如在TrainingJob的volumes中，增加一个`source: model://model-id/version`，由调度服务解析并挂载模型文件到容器。

这需要扩展VolumeMount的source解析逻辑，调用模型管理服务的下载API或直接从MinIO读取。

### 5.2 部署方式

* Controller 作为独立的Pod运行在K8s集群内（通过Operator模式）
* API服务可以作为Deployment运行，通过Service暴露

---

## 六、测试与验证

1. **创建CRD**：将上述API定义应用到集群

   ```
   make install
   ```
2. **运行Controller**（本地调试）：

   ```
   make run
   ```
3. **提交训练任务**：

   ```
   kubectl apply -f examples/pytorch-training.yaml
   ```
4. **查看状态**：

   ```
   kubectl get trainingjob
   kubectl describe trainingjob resnet-training-20250401
   ```
5. **通过API**：

   ```
   curl -X POST http://localhost:8080/api/v1/training-jobs -d @training.json
   ```

---

## 七、后续扩展建议

* **分布式训练支持**：利用PyTorchJob或TensorFlowJob的CRD，或自己实现多Pod协同（通过头服务、InitContainer等）
* **指标采集**：在训练容器中集成Prometheus exporter，收集训练指标
* **任务依赖**：支持DAG任务流（可基于Argo或Tekton）
* **资源配额**：对接K8s ResourceQuota，限制每个团队的训练资源
* **Web Terminal**：通过websocket提供终端访问（kubectl exec）

现在你已经有了模型管理服务和任务调度服务两个核心模块，接下来可以继续实现**推理服务网关**，形成完整的AI平台。需要继续细化推理网关吗？

继续

已思考（用时 3 秒）

我们收到了用户说“继续”，之前已经完成了模型管理服务和任务调度服务的详细设计。用户希望继续细化，现在需要提供推理服务网关的详细设计。

我们需要给出推理服务网关的详细方案，包括：目标、架构、CRD设计、Controller实现、API设计、代码示例等。这是AI平台的第三个核心模块，负责模型在线推理服务的部署、管理、灰度发布、弹性伸缩等。

我会结合之前模型管理服务、任务调度服务的风格，提供清晰的设计文档，包含关键代码片段和部署建议。

我们继续实现第三个核心模块：**推理服务网关**。这个模块负责将模型部署为在线推理服务，支持多版本、灰度发布、弹性伸缩，并通过统一的网关对外提供API。

---

## 一、推理服务网关目标与功能

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
