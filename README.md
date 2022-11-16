# Open-Object

`Open-Object`将对象存储能力以 PV/PVC 方式提供给 K8s 集群内的应用使用，即将一个 Bucket 挂载到容器目录中，使得应用像访问本地文件一般（通过 read/write/stat 等 POSIX 接口）访问 Bucket 中的 Object。

`Open-Object`已广泛用于生产环境，目前使用的产品包括：

- [ACK 发行版](https://github.com/AliyunContainerService/ackdistro)
- 阿里云 ACK 敏捷版
- [云原生 CNStack 产品](https://github.com/alibaba/CNStackCommunityEdition)

## 特性

- 支持S3对象存储作为后端存储
- 存储卷动态分配
- 存储卷容量隔离
- 存储卷扩容
- 存储卷监控

## 开发

```bash
# clone repo
mkdir -p $GOPATH/src/github.com/alibaba/
cd $GOPATH/src/github.com/alibaba/
git clone https://github.com/alibaba/open-object.git
cd open-object
# build binary
make
# build image
make image
```

## 部署

```bash
# 安装 s3fs 依赖
# 需要在 k8s 集群每个节点上安装
yum update
yum install -y s3fs

cd $GOPATH/src/github.com/alibaba/open-object
# 编辑 values.yaml 中的 .minio 字段
# minio:
#   host: "http://10.96.2.217:9000"
#   accesskey: "minio"
#   secretkey: "miniostorage"
#   region: "china"
vi values.yaml
# 通过 helm 安装 open-object
helm install open-object helm/
```

## 许可证

[Apache 2.0 License](LICENSE)