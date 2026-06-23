package main

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	BG_SHEET_NAME = "I. Budget"
	BG_TAB_COLOR  = "5B9BD5" // Medium Blue

	BG_COL_LABEL  = 2
	BG_COL_ID     = 3
	BG_COL_POS    = 4
	BG_COL_LC     = 5
	BG_COL_Y1     = 6
	BG_COL_Y2     = 7
	BG_COL_Y3     = 8
	BG_COL_EUR    = 9
	BG_COL_GAP    = 10
	BG_COL_STATUS = 11
	BG_COL_CHECK  = 12
	BG_COL_BEGR_1 = 11
	BG_COL_BEGR_2 = 13

	BG_HELP_LC  = 15
	BG_HELP_EUR = 16

	BG_COL_LIST_GEBER = 18
	BG_COL_LIST_ID    = 19

	BG_NAME_GEBER_LIST = "Geber_Liste"
	BG_NAME_ID_LIST    = "Budget_ID_Liste"

	BG_TABLE_NAME = "TblDrittmittel"
	BG_TABLE_AUSG = "TblBudgetAusgaben"

	BG_W_LABEL  = 25.0
	BG_W_ID     = 8.0
	BG_W_POS    = 30.0
	BG_W_LC     = 18.0
	BG_W_YEAR   = 14.0
	BG_W_EUR    = 18.0
	BG_W_GAP    = 3.0
	BG_W_STATUS = 25.0
	BG_W_CHECK  = 14.0
	BG_W_BEGR   = 16.0

	BG_FMT_LC   = "#,##0.00"
	BG_FMT_EUR  = `#,##0.00" €"`
	BG_FMT_RATE = "0.0000"

	BG_CLR_HEADER     = "D3D3D3"
	BG_CLR_SUBHEAD    = "F0F0F0"
	BG_CLR_INPUT      = "FFFAE5"
	BG_CLR_BORDER     = "808080"
	BG_CLR_GRID       = "D3D3D3"
	BG_CLR_FONT       = "3C3C3C"
	BG_CLR_BLACK      = "000000"
	BG_CLR_RES_OFF    = "F2F2F2"
	BG_CLR_RES_TXT    = "595959"
	BG_CLR_RES_ON     = "C6EFCE"
	BG_CLR_RES_ON_TXT = "006100"
	BG_CLR_BAD        = "FFC7CE"
	BG_CLR_BAD_TXT    = "9C0006"

	BG_NAME_RESERVE    = "Reserve_Freigabe"
	BG_NAME_KURS       = "Budget_Kurs"
	BG_NAME_EIGEN_LW   = "Eigenmittel_LW"
	BG_NAME_EIGEN_EUR  = "Eigenmittel_EUR"
	BG_NAME_DRITT_LW   = "Drittmittel_LW"
	BG_NAME_DRITT_EUR  = "Drittmittel_EUR"
	BG_NAME_KMW_LW     = "KMW_Mittel_LW"
	BG_NAME_KMW_EUR    = "KMW_Mittel_EUR"
	BG_NAME_GESAMT_LW  = "Gesamtprojektmittel_LW"
	BG_NAME_GESAMT_EUR = "Gesamtprojektmittel_EUR"
	BG_NAME_AUSG_LW    = "Gesamtausgaben_LW"
	BG_NAME_AUSG_EUR   = "Gesamtausgaben_EUR"
)

var BG_CATEGORIES = []string{
	"Bauausgaben", "Investitionen", "Personalkosten", "Projektaktivitaeten",
	"Projektverwaltung", "Evaluierung", "Audit", "Reserve",
}

var BG_YEARS = []string{"Jahr 1", "Jahr 2", "Jahr 3"}

func bgKostenName(cat string, cur string) string {
	return fmt.Sprintf("Kosten_%s_%s", cat, cur)
}

func falsePtr() *bool {
	b := false
	return &b
}

