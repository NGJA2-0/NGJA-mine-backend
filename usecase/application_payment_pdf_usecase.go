package usecase

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"my-fiber-app/domain"

	"github.com/go-pdf/fpdf"
)

// allRows is passed as the page size so the existing query returns every month (no pagination)
const allRows int64 = 1<<31 - 1

// Paid-at times are shown in Sri Lanka time (UTC+5:30)
var colomboTZ = time.FixedZone("+0530", 5*3600+30*60)

type pdfApplicant struct {
	FullName  string
	AccNumber string
	NIC       string
}

type applicationPaymentPDFUsecase struct {
	paymentRepo domain.ApplicationPaymentRepository
	cardRepo    domain.ReportCardRepository
}

func NewApplicationPaymentPDFUsecase(paymentRepo domain.ApplicationPaymentRepository, cardRepo domain.ReportCardRepository) domain.ApplicationPaymentPDFUsecase {
	return &applicationPaymentPDFUsecase{paymentRepo: paymentRepo, cardRepo: cardRepo}
}

func (u *applicationPaymentPDFUsecase) GeneratePDF(ctx context.Context, applicationID string) ([]byte, error) {
	applicationID = strings.TrimSpace(applicationID)
	if applicationID == "" {
		return nil, domain.ErrApplicationIDRequired
	}

	// Every month of the application + totals over all of them
	rows, summary, err := u.paymentRepo.GetByApplication(ctx, applicationID, 0, allRows)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrNoPaymentRecords
	}

	// Applicant details come from the latest report card of the application
	who := pdfApplicant{}
	cards, err := u.cardRepo.GetByApplicationID(ctx, applicationID, 1, 1)
	if err != nil {
		return nil, err
	}
	if cards != nil && len(cards.Data) > 0 {
		c := cards.Data[0]
		who = pdfApplicant{FullName: c.FullName, AccNumber: c.AccNumber, NIC: c.NIC}
	}

	summary.TotalAmount = roundPaymentAmount(summary.TotalAmount)
	summary.PaidAmount = roundPaymentAmount(summary.PaidAmount)
	summary.UnpaidAmount = roundPaymentAmount(summary.UnpaidAmount)

	return buildApplicationPaymentPDF(applicationID, who, *summary, groupPaymentsByYear(rows))
}

func roundPaymentAmount(v float64) float64 { return math.Round(v*100) / 100 }

// groupPaymentsByYear relies on the rows already being sorted by year
func groupPaymentsByYear(rows []*domain.ApplicationPaymentMonth) []domain.ApplicationPaymentYearGroup {
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
	return groups
}

func formatAmount(v float64) string {
	parts := strings.Split(fmt.Sprintf("%.2f", math.Abs(v)), ".")
	intPart := parts[0]

	var b strings.Builder
	if v < 0 {
		b.WriteByte('-')
	}
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return b.String() + "." + parts[1]
}

