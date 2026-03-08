package handlers

import (
	"fmt"
	"net/http"

	"product-mgmt/internal/database"
	"product-mgmt/internal/middleware"
	"product-mgmt/internal/models"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	var products []models.Product
	query := database.DB.Preload("Creator")
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	query.Find(&products)
	c.JSON(http.StatusOK, products)
}

func GetProduct(c *gin.Context) {
	var product models.Product
	if err := database.DB.Preload("Creator").First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, _ := c.Get("user")
	u := user.(models.User)
	product := models.Product{
		Name: req.Name, Description: req.Description, Price: req.Price,
		Stock: req.Stock, Category: req.Category, SKU: req.SKU,
		IsActive: true, CreatedBy: u.ID, UpdatedBy: u.ID,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		middleware.LogActivity(c, "create", "product", "", req.Name, "Failed to create product: "+err.Error(), "failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists or invalid data"})
		return
	}
	database.DB.Preload("Creator").First(&product, product.ID)
	middleware.LogActivity(c, "create", "product", fmt.Sprintf("%d", product.ID), product.Name,
		fmt.Sprintf("Created product '%s' (SKU: %s, Price: $%.2f)", product.Name, product.SKU, product.Price), "success")
	c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
	var product models.Product
	if err := database.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, _ := c.Get("user")
	u := user.(models.User)
	oldName := product.Name
	if req.Name != "" { product.Name = req.Name }
	if req.Description != "" { product.Description = req.Description }
	if req.Price > 0 { product.Price = req.Price }
	if req.Stock >= 0 { product.Stock = req.Stock }
	if req.Category != "" { product.Category = req.Category }
	if req.IsActive != nil { product.IsActive = *req.IsActive }
	product.UpdatedBy = u.ID
	database.DB.Save(&product)
	database.DB.Preload("Creator").First(&product, product.ID)
	middleware.LogActivity(c, "update", "product", fmt.Sprintf("%d", product.ID), oldName,
		fmt.Sprintf("Updated product '%s' — price: $%.2f, stock: %d", product.Name, product.Price, product.Stock), "success")
	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	var product models.Product
	if err := database.DB.First(&product, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	middleware.LogActivity(c, "delete", "product", fmt.Sprintf("%d", product.ID), product.Name,
		fmt.Sprintf("Deleted product '%s' (SKU: %s)", product.Name, product.SKU), "success")
	database.DB.Delete(&product)
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func GetCategories(c *gin.Context) {
	var categories []string
	database.DB.Model(&models.Product{}).Distinct("category").Pluck("category", &categories)
	c.JSON(http.StatusOK, categories)
}