func (g *Generator) CreateBudgetSheet() error {
	ws := BG_SHEET_NAME
	f := g.file

	_, _ = f.NewSheet(ws)
	tabColor := BG_TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	g.setColWidth(ws, BG_COL_LABEL, BG_W_LABEL)
	g.setColWidth(ws, BG_COL_ID, BG_W_ID)
	g.setColWidth(ws, BG_COL_POS, BG_W_POS)
	g.setColWidth(ws, BG_COL_LC, BG_W_LC)
	g.setColWidth(ws, BG_COL_Y1, BG_W_YEAR)
	g.setColWidth(ws, BG_COL_Y2, BG_W_YEAR)
	g.setColWidth(ws, BG_COL_Y3, BG_W_YEAR)
	g.setColWidth(ws, BG_COL_EUR, BG_W_EUR)
	g.setColWidth(ws, BG_COL_GAP, BG_W_GAP)
	g.setColWidth(ws, BG_COL_STATUS, BG_W_STATUS)
	g.setColWidth(ws, BG_COL_CHECK, BG_W_CHECK)
	g.setColWidth(ws, BG_COL_BEGR_2, BG_W_BEGR)

	r := 2

	// Title
	titleOpts := StyleOptions{Size: 14, Bold: true, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_HEADER, VAlign: "center", BorderTop: 2, BorderBottom: 2, BorderColor: BG_CLR_BORDER}
	for c := BG_COL_LABEL; c <= BG_COL_EUR; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), titleOpts)
	}
	g.setValue(ws, cellName(BG_COL_LABEL, r), "I. KERNDATEN BUDGET", titleOpts)
	_ = f.SetRowHeight(ws, r, 24)
	r += 2

	// Section 1
	g.bgSectionHeader(ws, r, "1. GEPLANTE EINNAHMEN / FINANZIERUNG")

	// Budget-Kurs (G4 und H4)
	g.setValue(ws, cellName(BG_COL_Y2, r), "€ Budget-Kurs:", StyleOptions{
		Size:         9,
		HAlign:       "right",
		VAlign:       "center",
		FillColor:    BG_CLR_HEADER,
		BorderTop:    2,
		BorderBottom: 1,
		BorderColor:  BG_CLR_BORDER,
	})
	rateCellOpts := StyleOptions{
		NumFormat:    BG_FMT_RATE,
		Italic:       true,
		BorderTop:    2, // Durchgehender dicker Rahmen oben
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  BG_CLR_BORDER,
	}
	g.setStyle(ws, cellName(BG_COL_Y3, r), cellName(BG_COL_Y3, r), rateCellOpts)
	g.upsertNamedRange(BG_NAME_KURS, BG_COL_Y3, r)
	r += 1

	g.setValue(ws, cellName(BG_COL_LABEL, r), "Finanzierungsquelle", StyleOptions{})
	g.bgValueHeaderCells(ws, r)
	g.bgTableHeader(ws, r, BG_COL_LABEL, BG_COL_EUR)
	r += 1

	// 1.1 Eigenmittel
	g.bgSubHeader(ws, r, "1.1 Eigenmittel")
	r += 1
	eigenRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, r), "Eigenmittel", StyleOptions{Size: 10})
	g.bgYearRow(ws, r)
	g.upsertNamedRange(BG_NAME_EIGEN_LW, BG_COL_LC, r)
	g.upsertNamedRange(BG_NAME_EIGEN_EUR, BG_COL_EUR, r)
	_ = f.SetRowHeight(ws, r, 22)
	r += 2

	// 1.2 Drittmittel
	g.bgSubHeader(ws, r, "1.2 Drittmittel")
	r += 1
	drittSummeRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, r), "Drittmittel (Summe):", StyleOptions{Size: 10})
	g.setValue(ws, cellName(BG_COL_POS, r), "Aufstellung je Geber → Tabelle rechts", StyleOptions{Size: 8, Italic: true, FontColor: BG_CLR_RES_TXT})
	g.bgSummeCell(ws, r, BG_COL_LC, fmt.Sprintf(`=SUM(%s[Betrag (LC)])`, BG_TABLE_NAME), BG_FMT_LC)
	g.bgSummeCell(ws, r, BG_COL_EUR, fmt.Sprintf(`=SUM(%s[Betrag (EUR)])`, BG_TABLE_NAME), BG_FMT_EUR)
	for _, c := range []int{BG_COL_Y1, BG_COL_Y2, BG_COL_Y3} {
		g.bgInput(ws, cellName(c, r), BG_FMT_LC)
	}
	g.upsertNamedRange(BG_NAME_DRITT_LW, BG_COL_LC, r)
	g.upsertNamedRange(BG_NAME_DRITT_EUR, BG_COL_EUR, r)
	_ = f.SetRowHeight(ws, r, 22)
	r += 2

	// 1.3 KMW-Mittel
	g.bgSubHeader(ws, r, "1.3 KMW-Mittel")
	r += 1
	kmwRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, r), "KMW-Mittel", StyleOptions{Size: 10})
	g.bgYearRow(ws, r)
	g.upsertNamedRange(BG_NAME_KMW_LW, BG_COL_LC, r)
	g.upsertNamedRange(BG_NAME_KMW_EUR, BG_COL_EUR, r)
	_ = f.SetRowHeight(ws, r, 22)
	r += 2

	// GESAMTPROJEKTMITTEL
	gesamtRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, r), "GESAMTPROJEKTMITTEL", StyleOptions{})
	sumOf := func(col int) string {
		return fmt.Sprintf("=%s+%s+%s", cellName(col, eigenRow), cellName(col, drittSummeRow), cellName(col, kmwRow))
	}
	g.setFormula(ws, cellName(BG_COL_LC, r), sumOf(BG_COL_LC), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y1, r), sumOf(BG_COL_Y1), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y2, r), sumOf(BG_COL_Y2), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y3, r), sumOf(BG_COL_Y3), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_EUR, r), sumOf(BG_COL_EUR), StyleOptions{NumFormat: BG_FMT_EUR})
	g.bgTotalRow(ws, r, BG_COL_LABEL, BG_COL_EUR)

	totalLoc := absName(BG_COL_LC, r)
	totalEur := absName(BG_COL_EUR, r)
	g.upsertNamedRange(BG_NAME_GESAMT_LW, BG_COL_LC, r)
	g.upsertNamedRange(BG_NAME_GESAMT_EUR, BG_COL_EUR, r)

	g.setFormula(ws, cellName(BG_COL_Y3, 4), fmt.Sprintf(`=IFERROR(%s/%s,0)`, totalLoc, totalEur), rateCellOpts)
	r += 2

	// Section 2: Ausgaben
	g.bgSectionHeader(ws, r, "2. GEPLANTE AUSGABEN")
	r += 1

	ausgHdrRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, r), "Kostenkategorie", StyleOptions{})
	g.setValue(ws, cellName(BG_COL_ID, r), "ID", StyleOptions{})
	g.setValue(ws, cellName(BG_COL_POS, r), "Kostenposition", StyleOptions{})
	g.bgValueHeaderCells(ws, r)
	r += 1

	ausgDataRows := len(BG_CATEGORIES)
	catArrayStr := `{"` + strings.Join(BG_CATEGORIES, `","`) + `"}`
	for i := 0; i < ausgDataRows; i++ {
		row := r + i
		g.setValue(ws, cellName(BG_COL_LABEL, row), BG_CATEGORIES[i], StyleOptions{FillColor: BG_CLR_INPUT, HAlign: "left", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

		formulaID := fmt.Sprintf(`=IF(B%d="","",MATCH(B%d,%s,0)&"."&COUNTIF(B$%d:B%d,B%d))`, row, row, catArrayStr, r, row, row)
		g.setFormula(ws, cellName(BG_COL_ID, row), formulaID, StyleOptions{FillColor: BG_CLR_INPUT, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

		g.setValue(ws, cellName(BG_COL_POS, row), "", StyleOptions{FillColor: BG_CLR_INPUT, HAlign: "left", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
		g.bgInput(ws, cellName(BG_COL_LC, row), BG_FMT_LC)
		g.bgInput(ws, cellName(BG_COL_Y1, row), BG_FMT_LC)
		g.bgInput(ws, cellName(BG_COL_Y2, row), BG_FMT_LC)
		g.bgInput(ws, cellName(BG_COL_Y3, row), BG_FMT_LC)
		g.bgInput(ws, cellName(BG_COL_EUR, row), BG_FMT_EUR)
	}

	g.bgTableHeader(ws, ausgHdrRow, BG_COL_LABEL, BG_COL_EUR)

	dv := excelize.NewDataValidation(true)
	dv.Sqref = fmt.Sprintf("%s:%s", cellName(BG_COL_LABEL, r), cellName(BG_COL_LABEL, r+ausgDataRows-1))
	dv.SetDropList(BG_CATEGORIES)
	_ = f.AddDataValidation(ws, dv)

	r += ausgDataRows
	ausgTotalsRow := r
	g.setValue(ws, cellName(BG_COL_LABEL, ausgTotalsRow), "Geplante Gesamtausgaben", StyleOptions{})
	g.setFormula(ws, cellName(BG_COL_LC, ausgTotalsRow), fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (LC)])`, BG_TABLE_AUSG), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y1, ausgTotalsRow), fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BG_TABLE_AUSG, BG_YEARS[0]), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y2, ausgTotalsRow), fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BG_TABLE_AUSG, BG_YEARS[1]), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_Y3, ausgTotalsRow), fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BG_TABLE_AUSG, BG_YEARS[2]), StyleOptions{NumFormat: BG_FMT_LC})
	g.setFormula(ws, cellName(BG_COL_EUR, ausgTotalsRow), fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (EUR)])`, BG_TABLE_AUSG), StyleOptions{NumFormat: BG_FMT_EUR})

	g.bgTotalRow(ws, ausgTotalsRow, BG_COL_LABEL, BG_COL_EUR)

	_ = f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(BG_COL_LABEL, ausgHdrRow), cellName(BG_COL_EUR, ausgTotalsRow-1)),
		Name:           BG_TABLE_AUSG,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})

	ausgLastRow := ausgTotalsRow

	reserveEurAddr := ""
	for i, cat := range BG_CATEGORIES {
		hr := 4 + i
		g.setFormula(ws, cellName(BG_HELP_LC, hr), fmt.Sprintf(`=SUMIFS(%s[Betrag (LC)],%s[Kostenkategorie],"%s")`, BG_TABLE_AUSG, BG_TABLE_AUSG, cat), StyleOptions{})
		g.setFormula(ws, cellName(BG_HELP_EUR, hr), fmt.Sprintf(`=SUMIFS(%s[Betrag (EUR)],%s[Kostenkategorie],"%s")`, BG_TABLE_AUSG, BG_TABLE_AUSG, cat), StyleOptions{})
		g.upsertNamedRange(bgKostenName(cat, "LW"), BG_HELP_LC, hr)
		g.upsertNamedRange(bgKostenName(cat, "EUR"), BG_HELP_EUR, hr)
		if cat == "Reserve" {
			reserveEurAddr = absName(BG_HELP_EUR, hr)
		}
	}

	gesHr := 4 + len(BG_CATEGORIES)
	g.setFormula(ws, cellName(BG_HELP_LC, gesHr), fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (LC)])`, BG_TABLE_AUSG), StyleOptions{})
	g.setFormula(ws, cellName(BG_HELP_EUR, gesHr), fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (EUR)])`, BG_TABLE_AUSG), StyleOptions{})
	g.upsertNamedRange(BG_NAME_AUSG_LW, BG_HELP_LC, gesHr)
	g.upsertNamedRange(BG_NAME_AUSG_EUR, BG_HELP_EUR, gesHr)
	expLocAddr := absName(BG_HELP_LC, gesHr)
	expEurAddr := absName(BG_HELP_EUR, gesHr)

	_ = f.SetColVisible(ws, colLetter(BG_HELP_LC), false)
	_ = f.SetColVisible(ws, colLetter(BG_HELP_EUR), false)
	_ = f.SetColVisible(ws, colLetter(BG_COL_LIST_GEBER), false)
	_ = f.SetColVisible(ws, colLetter(BG_COL_LIST_ID), false)

	g.bgBuildLookupLists(ws)
	g.bgDrawDrittmittelTable(ws, ausgHdrRow)

	g.styleOuterBorder(ws, 2, BG_COL_LABEL, ausgLastRow, BG_COL_EUR, 2, BG_CLR_BORDER)

	reserveCheckAddr := g.bgDrawReserveBox(ws, reserveEurAddr)
	g.bgDrawBegruendung(ws, reserveCheckAddr)

	incYearsAddr := fmt.Sprintf("%s+%s+%s", absName(BG_COL_Y1, gesamtRow), absName(BG_COL_Y2, gesamtRow), absName(BG_COL_Y3, gesamtRow))
	expYearsAddr := fmt.Sprintf("%s+%s+%s", absName(BG_COL_Y1, ausgLastRow), absName(BG_COL_Y2, ausgLastRow), absName(BG_COL_Y3, ausgLastRow))
	rateCellAddr := absName(BG_COL_Y3, 4)
	g.bgDrawChecks(ws, ausgLastRow+2, totalLoc, totalEur, incYearsAddr, expLocAddr, expEurAddr, expYearsAddr, rateCellAddr)

	return nil
}

