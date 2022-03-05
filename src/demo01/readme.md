### 测试

`go run client.go` 只有 `main.go` 内代码不使用`tls`启动才有效

### 编译

`GOOS=linux GOARCH=amd64 go build -o myhook main.go`

### 生成证书

进入虚拟机`node01`

创建目录 `mkdir -p ~/hook/certs && cd ~/hook/certs`

1. `vi ca-config.json` 内容如下

```
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "server": {
        "usages": ["signing"],
        "expiry": "8760h"
      }
    }
  }
}
```

2. `vi ca-csr.json` 内如如下

```
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "zh",
      "L": "bj",
      "O": "bj",
      "OU": "CA"
   }
  ]
}
```

3. 生成ca证书
   `cfssl gencert -initca ca-csr.json | cfssljson -bare ca`


4. 生成服务端证书 `vi server-csr.json`

```
{
  "CN": "admission",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "zh",
      "L": "bj",
      "O": "bj",
      "OU": "bj"
    }
  ]
}
```

5. 签发证书

```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname=myhook.kube-system.svc \
  -profile=server \
  server-csr.json | cfssljson -bare server
```

`-hostname` 很重要，是服务部署到`k8s`里的服务名称、命名空间

### 部署到`k8s`

下面使用`node01`来执行的原因主要是因为`yaml`文件内的挂载目录写的是`node01`上的目录

1. 复制证书内容`cat ca.pem | base64`，到`admconfig.yaml`文件内

2. 在`node01`上执行`kubectl create secret tls myhook --cert=server.pem --key=server-key.pem -n kube-system`
   ，将证书文件导入配置，最后挂载到`pod`内

3. 修改`yamls`对应的需要修改的内容(一般不用)

4. 在`node01`上执行`kubectl apply -f deploy.yaml`

5. 在`node01`上执行`kubectl apply -f admconfig.yaml`

6. 测试是否被拒绝`kubectl apply -f newpod.yaml`