package main

import (
	"log"

	"bundler/config"
	"bundler/controllers"
	"bundler/routes"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv() // 加载环境变量

	r := gin.Default()

	// 连接以太坊客户端
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// 创建 UserOpController 实例
	userOpController, err := controllers.NewUserOpController()
	if err != nil {
		log.Fatalf("Failed to create UserOpController: %v", err)
	}

	// 创建 PublicKeyOracleController 实例
	publicKeyOracleController := controllers.NewPublicKeyOracleController(client)

	// 创建 DepositController 实例
	depositController := controllers.NewDepositController(userOpController.Client)

	// 初始化路由
	routes.SetupRouter(r)
	routes.SetupUserOpRouter(r, userOpController)
	routes.SetupDepositRouter(r, depositController)
	routes.SetupPublicKeyOracleRouter(r, publicKeyOracleController)

	// 运行服务器
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