func (g *Generator) bgDrawDrittmittelTable(ws string, ausgHdrRow int) {
	cName, cLc, cEur := BG_COL_STATUS, BG_COL_CHECK, BG_COL_BEGR_2
	titleRow := ausgHdrRow - 1
	headerRow := ausgHdrRow
	dataRows := 3

	g.mergeCells(ws, cellName(cName, titleRow), cellName(cEur, titleRow), "Drittmittel – Aufstellung je Geber", StyleOptions{
		Bold: true, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_HEADER, HAlign: "center", VAlign: "center",
	})
	g.setValue(ws, cellName(cName, headerRow), "Name des Gebers", StyleOptions{})
	g.setValue(ws, cellName(cLc, headerRow), "Betrag (LC)", StyleOptions{})
	g.setValue(ws, cellName(cEur, headerRow), "Betrag (EUR)", StyleOptions{})

	for i := 0; i < dataRows; i++ {
		row := headerRow + 1 + i
		g.setValue(ws, cellName(cName, row), "", StyleOptions{FillColor: BG_CLR_INPUT, HAlign: "left", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
		g.bgInput(ws, cellName(cLc, row), BG_FMT_LC)
		g.bgInput(ws, cellName(cEur, row), BG_FMT_EUR)
	}

	g.bgTableHeader(ws, headerRow, cName, cEur)
	_ = g.file.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(cName, headerRow), cellName(cEur, headerRow+dataRows)),
		Name:           BG_TABLE_NAME,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})
	g.styleOuterBorder(ws, titleRow, cName, headerRow+dataRows, cEur, 2, BG_CLR_BORDER)
}

