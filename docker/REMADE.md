### 数据卷Vloume

#### 基本概念

- 是什么？

数据卷相当于是容器的虚拟文件系统和主机的真实文件系统之间的一个桥梁，建

立数据卷就相当于是打通了容器于主机之间的文件交互通道，可以让容器运行时

所产生的数据变更被保存到主机中，能够更方便的对数据进行备份以及保护

- 为什么需要？

当我们在容器中运行一些关键的应用如MySQL、Redis等服务，其中都会存储着

一些关键数据，这些数据是你希望即使删除容器也不应该被删除的，此时我们便

需要用到数据卷了。

- 能干什么？

你可以将数据卷理解为文件目录的映射，我们可以通过 Docker 提供的相关命

令，来将主机中的某一个文件目录映射到容器中，此时当你在容器中操作该目录

下的文件时，实际上操作的就是主机中的文件。

#### 绑定方式

- 匿名绑定

在启动容器时直接使用 -v /container\_dir 即可完成匿名绑定，匿名绑定的方式将

在 Docker 的 volumes 目录下生成一个 sha256 的字符串作为目录名，且指定的

/container\_dir 中的文件或目录会被保存在该处，匿名绑定的 volume 在容器被删

除的时候，数据卷也会被删除

```bash
docker run --rm -d -p 80: 80 -v /www/test nginx
```

匿名绑定方式由于不知道名称，因此如果需要查看数据卷在主机的哪个位置，需

要使用 docker inspect container\_id 来查看

- 具名绑定

同样是启动容器时绑定一个数据卷，不同的是可以为该数据卷起个名字 -v

volume-name:container\_dir，通过名字你可以快速的定位并管理这些 volume

```bash
docker run --rm -d -p 80:80 -v nginx-www:/www/test nginx
```

- Bind Mount

绑定并加载主机的某个文件目录到容器中，这种方式是平常最常用的。这种绑定

方式与前面两种一样，也是在容器启动时使用 -v host\_dir:container\_dir 的格式来

完成映射

```bash
docker run --rm -d -p 80: 80 -v /www/wolfcode:/www/wolfcode -v
```

/etc/nginx/nginx.conf:/etc/nginx/nginx.conf nginx数据卷管理

Docker 为我们提供了一些专门用于管理数据卷的命令 docker volume，通过下面的

Usage 来查看相关命令的使用

Usage:

docker volume command

Manage volumes

Commands:

create      创建一个数据卷

inspect     显示一个或多个数据卷的详细信息

ls          查看目前已有的数据卷列表

prune       删除所有本地没有被使用的数据卷

rm          删除一个或多个数据卷

### 网络 Network

#### 基本概念

- 是什么？

是 Docker 对容器网络隔离的一项技术，提供了多种不同的模式供用户使用，选择不同的网络模式来实现容器网络的互通以及彻底的隔离。

- 为什么需要？

容器间的网络隔离

实现部分容器之间的网络共享

管理多个子网下容器的 ip

- 能干什么？

提供了多种模式，可以定制化的为每个容器置顶不同的网络

自定义网络模式，划分不同的子网以及网关、dns等配置

网络互通

实现不同子网之间的网络互通

基于容器名（主机名）的方式在网络内访问

#### 网络模式

- bridge（桥接）

在主机中创建一个 Docker0 的虚拟网桥，在 Ddocker0 创建一对虚拟网卡，有一

半在主机上 vethxxx，还有一半在容器内 eth0

- 默认模式 host

容器不再拥有自己的网络空间，而是直接与主机共享网络空间，那么基于该模式

创建的容器对应的 ip 实际就是与主机同一个子网，同一个网段。

- none

Docker 会拥有自己的网络空间，不与主机共享，在这个网络模式下的容器，不会被分配网卡、ip、路由等相关信息。

特点：完全隔离，与外部任何机器都无网络访问，只有自己的 lo 本地网络127.0.0.1

- container

就是不会创建自己的网络空间，而是与其他容器共享网络空间，直接使用指定容器的ip/端口等

- 自定义（推荐）

不适用 Docker 自带的网络模式，而是自己去定制化自己特有的网络模式。

命令：

docker network command

### Dockerfile

#### 基本概念

- 是什么？

Docker 为我们提供的一个用于自定义构建镜像的一个配置文件：描述如何构建一个对象

利用 Docker 提供的 build 命令，指定 Dockerfile 文件，就可以按照配置的内容将镜像构建出来

- 为什么需要？

作为开发者需要将自己开发好的项目打包成 Docker 镜像，便于后面直接作为 Docker 容器运行作为运维人员需要构建更精简的基础设施服务镜像，满足公司的需求以及尽可能减少冗余的功能占用过多的资源

- 能干什么？

可以自定义镜像内容

构建公共基础镜像减少其他镜像配置

开源程序的快速部署

实现企业内部项目的快速交付

#### 常用指令

- FROM

指定以什么镜像作为基础镜像，在改进项的基础之上构建新的镜像。

如果不想以任何镜像作为基础：FROM scratch

语法：

FROM image<image>

FROM <image>:image:tag<tag>

FROM <image>:image:digest<digest>

以上为三种写法，后两者为指定具体版本，第一种则使用 latest 也就是最新版

- MAINTAINER

指定该镜像的作者

语法：

MAINTAINER name<name>

- LABEL

为镜像设置标签，一个 Dockerfile 中可以配置多个 LABEL

语法：

LABEL <key>key = value<value>

如：

LABEL "example.label"="Example Label"

LABEL label-value="LABEL"

LABEL version="1.0.0"

LABEL description="可以写成多行，使用 \ 符号可以拼接多行的 value"

PS：LABEL 指令会继承基础镜像（FROM）中的 LABEL，如果当前镜像 LABEL 的key 与其相同，则会将其覆盖

