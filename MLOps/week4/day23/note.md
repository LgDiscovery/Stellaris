围绕 MLOps 实战场景，完成 Docker 镜像的三大优化：**减小体积**、**规范多阶段构建**、**标准化版本标签管理**，并重新构建运行优化后的镜像。我会结合 MLOps 的最佳实践，给出可直接落地的完整方案。

### 一、MLOps 场景下的镜像优化核心思路

MLOps 中镜像优化不仅要 “小”，还要满足：版本可追溯、构建可复现、部署可验证。核心原则：

1. **体积最小化**：仅保留模型推理必需的依赖 / 文件，剥离所有构建工具；
2. **多阶段纯净构建**：构建阶段（编译依赖 / 模型）与运行阶段（仅推理）严格分离；
3. **版本标签标准化**：采用「语义化版本 + 模型版本 + 构建时间」的标签规则，便于追溯和部署。

### 二、完整优化方案（贴合 MLOps 实战）

#### 1. 前置准备：规范文件结构（MLOps 标准）

plaintext

```
ml-diabetes-api/
├── .dockerignore          # 排除无关文件，减小构建上下文
├── Dockerfile             # 多阶段优化版
├── requirements.txt       # 仅保留推理依赖
├── model_api_u.py         # API服务代码
├── pretrained_model.joblib # 模型文件
└── build.sh               # 一键构建脚本（含版本标签）
```

#### 2. 关键文件编写

##### (1) .dockerignore（必加！减少构建上下文体积）

plaintext

```
# 排除Python缓存
__pycache__/
*.pyc
*.pyo
*.pyd
.venv/
venv/

# 排除构建临时文件
wheels/
dist/
build/

# 排除版本控制/日志
.git/
.gitignore
logs/
*.log

# 排除本地配置
.env
*.yml
*.md
```

##### (2) 优化版 Dockerfile（MLOps 定制）

dockerfile

```
# ===================== 阶段1：构建阶段（Builder）- 仅编译依赖 =====================
# 选择轻量基础镜像，换国内源加速
FROM python:3.9-slim AS builder

# 配置环境变量（MLOps：固定依赖版本，避免漂移）
ENV PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple

# 安装编译依赖（仅模型库需要，安装后清理缓存）
WORKDIR /app
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends gcc g++ \
    && rm -rf /var/lib/apt/lists/*  # 关键：清理apt缓存，不残留

# 编译依赖为wheel（预编译，运行阶段无需编译）
COPY requirements.txt .
RUN pip wheel --wheel-dir /app/wheels -r requirements.txt

# ===================== 阶段2：运行阶段（Runner）- 仅推理环境 =====================
# 完全纯净的运行镜像，仅保留推理必需
FROM python:3.9-slim AS runner

# MLOps：非root用户运行，降低部署风险
RUN useradd -m mlops-user && mkdir -p /app && chown -R mlops-user:mlops-user /app

# 环境变量（MLOps：配置日志/编码，适配生产环境）
ENV PYTHONUNBUFFERED=1 \
    LC_ALL=C.UTF-8 \
    LANG=C.UTF-8 \
    UVICORN_ACCESS_LOGS=false  # 关闭访问日志，减少镜像运行时IO

WORKDIR /app

# 仅复制编译好的wheel包，安装后立即删除
COPY --from=builder /app/wheels /wheels
COPY --from=builder /app/requirements.txt .
RUN pip install --no-index /wheels/* \
    && rm -rf /wheels ~/.cache/pip  # 关键：删除临时文件，减小体积

# 仅复制模型和API代码（MLOps：最小化代码暴露）
COPY model_api_u.py .
COPY pretrained_model.joblib .

# 切换非root用户
USER mlops-user

# 暴露端口（MLOps：声明端口，适配K8s/容器编排）
EXPOSE 9000

# 启动命令（MLOps：生产级启动，禁用reload）
CMD ["uvicorn", "model_api_u:app", "--host", "0.0.0.0", "--port", "9000", "--workers", "2"]
```

##### (3) requirements.txt（MLOps：固定版本，可复现）

txt

```
fastapi==0.104.1
uvicorn==0.24.0
scikit-learn==1.3.2
joblib==1.3.2
numpy==1.26.1
# 仅保留推理必需，删除所有开发依赖（如pytest、jupyter等）
```