func (g *Generator) bgSectionHeader(ws string, r int, title string) {
	opts := StyleOptions{
		Bold: true, Size: 11, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_HEADER, HAlign: "left", VAlign: "center", BorderTop: 2, BorderBottom: 1, BorderColor: BG_CLR_BORDER,
	}
	for c := BG_COL_LABEL; c <= BG_COL_EUR; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), opts)
	}
	g.setValue(ws, cellName(BG_COL_LABEL, r), title, opts)
	_ = g.file.SetRowHeight(ws, r, 24)
}

func (g *Generator) bgSubHeader(ws string, r int, title string) {
	opts := StyleOptions{
		Bold: true, Size: 10, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_SUBHEAD, HAlign: "left", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_BORDER,
	}
	for c := BG_COL_LABEL; c <= BG_COL_EUR; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), opts)
	}
	g.setValue(ws, cellName(BG_COL_LABEL, r), title, opts)
	_ = g.file.SetRowHeight(ws, r, 20)
}

func (g *Generator) bgValueHeaderCells(ws string, r int) {
	g.setValue(ws, cellName(BG_COL_LC, r), "Betrag (LC)", StyleOptions{})
	g.setValue(ws, cellName(BG_COL_Y1, r), BG_YEARS[0], StyleOptions{})
	g.setValue(ws, cellName(BG_COL_Y2, r), BG_YEARS[1], StyleOptions{})
	g.setValue(ws, cellName(BG_COL_Y3, r), BG_YEARS[2], StyleOptions{})
	g.setValue(ws, cellName(BG_COL_EUR, r), "Betrag (EUR)", StyleOptions{})
}

