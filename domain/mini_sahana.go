package domain

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MiniSahanaAttachments struct {
	BankPassbookCopy     string `json:"bankPassbookCopy" bson:"bankPassbookCopy" validate:"required"`
	BirthCertificateCopy string `json:"birthCertificateCopy" bson:"birthCertificateCopy" validate:"required"`
}

type MiniSahanaForm struct {
	ID                           primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty"`
	FormType                     string                `json:"formType" bson:"formType" validate:"required"`
	Year                         string                `json:"year" bson:"year" validate:"required"`
	ApplicantFullNameSinhala     string                `json:"applicantFullNameSinhala" bson:"applicantFullNameSinhala" validate:"required"`
	ApplicantFullNameEnglish     string                `json:"applicantFullNameEnglish" bson:"applicantFullNameEnglish" validate:"required"`
	NameWithInitialsEnglish      string                `json:"nameWithInitialsEnglish" bson:"nameWithInitialsEnglish" validate:"required"`
	DateOfBirth                  string                `json:"dateOfBirth" bson:"dateOfBirth" validate:"required"` // Validated manually
	Gender                       string                `json:"gender" bson:"gender" validate:"required"`
	School                       string                `json:"school" bson:"school" validate:"required"`
	Grade                        string                `json:"grade" bson:"grade" validate:"required"`
	SchoolAddressAndPhone        string                `json:"schoolAddressAndPhone" bson:"schoolAddressAndPhone" validate:"required"`
	BankAccountName              string                `json:"bankAccountName" bson:"bankAccountName" validate:"required"`
	BankBranch                   string                `json:"bankBranch" bson:"bankBranch" validate:"required"`
	BankAccountNumber            string                `json:"bankAccountNumber" bson:"bankAccountNumber" validate:"required"`
	EligibilityCategory          []string              `json:"eligibilityCategory" bson:"eligibilityCategory" validate:"required,min=1"`
	ParentGuardianRelation       string                `json:"parentGuardianRelation" bson:"parentGuardianRelation" validate:"required"`
	ParentGuardianName           string                `json:"parentGuardianName" bson:"parentGuardianName" validate:"required"`
	PermanentAddress             string                `json:"permanentAddress" bson:"permanentAddress" validate:"required"`
	PhoneNumber                  string                `json:"phoneNumber" bson:"phoneNumber" validate:"required"`
	NIC                          string                `json:"nic" bson:"nic" validate:"required"` // Validated manually
	Age                          *int                  `json:"age" bson:"age" validate:"required"`
	Occupation                   string                `json:"occupation" bson:"occupation" validate:"required"`
	MonthlyIncome                *float64              `json:"monthlyIncome" bson:"monthlyIncome" validate:"required"`
	MaritalStatus                string                `json:"maritalStatus" bson:"maritalStatus" validate:"required"`
	SpouseRelation               string                `json:"spouseRelation" bson:"spouseRelation" validate:"required"`
	SpouseName                   string                `json:"spouseName" bson:"spouseName" validate:"required"`
	SpouseOccupation             string                `json:"spouseOccupation" bson:"spouseOccupation" validate:"required"`
	ChildrenCount                *int                  `json:"childrenCount" bson:"childrenCount" validate:"required"`
	LicenseNumber                string                `json:"licenseNumber" bson:"licenseNumber" validate:"required"`
	FileNumber                   string                `json:"fileNumber" bson:"fileNumber" validate:"required"`
	LicenseRegionalOfficeAndZone string                `json:"licenseRegionalOfficeAndZone" bson:"licenseRegionalOfficeAndZone" validate:"required"`
	Attachments                  MiniSahanaAttachments `json:"attachments" bson:"attachments" validate:"required"`
	DeclarationSigned            *bool                 `json:"declarationSigned" bson:"declarationSigned" validate:"required"`
	CreatedBy                    string                `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	RefNumber                    string                `json:"refNumber,omitempty" bson:"refNumber,omitempty"`
	CreatedAt                    time.Time             `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedBy                    string                 `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	Changes                      []MiniSahanaChangeEntry `json:"changes" bson:"changes"`
	Documents 					[]MiniSahanaDocument `json:"documents" bson:"documents"`
}

type MiniSahanaChangeEntry struct {
	EditType   string             `json:"editType" bson:"editType"`     // "normal" or "acc_number"
	ChangedBy  string             `json:"changedBy" bson:"changedBy"`
	ChangedAt  time.Time          `json:"changedAt" bson:"changedAt"`
	OldValues  map[string]interface{} `json:"oldValues" bson:"oldValues"`
	NewValues  map[string]interface{} `json:"newValues" bson:"newValues"`
}

type MiniSahanaSearchSuggestion struct {
	ID                       primitive.ObjectID `json:"id" bson:"_id"`
	ApplicantFullNameSinhala string             `json:"applicantFullNameSinhala" bson:"applicantFullNameSinhala"`
	BankAccountNumber        string             `json:"bankAccountNumber" bson:"bankAccountNumber"`
	NIC                      string             `json:"nic" bson:"nic"`
	Grade                    string             `json:"grade" bson:"grade"`
}

type MiniSahanaYearCount struct {
	Year  string `json:"year"`
	Count int64  `json:"count"`
}

type MiniSahanaGradeCount struct {
	Grade string `json:"grade"`
	Count int64  `json:"count"`
}

type MiniSahanaStats struct {
	Total   int64                  `json:"total"`
	ByYear  []MiniSahanaYearCount  `json:"byYear"`
	ByGrade []MiniSahanaGradeCount `json:"byGrade"`
}

type MiniSahanaRepository interface {
	Create(ctx context.Context, form *MiniSahanaForm) error
	GetLatestRefNumber(ctx context.Context) (string, error)
	GetByID(ctx context.Context, id string) (*MiniSahanaForm, error)
	Search(ctx context.Context, query string) ([]*MiniSahanaSearchSuggestion, error)
	Update(ctx context.Context, id string, form *MiniSahanaForm) error
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*MiniSahanaStats, error)
}

type MiniSahanaUsecase interface {
	Create(ctx context.Context, form *MiniSahanaForm, userID string, files map[string]*DocUpload) error
	GetByID(ctx context.Context, id string) (*MiniSahanaForm, error)
	Search(ctx context.Context, query string) ([]*MiniSahanaSearchSuggestion, error)
	Update(ctx context.Context, id string, form *MiniSahanaForm, userID string) error
	UpdateAccountNumber(ctx context.Context, id string, newAccountNumber string, userID string) error
	UpdateDocument(ctx context.Context, id, docKey string, file *DocUpload, userID string) (*MiniSahanaForm, error)
	OpenDocument(ctx context.Context, id, docKey string, version int) (*MiniSahanaDocVersion, error)
	Delete(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*MiniSahanaStats, error)
}

const DocumentsRoot = "minisahana_documents"

func DocumentDir(appID string) string { return DocumentsRoot + "/" + appID }

func DocumentPath(appID, docKey string, version int) string {
	return fmt.Sprintf("%s/%s/%s_v%d.pdf", DocumentsRoot, appID, docKey, version)
}

type MiniSahanaDocumentSlot struct {
	Key       string
	FormField string // multipart field name sent by the frontend
	Label     string
	Required  bool
}

var MiniSahanaDocumentSlots = []MiniSahanaDocumentSlot{
	{"hard_copy", "hardCopy", "Submitted Hard Copy", true},
	{"bank_passbook", "passbook", "Copy of the Bank Passbook", true},
	{"birth_certificate", "birthCert", "Copy of the Birth Certificate", true},
	{"additional", "additional", "O/L Certificate", false},
}

type MiniSahanaDocVersion struct {
	Version    int       `json:"version" bson:"version"`
	FileName   string    `json:"fileName" bson:"fileName"`
	StoredPath string    `json:"-" bson:"storedPath"` // never sent to the browser
	Size       int64     `json:"size" bson:"size"`
	UploadedBy string    `json:"uploadedBy" bson:"uploadedBy"`
	UploadedAt time.Time `json:"uploadedAt" bson:"uploadedAt"`
}

type MiniSahanaDocument struct {
	Key            string                 `json:"key" bson:"key"`
	Label          string                 `json:"label" bson:"label"`
	CurrentVersion int                    `json:"currentVersion" bson:"currentVersion"`
	Versions       []MiniSahanaDocVersion `json:"versions" bson:"versions"`
}

// Keeps the usecase free of Fiber: Save is the handler's c.SaveFile wrapper.
type DocUpload struct {
	FileName string
	Size     int64
	Save     func(relPath string) error
}