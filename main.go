package main

import (
	"log"

	"my-fiber-app/config"
	httpHandler "my-fiber-app/delivery/http"
	"my-fiber-app/middleware"
	"my-fiber-app/repository"
	"my-fiber-app/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load environment variables
	env := config.LoadEnv()

	// Connect to MongoDB
	db := config.ConnectDB(env)

	// Initialize Fiber app
	app := fiber.New()

	// Global Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Modify this for production
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Setup Repositories
	userRepo := repository.NewUserRepository(db)

	// Setup Usecases
	userUsecase := usecase.NewUserUsecase(userRepo, env.JWTSecret)

	// Setup Handlers (Routes)
	httpHandler.NewUserHandler(app, userUsecase)

	// Root GET Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Go Fiber!")
	})

	// JSON Response Route
	app.Get("/api/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "active",
			"message": "Fiber server is running smoothly",
		})
	})

	// Example of a protected route using JWT middleware
	api := app.Group("/api")
	api.Get("/protected", middleware.Protected(env.JWTSecret), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{
			"message": "This is a protected route",
			"user_id": userID,
		})
	})

	// Start server
	log.Printf("Server starting on port %s", env.Port)
	log.Fatal(app.Listen(":" + env.Port))
}