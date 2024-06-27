package routes

import (
	"bundler/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupPublicKeyOracleRouter 设置公钥预言机路由
func SetupPublicKeyOracleRouter(r *gin.Engine, publicKeyOracleController *controllers.PublicKeyOracleController) {
	r.POST("/publicKeyOracle/setPublicKey", func(c *gin.Context) {
		var request struct {
			Domain   string `json:"domain"`
			Selector string `json:"selector"`
			Modulus  []byte `json:"modulus"`
			Exponent []byte `json:"exponent"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		txHash, err := publicKeyOracleController.SetPublicKey(request.Domain, request.Selector, request.Modulus, request.Exponent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transaction sent successfully", "transactionHash": txHash})
	})

	r.GET("/publicKeyOracle/getRSAKey", func(c *gin.Context) {
		var request struct {
			Domain   string `form:"domain"`
			Selector string `form:"selector"`
		}
		if err := c.ShouldBindQuery(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 调用获取RSA密钥的方法
		txHash, modulus, exponent, err := publicKeyOracleController.GetRSAKey(request.Domain, request.Selector)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transaction sent successfully", "transactionHash": txHash, "modulus": modulus, "exponent": exponent})
	})
}
