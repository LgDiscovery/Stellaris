#### service资源对象的分类

> clusterIP: 默认类型的service资源对象；只能在kuberbetes cluster内部通过cluster IP +端口 来访问服务

> NodePort: clusterIP类型的service对象的超集，除了可以再kubernetes cluster内部通过 cluster IP + 端口访问之外，还可以通过整个集群中的每个中的每个node的ip+端口来访问服务

> loadbalancer: 它是NodePort类型service资源对象的超集，除了可以通过整个集群中的每个node的ip+端口来访问服务之外，还可以通过一个公网ip地址来访问服务

> ingress: 它是loadbalancer的超集，1个loadbalancer只能对应1个service资源对象；而1个ingress可以对应多个不同的服务
