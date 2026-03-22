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