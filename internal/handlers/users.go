package handlers

import (
	"fmt"
	"net/http"

	"product-mgmt/internal/database"
	"product-mgmt/internal/middleware"
	"product-mgmt/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	database.DB.Preload("Roles").Preload("Groups").Find(&users)
	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	var user models.User
	if err := database.DB.Preload("Roles.Permissions").Preload("Groups").First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := createSingleUser(req)
	if err != nil {
		middleware.LogActivity(c, "create", "user", "", req.Username, "Failed to create user: "+err.Error(), "failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	middleware.LogActivity(c, "create", "user", fmt.Sprintf("%d", user.ID), user.Username,
		fmt.Sprintf("Created user '%s' (%s)", user.Username, user.Email), "success")
	c.JSON(http.StatusCreated, user)
}

func BulkCreateUsers(c *gin.Context) {
	var req models.BulkUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	results := make([]gin.H, 0, len(req.Users))
	successCount, failCount := 0, 0
	for _, userReq := range req.Users {
		user, err := createSingleUser(userReq)
		if err != nil {
			results = append(results, gin.H{"username": userReq.Username, "status": "failed", "error": err.Error()})
			failCount++
		} else {
			results = append(results, gin.H{"username": user.Username, "id": user.ID, "status": "created"})
			successCount++
		}
	}
	middleware.LogActivity(c, "bulk_create", "user", "", "",
		fmt.Sprintf("Bulk created %d users (%d success, %d failed)", len(req.Users), successCount, failCount), "success")
	c.JSON(http.StatusOK, gin.H{"total": len(req.Users), "success": successCount, "failed": failCount, "results": results})
}

func createSingleUser(req models.CreateUserRequest) (models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		Username: req.Username, Email: req.Email, Password: string(hash),
		FirstName: req.FirstName, LastName: req.LastName, IsActive: true,
	}
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		database.DB.Find(&roles, req.RoleIDs)
		user.Roles = roles
	}
	if len(req.GroupIDs) > 0 {
		var groups []models.Group
		database.DB.Find(&groups, req.GroupIDs)
		user.Groups = groups
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return models.User{}, err
	}
	database.DB.Preload("Roles").Preload("Groups").First(&user, user.ID)
	return user, nil
}

func UpdateUser(c *gin.Context) {
	var user models.User
	if err := database.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		IsActive  *bool  `json:"is_active"`
		RoleIDs   []uint `json:"role_ids"`
		GroupIDs  []uint `json:"group_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.FirstName != "" { user.FirstName = req.FirstName }
	if req.LastName != "" { user.LastName = req.LastName }
	if req.Email != "" { user.Email = req.Email }
	if req.IsActive != nil { user.IsActive = *req.IsActive }
	database.DB.Save(&user)
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		database.DB.Find(&roles, req.RoleIDs)
		database.DB.Model(&user).Association("Roles").Replace(roles)
	}
	if len(req.GroupIDs) > 0 {
		var groups []models.Group
		database.DB.Find(&groups, req.GroupIDs)
		database.DB.Model(&user).Association("Groups").Replace(groups)
	}
	database.DB.Preload("Roles").Preload("Groups").First(&user, user.ID)
	middleware.LogActivity(c, "update", "user", fmt.Sprintf("%d", user.ID), user.Username,
		fmt.Sprintf("Updated user '%s'", user.Username), "success")
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	currentUser, _ := c.Get("user")
	cu := currentUser.(models.User)
	var user models.User
	if err := database.DB.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.ID == cu.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}
	middleware.LogActivity(c, "delete", "user", fmt.Sprintf("%d", user.ID), user.Username,
		fmt.Sprintf("Deleted user '%s' (%s)", user.Username, user.Email), "success")
	database.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetActivityLogs(c *gin.Context) {
	var logs []models.ActivityLog
	query := database.DB.Preload("User").Order("created_at desc")
	if uid := c.Query("user_id"); uid != "" { query = query.Where("user_id = ?", uid) }
	if status := c.Query("status"); status != "" { query = query.Where("status = ?", status) }
	if resource := c.Query("resource"); resource != "" { query = query.Where("resource = ?", resource) }
	if action := c.Query("action"); action != "" { query = query.Where("action = ?", action) }
	query.Limit(200).Find(&logs)
	c.JSON(http.StatusOK, logs)
}
