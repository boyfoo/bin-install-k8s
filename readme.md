

### 创建统一管理文件夹
`sudo mkdir /usr/k8s`
### 复制etcd执行文件
`sudo cp /vagrant/etcd-v3.4.14/etcd /vagrant/etcd-v3.4.14/etcdctl /usr/k8s/`
### 复制ssl工具
```
chmod +x /vagrant/ssltool/*
sudo mkdir /usr/cfssl
cd /vagrant/ssltool/
sudo cp cfssljson_linux-amd64 cfssl_linux-amd64 cfssl-certinfo_linux-amd64 /usr/cfssl/
cd /usr/cfssl/
sudo mv cfssl-certinfo_linux-amd64 cfssl-certinfo
sudo mv cfssljson_linux-amd64 cfssljson
sudo mv cfssl_linux-amd64 cfssl
# 修改环境变量
sudo vim /etc/profile
# 最后一行加入
export K8SBIN=/usr/k8s
export CFSSLBIN=/usr/cfssl
export PATH=$PATH:$K8SBIN
export PATH=$PATH:$CFSSLBIN
# 刷新环境变量
source /etc/profile
```


# mater:
## ETCD证书
### 创建证书临时文件夹
```
sudo mkdir -p ~/certs/{etcd,k8s}
cd ~/certs/etcd/
```
### 新增ca证书
```
cat > ca-config.json << EOF
{
  "signing": {
    "defaulf": {
      "expiry": "87600h"
    },
    "profiles": {
      "www": {
        "expiry": "87600h",
        "usages": [
          "signing",
          "key encipherment",
          "server auth",
          "client auth"
        ]
      }
    }
  }
}
EOF
```
### 新增证书请求文件
```
cat > ca-csr.json << EOF
{
  "CN": "etcd CA",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "Beijing",
      "ST": "Beijing"
    }
  ]
}
EOF
```
### 生成key和证书
`cfssl gencert -initca ca-csr.json | cfssljson -bare ca -`
### https证书请求文件 用于etcd对外https证书
```
cat > server-csr.json << EOF
{
  "CN": "etcd",
  "hosts": [
    "192.168.33.10",
    "192.168.33.11"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "Beijing",
      "ST": "Beijing"
    }
  ]
}
EOF
```
### 用ca正式去签名
`cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=www server-csr.json | cfssljson -bare server`
### 创建etcd配置和证书目录 和 etcd储存目录
```
sudo mkdir -p /etc/k8s/etcd/{config,certs}
sudo mkdir /var/lib/etcd
```
### 新增单机etcd启动文件
`sudo vi /etc/k8s/etcd/config/etcd.conf `
### 贴如下内容
```
#[Member]
ETCD_NAME="etcd1"
ETCD_DATA_DIR="/var/lib/etcd"
ETCD_LISTEN_PEER_URLS="https://192.168.33.10:2380"
ETCD_LISTEN_CLIENT_URLS="https://192.168.33.10:2379"
#[Clustering]
ETCD_INITIAL_ADVERTISE_PEER_URLS="https://192.168.33.10:2380"
ETCD_ADVERTISE_CLIENT_URLS="https://192.168.33.10:2379"
ETCD_INITIAL_CLUSTER="etcd1=https://192.168.33.10:2380 "
ETCD_INITIAL_CLUSTER_TOKEN="etcd-cluster"
ETCD_INITIAL_CLUSTER_STATE="new"
```
### 复制证书文件
`sudo cp ~/certs/etcd/*.pem /etc/k8s/etcd/certs/`
### 配置systemd管理etcd
`sudo vi /usr/lib/systemd/system/etcd.service`
### 贴如下内容
```
[Unit]
Description=Etcd Server
After=network.target
After=network-online.target
Wants=network-online.target
[Service]
Type=notify
EnvironmentFile=/etc/k8s/etcd/config/etcd.conf 
ExecStart=/usr/k8s/etcd \
--cert-file=/etc/k8s/etcd/certs/server.pem \
--key-file=/etc/k8s/etcd/certs/server-key.pem \
--peer-cert-file=/etc/k8s/etcd/certs/server.pem \
--peer-key-file=/etc/k8s/etcd/certs/server-key.pem \
--trusted-ca-file=/etc/k8s/etcd/certs/ca.pem \
--peer-trusted-ca-file=/etc/k8s/etcd/certs/ca.pem \
--logger=zap
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
```
### 启动
```
sudo systemctl daemon-reload
sudo systemctl start etcd
sudo systemctl enable etcd
```
### 请求尝试
`sudo /usr/k8s/etcdctl --endpoints=https://192.168.33.10:2379 --cert=/etc/k8s/etcd/certs/server.pem --cacert=/etc/k8s/etcd/certs/ca.pem --key=/etc/k8s/etcd/certs/server-key.pem member list`
### 配置k8s证书
`cd ~/certs/k8s/`
### 新增证书文件
```
cat > ca-config.json << EOF
{
  "signing": {
    "defaulf": {
      "expiry": "87600h"
    },
    "profiles": {
      "kubernetes": {
        "expiry": "87600h",
        "usages": [
          "signing",
          "key encipherment",
          "server auth",
          "client auth"
        ]
      }
    }
  }
}
EOF
```
### 新增请求文件
```
cat > ca-csr.json << EOF
{
  "CN": "kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "Beijing",
      "ST": "Beijing",
      "O": "k8s",
      "OU": "System"
    }
  ]
}
EOF
```
### 生成ca证书
`cfssl gencert -initca ca-csr.json | cfssljson -bare ca -`

