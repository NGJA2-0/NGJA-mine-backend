package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidReportCardID = errors.New("invalid report card id")
	ErrInvalidPeriod       = errors.New("year must be between 2000 and 2100, months must be between 1 and 12, and fromMonth must not be after toMonth")
	ErrMonthNotFound       = errors.New("report card not found, or it has no months in this range")
	ErrMonthAlreadyPaid    = errors.New("all months in this range are already paid")
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

const MaxReportCardsPerYear = 2

type GradeLimit struct {
	AppliedGrade  string `json:"appliedGrade"`
	MinGrade      int    `json:"minGrade"`
	CardsThisYear int    `json:"cardsThisYear"`
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
	SelectedMonths []ReportCardMonth `json:"selectedMonths"` // the card's months inside the searched range
	PaidCount      int               `json:"paidCount"`
	UnpaidCount    int               `json:"unpaidCount"`
}

// PayMonthsResult is the response of a pay action
type PayMonthsResult struct {
	Year             int               `json:"year"`
	FromMonth        int               `json:"fromMonth"`
	ToMonth          int               `json:"toMonth"`
	PaidMonths       []ReportCardMonth `json:"paidMonths"`       // months switched to paid by this click
	PaidCount        int               `json:"paidCount"`
	AlreadyPaidCount int               `json:"alreadyPaidCount"` // months in range that were already paid (left untouched)
	PaidAt           time.Time         `json:"paidAt"`
	PaidBy           string            `json:"paidBy"`
}

// MonthlyReportRow is a report card without its months array, plus only the selected month
type MonthlyReportRow struct {
	ReportCard `bson:",inline"`
	Month      *ReportCardMonth `json:"month" bson:"month"`
}

// MonthlyReportSummary holds the totals for ALL matching cards of the month (not just one page).
// Amounts use each card's monthly "amount" field.
type MonthlyReportSummary struct {
	TotalCount   int64   `json:"totalCount" bson:"totalCount"`
	PaidCount    int64   `json:"paidCount" bson:"paidCount"`
	UnpaidCount  int64   `json:"unpaidCount" bson:"unpaidCount"`
	TotalAmount  float64 `json:"totalAmount" bson:"totalAmount"`
	PaidAmount   float64 `json:"paidAmount" bson:"paidAmount"`
	UnpaidAmount float64 `json:"unpaidAmount" bson:"unpaidAmount"`
}

// MonthlyReport is the report for one year + month
type MonthlyReport struct {
	Year       int                  `json:"year"`
	Month      int                  `json:"month"`
	MonthLabel string               `json:"monthLabel"`
	Summary    MonthlyReportSummary `json:"summary"`
	Data       []*MonthlyReportRow  `json:"data"`
}

// PaginatedMonthlyReport is the paginated version of the report
type PaginatedMonthlyReport struct {
	MonthlyReport
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
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
	ListByMonth(ctx context.Context, year int, fromMonth int, toMonth int, grade string, page int, limit int) ([]*ReportCard, int64, error)
	MarkMonthsPaid(ctx context.Context, id string, year int, fromMonth int, toMonth int, paidBy string, paidByID string, paidAt time.Time) ([]ReportCardMonth, int, error)
	GetMonthlyReport(ctx context.Context, year int, month int, skip int64, limit int64) ([]*MonthlyReportRow, *MonthlyReportSummary, error)
	GetGradeLimit(ctx context.Context, applicationID string) (*GradeLimit, error)
}

// ReportCardUsecase defines the business logic interface
type ReportCardUsecase interface {
	Create(ctx context.Context, card *ReportCard) error
	Search(ctx context.Context, query string) ([]*ReportCardSearchSuggestion, error)
	GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*PaginatedReportCards, error)
	ListSubmissions(ctx context.Context, grade string, page int, limit int) (*PaginatedReportCards, error)
	GetOLCertificate(ctx context.Context, applicationID string) (*DocumentVersion, error)
	AddOLCertificate(ctx context.Context, applicationID string, v DocumentVersion) error
	ListByMonth(ctx context.Context, year int, fromMonth int, toMonth int, grade string, page int, limit int) (*PaginatedMonthReportCards, error)
	PayMonths(ctx context.Context, id string, year int, fromMonth int, toMonth int, userID string) (*PayMonthsResult, error)
	GetMonthlyReport(ctx context.Context, year int, month int, page int, limit int) (*PaginatedMonthlyReport, error)
	GetMonthlyReportAll(ctx context.Context, year int, month int) (*MonthlyReport, error)
	GetGradeLimit(ctx context.Context, applicationID string) (*GradeLimit, error)
}

