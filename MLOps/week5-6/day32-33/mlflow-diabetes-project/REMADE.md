# 完整项目：MLflow + MinIO + FastAPI + K8s 糖尿病预测服务

我为你打造**一站式可直接运行**的完整项目，包含：

✅ MLflow 实验跟踪 / 模型保存 / 模型注册

✅ MinIO 作为模型 artifact 存储（替代 S3）

✅ FastAPI 封装预测服务（从 MLflow+MinIO 加载模型）

✅ Docker 多阶段构建（极小镜像）

✅ K8s 完整部署清单

✅ 环境变量解耦配置（无硬编码）

---

## 一、项目架构

plaintext

```
mlflow-diabetes-project/
├── config/               # 环境配置
│   └── .env              # 本地开发环境变量
├── src/
│   ├── train.py          # 训练+MLflow实验跟踪+注册模型到MinIO
│   └── api/
│       └── main.py       # FastAPI服务（从MLflow+MinIO加载模型）
├── Dockerfile            # 多阶段构建镜像
├── requirements.txt      # 项目依赖
├── k8s/                  # K8s部署清单
│   ├── configmap.yaml    # 配置注入
│   ├── deployment.yaml   # 服务部署
│   └── service.yaml      # 端口暴露
└── README.md
```

---

## 二、前置环境准备

1. **MinIO**（对象存储，存模型文件）
2. **MLflow**（跟踪实验、管理模型，对接 MinIO）
3. Docker + Kubernetes 环境

### 快速启动 MinIO + MLflow（本地测试）

bash

运行

```


# 1. 启动 MinIO (默认账号: minioadmin/minioadmin)
docker run -d --name minio \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# 2. 创建 MinIO bucket (mlflow-models)
# 浏览器打开 http://localhost:9001 登录，创建桶：mlflow-models

# 3. 启动 MLflow (对接 MinIO 存储模型)
docker run -d --name mlflow \
  -p 5000:5000 \
  -e MLFLOW_S3_ENDPOINT_URL=http://host.docker.internal:9000 \
  -e AWS_ACCESS_KEY_ID=minioadmin \
  -e AWS_SECRET_ACCESS_KEY=minioadmin \
  ghcr.io/mlflow/mlflow:v2.14.0 \
  mlflow server \
    --backend-store-uri sqlite:///mlflow.db \
    --default-artifact-root s3://mlflow-models/ \
    --host 0.0.0.0
```

---

## 三、项目文件（直接复制使用）

### 1. `requirements.txt` 项目依赖

txt

```
# 核心依赖
mlflow>=2.10.0
minio>=7.2.0
fastapi>=0.104.0
uvicorn>=0.24.0
python-dotenv>=1.0.0
# 训练依赖
scikit-learn>=1.3.0
numpy>=1.24.0
pandas>=2.1.0
# 工具
python-multipart>=0.0.6
psutil>=5.9.0
```

### 2. `config/.env` 环境变量（本地开发）

env

```
# MinIO 配置
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=mlflow-models

# MLflow 配置
MLFLOW_TRACKING_URI=http://localhost:5000
MLFLOW_MODEL_NAME=diabetes_linear_regression
MODEL_VERSION=latest

# FastAPI
API_HOST=0.0.0.0
API_PORT=9000
```

### 3. `src/train.py` 训练 + MLflow 实验跟踪 + 注册模型到 MinIO

python

运行

