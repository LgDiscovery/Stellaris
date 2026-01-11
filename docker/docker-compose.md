## ✅ 方法一：安装【Docker Compose V2 官方正式版】（首选，强烈推荐）

这是 Docker 官方现在主推的版本，**2025 年最新有效**，优点：

✅ 无需 Python 依赖，纯二进制文件，安装快、不冲突

✅ 版本更新及时，兼容性强，支持所有新特性

✅ 命令兼容：既可以用新版的 `docker compose up`，也可以用旧版的 `docker-compose up`

✅ 稳定性远超 V1 版本，无各种依赖报错

### 完整安装步骤（一步一步执行即可）

bash

运行

```
# 1. 下载 Docker Compose V2 最新稳定版二进制包（官方地址，自动适配x86_64架构）
sudo curl -SL https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64 -o /usr/libexec/docker/cli-plugins/docker-compose

# 2. 给二进制文件添加【可执行权限】（关键步骤，必加）
sudo chmod +x /usr/libexec/docker/cli-plugins/docker-compose

# 3. 【可选但推荐】创建软链接，兼容旧版 docker-compose 命令（不加也能用 docker compose）
sudo ln -s /usr/libexec/docker/cli-plugins/docker-compose /usr/local/bin/docker-compose
```

### 验证安装是否成功

两种命令都能验证，任选其一即可，出现版本号即代表安装成功：

bash

运行

```
# 新版原生命令（推荐使用）
docker compose version

# 兼容旧版命令（和 V1 用法完全一致）
docker-compose --version
```

成功示例输出：`Docker Compose version v2.24.6`（版本号会随最新版更新）

echo "alias ll='ls -l'" >> \~/.bashrc source \~/.bashrc


在 Docker 中，“加载镜像” 对应的是导入**已导出的镜像 tar 包**（通常由`docker save`生成），使用`docker load`命令实现，具体步骤如下：

### 步骤 1：准备镜像 tar 包

确保你已拥有导出的镜像包（比如`nginx.tar`、`pause.tar`，通常是通过`docker save -o 包名.tar 镜像名:标签`生成的）。

### 步骤 2：执行加载命令

在终端中运行以下命令（二选一即可）：

bash

运行

```
# 方法1：通过-i参数指定tar包路径
docker load -i /路径/到/镜像包.tar

# 方法2：通过重定向方式加载
docker load < /路径/到/镜像包.tar
```

示例（加载当前目录下的`pause.tar`）：

bash

运行

```
docker load -i ./pause.tar
```

### 步骤 3：验证镜像是否加载成功

加载完成后，查看本地镜像列表，确认目标镜像已存在：

bash

运行

```
docker images
```

输出中会显示加载后的镜像名、标签、ID 等信息。

### 注意事项

1. 若加载的镜像没有指定标签，可能显示为`<none>`，需手动打标签：
   bash

   运行

   ```
   docker tag 镜像ID 镜像名:标签
   ```
2. 镜像包需与当前系统架构匹配（比如 x86 架构的镜像不能直接加载到 ARM 架构的机器）
