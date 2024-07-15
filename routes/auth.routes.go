package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/momokii/gin-crud-boilerplate/controllers"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", controllers.Login)
}
