# Sistem Inventaris - Inventory Management System

A robust, enterprise-grade Inventory Management System built with Go (Golang) using the Gin framework. This system provides comprehensive inventory control capabilities including stock management, purchase orders, goods receipts, warehouse operations, and real-time monitoring.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Project Architecture](#project-architecture)
- [API Endpoints](#api-endpoints)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Testing](#testing)
- [Documentation](#documentation)
- [Deployment](#deployment)
- [BNSP Senior Programmer Competencies](#bnsp-senior-programmer-competencies)

---

## Overview

This Inventory Management System is designed to handle complex inventory operations for businesses of all sizes. It provides real-time stock tracking, multi-warehouse support, purchase and sales workflow management, and comprehensive reporting capabilities.

### Key Capabilities

- **Real-time Inventory Tracking**: Monitor stock levels across multiple warehouses
- **Multi-Warehouse Support**: Manage inventory across different locations
- **Purchase Order Management**: Streamline procurement processes
- **Goods Receipt Handling**: Track incoming inventory with ease
- **Stock Operations**: Transfers, adjustments, and stock opname
- **Sales Order Processing**: Handle customer orders efficiently
- **Delivery Order Management**: Coordinate outbound logistics
- **Role-Based Access Control**: Secure user authentication and authorization
- **Dashboard & Monitoring**: Real-time insights and alerts

---

## Features

### Core Inventory Features

| Feature | Description |
|---------|-------------|
| Stock Movement | Track IN, OUT, TRANSFER, and ADJUSTMENT movements |
| Stock Transfer | Transfer inventory between warehouses |
| Stock Adjustment | Correct inventory discrepancies |
| Stock Opname | Perform physical stock counts and reconciliation |
| Low Stock Alerts | Automated notifications for low inventory levels |
| Stock History | Complete audit trail of inventory changes |

### Business Process Features

- **Purchase Orders**: Create, track, and manage purchase requests
- **Goods Receipts**: Record and validate incoming inventory
- **Sales Orders**: Manage customer orders and fulfillment
- **Delivery Orders**: Coordinate outbound shipments
- **Customer Management**: Maintain customer database
- **Supplier Management**: Manage vendor relationships

### System Features

- **JWT Authentication**: Secure API access with token-based auth
- **Role-Based Access**: Granular permission management
- **Structured Logging**: Comprehensive logging with zerolog
- **API Validation**: Request validation using go-playground/validator
- **Error Handling**: Custom error handling and response formatting
- **Graceful Shutdown**: Proper server shutdown handling
- **Multi-Environment Support**: Configurable for dev, staging, and prod

---

## Technology Stack

### Backend

| Technology | Purpose |
|------------|---------|
| **Go 1.21+** | Programming language |
| **Gin Gonic** | HTTP web framework |
| **GORM** | ORM for database operations |
| **PostgreSQL** | Primary database |
| **Zerolog** | Structured logging |
| **JWT** | Token-based authentication |
| **UUID** | Unique identifier generation |

### Development & Quality

| Tool | Purpose |
|------|---------|
| **GolangCI-Lint** | Code linting |
| **Swag** | Swagger documentation generation |
| **Docker** | Containerization |
| **Make** | Build automation |
| **Go Test** | Unit and integration testing |
| **Go Bench** | Performance benchmarking |

### Infrastructure

| Technology | Purpose |
|------------|---------|
| **Docker Compose** | Multi-container orchestration |
| **PostgreSQL** | Database server |
| **Nginx** (optional) | Reverse proxy |

---

## Project Architecture

```
sistem_inventaris/
├── cmd/
│   └── api/
│       ├── main.go                    # Application entry point
│       └── docs/                      # Swagger documentation
├── internal/
│   ├── common/                        # Shared utilities
│   │   ├── enum/                      # Enumeration types
│   │   └── type/                      # Common type definitions
│   ├── config/                        # Configuration management
│   ├── dto/                           # Data Transfer Objects
│   │   ├── auth/                      # Authentication DTOs
│   │   ├── role/                      # Role DTOs
│   │   ├── inventory.go
│   │   ├── monitoring.go
│   │   ├── product.go
│   │   ├── product_category.go
│   │   ├── product_stock.go
│   │   ├── purchase.go
│   │   ├── sales.go
│   │   └── warehouse.go
│   ├── entity/                        # Database entities
│   │   ├── assessment.go
│   │   ├── inventory.go
│   │   ├── monitoring.go
│   │   ├── permission.entity.go
│   │   ├── product.go
│   │   ├── purchase.go
│   │   ├── query_catalog.entity.go
│   │   ├── report.go
│   │   ├── role.entity.go
│   │   ├── sales.go
│   │   └── user.entity.go
│   ├── handler/                       # HTTP handlers/controllers
│   │   ├── auth/                      # Authentication handlers
│   │   ├── customer/                  # Customer management
│   │   ├── delivery_order/            # Delivery order handlers
│   │   ├── goods_receipt/             # Goods receipt handlers
│   │   ├── inventory/                 # Inventory handlers
│   │   ├── monitoring/                # Monitoring handlers
│   │   ├── product/                   # Product handlers
│   │   ├── product_category/          # Category handlers
│   │   ├── product_stock/             # Stock handlers
│   │   ├── purchase_order/            # Purchase order handlers
│   │   ├── query_catalog/             # Query catalog handlers
│   │   ├── role/                      # Role handlers
│   │   ├── sales_order/               # Sales order handlers
│   │   ├── supplier/                  # Supplier handlers
│   │   ├── warehouse/                 # Warehouse handlers
│   │   └── warehouse_location/        # Warehouse location handlers
│   ├── pkg/                           # External packages (db, logger, middleware, etc.)
│   ├── repository/                    # Data access layer
│   │   ├── assessment_item.repository.go
│   │   ├── assessment.repository.go
│   │   ├── auth.repository.go
│   │   ├── customer.repository.go
│   │   ├── delivery_order.repository.go
│   │   ├── goods_receipt.repository.go
│   │   ├── inventory.repository.go
│   │   ├── permission.repository.go
│   │   ├── product_category.repository.go
│   │   ├── product.repository.go
│   │   ├── product_stock.repository.go
│   │   ├── purchase_order.repository.go
│   │   ├── query_catalog.repository.go
│   │   ├── role.repository.go
│   │   ├── sales_order.repository.go
│   │   ├── supplier.repository.go
│   │   ├── warehouse.repository.go
│   │   └── warehouse_location.repository.go
│   ├── server/                        # Server configuration
│   │   └── api.go                     # API setup and routing
│   ├── service/                       # Business logic layer
│   │   ├── assessment_item.service.go
│   │   ├── assessment.service.go
│   │   ├── auth.service.go
│   │   ├── backup.service.go
│   │   ├── customer.service.go
│   │   ├── dashboard.service.go
│   │   ├── delivery_order.service.go
│   │   ├── goods_receipt.service.go
│   │   ├── inventory.service.go
│   │   ├── monitoring.service.go
│   │   ├── notification.service.go
│   │   ├── product.service.go
│   │   ├── product_category.service.go
│   │   ├── product_stock.service.go
│   │   ├── purchase_order.service.go
│   │   ├── purchase.service.go
│   │   ├── query_catalog.service.go
│   │   ├── report.service.go
│   │   ├── role.service.go
│   │   ├── sales_order.service.go
│   │   ├── supplier.service.go
│   │   ├── warehouse.service.go
│   │   └── warehouse_location.service.go
│   └── utils/                         # Utility functions
├── migrations/
│   └── dump.sql                       # Database schema
├── api/
│   └── v1/
│       └── product.swagger.json       # Swagger specification
├── postman/
│   └── import.json                    # Postman collection
├── tests/
│   ├── integration/                   # Integration tests
│   └── stress/                        # Stress tests
├── .env.example                       # Environment template
├── .gitignore                         # Git ignore rules
├── docker-compose.yml                 # Docker composition
├── Dockerfile                         # Docker image definition
├── go.mod                             # Go module dependencies
├── go.sum                             # Go module checksums
├── Makefile                           # Build targets
└── README.md                          # This file
```

### Architecture Pattern

This project follows a **Layered Architecture** pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                       │
│                    (Handlers/Routes)                         │
├─────────────────────────────────────────────────────────────┤
│                      Business Layer                          │
│                      (Services)                              │
├─────────────────────────────────────────────────────────────┤
│                       Data Layer                             │
│                  (Repositories/Entities)                     │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure                            │
│              (Config, Database, Logging)                     │
└─────────────────────────────────────────────────────────────┘
```

---

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/register` | User registration |
| GET | `/api/v1/auth/profile` | Get user profile |
| PUT | `/api/v1/auth/profile` | Update profile |

### Products

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/products` | List all products |
| POST | `/api/v1/products` | Create product |
| GET | `/api/v1/products/:id` | Get product by ID |
| PUT | `/api/v1/products/:id` | Update product |
| DELETE | `/api/v1/products/:id` | Delete product |
| GET | `/api/v1/products/:id/stock` | Get product stock |

### Product Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/categories` | List categories |
| POST | `/api/v1/categories` | Create category |
| GET | `/api/v1/categories/:id` | Get category |
| PUT | `/api/v1/categories/:id` | Update category |
| DELETE | `/api/v1/categories/:id` | Delete category |

### Inventory

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory/movements` | List stock movements |
| POST | `/api/v1/inventory/movements` | Create stock movement |
| POST | `/api/v1/inventory/transfer` | Transfer stock between warehouses |
| POST | `/api/v1/inventory/adjustment` | Adjust stock quantity |
| POST | `/api/v1/inventory/opname` | Perform stock opname |
| GET | `/api/v1/inventory/product/:id` | Get stock by product |
| GET | `/api/v1/inventory/warehouse/:id` | Get stock by warehouse |
| GET | `/api/v1/inventory/summary` | Get stock summary |
| GET | `/api/v1/inventory/alerts/low-stock` | Get low stock alerts |
| GET | `/api/v1/inventory/history/:product_id` | Get stock history |

### Purchase Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/purchase-orders` | List purchase orders |
| POST | `/api/v1/purchase-orders` | Create purchase order |
| GET | `/api/v1/purchase-orders/:id` | Get purchase order |
| PUT | `/api/v1/purchase-orders/:id` | Update purchase order |
| PUT | `/api/v1/purchase-orders/:id/status` | Update order status |

### Goods Receipts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/goods-receipts` | List goods receipts |
| POST | `/api/v1/goods-receipts` | Create goods receipt |
| GET | `/api/v1/goods-receipts/:id` | Get goods receipt |
| PUT | `/api/v1/goods-receipts/:id` | Update goods receipt |

### Sales Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/sales-orders` | List sales orders |
| POST | `/api/v1/sales-orders` | Create sales order |
| GET | `/api/v1/sales-orders/:id` | Get sales order |
| PUT | `/api/v1/sales-orders/:id` | Update sales order |

### Delivery Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/delivery-orders` | List delivery orders |
| POST | `/api/v1/delivery-orders` | Create delivery order |
| GET | `/api/v1/delivery-orders/:id` | Get delivery order |
| PUT | `/api/v1/delivery-orders/:id` | Update delivery order |

### Warehouse

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/warehouses` | List warehouses |
| POST | `/api/v1/warehouses` | Create warehouse |
| GET | `/api/v1/warehouses/:id` | Get warehouse |
| PUT | `/api/v1/warehouses/:id` | Update warehouse |
| DELETE | `/api/v1/warehouses/:id` | Delete warehouse |

### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/customers` | List customers |
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers/:id` | Get customer |
| PUT | `/api/v1/customers/:id` | Update customer |
| DELETE | `/api/v1/customers/:id` | Delete customer |

### Suppliers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/suppliers` | List suppliers |
| POST | `/api/v1/suppliers` | Create supplier |
| GET | `/api/v1/suppliers/:id` | Get supplier |
| PUT | `/api/v1/suppliers/:id` | Update supplier |
| DELETE | `/api/v1/suppliers/:id` | Delete supplier |

### Roles & Permissions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/roles` | List roles |
| POST | `/api/v1/roles` | Create role |
| GET | `/api/v1/roles/:id` | Get role |
| PUT | `/api/v1/roles/:id` | Update role |
| DELETE | `/api/v1/roles/:id` | Delete role |
| GET | `/api/v1/roles/:id/permissions` | Get role permissions |

### Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/monitoring/dashboard` | Dashboard data |
| GET | `/api/v1/monitoring/stats` | System statistics |
| GET | `/api/v1/monitoring/health` | Health check |

---

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Docker & Docker Compose (optional)
- Make

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd sistem_inventaris
   ```

2. **Install dependencies**
   ```bash
   make install
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Setup database**
   ```bash
   # Using Docker Compose
   docker-compose up -d postgres
   
   # Or connect to your PostgreSQL instance and run migrations
   psql -U postgres -d inventory < migrations/dump.sql
   ```

5. **Run the application**
   ```bash
   make run_api
   ```

### Development Commands

| Command | Description |
|---------|-------------|
| `make install` | Download and verify Go modules |
| `make run_api` | Run the API server |
| `make build_api` | Build the API binary |
| `make build_run_api` | Build and run the API |
| `make lint` | Run linter |
| `make lint_fix` | Run linter with auto-fix |
| `make test` | Run all tests |
| `make test_coverage` | Run tests with coverage report |
| `make benchmark` | Run benchmarks |
| `make docs` | Generate Swagger documentation |
| `make clean` | Clean build artifacts |

---

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```env
# Application
APP_NAME=api-jellyfish
APP_ENV=development
APP_VERSION=1.0.0
DEBUG=true
LOG_LEVEL=debug

# Server
APP_PORT=9000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=inventory
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRY_HOURS=24

# Redis (optional)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Configuration Structure

The configuration is managed through the [`internal/config/config.go`](internal/config/config.go) package, which loads settings from environment variables and provides type-safe access throughout the application.

---

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run with coverage
make test_coverage

# View coverage report
open coverage.html
```

### Benchmark Tests

```bash
# Run benchmarks
make benchmark
```

### Integration Tests

Integration tests are located in `tests/integration/` and test the complete API workflow including database interactions.

### Stress Tests

Stress tests are located in `tests/stress/` and test the API under high load conditions.

---

## Documentation

### Swagger API Documentation

The API is documented using Swagger/OpenAPI 2.0. To generate and view the documentation:

```bash
# Generate Swagger docs
make docs
```

Generated documentation is available at:
- `cmd/api/docs/docs.go`
- `cmd/api/docs/swagger.json`
- `cmd/api/docs/swagger.yaml`

### API Documentation Endpoint

When the server is running, access Swagger UI at:
```
http://localhost:9000/swagger/index.html
```

### Postman Collection

A Postman collection is available in `postman/import.json` for testing the API endpoints.

---

## Deployment

### Docker Deployment

1. **Build the image**
   ```bash
   docker build -t inventory-api .
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

### Production Build

```bash
# Build for multiple platforms
make build_api

# The binaries will be in bin/api/
ls bin/api/
```

### Environment Considerations

- Set `DEBUG=false` in production
- Use a secure `JWT_SECRET`
- Configure proper logging level (`info` or `warn`)
- Set up proper database connection pooling
- Use HTTPS in production
- Configure proper CORS settings

---

## BNSP Senior Programmer Competencies

This project demonstrates the following competencies required for BNSP Senior Programmer certification:

### 1. Programming Fundamentals

- **Go Language Proficiency**: Clean, idiomatic Go code using proper conventions
- **Data Structures**: Proper use of slices, maps, structs, and interfaces
- **Concurrency**: Goroutines for concurrent operations, sync.WaitGroup for synchronization
- **Error Handling**: Custom error types, proper error propagation, and graceful failure handling

### 2. Object-Oriented Design

- **Encapsulation**: Proper package organization with clear public/private boundaries
- **Interface Segregation**: Small, focused interfaces (e.g., `IHandler`, `IInventoryService`)
- **Dependency Injection**: Services injected into handlers via interfaces
- **SOLID Principles**: Single responsibility, open/closed, Liskov substitution, interface segregation, dependency inversion

### 3. Database Operations

- **ORM Usage**: GORM for database operations with proper model definitions
- **Transactions**: Database transactions for atomic operations
- **Migrations**: Database schema management via SQL migrations
- **Query Optimization**: Efficient queries with proper indexing considerations

### 4. API Development

- **RESTful Design**: Proper HTTP methods, status codes, and resource naming
- **Middleware**: Custom middleware for logging, auth, validation, and error handling
- **Validation**: Request validation using go-playground/validator
- **Documentation**: Swagger/OpenAPI documentation with swag

### 5. Security

- **Authentication**: JWT-based authentication with secure token handling
- **Authorization**: Role-based access control (RBAC) implementation
- **Input Validation**: All user inputs validated before processing
- **Secure Headers**: Proper HTTP headers for security

### 6. Testing

- **Unit Testing**: Test functions for individual components
- **Integration Testing**: Tests for complete workflows
- **Benchmark Testing**: Performance testing with go test -bench
- **Test Coverage**: Code coverage reporting

### 7. DevOps & CI/CD

- **Docker**: Containerization for consistent deployments
- **Docker Compose**: Multi-container orchestration
- **Makefile**: Automated build and deployment commands
- **Linting**: golangci-lint for code quality

### 8. Documentation

- **Code Documentation**: Godoc-style comments for all public APIs
- **README**: Comprehensive project documentation
- **Swagger**: API documentation with examples
- **Comments**: Inline comments for complex logic

### 9. Error Handling & Logging

- **Structured Logging**: Zerolog for structured, contextual logging
- **Error Types**: Custom error types with error wrapping
- **Response Formatting**: Consistent API response structure
- **Debugging**: Proper log levels and context

### 10. Performance Optimization

- **Connection Pooling**: Database connection pooling configuration
- **Graceful Shutdown**: Proper server shutdown handling
- **Middleware Efficiency**: Lightweight middleware without unnecessary overhead
- **Benchmark Tests**: Performance benchmarks for critical paths

---

## License

This project is developed for BNSP Senior Programmer certification purposes.

---

## Contributing

For certification purposes, this project demonstrates best practices in Go backend development. For any issues or improvements, please contact the development team.

---

## Support

For questions about this project or BNSP Senior Programmer certification requirements, please refer to the certification guidelines or contact your program administrator.
