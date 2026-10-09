package domain

import (
	"context"
	"errors"
)

var (
	ErrInvalidYear  = errors.New("year is required and must be between 2000 and 2100")
	ErrInvalidGrade = errors.New("grade must be a number (1-2 digits)")
)

// AmountTotals holds payment counts and amounts. "Count" is the number of monthly payments
// (month entries); each payment is worth the report card's monthly "amount".
type AmountTotals struct {
	Count        int64   `json:"count" bson:"count"`
	PaidCount    int64   `json:"paidCount" bson:"paidCount"`
	UnpaidCount  int64   `json:"unpaidCount" bson:"unpaidCount"`
	TotalAmount  float64 `json:"totalAmount" bson:"totalAmount"`
	PaidAmount   float64 `json:"paidAmount" bson:"paidAmount"`
	UnpaidAmount float64 `json:"unpaidAmount" bson:"unpaidAmount"`
}

// AnnualSummary is the grand total of the year (all matching cards, never just one page)
type AnnualSummary struct {
	TotalCards   int64 `json:"totalCards" bson:"totalCards"`
	AmountTotals `bson:",inline"`
}

// AnnualMonthTotal is the total of one calendar month of the year
type AnnualMonthTotal struct {
	Month        int    `json:"month" bson:"_id"`
	MonthLabel   string `json:"monthLabel" bson:"-"`
	AmountTotals `bson:",inline"`
}

// AnnualGradeTotal is the total of one grade (current_grade) for the year
type AnnualGradeTotal struct {
	Grade        string `json:"grade" bson:"_id"`
	Cards        int64  `json:"cards" bson:"cards"`
	AmountTotals `bson:",inline"`
}

// AnnualAggregates is what the repository computes in MongoDB over ALL matching cards
type AnnualAggregates struct {
	Summary AnnualSummary
	ByMonth []AnnualMonthTotal
	ByGrade []AnnualGradeTotal
}

// AnnualReportRow is one report card with ONLY its months of the selected year, plus subtotals
type AnnualReportRow struct {
	ReportCard   `bson:",inline"`
	MonthsInYear int     `json:"monthsInYear" bson:"-"`
	PaidMonths   int     `json:"paidMonths" bson:"-"`
	UnpaidMonths int     `json:"unpaidMonths" bson:"-"`
	YearAmount   float64 `json:"yearAmount" bson:"-"`
	PaidAmount   float64 `json:"paidAmount" bson:"-"`
	UnpaidAmount float64 `json:"unpaidAmount" bson:"-"`
}

// AnnualReport is the report for one year (optionally one grade)
type AnnualReport struct {
	Year    int                `json:"year"`
	Grade   string             `json:"grade,omitempty"` // the grade filter used, if any
	Summary AnnualSummary      `json:"summary"`
	ByMonth []AnnualMonthTotal `json:"byMonth"` // always 12 entries, January to December
	ByGrade []AnnualGradeTotal `json:"byGrade"`
	Data    []*AnnualReportRow `json:"data"`
}

// PaginatedAnnualReport is the paginated version of the report
type PaginatedAnnualReport struct {
	AnnualReport
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// AnnualReportRepository defines the data access for the annual report
type AnnualReportRepository interface {
	// GetAnnualReport returns the rows (limit = 0 means all) and the aggregates over ALL matching cards.
	GetAnnualReport(ctx context.Context, year int, grade string, skip int64, limit int64) ([]*AnnualReportRow, *AnnualAggregates, error)
}

// AnnualReportUsecase defines the business logic for the annual report
type AnnualReportUsecase interface {
	GetAnnualReport(ctx context.Context, year int, grade string, page int, limit int) (*PaginatedAnnualReport, error)
	GetAnnualReportAll(ctx context.Context, year int, grade string) (*AnnualReport, error)
	GenerateAnnualReportPDF(ctx context.Context, year int, grade string) ([]byte, error)
}