##### (4) build.sh（MLOps：一键构建 + 标准化版本标签）

bash

运行

```
#!/bin/bash
# MLOps镜像版本规则：{项目名}-{模型版本}-{API版本}-{构建日期}
PROJECT_NAME="diabetes-api"
MODEL_VERSION="v1.0"  # 模型版本（对应第22天的模型）
API_VERSION="v2.0"    # API版本
BUILD_DATE=$(date +%Y%m%d)
TAG="${PROJECT_NAME}:${MODEL_VERSION}-${API_VERSION}-${BUILD_DATE}"
LATEST_TAG="${PROJECT_NAME}:latest"

# 停止并删除旧容器
echo "=== 清理旧容器 ==="
docker stop ${PROJECT_NAME}-container || true
docker rm ${PROJECT_NAME}-container || true

# 构建镜像（多阶段优化）
echo "=== 构建镜像：${TAG} ==="
docker build -t ${TAG} -t ${LATEST_TAG} .

# 查看镜像体积
echo "=== 镜像体积 ==="
docker images | grep ${PROJECT_NAME}

# 运行新容器（MLOps：挂载日志目录，便于监控）
echo "=== 启动容器 ==="
docker run -d -p 9000:9000 \
  --name ${PROJECT_NAME}-container \
  --restart=on-failure:3 \  # MLOps：故障自动重启
  -v $(pwd)/logs:/app/logs \ # 挂载日志目录到宿主机
  ${TAG}

# 验证接口
echo "=== 验证接口 ==="
sleep 5  # 等待服务启动
curl -s http://localhost:9000/health && echo -e "\n=== 接口验证成功 ==="
```

### 三、执行步骤（MLOps 实战流程）

#### 1. 赋予构建脚本执行权限

bash

运行

```
chmod +x build.sh
```

#### 2. 一键构建 + 运行

bash

运行

```
./build.sh
```

#### 3. 验证结果

* 镜像体积：优化后应从 699MB 降至**200MB 以内**；
* 版本标签：生成如 `diabetes-api:v1.0-v2.0-20260321` 的标签，便于追溯；
* 接口验证：`http://localhost:9000/health` 返回正常，推理接口可用。

### 四、MLOps 场景下的镜像版本管理最佳实践

#### 1. 标签命名规范（核心）

plaintext

```
{项目名}:{模型版本}-{API版本}-{构建时间}-{环境}
# 示例：diabetes-api:v1.0-v2.0-20260321-prod
```

* 模型版本：对应训练的模型版本（如 v1.0、v1.1）；
* API 版本：对应接口的版本（如 v2.0）；
* 构建时间：YYYYMMDD，便于区分同版本不同构建；
* 环境：dev/test/prod，区分部署环境。

#### 2. 版本追溯与清理

bash

运行

```
# 查看镜像构建历史（MLOps：追溯每一层的构建信息）
docker history diabetes-api:v1.0-v2.0-20260321

# 清理无用镜像（MLOps：定期清理，节省空间）
docker image prune -f
docker rmi $(docker images -f "dangling=true" -q)
```

#### 3. 镜像推送（可选，适配 MLOps 仓库）

bash

运行

```
# 标记镜像（适配私有仓库）
docker tag diabetes-api:v1.0-v2.0-20260321 registry.example.com/mlops/diabetes-api:v1.0-v2.0-20260321

# 推送至仓库（MLOps：统一镜像仓库管理）
docker push registry.example.com/mlops/diabetes-api:v1.0-v2.0-20260321
```

### 总结

1. **体积优化核心**：选`slim`基础镜像 + 多阶段仅复制运行产物 + 每层即时清理临时文件，体积从 699MB 降至 200MB 内；
2. **多阶段构建**：构建阶段仅编译依赖，运行阶段仅保留推理环境，剥离所有编译工具；
3. **MLOps 版本管理**：采用「模型版本 + API 版本 + 构建时间」的标签规则，配合一键构建脚本，实现版本可追溯、部署可复现。

按这套方案优化后，镜像既满足 MLOps 实战的标准化要求，又能保证体积最小、运行稳定，完全适配模型 API 的生产部署场景。
