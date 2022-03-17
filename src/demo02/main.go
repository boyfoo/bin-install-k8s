package main

import (
	"fmt"
	"src/demo02/clients"
)

func main() {
	client := clients.NewCertClient()
	// 查看版本
	version, _ := client.ServerVersion()
	fmt.Println(version)
}
