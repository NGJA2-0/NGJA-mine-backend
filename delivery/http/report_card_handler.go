package http

import (
	"strings"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type ReportCardHandler struct {
	Usecase domain.ReportCardUsecase
}

func NewReportCardHandler(app *fiber.App, us domain.ReportCardUsecase, jwtSecret string) {
	handler := &ReportCardHandler{
		Usecase: us,
	}

	api := app.Group("/api/report-cards", middleware.Protected(jwtSecret))
	api.Post("/", handler.Create)
}

func (h *ReportCardHandler) Create(c *fiber.Ctx) error {
	var card domain.ReportCard

	if err := c.BodyParser(&card); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Basic presence validation before passing to usecase
	errorsList := []string{}
	checkEmpty := func(field string, name string) {
		if strings.TrimSpace(field) == "" {
			errorsList = append(errorsList, name+" is required")
		}
	}

	checkEmpty(card.ApplicationID, "applicationId")
	checkEmpty(card.FullName, "fullName")
	checkEmpty(card.AccNumber, "accNumber")
	checkEmpty(card.NIC, "nic")
	checkEmpty(card.StartDate, "startDate")
	checkEmpty(card.EndDate, "endDate")

	if card.Amount <= 0 {
		errorsList = append(errorsList, "amount must be greater than zero")
	}

	if len(errorsList) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Validation failed",
			"fields": errorsList,
		})
	}

	err := h.Usecase.Create(c.Context(), &card)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "Report card saved successfully",
		"id":            card.ID.Hex(),
		"refNumber":     card.RefNumber,
		"pdfUrl":        card.PdfUrl,
		"totalDuration": card.TotalDuration,
		"totalAmount":   card.TotalAmount,
	})
}
