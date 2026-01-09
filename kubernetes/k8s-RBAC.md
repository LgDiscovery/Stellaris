## 01 RBAC是什么？解决了什么问题

RBAC,全称 基于角色的访问控制，解决的是：

> 谁 who 可以对什么资源 what 做什么操作 verb

在 K8s 中

1 谁 ：user  / ServiceAccount

2 什么资源：Pod Deployment service ConfigMap

3 做什么：get/list/watch/creat/update/delete

## 02 RBAC的4个核心对象

① Role / ClusterRole ---权限规则本身

定义【能干什么】

```yaml
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get","list"]

```

区别

Role    仅限某个Namespace     ClusterRole  整个集群


② RoleBinding / ClusterRoleBinding —— 授权关系

定义 【把权限给谁】

```yaml
subjects:
- kind: ServiceAccount
  name: app-sa
roleRef:
  kind: Role
  name: pod-reader
```

③ Subject —— 被授权的对象

可以是：

User (外部用户)  Group (用户组) ServiceAccount(Pod使用)

生产环境90%是ServiceAccount

④ Verb —— 操作动作

最常见的：

* 读：get / list / watch
* 写：create / update / patch
* 删：delete

## 03 一个完整 RBAC 授权流程

用一句话串起来：

Role 定义权限 → Binding 绑定对象 → Subject 获得权限

示意：

```

ServiceAccount
      ↓
RoleBinding
      ↓
Role / ClusterRole
```

## 04实战1：只允许Pod读取Pod信息（Namespace级）

setp1: 创建Role

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
 name: pod-reader
 namespace: dev
rules:
- apiGroups: [""]
  resources: ["pods"]
  vrebs: ["get","list"]
```

step2: 创建ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: dev
```

step3: 绑定 Role

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
 name: read-pods
 namespace: dev
subjects:
- kind: ServiceAccount
  name: app-sa
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io

```

## 05实战2： 集群只读权限（ClusterRole）

ClusterRole

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
 name: view-all
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get","list","watch"]
```

ClusterRoleBinding

```yaml

apiVersion: rbac.authorization.k8s.io/v1
metadata:
 name: view-all-binding
subjects:
- kind: User
  name: dev-user
roleRef:
  kind: ClustrRole
  name: view-all

```
