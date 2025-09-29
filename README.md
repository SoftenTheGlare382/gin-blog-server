# gin-blog-server

A modern blog system backend service built with Go language and Gin framework.

## Project Overview

gin-blog-server is a complete blog system backend implementation that provides core functionalities such as blog management, user authentication, content publishing, and comment interaction. This project adopts a modern Go web development architecture with well-structured modular design and extensibility.

## Technology Stack

- **Core Framework**: [Gin](https://gin-gonic.com/) - High-performance HTTP web framework
- **Database**: [GORM](https://gorm.io/) - Full-featured ORM library supporting MySQL and SQLite
- **Cache**: [Redis](https://redis.io/) - Used for session management, caching, and statistical data
- **Configuration Management**: [Viper](https://github.com/spf13/viper) - Complete configuration solution
- **Logging System**: [Slog](https://pkg.go.dev/log/slog) - Structured logging
- **Authentication**: [JWT](https://jwt.io/) - User authentication based on JSON Web Token
- **API Documentation**: [Swagger](https://swagger.io/) - Automatic API documentation generation

## Core Features

### User System
- User registration with email verification
- JWT Token authentication and refresh
- Role-based access control (RBAC)
- User information management and online status monitoring
- Password encryption storage (BCrypt)

### Blog Management
- Article publishing, editing, deletion, and categorization
- Tag system management
- Article topping and status management
- Article comments and reply functionality
- Friendship link management

### Content Interaction
- Comment moderation mechanism
- Like functionality (articles/comments)
- Message board system
- Visit statistics and data analysis

### System Management
- Menu and permission management
- Role-resource allocation
- Operation log recording
- File upload and management (local/Qiniu Cloud)
- System configuration management

## Project Structure

```
gin-blog-server/
├── assets/              # Static resource files
│   ├── ip2region.xdb    # IP address database
│   └── templates/       # Email templates
├── config.yml           # Configuration file
├── docs/                # API documentation
├── internal/            # Core code
│   ├── global/          # Global configuration and constants
│   ├── handle/          # API handler functions
│   ├── middleware/      # Middleware
│   ├── model/           # Data models
│   ├── utils/           # Utility functions
│   └── sql/             # SQL initialization scripts
├── go.mod               # Go module definition
└── main.go              # Application entry point
```


## Quick Start

### Requirements
- Go 1.23+
- MySQL 5.7+ or SQLite 3+
- Redis 5.0+

### Installation Steps

1. Clone the project
```bash
git clone <repository-url>
cd gin-blog-server
```


2. Configure environment
```bash
# Copy configuration file and modify as needed
cp config.yml.example config.yml
```


3. Install dependencies
```bash
go mod tidy
```


4. Run the project
```bash
go run main.go
```


### Configuration Guide

Configure the following key items in [config.yml](file://E:\goland\workplace\gin-blog\gin-blog-server\config.yml):

```yaml
Server:
  Port: localhost:8765         # Service port
  DbType: "mysql"              # Database type
  Mode: debug                  # Running mode

Mysql:
  Host: "localhost"            # MySQL host
  Port: 3306                   # MySQL port
  DbName: "gin_blog_db"        # Database name
  UserName: "root"             # Username
  Password: "root"             # Password

Redis:
  Addr: "localhost:6379"       # Redis address
  Password: ""                 # Redis password

JWT:
  Secret: "your-secret-key"    # JWT secret key
  Expire: 24                   # Expiration time (hours)
```


## API Documentation

The project integrates Swagger API documentation. After starting the service, visit:
```
http://localhost:8765/swagger/index.html
```


## Deployment

### Build Binary
```bash
go build -o gin-blog-server main.go
```


### Run Service
```bash
./gin-blog-server
```


### Docker Deployment (Optional)
```dockerfile
# Example Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o gin-blog-server main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/gin-blog-server .
COPY --from=builder /app/config.yml .
COPY --from=builder /app/assets/ ./assets/
EXPOSE 8765
CMD ["./gin-blog-server"]
```



