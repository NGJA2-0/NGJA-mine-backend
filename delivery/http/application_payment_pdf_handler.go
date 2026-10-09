package http

import (
	"errors"
	"fmt"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type ApplicationPaymentPDFHandler struct {
	Usecase domain.ApplicationPaymentPDFUsecase
}

func NewApplicationPaymentPDFHandler(app *fiber.App, us domain.ApplicationPaymentPDFUsecase, jwtSecret string) {
	handler := &ApplicationPaymentPDFHandler{Usecase: us}
	app.Get("/api/application-payments/:applicationId/pdf", middleware.Protected(jwtSecret), handler.DownloadPDF)
}

func (h *ApplicationPaymentPDFHandler) DownloadPDF(c *fiber.Ctx) error {
	applicationID := c.Params("applicationId")

	pdfBytes, err := h.Usecase.GeneratePDF(c.Context(), applicationID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrApplicationIDRequired):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrNoPaymentRecords):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="application-payments-%s.pdf"`, applicationID))
	return c.Send(pdfBytes)
}