package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	CreatedAt     time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// ReportCardSearchSuggestion is the response shape for search/dropdown results
type ReportCardSearchSuggestion struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	FullName      string             `json:"fullName" bson:"fullName"`
	AccNumber     string             `json:"accNumber" bson:"accNumber"`
	NIC           string             `json:"nic" bson:"nic"`
	AppliedGrade  string             `json:"appliedGrade" bson:"applied_grade"`
	CurrentGrade  string             `json:"currentGrade" bson:"current_grade"`
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

// ReportCardRepository defines the data access interface
type ReportCardRepository interface {
	Create(ctx context.Context, card *ReportCard) error
	GetLatestRefNumber(ctx context.Context) (string, error)
	Search(ctx context.Context, query string) ([]*ReportCardSearchSuggestion, error)
	GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*PaginatedReportCards, error)
}

// ReportCardUsecase defines the business logic interface
type ReportCardUsecase interface {
	Create(ctx context.Context, card *ReportCard) error
	Search(ctx context.Context, query string) ([]*ReportCardSearchSuggestion, error)
	GetByApplicationID(ctx context.Context, applicationID string, page int, limit int) (*PaginatedReportCards, error)
}
