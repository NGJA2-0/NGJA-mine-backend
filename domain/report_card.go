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
	RefNumber     string             `json:"refNumber,omitempty" bson:"refNumber,omitempty"`
	PdfUrl        string             `json:"pdfUrl,omitempty" bson:"pdfUrl,omitempty"`
	StartDate     string             `json:"startDate" bson:"startDate"`
	EndDate       string             `json:"endDate" bson:"endDate"`
	Amount        float64            `json:"amount" bson:"amount"`
	TotalDuration int                `json:"totalDuration" bson:"totalDuration"`
	TotalAmount   float64            `json:"totalAmount" bson:"totalAmount"`
	CreatedAt     time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// ReportCardRepository defines the data access interface
type ReportCardRepository interface {
	Create(ctx context.Context, card *ReportCard) error
	GetLatestRefNumber(ctx context.Context) (string, error)
}

// ReportCardUsecase defines the business logic interface
type ReportCardUsecase interface {
	Create(ctx context.Context, card *ReportCard) error
}