func (g *Generator) bgTableHeader(ws string, r int, c1 int, c2 int) {
	opts := StyleOptions{
		Bold: true, Size: 9, FontColor: BG_CLR_FONT, FillColor: BG_CLR_HEADER, HAlign: "center", VAlign: "center", BorderBottom: 2, BorderColor: BG_CLR_BORDER,
	}
	for c := c1; c <= c2; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), opts)
	}
}

func (g *Generator) bgYearRow(ws string, r int) {
	for _, c := range []int{BG_COL_LC, BG_COL_Y1, BG_COL_Y2, BG_COL_Y3} {
		g.bgInput(ws, cellName(c, r), BG_FMT_LC)
	}
	g.bgInput(ws, cellName(BG_COL_EUR, r), BG_FMT_EUR)
}

func (g *Generator) bgSummeCell(ws string, r int, c int, formula string, fmtStr string) {
	g.setFormula(ws, cellName(c, r), formula, StyleOptions{
		Bold: true, HAlign: "right", VAlign: "center", NumFormat: fmtStr,
	})
}

func (g *Generator) bgInput(ws string, cell string, numFmt string) {
	g.setStyle(ws, cell, cell, StyleOptions{
		FillColor: BG_CLR_INPUT, HAlign: "right", VAlign: "center", NumFormat: numFmt, BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID,
	})
}