func buildApplicationPaymentPDF(applicationID string, who pdfApplicant, s domain.ApplicationPaymentSummary, groups []domain.ApplicationPaymentYearGroup) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AliasNbPages("")
	tr := pdf.UnicodeTranslatorFromDescriptor("") // UTF-8 -> the built-in font's encoding

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 6, fmt.Sprintf("Page %d of {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	// ---- Title ----
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(15, 23, 42)
	pdf.CellFormat(0, 10, "Application Payment Report", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(100, 116, 139)
	pdf.CellFormat(0, 5, "Generated on "+time.Now().In(colomboTZ).Format("02 Jan 2006, 15:04"), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	// ---- Applicant details ----
	kv := func(label, value string) {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(71, 85, 105)
		pdf.CellFormat(35, 6, label, "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 6, tr(value), "", 1, "L", false, 0, "")
	}
	kv("Applicant", who.FullName)
	kv("Account number", who.AccNumber)
	kv("NIC", who.NIC)
	kv("Application ID", applicationID)
	pdf.Ln(4)

	// ---- Summary boxes (totals over ALL months) ----
	boxes := []struct {
		label, value, sub string
		r, g, b           int
	}{
		{"PAID", formatAmount(s.PaidAmount), fmt.Sprintf("%d months", s.PaidMonths), 220, 252, 231},
		{"UNPAID", formatAmount(s.UnpaidAmount), fmt.Sprintf("%d months", s.UnpaidMonths), 254, 226, 226},
		{"TOTAL", formatAmount(s.TotalAmount), fmt.Sprintf("%d months", s.TotalMonths), 226, 232, 240},
	}
	boxY := pdf.GetY()
	for i, b := range boxes {
		x := 12 + float64(i)*(58+6)
		pdf.SetFillColor(b.r, b.g, b.b)
		pdf.RoundedRect(x, boxY, 58, 20, 2, "1234", "F")

		pdf.SetXY(x+4, boxY+3)
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(71, 85, 105)
		pdf.CellFormat(50, 4, b.label, "", 0, "L", false, 0, "")

		pdf.SetXY(x+4, boxY+8)
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(50, 6, b.value, "", 0, "L", false, 0, "")

		pdf.SetXY(x+4, boxY+14)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(71, 85, 105)
		pdf.CellFormat(50, 4, b.sub, "", 0, "L", false, 0, "")
	}
	pdf.SetXY(12, boxY+28)

	// ---- Month tables, one per year ----
	cols := []struct {
		title string
		w     float64
		align string
	}{
		{"Month", 30, "L"}, {"Ref No", 28, "L"}, {"Grade", 16, "C"}, {"Amount", 26, "R"},
		{"Status", 20, "C"}, {"Paid By", 32, "L"}, {"Paid At", 34, "L"},
	}

	// cuts the text so it never overflows its column
	fit := func(text string, w float64) string {
		text = tr(text)
		if pdf.GetStringWidth(text) <= w-2 {
			return text
		}
		for len(text) > 0 && pdf.GetStringWidth(text+"...") > w-2 {
			text = text[:len(text)-1]
		}
		return text + "..."
	}

	yearHeading := func(year int, continued bool) {
		title := fmt.Sprintf("%d", year)
		if continued {
			title += " (continued)"
		}
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(15, 23, 42)
		pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "B", 8.5)
		pdf.SetFillColor(30, 41, 59)
		pdf.SetTextColor(255, 255, 255)
		for _, c := range cols {
			pdf.CellFormat(c.w, 7, c.title, "", 0, c.align, true, 0, "")
		}
		pdf.Ln(7)
	}

	pdf.SetDrawColor(226, 232, 240)
	for _, g := range groups {
		if pdf.GetY() > 255 {
			pdf.AddPage() // don't leave a year heading alone at the bottom of a page
		}
		yearHeading(g.Year, false)

		for i, m := range g.Months {
			if pdf.GetY()+7 > 280 {
				pdf.AddPage()
				yearHeading(g.Year, true)
			}

			if i%2 == 1 {
				pdf.SetFillColor(248, 250, 252)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			paidBy, paidAt, status := "-", "-", "Unpaid"
			if m.Paid {
				status = "Paid"
				if m.PaidBy != "" {
					paidBy = m.PaidBy
				}
				if m.PaidAt != nil {
					paidAt = m.PaidAt.In(colomboTZ).Format("02 Jan 2006 15:04")
				}
			}

			pdf.SetFont("Helvetica", "", 8.5)
			pdf.SetTextColor(15, 23, 42)
			pdf.CellFormat(cols[0].w, 7, fit(m.Label, cols[0].w), "B", 0, "L", true, 0, "")
			pdf.CellFormat(cols[1].w, 7, fit(m.RefNumber, cols[1].w), "B", 0, "L", true, 0, "")
			pdf.CellFormat(cols[2].w, 7, fit(m.CurrentGrade, cols[2].w), "B", 0, "C", true, 0, "")
			pdf.CellFormat(cols[3].w, 7, formatAmount(m.Amount), "B", 0, "R", true, 0, "")

			pdf.SetFont("Helvetica", "B", 8.5)
			if m.Paid {
				pdf.SetTextColor(22, 101, 52)
			} else {
				pdf.SetTextColor(185, 28, 28)
			}
			pdf.CellFormat(cols[4].w, 7, status, "B", 0, "C", true, 0, "")

			pdf.SetFont("Helvetica", "", 8.5)
			pdf.SetTextColor(15, 23, 42)
			pdf.CellFormat(cols[5].w, 7, fit(paidBy, cols[5].w), "B", 0, "L", true, 0, "")
			pdf.CellFormat(cols[6].w, 7, paidAt, "B", 0, "L", true, 0, "")
			pdf.Ln(7)
		}
		pdf.Ln(6)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}