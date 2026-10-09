package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrApplicationIDRequired = errors.New("applicationId is required")

// ApplicationPaymentMonth is ONE month of ONE report card, with the card info next to it
type ApplicationPaymentMonth struct {
	CardID       primitive.ObjectID `json:"cardId" bson:"cardId"`
	RefNumber    string             `json:"refNumber" bson:"refNumber"`
	CurrentGrade string             `json:"currentGrade" bson:"currentGrade"`
	Amount       float64            `json:"amount" bson:"amount"`
	Year         int                `json:"year" bson:"year"`
	Month        int                `json:"month" bson:"month"`
	Label        string             `json:"label" bson:"label"`
	Paid         bool               `json:"paid" bson:"paid"`
	PaidAt       *time.Time         `json:"paidAt,omitempty" bson:"paidAt,omitempty"`
	PaidBy       string             `json:"paidBy,omitempty" bson:"paidBy,omitempty"`
	PaidByID     string             `json:"paidById,omitempty" bson:"paidById,omitempty"`
}

// ApplicationPaymentSummary holds the totals over ALL months of the application (never affected by pagination)
type ApplicationPaymentSummary struct {
	TotalMonths  int64   `json:"totalMonths" bson:"totalMonths"`
	PaidMonths   int64   `json:"paidMonths" bson:"paidMonths"`
	UnpaidMonths int64   `json:"unpaidMonths" bson:"unpaidMonths"`
	TotalAmount  float64 `json:"totalAmount" bson:"totalAmount"`
	PaidAmount   float64 `json:"paidAmount" bson:"paidAmount"`
	UnpaidAmount float64 `json:"unpaidAmount" bson:"unpaidAmount"`
}

type ApplicationPaymentYearGroup struct {
	Year   int                        `json:"year"`
	Months []*ApplicationPaymentMonth `json:"months"`
}

type PaginatedApplicationPayments struct {
	ApplicationID string                        `json:"applicationId"`
	Summary       ApplicationPaymentSummary     `json:"summary"`
	Data          []ApplicationPaymentYearGroup `json:"data"`
	Total         int64                         `json:"total"` // total months, not total cards
	Page          int                           `json:"page"`
	Limit         int                           `json:"limit"`
	TotalPages    int                           `json:"totalPages"`
}

type ApplicationPaymentRepository interface {
	GetByApplication(ctx context.Context, applicationID string, skip int64, limit int64) ([]*ApplicationPaymentMonth, *ApplicationPaymentSummary, error)
}

type ApplicationPaymentUsecase interface {
	GetByApplication(ctx context.Context, applicationID string, page int, limit int) (*PaginatedApplicationPayments, error)
}