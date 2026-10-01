package http

import (
	"encoding/json" 
	"errors" 
	"fmt" 
	"io" 
	"mime/multipart" 
	"net/url" 
	"os" 
	"path/filepath"
	"strconv"
	"strings"

	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

type MiniSahanaHandler struct {
	Usecase domain.MiniSahanaUsecase
}

func NewMiniSahanaHandler(app *fiber.App, us domain.MiniSahanaUsecase, jwtSecret string) {
	handler := &MiniSahanaHandler{
		Usecase: us,
	}

	api := app.Group("/api/mini-sahana-form", middleware.Protected(jwtSecret))
	api.Get("/search", handler.Search)
	api.Get("/stats", handler.Stats)
	api.Post("/", handler.Create)
	api.Get("/:id/documents/:key/versions/:version", handler.GetDocument)
	api.Put("/:id/documents/:key", handler.UpdateDocument)
	api.Get("/:id", handler.GetByID)
	api.Put("/:id", handler.Update)
	api.Put("/:id/account-number", handler.UpdateAccountNumber)
	api.Delete("/:id", handler.Delete)
}

func (h *MiniSahanaHandler) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	results, err := h.Usecase.Search(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(results)
}

func (h *MiniSahanaHandler) Create(c *fiber.Ctx) error {
	var form domain.MiniSahanaForm
	if err := json.Unmarshal([]byte(c.FormValue("data")), &form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate all fields manually for empty values
	errorsList := []string{}
	
	// Helper to add error
	checkEmpty := func(field string, name string) {
		if strings.TrimSpace(field) == "" {
			errorsList = append(errorsList, name+" is required")
		}
	}

	checkEmpty(form.FormType, "formType")
	checkEmpty(form.Year, "year")
	checkEmpty(form.ApplicantFullNameSinhala, "applicantFullNameSinhala")
	checkEmpty(form.ApplicantFullNameEnglish, "applicantFullNameEnglish")
	checkEmpty(form.NameWithInitialsEnglish, "nameWithInitialsEnglish")
	checkEmpty(form.DateOfBirth, "dateOfBirth")
	checkEmpty(form.Gender, "gender")
	checkEmpty(form.School, "school")
	checkEmpty(form.Grade, "grade")
	checkEmpty(form.SchoolAddressAndPhone, "schoolAddressAndPhone")
	checkEmpty(form.BankAccountName, "bankAccountName")
	checkEmpty(form.BankBranch, "bankBranch")
	checkEmpty(form.BankAccountNumber, "bankAccountNumber")
	checkEmpty(form.ParentGuardianRelation, "parentGuardianRelation")
	checkEmpty(form.ParentGuardianName, "parentGuardianName")
	checkEmpty(form.PermanentAddress, "permanentAddress")
	checkEmpty(form.PhoneNumber, "phoneNumber")
	checkEmpty(form.NIC, "nic")
	checkEmpty(form.Occupation, "occupation")
	checkEmpty(form.MaritalStatus, "maritalStatus")
	checkEmpty(form.SpouseRelation, "spouseRelation")
	checkEmpty(form.SpouseName, "spouseName")
	checkEmpty(form.SpouseOccupation, "spouseOccupation")
	checkEmpty(form.LicenseNumber, "licenseNumber")
	checkEmpty(form.FileNumber, "fileNumber")
	checkEmpty(form.LicenseRegionalOfficeAndZone, "licenseRegionalOfficeAndZone")
	checkEmpty(form.Attachments.BankPassbookCopy, "attachments.bankPassbookCopy")
	checkEmpty(form.Attachments.BirthCertificateCopy, "attachments.birthCertificateCopy")

	if len(form.EligibilityCategory) == 0 {
		errorsList = append(errorsList, "eligibilityCategory is required")
	}
	if form.Age == nil {
		errorsList = append(errorsList, "age is required")
	}
	if form.MonthlyIncome == nil {
		errorsList = append(errorsList, "monthlyIncome is required")
	}
	if form.ChildrenCount == nil {
		errorsList = append(errorsList, "childrenCount is required")
	}
	if form.DeclarationSigned == nil {
		errorsList = append(errorsList, "declarationSigned is required")
	}

	if len(errorsList) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Validation failed",
			"fields": errorsList,
		})
	}

	files := map[string]*domain.DocUpload{}
	for _, slot := range domain.MiniSahanaDocumentSlots {
		fh, err := c.FormFile(slot.FormField)
		if err != nil {
			continue
		}
		if err := validatePDF(fh); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": slot.Label + ": " + err.Error()})
		}
		files[slot.Key] = &domain.DocUpload{FileName: fh.Filename, Size: fh.Size, Save: savePDF(c, fh)}
	}


	userID := c.Locals("user_id").(string)
	if err := h.Usecase.Create(c.Context(), &form, userID, files); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Form submitted successfully",
		"id":        form.ID.Hex(),
		"refNumber": form.RefNumber,
	})
}