### 服务端http证书请求文件
```
cat > server-csr.json << EOF
{
    "CN": "kubernetes",
    "hosts": [
        "10.0.0.1",
        "127.0.0.1",
        "192.168.33.10",
        "192.168.33.11",
        "kubernetes",
        "kubernetes.default",
        "kubernetes.default.svc",
        "kubernetes.default.svc.cluster",
        "kubernetes.default.svc.cluster.local"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "CN",
            "L": "BeiJing",
            "ST": "BeiJing",
            "O": "k8s",
            "OU": "System"
        }
    ]
}
EOF
```
### 签发
`cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes server-csr.json | cfssljson -bare server`
### 复制k8s执行文件
`sudo cp /vagrant/master/kubernetes/server/bin/kube-apiserver /vagrant/master/kubernetes/server/bin/kubectl /usr/k8s`
### 创建k8s相关的配置文件目录、证书目录和日志目录
`sudo mkdir -p /etc/k8s/{configs,logs,certs}`
### 拷贝证书到证书目录里
`sudo cp ~/certs/k8s/*.pem /etc/k8s/certs`
### 新增token文件
`sudo vi /etc/k8s/configs/token.csv`

```
239309f7162e1fefdfa8ff63932fdbc4,"kubelet-bootstrap",10001,"system:node-bootstrapper"
```

### 新增运行文件

创建空文件: `sudo vi /etc/k8s/configs/kube-apiserver.conf`

`sudo vi /usr/lib/systemd/system/kube-apiserver.service`
### 内容如下
```
[Unit]
Description=Kubernetes API Server
Documentation=https://github.com/kubernetes/kubernetes
[Service]
EnvironmentFile=/etc/k8s/configs/kube-apiserver.conf
ExecStart=/usr/k8s/kube-apiserver \
--logtostderr=false \
--v=4 \
--log-dir=/etc/k8s/logs \
--etcd-servers=https://192.168.33.10:2379 \
--bind-address=192.168.33.10 \
--secure-port=6443 \
--advertise-address=192.168.33.10 \
--allow-privileged=true \
--service-cluster-ip-range=10.0.0.0/24 \
--enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota,NodeRestriction \
--authorization-mode=RBAC,Node \
--enable-bootstrap-token-auth=true \
--token-auth-file=/etc/k8s/configs/token.csv \
--service-node-port-range=30000-32767 \
--kubelet-client-certificate=/etc/k8s/certs/server.pem \
--kubelet-client-key=/etc/k8s/certs/server-key.pem \
--tls-cert-file=/etc/k8s/certs/server.pem  \
--tls-private-key-file=/etc/k8s/certs/server-key.pem \
--client-ca-file=/etc/k8s/certs/ca.pem \
--service-account-key-file=/etc/k8s/certs/ca-key.pem \
--service-account-signing-key-file=/etc/k8s/certs/ca-key.pem \
--service-account-issuer=https://kubernetes.default.svc \
--etcd-cafile=/etc/k8s/etcd/certs/ca.pem \
--etcd-certfile=/etc/k8s/etcd/certs/server.pem \
--etcd-keyfile=/etc/k8s/etcd/certs/server-key.pem \
--audit-log-maxage=30 \
--audit-log-maxbackup=3 \
--audit-log-maxsize=100 \
--audit-log-path=/etc/k8s/logs/k8s-audit.log 
Restart=on-failure
[Install]
WantedBy=multi-user.target
```

