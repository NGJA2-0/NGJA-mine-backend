package usecase

import (
	"context"
	"math"
	"strings"

	"my-fiber-app/domain"
)

type applicationPaymentUsecase struct {
	repo domain.ApplicationPaymentRepository
}

func NewApplicationPaymentUsecase(repo domain.ApplicationPaymentRepository) domain.ApplicationPaymentUsecase {
	return &applicationPaymentUsecase{repo: repo}
}

func (u *applicationPaymentUsecase) GetByApplication(ctx context.Context, applicationID string, page int, limit int) (*domain.PaginatedApplicationPayments, error) {
	applicationID = strings.TrimSpace(applicationID)
	if applicationID == "" {
		return nil, domain.ErrApplicationIDRequired
	}

	switch limit {
	case 10, 15, 20:
		// valid
	default:
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	rows, summary, err := u.repo.GetByApplication(ctx, applicationID, int64((page-1)*limit), int64(limit))
	if err != nil {
		return nil, err
	}

	round2 := func(v float64) float64 { return math.Round(v*100) / 100 }
	summary.TotalAmount = round2(summary.TotalAmount)
	summary.PaidAmount = round2(summary.PaidAmount)
	summary.UnpaidAmount = round2(summary.UnpaidAmount)

	// Rows are already sorted by year, so consecutive rows with the same year go in the same group
	groups := []domain.ApplicationPaymentYearGroup{}
	for _, row := range rows {
		if n := len(groups); n == 0 || groups[n-1].Year != row.Year {
			groups = append(groups, domain.ApplicationPaymentYearGroup{
				Year:   row.Year,
				Months: []*domain.ApplicationPaymentMonth{},
			})
		}
		g := &groups[len(groups)-1]
		g.Months = append(g.Months, row)
	}

	totalPages := int((summary.TotalMonths + int64(limit) - 1) / int64(limit))
	return &domain.PaginatedApplicationPayments{
		ApplicationID: applicationID,
		Summary:       *summary,
		Data:          groups,
		Total:         summary.TotalMonths,
		Page:          page,
		Limit:         limit,
		TotalPages:    totalPages,
	}, nil
}