package usecase

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"my-fiber-app/domain"

	"github.com/go-pdf/fpdf"
)

const (
	pdfPageBottom = 195.0 // rows must end above this (A4 landscape is 210 mm high)
	pdfRowHeight  = 6.0
)

// Sri Lanka time (UTC+5:30), used for the "Generated" stamp in the footer
var reportZone = time.FixedZone("+0530", 5*3600+30*60)

type pdfCol struct {
	title string
	w     float64
	align string // "L", "R" or "C"
}

// pdfTable draws a table and repeats its header row on every new page
type pdfTable struct {
	pdf  *fpdf.Fpdf
	tr   func(string) string
	cols []pdfCol
}

func (t *pdfTable) header() {
	t.pdf.SetFont("Helvetica", "B", 8)
	t.pdf.SetFillColor(38, 70, 83)
	t.pdf.SetTextColor(255, 255, 255)
	t.pdf.SetDrawColor(38, 70, 83)
	for _, c := range t.cols {
		t.pdf.CellFormat(c.w, pdfRowHeight, c.title, "1", 0, "C", true, 0, "")
	}
	t.pdf.Ln(pdfRowHeight)
}

// fit translates the text to the PDF code page and shortens it with "..." if it is wider than the cell
func (t *pdfTable) fit(s string, w float64) string {
	s = t.tr(s) // one byte per character from here on
	max := w - 2
	if t.pdf.GetStringWidth(s) <= max {
		return s
	}
	for len(s) > 0 && t.pdf.GetStringWidth(s+"...") > max {
		s = s[:len(s)-1]
	}
	return s + "..."
}

func (t *pdfTable) row(vals []string, bold bool, shade bool) {
	if t.pdf.GetY()+pdfRowHeight > pdfPageBottom {
		t.pdf.AddPage()
		t.header()
	}
	style := ""
	if bold {
		style = "B"
	}
	t.pdf.SetFont("Helvetica", style, 8)
	t.pdf.SetTextColor(30, 30, 30)
	t.pdf.SetDrawColor(210, 210, 210)
	if shade {
		t.pdf.SetFillColor(236, 241, 244)
	} else {
		t.pdf.SetFillColor(255, 255, 255)
	}
	for i, c := range t.cols {
		t.pdf.CellFormat(c.w, pdfRowHeight, t.fit(vals[i], c.w), "1", 0, c.align, true, 0, "")
	}
	t.pdf.Ln(pdfRowHeight)
}

func groupDigits(digits string) string {
	var b strings.Builder
	for i := 0; i < len(digits); i++ {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(digits[i])
	}
	return b.String()
}

func formatInt(n int64) string {
	if n < 0 {
		return "-" + groupDigits(strconv.FormatInt(-n, 10))
	}
	return groupDigits(strconv.FormatInt(n, 10))
}

func formatMoney(v float64) string {
	s := strconv.FormatFloat(math.Abs(v), 'f', 2, 64)
	dot := strings.IndexByte(s, '.')
	out := groupDigits(s[:dot]) + s[dot:]
	if v < 0 {
		out = "-" + out
	}
	return out
}

