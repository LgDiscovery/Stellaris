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