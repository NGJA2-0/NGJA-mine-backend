package domain

import "context"

type MonthlyPaymentPoint struct {
	Month  string  `json:"month"` // "Jan".."Dec"
	Paid   float64 `json:"paid"`
	Unpaid float64 `json:"unpaid"`
}

type AnnualPaymentPoint struct {
	Year   int     `json:"year"`
	Total  float64 `json:"total"` // paid amount, used by the donut
	Unpaid float64 `json:"unpaid"`
}

type PaymentOverview struct {
	Year    int                   `json:"year"`
	Monthly []MonthlyPaymentPoint `json:"monthly"` // always 12 items
	Annual  []AnnualPaymentPoint  `json:"annual"`  // always 5 items, oldest first
}

// PaymentOverviewRow is one (year, month) bucket straight from MongoDB
type PaymentOverviewRow struct {
	Year   int     `bson:"year"`
	Month  int     `bson:"month"`
	Paid   float64 `bson:"paid"`
	Unpaid float64 `bson:"unpaid"`
}

type PaymentOverviewRepository interface {
	GetOverview(ctx context.Context, fromYear, toYear int) ([]PaymentOverviewRow, error)
}

type PaymentOverviewUsecase interface {
	GetOverview(ctx context.Context, year int) (*PaymentOverview, error)
}