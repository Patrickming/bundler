package routes

import (
	"bundler/controllers"

	"github.com/gin-gonic/gin"
)

func SetupUserOpRouter(r *gin.Engine, userOpController *controllers.UserOpController) {
	r.POST("/userOp", userOpController.StoreUserOp)
}
