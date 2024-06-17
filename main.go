package main

import (
	"log"

	"bundler/config"
	"bundler/controllers"

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

	// 设置路由和处理方法
	r.POST("/userOp", userOpController.StoreUserOp)

	// 运行服务器
	if err := r.Run(":3000"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
