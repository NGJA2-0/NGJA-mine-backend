package http

import (
	"my-fiber-app/domain"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
}

// NewUserHandler sets up routing for user endpoints
func NewUserHandler(app *fiber.App, us domain.UserUsecase) {
	handler := &UserHandler{
		UserUsecase: us,
	}

	api := app.Group("/api")
	api.Post("/signup", handler.Signup)
	api.Post("/login", handler.Login)
}

// Signup handles the user registration
func (h *UserHandler) Signup(c *fiber.Ctx) error {
	var req domain.UserSignupRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.NIC == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, NIC, and Password are required"})
	}

	user, token, err := h.UserUsecase.Signup(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    user,
		"token":   token,
	})
}

// Login handles user authentication
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req domain.UserLoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.NIC == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "NIC and Password are required"})
	}

	user, token, err := h.UserUsecase.Login(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}
