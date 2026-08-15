package http

import (
	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

// MiningLicenseHandler handles HTTP requests for mining license applications.
type MiningLicenseHandler struct {
	Usecase     domain.MiningLicenseUsecase
	UserUsecase domain.UserUsecase
	JWTSecret   string
}

// NewMiningLicenseHandler registers all routes for the mining license API.
// All routes are protected — a valid Bearer JWT token is required.
func NewMiningLicenseHandler(app *fiber.App, uc domain.MiningLicenseUsecase, userUc domain.UserUsecase, jwtSecret string) {
	handler := &MiningLicenseHandler{Usecase: uc, UserUsecase: userUc, JWTSecret: jwtSecret}

	auth := middleware.Protected(jwtSecret)

	api := app.Group("/api/mining-licenses")
	api.Post("/", auth, handler.Submit)
	api.Get("/", auth, handler.GetAll)
	api.Get("/tin/:tin", auth, handler.GetByTIN)
	api.Get("/map", auth, handler.GetForMap) 
	api.Get("/:id", auth, handler.GetByID)
	api.Patch("/:id/status", auth, handler.UpdateStatus)
}

// Submit godoc
// POST /api/mining-licenses
// Creates a new mechanized gem mining license application.
// The createdBy field is resolved automatically from the JWT token.
func (h *MiningLicenseHandler) Submit(c *fiber.Ctx) error {
	var license domain.MechanizedGemMiningLicense

	if err := c.BodyParser(&license); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Resolve the submitting user's name from the JWT token
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Could not identify the requesting user",
		})
	}

	user, err := h.UserUsecase.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found: " + err.Error(),
		})
	}

	// Override whatever the client may have sent — always use the token owner's name
	license.CreatedBy = user.Name

	result, err := h.Usecase.Submit(c.Context(), &license)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "License application submitted successfully",
		"data":    result,
	})
}

// GetAll godoc
// GET /api/mining-licenses
// Returns all mining license applications.
func (h *MiningLicenseHandler) GetAll(c *fiber.Ctx) error {
	licenses, err := h.Usecase.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "License applications retrieved successfully",
		"data":    licenses,
	})
}

// GetByID godoc
// GET /api/mining-licenses/:id
// Returns a single mining license application by its MongoDB ObjectID.
func (h *MiningLicenseHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "License ID is required",
		})
	}

	license, err := h.Usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "License application retrieved successfully",
		"data":    license,
	})
}

// GetByTIN godoc
// GET /api/mining-licenses/tin/:tin
// Returns all mining license applications by the applicant's TIN number.
func (h *MiningLicenseHandler) GetByTIN(c *fiber.Ctx) error {
	tin := c.Params("tin")
	if tin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "TIN number is required",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	paginatedLicenses, err := h.Usecase.GetByTIN(c.Context(), tin, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "License applications retrieved successfully",
		"data":    paginatedLicenses,
	})
}

// GetForMap godoc
// GET /api/mining-licenses/map
// Returns license applications with GPS points, optionally filtered by
// district, village, tin, nic, gmlNumber, or landName query params.
func (h *MiningLicenseHandler) GetForMap(c *fiber.Ctx) error {
	filters := domain.MapFilters{
		District:  c.Query("district"),
		Village:   c.Query("village"),
		TIN:       c.Query("tin"),
		NIC:       c.Query("nic"),
		GMLNumber: c.Query("gmlNumber"),
		LandName:  c.Query("landName"),
	}

	licenses, err := h.Usecase.GetForMap(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Map license applications retrieved successfully",
		"data":    licenses,
	})
}

// UpdateStatus godoc
// PATCH /api/mining-licenses/:id/status
// Body: { "status": "submitted" | "approved" | "rejected" | "draft" }
// Updates only the status field of a license application.
func (h *MiningLicenseHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "License ID is required",
		})
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	if body.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "status field is required",
		})
	}

	if err := h.Usecase.UpdateStatus(c.Context(), id, body.Status); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "License status updated successfully",
	})
}
