package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/pedersandvoll/Practice-Exam-BE/config"
	"github.com/pedersandvoll/Practice-Exam-BE/tables"
	"github.com/pedersandvoll/Practice-Exam-BE/utils"
)

type Handlers struct {
	db        *config.Database
	JWTSecret []byte
}

func NewHandlers(db *config.Database, jwtSecret string) *Handlers {
	return &Handlers{
		db:        db,
		JWTSecret: []byte(jwtSecret),
	}
}

type UserBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (h *Handlers) RegisterUser(c *fiber.Ctx) error {
	var body UserBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Name == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	hashedPassword, err := utils.HashPassword(body.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user := tables.Users{
		Email:    body.Email,
		Name:     body.Name,
		Password: hashedPassword,
	}
	result := h.db.DB.Create(&user)

	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already exists",
			})
		}
		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"userid":  user.ID,
	})
}

type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handlers) LoginUser(c *fiber.Ctx) error {
	var body LoginBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	var user tables.Users
	result := h.db.DB.Where("email = ?", body.Email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve user",
			"msg":   result.Error.Error(),
		})
	}

	isValid := utils.VerifyPassword(body.Password, user.Password)
	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Email or password are wrong",
		})
	}

	claims := jwt.MapClaims{
		"username": user.Name,
		"userid":   user.ID,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString(h.JWTSecret)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

type CustomerBody struct {
	Name string `json:"name"`
}

func (h *Handlers) RegisterCustomer(c *fiber.Ctx) error {
	var body CustomerBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer name is required",
		})
	}

	customer := tables.Customers{
		Name: body.Name,
	}
	result := h.db.DB.Create(&customer)

	if result.Error != nil {
		fmt.Println("result error", result.Error.Error())
		if strings.Contains(result.Error.Error(), "duplicate key value") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Customer already exists",
			})
		}
		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create customer",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Customer created successfully",
		"customerid": customer.ID,
	})
}

func (h *Handlers) GetCustomers(c *fiber.Ctx) error {
	var customers []tables.Customers
	result := h.db.DB.Find(&customers)

	if result.Error != nil {
		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get customers",
			"msg":   result.Error.Error(),
		})
	}

	return c.JSON(customers)
}

type ComplaintsBody struct {
	CustomerName string          `json:"customername"`
	Description  string          `json:"description"`
	CategoryId   uint            `json:"category"`
	Priority     tables.Priority `json:"priority"`
}

func (h *Handlers) RegisterComplaint(c *fiber.Ctx) error {
	var body ComplaintsBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.CustomerName == "" || body.Description == "" || body.CategoryId == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Customer and description is required",
		})
	}

	token := c.Locals("user").(*jwt.Token)
	if token == nil {
		fmt.Println("JWT token not found in locals")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - Missing JWT token",
		})
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDFloat, ok := claims["userid"].(float64)
	if !ok {
		fmt.Println("Invalid User ID type in JWT.  Expected float64.")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid User ID type in JWT",
		})
	}
	userID := uint(userIDFloat)

	var customer tables.Customers
	resultCustomer := h.db.Where("name = ?", body.CustomerName).First(&customer)

	if resultCustomer.Error != nil {
		if errors.Is(resultCustomer.Error, gorm.ErrRecordNotFound) {
			newCustomer := tables.Customers{Name: body.CustomerName}
			resultNewCustomer := h.db.DB.Create(&newCustomer)

			if resultNewCustomer.Error != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create customer",
					"msg":   resultNewCustomer.Error.Error(),
				})
			}

			customer = newCustomer
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve user",
				"msg":   resultCustomer.Error.Error(),
			})
		}
	}

	complaint := tables.Complaints{
		CustomerID:  customer.ID,
		Description: body.Description,
		CreatedByID: userID,
		CategoryId:  body.CategoryId,
		Priority:    body.Priority,
	}
	result := h.db.DB.Create(&complaint)

	if result.Error != nil {
		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create complaint",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Complaint created successfully",
		"complaintid": complaint.ID,
	})
}

type EditComplaintBody struct {
	Description string          `json:"description"`
	CategoryId  uint            `json:"category"`
	Priority    tables.Priority `json:"priority"`
}