```
sudo systemctl daemon-reload
sudo systemctl start kube-apiserver
sudo systemctl enable kube-apiserver
```









## 创建kubectl证书

进入目录 `cd ~/certs/k8s/`

### 创建请求文件

```
cat > admin-csr.json << EOF
{
  "CN": "admin",
  "hosts": [],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "Beijing",
      "L": "Beijing",
      "O": "system:masters"
    }
  ]
}
EOF
```

因为是本地请求，所以`hosts`字段可以为空

`names.O` 字段主要是请求后会解析我们的证书，拿的就是证书里的这个字段，`kube-apiserver`收到请求会识别`group`为`system:masters`,内置的`clusterRoleBinding: clusrer-admin` 将`system:masters`与`Rolecluster-admin`绑定，该`Role`授予所有api的权限


### 签发证书

`cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes admin-csr.json | cfssljson -bare admin`

### 设置kubectl集群

`kubectl config set-cluster kubernetes --certificate-authority=/etc/k8s/certs/ca.pem --embed-certs=true --server=https://192.168.33.10:6443`

### 设置客户端认证参数

`kubectl config set-credentials kube-admin --client-certificate=/home/vagrant/certs/k8s/admin.pem --client-key=/home/vagrant/certs/k8s/admin-key.pem --embed-certs=true`

就是上面刚签发的证书

### 设置上下文

`kubectl config set-context kube-admin@kubernetes --cluster=kubernetes --user=kube-admin`

`kubectl config use-context kube-admin@kubernetes`

查看是否成功 `kubectl cluster-info`


## 启动controller-manager

`sudo cp /vagrant/master/kubernetes/server/bin/kube-controller-manager /usr/k8s`

`cd ~/certs/k8s/`


### 证书请求文件

```
cat > kube-controller-manager-csr.json << EOF
{
    "CN": "system:kube-controller-manager",
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "hosts": [
      "127.0.0.1",
      "192.168.33.10"
    ],
    "names": [
      {
        "C": "CN",
        "ST": "BeiJing",
        "L": "BeiJing",
        "O": "system:kube-controller-manager"
      }
    ]
}
EOF
```

### 生成证书

`cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-controller-manager-csr.json | cfssljson -bare kube-controller-manager`

```
sudo cp kube-controller-manager.pem /etc/k8s/certs
sudo cp kube-controller-manager-key.pem /etc/k8s/certs
```

### 创建配置文件

```
cat > /etc/k8s/configs/kube-controller-manager.kubeconfig << EOF
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /etc/k8s/certs/ca.pem
    server: https://192.168.33.10:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: system:kube-controller-manager
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: system:kube-controller-manager
  user:
    client-certificate: /etc/k8s/certs/kube-controller-manager.pem
    client-key: /etc/k8s/certs/kube-controller-manager-key.pem
EOF
```

### 服务启动文件

`sudo vi /usr/lib/systemd/system/kube-controller-manager.service`

如下

```
[Unit]
Description=Kubernetes Controller Manager
Documentation=https://github.com/kubernetes/kubernetes
[Service]
ExecStart=/usr/k8s/kube-controller-manager \
--logtostderr=false \
--kubeconfig=/etc/k8s/configs/kube-controller-manager.kubeconfig \
--v=4 \
--log-dir=/etc/k8s/logs \
--leader-elect=false \
--master=https://192.168.33.10:6443 \
--bind-address=127.0.0.1 \
--allocate-node-cidrs=true \
--cluster-cidr=10.244.0.0/16 \
--service-cluster-ip-range=10.0.0.0/24 \
--cluster-signing-cert-file=/etc/k8s/certs/ca.pem \
--cluster-signing-key-file=/etc/k8s/certs/ca-key.pem  \
--root-ca-file=/etc/k8s/certs/ca.pem \
--service-account-private-key-file=/etc/k8s/certs/ca-key.pem \
--client-ca-file=/etc/k8s/certs/ca.pem \
--tls-cert-file=/etc/k8s/certs/kube-controller-manager.pem \
--tls-private-key-file=/etc/k8s/certs/kube-controller-manager-key.pem \
--cluster-signing-duration=87600h0m0s \
--use-service-account-credentials=true
Restart=on-failure
[Install]
WantedBy=multi-user.target
```


