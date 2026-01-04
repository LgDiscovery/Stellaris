### Operator = CRD + Controller

一个标准的CRD的 YAML 定义会是什么样的？

```yaml
 
apiVersion: apiextensions.k8s.io/v1
kind:CustomResourceDefinition
metadata:
# 这个名称的格式必须是: <CRD 的复数名称>.<group>
name:vehicles.iov.domain.io
spec:
# group 是我们自定义的，通常使用域名保持唯一性
group:iov.domain.io

# CRD 的种类，名称复数，名称单数，简称
# 这允许我们执行 `kubectl get vehicles/vehicle/vh` 这样的命令
names:
    kind:Vehicle
    plural:vehicles
    singular:vehicle
    shortNames:
    -vh

# Scope 定义了这个资源是 命名空间资源 还是 集群资源
scope:Namespaced

# CRD 可以定义多个版本
versions:
    -name:v1alpha1
      # 'served' 是否启用
      served:true
      # 'storage' 是否存储在 etcd 中，只能有一个版本被设置为 true
      storage:true

      # 这是构建健壮 Operator 的关键：启用 /status 子资源（下文详细解释）
      subresources:
        status:{}# 声明一个空的 status 子资源

      # 这里就是遵循了 OpenAPI v3 标准的 Schema 的定义了
      schema:
        openAPIV3Schema:
          type:object
          properties:
            # 这里的 spec 就是 期望状态 字段信息
            spec:
              type:object
              properties:
                firmwareVersion:
                  type:string
                  description:Thedesiredfirmwareversionforthisvehicle.
            # 这里的 status 就是 实际状态 字段信息
            status:
              type:object
              properties:
                phase:
                  type:string
                lastSeenTime:
                  type:string
                  format:date-time

```

### 关键字段剖析

1. `<span leaf="">group</span>`：设置 API 所属的组，会被映射为 `<span leaf="">/apis/</span>` 的下一级目录，这里是 `<span leaf="">/apis/iov.domain.io</span>`。强烈推荐使用**域名**，保持唯一性，避免命名冲突。
2. `<span leaf="">names</span>`：定义了 `<span leaf="">kind</span>`（Vehicle）、`<span leaf="">plural</span>`（vehicles）等。这让 Kubernetes 知道如何称呼你的资源，并允许 `<span leaf="">kubectl</span>` 通过多种名称（`<span leaf="">kubectl get vehicles</span>`、`<span leaf="">kubectl get vh</span>`）来查询它。
3. `<span leaf="">scope</span>`：设置 API 的生效范围，有两个可选项：`<span leaf="">Namespaced</span>` 和 `<span leaf="">Cluster</span>`。前者意味着 `<span leaf="">Vehicle</span>` CR 必须存在于一个命名空间中。后者意味着在集群范围内全局生效，不局限于任何命名空间（例如 Node、ClusterRole 等）。默认值是 `<span leaf="">Namespaced</span>`。
4. `<span leaf="">versions</span>`：设置此 CRD 支持的版本，可以设置多个版本。这允许我们的 API 随着时间演进。我们可以从 `<span leaf="">v1alpha1</span>` 开始，未来再添加 `<span leaf="">v1beta1</span>` 或 `<span leaf="">v1</span>`，并提供版本间的转换（conversion）。
5. `<span leaf="">subresources</span>`：**（重点）** 这是一个至关重要的设置。启用它（即使只是一个空对象 `<span leaf="">{}</span>`），`<span leaf="">kube-apiserver</span>` 就会为该资源额外暴露一个 `<span leaf="">/status</span>` 子路径。这允许 Controller（控制器）只更新 `<span leaf="">status</span>` 字段，而不会意外覆盖用户（或其它系统）设置的 `<span leaf="">spec</span>` 字段。这是实现健壮的声明式 API 的最佳实践。
6. `<span leaf="">schema.openAPIV3Schema</span>`：这是 CRD 的核心。它使用 OpenAPI v3 规范定义了 `<span leaf="">spec</span>` 和 `<span leaf="">status</span>` 中每个字段的**数据类型**、**描述**和**约束**（validation）。这就是为什么当你在 `<span leaf="">Vehicle</span>` CR 的 `<span leaf="">firmwareVersion</span>` 字段中写入一个数字（而不是字符串）时，`<span leaf="">kubectl apply</span>` 会立即拒绝你的请求——因为这个模式（schema）起到了校验作用。
