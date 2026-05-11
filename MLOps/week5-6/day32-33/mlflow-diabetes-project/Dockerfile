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
CMD ["uvicorn", "api.main:app", "--host", "0.0.0.0", "--port", "9090"]