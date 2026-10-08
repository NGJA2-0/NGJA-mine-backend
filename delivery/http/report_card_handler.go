package http

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"mime/multipart" 
	"os"
	"regexp"
	"time"

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
	api.Get("/by-month", handler.ListByMonth)
	api.Get("/monthly-report", handler.GetMonthlyReport)
	api.Get("/monthly-report/all", handler.GetMonthlyReportAll)
	api.Patch("/:id/pay", handler.PayMonths)
	api.Get("/by-application/:applicationId", handler.GetByApplicationID)
	api.Get("/", handler.ListSubmissions)
	api.Post("/", handler.Create)
	api.Get("/ol-certificate/:applicationId", handler.GetOLCertificateStatus)
	api.Get("/ol-certificate/:applicationId/file", handler.GetOLCertificateFile)
}

func (h *ReportCardHandler) Create(c *fiber.Ctx) error {
	// Parse multipart/form-data fields
	card := domain.ReportCard{}

	card.ApplicationID = strings.TrimSpace(c.FormValue("applicationId"))
	card.FullName = strings.TrimSpace(c.FormValue("fullName"))
	card.AccNumber = strings.TrimSpace(c.FormValue("accNumber"))
	card.NIC = strings.TrimSpace(c.FormValue("nic"))
	card.AppliedGrade = strings.TrimSpace(c.FormValue("appliedGrade"))
	card.CurrentGrade = strings.TrimSpace(c.FormValue("currentGrade"))
	if !regexp.MustCompile(`^\d{1,2}$`).MatchString(card.CurrentGrade) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "currentGrade must be a number"})
	}
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
	checkEmpty(card.AppliedGrade, "appliedGrade")
	checkEmpty(card.CurrentGrade, "currentGrade")
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

	// Grade 12 requires an O/L certificate (existing one on the application, or a new upload)
	if card.CurrentGrade == "12" {
		existing, err := h.Usecase.GetOLCertificate(c.Context(), card.ApplicationID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		if existing == nil {
			olFile, err := c.FormFile("olCertificate")
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "O/L certificate is required for grade 12"})
			}
			if strings.ToLower(filepath.Ext(olFile.Filename)) != ".pdf" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "O/L certificate must be a PDF"})
			}
			if err := h.saveOLCertificate(c, card.ApplicationID, olFile); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
		}
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

func (h *ReportCardHandler) GetByApplicationID(c *fiber.Ctx) error {
	applicationID := c.Params("applicationId")
	if applicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "applicationId is required"})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "5"))
	if err != nil {
		limit = 5
	}

	result, err := h.Usecase.GetByApplicationID(c.Context(), applicationID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
func (h *ReportCardHandler) ListSubmissions(c *fiber.Ctx) error {
	grade := c.Query("grade")

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	result, err := h.Usecase.ListSubmissions(c.Context(), grade, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *ReportCardHandler) saveOLCertificate(c *fiber.Ctx, applicationID string, file *multipart.FileHeader) error {
	dir := filepath.Join("minisahana_documents", applicationID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create folder")
	}
	storedPath := filepath.ToSlash(filepath.Join(dir, "ol_certificate_v1.pdf"))
	if err := c.SaveFile(file, storedPath); err != nil {
		return fmt.Errorf("failed to save O/L certificate")
	}

	uploader, _ := c.Locals("username").(string) // adjust to however your middleware stores the user
	return h.Usecase.AddOLCertificate(c.Context(), applicationID, domain.DocumentVersion{
		Version:    1,
		FileName:   file.Filename,
		StoredPath: storedPath,
		Size:       file.Size,
		UploadedBy: uploader,
		UploadedAt: time.Now(),
	})
}

func (h *ReportCardHandler) GetOLCertificateStatus(c *fiber.Ctx) error {
	doc, err := h.Usecase.GetOLCertificate(c.Context(), c.Params("applicationId"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if doc == nil {
		return c.JSON(fiber.Map{"exists": false})
	}
	return c.JSON(fiber.Map{
		"exists":     true,
		"fileName":   doc.FileName,
		"uploadedBy": doc.UploadedBy,
		"uploadedAt": doc.UploadedAt,
	})
}

func (h *ReportCardHandler) GetOLCertificateFile(c *fiber.Ctx) error {
	doc, err := h.Usecase.GetOLCertificate(c.Context(), c.Params("applicationId"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if doc == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "O/L certificate not found"})
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline")
	return c.SendFile(doc.StoredPath)
}

func (h *ReportCardHandler) ListByMonth(c *fiber.Ctx) error {
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "year is required and must be a number"})
	}

	fromStr, toStr := c.Query("fromMonth"), c.Query("toMonth")
	if fromStr == "" && toStr == "" {
		// single-month shortcut: ?month=6 is the same as fromMonth=6&toMonth=6
		fromStr = c.Query("month")
		toStr = fromStr
	}
	fromMonth, err := strconv.Atoi(fromStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "fromMonth is required and must be a number (1-12)"})
	}
	toMonth, err := strconv.Atoi(toStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "toMonth is required and must be a number (1-12)"})
	}

	grade := strings.TrimSpace(c.Query("grade")) // optional, matched against current_grade

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	result, err := h.Usecase.ListByMonth(c.Context(), year, fromMonth, toMonth, grade, page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriod) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

func (h *ReportCardHandler) PayMonths(c *fiber.Ctx) error {
	var body struct {
		Year      int `json:"year"`
		FromMonth int `json:"fromMonth"`
		ToMonth   int `json:"toMonth"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Set by middleware.Protected from the JWT
	userID, _ := c.Locals("user_id").(string)

	result, err := h.Usecase.PayMonths(c.Context(), c.Params("id"), body.Year, body.FromMonth, body.ToMonth, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidReportCardID), errors.Is(err, domain.ErrInvalidPeriod):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrMonthNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrMonthAlreadyPaid):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, domain.ErrUserNotFound):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Months marked as paid",
		"result":  result,
	})
}

func parseYearMonth(c *fiber.Ctx) (int, int, error) {
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return 0, 0, errors.New("year is required and must be a number")
	}
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil {
		return 0, 0, errors.New("month is required and must be a number (1-12)")
	}
	return year, month, nil
}

// GetMonthlyReport: paginated list + full totals for one year and month
func (h *ReportCardHandler) GetMonthlyReport(c *fiber.Ctx) error {
	year, month, err := parseYearMonth(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil {
		limit = 10
	}

	result, err := h.Usecase.GetMonthlyReport(c.Context(), year, month, page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriod) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

// GetMonthlyReportAll: every row + full totals, for the report download button
func (h *ReportCardHandler) GetMonthlyReportAll(c *fiber.Ctx) error {
	year, month, err := parseYearMonth(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	result, err := h.Usecase.GetMonthlyReportAll(c.Context(), year, month)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPeriod) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(result)
}