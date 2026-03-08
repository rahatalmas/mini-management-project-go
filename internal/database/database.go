package database

import (
	"fmt"
	"log"
	"os"

	"product-mgmt/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.Group{},
		&models.User{},
		&models.ActivityLog{},
		&models.Product{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database migrated successfully")
}

func Seed() {
	permissions := []models.Permission{
		{Name: "products:read", Resource: "products", Action: "read", Description: "View products"},
		{Name: "products:create", Resource: "products", Action: "create", Description: "Create products"},
		{Name: "products:update", Resource: "products", Action: "update", Description: "Update products"},
		{Name: "products:delete", Resource: "products", Action: "delete", Description: "Delete products"},
		{Name: "users:read", Resource: "users", Action: "read", Description: "View users"},
		{Name: "users:create", Resource: "users", Action: "create", Description: "Create users"},
		{Name: "users:update", Resource: "users", Action: "update", Description: "Update users"},
		{Name: "users:delete", Resource: "users", Action: "delete", Description: "Delete users"},
		{Name: "roles:manage", Resource: "roles", Action: "manage", Description: "Manage roles"},
		{Name: "logs:view", Resource: "logs", Action: "read", Description: "View activity logs"},
	}
	for _, p := range permissions {
		DB.FirstOrCreate(&p, models.Permission{Name: p.Name})
	}

	var allPerms, readPerms, productPerms []models.Permission
	DB.Find(&allPerms)
	DB.Where("action = ?", "read").Find(&readPerms)
	DB.Where("resource = ?", "products").Find(&productPerms)

	roleData := []struct {
		role  models.Role
		perms []models.Permission
	}{
		{models.Role{Name: "admin", Description: "Full system access"}, allPerms},
		{models.Role{Name: "manager", Description: "Manage products and view users"}, productPerms},
		{models.Role{Name: "viewer", Description: "Read-only access"}, readPerms},
	}
	for _, r := range roleData {
		var role models.Role
		DB.FirstOrCreate(&role, models.Role{Name: r.role.Name})
		role.Description = r.role.Description
		role.Permissions = r.perms
		DB.Save(&role)
	}

	groups := []models.Group{
		{Name: "Engineering", Description: "Engineering team"},
		{Name: "Marketing", Description: "Marketing team"},
		{Name: "Operations", Description: "Operations team"},
	}
	for _, g := range groups {
		DB.FirstOrCreate(&g, models.Group{Name: g.Name})
	}

	var adminRole models.Role
	DB.Where("name = ?", "admin").First(&adminRole)

	var adminUser models.User
	if DB.Where("username = ?", "admin").First(&adminUser).Error != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser = models.User{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  string(hash),
			FirstName: "System",
			LastName:  "Admin",
			IsActive:  true,
			Roles:     []models.Role{adminRole},
		}
		DB.Create(&adminUser)
		log.Println("Admin user created: admin / admin123")
	}

	var productCount int64
	DB.Model(&models.Product{}).Count(&productCount)
	if productCount == 0 {
		DB.Create(&[]models.Product{
			{Name: "Laptop Pro X1", Description: "High-performance laptop", Price: 1299.99, Stock: 50, Category: "Electronics", SKU: "LAP-001", CreatedBy: adminUser.ID},
			{Name: "Wireless Mouse", Description: "Ergonomic wireless mouse", Price: 29.99, Stock: 200, Category: "Accessories", SKU: "MOU-001", CreatedBy: adminUser.ID},
			{Name: "USB-C Hub", Description: "7-in-1 USB-C hub", Price: 49.99, Stock: 150, Category: "Accessories", SKU: "HUB-001", CreatedBy: adminUser.ID},
			{Name: "4K Monitor", Description: "27-inch 4K display", Price: 599.99, Stock: 30, Category: "Electronics", SKU: "MON-001", CreatedBy: adminUser.ID},
			{Name: "Mechanical Keyboard", Description: "RGB mechanical keyboard", Price: 89.99, Stock: 75, Category: "Accessories", SKU: "KEY-001", CreatedBy: adminUser.ID},
		})
	}
	log.Println("Database seeded successfully")
}
