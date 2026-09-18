package http

import (
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
	api.Post("/", handler.Create)
	api.Get("/:id", handler.GetByID)
	api.Put("/:id", handler.Update)
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

	if err := c.BodyParser(&form); err != nil {
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

	userID := c.Locals("user_id").(string)
	err := h.Usecase.Create(c.Context(), &form, userID)
	if err != nil {
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

func (h *MiniSahanaHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.Usecase.Delete(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Form deleted successfully",
	})
}
