核心需求是基于 MLOps 实战场景，把第 23 天优化好的模型镜像部署到 K8s 集群，核心要完成**Deployment 编排部署**、**Service 暴露服务**、**端口映射访问接口**这三个关键步骤。

### 一、前置准备（确保 K8s 环境可用）

1. 确认 K8s 集群已正常运行（单节点 / 多节点均可）：
   bash

   运行

   ```
   # 检查集群状态
   kubectl get nodes
   # 确保节点状态为Ready
   ```
2. 确保优化后的模型镜像已推送到 K8s 集群可访问的镜像仓库（如私有仓库 / 阿里云 ACR / 本地仓库）：

   * 本地集群（如 Minikube）：需在每个节点加载镜像
     bash

     运行

     ```
     docker save diabetes-api:v1.0-v2.0-20260321 -o diabetes-api.tar
     kubectl cp diabetes-api.tar <节点名称>:/tmp/
     # 进入节点加载镜像（Minikube可直接用minikube image load）
     minikube image load diabetes-api.tar
     ```
   * 远程集群：推送镜像到仓库（如阿里云 ACR）
     bash

     运行

     ```
     docker tag diabetes-api:v1.0-v2.0-20260321 registry.cn-hangzhou.aliyuncs.com/your-namespace/diabetes-api:v1.0-v2.0-20260321
     docker push registry.cn-hangzhou.aliyuncs.com/your-namespace/diabetes-api:v1.0-v2.0-20260321
     ```

### 二、K8s 配置文件编写（MLOps 实战版）

创建`k8s-deploy`目录，存放以下 2 个核心配置文件：

plaintext

```
k8s-deploy/
├── diabetes-deployment.yaml  # 部署配置（管理Pod）
└── diabetes-service.yaml     # 服务配置（暴露端口）
```

#### 1. Deployment 配置文件（diabetes-deployment.yaml）

核心作用：管理 Pod 的创建、副本数、镜像、资源限制（MLOps 需关注资源适配模型推理）

yaml

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: diabetes-api-deployment  # Deployment名称
  namespace: mlops               # MLOps专属命名空间（需提前创建）
  labels:
    app: diabetes-api
    mlops/model-version: v1.0
    mlops/api-version: v2.0
spec:
  replicas: 2  # 副本数（MLOps：高可用，至少2个）
  selector:
    matchLabels:
      app: diabetes-api
  strategy:
    rollingUpdate:  # 滚动更新（MLOps：无停机更新）
      maxSurge: 1
      maxUnavailable: 0
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: diabetes-api
        mlops/model-version: v1.0
        mlops/api-version: v2.0
    spec:
      containers:
      - name: diabetes-api-container
        # 替换为你的镜像地址（本地/私有仓库）
        image: docker.io/library/diabetes-api:v1.0-v2.0-20260321
        # 镜像拉取策略（本地镜像用Never，远程用Always）
        imagePullPolicy: IfNotPresent
        # 资源限制（MLOps：根据模型推理需求配置）
        resources:
          requests:  # 最小资源需求
            cpu: 500m
            memory: 512Mi
          limits:    # 最大资源限制
            cpu: 1000m
            memory: 1Gi
        # 端口配置（容器内端口，需和镜像内一致）
        ports:
        - containerPort: 9000
          name: api-port
          protocol: TCP
        # 健康检查（MLOps：关键，确保服务可用）
        livenessProbe:  # 存活探针（服务挂了重启Pod）
          httpGet:
            path: /health
            port: api-port
          initialDelaySeconds: 10  # 启动后10秒开始检查
          periodSeconds: 10        # 每10秒检查一次
          timeoutSeconds: 5        # 超时时间
        readinessProbe: # 就绪探针（Pod就绪后才接收流量）
          httpGet:
            path: /health
            port: api-port
          initialDelaySeconds: 5
          periodSeconds: 5
        # 环境变量（MLOps：配置日志/模型路径）
        env:
        - name: UVICORN_ACCESS_LOGS
          value: "false"
        - name: MODEL_PATH
          value: /app/pretrained_model.joblib
      # 镜像拉取密钥（私有仓库需配置）
      # imagePullSecrets:
      # - name: aliyun-acr-secret