func (h *Handlers) EditComplaint(c *fiber.Ctx) error {
	complaintID := c.Params("id")
	if complaintID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required in the URL",
		})
	}
	var body EditComplaintBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Description == "" || body.Priority == -1 || body.CategoryId == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description, priority and category is required",
		})
	}

	if body.Priority != -1 {
		isValidPriority := body.Priority == tables.High ||
			body.Priority == tables.Medium ||
			body.Priority == tables.Low
		if !isValidPriority {
			fmt.Println("Invalid priority value:", body.Priority)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid priority value. Must be High, Medium, or Low.",
			})
		}
	}

	var complaint tables.Complaints
	result := h.db.First(&complaint, complaintID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Complaint not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load complaint",
		})
	}

	complaint.Description = body.Description
	complaint.Priority = body.Priority
	complaint.CategoryId = body.CategoryId

	result = h.db.Save(&complaint)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update complaint",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Complaint updated successfully",
		"complaintid": complaintID,
	})
}

func (h Handlers) GetComplaintById(c *fiber.Ctx) error {
	complaintID := c.Params("id")
	if complaintID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required in the URL",
		})
	}
	var complaint tables.Complaints
	result := h.db.
		Preload("CreatedBy").
		Preload("Customer").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("comments.created_at DESC") // Sort by CreatedAt in descending order
		}).
		Preload("Comments.CreatedBy").
		Preload("Category").
		First(&complaint, complaintID)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get complaint",
			"msg":   result.Error.Error(),
		})
	}

	return c.JSON(complaint)
}

func (h *Handlers) GetComplaints(c *fiber.Ctx) error {
	userId := c.Query("userId")
	customerId := c.Query("customerId")
	sortBy := c.Query("sortBy", "created_at")
	sortOrder := c.Query("sortOrder", "desc")

	query := h.db.
		Preload("CreatedBy").
		Preload("Customer").
		Preload("Comments").
		Preload("Comments.CreatedBy").
		Preload("Category")
	if userId != "" {
		query = query.Where("created_by_id = ?", userId)
	}
	if customerId != "" {
		query = query.Where("customer_id = ?", customerId)
	}

	allowedSortColumns := map[string]bool{
		"created_at":  true,
		"modified_at": true,
	}

	if _, exists := allowedSortColumns[sortBy]; !exists {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	query = query.Order(orderClause)

	var complaints []tables.Complaints
	result := query.Find(&complaints)

	if result.Error != nil {
		fmt.Println("Database error:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get complaints",
			"msg":   result.Error.Error(),
		})
	}

	return c.JSON(complaints)
}

type CommentBody struct {
	Comment string `json:"comment"`
}

func (h *Handlers) AddComplaintComment(c *fiber.Ctx) error {
	complaintIDStr := c.Params("id")
	if complaintIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID is required in the URL",
		})
	}

	complaintID, err := strconv.ParseUint(complaintIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid complaint ID format",
		})
	}

	var body CommentBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Comment == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Comment is required",
		})
	}

	token := c.Locals("user").(*jwt.Token)
	if token == nil {
		fmt.Println("JWT token not found in locals")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized - Missing JWT token",
		})
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDFloat, ok := claims["userid"].(float64)
	if !ok {
		fmt.Println("Invalid User ID type in JWT.  Expected float64.")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid User ID type in JWT",
		})
	}
	userID := uint(userIDFloat)

	comment := tables.Comments{
		ComplaintID: uint(complaintID),
		Comment:     body.Comment,
		CreatedByID: userID,
	}
	result := h.db.DB.Create(&comment)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create comment",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "comment created successfully",
		"commentid": comment.ID,
	})
}

type CategoryBody struct {
	Name string `json:"name"`
}

func (h *Handlers) RegisterCategory(c *fiber.Ctx) error {
	var body CategoryBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category name is required",
		})
	}

	category := tables.Categories{
		Name: body.Name,
	}
	result := h.db.DB.Create(&category)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create category",
			"msg":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Category created successfully",
		"customerid": category.ID,
	})
}

func (h *Handlers) GetCategories(c *fiber.Ctx) error {
	var categories []tables.Categories
	result := h.db.DB.Find(&categories)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get categories",
			"msg":   result.Error.Error(),
		})
	}

	return c.JSON(categories)
}