```
sudo systemctl daemon-reload
sudo systemctl enable kube-controller-manager
sudo systemctl start kube-controller-manager
```

## 部署kube-scheduler

`sudo cp /vagrant/master/kubernetes/server/bin/kube-scheduler  /usr/k8s`

`cd ~/certs/k8s/`

### 证书请求文件

```
cat > kube-scheduler-csr.json << EOF
{
    "CN": "system:kube-scheduler",
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "hosts": [
      "127.0.0.1",
      "192.168.33.10"
    ],
    "names": [
      {
        "C": "CN",
        "ST": "BeiJing",
        "L": "BeiJing",
        "O": "system:kube-scheduler"
      }
    ]
}
EOF
```


生成证书

`cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-scheduler-csr.json  | cfssljson -bare kube-scheduler`

拷贝文件

`sudo cp kube-scheduler.pem /etc/k8s/certs && sudo cp kube-scheduler-key.pem /etc/k8s/certs`

### 新增配置文件

`sudo vi /etc/k8s/configs/kube-scheduler.kubeconfig`

内容如下:

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /etc/k8s/certs/ca.pem
    server: https://192.168.33.10:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: system:kube-scheduler
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: system:kube-scheduler
  user:
    client-certificate: /etc/k8s/certs/kube-scheduler.pem
    client-key: /etc/k8s/certs/kube-scheduler-key.pem
```

### 新增服务文件

`sudo vi /usr/lib/systemd/system/kube-scheduler.service`

内容如下：

```
[Unit]
Description=Kubernetes Scheduler
Documentation=https://github.com/kubernetes/kubernetes
[Service]
ExecStart=/usr/k8s/kube-scheduler \
--logtostderr=true \
--v=4 \
--kubeconfig=/etc/k8s/configs/kube-scheduler.kubeconfig \
--log-dir=/etc/k8s/logs \
--master=https://192.168.33.10:6443 \
--bind-address=127.0.0.1
--tls-cert-file=/etc/k8s/certs/kube-scheduler.pem \
--tls-private-key-file=/etc/k8s/certs/kube-scheduler-key.pem \
--client-ca-file=/etc/k8s/certs/ca.pem
Restart=on-failure
[Install]
WantedBy=multi-user.target
```

启动：

```
sudo systemctl daemon-reload
sudo systemctl restart kube-scheduler
sudo systemctl enable kube-scheduler
```

查看是否正常

```
kubectl get cs
```

## 部署kubelet

使用 `K8S1.4` 之后的 `TLS bootstraping` 方式动态签署发布证书


`sudo cp /vagrant/master/kubernetes/server/bin/kubelet /usr/k8s`


### 修改docker

`sudo vim /etc/docker/daemon.json`

内容如下

```
{
  "exec-opts": ["native.cgroupdriver=systemd"]
}
```

`sudo systemctl daemon-reload && sudo systemctl restart docker.service`

确保成功 `docker info | grep Cgroup`

### 配置文件

`sudo vi /etc/k8s/configs/kubelet-config.yaml`

内容

```
kind: KubeletConfiguration
apiVersion: kubelet.config.k8s.io/v1beta1
address: 0.0.0.0
port: 10250
readOnlyPort: 10255
cgroupDriver: systemd
clusterDNS:
- 10.0.0.2
clusterDomain: cluster.local 
failSwapOn: true
authentication:
  anonymous:
    enabled: false
  webhook:
    cacheTTL: 2m0s
    enabled: true
  x509:
    clientCAFile: /etc/k8s/certs/ca.pem 
authorization:
  mode: Webhook
  webhook:
    cacheAuthorizedTTL: 5m0s
    cacheUnauthorizedTTL: 30s
evictionHard:
  imagefs.available: 15%
  memory.available: 100Mi
  nodefs.available: 10%
  nodefs.inodesFree: 5%
maxOpenFiles: 1000000
maxPods: 110
```

创建目录：`sudo mkdir /etc/k8s/configs/kubelet`

帮`token.csv`内定义的角色绑定权限:

`kubectl create clusterrolebinding kubelet-bootstrap --clusterrole=system:node-bootstrapper --user=kubelet-bootstrap`

### 创建启动配置文件

新增启动文件: `sudo vim /etc/k8s/configs/bootstrap.kubeconfig`

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /etc/k8s/certs/ca.pem
    server: https://192.168.33.10:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubelet-bootstrap
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: kubelet-bootstrap
  user:
    token: 239309f7162e1fefdfa8ff63932fdbc4
```