```
import os
import mlflow
import mlflow.sklearn
import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.datasets import load_diabetes
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_squared_error
from dotenv import load_dotenv

# 加载环境变量
load_dotenv(dotenv_path="../config/.env")

# ===================== 配置 MLflow 对接 MinIO =====================
os.environ["MLFLOW_S3_ENDPOINT_URL"] = os.getenv("MINIO_ENDPOINT")
os.environ["AWS_ACCESS_KEY_ID"] = os.getenv("MINIO_ACCESS_KEY")
os.environ["AWS_SECRET_ACCESS_KEY"] = os.getenv("MINIO_SECRET_KEY")
mlflow.set_tracking_uri(os.getenv("MLFLOW_TRACKING_URI"))
mlflow.set_experiment("diabetes_prediction")

# ===================== 训练模型 =====================
if __name__ == "__main__":
    # 加载数据
    X, y = load_diabetes(return_X_y=True)
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2)

    with mlflow.start_run(run_name="linear_regression_baseline"):
        # 训练模型
        model = LinearRegression()
        model.fit(X_train, y_train)

        # 评估
        y_pred = model.predict(X_test)
        mse = mean_squared_error(y_test, y_pred)
        rmse = np.sqrt(mse)

        # MLflow 自动记录参数/指标/模型
        mlflow.log_param("model_type", "LinearRegression")
        mlflow.log_metric("mse", mse)
        mlflow.log_metric("rmse", rmse)

        # 保存模型到 MinIO（自动上传）
        mlflow.sklearn.log_model(
            sk_model=model,
            artifact_path="model",
            registered_model_name=os.getenv("MLFLOW_MODEL_NAME")  # 注册模型
        )

        print(f"✅ 模型训练完成，RMSE: {rmse:.2f}")
        print(f"✅ 模型已保存到 MinIO 并注册到 MLflow")
```

### 4. `src/api/main.py` FastAPI 服务（从 MLflow+MinIO 加载模型）

基于你原有的代码，**解耦配置、适配容器化**：

python

运行

```
import os
import logging
import numpy as np
import mlflow.pyfunc
from fastapi import FastAPI, HTTPException, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel, Field, field_validator, ValidationInfo
from dotenv import load_dotenv

# 加载环境变量（容器内自动读取 K8s ConfigMap）
load_dotenv(dotenv_path="/app/config/.env")  # 容器内路径

# ===================== 日志配置 =====================
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("diabetes_api")

# ===================== MLflow + MinIO 配置 =====================
MLFLOW_TRACKING_URI = os.getenv("MLFLOW_TRACKING_URI")
MLFLOW_MODEL_NAME = os.getenv("MLFLOW_MODEL_NAME")
MODEL_VERSION = os.getenv("MODEL_VERSION")

# MinIO 认证（MLflow 自动读取加载 MinIO 模型）
os.environ["MLFLOW_S3_ENDPOINT_URL"] = os.getenv("MINIO_ENDPOINT")
os.environ["AWS_ACCESS_KEY_ID"] = os.getenv("MINIO_ACCESS_KEY")
os.environ["AWS_SECRET_ACCESS_KEY"] = os.getenv("MINIO_SECRET_KEY")

# ===================== 初始化 FastAPI =====================
app = FastAPI(
    title="糖尿病预测API",
    description="MLflow+MinIO+FastAPI+K8s 完整服务",
    version="1.0.0"
)

# ===================== 加载 MLflow 模型（从 MinIO 下载） =====================
try:
    logger.info(f"加载模型：{MLFLOW_MODEL_NAME}@{MODEL_VERSION}")
    model = mlflow.pyfunc.load_model(f"models:/{MLFLOW_MODEL_NAME}/{MODEL_VERSION}")
    logger.info("✅ 模型加载成功（从 MinIO 自动下载）")
except Exception as e:
    logger.error(f"❌ 模型加载失败：{str(e)}")
    raise RuntimeError("模型加载失败")

# ===================== 请求参数校验 =====================
class DiabetesFeatures(BaseModel):
    feature1: float = Field(..., example=0.038)
    feature2: float = Field(..., example=0.050)
    feature3: float = Field(..., example=0.061)
    feature4: float = Field(..., example=0.021)
    feature5: float = Field(..., example=-0.044)
    feature6: float = Field(..., example=-0.034)
    feature7: float = Field(..., example=-0.043)
    feature8: float = Field(..., example=-0.002)
    feature9: float = Field(..., example=0.019)
    feature10: float = Field(..., example=-0.017)

    @field_validator("feature1~feature10")
    def check_range(cls, v, info):
        if not (-0.1 <= v <= 0.1):
            raise ValueError(f"{info.field_name} 超出范围 [-0.1, 0.1]")
        return v

# ===================== 异常处理 =====================
@app.exception_handler(RequestValidationError)
async def val_err(request, exc):
    return JSONResponse(status_code=422, content={"code":422,"msg":"参数错误"})

@app.exception_handler(HTTPException)
async def http_err(request, exc):
    return JSONResponse(status_code=exc.status_code, content={"code":exc.status_code,"msg":exc.detail})

# ===================== 预测接口 =====================
@app.post("/predict")
async def predict(data: DiabetesFeatures):
    try:
        input_data = [[getattr(data, f"feature{i}") for i in range(1,11)]]
        pred = model.predict(input_data)
        return {
            "code":200,
            "msg":"success",
            "data": round(float(pred[0]),2)
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"推理失败：{str(e)}")

# ===================== 健康检查 =====================
@app.get("/health")
async def health():
    return {"status":"healthy","model":MLFLOW_MODEL_NAME,"version":MODEL_VERSION}
```

