package http

import (
	"errors"
	"strconv"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type ApplicationPaymentHandler struct {
	Usecase domain.ApplicationPaymentUsecase
}

func NewApplicationPaymentHandler(app *fiber.App, us domain.ApplicationPaymentUsecase, jwtSecret string) {
	handler := &ApplicationPaymentHandler{Usecase: us}

	api := app.Group("/api/application-payments", middleware.Protected(jwtSecret))
	api.Get("/:applicationId", handler.GetByApplication)
}

func (h *ApplicationPaymentHandler) GetByApplication(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	result, err := h.Usecase.GetByApplication(c.Context(), c.Params("applicationId"), page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrApplicationIDRequired) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}