package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GPSPoint represents a single GPS coordinate pair
type GPSPoint struct {
	Latitude  string `json:"latitude" bson:"latitude"`
	Longitude string `json:"longitude" bson:"longitude"`
}

// MapFilters holds optional query filters for the map view.
type MapFilters struct {
	District  string
	Village   string
	RegionalOffice string
	TIN       string
	NIC       string
	GMLNumber string
	LandName  string
	Search        string
}

// MechanizedGemMiningLicense represents the full license application document stored in MongoDB
type MechanizedGemMiningLicense struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	ReferenceNumber string       `json:"referenceNumber" bson:"referenceNumber"`

	// Applicant details
	ApplicantName    string `json:"applicantName" bson:"applicantName"`
	ApplicantAddress string `json:"applicantAddress" bson:"applicantAddress"`
	ApplicantPhone   string `json:"applicantPhone" bson:"applicantPhone"`
	NIC              string `json:"nic" bson:"nic"`
	TIN              string `json:"tin,omitempty" bson:"tin,omitempty"`

	// Expense party (optional)
	HasExpenseParty bool   `json:"hasExpenseParty" bson:"hasExpenseParty"`
	ExpenseName     string `json:"expenseName,omitempty" bson:"expenseName,omitempty"`
	ExpenseAddress  string `json:"expenseAddress,omitempty" bson:"expenseAddress,omitempty"`
	ExpensePhone    string `json:"expensePhone,omitempty" bson:"expensePhone,omitempty"`
	ExpenseTIN      string `json:"expenseTin,omitempty" bson:"expenseTin,omitempty"`

	// License core
	GMLNumber string     `json:"gmlNumber" bson:"gmlNumber"`
	GPSPoints []GPSPoint `json:"gpsPoints" bson:"gpsPoints"`

	// Land details
	LandName   string `json:"landName" bson:"landName"`
	LandNature string `json:"landNature" bson:"landNature"`

	// Ratnapura land evidence
	IsRatnapuraLand             string `json:"isRatnapuraLand" bson:"isRatnapuraLand"` // "yes" | "no"
	WrittenEvidenceAttachmentUrl string `json:"writtenEvidenceAttachmentUrl,omitempty" bson:"writtenEvidenceAttachmentUrl,omitempty"`
	AffidavitAttachmentUrl       string `json:"affidavitAttachmentUrl,omitempty" bson:"affidavitAttachmentUrl,omitempty"`

	// Location
	District       string `json:"district" bson:"district"`
	Village        string `json:"village,omitempty" bson:"village,omitempty"`
	RegionalOffice string `json:"regionalOffice" bson:"regionalOffice"`

	// Land info
	LandExtent            string `json:"landExtent,omitempty" bson:"landExtent,omitempty"`
	LicenseeType          string `json:"licenseeType" bson:"licenseeType"`
	ConsentLetterAttached string `json:"consentLetterAttached,omitempty" bson:"consentLetterAttached,omitempty"` // "yes" | "no"
	GovLandAct            string `json:"govLandAct,omitempty" bson:"govLandAct,omitempty"`                       // "yes" | "no"

	// Existing pits history (conditional on existingPits == "yes")
	ExistingPits          string `json:"existingPits" bson:"existingPits"` // "yes" | "no"
	PrevLicenseFirstDate  string `json:"prevLicenseFirstDate,omitempty" bson:"prevLicenseFirstDate,omitempty"`
	ExtensionCount        *int   `json:"extensionCount,omitempty" bson:"extensionCount,omitempty"`
	MinedGemValue         string `json:"minedGemValue,omitempty" bson:"minedGemValue,omitempty"`
	ConditionBreach       string `json:"conditionBreach,omitempty" bson:"conditionBreach,omitempty"` // "yes" | "no"
	ConditionBreachDetails string `json:"conditionBreachDetails,omitempty" bson:"conditionBreachDetails,omitempty"`
	OwnershipComplaint    string `json:"ownershipComplaint,omitempty" bson:"ownershipComplaint,omitempty"` // "yes" | "no"
	ComplaintDetails      string `json:"complaintDetails,omitempty" bson:"complaintDetails,omitempty"`
	ExistingPitsArea       string `json:"existingPitsArea,omitempty" bson:"existingPitsArea,omitempty"`
	MudPitsArea            string `json:"mudPitsArea,omitempty" bson:"mudPitsArea,omitempty"`
	DepthSize              string `json:"depthSize,omitempty" bson:"depthSize,omitempty"`
	BreachesInLast3Months  string `json:"breachesInLast3Months,omitempty" bson:"breachesInLast3Months,omitempty"`
	ReportsSubmitted       string `json:"reportsSubmitted,omitempty" bson:"reportsSubmitted,omitempty"` // "yes" | "no"
	PrivateSaleValue       string `json:"privateSaleValue,omitempty" bson:"privateSaleValue,omitempty"`
	AuctionSaleValue       string `json:"auctionSaleValue,omitempty" bson:"auctionSaleValue,omitempty"`

	// Mining proposal
	ProposedDepth   *float64 `json:"proposedDepth,omitempty" bson:"proposedDepth,omitempty"`
	LandCultivation string   `json:"landCultivation,omitempty" bson:"landCultivation,omitempty"`

	// Boundaries
	BoundaryNorth         string `json:"boundaryNorth,omitempty" bson:"boundaryNorth,omitempty"`
	BoundarySouth         string `json:"boundarySouth,omitempty" bson:"boundarySouth,omitempty"`
	BoundaryEast          string `json:"boundaryEast,omitempty" bson:"boundaryEast,omitempty"`
	BoundaryWest          string `json:"boundaryWest,omitempty" bson:"boundaryWest,omitempty"`
	BoundaryHouses        string `json:"boundaryHouses,omitempty" bson:"boundaryHouses,omitempty"`
	BoundaryElectricPoles string `json:"boundaryElectricPoles,omitempty" bson:"boundaryElectricPoles,omitempty"`
	BoundaryWater         string `json:"boundaryWater,omitempty" bson:"boundaryWater,omitempty"`
	BoundaryRoads         string `json:"boundaryRoads,omitempty" bson:"boundaryRoads,omitempty"`
	BoundaryOther         string `json:"boundaryOther,omitempty" bson:"boundaryOther,omitempty"`

	// Financial / administrative
	ProposedExtent           string `json:"proposedExtent,omitempty" bson:"proposedExtent,omitempty"`
	RefundServiceFee         string `json:"refundServiceFee,omitempty" bson:"refundServiceFee,omitempty"`
	NGJARefNumber            string `json:"ngjaRefNumber,omitempty" bson:"ngjaRefNumber,omitempty"`
	MaxExtentVA              string `json:"maxExtentVA,omitempty" bson:"maxExtentVA,omitempty"`
	MaxPcCount               string `json:"maxPcCount,omitempty" bson:"maxPcCount,omitempty"`
	BackhoeCount             string `json:"backhoeCount,omitempty" bson:"backhoeCount,omitempty"`
	GerumCount               string `json:"gerumCount,omitempty" bson:"gerumCount,omitempty"`

	// Deprecated: no longer editable from the frontend, but kept so older documents
	// that still have these values return them correctly in API responses.
	AdumMachineCount          string `json:"adumMachineCount,omitempty" bson:"adumMachineCount,omitempty"`
	SilageExtent              string `json:"silageExtent,omitempty" bson:"silageExtent,omitempty"`
	DepositAmount             string `json:"depositAmount,omitempty" bson:"depositAmount,omitempty"`
	RiverbankProtectionAmount string `json:"riverbankProtectionAmount,omitempty" bson:"riverbankProtectionAmount,omitempty"`
	SpecialCaseAmount         string `json:"specialCaseAmount,omitempty" bson:"specialCaseAmount,omitempty"`

	// Metadata
	CreatedBy string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	Status    string `json:"status" bson:"status"` // "draft" | "submitted" | "approved" | "rejected"
}