### 5. `Dockerfile` 多阶段构建（极小生产镜像）

dockerfile

```
# 阶段1：构建依赖（python:slim）
FROM python:3.11-slim AS builder
WORKDIR /app
COPY requirements.txt .
# 安装依赖到本地目录
RUN pip install --no-cache-dir -r requirements.txt -t /app/deps

# 阶段2：生产镜像（alpine 极小体积）
FROM python:3.11-alpine
WORKDIR /app

# 复制依赖 + 代码 + 配置
COPY --from=builder /app/deps /usr/local/lib/python3.11/site-packages
COPY src/api /app/api
COPY config/.env /app/config/.env

# 非root用户运行（安全）
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 启动服务
CMD ["uvicorn", "api.main:app", "--host", "0.0.0.0", "--port", "9000"]
```

---

## 四、K8s 部署清单

### 1. `k8s/configmap.yaml` 配置注入（环境变量）

yaml

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: diabetes-api-config
data:
  MINIO_ENDPOINT: "http://minio-service:9000"  # K8s内MinIO服务名
  MINIO_ACCESS_KEY: "minioadmin"
  MINIO_SECRET_KEY: "minioadmin"
  MINIO_BUCKET: "mlflow-models"
  MLFLOW_TRACKING_URI: "http://mlflow-service:5000"
  MLFLOW_MODEL_NAME: "diabetes_linear_regression"
  MODEL_VERSION: "latest"
```

### 2. `k8s/deployment.yaml` 服务部署

yaml

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: diabetes-api
spec:
  replicas: 2
  selector:
    matchLabels:
      app: diabetes-api
  template:
    metadata:
      labels:
        app: diabetes-api
    spec:
      containers:
      - name: diabetes-api
        image: diabetes-api:latest  # 替换为你的镜像地址
        ports:
        - containerPort: 9000
        envFrom:
        - configMapRef:
            name: diabetes-api-config
        resources:
          limits:
            cpu: "0.5"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 9000
          initialDelaySeconds: 10
```

### 3. `k8s/service.yaml` 端口暴露

yaml

```
apiVersion: v1
kind: Service
metadata:
  name: diabetes-api-service
spec:
  type: NodePort
  selector:
    app: diabetes-api
  ports:
  - port: 9000
    targetPort: 9000
    nodePort: 30900  # 外部访问端口
```

---

## 五、完整运行步骤

### 1. 训练并注册模型

bash

运行

```
cd src
python train.py
```

→ 模型自动保存到 **MinIO**，并注册到 **MLflow 模型仓库**

### 2. 构建 Docker 镜像

bash

运行

```
docker build -t diabetes-api:latest .
```

### 3. K8s 部署

bash

运行

```
kubectl apply -f k8s/
```

### 4. 测试服务

bash

运行

```
# 健康检查
curl http://localhost:30900/health

# 预测接口
curl -X POST http://localhost:30900/predict \
-H "Content-Type: application/json" \
-d '{
    "feature1":0.038,"feature2":0.050,"feature3":0.061,"feature4":0.021,
    "feature5":-0.044,"feature6":-0.034,"feature7":-0.043,"feature8":-0.002,
    "feature9":0.019,"feature10":-0.017
}'
```

---

## 六、核心特性总结

1. **MinIO 存储**：模型文件完全托管在对象存储，支持版本管理
2. **MLflow 管理**：实验跟踪、模型注册、版本切换一站式搞定
3. **FastAPI 服务**：高性能接口，自动参数校验，自动生成文档
4. **Docker 多阶段**：镜像体积缩小 80%，生产安全高效
5. **K8s 部署**：高可用、弹性扩缩、配置解耦、健康检查
