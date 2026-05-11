## 一、基础概念题（必问）

### 1. MLflow 是什么？在 MLOps 里有什么用？

**答**：

MLflow 是开源的 MLOps 工具，用于**模型全生命周期管理**。主要做三件事：

* 记录实验：参数、指标、模型、日志
* 管理模型：版本、阶段、回滚
* 支持推理：直接加载模型上线
  它解决了 “模型乱存、版本混乱、不好上线” 的问题。

### 2. MLflow 四大组件是什么？你常用哪两个？

**答**：

* Tracking：实验跟踪（最常用）
* Model Registry：模型注册与版本管理（最常用）
* Projects：代码打包
* Models：模型部署
  我在项目里主要用 **Tracking + Registry**。

### 3. Run / Param / Metric / Artifact 分别是什么？

**答**：

* **Run**：一次训练 / 实验过程
* **Param**：输入参数（如学习率、迭代次数）
* **Metric**：输出指标（准确率、R2、损失）
* **Artifact**：产出物（模型文件、图片、日志）

---

## 二、Tracking 实验跟踪题（第 29 天）

### 4. 怎么用 MLflow 记录一次实验？

**答**：

1. 设置实验名 `mlflow.set_experiment()`
2. 用 `mlflow.start_run()` 开启一次运行
3. 用 `log_param` 记参数、`log_metric` 记指标
4. 用 `log_model` 存模型
5. 打开 UI 查看结果

### 5. MLflow UI 能看到什么？

**答**：

* 所有实验与运行记录
* 参数、指标对比
* 模型文件与 artifact
* run\_id（用于注册模型）

---

## 三、Model Registry 模型注册题（第 30 天・高频）

### 6. 模型注册（Registry）有什么用？

**答**：

* 集中存放模型，方便查找
* 管理多版本，支持回滚
* 区分环境：Staging / Production / Archived
* 保证上线模型可追溯、可复用

### 7. Staging / Production / Archived 分别代表什么？

**答**：

* **Staging**：测试环境，用于验证
* **Production**：生产环境，正式上线
* **Archived**：归档，废弃不用

### 8. 怎么从 MLflow 加载指定版本 / 阶段的模型？

**答**：

* 按版本：`models:/MyModel/1`
* 按阶段：`models:/MyModel/Production`
  用 `mlflow.pyfunc.load_model()` 加载后直接推理。

### 9. 模型版本冲突怎么处理？

**答**：

* 用版本号区分
* 用阶段隔离测试与生产
* 异常时直接回滚到上一个稳定版本
* 保留所有 run 记录，方便溯源

---

## 四、实战场景题（面试加分）

### 10. 你在项目里怎么用 MLflow？

**答**：

1. 训练时用 Tracking 记录参数、指标、模型
2. 把验证通过的模型注册到 Registry
3. 模型分 Staging/Production 管理
4. 推理服务直接从 Registry 加载模型
5. 新版本有问题，一键回滚

### 11. MLflow 对比其他模型管理工具（如 ModelArts）有什么优势？

**答**：

* 开源免费、轻量、部署简单
* 不绑定云厂商，本地 / 私有云都能用
* 与 K8s、Docker、CI/CD 无缝配合
* 适合云原生 MLOps 架构

---

## 五、口试题（10 秒速答版）

* MLflow 核心：**实验跟踪 + 模型版本管理**
* 模型阶段：**Staging → Production → Archived**
* 加载模型：`models:/模型名/Production`
* 关键 ID：**run\_id**
* 核心价值：**模型可追溯、可版本、可回滚**
