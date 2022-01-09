# lyyops-cicd

在Kubernetes集群中使用工作流引擎，实现自动化应用发布功能，基于:
- argo workflow
- argo CD

## 1. 准备环境


- go 1.16

```shell
GOROOT=/opt/soft/go
GOBIN=$GOROOT/bin
GOPATH=/opt/soft/gopath
GOPROXY=https://goproxy.cn,direct
```

- [argo workflow](https://argoproj.github.io/argo-workflows/)
- [argo cd](https://argo-cd.readthedocs.io/en/stable/)

## 2. 编译

```shell
cd lyyops-cicd
go build -o cicd
```

## 3. 运行

```shell
nohup ./cicd &
```

## 4. 示例

- [CICD 接口](http://192.168.101.211:9090/swagger/index.html)
- [Argo Workflow](https://192.168.101.211:30010/workflows/argo)
- [Argocd](http://192.168.101.211:30011/) (admin/80ur2Xn-o2doT2Vx)