package controller

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"src/demo02/common"
)

type DeployController struct {
	client *kubernetes.Clientset
}

// List 获取全部deploy
func (d *DeployController) List() (*appsv1.DeploymentList, error) {
	list, err := d.client.AppsV1().Deployments(common.Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (d *DeployController) Show(name string) (*appsv1.Deployment, error) {
	return d.client.AppsV1().Deployments(common.Namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (d *DeployController) getPodsByDeployment(dep *appsv1.Deployment) (*corev1.PodList, error) {
	// 标签选择 获取选择
	selector, err := metav1.LabelSelectorAsSelector(dep.Spec.Selector)
	if err != nil {
		return nil, err
	}

	// deployment 调度 ReplicaSets ， ReplicaSets 调度 pod
	// 不能使用deployment直接获取pod，可能因为标签一样获取到其他的pod
	// 可能获取到多个ReplicaSets 应该deployment控制的最新的ReplicaSets
	// 通过命令行获取最新，例: kubectl describe deployment nginx-deployment-name01 | grep NewReplicaSet
	ReplicaList, err := d.client.AppsV1().ReplicaSets(dep.Namespace).List(
		context.Background(), metav1.ListOptions{LabelSelector: selector.String()},
	)
	if err != nil {
		return nil, err
	}

	replicaListSelector := ""
	for _, item := range ReplicaList.Items {
		// 是否是dep当前控制的set
		if IsCurrentRsByDep(dep, &item) {
			retSelector, err := metav1.LabelSelectorAsSelector(item.Spec.Selector)
			if err != nil {
				break
			}
			replicaListSelector = retSelector.String()
			break
		}
	}

	podList, err := d.client.CoreV1().Pods(dep.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: replicaListSelector,
	})

	if err != nil {
		return nil, err
	}

	return podList, nil
}

// CreateForYaml 通过文件创建
func (d *DeployController) CreateForYaml(path string) (*appsv1.Deployment, error) {
	dep, err := d.getJsonByYamlFile(path)
	if err != nil {
		return nil, err
	}
	create, err := d.client.AppsV1().Deployments(common.Namespace).Create(context.TODO(), dep, metav1.CreateOptions{})
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
	return d.client.AppsV1().Deployments(common.Namespace).Delete(context.TODO(), dep.Name, metav1.DeleteOptions{})
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

			res = append(res, map[string]interface{}{
				"名称": item.Name,
				"副本数": map[string]interface{}{
					"总副本数":   item.Status.Replicas,
					"可用副本数":  item.Status.AvailableReplicas,
					"不可用副本数": item.Status.UnavailableReplicas,
				},
				"镜像名称": getImages(&item),
			})
		}
		ctx.JSON(200, res)
	})
	r.GET("deploy/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		show, err := deployController.Show(name)
		if err != nil {
			ctx.JSON(200, err)
			return
		}

		// 获取pod
		podList, err := deployController.getPodsByDeployment(show)
		if err != nil {
			ctx.JSON(200, err)
			return
		}

		var podsMap []map[string]interface{}
		for _, pod := range podList.Items {
			podsMap = append(podsMap, map[string]interface{}{
				"1.名称":   pod.Name,
				"2.镜像":   getImagesByPod(&pod),
				"3.所属节点": pod.Spec.NodeName,
				"4.创建时间": pod.CreationTimestamp.Format(common.GoTime),
				"5.阶段":   pod.Status.Phase,
				"6.信息":   GetMessage(&pod),
			})
		}

		ctx.JSON(200, map[string]interface{}{
			"1.名称":   show.Name,
			"2.命名空间": show.Namespace,
			"3.镜像":   getImages(show),
			"4.创建时间": show.CreationTimestamp.Format(common.GoTime),
			"5.pods": podsMap,
		})
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

	// 更新副本数
	r.POST("deploy-scale", func(ctx *gin.Context) {
		req := &struct {
			DeploymentName string `json:"deployment_name"`
			Dec            bool   `json:"dec"`
		}{}
		err := ctx.ShouldBindJSON(req)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}
		// 获取原本的
		scale, err := deployController.client.AppsV1().Deployments(common.Namespace).GetScale(
			context.Background(),
			req.DeploymentName,
			metav1.GetOptions{},
		)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}

		if req.Dec {
			scale.Spec.Replicas--
		} else {
			scale.Spec.Replicas++
		}
		updateScale, err := deployController.client.AppsV1().Deployments(common.Namespace).UpdateScale(
			context.Background(),
			req.DeploymentName,
			scale,
			metav1.UpdateOptions{},
		)
		if err != nil {
			ctx.JSON(http.StatusOK, err)
			return
		}

		ctx.JSON(http.StatusOK, updateScale)
	})
}

// IsCurrentRsByDep set是否是dep的当前set
func IsCurrentRsByDep(dep *appsv1.Deployment, set *appsv1.ReplicaSet) bool {
	// 版本不相等直接不是
	if dep.ObjectMeta.Annotations["deployment.kubernetes.io/revision"] == set.ObjectMeta.Annotations["deployment.kubernetes.io/revision"] {
		return false
	}

	// set引用是否是对应dep
	for _, reference := range set.OwnerReferences {
		if reference.Kind == "Deployment" && reference.Name == dep.Name {
			return true
		}
	}

	return false
}

func GetMessage(pod *corev1.Pod) string {
	msg := ""
	for _, condition := range pod.Status.Conditions {
		if condition.Status != "True" {
			msg += condition.Message
		}
	}
	return msg
}

// 获取镜像
func getImages(item *appsv1.Deployment) string {
	images := ""
	for _, c := range item.Spec.Template.Spec.Containers {
		images += c.Image
	}
	return images
}

// 获取pod的镜像
func getImagesByPod(item *corev1.Pod) string {
	images := ""
	for _, c := range item.Spec.Containers {
		images += c.Image
	}
	return images
}

func NewDeployController(client *kubernetes.Clientset) *DeployController {
	return &DeployController{client: client}
}
