package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/pedersandvoll/Practice-Exam-BE/config"
	"github.com/pedersandvoll/Practice-Exam-BE/handlers"
	"github.com/pedersandvoll/Practice-Exam-BE/routes"
	"github.com/pedersandvoll/Practice-Exam-BE/tables"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbConfig := config.NewConfig()
	db, err := config.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	tables.RunMigrations(db.DB)

	app := fiber.New()

	h := handlers.NewHandlers(db, dbConfig.JWTSecret)

	routes.Routes(app, h)

	app.Listen(":3000")
}
