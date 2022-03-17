package clients

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

// NewTokenClient 通过token创建客户端
func NewTokenClient() *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:        "https://192.168.10.10:8443/k8s/clusters/c-xl9h8",
		BearerToken: "kubeconfig-user-lmfrd:ns2k2xqtq2np7s2tb9jfbxgwwtjjdm8j27gh7gdqtjnhlvqn2rflv2",
		TLSClientConfig: rest.TLSClientConfig{ // 不验证响应tls
			Insecure: true,
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	return client
}
