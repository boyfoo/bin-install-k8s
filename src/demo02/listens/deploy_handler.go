package listens

import (
	"fmt"
	v1 "k8s.io/api/apps/v1"
)

type DepHandler struct {
}

func (d DepHandler) OnAdd(obj interface{}) {
	if deployment, ok := obj.(*v1.Deployment); ok {
		fmt.Println("add")
		fmt.Println(deployment.Name)
	}
}

func (d DepHandler) OnUpdate(oldObj, newObj interface{}) {
	if deployment, ok := newObj.(*v1.Deployment); ok {
		fmt.Println("update")
		fmt.Println(deployment.Name)
	}
}

func (d DepHandler) OnDelete(obj interface{}) {
	//TODO implement me
	fmt.Println("OnDelete")
}
