package lib

import (
	"fmt"
	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func AdmitPods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	fmt.Println("request")
	fmt.Println(ar.Request)
	// 判断是否是一个pod的数据
	if ar.Request.Resource != podResource {
		err := fmt.Errorf("expect resource to be %s", podResource)
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}
	// 正常业务
	// 请求的内容格式
	// https://kubernetes.io/zh/docs/reference/access-authn-authz/extensible-admission-controllers/#request
	// Raw内的是用yaml文件创建内容的json格式
	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	fmt.Println("pod")
	fmt.Println(pod)
	reviewResponse := v1.AdmissionResponse{}
	// 判断pod名称是否正常 错误返回false不允许部署
	if pod.Name == "shenyi" {
		reviewResponse.Allowed = false
		reviewResponse.Result = &metav1.Status{Code: 503, Message: "pod name cannot be shenyi"}
	} else {
		reviewResponse.Allowed = true
	}

	return &reviewResponse
}
