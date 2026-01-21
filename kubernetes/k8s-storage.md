# K8s 运维必备知识：Pod 持久化存储（PV、PVC、StorageClass）

**原创** **SRE运维小张** [SRE运维小张](javascript:void(0);)

*2026年1月19日 11:46* *湖北*

![](http://mmbiz.qpic.cn/sz_mmbiz_png/fxUeQg4XWEXXf9nL3CHGKx7CZP9G9zSQnlDh6YBhj8h6AzlkU4lfb7PpcaOo7Pz4DTmlcMBkFUibcksxfgQOr0g/300?wx_fmt=png&wxfrom=19)

**SRE运维小张**

10年linux运维，分享k8s、docker、devops等技术，“不断学习” 是运维这行的基本生存法则，信奉 “日拱一卒”：今天比昨天多懂一个命令，每周比上周少踩一个坑，关注我，一起在云原生的路上慢慢 “卷”～

**33篇原创内容**

公众号

k8s运维，如何保证pod故障数据不丢失，pod重启后能恢复之前的数据状态，那就必须把数据持久化到磁盘上，这篇文章介绍k8s的几种持久化存储方式和使用方法。

---

一、持久化基础

k8s 持久化存储的核心围绕「**存储生命周期独立于 Pod**」展开，主要分为静态持久化（PV+PVC）**和**动态持久化（StorageClass+PVC）两大类

明确三个核心组件的关系：

1. **PV（Persistent Volume）：持久化存储卷，由管理员提前创建，是集群中的 “存储资源”（对应实际磁盘 / 共享目录），生命周期独立于 Pod。**
2. **PVC（**Persistent Volume Claim**）：持久化存储申请，由用户创建，是 “存储资源申请单”，用于申请 PV 的存储大小、访问模式等。**
3. **StorageClass：存储类，用于动态创建 PV 的 “模板”，解决静态 PV 手动创建的运维痛点，是生产环境首选。**

---

## 二、静态持久化存储（PV+PVC）

## 

### 1. 核心原理

* 管理员**手动提前创建 PV**（绑定实际的存储介质，如宿主机目录、NFS 共享目录等），PV 中定义存储大小、访问模式、回收策略等属性。
* 用户创建**PVC（存储申请）**，k8s 会自动匹配满足条件的 PV（大小≤PV 容量、访问模式一致、存储类匹配），匹配成功后二者绑定（Bound 状态）。
* Pod 通过引用 PVC，间接挂载对应的 PV 存储，实现数据持久化，Pod 销毁后，PV/PVC 及数据仍保留，可被其他 Pod 复用。

核心特点：**手动创建 PV、静态匹配、适用于存储资源固定的场景**，测试环境或小规模集群常用。

![图片](https://mmbiz.qpic.cn/sz_mmbiz_png/fxUeQg4XWEUNVaYnucEEMc30tQM9iclx6dx6Ma6yLlWT5dNibJRLbu2HT6icuINcxb0gryb4iaSk6zoDOmtILXDtlA/640?wx_fmt=png&from=appmsg&watermark=1&tp=wxpic&wxfrom=5&wx_lazy=1#imgIndex=0)

### 2. 具体实现

## 🛠️ 场景 1：测试环境（hostPath 类型 PV，单节点集群）

`<span leaf="">hostPath</span>`将宿主机的本地目录 / 磁盘映射为 PV，仅支持单节点，节点故障数据丢失，**生产环境数据不重要可使用**。

##### 👉 步骤 1：创建 hostPath 类型 PV（手动提供存储资源）

```
# demo-hostpath-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: hostpath-demo-pv # PV名称
spec:
  capacity:
    storage: 10Gi # PV存储容量
  accessModes:
  - ReadWriteOnce # 访问模式：单节点可读写（RWO）
  hostPath:
    path: /data/k8s/hostpath-pv # 宿主机上的实际目录（需提前创建：mkdir -p /data/k8s/hostpath-pv）
    type: DirectoryOrCreate # 目录不存在则自动创建
  persistentVolumeReclaimPolicy: Retain # 回收策略：保留数据（Pod删除后数据不丢失）
  storageClassName: ""  # 存储类留空，用于静态匹配
```

👉 步骤 2：创建 PVC（申请存储资源）

```
# demo-hostpath-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: hostpath-demo-pvc # PVC名称
spec:
  resources:
    requests:
      storage: 5Gi # 申请容量（需≤PV的10Gi）
  accessModes:
  - ReadWriteOnce # 访问模式必须与PV匹配
  storageClassName: ""  # 与PV存储类一致（空字符串）
```

👉 步骤 3：创建 Pod，挂载 PVC 实现持久化

```
# demo-hostpath-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: hostpath-demo-pod
spec:
  containers:
  - name: nginx-demo
    image: nginx:alpine
    volumeMounts:
    - name: persistent-storage # 与下方volumes名称对应
      mountPath: /usr/share/nginx/html # 容器内挂载路径
  volumes:
  - name: persistent-storage
    persistentVolumeClaim:
      claimName: hostpath-demo-pvc # 引用创建好的PVC
```

👉 步骤 4：操作与验证（持久化效果）

依次创建资源

```
kubectl apply -f demo-hostpath-pv.yaml
kubectl apply -f demo-hostpath-pvc.yaml
kubectl apply -f demo-hostpath-pod.yaml
```

验证 PVC 绑定状态（`<span leaf="">Bound</span>`即为成功）：

```
kubectl get pvc
```

进入 Pod 创建测试文件，验证数据持久化

```
# 进入容器创建测试文件
kubectl exec -it hostpath-demo-pod -- touch /usr/share/nginx/html/test-hostpath.txt
# 查看宿主机目录，确认文件同步存在
ls /data/k8s/hostpath-pv/
# 删除Pod，验证数据不丢失
kubectl delete pod hostpath-demo-pod
# 重新创建Pod，进入容器查看文件是否存在
kubectl apply -f demo-hostpath-pod.yaml
kubectl exec -it hostpath-demo-pod -- ls /usr/share/nginx/html/
```

## 🛠️ 场景 2：生产环境（NFS 类型 PV，多节点集群）

`<span leaf="">NFS</span>`（网络文件系统）是多节点 k8s 集群的常用持久化方案，通过共享网络目录实现跨节点数据共享，Pod 可在不同节点挂载同一 NFS 存储，**支持生产环境使用**。

##### 前置准备：搭建 NFS 服务器（单台服务器作为 NFS 服务端，所有 k8s 节点作为客户端）

1. NFS 服务端（以 CentOS 为例）：

```
# 1. 安装NFS服务
yum install -y nfs-utils rpcbind
# 2. 创建NFS共享目录
mkdir -p /data/k8s/nfs-share
# 3. 配置NFS共享权限（允许所有k8s节点访问）
echo "/data/k8s/nfs-share *(rw,sync,no_root_squash,no_all_squash)" >> /etc/exports
# 4. 启动服务并设置开机自启
systemctl start rpcbind nfs-server
systemctl enable rpcbind nfs-server
# 5. 生效配置
exportfs -rv
```

所有 k8s 节点（NFS 客户端）安装依赖：

```
yum install -y nfs-utils
```

👉 步骤 1：创建 NFS 类型 PV（手动绑定 NFS 共享目录）

```
# demo-nfs-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-demo-pv
spec:
  capacity:
    storage: 20Gi # NFS共享目录提供的存储容量
  accessModes:
  - ReadWriteMany # 访问模式：多节点可读写（RWX，NFS核心优势）
  nfs:
    server: 192.168.1.100 # 替换为你的NFS服务端IP
    path: /data/k8s/nfs-share # NFS服务端的共享目录
  persistentVolumeReclaimPolicy: Retain
  storageClassName: ""
```

👉 步骤 2：创建 PVC（申请 NFS 存储资源）

```
# demo-nfs-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-demo-pvc
spec:
  resources:
    requests:
      storage: 10Gi # 申请容量≤PV的20Gi
  accessModes:
  - ReadWriteMany # 与PV访问模式匹配
  storageClassName: ""
```

👉 步骤 3：创建 Pod，挂载 NFS 类型 PVC

```
# demo-nfs-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nfs-demo-pod
spec:
  containers:
  - name: nginx-demo
    image: nginx:alpine
    volumeMounts:
    - name: nfs-storage
      mountPath: /usr/share/nginx/html
  volumes:
  - name: nfs-storage
    persistentVolumeClaim:
      claimName: nfs-demo-pvc
```

##### 👉 步骤 4：操作与验证（跨节点持久化）

1. 依次创建资源，验证 PVC 绑定状态：

```
kubectl apply -f demo-nfs-pv.yaml
kubectl apply -f demo-nfs-pvc.yaml
kubectl apply -f demo-nfs-pod.yaml
kubectl get pvc
```

验证跨节点持久化：* 在当前 Pod 创建测试文件，然后删除 Pod。

* 在另一台 k8s 节点上重新创建该 Pod（可通过`<span leaf="">nodeName</span>`指定节点），进入容器后可看到测试文件仍存在。
* 直接查看 NFS 服务端`<span leaf="">/data/k8s/nfs-share</span>`目录，文件同步保留。
*

---

## 三、动态持久化存储（StorageClass+PVC）

## 

### 1. 核心原理

* 管理员**提前创建 StorageClass（存储类）**，作为动态创建 PV 的 “模板”，其中定义了存储介质类型（NFS/Ceph/ 云盘）、存储大小、回收策略、存储提供者等配置。
* 用户创建 PVC，在 PVC 中**引用该 StorageClass**，无需管理员手动创建 PV。
* k8s 通过 StorageClass 关联的 “存储供应器”（Provisioner），**自动创建符合 PVC 要求的 PV**，并完成 PVC 与 PV 的绑定。
* Pod 挂载 PVC，间接使用动态创建的 PV，实现数据持久化。

  核心特点：**自动创建 PV、动态供应、运维效率高**，是**生产环境的首选方案**，解决了静态 PV 手动创建、资源浪费的问题。

![图片](https://mmbiz.qpic.cn/sz_mmbiz_png/fxUeQg4XWEUNVaYnucEEMc30tQM9iclx6AGxe5pIM9GTA3TbyIH5U4RWMCHPYleLU3JxLLOxr89y1ctSWP1NOKA/640?wx_fmt=png&from=appmsg&watermark=1&tp=wxpic&wxfrom=5&wx_lazy=1#imgIndex=1)

### 2. 具体实现（以 NFS 动态供应为例，多节点生产环境）

k8s 原生不支持 NFS 动态供应，需要部署`<span leaf="">nfs-subdir-external-provisioner</span>`插件（存储供应器），用于关联 NFS 服务器并自动创建 PV。

#### 前置准备

1. 已完成 NFS 服务器搭建（复用前文的 NFS 服务端 / 客户端配置）。
2. 部署 nfs-subdir-external-provisioner 插件（以 YAML 方式部署，也可使用 Helm）：

```
# nfs-provisioner.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-client-provisioner
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: nfs-client-provisioner-runner
rules:
- apiGroups: [""]
  resources: ["persistentvolumes"]
verbs: ["get", "list", "watch", "create", "delete"]
- apiGroups: [""]
  resources: ["persistentvolumeclaims"]
verbs: ["get", "list", "watch", "update"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["events"]
verbs: ["create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: run-nfs-client-provisioner
subjects:
- kind: ServiceAccount
  name: nfs-client-provisioner
  namespace: default
roleRef:
  kind: ClusterRole
  name: nfs-client-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nfs-client-provisioner
  namespace: default
spec:
  replicas:1
  selector:
    matchLabels:
      app: nfs-client-provisioner
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: nfs-client-provisioner
    spec:
      serviceAccountName: nfs-client-provisioner
      containers:
      - name: nfs-client-provisioner
        image: registry.cn-hangzhou.aliyuncs.com/google_containers/nfs-subdir-external-provisioner:v4.0.2
        volumeMounts:
        - name: nfs-client-root
          mountPath: /persistentvolumes
        env:
        - name: PROVISIONER_NAME
          value: k8s-sigs.io/nfs-subdir-external-provisioner # 供应器名称，后续StorageClass需引用
        - name: NFS_SERVER
          value: 192.168.1.100  # 替换为你的NFS服务端IP
        - name: NFS_PATH
          value: /data/k8s/nfs-share # NFS服务端共享目录
      volumes:
      - name: nfs-client-root
        nfs:
          server: 192.168.1.100
          path: /data/k8s/nfs-share
```

部署插件：

```
kubectl apply -f nfs-provisioner.yaml
# 验证插件运行状态（Running即为成功）
kubectl get pods | grep nfs-client-provisioner
```

👉 步骤 1：创建 StorageClass（动态 PV 模板）

```
# demo-nfs-storageclass.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfs-dynamic-sc # StorageClass名称，后续PVC需引用
provisioner: k8s-sigs.io/nfs-subdir-external-provisioner # 与nfs-provisioner的PROVISIONER_NAME一致
parameters:
  archiveOnDelete: "true"  # 删除PVC时，是否归档PV数据（保留数据，避免丢失）
reclaimPolicy: Retain # 动态PV的回收策略
allowVolumeExpansion: true  # 允许扩容存储容量
```

👉 步骤 2：创建 PVC（引用 StorageClass，自动申请 PV）

```
# demo-nfs-dynamic-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-dynamic-pvc
spec:
  resources:
    requests:
      storage: 8Gi # 申请存储容量（由StorageClass自动创建对应大小的PV）
  accessModes:
  - ReadWriteMany # 支持多节点可读写
  storageClassName: nfs-dynamic-sc # 引用创建好的StorageClass（核心：实现动态PV）
```

👉 步骤 3：创建 Pod，挂载动态 PVC

```
# demo-nfs-dynamic-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nfs-dynamic-pod
spec:
  containers:
  - name: nginx-demo
    image: nginx:alpine
    volumeMounts:
    - name: nfs-dynamic-storage
      mountPath: /usr/share/nginx/html
  volumes:
  - name: nfs-dynamic-storage
    persistentVolumeClaim:
      claimName: nfs-dynamic-pvc
```

#### 👉 步骤 4：操作与验证（动态 PV 创建）

1. 依次创建资源：

```
kubectl apply -f demo-nfs-storageclass.yaml
kubectl apply -f demo-nfs-dynamic-pvc.yaml
kubectl apply -f demo-nfs-dynamic-pod.yaml
```

验证动态 PV 创建（k8s 自动创建 PV，名称以`<span leaf="">pvc-</span>`开头）：

```
kubectl get pv # 可看到状态为Bound的PV，容量8Gi
kubectl get pvc # 状态为Bound，已成功绑定动态PV
```

1. 验证持久化效果：

   * 进入 Pod 创建测试文件，删除 Pod 后重新创建，文件仍存在。
   * 查看 NFS 服务端`<span leaf="">/data/k8s/nfs-share</span>`目录，可看到 k8s 自动创建的子目录（对应动态 PV），测试文件存储在该目录中。

### 

---

### 

## 四、其他常见持久化存储方式

## 

1. **Ceph RBD/CSI：分布式存储系统，将数据分散存储在多个节点的磁盘中，提供高可用、高扩展性的存储服务，通过 CSI 插件与 k8s 集成。**

* 适用场景：大规模生产集群、对数据可靠性要求极高的场景（如核心数据库、大数据存储）。

1. **GlusterFS：分布式文件系统，无中心节点，通过集群节点的本地存储组成共享存储池，支持 RWX 访问模式。**

* 适用场景：中小规模生产集群、需要跨节点共享文件的场景。


## 五、总结

## 

1. k8s 持久化存储核心分为**静态（PV+PVC）**和**动态（StorageClass+PVC）**，动态存储是生产环境首选，提升运维效率。
2. 存储介质分**本地存储（hostPath，仅测试）**和**网络存储（NFS/Ceph/ 云盘，生产）**，网络存储支持跨节点共享，保障高可用。
3. 核心操作流程：「提供存储资源（静态 PV / 动态 StorageClass）→ 申请存储（PVC）→ Pod 挂载 PVC」，三者的访问模式、存储类必须匹配。
4. 生产环境关键配置：选择支持`<span leaf="">ReadWriteMany（RWX）</span>`的存储介质、开启存储扩容、设置`<span leaf="">Retain</span>`回收策略保障数据安全。
