package routes

import (
	"bundler/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupPublicKeyOracleRouter 初始化公钥路由
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

		c.JSON(http.StatusOK, gin.H{"txHash": txHash})
	})

	r.GET("/publicKeyOracle/getRSAKey", func(c *gin.Context) {
		domain := c.Query("domain")
		selector := c.Query("selector")

		modulus, exponent, err := publicKeyOracleController.GetRSAKey(domain, selector)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{
			"modulus":  modulus,
			"exponent": exponent,
		}

		c.JSON(http.StatusOK, response)
	})
}
