package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/momokii/gin-crud-boilerplate/controllers"
	"github.com/momokii/gin-crud-boilerplate/middlewares"
)

func SetupUserRoutes(router *gin.RouterGroup) {
	router.GET("/self", middlewares.IsAuth, controllers.GetSelf)
	router.GET("/", middlewares.IsAuth, middlewares.IsAdmin, controllers.GetAllUsers)
	router.GET("/:id", middlewares.IsAuth, middlewares.IsAdmin, controllers.GetOneUser)
	router.POST("/", middlewares.IsAuth, middlewares.IsAdmin, controllers.CreateUser)
	router.PATCH("/", middlewares.IsAuth, controllers.EditUser)
	router.PATCH("/password", middlewares.IsAuth, controllers.EditUserPassword)
	router.PATCH("/:id/status", middlewares.IsAuth, middlewares.IsAdmin, controllers.EditUserStatus)
	router.DELETE("/:id", middlewares.IsAuth, middlewares.IsAdmin, controllers.DeleteUser)
}
