# Customer Complaint Management System API

A RESTful API backend for managing customer complaints, built with Go and Fiber framework.

## Features

- User authentication (registration/login) with JWT
- Customer management
- Complaint tracking with priority levels and status updates
- Comment system for complaints
- Category management for complaints

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: [Fiber](https://github.com/gofiber/fiber)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL
- **Authentication**: JWT (JSON Web Tokens)
- **Environment**: Docker and Docker Compose

## Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- Make (for using the Makefile commands)
- PostgreSQL

## Environment Setup

Create a `.env` file in the root directory with the following variables:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=complaint_db
DB_SSLMODE=disable
JWT_SECRET=your-secret-key
```

## Installation and Running

### Run Docker

```bash
# Run with Docker Compose (with make)
make docker-compose

# Run with Docker Compose (without make)
docker compose up -d --build
```

### Local Development

```bash
# Build the project (with make)
make build

# Build the project (without make)
go build -o bin/main main.go


# Run the application (with make)
make run

# Run the application (without make)
go run main.go
```

## API Endpoints

### Authentication
- `POST /register` - Register a new user
- `POST /login` - Login and get JWT token

### Protected Routes (require authentication)
- `GET /api/users` - Get all users

- `POST /api/customers/create` - Create a new customer
- `GET /api/customers` - Get all customers

- `POST /api/complaints/create` - Create a new complaint
- `PUT /api/complaints/edit/:id` - Edit a complaint
- `GET /api/complaints/:id` - Get a specific complaint
- `GET /api/complaints` - Get all complaints

- `POST /api/comments/create/:id` - Add a comment to a complaint

- `POST /api/categories/create` - Create a new category
- `GET /api/categories` - Get all categories

## Data Models

### Users
- ID
- Name
- Email
- Password
- CreatedAt

### Customers
- ID
- Name
- CreatedAt

### Complaints
- ID
- CustomerID (foreign key to Customers)
- Description
- CreatedAt
- ModifiedAt
- CreatedByID (foreign key to Users)
- Priority (High, Medium, Low)
- Status (New, UnderTreatment, Solved)
- CategoryId (foreign key to Categories)

### Comments
- ID
- Comment
- ComplaintID (foreign key to Complaints)
- CreatedAt
- CreatedByID (foreign key to Users)

### Categories
- ID
- Name
- CreatedAt
