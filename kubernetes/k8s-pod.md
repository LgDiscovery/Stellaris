#### 如何查看pod和pod中的container的log

> Docker的世界里，我们通过
>
> docker logs container\_id|container\_name来获取container的日志，
>
> 类似的Kubernetes里，我们通过
>
> kubectl logs pod\_name来获取pod的日志，
>
> 如果1个pod里运行了多个container的话，我们则需要通过
>
> kubectl logs pod\_name -c container\_name来查看每个不同的container的日志。


#### 如何进入1个pod中不同container执行命令

> 通过kubectl exec pod\_name -c container\_name -it -- /bin/bash 格式的命令，
>
> exec 后面跟pod名；
>
> -c 表示pod中指定的container名；
>
> -it交互式的方式执行命令；
>
> -- 后面表示要执行的命令，这里是执行一个bash shell;注意--和命令之间要留空格。


#### 如何查看pod的annotation

> 可以通过kubectl describe pods pod_name 或者kubectl get pods pod_name -oyaml的方式查看


#### 如何给pod做annotation

> 可以通过kubectl annotate pod pod\_name annotate\_key=”annotation info”;来给pod贴上annotation。


#### annotation和label的对比

> 作用：label用于调度/选择、过滤管理维护pod；
>
> annotation用于描述对象自身的信息，如作者、版本号等，不能用于选择过滤对象；

> 范围：都可以用于pod也可以用于其它对象；比如可以给node分别贴label和annotation；

> 形式：都是key=value的形式；

> 查看：kubect get pods –show-labels ; kubectl get pods -oyaml;

> 长度：label存放内容比较短；annotation可以用于存储更长更多信息；
