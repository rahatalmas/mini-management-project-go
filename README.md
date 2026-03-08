# ProductVault — Full Stack Management System

A complete Go + MySQL management system built for academic coursework.

## Tech Stack
- **Backend**: Go (Gin framework), GORM ORM
- **Database**: MySQL 8.0
- **Auth**: JWT (HS256) with HttpOnly cookie fallback
- **Frontend**: HTML5, Vanilla JS, Tailwind CSS

## Features
- ✅ JWT Token Authentication
- ✅ Role-Based Access Control (RBAC)
- ✅ Permission-based resource protection
- ✅ Bulk User Creation (JSON import)
- ✅ Login Activity Monitoring
- ✅ Product CRUD with category filtering
- ✅ User, Role, Group management
- ✅ Soft deletes (GORM)

## Default Credentials
```
Username: admin
Password: admin123
```

## Quick Start (Docker)
```bash
# Clone and start everything
docker-compose up --build

# Visit http://localhost:8080
```

## Manual Setup
```bash
# 1. Create MySQL database
mysql -u root -p -e "CREATE DATABASE product_mgmt;"

# 2. Copy and configure env
cp .env.example .env
# Edit .env with your DB credentials

# 3. Download dependencies
go mod tidy

# 4. Run
go run cmd/main.go
```

## API Endpoints

### Auth
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/auth/login | No | Login |
| POST | /api/v1/auth/logout | Yes | Logout |
| GET | /api/v1/auth/me | Yes | Current user |

### Products
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/products | products:read | List products |
| POST | /api/v1/products | products:create | Create product |
| PUT | /api/v1/products/:id | products:update | Update product |
| DELETE | /api/v1/products/:id | products:delete | Delete product |

### Users
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/users | users:read | List users |
| POST | /api/v1/users | users:create | Create single user |
| POST | /api/v1/users/bulk | users:create | Bulk create users |
| PUT | /api/v1/users/:id | users:update | Update user |
| DELETE | /api/v1/users/:id | users:delete | Delete user |

### Security
| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | /api/v1/roles | roles:manage | List roles |
| POST | /api/v1/roles | roles:manage | Create role |
| GET | /api/v1/permissions | roles:manage | List permissions |
| GET | /api/v1/groups | Any auth | List groups |
| GET | /api/v1/logs | logs:read | Activity logs |

## Bulk User Creation Example
```json
POST /api/v1/users/bulk
{
  "users": [
    {
      "username": "alice",
      "email": "alice@example.com",
      "password": "secret123",
      "first_name": "Alice",
      "last_name": "Smith",
      "role_ids": [3],
      "group_ids": [1]
    }
  ]
}
```

## Default Roles & Permissions
| Role | Permissions |
|------|------------|
| admin | All permissions |
| manager | products:* |
| viewer | *:read |

## Project Structure
```
product-mgmt/
├── cmd/
│   └── main.go              # Entry point & routes
├── internal/
│   ├── models/
│   │   └── models.go        # GORM models
│   ├── database/
│   │   └── database.go      # DB connection & seeding
│   ├── middleware/
│   │   └── auth.go          # JWT middleware, RBAC
│   └── handlers/
│       ├── auth.go          # Login/logout
│       ├── users.go         # User CRUD + bulk
│       ├── products.go      # Product CRUD
│       └── roles.go         # Roles, groups, dashboard
├── frontend/
│   └── templates/
│       └── index.html       # Single-page app
├── docker-compose.yml
├── Dockerfile
└── go.mod
```