func (g *Generator) bgTotalRow(ws string, r int, c1 int, c2 int) {
	opts := StyleOptions{
		Bold: true, Size: 10, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_SUBHEAD, VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BG_CLR_BORDER,
	}
	for c := c1; c <= c2; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), opts)
	}
	_ = g.file.SetRowHeight(ws, r, 20)
}

func (g *Generator) bgBuildLookupLists(ws string) {
	g.setDynArrayFormula(ws, cellName(BG_COL_LIST_GEBER, 1), fmt.Sprintf(`_xlfn.VSTACK({"Projektpartner";"Bank"},IFERROR(_xlfn._xlws.FILTER(%s[Name des Gebers],%s[Name des Gebers]<>""),""))`, BG_TABLE_NAME, BG_TABLE_NAME), StyleOptions{})
	g.upsertNamedFormula(BG_NAME_GEBER_LIST, fmt.Sprintf("OFFSET('%s'!%s, 0, 0, COUNTA('%s'!%s:%s), 1)", ws, absName(BG_COL_LIST_GEBER, 1), ws, colLetter(BG_COL_LIST_GEBER), colLetter(BG_COL_LIST_GEBER)))

	g.setDynArrayFormula(ws, cellName(BG_COL_LIST_ID, 1), fmt.Sprintf(`IFERROR(_xlfn._xlws.FILTER(%s[ID],%s[ID]<>""),"")`, BG_TABLE_AUSG, BG_TABLE_AUSG), StyleOptions{})
	g.upsertNamedFormula(BG_NAME_ID_LIST, fmt.Sprintf("OFFSET('%s'!%s, 0, 0, COUNTA('%s'!%s:%s), 1)", ws, absName(BG_COL_LIST_ID, 1), ws, colLetter(BG_COL_LIST_ID), colLetter(BG_COL_LIST_ID)))
}

