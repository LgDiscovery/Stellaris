### 一、Dockerfile 编写（适配模型 API）

首先针对 model\_api\_u.py 编写 Dockerfile，确保镜像能完整运行 API 服务：

dockerfile

```
# 1. 选择基础镜像（Python官方轻量版，适配FastAPI+机器学习依赖）
FROM python:3.9-slim

# 2. 设置工作目录（容器内的代码存放路径）
WORKDIR /app

# 3. 复制依赖文件（先复制requirements.txt，利用Docker缓存优化）
# 先手动创建requirements.txt，包含所有依赖：
# fastapi==0.104.1
# uvicorn==0.24.0
# scikit-learn==1.3.2
# joblib==1.3.2
# numpy==1.26.1
COPY requirements.txt .

# 4. 安装依赖（国内源加速，避免下载超时）
RUN pip install --no-cache-dir -i https://pypi.tuna.tsinghua.edu.cn/simple -r requirements.txt

# 5. 复制项目代码（模型文件+API代码）
COPY model_api_u.py .
COPY pretrained_linear_regression_model.joblib .  # 确保模型文件和代码同目录

# 6. 暴露端口（和API服务的9000端口一致）
EXPOSE 9000

# 7. 启动命令（生产模式，关闭reload）
CMD ["uvicorn", "model_api_u:app", "--host", "0.0.0.0", "--port", "9000", "--log-level", "info"]
```

### 二、配套操作步骤（本地运行 + 调试）

#### 1. 准备文件目录

确保本地目录结构如下：

plaintext

```
├── model_api_u.py       # 你的API代码
├── pretrained_linear_regression_model.joblib  # 预训练模型文件
├── requirements.txt     # 依赖清单
└── Dockerfile           # 上面编写的Dockerfile
```

#### 2. 创建 `requirements.txt`

txt

```
fastapi==0.104.1
uvicorn==0.24.0
scikit-learn==1.3.2
joblib==1.3.2
numpy==1.26.1
```

#### 3. 构建 Docker 镜像

打开终端，进入文件目录，执行：

bash

运行

```
# 构建镜像，命名为diabetes-api，标签v1
docker build -t diabetes-api:v1 .
```

#### 4. 运行 Docker 容器

bash

运行

```
# 启动容器，映射本地9000端口到容器9000端口
docker run -d -p 9000:9000 --name diabetes-api-container diabetes-api:v1
```

#### 5. 调试接口是否可用

* **健康检查接口**：访问 `http://localhost:9000/health`，正常返回如下：
  json

  ```
  {
    "code": 200,
    "message": "服务正常",
    "data": {
      "model_loaded": true,
      "api_version": "2.0.0"
    }
  }
  ```
* **预测接口调试**（Postman/Curl）：
  bash

  运行

  ```
  curl -X POST "http://localhost:9000/predict" -H "Content-Type: application/json" -d '{
    "feature1": 0.03807591,
    "feature2": 0.05068012,
    "feature3": 0.06169621,
    "feature4": 0.02187235,
    "feature5": -0.0442235,
    "feature6": -0.03482076,
    "feature7": -0.04340085,
    "feature8": -0.00259226,
    "feature9": 0.01990749,
    "feature10": -0.01764613
  }'
  ```

  正常返回包含 `prediction_result` 的 JSON 结果。

#### 6. 常用调试命令

bash

运行

```
# 查看容器日志（排查启动/推理错误）
docker logs diabetes-api-container

# 进入容器内部（检查文件/依赖）
docker exec -it diabetes-api-container /bin/bash

# 停止/删除容器
docker stop diabetes-api-container
docker rm diabetes-api-container
```

### 三、Dockerfile 核心知识讲解

#### 1. 基础镜像（Base Image）

