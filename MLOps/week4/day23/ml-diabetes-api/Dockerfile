# ===================== 阶段1：构建阶段（Builder）- 仅编译依赖 =====================
# 选择轻量基础镜像，换国内源加速
FROM python:3.9-slim AS builder

# 配置环境变量（MLOps：固定依赖版本，避免漂移）
ENV PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple


# 安装编译依赖（仅模型库需要，安装后清理缓存） # 关键：清理apt缓存，不残留
WORKDIR /app
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends gcc g++ \
    && rm -rf /var/lib/apt/lists/*


# 编译依赖为wheel（预编译，运行阶段无需编译）
COPY requirements.txt .
RUN pip wheel --wheel-dir /app/wheels -r requirements.txt


# ===================== 阶段2：运行阶段（Runner）- 仅推理环境 =====================
# 完全纯净的运行镜像，仅保留推理必需
FROM python:3.9-slim AS runner

# MLOps：非root用户运行，降低部署风险
RUN useradd -m mlops-user && mkdir -p /app && chown -R mlops-user:mlops-user /app

# 环境变量（MLOps：配置日志/编码，适配生产环境）  # 关闭访问日志，减少镜像运行时IO
ENV PYTHONUNBUFFERED=1 \
    LC_ALL=C.UTF-8 \
    LANG=C.UTF-8 \
    UVICORN_ACCESS_LOGS=false

WORKDIR /app

# 仅复制编译好的wheel包，安装后立即删除 # 关键：删除临时文件，减小体积
COPY --from=builder /app/wheels /wheels
COPY --from=builder /app/requirements.txt .
RUN pip install --no-index /wheels/* \
    && rm -rf /wheels ~/.cache/pip

# 仅复制模型和API代码（MLOps：最小化代码暴露）
COPY model_api_u.py .
COPY pretrained_model.joblib .

# 切换非root用户
USER mlops-user

# 暴露端口（MLOps：声明端口，适配K8s/容器编排）
EXPOSE 9000

# 启动命令（MLOps：生产级启动，禁用reload）
CMD ["uvicorn", "model_api_u:app", "--host", "0.0.0.0", "--port", "9000", "--workers", "2"]