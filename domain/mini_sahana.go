package domain

import (
	"context"
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
}

type MiniSahanaSearchSuggestion struct {
	ID                       primitive.ObjectID `json:"id" bson:"_id"`
	ApplicantFullNameSinhala string             `json:"applicantFullNameSinhala" bson:"applicantFullNameSinhala"`
	BankAccountNumber        string             `json:"bankAccountNumber" bson:"bankAccountNumber"`
	NIC                      string             `json:"nic" bson:"nic"`
}

type MiniSahanaRepository interface {
	Create(ctx context.Context, form *MiniSahanaForm) error
	GetLatestRefNumber(ctx context.Context) (string, error)
	GetByID(ctx context.Context, id string) (*MiniSahanaForm, error)
	Search(ctx context.Context, query string) ([]*MiniSahanaSearchSuggestion, error)
	Update(ctx context.Context, id string, form *MiniSahanaForm) error
	Delete(ctx context.Context, id string) error
}

type MiniSahanaUsecase interface {
	Create(ctx context.Context, form *MiniSahanaForm, userID string) error
	GetByID(ctx context.Context, id string) (*MiniSahanaForm, error)
	Search(ctx context.Context, query string) ([]*MiniSahanaSearchSuggestion, error)
	Update(ctx context.Context, id string, form *MiniSahanaForm, userID string) error
	Delete(ctx context.Context, id string) error
}