* **定义**：Docker 镜像的 “底层基础”，包含运行程序所需的操作系统、核心依赖，所有自定义镜像都基于基础镜像构建。
* **分类**（按场景）：
  表格


  | 类型         | 示例                              | 适用场景                          |
  | ------------ | --------------------------------- | --------------------------------- |
  | 官方基础镜像 | python:3.9-slim                   | Python 项目（轻量，仅含核心依赖） |
  | 操作系统镜像 | ubuntu:22.04/alpine               | 需要自定义系统环境的项目          |
  | 框架封装镜像 | tiangolo/uvicorn-gunicorn-fastapi | FastAPI 生产环境（已封装依赖）    |
* **选择原则**：

  * 优先选**官方镜像**（安全、维护及时）；
  * 优先选**slim/alpine 版**（体积小，减少攻击面），比如 `python:3.9-slim` 比 `python:3.9` 体积小 80%+；
  * 版本固定（避免用`python:latest`，防止依赖兼容问题）。

#### 2. Dockerfile 核心指令

表格


| 指令      | 作用                                                                      |
| --------- | ------------------------------------------------------------------------- |
| `FROM`    | 指定基础镜像（必须是第一个指令）                                          |
| `WORKDIR` | 设置容器内工作目录（后续指令基于此目录执行）                              |
| `COPY`    | 从本地复制文件 / 目录到容器内                                             |
| `RUN`     | 在构建镜像时执行命令（如安装依赖、创建目录）                              |
| `EXPOSE`  | 声明容器暴露的端口（仅文档作用，实际映射需`docker run -p`）               |
| `CMD`     | 容器启动时执行的命令（一个 Dockerfile 只能有一个 CMD，会被`run`参数覆盖） |
| `ENV`     | 设置容器内环境变量（如`ENV PYTHONPATH=/app`）                             |

#### 3. 镜像构建核心逻辑

* **分层构建**：Dockerfile 每个指令对应一个 “镜像层”，层会被缓存（比如`COPY requirements.txt`后，只要依赖文件不变，再次构建会跳过`RUN pip install`）；
* **缓存优化**：先复制依赖文件（`requirements.txt`），再安装依赖，最后复制代码（代码频繁修改不会触发依赖重装）；
* **轻量性**：避免安装无用依赖（比如`python-slim`已去除编译工具，如需编译依赖可临时安装后删除）。

#### 4. 生产环境优化建议

* 关闭`reload`（开发模式才用，生产用`uvicorn`默认模式）；
* 使用多阶段构建（构建依赖和运行依赖分离，进一步减小镜像体积）；
* 非 root 用户运行容器（避免权限风险）：
  dockerfile

  ```
  # 新增：创建非root用户
  RUN useradd -m appuser
  USER appuser
  ```

### 四、常见问题排查

1. **模型文件找不到**：确保`COPY pretrained_linear_regression_model.joblib .` 路径正确，容器内执行`ls /app`检查文件是否存在；
2. **端口映射失败**：检查本地 9000 端口是否被占用（`netstat -tulpn | grep 9000`），或换端口（`docker run -p 9001:9000 ...`）；
3. **依赖安装失败**：换国内源（如清华 pypi），或检查`requirements.txt`版本是否兼容；
4. **特征范围校验失败**：确保请求参数在`-0.1~0.1`之间，查看容器日志（`docker logs`）定位具体错误特征。

### 六、多阶段构建的 Dockerfile（优化版）

多阶段构建核心是**拆分 “构建依赖” 和 “运行依赖”**，最终镜像仅保留运行所需文件，大幅减小体积（原镜像～800MB → 优化后～200MB）。

dockerfile

