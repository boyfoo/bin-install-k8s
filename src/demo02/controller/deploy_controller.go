package controller

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

type DeployController struct {
	client *kubernetes.Clientset
}

// List 获取全部deploy
func (d *DeployController) List() (*appsv1.DeploymentList, error) {
	list, err := d.client.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CreateForYaml 通过文件创建
func (d *DeployController) CreateForYaml(path string) (*appsv1.Deployment, error) {
	dep, err := d.getJsonByYamlFile(path)
	if err != nil {
		return nil, err
	}
	create, err := d.client.AppsV1().Deployments("default").Create(context.TODO(), dep, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return create, nil
}

// DeleteForYaml 通过文件删除
func (d *DeployController) DeleteForYaml(path string) error {
	dep, err := d.getJsonByYamlFile(path)
	if err != nil {
		return err
	}
	return d.client.AppsV1().Deployments("default").Delete(context.TODO(), dep.Name, metav1.DeleteOptions{})
}

// 根据文件路径获取json
func (d *DeployController) getJsonByYamlFile(path string) (*appsv1.Deployment, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// 把ymal格式转换成json格式
	ngxJson, err := yaml.ToJSON(b)
	if err != nil {
		return nil, err
	}
	dep := &appsv1.Deployment{}
	err = json.Unmarshal(ngxJson, dep)
	if err != nil {
		return nil, err
	}
	return dep, nil
}

func DeployRoute(client *kubernetes.Clientset, r *gin.Engine) {
	deployController := NewDeployController(client)
	r.GET("deploy", func(ctx *gin.Context) {
		list, err := deployController.List()
		if err != nil {
			ctx.JSON(200, err)
			return
		}
		res := []map[string]interface{}{}
		for _, item := range list.Items {

			images := ""
			for _, c := range item.Spec.Template.Spec.Containers {
				images += c.Image
			}

			res = append(res, map[string]interface{}{
				"名称": item.Name,
				"副本数": map[string]interface{}{
					"总副本数":   item.Status.Replicas,
					"可用副本数":  item.Status.AvailableReplicas,
					"不可用副本数": item.Status.UnavailableReplicas,
				},
				"镜像名称": images,
			})
		}
		ctx.JSON(200, res)
	})
	r.POST("deploy", func(ctx *gin.Context) {
		path := &struct {
			Path string `json:"path"`
		}{}
		err := ctx.ShouldBindJSON(path)
		yamlBytes, err := deployController.CreateForYaml(path.Path)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}
		ctx.JSON(http.StatusOK, yamlBytes)
	})
	r.DELETE("deploy", func(ctx *gin.Context) {
		path := &struct {
			Path string `json:"path"`
		}{}
		err := ctx.ShouldBindJSON(path)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
		}
		err = deployController.DeleteForYaml(path.Path)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}
		ctx.JSON(http.StatusOK, nil)
	})
}

func NewDeployController(client *kubernetes.Clientset) *DeployController {
	return &DeployController{client: client}
}
