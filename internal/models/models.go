package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Resource    string    `gorm:"size:100" json:"resource"`
	Action      string    `gorm:"size:100" json:"action"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Group struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:100" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Users       []User    `gorm:"many2many:user_groups;" json:"users,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password     string         `gorm:"not null;size:255" json:"-"`
	FirstName    string         `gorm:"size:100" json:"first_name"`
	LastName     string         `gorm:"size:100" json:"last_name"`
	Bio          string         `gorm:"size:500" json:"bio"`
	Avatar       string         `gorm:"type:mediumtext" json:"avatar"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Roles        []Role         `gorm:"many2many:user_roles;" json:"roles"`
	Groups       []Group        `gorm:"many2many:user_groups;" json:"groups"`
	ActivityLogs []ActivityLog  `gorm:"foreignKey:UserID" json:"activity_logs,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// ActivityLog tracks ALL system events
type ActivityLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action       string    `gorm:"size:50" json:"action"`
	Resource     string    `gorm:"size:50" json:"resource"`
	ResourceID   string    `gorm:"size:50" json:"resource_id"`
	ResourceName string    `gorm:"size:255" json:"resource_name"`
	Detail       string    `gorm:"size:1000" json:"detail"`
	Status       string    `gorm:"size:20" json:"status"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `gorm:"size:512" json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
}

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Price       float64        `gorm:"not null" json:"price"`
	Stock       int            `gorm:"default:0" json:"stock"`
	Category    string         `gorm:"size:100" json:"category"`
	SKU         string         `gorm:"uniqueIndex;size:100" json:"sku"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedBy   uint           `json:"created_by"`
	UpdatedBy   uint           `json:"updated_by"`
	Creator     User           `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Claims struct {
	UserID   uint     `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expires_at"`
}

type BulkUserRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required"`
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleIDs   []uint `json:"role_ids"`
	GroupIDs  []uint `json:"group_ids"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar"`
	Password  string `json:"password"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	IsActive    *bool   `json:"is_active"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
	SKU         string  `json:"sku" binding:"required"`
}
