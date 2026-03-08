package handlers

import (
	"fmt"
	"net/http"

	"product-mgmt/internal/database"
	"product-mgmt/internal/middleware"
	"product-mgmt/internal/models"

	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	var roles []models.Role
	database.DB.Preload("Permissions").Find(&roles)
	c.JSON(http.StatusOK, roles)
}

func GetRole(c *gin.Context) {
	var role models.Role
	if err := database.DB.Preload("Permissions").First(&role, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

func CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		PermIDs     []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role := models.Role{Name: req.Name, Description: req.Description}
	if len(req.PermIDs) > 0 {
		var perms []models.Permission
		database.DB.Find(&perms, req.PermIDs)
		role.Permissions = perms
	}
	if err := database.DB.Create(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role name already exists"})
		return
	}
	middleware.LogActivity(c, "create", "role", fmt.Sprintf("%d", role.ID), role.Name,
		fmt.Sprintf("Created role '%s' with %d permissions", role.Name, len(role.Permissions)), "success")
	c.JSON(http.StatusCreated, role)
}

func GetPermissions(c *gin.Context) {
	var perms []models.Permission
	database.DB.Find(&perms)
	c.JSON(http.StatusOK, perms)
}

func GetGroups(c *gin.Context) {
	var groups []models.Group
	database.DB.Preload("Users").Find(&groups)
	c.JSON(http.StatusOK, groups)
}

func CreateGroup(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group := models.Group{Name: req.Name, Description: req.Description}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group name already exists"})
		return
	}
	middleware.LogActivity(c, "create", "group", fmt.Sprintf("%d", group.ID), group.Name,
		fmt.Sprintf("Created group '%s'", group.Name), "success")
	c.JSON(http.StatusCreated, group)
}

func GetDashboardStats(c *gin.Context) {
	var userCount, productCount, roleCount, groupCount, logCount int64
	var recentLogs []models.ActivityLog
	database.DB.Model(&models.User{}).Count(&userCount)
	database.DB.Model(&models.Product{}).Count(&productCount)
	database.DB.Model(&models.Role{}).Count(&roleCount)
	database.DB.Model(&models.Group{}).Count(&groupCount)
	database.DB.Model(&models.ActivityLog{}).Count(&logCount)
	database.DB.Preload("User").Order("created_at desc").Limit(12).Find(&recentLogs)
	var successLogins, failedLogins int64
	database.DB.Model(&models.ActivityLog{}).Where("action = ? AND status = ?", "login", "success").Count(&successLogins)
	database.DB.Model(&models.ActivityLog{}).Where("action = ? AND status = ?", "login", "failed").Count(&failedLogins)
	c.JSON(http.StatusOK, gin.H{
		"users": userCount, "products": productCount,
		"roles": roleCount, "groups": groupCount,
		"total_logs": logCount, "success_logins": successLogins,
		"failed_logins": failedLogins, "recent_logs": recentLogs,
	})
}
