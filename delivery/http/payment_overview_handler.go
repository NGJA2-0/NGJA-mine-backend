package http

import (
	"strconv"
	"time"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type PaymentOverviewHandler struct {
	Usecase domain.PaymentOverviewUsecase
}

func NewPaymentOverviewHandler(app *fiber.App, us domain.PaymentOverviewUsecase, jwtSecret string) {
	h := &PaymentOverviewHandler{Usecase: us}
	api := app.Group("/api/payments", middleware.Protected(jwtSecret))
	api.Get("/overview", h.GetOverview)
}

func (h *PaymentOverviewHandler) GetOverview(c *fiber.Ctx) error {
	year, err := strconv.Atoi(c.Query("year", strconv.Itoa(time.Now().Year())))
	if err != nil {
		year = time.Now().Year()
	}
	result, err := h.Usecase.GetOverview(c.Context(), year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}