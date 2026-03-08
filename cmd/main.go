package main

import (
	"log"
	"os"

	"product-mgmt/internal/database"
	"product-mgmt/internal/handlers"
	"product-mgmt/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	database.Connect()
	database.Migrate()
	database.Seed()

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" { c.AbortWithStatus(204); return }
		c.Next()
	})

	r.Static("/static", "./frontend/static")
	r.LoadHTMLGlob("frontend/templates/*")
	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", nil) })

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", handlers.Login)
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			auth.POST("/auth/logout", handlers.Logout)
			auth.GET("/auth/me", handlers.GetMe)
			auth.PUT("/auth/profile", handlers.UpdateProfile)

			auth.GET("/dashboard", handlers.GetDashboardStats)

			auth.GET("/products", middleware.RequirePermission("products", "read"), handlers.GetProducts)
			auth.GET("/products/categories", middleware.RequirePermission("products", "read"), handlers.GetCategories)
			auth.GET("/products/:id", middleware.RequirePermission("products", "read"), handlers.GetProduct)
			auth.POST("/products", middleware.RequirePermission("products", "create"), handlers.CreateProduct)
			auth.PUT("/products/:id", middleware.RequirePermission("products", "update"), handlers.UpdateProduct)
			auth.DELETE("/products/:id", middleware.RequirePermission("products", "delete"), handlers.DeleteProduct)

			auth.GET("/users", middleware.RequirePermission("users", "read"), handlers.GetUsers)
			auth.GET("/users/:id", middleware.RequirePermission("users", "read"), handlers.GetUser)
			auth.POST("/users", middleware.RequirePermission("users", "create"), handlers.CreateUser)
			auth.POST("/users/bulk", middleware.RequirePermission("users", "create"), handlers.BulkCreateUsers)
			auth.PUT("/users/:id", middleware.RequirePermission("users", "update"), handlers.UpdateUser)
			auth.DELETE("/users/:id", middleware.RequirePermission("users", "delete"), handlers.DeleteUser)

			auth.GET("/roles", middleware.RequirePermission("roles", "manage"), handlers.GetRoles)
			auth.POST("/roles", middleware.RequirePermission("roles", "manage"), handlers.CreateRole)
			auth.GET("/roles/:id", middleware.RequirePermission("roles", "manage"), handlers.GetRole)
			auth.GET("/permissions", middleware.RequirePermission("roles", "manage"), handlers.GetPermissions)

			auth.GET("/groups", handlers.GetGroups)
			auth.POST("/groups", middleware.RequirePermission("users", "create"), handlers.CreateGroup)

			auth.GET("/logs", middleware.RequirePermission("logs", "read"), handlers.GetActivityLogs)
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" { port = "8080" }
	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}
