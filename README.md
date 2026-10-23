# LUSCSI - Lustre CSI Driver

## Introduction
LUSCSI is a driver based on the [Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec)
that enables support for the Lustre file system in Kubernetes clusters. With this driver, users can easily mount Lustre file systems
to Kubernetes Pods and achieve dynamic storage volume creation and management.

## Features
- **Dynamic Volume Creation**: Supports dynamic creation of Lustre storage volumes via the CSI interface.
- **Parameterized Configuration**: Allows users to specify MGS address, file system name, and subdirectory path through parameters.
- **Static Data Volumes**: Supports mounting specific Lustre directories to Pods, enabling static storage volume creation and management.
- **Custom Volume Names**: Allows users to define PV names based on the namespace and name of the PVC.
- **Volume Quota**: Supports setting capacity limits for volumes (requires Lustre 2.16.0 or later).
- **Volume Usage Statistics**: Supports tracking the usage of storage volumes.

## Core Concepts
### Parameter Description
The following parameters are used to configure Lustre volumes:
- `mgsAddress`: The address of the MGS (Management Service).
- `fsName`: The name of the Lustre file system.
- `sharePath`: The shared storage path where new volumes will be created, defaulting to `/csi~volume` (optional).
> Note: The sharePath (e.g., `/csi~volume`) must be pre-created in the Lustre file system.

## Usage

### 1. Deploy the Driver
Ensure the Kubernetes cluster has the Helm plugin installed, and follow these steps to deploy LUSCSI:

```bash
# Add luscsi helm repo
helm repo add luscsi https://luskits.github.io/luscsi/

# Deploy luscsi using Helm
helm install luscsi luscsi/luscsi -n luscsi --create-namespace \
--set global.luscsiImageRegistry=ghcr.m.daocloud.io \
--set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io
```

The diagnostic commands below use the `luscsi` namespace from this installation.
If you install into a different namespace, use that namespace in those commands.

### 2. Create StorageClass
Define a StorageClass with the required parameters:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: lustre-sc
provisioner: luscsi.luskits.io
parameters:
  mgsAddress: "example.mgs.address@o2ib"
  fsName: "lustrefs"
  sharePath: "/path/to/share" # optional, defaults to /csi~volume
```
### 3. Dynamic Create PVC
Create a PersistentVolumeClaim (PVC), dynamically allocating a Lustre volume:

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

### 4. Mount to Pod
Mount the dynamically created volume to a Pod:

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

## Development Guide
## Dependencies
Ensure you have the following dependencies installed:
- Go 1.23+
- Docker
- kubectl

Run the following command to install dependencies:
```bash
go mod tidy
```

### Debug
Enable debug logs:
```bash
kubectl logs -n luscsi <driver-pod> -c luscsi
```
## FAQ
### Q: How to check if the driver is running correctly?
A: Use the following command to check the status of the CSI driver:

```bash
kubectl get pods -n luscsi | grep luscsi
````

### Q: How to troubleshoot mount failures?
A: Check the logs of the driver pod for any error messages:

```bash
kubectl logs -n luscsi <driver-pod> -c luscsi
```
