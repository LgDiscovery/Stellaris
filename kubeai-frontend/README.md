# KubeAI 前端项目

> 基于 React + TypeScript + Vite 的企业级云原生 AI 平台管理控制台

## 项目简介

KubeAI 前端是 [KubeAI 云原生 AI 平台](https://github.com/LgDiscovery/kubeai) 的官方控制台界面，提供模型管理、离线训练、在线推理的全链路可视化操作。支持 JWT 认证、基于 RBAC 的权限控制、实时任务监控与服务健康检查，旨在帮助算法工程师和运维人员高效管理 AI 工作负载。

## 技术栈

表格


| 类别     | 技术选型                          |
| -------- | --------------------------------- |
| 框架     | React 18 + TypeScript             |
| 构建工具 | Vite 5                            |
| UI 库    | Ant Design 5 + @ant-design/icons  |
| 路由     | React Router v6                   |
| HTTP     | Axios（拦截器 + JWT 自动注入）    |
| 状态管理 | React Context（Auth）             |
| 样式     | CSS Modules + Ant Design 主题定制 |
| 代理     | Vite 内置 proxy（开发环境）       |
| 部署     | Nginx + Docker + Kubernetes       |

## 目录结构

```
kubeai-frontend/
├── index.html                # 入口 HTML
├── package.json              # 项目依赖与脚本
├── tsconfig.json             # TypeScript 配置
├── vite.config.ts            # Vite 配置（含代理）
├── deploy/                   # 部署资源
│   ├── Dockerfile            # 多阶段构建镜像
│   ├── nginx.conf            # Nginx 配置模板（生产）
│   ├── configmap.yaml        # Nginx 配置的 ConfigMap
│   ├── deployment.yaml       # Kubernetes Deployment
│   └── service.yaml          # Kubernetes Service (NodePort:39090)
└── src/
    ├── main.tsx              # 应用入口
    ├── App.tsx               # 路由定义
    ├── vite-env.d.ts         # Vite 类型声明
    ├── api/                  # API 接口封装
    │   ├── axios.ts          # Axios 实例与拦截器
    │   ├── auth.ts           # 认证接口（登录/注册）
    │   ├── models.ts         # 模型管理接口
    │   ├── inference.ts      # 推理服务接口
    │   └── jobs.ts           # 训练/推理任务接口
    ├── contexts/
    │   └── AuthContext.tsx   # 全局认证上下文
    ├── components/
    │   └── ProtectedRoute.tsx # 路由守卫
    ├── layouts/
    │   └── MainLayout.tsx    # 主框架布局（侧边栏+头部）
    ├── pages/
    │   ├── Login.tsx         # 登录页
    │   ├── Register.tsx      # 注册页
    │   ├── Dashboard.tsx     # 仪表盘（总览）
    │   ├── models/
    │   │   ├── ModelList.tsx # 模型列表
    │   │   ├── ModelDetail.tsx # 模型详情+版本管理
    │   │   ├── ModelCreate.tsx # 创建模型
    │   │   └── VersionCreate.tsx # 添加版本
    │   ├── inference/
    │   │   └── InferenceExecute.tsx # 执行推理
    │   ├── jobs/
    │   │   ├── JobList.tsx   # 任务列表（训练/推理）
    │   │   ├── TrainingJobCreate.tsx # 提交训练任务
    │   │   └── InferenceJobCreate.tsx # 提交推理任务
    │   └── health/
    │       └── HealthCheck.tsx # 服务健康检查
    └── styles/
        └── global.css        # 全局样式
```

## 快速开始

### 环境要求

* Node.js >= 18.x
* npm >= 9.x

### 安装与启动

```
# 克隆仓库（假设已获取前端代码）
cd kubeai-frontend

# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:3000）
npm run dev
```

开发环境下，所有以 /api 开头的请求将被自动代理到 [http://localhost:8080](http://localhost:8080)（API 网关默认地址）。

如需修改后端地址，请编辑 vite.config.ts 中的 server.proxy 配置。

### 构建与部署

#### 本地构建

```
npm run build        # 输出到 dist/ 目录
npm run preview      # 预览构建结果
```

#### Docker 镜像构建

```bash
docker build -f deploy/Dockerfile -t kubeai-frontend:latest .
```

#### Kubernetes 部署

项目中已提供完整的部署资源文件，位于 deploy/ 目录。

1. 创建命名空间（若不存在）

```
kubectl create namespace kubeai
```

2. 部署 ConfigMap（动态 Nginx 配置）

```
kubectl apply -f deploy/configmap.yaml
```

3. 部署应用

```
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/service.yaml
```

4. 查看状态

```
kubectl get pods,svc -n kubeai -l app=kubeai-frontend
```

#### 访问

部署成功后，可通过集群任意节点的 IP + 39090 端口访问控制台：

```
http://<任意节点IP>:39090
```

默认账号密码请参见 KubeAI 后端部署文档（例如：postgres /postgres）。

### Nginx 配置动态更新

生产环境使用 ConfigMap 挂载 Nginx 配置文件，无需重新构建镜像即可修改后端代理地址等设置。

1. 修改 ConfigMap：

```
kubectl edit configmap kubeai-frontend-nginx -n kubeai
```

2. 滚动重启 Pod 生效：

```
kubectl rollout restart deployment/kubeai-frontend -n kubeai
```

### 健康检查

Kubernetes 探针配置于 /health 端点，返回 200。

开发模式下可通过访问以下地址手动检查：

* 前端：[http://localhost:3000/health](http://localhost:3000/health)
* 后端（需后端运行）：
  * API 网关：/api/v1/auth/metrics
  * 模型管理：/api/v1/models/health
  * 任务调度：/api/v1/jobs/health
  * 推理网关：/api/v1/inference/health

## 环境变量与配置

表格


| 配置项               | 说明                 | 默认值                                                                                               |
| -------------------- | -------------------- | ---------------------------------------------------------------------------------------------------- |
| Vite 代理目标        | 后端 API 网关地址    | [http://localhost:8080](http://localhost:8080)                                                       |
| 后端内部域名（生产） | Nginx 反向代理地址   | [http://api-gateway.kubeai.svc.cluster.local:8080](http://api-gateway.kubeai.svc.cluster.local:8080) |
| NodePort             | K8s Service 暴露端口 | 39090                                                                                                |
| 鉴权 Token 存储 Key  | localStorage 键名    | kubeai\_token                                                                                        |

## 开发约定

1. 使用 ESLint + TypeScript 严格模式
2. API 接口封装遵循 “一个模块一个文件” 原则（src/api/\*.ts）
3. 所有路径引用使用相对导入
4. 路由守卫通过 ProtectedRoute 组件实现未登录重定向
5. 功能页面统一放置在 src/pages/<模块>/ 下

## 常见问题

**Q: 开发模式下代理不生效？**

A: 请检查 vite.config.ts 中的 target 是否正确，并确保后端 API 网关已在对应端口启动。

**Q: 页面打开后空白或路由 404？**

A: 确认 nginx.conf 中配置了 try\_files \$uri \$uri//index.html，以支持 SPA 路由。

**Q: 如何更换后端地址？**

A: 开发环境修改 vite.config.ts 的代理；生产环境修改 deploy/configmap.yaml 中的 proxy\_pass 地址，然后重启 Pod。

## 相关链接

* KubeAI 后端仓库：[https://github.com/LgDiscovery/kubeai](https://github.com/LgDiscovery/kubeai)
* Ant Design 官方文档：[https://ant.design/](https://ant.design/)
* Vite 官方文档：[https://vitejs.dev/](https://vitejs.dev/)

## License

本项目遵循与 KubeAI 主仓库相同的开源协议。