// PaginatedMiningLicenses represents a paginated list of license applications
type PaginatedMiningLicenses struct {
	Data       []MechanizedGemMiningLicense `json:"data"`
	Total      int64                        `json:"total"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"totalPages"`
}

// MiningLicenseSummary is a slim projection of MechanizedGemMiningLicense
// used for listing every edition (version) of a given base reference number.
type MiningLicenseSummary struct {
	ID               primitive.ObjectID `json:"id" bson:"_id"`
	ReferenceNumber  string             `json:"referenceNumber" bson:"referenceNumber"`
	ApplicantName    string             `json:"applicantName" bson:"applicantName"`
	PrivateSaleValue *string            `json:"privateSaleValue" bson:"privateSaleValue"`
	CreatedBy        string             `json:"createdBy" bson:"createdBy"`
	CreatedAt        time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt" bson:"updatedAt"`
	Status           string             `json:"status" bson:"status"`
}

// PaginatedMiningLicenseSummaries represents a paginated list of license
// summaries, all sharing the same base reference number.
type PaginatedMiningLicenseSummaries struct {
	Data       []MiningLicenseSummary `json:"data"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"totalPages"`
}

// MiningLicenseRepository defines DB operations for mining license applications
type MiningLicenseRepository interface {
	Create(ctx context.Context, license *MechanizedGemMiningLicense) error
	GetByID(ctx context.Context, id string) (*MechanizedGemMiningLicense, error)
	GetAll(ctx context.Context) ([]MechanizedGemMiningLicense, error)
	GetByTIN(ctx context.Context, tin string, page int, limit int) (*PaginatedMiningLicenses, error)
	GetForMap(ctx context.Context, filters MapFilters) ([]MechanizedGemMiningLicense, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetNextBaseReferenceNumber(ctx context.Context) (int64, error)
	// GetMaxVersionByBaseRef returns the highest version suffix (e.g. 2 for "REF_2.2") stored
	// for the given base reference number (e.g. "REF_2"). Returns 0 if none exist yet.
	GetMaxVersionByBaseRef(ctx context.Context, baseRef string) (int, error)
	// GetByBaseReferenceNumber returns a slim, paginated view of every edition
	// (base ref + all its versioned suffixes) for a given base reference number.
	GetByBaseReferenceNumber(ctx context.Context, baseRef string, page int, limit int) (*PaginatedMiningLicenseSummaries, error)
}

// MiningLicenseUsecase defines business logic for mining license applications
type MiningLicenseUsecase interface {
	Submit(ctx context.Context, license *MechanizedGemMiningLicense) (*MechanizedGemMiningLicense, error)
	GetByID(ctx context.Context, id string) (*MechanizedGemMiningLicense, error)
	GetAll(ctx context.Context) ([]MechanizedGemMiningLicense, error)
	GetByTIN(ctx context.Context, tin string, page int, limit int) (*PaginatedMiningLicenses, error)
	GetForMap(ctx context.Context, filters MapFilters) ([]MechanizedGemMiningLicense, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Edit(ctx context.Context, id string, updatedLicense *MechanizedGemMiningLicense) (*MechanizedGemMiningLicense, error)
	// GetByReferenceNumber returns every edition of a license sharing the
	// same base reference number (e.g. "REF_4" -> REF_4, REF_4.1, REF_4.2 ...).
	GetByReferenceNumber(ctx context.Context, refNumber string, page int, limit int) (*PaginatedMiningLicenseSummaries, error)
}
