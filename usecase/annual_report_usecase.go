package usecase

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"my-fiber-app/domain"
)

var gradePattern = regexp.MustCompile(`^\d{1,2}$`)

type annualReportUsecase struct {
	repo domain.AnnualReportRepository
}

func NewAnnualReportUsecase(repo domain.AnnualReportRepository) domain.AnnualReportUsecase {
	return &annualReportUsecase{repo: repo}
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundTotals(t *domain.AmountTotals) {
	t.TotalAmount = roundMoney(t.TotalAmount)
	t.PaidAmount = roundMoney(t.PaidAmount)
	t.UnpaidAmount = roundMoney(t.UnpaidAmount)
}

func validateAnnualInput(year int, grade string) (string, error) {
	if year < 2000 || year > 2100 {
		return "", domain.ErrInvalidYear
	}
	grade = strings.TrimSpace(grade)
	if grade != "" && !gradePattern.MatchString(grade) {
		return "", domain.ErrInvalidGrade
	}
	return grade, nil
}

// fillAnnualRowTotals calculates the per-card subtotals from the card's months of the year
func fillAnnualRowTotals(row *domain.AnnualReportRow) {
	row.MonthsInYear, row.PaidMonths, row.UnpaidMonths = 0, 0, 0
	for _, m := range row.Months {
		row.MonthsInYear++
		if m.Paid {
			row.PaidMonths++
		} else {
			row.UnpaidMonths++
		}
	}
	row.YearAmount = roundMoney(row.Amount * float64(row.MonthsInYear))
	row.PaidAmount = roundMoney(row.Amount * float64(row.PaidMonths))
	row.UnpaidAmount = roundMoney(row.Amount * float64(row.UnpaidMonths))
}

// buildAnnualReport loads the rows + aggregates and shapes them into the report
func (u *annualReportUsecase) buildAnnualReport(ctx context.Context, year int, grade string, skip int64, limit int64) (*domain.AnnualReport, error) {
	rows, agg, err := u.repo.GetAnnualReport(ctx, year, grade, skip, limit)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		fillAnnualRowTotals(row)
	}

	summary := agg.Summary
	roundTotals(&summary.AmountTotals)

	// Month-wise: always 12 entries, January to December (zeros for months with no payments)
	byMonthMap := make(map[int]domain.AnnualMonthTotal, 12)
	for _, m := range agg.ByMonth {
		byMonthMap[m.Month] = m
	}
	byMonth := make([]domain.AnnualMonthTotal, 0, 12)
	for m := 1; m <= 12; m++ {
		t := byMonthMap[m]
		t.Month = m
		t.MonthLabel = time.Month(m).String()
		roundTotals(&t.AmountTotals)
		byMonth = append(byMonth, t)
	}

	// Grade-wise: sorted numerically (2, 10, 11, 12 ...)
	byGrade := make([]domain.AnnualGradeTotal, 0, len(agg.ByGrade))
	for _, g := range agg.ByGrade {
		roundTotals(&g.AmountTotals)
		byGrade = append(byGrade, g)
	}
	sort.Slice(byGrade, func(i, j int) bool {
		a, aErr := strconv.Atoi(byGrade[i].Grade)
		b, bErr := strconv.Atoi(byGrade[j].Grade)
		if aErr == nil && bErr == nil {
			return a < b
		}
		return byGrade[i].Grade < byGrade[j].Grade
	})

	return &domain.AnnualReport{
		Year:    year,
		Grade:   grade,
		Summary: summary,
		ByMonth: byMonth,
		ByGrade: byGrade,
		Data:    rows,
	}, nil
}

func (u *annualReportUsecase) GetAnnualReport(ctx context.Context, year int, grade string, page int, limit int) (*domain.PaginatedAnnualReport, error) {
	grade, err := validateAnnualInput(year, grade)
	if err != nil {
		return nil, err
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

	report, err := u.buildAnnualReport(ctx, year, grade, int64((page-1)*limit), int64(limit))
	if err != nil {
		return nil, err
	}

	totalPages := int((report.Summary.TotalCards + int64(limit) - 1) / int64(limit))
	return &domain.PaginatedAnnualReport{
		AnnualReport: *report,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
	}, nil
}

func (u *annualReportUsecase) GetAnnualReportAll(ctx context.Context, year int, grade string) (*domain.AnnualReport, error) {
	grade, err := validateAnnualInput(year, grade)
	if err != nil {
		return nil, err
	}
	return u.buildAnnualReport(ctx, year, grade, 0, 0)
}

func (u *annualReportUsecase) GenerateAnnualReportPDF(ctx context.Context, year int, grade string) ([]byte, error) {
	report, err := u.GetAnnualReportAll(ctx, year, grade)
	if err != nil {
		return nil, err
	}
	return buildAnnualReportPDF(report)
}