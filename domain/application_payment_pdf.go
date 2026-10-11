package domain

import (
	"context"
	"errors"
)

var ErrNoPaymentRecords = errors.New("no payment records found for this application")

type ApplicationPaymentPDFUsecase interface {
	GeneratePDF(ctx context.Context, applicationID string) ([]byte, error)
}