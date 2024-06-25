package main

import (
	"log"

	"bundler/config"
	"bundler/controllers"
	"bundler/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv() // 加载环境变量

	r := gin.Default()

	// 创建 UserOpController 实例
	userOpController, err := controllers.NewUserOpController()
	if err != nil {
		log.Fatalf("Failed to create UserOpController: %v", err)
	}

	// 创建 DepositController 实例
	depositController := controllers.NewDepositController(userOpController.Client)

	// 初始化路由
	routes.SetupRouter(r)
	routes.SetupUserOpRouter(r, userOpController)
	routes.SetupDepositRouter(r, depositController) // 新增的路由设置

	// 运行服务器
	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
