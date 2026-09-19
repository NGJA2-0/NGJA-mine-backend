package http

import (
	"fmt"
	"path/filepath"
	"strconv"
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
	api.Get("/search", handler.Search)
	api.Post("/", handler.Create)
}

func (h *ReportCardHandler) Create(c *fiber.Ctx) error {
	// Parse multipart/form-data fields
	card := domain.ReportCard{}

	card.ApplicationID = strings.TrimSpace(c.FormValue("applicationId"))
	card.FullName = strings.TrimSpace(c.FormValue("fullName"))
	card.AccNumber = strings.TrimSpace(c.FormValue("accNumber"))
	card.NIC = strings.TrimSpace(c.FormValue("nic"))
	card.Grade = strings.TrimSpace(c.FormValue("grade"))
	card.StartDate = strings.TrimSpace(c.FormValue("startDate"))
	card.EndDate = strings.TrimSpace(c.FormValue("endDate"))

	amountStr := strings.TrimSpace(c.FormValue("amount"))
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "amount must be a valid number greater than zero"})
	}
	card.Amount = amount

	// Validate required text fields
	errorsList := []string{}
	checkEmpty := func(field string, name string) {
		if field == "" {
			errorsList = append(errorsList, name+" is required")
		}
	}

	checkEmpty(card.ApplicationID, "applicationId")
	checkEmpty(card.FullName, "fullName")
	checkEmpty(card.AccNumber, "accNumber")
	checkEmpty(card.NIC, "nic")
	checkEmpty(card.Grade, "grade")
	checkEmpty(card.StartDate, "startDate")
	checkEmpty(card.EndDate, "endDate")

	if len(errorsList) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Validation failed",
			"fields": errorsList,
		})
	}

	// Handle PDF file upload
	file, err := c.FormFile("pdf")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "pdf file is required"})
	}

	// Only allow PDF files
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "only PDF files are allowed"})
	}

	// Build the save path — usecase will generate the refNumber, so we use a temp name here.
	// We pass the file through to the usecase after the refNumber is known.
	// Step 1: Run business logic (calculates refNumber, totalDuration, totalAmount)
	if err := h.Usecase.Create(c.Context(), &card); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Step 2: Save the PDF using the generated refNumber as the filename
	savePath := fmt.Sprintf("./report_cards/%s.pdf", card.RefNumber)
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save PDF file"})
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

func (h *ReportCardHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	results, err := h.Usecase.Search(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(results)
}

