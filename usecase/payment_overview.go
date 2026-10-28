package usecase

import (
	"context"
	"math"
	"time"

	"my-fiber-app/domain"
)

var monthShort = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

type paymentOverviewUsecase struct {
	repo domain.PaymentOverviewRepository
}

func NewPaymentOverviewUsecase(repo domain.PaymentOverviewRepository) domain.PaymentOverviewUsecase {
	return &paymentOverviewUsecase{repo: repo}
}

func (u *paymentOverviewUsecase) GetOverview(ctx context.Context, year int) (*domain.PaymentOverview, error) {
	if year < 2000 || year > 2100 {
		year = time.Now().Year()
	}
	fromYear := year - 4

	rows, err := u.repo.GetOverview(ctx, fromYear, year)
	if err != nil {
		return nil, err
	}

	round2 := func(v float64) float64 { return math.Round(v*100) / 100 }

	// Pre-fill so months / years without data come back as zeros
	monthly := make([]domain.MonthlyPaymentPoint, 12)
	for i := range monthly {
		monthly[i].Month = monthShort[i]
	}
	annual := make([]domain.AnnualPaymentPoint, 5)
	for i := range annual {
		annual[i].Year = fromYear + i
	}

	for _, r := range rows {
		if r.Year == year && r.Month >= 1 && r.Month <= 12 {
			monthly[r.Month-1].Paid += r.Paid
			monthly[r.Month-1].Unpaid += r.Unpaid
		}
		if idx := r.Year - fromYear; idx >= 0 && idx < 5 {
			annual[idx].Total += r.Paid
			annual[idx].Unpaid += r.Unpaid
		}
	}
	for i := range monthly {
		monthly[i].Paid = round2(monthly[i].Paid)
		monthly[i].Unpaid = round2(monthly[i].Unpaid)
	}
	for i := range annual {
		annual[i].Total = round2(annual[i].Total)
		annual[i].Unpaid = round2(annual[i].Unpaid)
	}

	return &domain.PaymentOverview{Year: year, Monthly: monthly, Annual: annual}, nil
}