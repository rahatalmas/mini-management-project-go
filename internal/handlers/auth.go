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

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	result := database.DB.Preload("Roles.Permissions").Preload("Groups").
		Where("username = ? OR email = ?", req.Username, req.Username).First(&user)

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	if result.Error != nil || !user.IsActive {
		if user.ID != 0 {
			database.DB.Create(&models.ActivityLog{
				UserID: user.ID, Action: "login", Resource: "auth",
				Detail: "Login failed - account inactive or not found", Status: "failed",
				IPAddress: ip, UserAgent: ua,
			})
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		database.DB.Create(&models.ActivityLog{
			UserID: user.ID, Action: "login", Resource: "auth",
			Detail: "Login failed - wrong password", Status: "failed",
			IPAddress: ip, UserAgent: ua,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	token, expiresAt, err := middleware.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	database.DB.Create(&models.ActivityLog{
		UserID: user.ID, Action: "login", Resource: "auth",
		ResourceName: user.Username, Detail: "Successful login", Status: "success",
		IPAddress: ip, UserAgent: ua,
	})
	c.SetCookie("auth_token", token, 86400, "/", "", false, true)
	c.JSON(http.StatusOK, models.LoginResponse{Token: token, User: user, ExpiresAt: expiresAt})
}

func Logout(c *gin.Context) {
	middleware.LogActivity(c, "logout", "auth", "", "", "User logged out", "success")
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func GetMe(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, user)
}

func UpdateProfile(c *gin.Context) {
	user, _ := c.Get("user")
	u := user.(models.User)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	changes := []string{}
	if req.FirstName != "" && req.FirstName != u.FirstName {
		u.FirstName = req.FirstName
		changes = append(changes, "first_name")
	}
	if req.LastName != "" && req.LastName != u.LastName {
		u.LastName = req.LastName
		changes = append(changes, "last_name")
	}
	if req.Email != "" && req.Email != u.Email {
		u.Email = req.Email
		changes = append(changes, "email")
	}
	if req.Bio != u.Bio {
		u.Bio = req.Bio
		changes = append(changes, "bio")
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
		changes = append(changes, "avatar")
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		u.Password = string(hash)
		changes = append(changes, "password")
	}

	database.DB.Save(&u)
	database.DB.Preload("Roles").Preload("Groups").First(&u, u.ID)

	detail := fmt.Sprintf("Profile updated: %v", changes)
	middleware.LogActivity(c, "update", "profile", fmt.Sprintf("%d", u.ID), u.Username, detail, "success")

	c.JSON(http.StatusOK, u)
}
