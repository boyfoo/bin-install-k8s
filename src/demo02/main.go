package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"src/demo02/clients"
	"src/demo02/controller"
)

func main() {
	client := clients.NewCertClient()
	r := gin.Default()
	r.GET("version", func(ctx *gin.Context) {
		version, _ := client.ServerVersion()
		ctx.String(http.StatusOK, version.String())
	})
	controller.DeployRoute(client, r)
	r.Run(":8000")
}
