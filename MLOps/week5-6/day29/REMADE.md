### 一、MLflow基础概念

1、MLflow是什么？

MLOps里最核心、面试必问的工具，用来管理全生命周期:

* 记录：模型参数、指标、日志、文件
* 管理：模型版本、上线 / 测试 / 归档
* 推理：直接加载模型用

### 2、四大核心组件

* **Tracking（实验跟踪）**
* **Model Registry（模型注册）**
* Projects、Models 暂时不用学

### 3. 关键名词

* **Run**：一次训练 / 实验过程
* **Param**：参数（如学习率、迭代次数）
* **Metric**：指标（准确率、损失、MAE）
* **Artifact**：产出物（模型文件、图片、日志）
* **Model Registry**：模型仓库（管版本）
* **Stage**：模型阶段
  * Staging：测试
  * Production：上线
  * Archived：归档

---

## 二、环境准备（一步一步来）

### 1. 安装依赖

bash

运行

```
pip install mlflow scikit-learn pandas numpy
```

### 2. 启动 MLflow UI（关键）

打开终端运行：

bash

运行

```
mlflow ui --host 0.0.0.0 --port 5000
```

然后浏览器打开：

[http://localhost:5000](http://localhost:5000)
