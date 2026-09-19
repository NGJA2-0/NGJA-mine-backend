package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"my-fiber-app/domain"
)

type reportCardUsecase struct {
	repo domain.ReportCardRepository
}

func NewReportCardUsecase(repo domain.ReportCardRepository) domain.ReportCardUsecase {
	return &reportCardUsecase{repo: repo}
}

func (u *reportCardUsecase) Create(ctx context.Context, card *domain.ReportCard) error {
	// Validate required client-side fields
	if strings.TrimSpace(card.ApplicationID) == "" {
		return errors.New("applicationId is required")
	}
	if strings.TrimSpace(card.FullName) == "" {
		return errors.New("fullName is required")
	}
	if strings.TrimSpace(card.AccNumber) == "" {
		return errors.New("accNumber is required")
	}
	if strings.TrimSpace(card.NIC) == "" {
		return errors.New("nic is required")
	}
	if strings.TrimSpace(card.StartDate) == "" {
		return errors.New("startDate is required")
	}
	if strings.TrimSpace(card.EndDate) == "" {
		return errors.New("endDate is required")
	}
	if card.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	// Parse and validate dates
	start, err := time.Parse("2006-01-02", card.StartDate)
	if err != nil {
		return errors.New("startDate must be a valid ISO date (YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", card.EndDate)
	if err != nil {
		return errors.New("endDate must be a valid ISO date (YYYY-MM-DD)")
	}
	if !end.After(start) {
		return errors.New("endDate must be after startDate")
	}

	// Calculate total duration in months
	months := (end.Year()-start.Year())*12 + int(end.Month()-start.Month())
	card.TotalDuration = months

	// Calculate total amount
	card.TotalAmount = float64(months) * card.Amount

	// Auto-generate unique ref number: REF-rec-1, REF-rec-2, ...
	latestRef, _ := u.repo.GetLatestRefNumber(ctx)
	nextNum := 1
	if latestRef != "" {
		// Parse the trailing number from "REF-rec-N"
		parts := strings.Split(latestRef, "-")
		if len(parts) == 3 {
			if num, err := strconv.Atoi(parts[2]); err == nil {
				nextNum = num + 1
			}
		}
	}
	card.RefNumber = fmt.Sprintf("REF-rec-%d", nextNum)

	// Save PDF in root folder (placeholder path)
	card.PdfUrl = fmt.Sprintf("report_cards/%s.pdf", card.RefNumber)

	card.CreatedAt = time.Now()

	return u.repo.Create(ctx, card)
}
