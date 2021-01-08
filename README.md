# codeci

使用 codeci 可以快速部署 k8s 服务, 包括服务间依赖分层部署 (生成依赖关系树, 循环依赖检测, 分组配置启动应用等)


## 服务依赖定义

服务依赖需要在 metadata.annotations.dependOn 声明依赖项(多个服务用逗号隔开) <br>
如:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-docs-backend
  annotations:
    dependOn: mariadb,etcd,redis,rabbitmq
spec:
  selector:
    matchLabels:
      app: api-docs-backend
...

```