func (h *MiniSahanaHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	form, err := h.Usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(form)
}

func (h *MiniSahanaHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var form domain.MiniSahanaForm

	if err := c.BodyParser(&form); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(string)
	err := h.Usecase.Update(c.Context(), id, &form, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Form updated successfully",
		"id":        form.ID.Hex(),
		"refNumber": form.RefNumber,
	})
}

func (h *MiniSahanaHandler) UpdateAccountNumber(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		BankAccountNumber string `json:"bankAccountNumber"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID := c.Locals("user_id").(string)
	err := h.Usecase.UpdateAccountNumber(c.Context(), id, body.BankAccountNumber, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Account number updated successfully"})
}

func (h *MiniSahanaHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.Usecase.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// remove this application's PDF folder too
	_ = os.RemoveAll(filepath.FromSlash(domain.DocumentDir(id)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Form deleted successfully",
	})
}

const maxPDFSize = 10 << 20

func validatePDF(fh *multipart.FileHeader) error {
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".pdf" {
		return errors.New("only PDF files are allowed")
	}
	if fh.Size > maxPDFSize {
		return errors.New("file exceeds 10MB")
	}
	f, err := fh.Open()
	if err != nil {
		return errors.New("could not read file")
	}
	defer f.Close()
	head := make([]byte, 5)
	if _, err := io.ReadFull(f, head); err != nil || string(head) != "%PDF-" {
		return errors.New("file is not a valid PDF")
	}
	return nil
}

func savePDF(c *fiber.Ctx, fh *multipart.FileHeader) func(rel string) error {
	return func(rel string) error {
		full := filepath.Join(".", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		return c.SaveFile(fh, full)
	}
}

func (h *MiniSahanaHandler) GetDocument(c *fiber.Ctx) error {
	version, err := strconv.Atoi(c.Params("version"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid version"})
	}
	meta, err := h.Usecase.OpenDocument(c.Context(), c.Params("id"), c.Params("key"), version)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	rel := filepath.Clean(filepath.FromSlash(meta.StoredPath))
	if !strings.HasPrefix(rel, domain.DocumentsRoot+string(filepath.Separator)) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	disp := "inline"
	if c.Query("download") == "1" {
		disp = "attachment"
	}
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf("%s; filename*=UTF-8''%s", disp, url.PathEscape(meta.FileName)))
	return c.SendFile(rel)
}

func (h *MiniSahanaHandler) UpdateDocument(c *fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}
	if err := validatePDF(fh); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	userID := c.Locals("user_id").(string)
	updated, err := h.Usecase.UpdateDocument(c.Context(), c.Params("id"), c.Params("key"),
		&domain.DocUpload{FileName: fh.Filename, Size: fh.Size, Save: savePDF(c, fh)}, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(updated)
}

func (h *MiniSahanaHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.Usecase.GetStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(stats)
}