```
# ===================== 阶段1：构建阶段（复用缓存） =====================
FROM python:3.9-slim AS builder

# 提前安装编译依赖（这层只要不修改，永远复用缓存）
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends gcc g++ \
    && rm -rf /var/lib/apt/lists/*

# 先复制requirements.txt（仅当依赖文件修改时，才重新编译wheel）
WORKDIR /app
COPY requirements.txt .

# 用国内源+缓存wheel文件（编译一次后，后续复用）
RUN pip wheel --no-cache-dir --wheel-dir /app/wheels \
    -i https://pypi.tuna.tsinghua.edu.cn/simple \
    -r requirements.txt

# ===================== 阶段2：运行阶段（只复制运行必需的东西） =====================
FROM python:3.9-slim AS runner

# 安全优化：非root用户
RUN useradd -m appuser
WORKDIR /app

# 复制wheel并安装，然后立刻删除wheel文件
COPY --from=builder /app/wheels /wheels
COPY --from=builder /app/requirements.txt .
# ✅ 关键：删除临时wheel，减小体积
RUN pip install --no-cache-dir --no-index /wheels/* \
    && rm -rf /wheels  

# 只复制代码和模型
COPY model_api_u.py .
COPY pretrained_linear_regression_model.joblib .

# 授权给非root用户
RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 9000
CMD ["uvicorn", "model_api_u:app", "--host", "0.0.0.0", "--port", "9000"]
```

### 七、多阶段构建核心优势

表格


| 优化点       | 传统单阶段                   | 多阶段构建                 |
| ------------ | ---------------------------- | -------------------------- |
| 镜像体积     | \~800MB（含 gcc 等编译工具） | \~200MB（仅保留运行依赖）  |
| 安全风险     | 含编译工具，攻击面大         | 无编译工具，仅保留运行文件 |
| 依赖安装效率 | 每次构建重新编译             | 预编译 wheel，复用缓存     |
| 权限安全     | 默认为 root 用户             | 非 root 用户运行，降低风险 |

### 八、配套文件与操作说明

#### 1. 优化后的 requirements.txt（无需修改，保持原内容）

txt

```
fastapi==0.104.1
uvicorn==0.24.0
scikit-learn==1.3.2
joblib==1.3.2
numpy==1.26.1
```

#### 2. 构建与运行命令（和单阶段一致，无需修改）

bash

运行

```
# 构建镜像（多阶段构建无需额外参数，Docker自动识别阶段）
docker build -t diabetes-api:v2 .

# 运行容器（非root用户+端口映射）
docker run -d -p 9000:9000 --name diabetes-api-v2 diabetes-api:v2

# 验证接口（和之前一致）
curl http://localhost:9000/health
```

#### 3. 关键优化点解释

1. **构建阶段（builder）**：
   * 安装`gcc/g++`（scikit-learn/numpy 等库编译依赖）；
   * 用`pip wheel`将依赖打包为 wheel 文件（预编译，避免运行阶段重复编译）；
   * 编译依赖仅存在于构建阶段，不进入最终镜像。
2. **运行阶段（runner）**：
   * 仅从构建阶段复制预编译的 wheel 文件，快速安装（无需编译）；
   * 创建`appuser`非 root 用户，避免容器以 root 权限运行（降低黑客攻击风险）；
   * 删除临时`wheels`目录，进一步减小镜像体积；
   * 设置`UVICORN_ACCESS_LOGS=false`关闭访问日志（生产环境可自定义日志配置）。

### 九、多阶段构建进阶技巧

1. **按需拆分阶段**：如果项目有前端 / 后端分离，可新增 “前端构建阶段”（如 node 镜像编译前端静态文件），再复制到后端运行镜像。
2. **复用缓存**：构建阶段的`pip wheel`会缓存已编译的依赖包，只要`requirements.txt`不变，再次构建会直接复用缓存。
3. **精简运行镜像**：可选择`python:3.9-alpine`（更小的 Alpine 系统），但需注意 Alpine 的 libc 和 Ubuntu 不同，部分编译库需适配。

### 十、调试命令（适配多阶段镜像）

bash

运行

```
# 查看镜像分层（验证多阶段是否精简）
docker history diabetes-api:v2

# 进入容器（验证非root用户）
docker exec -it diabetes-api-v2 whoami  # 输出应为appuser

# 检查依赖是否仅安装运行包
docker exec -it diabetes-api-v2 pip list  # 无gcc等编译工具相关包
```
