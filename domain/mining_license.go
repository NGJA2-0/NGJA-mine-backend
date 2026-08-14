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

// MechanizedGemMiningLicense represents the full license application document stored in MongoDB
type MechanizedGemMiningLicense struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`

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
	AdumMachineCount         string `json:"adumMachineCount,omitempty" bson:"adumMachineCount,omitempty"`
	SilageExtent             string `json:"silageExtent,omitempty" bson:"silageExtent,omitempty"`
	DepositAmount            string `json:"depositAmount,omitempty" bson:"depositAmount,omitempty"`
	RiverbankProtectionAmount string `json:"riverbankProtectionAmount,omitempty" bson:"riverbankProtectionAmount,omitempty"`
	SpecialCaseAmount        string `json:"specialCaseAmount,omitempty" bson:"specialCaseAmount,omitempty"`

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

// MiningLicenseRepository defines DB operations for mining license applications
type MiningLicenseRepository interface {
	Create(ctx context.Context, license *MechanizedGemMiningLicense) error
	GetByID(ctx context.Context, id string) (*MechanizedGemMiningLicense, error)
	GetAll(ctx context.Context) ([]MechanizedGemMiningLicense, error)
	GetByTIN(ctx context.Context, tin string, page int, limit int) (*PaginatedMiningLicenses, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

// MiningLicenseUsecase defines business logic for mining license applications
type MiningLicenseUsecase interface {
	Submit(ctx context.Context, license *MechanizedGemMiningLicense) (*MechanizedGemMiningLicense, error)
	GetByID(ctx context.Context, id string) (*MechanizedGemMiningLicense, error)
	GetAll(ctx context.Context) ([]MechanizedGemMiningLicense, error)
	GetByTIN(ctx context.Context, tin string, page int, limit int) (*PaginatedMiningLicenses, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
