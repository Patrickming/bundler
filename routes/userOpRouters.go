// userOpRouters.go

package routes

import (
	"bundler/controllers"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func SetupUserOpRouter(r *gin.Engine, userOpController *controllers.UserOpController) {
	r.POST("/userOp", userOpController.StoreUserOp)
}

// SetupDepositRouter 初始化存款路由
func SetupDepositRouter(r *gin.Engine, depositController *controllers.DepositController) {
	r.POST("/deposit", func(c *gin.Context) {
		var request struct {
			Address string `json:"address"`
			Amount  string `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		address := common.HexToAddress(request.Address)
		amount, ok := new(big.Int).SetString(request.Amount, 10)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount format"})
			return
		}

		// 创建 DepositController 实例并调用 DepositToAddress 方法
		txHash, err := depositController.DepositToAddress(address, amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deposit to address completed", "transactionHash": txHash})
	})
}
