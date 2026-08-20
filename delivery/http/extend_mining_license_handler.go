package http

import (
	"my-fiber-app/domain"
	"my-fiber-app/middleware"

	"github.com/gofiber/fiber/v2"
)

// ExtendMiningLicenseHandler handles HTTP requests for extend mining license applications.
type ExtendMiningLicenseHandler struct {
	Usecase     domain.ExtendMiningLicenseUsecase
	UserUsecase domain.UserUsecase
	JWTSecret   string
}

// NewExtendMiningLicenseHandler registers all routes for the extend mining license API.
// All routes are protected — a valid Bearer JWT token is required.
func NewExtendMiningLicenseHandler(app *fiber.App, uc domain.ExtendMiningLicenseUsecase, userUc domain.UserUsecase, jwtSecret string) {
	handler := &ExtendMiningLicenseHandler{Usecase: uc, UserUsecase: userUc, JWTSecret: jwtSecret}

	auth := middleware.Protected(jwtSecret)

	api := app.Group("/api/extend-mining-licenses")
	api.Post("/", auth, handler.Submit)
	api.Get("/", auth, handler.GetAll)
	api.Get("/tin/:tin", auth, handler.GetByTIN)
	api.Get("/map", auth, handler.GetForMap) 
	api.Get("/:id", auth, handler.GetByID)
	api.Post("/:id/edit", auth, handler.Edit)
	api.Patch("/:id/status", auth, handler.UpdateStatus)
}

// Submit godoc
// POST /api/extend-mining-licenses
// Creates a new mechanized gem mining license extension application.
// The createdBy field is resolved automatically from the JWT token.
func (h *ExtendMiningLicenseHandler) Submit(c *fiber.Ctx) error {
	var license domain.ExtendMiningLicense

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
		"message": "Extend license application submitted successfully",
		"data":    result,
	})
}

// GetAll godoc
// GET /api/extend-mining-licenses
// Returns all extend mining license applications.
func (h *ExtendMiningLicenseHandler) GetAll(c *fiber.Ctx) error {
	licenses, err := h.Usecase.GetAll(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Extend license applications retrieved successfully",
		"data":    licenses,
	})
}

// GetByID godoc
// GET /api/extend-mining-licenses/:id
// Returns a single extend mining license application by its MongoDB ObjectID.
func (h *ExtendMiningLicenseHandler) GetByID(c *fiber.Ctx) error {
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
		"message": "Extend license application retrieved successfully",
		"data":    license,
	})
}

// GetByTIN godoc
// GET /api/extend-mining-licenses/tin/:tin
// Returns all extend mining license applications by the applicant's TIN number.
func (h *ExtendMiningLicenseHandler) GetByTIN(c *fiber.Ctx) error {
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
		"message": "Extend license applications retrieved successfully",
		"data":    paginatedLicenses,
	})
}

// GetForMap godoc
// GET /api/extend-mining-licenses/map
// Returns extend license applications with GPS points, optionally filtered by
// district, village, tin, nic, gmlNumber, or landName query params.
func (h *ExtendMiningLicenseHandler) GetForMap(c *fiber.Ctx) error {
	filters := domain.MapFilters{
		District:       c.Query("district"),
		Village:        c.Query("village"),
		RegionalOffice: c.Query("regionalOffice"),
		TIN:            c.Query("tin"),
		NIC:            c.Query("nic"),
		GMLNumber:      c.Query("gmlNumber"),
		LandName:       c.Query("landName"),
		Search:         c.Query("q"),
	}

	licenses, err := h.Usecase.GetForMap(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Map extend license applications retrieved successfully",
		"data":    licenses,
	})
}

// UpdateStatus godoc
// PATCH /api/extend-mining-licenses/:id/status
// Body: { "status": "submitted" | "approved" | "rejected" | "draft" }
// Updates only the status field of an extend license application.
func (h *ExtendMiningLicenseHandler) UpdateStatus(c *fiber.Ctx) error {
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
		"message": "Extend license status updated successfully",
	})
}

// Edit godoc
// POST /api/extend-mining-licenses/:id/edit
// Creates a new versioned edit of an existing extend mining license application.
// The createdBy field is resolved automatically from the JWT token.
func (h *ExtendMiningLicenseHandler) Edit(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "License ID is required",
		})
	}

	var updatedLicense domain.ExtendMiningLicense
	if err := c.BodyParser(&updatedLicense); err != nil {
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

	// Always use the token owner's name
	updatedLicense.CreatedBy = user.Name

	result, err := h.Usecase.Edit(c.Context(), id, &updatedLicense)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Extend license application edited successfully",
		"data":    result,
	})
}
