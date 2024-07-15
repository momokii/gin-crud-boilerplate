package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/momokii/gin-crud-boilerplate/db"
	"github.com/momokii/gin-crud-boilerplate/middlewares"
	"github.com/momokii/gin-crud-boilerplate/routes"
)

func main() {

	db.InitPostgres() // init db (postgres)

	is_production := os.Getenv("PRODUCTION")
	if is_production == "true" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middlewares.CORS())

	// api routing
	apiV1 := r.Group("/api/v1")

	routes.SetupAuthRoutes(apiV1.Group("/auth"))
	routes.SetupUserRoutes(apiV1.Group("/users"))

	// start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	err := r.Run(":" + port)
	if err != nil {
		panic(err)
	}

}
