### 安装kubebuilder

1. 下载`https://github.com/kubernetes-sigs/kubebuilder`
2. 运行权限 `chmod +x kubebuilder3.1.0`
3. 复制到可运行的环境变量目录

### 安装 kustomize

1. 下载 `https://github.com/kubernetes-sigs/kustomize`
2. 运行权限 `chmod +x kustomize3.9.3`
3. 复制到可运行的环境变量目录

### get starting

必须在只有`go.mod`的空目录

```
# 初始化项目
$ kubebuilder init --domain boyfoo.com

# 创建一个类型
$ kubebuilder create api --group myapp --version v1 --kind RedisBoyfoo
```

以上两条内容等于创建了一个自定义的资源为:

```yaml
apiVersion: myapp.boyfoo.com/v1
kind: RedisBoyfoo
```

修改结构体`redisboyfoo_types.go`内的结构体：

```go
package v1

type RedisBoyfooSpec struct {
	Port int `json:"port,omitempty"`
}
```

意思为`spec`内有个`port`参数

于是部署`yaml`文件目前为:

```yaml
apiVersion: myapp.boyfoo.com/v1
kind: RedisBoyfoo
metadata:
  name: myredis
# 上面的都是基本格式，下面的是自定义格式(但是也要按照规范自定义)
spec: # 对应RedisBoyfooSpec结构体
  post: 1011
```

按照`crd`到`k8s`中: `make install` 

查看按照结果: `kb get crd`