里面 `user` 为 `token.csv` 里定义的用户名称，`token` 也是 `token.csv` 里定义的 `token`

### 配置服务文件

`sudo vi /usr/lib/systemd/system/kubelet.service`

内容如下:

```
[Unit]
Description=Kubernetes Kubelet
After=docker.service
[Service]
ExecStart=/usr/k8s/kubelet \
--logtostderr=false \
--v=4 \
--log-dir=/etc/k8s/logs \
--hostname-override=node01 \
--network-plugin=cni \
--cni-bin-dir=/usr/k8s/cni \
--kubeconfig=/etc/k8s/configs/kubelet.kubeconfig \
--bootstrap-kubeconfig=/etc/k8s/configs/bootstrap.kubeconfig \
--config=/etc/k8s/configs/kubelet-config.yaml \
--cert-dir=/etc/k8s/certs/kubelet \
--pod-infra-container-image=mirrorgooglecontainers/pause-amd64:3.0
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
```

启动：

`sudo systemctl daemon-reload && sudo systemctl start kubelet && sudo systemctl enable kubelet`

## kubelet授权

`kubectl get csr`

允许加入：`kubectl certificate approve node-csr-9o-XimTiREqG_FslHEbl1jI4DbAxqv3f1xCYK2F3Dfs`


获取加入是否成功 `kubectl get nodes`


## 部署 kube-proxy

`sudo cp /vagrant/master/kubernetes/server/bin/kube-proxy /usr/k8s/`

### 创建证书请求文件

`cd ~/certs/k8s/`

```
{
  "CN": "system:kube-proxy",
  "hosts": [],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "BeiJing",
      "ST": "BeiJing",
      "O": "kube-proxy",
      "OU": "System"
    }
  ]
}
```

生成： `cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes kube-proxy-csr.json | cfssljson -bare kube-proxy`


拷贝至指定目录: `sudo cp kube-proxy.pem /etc/k8s/certs && sudo cp kube-proxy-key.pem /etc/k8s/certs`

### 创建config配置文件

这部分最好在`root`环境下执行 不然容易读取`crt`证书失败

```
kubectl config set-cluster kubernetes \
  --certificate-authority=/etc/k8s/certs/ca.pem \
  --server=https://192.168.33.10:6443 \
  --embed-certs=true \
  --kubeconfig=kube-proxy.kubeconfig
  
kubectl config set-credentials kube-proxy \
  --client-certificate=/etc/k8s/certs/kube-proxy.pem \
  --client-key=/etc/k8s/certs/kube-proxy-key.pem \
  --embed-certs=true \
  --kubeconfig=kube-proxy.kubeconfig
 
kubectl config set-context default \
  --cluster=kubernetes \
  --user=kube-proxy \
  --kubeconfig=kube-proxy.kubeconfig

kubectl config use-context default --kubeconfig=kube-proxy.kubeconfig
```

在当前文件夹生成了一个 `kube-proxy.kubeconfig` 文件

复制当目录下 `sudo cp kube-proxy.kubeconfig /etc/k8s/configs`


创建 `sudo vi /etc/k8s/configs/kube-proxy-config.yml`

```
kind: KubeProxyConfiguration
apiVersion: kubeproxy.config.k8s.io/v1alpha1
bindAddress: 0.0.0.0
metricsBindAddress: 0.0.0.0:10249
clientConnection:
  kubeconfig: /etc/k8s/configs/kube-proxy.kubeconfig
hostnameOverride: node01
clusterCIDR: 10.0.0.0/24
```

### 创建服务文件

`sudo vi /usr/lib/systemd/system/kube-proxy.service`

```
[Unit]
Description=Kubernetes Proxy
After=network.target
[Service]
ExecStart=/usr/k8s/kube-proxy \
--logtostderr=false \
--v=4 \
--log-dir=/etc/k8s/logs \
--config=/etc/k8s/configs/kube-proxy-config.yml
Restart=on-failure
LimitNOFILE=65536
[Install]
WantedBy=multi-user.target
```

`sudo systemctl daemon-reload && sudo systemctl start kube-proxy && sudo systemctl enable kube-proxy`

# node:










