# LUSCSI - Lustre CSI Driver

## 简介
LUSCSI 是一个基于 [Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec) 的驱动程序，用于在 Kubernetes 集群中支持 Lustre 文件系统。通过此驱动程序，用户可以轻松地将 Lustre 文件系统挂载到 Kubernetes Pod 中，并实现动态存储卷的创建与管理。

## 功能特性
- **动态存储卷创建**：支持通过 CSI 接口动态创建 Lustre 存储卷。
- **参数化配置**：允许用户通过参数指定 MGS 地址、文件系统名称和子目录路径。
- **静态数据卷**：支持将 Lustre 指定目录挂载到 Pod 中，实现静态存储卷的创建与管理。
- **自定义数据卷名称**：允许用户根据 PVC 的命名空间和名称来定义 PV 名称。
- **数据卷 Quota**：支持设置数据卷的容量限制（Lustre 2.16.0 及以上版本支持）。
- **数据卷用量统计**：支持统计数据卷的使用情况。

## 核心概念
### 参数说明
以下参数用于配置 Lustre 卷：
- `mgsAddress`：MGS（Management Service）地址。
- `fsName`：Lustre 文件系统名称。
- `sharePath`：共享存储路径，用于在此目录下新建数据卷，默认为 /csi~volume (可选)。
> 注意：该路径（比如 /csi~volume）必须提前在 Lustre 文件系统中创建。

## 使用方法

### 1. 部署驱动程序
确保 Kubernetes 集群已安装 Helm 插件，并按照以下步骤部署 LUSCSI：

```bash
# 添加 Helm 仓库
helm repo add luscsi https://luskits.github.io/luscsi/

# 部署 luscsi
helm install luscsi luscsi/luscsi -n luscsi --create-namespace \
--set global.luscsiImageRegistry=ghcr.m.daocloud.io \
--set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
```

下面的诊断命令使用本次安装指定的 `luscsi` 命名空间。
如果安装时选择了其他命名空间，请相应调整诊断命令。

### 2. 创建 StorageClass
定义一个 StorageClass，指定所需的参数：

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: lustre-sc
provisioner: luscsi.luskits.io
parameters:
  mgsAddress: "example.mgs.address@o2ib"
  fsName: "lustrefs"
  sharePath: "/path/to/share" # 可选，默认为 /csi~volume
```

### 3. 动态创建 PVC
创建 PersistentVolumeClaim (PVC)，动态分配 Lustre 卷：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: lustre-pvc
spec:
  accessModes:
  - ReadWriteMany
  storageClassName: lustre-sc
  resources:
    requests:
      storage: 10Gi
  volumeMode: Filesystem
```

### 4. 挂载到 Pod
在 Pod 中挂载动态创建的卷：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: example-pod
spec:
  containers:
    - name: example-container
      image: nginx
      volumeMounts:
      - mountPath: "/mnt/lustre"
        name: lustre-volume
  volumes:
  - name: lustre-volume
    persistentVolumeClaim:
      claimName: lustre-pvc
```

## 开发指南

### 依赖安装
确保安装以下依赖：
- Go 1.23+
- Docker
- kubectl

运行以下命令安装依赖：
```bash
go mod tidy
```

### 调试
启用调试日志：
```bash
kubectl logs -n luscsi <driver-pod> -c luscsi
```


## 常见问题

### Q: 如何检查驱动程序是否正常运行？
A: 使用以下命令检查 CSI 驱动程序的状态：
    ```bash
    kubectl get pods -n luscsi | grep luscsi
    ```


### Q: 如何排查挂载失败的问题？
A: 查看驱动程序的日志，定位具体的错误信息：
```bash
kubectl logs -n luscsi <driver-pod> -c luscsi
```


## 贡献

欢迎提交 Issue 和 Pull Request！请遵循以下步骤：
1. Fork 仓库。
2. 创建新分支：`git checkout -b feature/new-feature`。
3. 提交更改并推送至远程分支。
4. 提交 Pull Request。
   
## 许可证

本项目采用 [Apache License 2.0](LICENSE) 发布。
