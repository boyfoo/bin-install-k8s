package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"src/demo02/clients"
	"src/demo02/controller"
	"src/demo02/listens"
)

func main() {
	client := clients.NewCertClient()

	// 监听
	go listens.NewSharedInformer(client)

	// http
	r := gin.Default()
	r.GET("version", func(ctx *gin.Context) {
		version, _ := client.ServerVersion()
		ctx.String(http.StatusOK, version.String())
	})
	// Deployments 相关api操作
	controller.DeployRoute(client, r)
	r.Run(":8000")
}
