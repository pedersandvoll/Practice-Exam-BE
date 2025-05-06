package handlers

import (
	"errors"
	"fmt"
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
			"error": "Username and password are required",
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
			"error": "User or password are wrong",
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
	CustomerID  uint   `json:"customerid"`
	Description string `json:"description"`
}
