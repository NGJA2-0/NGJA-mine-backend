package http

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type AnnualReportHandler struct {
	Usecase domain.AnnualReportUsecase
}

func NewAnnualReportHandler(app *fiber.App, us domain.AnnualReportUsecase, jwtSecret string) {
	handler := &AnnualReportHandler{Usecase: us}

	api := app.Group("/api/annual-reports", middleware.Protected(jwtSecret))
	api.Get("/", handler.GetAnnualReport)
	api.Get("/all", handler.GetAnnualReportAll)
	api.Get("/pdf", handler.DownloadPDF)
}

func parseYearParam(c *fiber.Ctx) (int, error) {
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return 0, domain.ErrInvalidYear
	}
	return year, nil
}

func annualReportError(c *fiber.Ctx, err error) error {
	if errors.Is(err, domain.ErrInvalidYear) || errors.Is(err, domain.ErrInvalidGrade) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}

// GetAnnualReport: paginated list + full totals (grand total, month-wise, grade-wise) for one year
func (h *AnnualReportHandler) GetAnnualReport(c *fiber.Ctx) error {
	year, err := parseYearParam(c)
	if err != nil {
		return annualReportError(c, err)
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	result, err := h.Usecase.GetAnnualReport(c.Context(), year, c.Query("grade"), page, limit)
	if err != nil {
		return annualReportError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

// GetAnnualReportAll: every row + the same totals, as JSON
func (h *AnnualReportHandler) GetAnnualReportAll(c *fiber.Ctx) error {
	year, err := parseYearParam(c)
	if err != nil {
		return annualReportError(c, err)
	}

	result, err := h.Usecase.GetAnnualReportAll(c.Context(), year, c.Query("grade"))
	if err != nil {
		return annualReportError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

// DownloadPDF: the whole report as a ready-made PDF file
func (h *AnnualReportHandler) DownloadPDF(c *fiber.Ctx) error {
	year, err := parseYearParam(c)
	if err != nil {
		return annualReportError(c, err)
	}
	grade := strings.TrimSpace(c.Query("grade"))

	pdfBytes, err := h.Usecase.GenerateAnnualReportPDF(c.Context(), year, grade)
	if err != nil {
		return annualReportError(c, err)
	}

	// grade was validated as 1-2 digits by the usecase, so it is safe in the file name
	fileName := fmt.Sprintf("annual-report-%d.pdf", year)
	if grade != "" {
		fileName = fmt.Sprintf("annual-report-%d-grade-%s.pdf", year, grade)
	}
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, fileName))
	return c.Send(pdfBytes)
}