package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"product-mgmt/internal/database"
	"product-mgmt/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateToken(user models.User) (string, int64, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expiresAt, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString, expiresAt, err
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			cookie, err := c.Cookie("auth_token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token"})
				c.Abort()
				return
			}
			authHeader = "Bearer " + cookie
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		var user models.User
		database.DB.Preload("Roles.Permissions").Preload("Groups").First(&user, claims.UserID)
		if user.ID == 0 || !user.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or inactive"})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		u := user.(models.User)
		for _, role := range u.Roles {
			if role.Name == "admin" {
				c.Next()
				return
			}
			for _, perm := range role.Permissions {
				if perm.Resource == resource && (perm.Action == action || perm.Action == "manage") {
					c.Next()
					return
				}
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions", "required": resource + ":" + action})
		c.Abort()
	}
}

// LogActivity helper - call from any handler
func LogActivity(c *gin.Context, action, resource, resourceID, resourceName, detail, status string) {
	user, exists := c.Get("user")
	if !exists {
		return
	}
	u := user.(models.User)
	database.DB.Create(&models.ActivityLog{
		UserID:       u.ID,
		Action:       action,
		Resource:     resource,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Detail:       detail,
		Status:       status,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	})
}
