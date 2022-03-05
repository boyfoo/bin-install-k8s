package main

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"log"
	"src/demo01/lib"

	"io/ioutil"
	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// 准入控制器 https://kubernetes.io/zh/docs/reference/access-authn-authz/extensible-admission-controllers/
// 完成内容 部署的pod只能是指定的名字
func main() {
	http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		// 读取请求的内容
		var body []byte
		if r.Body != nil {
			if data, err := ioutil.ReadAll(r.Body); err == nil {
				body = data
			}
		}
		//第二步
		reqAdmissionReview := v1.AdmissionReview{} //请求过来的内容
		rspAdmissionReview := v1.AdmissionReview{  //准备响应的内容
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
		}
		//第三步  把body decode 成请求对象
		deserializer := lib.Codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(body, nil, &reqAdmissionReview); err != nil {
			// decode响应失败
			klog.Error(err)
			// 构建一个带error内容的响应
			rspAdmissionReview.Response = lib.ToV1AdmissionResponse(err)
		} else {
			// decode成功 执行我们业务内容
			rspAdmissionReview.Response = lib.AdmitPods(reqAdmissionReview) //我们的业务
		}
		// 不管是否通过都要设置id
		rspAdmissionReview.Response.UID = reqAdmissionReview.Request.UID
		respBytes, err := json.Marshal(rspAdmissionReview)
		if err != nil {
			klog.Error(err)
		} else {
			if _, err := w.Write(respBytes); err != nil {
				klog.Error(err)
			}
		}
	})

	fmt.Println("启动")
	// 不使用tls 简单测试启动 这个启动方法才能用client.go测试 如果使用后面的tls启动client.go无效
	//http.ListenAndServe(":8080", nil)

	tlsConfig := lib.Config{
		CertFile: "/etc/webhook/certs/tls.crt", // 为什么是这个路径，没有为什么，个人喜好随意来
		KeyFile:  "/etc/webhook/certs/tls.key",
	}
	server := &http.Server{
		Addr:      ":443", // TLS默认443
		TLSConfig: lib.ConfigTLS(tlsConfig),
	}
	server.ListenAndServeTLS("", "") // 没有传参数，因为上面已经配置了tls路径
}
