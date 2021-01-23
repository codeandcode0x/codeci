# codeci

使用 codeci 可以快速部署 k8s 服务, 包括服务间依赖分层部署 (生成依赖关系树, 循环依赖检测, 分组配置启动应用等)


## 服务依赖定义

服务依赖需要在 **metadata.annotations.dependOn** 声明依赖项(多个服务用逗号隔开) <br>

如:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: serviceA
  annotations:
    dependOn: serviceB, serviceC
spec:
  selector:
    matchLabels:
      app: serviceA
...

```
codeci deploy serviceA  进行部署 serviceA 的时候会检测 serviceB、C 是否正常启动，如果没有则先启动 serviceB、C

## 安装

```sh
curl -fsSl https://codeandcode.cn/apps/codeci/install.sh | bash
```

## 使用
修改配置文件
```sh
vim ~/.codeci/deployconfig
```

2. cli 命令

部署服务
```
codeci deploy [service name]
```

删除服务
```
codeci reset [service name]
```

分析服务
```
codeci analyze [service name]
```