// buildAnnualReportPDF renders the whole report (summary, month-wise, grade-wise and the card list)
// into an A4 landscape PDF. It builds the file in memory, which is fine for thousands of rows.
func buildAnnualReportPDF(report *domain.AnnualReport) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 12, 10)
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetTitle(fmt.Sprintf("Annual Payment Report %d", report.Year), false)
	pdf.SetCreator("Minisahana", false)
	pdf.AliasNbPages("{nb}")
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	generated := time.Now().In(reportZone).Format("2006-01-02 15:04")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-10)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(110, 110, 110)
		pdf.CellFormat(0, 5, "Generated "+generated+"  |  Page "+strconv.Itoa(pdf.PageNo())+" of {nb}", "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	section := func(title string, minSpace float64) {
		if pdf.GetY()+minSpace > pdfPageBottom {
			pdf.AddPage()
		}
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(38, 70, 83)
		pdf.CellFormat(0, 8, title, "", 1, "L", false, 0, "")
	}

	// Title
	s := report.Summary
	pdf.SetFont("Helvetica", "B", 16)
	pdf.SetTextColor(38, 70, 83)
	pdf.CellFormat(0, 9, fmt.Sprintf("Annual Payment Report - %d", report.Year), "", 1, "L", false, 0, "")
	scope := "All grades"
	if report.Grade != "" {
		scope = "Grade " + report.Grade
	}
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(0, 6, scope, "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Grand total
	section("Summary", 30)
	sum := &pdfTable{pdf: pdf, tr: tr, cols: []pdfCol{
		{"Report cards", 39.5, "C"}, {"Payments", 39.5, "C"}, {"Paid payments", 39.5, "C"}, {"Unpaid payments", 39.5, "C"},
		{"Total amount", 39.5, "C"}, {"Paid amount", 39.5, "C"}, {"Unpaid amount", 39.5, "C"},
	}}
	sum.header()
	sum.row([]string{
		formatInt(s.TotalCards), formatInt(s.Count), formatInt(s.PaidCount), formatInt(s.UnpaidCount),
		formatMoney(s.TotalAmount), formatMoney(s.PaidAmount), formatMoney(s.UnpaidAmount),
	}, true, true)

	if len(report.Data) == 0 {
		pdf.Ln(6)
		pdf.SetFont("Helvetica", "I", 10)
		pdf.SetTextColor(90, 90, 90)
		pdf.CellFormat(0, 6, "No report cards found for this period.", "", 1, "L", false, 0, "")
	} else {
		// Month-wise totals
		pdf.Ln(4)
		section("Month-wise totals", 40)
		mt := &pdfTable{pdf: pdf, tr: tr, cols: []pdfCol{
			{"Month", 40, "L"}, {"Payments", 30, "R"}, {"Paid", 30, "R"}, {"Unpaid", 30, "R"},
			{"Total amount", 49, "R"}, {"Paid amount", 49, "R"}, {"Unpaid amount", 49, "R"},
		}}
		mt.header()
		for i, m := range report.ByMonth {
			mt.row([]string{
				m.MonthLabel, formatInt(m.Count), formatInt(m.PaidCount), formatInt(m.UnpaidCount),
				formatMoney(m.TotalAmount), formatMoney(m.PaidAmount), formatMoney(m.UnpaidAmount),
			}, false, i%2 == 1)
		}
		mt.row([]string{
			"Total", formatInt(s.Count), formatInt(s.PaidCount), formatInt(s.UnpaidCount),
			formatMoney(s.TotalAmount), formatMoney(s.PaidAmount), formatMoney(s.UnpaidAmount),
		}, true, true)

		// Grade-wise totals
		pdf.Ln(4)
		section("Grade-wise totals", 40)
		gt := &pdfTable{pdf: pdf, tr: tr, cols: []pdfCol{
			{"Grade", 25, "L"}, {"Cards", 25, "R"}, {"Payments", 30, "R"}, {"Paid", 30, "R"}, {"Unpaid", 30, "R"},
			{"Total amount", 46, "R"}, {"Paid amount", 46, "R"}, {"Unpaid amount", 45, "R"},
		}}
		gt.header()
		for i, g := range report.ByGrade {
			gt.row([]string{
				g.Grade, formatInt(g.Cards), formatInt(g.Count), formatInt(g.PaidCount), formatInt(g.UnpaidCount),
				formatMoney(g.TotalAmount), formatMoney(g.PaidAmount), formatMoney(g.UnpaidAmount),
			}, false, i%2 == 1)
		}
		gt.row([]string{
			"Total", formatInt(s.TotalCards), formatInt(s.Count), formatInt(s.PaidCount), formatInt(s.UnpaidCount),
			formatMoney(s.TotalAmount), formatMoney(s.PaidAmount), formatMoney(s.UnpaidAmount),
		}, true, true)

		// Report card list (starts on its own page; the header repeats on every page)
		pdf.AddPage()
		section("Report cards", 30)
		ct := &pdfTable{pdf: pdf, tr: tr, cols: []pdfCol{
			{"#", 10, "R"}, {"Ref No", 24, "L"}, {"Full name", 56, "L"}, {"NIC", 26, "L"}, {"Grade", 12, "C"},
			{"Months", 14, "R"}, {"Paid", 13, "R"}, {"Unpaid", 14, "R"}, {"Monthly amt", 24, "R"},
			{"Year total", 28, "R"}, {"Paid amount", 28, "R"}, {"Unpaid amount", 28, "R"},
		}}
		ct.header()
		for i, r := range report.Data {
			ct.row([]string{
				strconv.Itoa(i + 1), r.RefNumber, r.FullName, r.NIC, r.CurrentGrade,
				strconv.Itoa(r.MonthsInYear), strconv.Itoa(r.PaidMonths), strconv.Itoa(r.UnpaidMonths),
				formatMoney(r.Amount), formatMoney(r.YearAmount), formatMoney(r.PaidAmount), formatMoney(r.UnpaidAmount),
			}, false, i%2 == 1)
		}
		ct.row([]string{
			"", "", fmt.Sprintf("Grand total (%s cards)", formatInt(s.TotalCards)), "", "",
			formatInt(s.Count), formatInt(s.PaidCount), formatInt(s.UnpaidCount),
			"", formatMoney(s.TotalAmount), formatMoney(s.PaidAmount), formatMoney(s.UnpaidAmount),
		}, true, true)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}