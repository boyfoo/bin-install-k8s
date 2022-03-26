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
  post: 101122
```

发布`crd`到`k8s`中: `make install`

查看按照结果: `kb get crd`

若要重新发布最新的`crd`: `make uninstall && make install`

### 修改逻辑

修改业务文件`controllers\redisboyfoo_controller.go`

```go
package controllers

func (r *RedisBoyfooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	redisBoyfoo := &myappv1.RedisBoyfoo{}
	// 获取yaml提交来创建的信息 
	if err := r.Get(ctx, req.NamespacedName, redisBoyfoo); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("得到对象")
		fmt.Println(redisBoyfoo)
	}

	return ctrl.Result{}, nil
}
```

本地运行控制器调试 `make run`，另起一个终端发布资源`kb apply -f test/redis.yaml`，可以在前一个终端看见打印的数据内容

### spec 验证

文档地址 `https://book.kubebuilder.io/reference/markers/crd-validation.html`

```go
package v1

type RedisBoyfooSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of RedisBoyfoo. Edit redisboyfoo_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	// 校验文档地址 https://book.kubebuilder.io/reference/markers/crd-validation.html
	// +kubebuilder:validation:Minimum:=80
	// +kubebuilder:validation:Maximum:=1000
	Port int `json:"port,omitempty"`
}
```

重新安装`make uninstall && make install`

启动`make run`

新增资源: `kubectl apply -f test/redis.yaml`，会提示验证错误 `port` 值太大

可以查看控制器有哪些规则 `kb describe crd redisboyfooes.myapp.boyfoo.com`

### 为资源新增创建pod功能

修改文件`redisboyfoo_controller.go`：

```go
package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	myappv1 "k8sapi2/api/v1"
)

func CreateRedis(client client.Client, redisBoyfooConfig *myappv1.RedisBoyfoo) error {
	newpod := &corev1.Pod{}
	newpod.Namespace = redisBoyfooConfig.Namespace
	newpod.Name = redisBoyfooConfig.Name
	newpod.Spec.Containers = []corev1.Container{
		{
			Name:            redisBoyfooConfig.Name,
			Image:           "redis:5-alpine",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: int32(redisBoyfooConfig.Spec.Port),
				},
			},
		},
	}
	return client.Create(context.Background(), newpod)
}
```

### 删除pod

使用 `Finalizers` 原理删除