func (g *Generator) bgDrawReserveBox(ws string, reserveEurAddr string) string {
	c, col := BG_COL_STATUS, BG_COL_STATUS
	rHead, rAmount, rCapt, rCheck, rStatus := 2, 3, 4, 5, 6

	g.setValue(ws, cellName(col, rHead), "Reserve Freigabe", StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_HEADER, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

	if reserveEurAddr != "" {
		g.setFormula(ws, cellName(col, rAmount), fmt.Sprintf("=%s", reserveEurAddr), StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_FONT, HAlign: "center", VAlign: "center", NumFormat: BG_FMT_EUR, BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
	} else {
		g.setValue(ws, cellName(col, rAmount), 0, StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_FONT, HAlign: "center", VAlign: "center", NumFormat: BG_FMT_EUR, BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
	}

	g.setValue(ws, cellName(col, rCapt), "Reserve freigeben:", StyleOptions{Size: 9, FontColor: BG_CLR_RES_TXT, Italic: true, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

	g.setValue(ws, cellName(col, rCheck), "Nein", StyleOptions{FillColor: BG_CLR_INPUT, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
	checkAddr := absName(c, rCheck)

	// Dropdown-Auswahl Ja/Nein hinzufügen
	dv := excelize.NewDataValidation(true)
	dv.Sqref = cellName(col, rCheck)
	dv.SetDropList([]string{"Ja", "Nein"})
	_ = g.file.AddDataValidation(ws, dv)

	statusFormula := fmt.Sprintf(`=IF(%s="Ja","FREIGEGEBEN","NICHT FREIGEGEBEN")`, checkAddr)
	statusStyleId, _ := g.getOrCreateStyle(StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_RES_TXT, FillColor: BG_CLR_RES_OFF, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
	g.file.SetCellFormula(ws, cellName(col, rStatus), statusFormula)
	g.file.SetCellStyle(ws, cellName(col, rStatus), cellName(col, rStatus), statusStyleId)

	onStyleOpts := StyleOptions{
		Bold:         true,
		Size:         9,
		FontColor:    BG_CLR_RES_ON_TXT,
		FillColor:    BG_CLR_RES_ON,
		BorderLeft:   2,
		BorderRight:  2,
		BorderTop:    1,
		BorderBottom: 2,
		BorderColor:  BG_CLR_BORDER,
	}
	g.addConditionalFormat(ws, cellName(col, rStatus), fmt.Sprintf(`%s="Ja"`, checkAddr), onStyleOpts)

	g.styleOuterBorder(ws, rHead, col, rStatus, col, 2, BG_CLR_BORDER)
	g.upsertNamedRange(BG_NAME_RESERVE, c, rCheck)
	return checkAddr
}

func (g *Generator) bgDrawBegruendung(ws string, reserveCheckAddr string) {
	c1, c2 := BG_COL_BEGR_1, BG_COL_BEGR_2
	hdrRow, areaTop, areaRows := 8, 9, 4

	// Inaktiver Standard-Stil: Grau hinterlegt, grauer Text, dünner grauer Rahmen
	g.mergeCells(ws, cellName(c1, hdrRow), cellName(c2, hdrRow), "Begruendung", StyleOptions{Bold: true, Size: 9, FontColor: "7F7F7F", FillColor: "F2F2F2", HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: "D3D3D3"})
	g.mergeCells(ws, cellName(c1, areaTop), cellName(c2, areaTop+areaRows-1), "", StyleOptions{FillColor: "F2F2F2", HAlign: "left", VAlign: "top", WrapText: true, BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: "D3D3D3"})

	condFormula := fmt.Sprintf(`%s="Ja"`, reserveCheckAddr)

	// Aktiver Stil für Header bei TRUE: Dunkelgrauer Header mit weißem Text und deutlichem Rahmen
	styleActiveHeaderOpts := StyleOptions{Bold: true, Size: 9, FontColor: "FFFFFF", FillColor: "3C3C3C", HAlign: "center", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BG_CLR_BORDER}
	g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(c1, hdrRow), cellName(c2, hdrRow)), condFormula, styleActiveHeaderOpts)

	// Aktiver Stil für Eingabebereich bei TRUE: Zartgelb (BG_CLR_INPUT) mit deutlichem Rahmen
	styleActiveAreaOpts := StyleOptions{FillColor: BG_CLR_INPUT, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BG_CLR_BORDER}
	g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(c1, areaTop), cellName(c2, areaTop+areaRows-1)), condFormula, styleActiveAreaOpts)
}

func (g *Generator) bgDrawChecks(ws string, top int, incLocAddr, incEurAddr, incYearsAddr, expLocAddr, expEurAddr, expYearsAddr, rateCellAddr string) {
	cLbl, cVal := BG_COL_LABEL, BG_COL_LC

	g.mergeCells(ws, cellName(cLbl, top), cellName(cVal, top), "Budgetpruefung", StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_BLACK, FillColor: BG_CLR_HEADER, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

	checks := []struct{ lbl, fml string }{
		{"Einnahmen = Ausgaben (LC)", fmt.Sprintf(`=IF(ROUND(%s-%s,2)=0,"OK","Abweichung")`, incLocAddr, expLocAddr)},
		{"Einnahmen = Ausgaben (EUR)", fmt.Sprintf(`=IF(ROUND(%s-%s,2)=0,"OK","Abweichung")`, incEurAddr, expEurAddr)},
		{"Gleicher Budget-Kurs", fmt.Sprintf(`=IF(ROUND(%s,4)=ROUND(IFERROR(%s/%s,0),4),"OK","Abweichung")`, rateCellAddr, expLocAddr, expEurAddr)},
		{"Einnahmen: Jahre = Gesamt (LC)", fmt.Sprintf(`=IF(ROUND((%s)-%s,2)=0,"OK","Abweichung")`, incYearsAddr, incLocAddr)},
		{"Ausgaben: Jahre = Gesamt (LC)", fmt.Sprintf(`=IF(ROUND((%s)-%s,2)=0,"OK","Abweichung")`, expYearsAddr, expLocAddr)},
	}

	for i, chk := range checks {
		rr := top + 1 + i
		g.mergeCells(ws, cellName(cLbl, rr), cellName(BG_COL_POS, rr), chk.lbl, StyleOptions{Size: 9, FontColor: BG_CLR_RES_TXT, HAlign: "left", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})

		valCell := cellName(cVal, rr)
		valStyleId, _ := g.getOrCreateStyle(StyleOptions{Bold: true, Size: 9, FontColor: BG_CLR_RES_TXT, FillColor: BG_CLR_RES_OFF, HAlign: "center", VAlign: "center", BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BG_CLR_GRID})
		g.file.SetCellFormula(ws, valCell, chk.fml)
		g.file.SetCellStyle(ws, valCell, valCell, valStyleId)

		valAddr := absName(cVal, rr)
		bBottom := 1
		if i == len(checks)-1 {
			bBottom = 2
		}

		okStyleOpts := StyleOptions{
			Bold:         true,
			Size:         9,
			FontColor:    BG_CLR_RES_ON_TXT,
			FillColor:    BG_CLR_RES_ON,
			BorderLeft:   1,
			BorderTop:    1,
			BorderRight:  2,
			BorderBottom: bBottom,
			BorderColor:  BG_CLR_BORDER,
		}
		g.addConditionalFormat(ws, valCell, fmt.Sprintf(`%s="OK"`, valAddr), okStyleOpts)

		badStyleOpts := StyleOptions{
			Bold:         true,
			Size:         9,
			FontColor:    BG_CLR_BAD_TXT,
			FillColor:    BG_CLR_BAD,
			BorderLeft:   1,
			BorderTop:    1,
			BorderRight:  2,
			BorderBottom: bBottom,
			BorderColor:  BG_CLR_BORDER,
		}
		g.addConditionalFormat(ws, valCell, fmt.Sprintf(`%s<>"OK"`, valAddr), badStyleOpts)
	}

	g.styleOuterBorder(ws, top, cLbl, top+len(checks), cVal, 2, BG_CLR_BORDER)
}