```

#### 2. Service 配置文件（diabetes-service.yaml）

核心作用：暴露 Deployment 的 Pod，实现端口映射和稳定访问（ClusterIP/NodePort/LoadBalancer）

yaml

```
apiVersion: v1
kind: Service
metadata:
  name: diabetes-api-service
  namespace: mlops
  labels:
    app: diabetes-api
    mlops/model-version: v1.0
    mlops/api-version: v2.0
spec:
  type: NodePort  # 对外暴露（生产可用LoadBalancer/Ingress）
  selector:
    app: diabetes-api  # 关联Deployment的Pod标签
  ports:
  - name: api-port
    port: 9000          # Service内部端口
    targetPort: 9000    # 容器端口（和Deployment一致）
    nodePort: 30090     # 节点端口（范围30000-32767，固定便于访问）
    protocol: TCP
```

### 三、K8s 部署操作流程（MLOps 实战）

#### 1. 创建专属命名空间（MLOps：隔离资源）

bash

运行

```
kubectl create namespace mlops
```

#### 2. 部署 Deployment 和 Service

bash

运行

```
# 部署Deployment
kubectl apply -f diabetes-deployment.yaml

# 部署Service
kubectl apply -f diabetes-service.yaml

# 查看部署状态
kubectl get all -n mlops
```

#### 3. 验证部署结果

##### (1) 检查 Pod 状态（确保 Running）

bash

运行

```
kubectl get pods -n mlops
# 输出示例（2个副本均为Running）：
# NAME                                          READY   STATUS    RESTARTS   AGE
# diabetes-api-deployment-7f98d768c4-2x789      1/1     Running   0          50s
# diabetes-api-deployment-7f98d768c4-8k97s      1/1     Running   0          50s
```

##### (2) 检查 Service 状态（确认 NodePort）

bash

运行

```
kubectl get svc -n mlops
# 输出示例（NodePort为30090）：
# NAME                     TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
# diabetes-api-service     NodePort   10.96.123.45    <none>        9000:30090/TCP   1m
```

##### (3) 访问接口（验证服务可用）

bash

运行

```
# 方式1：通过NodePort访问（替换为你的节点IP）
curl http://<节点IP>:30090/health

# 方式2：通过ClusterIP在集群内访问
curl http://10.96.123.45:9000/health

# 正常返回健康检查JSON，说明部署成功
```

#### 4. 常用调试命令（MLOps 排障）

bash

运行

```
# 查看Pod日志（排查启动/推理错误）
kubectl logs -f <Pod名称> -n mlops

# 进入Pod内部（检查文件/依赖）
kubectl exec -it <Pod名称> -n mlops -- /bin/bash

# 查看Deployment详情（检查配置）
kubectl describe deployment diabetes-api-deployment -n mlops

# 扩容/缩容副本数（MLOps：弹性伸缩）
kubectl scale deployment diabetes-api-deployment --replicas=3 -n mlops
```

### 四、MLOps 场景下的 K8s 部署优化（进阶）

#### 1. 配置 Ingress（生产环境域名访问）

如果需要域名访问（如`api.mlops.com`），可添加 Ingress 配置：

yaml

```
# diabetes-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: diabetes-api-ingress
  namespace: mlops
spec:
  rules:
  - host: api.mlops.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: diabetes-api-service
            port:
              number: 9000
```

部署：`kubectl apply -f diabetes-ingress.yaml`

#### 2. 配置 HPA（自动弹性伸缩）

根据 CPU / 内存使用率自动扩缩容：

yaml

```
# diabetes-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: diabetes-api-hpa
  namespace: mlops
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: diabetes-api-deployment
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

部署：`kubectl apply -f diabetes-hpa.yaml`

### 总结

1. **核心配置**：Deployment 负责 Pod 的多副本管理（含健康检查、资源限制），Service 通过 NodePort 暴露端口（9000→30090），实现外部访问；
2. **MLOps 关键**：使用专属命名空间隔离资源、滚动更新保证无停机、健康检查确保服务可用、版本标签便于追溯；
3. **访问验证**：通过`节点IP:30090/health`访问健康检查接口，返回正常即部署成功。

按这套配置部署后，你的模型 API 服务将在 K8s 集群中实现高可用、可扩展的运行，完全贴合 MLOps 实战场景的部署要求。
