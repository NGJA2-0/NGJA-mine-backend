package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidReportCardID = errors.New("invalid report card id")
	ErrInvalidPeriod       = errors.New("year must be between 2000 and 2100 and month must be between 1 and 12")
	ErrMonthNotFound       = errors.New("report card or month not found")
	ErrMonthAlreadyPaid    = errors.New("this month is already paid")
	ErrUserNotFound        = errors.New("user not found")
)

// ReportCard represents a report card document stored in MongoDB
type ReportCard struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ApplicationID string             `json:"applicationId" bson:"applicationId"`
	FullName      string             `json:"fullName" bson:"fullName"`
	AccNumber     string             `json:"accNumber" bson:"accNumber"`
	NIC           string             `json:"nic" bson:"nic"`
	AppliedGrade  string             `json:"appliedGrade" bson:"applied_grade"`
	CurrentGrade  string             `json:"currentGrade" bson:"current_grade"`
	RefNumber     string             `json:"refNumber,omitempty" bson:"refNumber,omitempty"`
	PdfUrl        string             `json:"pdfUrl,omitempty" bson:"pdfUrl,omitempty"`
	StartDate     string             `json:"startDate" bson:"startDate"`
	EndDate       string             `json:"endDate" bson:"endDate"`
	Amount        float64            `json:"amount" bson:"amount"`
	TotalDuration int                `json:"totalDuration" bson:"totalDuration"`
	TotalAmount   float64            `json:"totalAmount" bson:"totalAmount"`
	Months        []ReportCardMonth  `json:"months,omitempty" bson:"months,omitempty"`
	CreatedAt     time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// ReportCardMonth is one payable month of a report card
type ReportCardMonth struct {
	Year     int        `json:"year" bson:"year"`
	Month    int        `json:"month" bson:"month"`
	Label    string     `json:"label" bson:"label"` // e.g. "2026-October"
	Paid     bool       `json:"paid" bson:"paid"`
	PaidAt   *time.Time `json:"paidAt,omitempty" bson:"paidAt,omitempty"`
	PaidBy   string     `json:"paidBy,omitempty" bson:"paidBy,omitempty"`
	PaidByID string     `json:"paidById,omitempty" bson:"paidById,omitempty"`
}

// ReportCardSearchSuggestion is the response shape for search/dropdown results
type ReportCardSearchSuggestion struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	FullName      string             `json:"fullName" bson:"fullName"`
	AccNumber     string             `json:"accNumber" bson:"accNumber"`
	NIC           string             `json:"nic" bson:"nic"`
	AppliedGrade  string             `json:"appliedGrade" bson:"applied_grade"`
	CurrentGrade  string             `json:"currentGrade" bson:"current_grade"`
	RegionalOffice string             `json:"regionalOffice" bson:"regionalOffice"`
	StartDate     string             `json:"startDate" bson:"startDate"`
	EndDate       string             `json:"endDate" bson:"endDate"`
	Amount        float64            `json:"amount" bson:"amount"`
	TotalDuration int                `json:"totalDuration" bson:"totalDuration"`
	TotalAmount   float64            `json:"totalAmount" bson:"totalAmount"`
}

// PaginatedReportCards represents a paginated list of report cards
type PaginatedReportCards struct {
	Data       []*ReportCard `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"totalPages"`
}

// ReportCardMonthRow is a report card plus the status of the month that was searched
type ReportCardMonthRow struct {
	ReportCard
	SelectedMonth *ReportCardMonth `json:"selectedMonth"`
}

// PaginatedMonthReportCards is the response of the year/month search
type PaginatedMonthReportCards struct {
	Data       []*ReportCardMonthRow `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"totalPages"`
}

const OLCertificateKey = "additional"

type DocumentVersion struct {
	Version    int32     `json:"version" bson:"version"`
	FileName   string    `json:"fileName" bson:"fileName"`
	StoredPath string    `json:"-" bson:"storedPath"`
	Size       int64     `json:"size" bson:"size"`
	UploadedBy string    `json:"uploadedBy" bson:"uploadedBy"`
	UploadedAt time.Time `json:"uploadedAt" bson:"uploadedAt"`
}

// ReportCardRepository defines the data access interface
type ReportCardRepository interface {
	Create(ctx context.Context, card *ReportCard) error
	GetLatestRefNumber(ctx context.Context) (string, error)
	Search(ctx context.Context, query string) ([]*ReportCardSearchSuggestion, error)
	GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*PaginatedReportCards, error)
	ListSubmissions(ctx context.Context, grade string, year int, page int, limit int) (*PaginatedReportCards, error)
	GetOLCertificate(ctx context.Context, applicationID string) (*DocumentVersion, error)
	AddOLCertificate(ctx context.Context, applicationID string, v DocumentVersion) error
	ListByMonth(ctx context.Context, year int, month int, grade string, page int, limit int) ([]*ReportCard, int64, error)
	MarkMonthPaid(ctx context.Context, id string, year int, month int, paidBy string, paidByID string, paidAt time.Time) error
}

// ReportCardUsecase defines the business logic interface
type ReportCardUsecase interface {
	Create(ctx context.Context, card *ReportCard) error
	Search(ctx context.Context, query string) ([]*ReportCardSearchSuggestion, error)
	GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*PaginatedReportCards, error)
	ListSubmissions(ctx context.Context, grade string, page int, limit int) (*PaginatedReportCards, error)
	GetOLCertificate(ctx context.Context, applicationID string) (*DocumentVersion, error)
	AddOLCertificate(ctx context.Context, applicationID string, v DocumentVersion) error
	ListByMonth(ctx context.Context, year int, month int, grade string, page int, limit int) (*PaginatedMonthReportCards, error)
	PayMonth(ctx context.Context, id string, year int, month int, userID string) (*ReportCardMonth, error)
}