- ENV

设置容器的环境变量，可以设置多个

语法：

ENV <key> key value

ENV key=value key=value

两种语法的区别为第一种一次只能设置一个环境变量，第二种可以一次设置多个

- RUN

构建镜像的过程中要执行的命令

语法：

RUN command

RUN ["executable", "param1", "param2"]

第一种写法就是直接写 Shell 脚本即可

第二种写法类似函数调用，第一个参数为可执行文件，后面的都是参数

- ADD

复制命令，把 src 的文件复制到镜像的 dest 位置

语法：

ADD <src> dest

ADD ["src", "dest"]

- WORKDIR

设置工作目录，可以简单理解为 cd 到指定目录，如果该目录不存在会自动创建，对 RUN、CMD、ENTRYPOINT、COPY、ADD 生效，可以设置多次 WORKDIR

如：

WORKDIR dir

表示在容器内创建了 dir 目录，并且当前目录已经是 dir 目录了

- VOLUME

设置挂载目录，可以将主机中的指定目录挂载到容器中

语法：

VOLUME ["dir"]

VOLUME dir

VOLUME dir dir

以上三种写法都可

- EXPOSE

改镜像运行容器后，需要暴露给外部的端口，但仅仅表示该容器想要暴露某些端口，并不会与主机端口有映射关系，如果想将容器暴露的端口与主机映射则需要

使用 -p 或 -P 参数来映射，可以暴露多个端口

语法：

EXPOSE <port>[/<tcp/udp>]

- CMD

该镜像启动容器时默认执行的命令或参数

语法：

CMD ["executable", "param1", "param2"]

CMD ["param1", "param2"]

CMD command param1 param2

以上为该命令的三种写法，第三种与普通 Shell 命令类似，第一、二两种都是可执行文件 + 参数的形式，另外数组内的参数必须使用双引号。

案例：

第一种：CMD ["sh", "-c", "echo \$HOME"] 等同于 sh -c "echo \$HOME"

第二种：CMD ["echo", "\$HOME"] 等同于 echo \$HOME

- ENTRYPOINT

运行容器时的启动命令，感觉与 CMD 命令会很像，实际上还是有很大区别，简

单对比一下：

相同点：

在整个 Dockerfile 中只能设置一次，如果写了多次则只有最后一次生效

不同点：

ENTRYPOINT 不会被运行容器时指定的命令所覆盖，而 CMD 会被覆盖

如果同时设置了这两个指令，且 CMD 仅仅是选项而不是参数，CMD 中的内

容会作为 ENTRYPOINT 的参数（一般不这么做）

如果两个都是完整命令，那么只会执行最后一条

语法：

ENTRYPOINT ["executable", "param1", "param2"]

ENTRYPOINT command param1 param2

#### 拓展指令

- ARG

设置变量，在镜像中定义一个变量，当使用 docker build 命令构建镜像时，带上 --build-arg <name>=<value> 来指定参数值，如果该变量名在 Dockerfile中不存在则会抛出一个警告

语法：

ARG name[=default value]

- USER

设置容器的用户，可以是用户名或 UID，如果容器设置了以 daemon 用户去运

行，那么 RUN、CMD 和 ENTRYPOINT 都会以这个用户去运行，一定要先确定

容器中有这个用户，并且拥有对应的操作权限。

语法：

USER username

USER PID

- ONBUILD

表示在构建镜像时做某操作，不过不对当前 Dockerfile 的镜像生效，而是对以当前 Dockerfile 镜像作为基础镜像的子镜像生效

语法：

ONBUILD [INSTRUCTION]

例：

当前镜像为 A，设置了如下指令

ONBUILD RUN ls -al

镜像 B：

FROM 镜像A

......

构建镜像 B 时，会执行 ls -al 命令

- STOPSIGNAL

STOPSIGNAL 指令设置将发送到容器的系统调用信号以退出。此信号可以是与内核的系统调用表中的位置匹配的有效无符号数，例如 9，或 SIGNAME 格式的信号名，例如 SIGKILL。默认的stop-signal是SIGTERM，在docker stop的时候会给容器内PID为1的进程发送这个signal，通过--stop-signal可以设置自己需要的signal，主要的目的是为了让容器内的应用程序在接收到signal之后可以先做一些事情，实现容器的平滑退出，如果不做任何处理，容器将在一段时间之后强制退出，会造成业

务的强制中断，这个时间默认是10s。

- STOPSIGNAL <signal>HEALTHCHECK

容器健康状况检查，可以指定周期检查容器当前的健康状况，该命令只能出现一次，如果有多次则只有最后一次生效。

语法：

HEALTHCHECK [OPTIONS] CMD command

HEALTHCHECK NONE

第一种：在容器内部按照指定周期运行指定命令来检测容器健康状况

第二种：取消在基础镜像

OPTIONS 选项：

--interval=DURATION 两次检查的间隔时间，默认30s

--timeout=DURATION 命令执行的超时时间，默认30s

--retries=N 当连续失败指定次数，容器会被认定为不健康，默认为3次

返回参数：

0：success => 健康状态

1：unhealthy => 不健康状态

2：reserved => 保留值

#### 构建镜像

- commit

基于一个现有的容器，构建一个新的镜像

命令：

docker commit [OPTIONS] CONTAINER [REPOSITORY[:TAG]]

例：

docker commit -a="wolfcode" -m="first image" centos7 mycentos:7

OPTIONS：

-a：镜像的作者

-c：使用 Dockerfile 指令来构建镜像

-m：提交时的描述

-p：在 commit 时暂停容器

- build

基于一个 Dockerfile 构建镜像

语法：

docker build -t ImageName:TagName dockerfile dir

Spring Boot 镜像

Tomcat 镜像

Nginx 镜像
