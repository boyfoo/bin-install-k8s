package listens

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

//// NewDeploymentInformer 创建一个监听单独命名空间deployment的坚挺者
//func NewDeploymentInformer(clientset *kubernetes.Clientset) {
//	client := cache.NewListWatchFromClient(clientset.AppsV1().RESTClient(), "deployments", common.Namespace, fields.Everything())
//	_, controller := cache.NewInformer(client, &v1.Deployment{}, 0, &DepHandler{})
//	controller.Run(wait.NeverStop)
//}

func NewSharedInformer(clientset *kubernetes.Clientset) {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// 创建dep监听
	deployments := factory.Apps().V1().Deployments()
	deployments.Informer().AddEventHandler(&DepHandler{})
	//depInformer.Run(wait.NeverStop)

	factory.Start(wait.NeverStop)
}
