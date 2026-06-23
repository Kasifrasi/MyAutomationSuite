package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ==================================================================================
// Blatt "V. AUSWERTUNG"
//
// Erzeugt zwei Prüf-Sektionen:
//   A) MITTELANFORDERUNGSPRÜFUNG (Basis: aktuelle Mittelanforderung / Prognose)
//   B) FINANZBERICHTSPRÜFUNG     (Basis: aktuelle Finanzberichte / kumulativ)
//
// Anders als die Office-Script-Vorlage (ExcelScript, Laufzeit im Excel) schreibt
// dieser Generator die Formeln vorab. Es werden ausschließlich benannte Bereiche
// und die bereits aufbereiteten Spill-Stacks des "Daten"-Blatts verwendet – keine
// Volltext-Spaltenbezüge (B:B) und kein 36×INDIRECT-Konstrukt.
// ==================================================================================

const (
	EVAL_SHEET_NAME   = "V. AUSWERTUNG"
	EVAL_TAB_COLOR    = "FFFF00" // Gelb
	EVAL_DATEN_SHEET  = "Daten"
	EVAL_STACK_MAXROW = 500 // großzügige Obergrenze für die Spill-Stacks

	// Spalten der Vergleichstabellen (B … J)
	EV_COL_LABEL   = 2  // B Kategorie
	EV_COL_ACT_LC  = 3  // C Prognose/Kumulativ (LC)
	EV_COL_BUD_LC  = 4  // D Budget (LC)
	EV_COL_DIF_LC  = 5  // E Differenz (LC)
	EV_COL_ABW_LC  = 6  // F Abweichung (LC)
	EV_COL_ACT_EUR = 7  // G Prognose/Kumulativ (EUR)
	EV_COL_BUD_EUR = 8  // H Budget (EUR)
	EV_COL_DIF_EUR = 9  // I Differenz (EUR)
	EV_COL_ABW_EUR = 10 // J Abweichung (EUR)

	EV_TABLE_GAP = 2

	// Spaltenbuchstaben der "Daten"-Stacks
	EV_DTN_EIN1_TYP = "C" // Einnahmen explizit – Typ
	EV_DTN_EIN1_LC  = "E" // Einnahmen explizit – LC
	EV_DTN_EIN1_EUR = "F" // Einnahmen explizit – EUR
	EV_DTN_EIN2_TYP = "I" // Einnahmen Durchschnittskurs – Typ
	EV_DTN_EIN2_LC  = "K" // Einnahmen Durchschnittskurs – LC
	EV_DTN_EIN2_EUR = "L" // Einnahmen Durchschnittskurs – EUR
	EV_DTN_AUS_ID   = "O" // Ausgaben (FB) – ID
	EV_DTN_AUS_LC   = "P" // Ausgaben (FB) – LC
	EV_DTN_AUS_EUR  = "Q" // Ausgaben (FB) – EUR
	EV_DTN_MA_CAT   = "U" // Mittelanforderung – Kostenkategorie
	EV_DTN_MA_LC    = "V" // Mittelanforderung – LC
	EV_DTN_MA_EUR   = "W" // Mittelanforderung – EUR

	// Farben
	EV_CLR_BANNER     = "212F3D" // Dark Slate
	EV_CLR_BANNER_TXT = "FFFFFF"
	EV_CLR_BANNER_SUB = "B4BEC8"
	EV_CLR_HEADER     = "D3D3D3"
	EV_CLR_TOTAL      = "212F3D"
	EV_CLR_TOTAL_TXT  = "FFFFFF"
	EV_CLR_INPUT      = "FFFAE5"
	EV_CLR_CALC       = "F2F2F2"
	EV_CLR_BORDER     = "808080"
	EV_CLR_GRID       = "D3D3D3"
	EV_CLR_BLACK      = "000000"
	EV_CLR_GOOD       = "C6EFCE"
	EV_CLR_GOOD_TXT   = "006100"
	EV_CLR_BAD        = "FFC7CE"
	EV_CLR_BAD_TXT    = "9C0006"
	EV_CLR_WARN       = "FCF3CF"
	EV_CLR_WARN_TXT   = "9C640C"

	EV_FMT_LC  = "#,##0.00"
	EV_FMT_EUR = `#,##0.00" €"`
	EV_FMT_PCT = "0.0%"
)

// evalCompResult bündelt die Adressen, die nach dem Bau einer Vergleichstabelle
// noch von außen gebraucht werden (z. B. für die Mehreinnahmen-Rückkopplung).
type evalCompResult struct {
	nextRow     int
	actEURRange string // z. B. $G$30:$G$33
	budEURRange string // z. B. $H$30:$H$33
	kmwActEUR   string // Zelle der KMW-Mittel-Ist (EUR)
	kmwBudEUR   string // Zelle des KMW-Mittel-Budgets (EUR)
}

// CreateAuswertungSheet baut das Blatt "V. AUSWERTUNG".
func (g *Generator) CreateAuswertungSheet() error {
	ws := EVAL_SHEET_NAME
	f := g.file

	if _, err := f.NewSheet(ws); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Auswertungs-Blatts: %w", err)
	}
	tabColor := EVAL_TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	// Spaltenbreiten
	g.setColWidth(ws, 1, 3.0)
	g.setColWidth(ws, EV_COL_LABEL, 34.0)
	for c := EV_COL_ACT_LC; c <= EV_COL_ABW_EUR; c++ {
		g.setColWidth(ws, c, 16.0)
	}

	r := 2

	// ─── Titel-Banner ─────────────────────────────────────────────────────────
	g.evalBanner(ws, r, "V. AUSWERTUNG", "Automatische Prüfung von Mittelanforderung und Finanzberichten")
	r += 3

	// ========================================================================
	// A. MITTELANFORDERUNGSPRÜFUNG
	// ========================================================================
	g.evalMainHeader(ws, r, "MITTELANFORDERUNGSPRÜFUNG", "Basis: aktuelle Mittelanforderung")
	r += 3

	kmwMA := g.evalDrawKMWSektion(ws, r, true)
	r = kmwMA.nextRow + EV_TABLE_GAP

	r = g.evalDrawMonatslimit(ws, r) + EV_TABLE_GAP

	resMAInc := g.evalDrawComparisonTable(ws, r, "Prognostizierte Finanzierungsanteile", true, true)
	r = resMAInc.nextRow + EV_TABLE_GAP

	resMAExp := g.evalDrawComparisonTable(ws, r, "Prognoseprüfung (Ausgaben)", false, true)
	r = resMAExp.nextRow + EV_TABLE_GAP

	// ========================================================================
	// B. FINANZBERICHTSPRÜFUNG
	// ========================================================================
	g.evalMainHeader(ws, r, "FINANZBERICHTSPRÜFUNG", "Basis: aktuelle Finanzberichte")
	r += 3

	kmwFB := g.evalDrawKMWSektion(ws, r, false)
	r = kmwFB.nextRow + EV_TABLE_GAP

	resFBInc := g.evalDrawComparisonTable(ws, r, "Finanzierungsanteile", true, false)
	r = resFBInc.nextRow + EV_TABLE_GAP

	// Mehreinnahmen = MAX(0, Ist-Einnahmen ohne KMW − Budget-Einnahmen ohne KMW)
	if kmwFB.mehrCell != "" {
		formula := fmt.Sprintf(
			`=IFERROR(ROUND(MAX(0,(SUM(%s)-%s)-(SUM(%s)-%s)),2),0)`,
			resFBInc.actEURRange, resFBInc.kmwActEUR,
			resFBInc.budEURRange, resFBInc.kmwBudEUR,
		)
		_ = g.setFormula(ws, kmwFB.mehrCell, formula, StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_CALC,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})
	}

	resFBExp := g.evalDrawComparisonTable(ws, r, "Soll-Ist Abweichungsprüfung", false, false)
	r = resFBExp.nextRow

	return nil
}

// ==================================================================================
// BANNER & TITEL
// ==================================================================================

func (g *Generator) evalBanner(ws string, row int, title, subtitle string) {
	titleOpts := StyleOptions{
		Bold: true, Size: 18.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER,
		HAlign: "left", VAlign: "center",
	}
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, titleOpts)
	_ = g.file.SetRowHeight(ws, row, 30.0)

	subOpts := StyleOptions{
		Italic: true, Size: 9.0, FontColor: EV_CLR_BANNER_SUB, FillColor: EV_CLR_BANNER,
		HAlign: "left", VAlign: "center",
	}
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, subOpts)
	_ = g.file.SetRowHeight(ws, row+1, 18.0)
}

func (g *Generator) evalMainHeader(ws string, row int, title, subtitle string) {
	hdrOpts := StyleOptions{
		Bold: true, Size: 13.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER,
		HAlign: "center", VAlign: "center",
	}
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, hdrOpts)
	_ = g.file.SetRowHeight(ws, row, 26.0)

	subOpts := StyleOptions{
		Italic: true, Size: 9.0, FontColor: "595959", HAlign: "center", VAlign: "center",
	}
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, subOpts)
}

func (g *Generator) evalSectionTitle(ws string, row int, title string) {
	opts := StyleOptions{
		Bold: true, Size: 11.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER,
		HAlign: "left", VAlign: "center", BorderTop: 2, BorderBottom: 1, BorderColor: EV_CLR_BORDER,
	}
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, opts)
	_ = g.file.SetRowHeight(ws, row, 22.0)
}

// ==================================================================================
// KMW-MITTELPRÜFUNG
// ==================================================================================

type evalKMWResult struct {
	nextRow  int
	mehrCell string // Zelle "- Mehreinnahmen" (nur FB-Variante befüllt)
}

func (g *Generator) evalDrawKMWSektion(ws string, r int, isMA bool) evalKMWResult {
	title := "KMW-Mittelprüfung"
	if isMA {
		title = "KMW-Mittelprüfung (Mittelanforderung)"
	}
	g.evalSectionTitle(ws, r, title)
	r += 2

	// Linker Block: Label B:C (merged), Wert D
	const lblL1, lblL2, valL = EV_COL_LABEL, EV_COL_LABEL + 1, EV_COL_LABEL + 2
	// Rechter Block: Label F:G (merged), Wert H
	const lblR1, lblR2, valR = EV_COL_LABEL + 4, EV_COL_LABEL + 5, EV_COL_LABEL + 6

	startRow := r

	// --- LINKER BLOCK (Basis) ---
	rBewilligt := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Bewilligte KMW-Mittel", false)
	g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", BG_NAME_KMW_EUR), false, EV_CLR_CALC)
	r++

	rReserve := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Davon Reserve", false)
	g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", bgKostenName("Reserve", "EUR")), false, EV_CLR_CALC)
	r++

	rOperativ := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Operatives Budget (abzgl. Reserve)", false)
	g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-%s,2)", absName(valL, rBewilligt), absName(valL, rReserve)), false, EV_CLR_CALC)
	r++

	rBereit := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Bereitgestellte KMW-Mittel", false)
	g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(SUBTOTAL(109,%s[Betrag]),2),0)", KMW_TABLE_NAME), false, EV_CLR_CALC)
	r++

	rVerfuegbar := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Verfügbare KMW-Mittel", true)
	g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-%s,2)", absName(valL, rOperativ), absName(valL, rBereit)), true, EV_CLR_CALC)
	addrVerfuegbar := absName(valL, rVerfuegbar)
	leftBottom := r

	// --- RECHTER BLOCK (Abzugsoptionen) ---
	rr := startRow
	g.evalKmwLabel(ws, rr, lblR1, valR, "Abzugsoptionen KMW-Mittel", true)
	rr++

	rSaldo := rr
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "− Saldovortrag", false)
	g.evalKmwFormula(ws, cellName(valR, rr), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", DB_NAME_SALDOVORTRAG_EUR), false, EV_CLR_INPUT)
	addrSaldo := absName(valR, rSaldo)
	rr++

	rMehr := rr
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "− Mehreinnahmen", false)
	mehrCell := cellName(valR, rMehr)
	if isMA {
		// In der MA-Variante gibt es keine Ist-Einnahmen → reines Eingabefeld.
		g.evalKmwInput(ws, mehrCell, 0)
		mehrCell = "" // kein Rückkopplungs-Ziel
	} else {
		// Wird nach dem Bau der FB-Finanzierungstabelle per Formel befüllt.
		g.evalKmwInput(ws, mehrCell, 0)
	}
	addrMehr := absName(valR, rMehr)
	rr += 2

	rBereinigt := rr
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Verfügbare KMW-Mittel (bereinigt)", true)
	g.evalKmwFormula(ws, cellName(valR, rr),
		fmt.Sprintf("=ROUND(%s-MAX(0,%s)-MAX(0,%s),2)", addrVerfuegbar, addrSaldo, addrMehr), true, EV_CLR_CALC)
	addrBereinigt := absName(valR, rBereinigt)
	rightBottom := rr

	g.styleOuterBorder(ws, startRow, lblL1, leftBottom, valL, 2, EV_CLR_BORDER)
	g.styleOuterBorder(ws, startRow, lblR1, rightBottom, valR, 2, EV_CLR_BORDER)

	r = leftBottom
	if rightBottom > r {
		r = rightBottom
	}
	r += 2

	// --- MA-spezifische Paare (Anforderung / manueller Betrag) ---
	if isMA {
		// Paar 1: Abzüglich aktuelle Anforderung → Verbleibende KMW-Mittel
		p1Top := r
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwInputColored(ws, cellName(valL, r), 0, EV_CLR_BAD_TXT)
		addrReqL := absName(valL, r)
		g.evalKmwLabel(ws, r, lblR1, lblR2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwInputColored(ws, cellName(valR, r), 0, EV_CLR_BAD_TXT)
		addrReqR := absName(valR, r)
		r++

		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrVerfuegbar, addrReqL), true, EV_CLR_CALC)
		g.evalKmwLabel(ws, r, lblR1, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwFormula(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrBereinigt, addrReqR), true, EV_CLR_CALC)
		p1Bottom := r
		g.styleOuterBorder(ws, p1Top, lblL1, p1Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p1Top, lblR1, p1Bottom, valR, 2, EV_CLR_BORDER)
		r += 2

		// Paar 2: Abzüglich manueller Betrag → Verbleibende KMW-Mittel
		p2Top := r
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich manueller Betrag", false)
		g.evalKmwInputColored(ws, cellName(valL, r), 0, EV_CLR_BAD_TXT)
		addrManL := absName(valL, r)
		g.evalKmwLabel(ws, r, lblR1, lblR2, "Abzüglich manueller Betrag", false)
		g.evalKmwInputColored(ws, cellName(valR, r), 0, EV_CLR_BAD_TXT)
		addrManR := absName(valR, r)
		r++

		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwFormula(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s)-MAX(0,%s),2)", addrVerfuegbar, addrReqL, addrManL), true, EV_CLR_CALC)
		g.evalKmwLabel(ws, r, lblR1, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwFormula(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s)-MAX(0,%s),2)", addrBereinigt, addrReqR, addrManR), true, EV_CLR_CALC)
		p2Bottom := r
		g.styleOuterBorder(ws, p2Top, lblL1, p2Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p2Top, lblR1, p2Bottom, valR, 2, EV_CLR_BORDER)
		r += 1
	}

	return evalKMWResult{nextRow: r, mehrCell: mehrCell}
}

func (g *Generator) evalKmwLabel(ws string, row, c1, c2 int, text string, bold bool) {
	_ = g.mergeCells(ws, cellName(c1, row), cellName(c2, row), text, StyleOptions{
		Bold: bold, Size: 10.0, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalKmwFormula(ws, cell, formula string, bold bool, fill string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		Bold: bold, HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: fill,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalKmwInput(ws, cell string, val interface{}) {
	_ = g.setValue(ws, cell, val, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalKmwInputColored(ws, cell string, val interface{}, fontColor string) {
	_ = g.setValue(ws, cell, val, StyleOptions{
		FontColor: fontColor, HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// ==================================================================================
// MONATSLIMIT-PRÜFUNG (6-Monats-Annahme / 8-Monats-Limit, Lokalwährung)
// ==================================================================================

func (g *Generator) evalDrawMonatslimit(ws string, r int) int {
	g.evalSectionTitle(ws, r, "Monatslimit-Prüfung (Lokalwährung)")
	r += 2

	const lbl1, lbl2, val = EV_COL_LABEL, EV_COL_LABEL + 1, EV_COL_LABEL + 2
	top := r

	// Aktuelle Periode (Dropdown 'Daten'!A1:A18)
	rPeriode := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Aktuelle Periode", false)
	periodCell := cellName(val, r)
	_ = g.setValue(ws, periodCell, "Periode 1", StyleOptions{
		HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	dvPer := excelize.NewDataValidation(true)
	dvPer.Sqref = periodCell
	dvPer.SetSqrefDropList(fmt.Sprintf("'%s'!$A$1:$A$%d", EVAL_DATEN_SHEET, MA_PERIOD_COUNT))
	_ = g.file.AddDataValidation(ws, dvPer)
	periodAddr := absName(val, rPeriode)
	r++

	// Anforderungshöhe (LC) = SUMME LC der gewählten MA-Periode (Gesamtbedarf, apples-to-apples)
	rAnforderung := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Anforderungshöhe (LC)", false)
	numExpr := fmt.Sprintf(`VALUE(TRIM(MID(%s,9,5)))`, periodAddr)
	anfFormula := fmt.Sprintf(
		`=IFERROR(ROUND(SUBTOTAL(109,INDIRECT("MA_"&%s&"[Angefordert (LC)]")),2),0)`, numExpr)
	g.evalLimitValue(ws, cellName(val, r), anfFormula, EV_FMT_LC, false, EV_CLR_CALC)
	anfAddr := absName(val, rAnforderung)
	r++

	// Jahr (Dropdown Jahr 1/2/3)
	rJahr := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Jahr", false)
	jahrCell := cellName(val, r)
	_ = g.setValue(ws, jahrCell, BG_YEARS[0], StyleOptions{
		HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	dvJahr := excelize.NewDataValidation(true)
	dvJahr.Sqref = jahrCell
	dvJahr.SetDropList(BG_YEARS)
	_ = g.file.AddDataValidation(ws, dvJahr)
	jahrAddr := absName(val, rJahr)
	r++

	// Jahresbudget (LC) = CHOOSE(MATCH(Jahr, {...}), SUBTOTAL je Jahresspalte)
	rJahresbudget := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Jahresbudget (LC)", false)
	jbFormula := fmt.Sprintf(
		`=IFERROR(ROUND(CHOOSE(MATCH(%s,{"%s";"%s";"%s"},0),SUBTOTAL(109,%s[%s]),SUBTOTAL(109,%s[%s]),SUBTOTAL(109,%s[%s])),2),0)`,
		jahrAddr, BG_YEARS[0], BG_YEARS[1], BG_YEARS[2],
		BG_TABLE_AUSG, BG_YEARS[0], BG_TABLE_AUSG, BG_YEARS[1], BG_TABLE_AUSG, BG_YEARS[2])
	g.evalLimitValue(ws, cellName(val, r), jbFormula, EV_FMT_LC, false, EV_CLR_CALC)
	jbAddr := absName(val, rJahresbudget)
	r++

	// Limit-Monate (8) – Eingabe
	rLimitMon := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Limit-Monate", false)
	g.evalLimitInput(ws, cellName(val, r), 8, "0")
	limitMonAddr := absName(val, rLimitMon)
	r++

	// Bezugszeitraum (Monate) (12) – Eingabe
	rBezug := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Bezugszeitraum (Monate)", false)
	g.evalLimitInput(ws, cellName(val, r), 12, "0")
	bezugAddr := absName(val, rBezug)
	r++

	// Anforderung deckt (Monate) (6) – rein informativ
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Anforderung deckt (Monate)", false)
	g.evalLimitInput(ws, cellName(val, r), 6, "0")
	r++

	// 8-Monats-Limit (LC)
	rLimit := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "8-Monats-Limit (LC)", true)
	limitFormula := fmt.Sprintf("=IFERROR(ROUND(%s*%s/%s,2),0)", jbAddr, limitMonAddr, bezugAddr)
	g.evalLimitValue(ws, cellName(val, r), limitFormula, EV_FMT_LC, true, EV_CLR_CALC)
	limitAddr := absName(val, rLimit)
	r++

	// Status
	rStatus := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Status", true)
	statusCell := cellName(val, r)
	statusFormula := fmt.Sprintf(`=IF(%s<=%s,"OK","ÜBERSCHRITTEN")`, anfAddr, limitAddr)
	_ = g.setFormula(ws, statusCell, statusFormula, StyleOptions{
		Bold: true, HAlign: "center", VAlign: "center", FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	statusAddr := absName(val, rStatus)
	g.addConditionalFormat(ws, statusCell, fmt.Sprintf(`%s="OK"`, statusAddr), StyleOptions{
		Bold: true, FontColor: EV_CLR_GOOD_TXT, FillColor: EV_CLR_GOOD, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	g.addConditionalFormat(ws, statusCell, fmt.Sprintf(`%s<>"OK"`, statusAddr), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	r++

	// Überschreitung (LC)
	rUeber := r
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Überschreitung (LC)", false)
	ueberCell := cellName(val, r)
	g.evalLimitValue(ws, ueberCell, fmt.Sprintf("=ROUND(MAX(0,%s-%s),2)", anfAddr, limitAddr), EV_FMT_LC, false, EV_CLR_CALC)
	ueberAddr := absName(val, rUeber)
	g.addConditionalFormat(ws, ueberCell, fmt.Sprintf(`%s>0`, ueberAddr), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD, HAlign: "right", VAlign: "center",
		NumFormat: EV_FMT_LC, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	r++

	// Auslastung %
	g.evalKmwLabel(ws, r, lbl1, lbl2, "Auslastung %", false)
	g.evalLimitValue(ws, cellName(val, r), fmt.Sprintf("=IFERROR(%s/%s,0)", anfAddr, limitAddr), EV_FMT_PCT, false, EV_CLR_CALC)
	bottom := r

	g.styleOuterBorder(ws, top, lbl1, bottom, val, 2, EV_CLR_BORDER)
	return r
}

func (g *Generator) evalLimitValue(ws, cell, formula, numFmt string, bold bool, fill string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		Bold: bold, HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: fill,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalLimitInput(ws, cell string, val interface{}, numFmt string) {
	_ = g.setValue(ws, cell, val, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// ==================================================================================
// VERGLEICHSTABELLEN (Einnahmen / Ausgaben)
// ==================================================================================

func (g *Generator) evalDrawComparisonTable(ws string, r int, title string, isIncome, isMA bool) evalCompResult {
	g.evalSectionTitle(ws, r, title)
	r += 2

	actualsLabel := "Kumulativ"
	if isMA {
		actualsLabel = "Prognose"
	}

	headers := []string{
		"Kategorie",
		actualsLabel + " (LC)", "Budget (LC)", "Differenz (LC)", "Abweichung (LC)",
		actualsLabel + " (EUR)", "Budget (EUR)", "Differenz (EUR)", "Abweichung (EUR)",
	}
	hdrRow := r
	for i, h := range headers {
		_ = g.setValue(ws, cellName(EV_COL_LABEL+i, hdrRow), h, StyleOptions{
			Bold: true, Size: 9.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER,
			HAlign: "center", VAlign: "center", WrapText: true,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
		})
	}
	_ = g.file.SetRowHeight(ws, hdrRow, 28.0)
	r++

	rowNames := TYPE_NAMES
	if !isIncome {
		rowNames = EXPENSE_CATEGORIES
	}
	dataStart := r
	kmwActEUR, kmwBudEUR := "", ""

	for i, name := range rowNames {
		row := dataStart + i

		// Label
		_ = g.setValue(ws, cellName(EV_COL_LABEL, row), name, StyleOptions{
			HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})

		// Zinsertraege (Einnahmen) hat keine Budgetzeile → komplette Zeile leer lassen.
		if isIncome && name == "Zinsertraege" {
			for c := EV_COL_ACT_LC; c <= EV_COL_ABW_EUR; c++ {
				_ = g.setStyle(ws, cellName(c, row), cellName(c, row), StyleOptions{
					BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
				})
			}
			continue
		}

		// ── Actuals (Prognose / Kumulativ) ──
		actLC := cellName(EV_COL_ACT_LC, row)
		actEUR := cellName(EV_COL_ACT_EUR, row)
		if isMA && isIncome {
			// Keine automatische Prognose-Quelle für Einnahmen → Eingabefelder.
			g.evalCellInput(ws, actLC, EV_FMT_LC)
			g.evalCellInput(ws, actEUR, EV_FMT_EUR)
		} else {
			lcF, eurF := g.evalActualFormulas(isIncome, isMA, name, i)
			g.evalCellFormula(ws, actLC, lcF, EV_FMT_LC, EV_CLR_CALC)
			g.evalCellFormula(ws, actEUR, eurF, EV_FMT_EUR, EV_CLR_CALC)
		}

		// ── Budget ──
		budLCName, budEURName := g.evalBudgetNames(isIncome, name, i)
		budLC := cellName(EV_COL_BUD_LC, row)
		budEUR := cellName(EV_COL_BUD_EUR, row)
		g.evalCellFormula(ws, budLC, fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", budLCName), EV_FMT_LC, EV_CLR_CALC)
		g.evalCellFormula(ws, budEUR, fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", budEURName), EV_FMT_EUR, EV_CLR_CALC)

		// ── Differenz / Abweichung (LC) ──
		g.evalCellFormula(ws, cellName(EV_COL_DIF_LC, row),
			fmt.Sprintf("=ROUND(IFERROR(%s-%s,0),2)", absName(EV_COL_ACT_LC, row), absName(EV_COL_BUD_LC, row)), EV_FMT_LC, "")
		g.evalCellFormula(ws, cellName(EV_COL_ABW_LC, row),
			fmt.Sprintf("=ROUND(IFERROR((%s/%s)-1,0),4)", absName(EV_COL_ACT_LC, row), absName(EV_COL_BUD_LC, row)), EV_FMT_PCT, "")

		// ── Differenz / Abweichung (EUR) ──
		g.evalCellFormula(ws, cellName(EV_COL_DIF_EUR, row),
			fmt.Sprintf("=ROUND(IFERROR(%s-%s,0),2)", absName(EV_COL_ACT_EUR, row), absName(EV_COL_BUD_EUR, row)), EV_FMT_EUR, "")
		g.evalCellFormula(ws, cellName(EV_COL_ABW_EUR, row),
			fmt.Sprintf("=ROUND(IFERROR((%s/%s)-1,0),4)", absName(EV_COL_ACT_EUR, row), absName(EV_COL_BUD_EUR, row)), EV_FMT_PCT, "")

		if isIncome && name == "KMW-Mittel" {
			kmwActEUR = absName(EV_COL_ACT_EUR, row)
			kmwBudEUR = absName(EV_COL_BUD_EUR, row)
		}
	}

	dataEnd := dataStart + len(rowNames) - 1

	// ── GESAMT-Zeile ──
	totalRow := dataEnd + 1
	g.evalTotalRow(ws, totalRow, dataStart, dataEnd)

	// ── Ampel-Conditional-Format auf Abweichungsspalten (nur Ausgaben) ──
	if !isIncome {
		g.evalDeviationConditional(ws, EV_COL_ABW_LC, dataStart, dataEnd)
		g.evalDeviationConditional(ws, EV_COL_ABW_EUR, dataStart, dataEnd)
	}

	g.styleOuterBorder(ws, hdrRow, EV_COL_LABEL, totalRow, EV_COL_ABW_EUR, 2, EV_CLR_BORDER)

	return evalCompResult{
		nextRow:     totalRow,
		actEURRange: fmt.Sprintf("%s:%s", absName(EV_COL_ACT_EUR, dataStart), absName(EV_COL_ACT_EUR, dataEnd)),
		budEURRange: fmt.Sprintf("%s:%s", absName(EV_COL_BUD_EUR, dataStart), absName(EV_COL_BUD_EUR, dataEnd)),
		kmwActEUR:   kmwActEUR,
		kmwBudEUR:   kmwBudEUR,
	}
}

// evalActualFormulas liefert die LC-/EUR-Formeln für die Ist-/Prognose-Spalten,
// berechnet über die Spill-Stacks des "Daten"-Blatts.
func (g *Generator) evalActualFormulas(isIncome, isMA bool, name string, idx int) (string, string) {
	if isIncome {
		// Kumulative Ist-Einnahmen je Typ = explizite + Durchschnittskurs-Stacks.
		lc := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s)+SUMIF(%s,"%s",%s),2)`,
			evalStackRange(EV_DTN_EIN1_TYP), name, evalStackRange(EV_DTN_EIN1_LC),
			evalStackRange(EV_DTN_EIN2_TYP), name, evalStackRange(EV_DTN_EIN2_LC))
		eur := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s)+SUMIF(%s,"%s",%s),2)`,
			evalStackRange(EV_DTN_EIN1_TYP), name, evalStackRange(EV_DTN_EIN1_EUR),
			evalStackRange(EV_DTN_EIN2_TYP), name, evalStackRange(EV_DTN_EIN2_EUR))
		return lc, eur
	}

	if isMA {
		// Prognose der Ausgaben = Mittelanforderung je Kategorie (MA-Stack, nach Name).
		lc := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s),2)`,
			evalStackRange(EV_DTN_MA_CAT), name, evalStackRange(EV_DTN_MA_LC))
		eur := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s),2)`,
			evalStackRange(EV_DTN_MA_CAT), name, evalStackRange(EV_DTN_MA_EUR))
		return lc, eur
	}

	// Kumulative Ist-Ausgaben je Kategorie = Ausgaben-Stack, ID beginnt mit "<k>.".
	catNo := idx + 1
	lc := fmt.Sprintf(`=ROUND(SUMIF(%s,"%d.*",%s),2)`,
		evalStackRange(EV_DTN_AUS_ID), catNo, evalStackRange(EV_DTN_AUS_LC))
	eur := fmt.Sprintf(`=ROUND(SUMIF(%s,"%d.*",%s),2)`,
		evalStackRange(EV_DTN_AUS_ID), catNo, evalStackRange(EV_DTN_AUS_EUR))
	return lc, eur
}

// evalBudgetNames liefert die benannten Budget-Bereiche (LC, EUR) je Zeile.
func (g *Generator) evalBudgetNames(isIncome bool, name string, idx int) (string, string) {
	if isIncome {
		switch name {
		case "Eigenmittel":
			return BG_NAME_EIGEN_LW, BG_NAME_EIGEN_EUR
		case "Drittmittel":
			return BG_NAME_DRITT_LW, BG_NAME_DRITT_EUR
		case "KMW-Mittel":
			return BG_NAME_KMW_LW, BG_NAME_KMW_EUR
		default:
			return "0", "0"
		}
	}
	return bgKostenName(name, "LW"), bgKostenName(name, "EUR")
}

func (g *Generator) evalTotalRow(ws string, totalRow, dataStart, dataEnd int) {
	// Label
	_ = g.setValue(ws, cellName(EV_COL_LABEL, totalRow), "GESAMT", StyleOptions{
		Bold: true, FontColor: EV_CLR_TOTAL_TXT, FillColor: EV_CLR_TOTAL, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
	})

	sumCol := func(col int, numFmt string) {
		rng := fmt.Sprintf("%s:%s", absName(col, dataStart), absName(col, dataEnd))
		_ = g.setFormula(ws, cellName(col, totalRow), fmt.Sprintf("=ROUND(SUM(%s),2)", rng), StyleOptions{
			Bold: true, FontColor: EV_CLR_TOTAL_TXT, FillColor: EV_CLR_TOTAL, HAlign: "right", VAlign: "center",
			NumFormat: numFmt, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
		})
	}
	pctCol := func(col, actCol, budCol int) {
		_ = g.setFormula(ws, cellName(col, totalRow),
			fmt.Sprintf("=ROUND(IFERROR((%s/%s)-1,0),4)", absName(actCol, totalRow), absName(budCol, totalRow)),
			StyleOptions{
				Bold: true, FontColor: EV_CLR_TOTAL_TXT, FillColor: EV_CLR_TOTAL, HAlign: "right", VAlign: "center",
				NumFormat: EV_FMT_PCT, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
	}

	sumCol(EV_COL_ACT_LC, EV_FMT_LC)
	sumCol(EV_COL_BUD_LC, EV_FMT_LC)
	sumCol(EV_COL_DIF_LC, EV_FMT_LC)
	pctCol(EV_COL_ABW_LC, EV_COL_ACT_LC, EV_COL_BUD_LC)
	sumCol(EV_COL_ACT_EUR, EV_FMT_EUR)
	sumCol(EV_COL_BUD_EUR, EV_FMT_EUR)
	sumCol(EV_COL_DIF_EUR, EV_FMT_EUR)
	pctCol(EV_COL_ABW_EUR, EV_COL_ACT_EUR, EV_COL_BUD_EUR)

	_ = g.file.SetRowHeight(ws, totalRow, 20.0)
}

// evalDeviationConditional färbt Abweichungen: > 20 % rot, 10–20 % gelb.
func (g *Generator) evalDeviationConditional(ws string, col, dataStart, dataEnd int) {
	rng := fmt.Sprintf("%s:%s", cellName(col, dataStart), cellName(col, dataEnd))
	topRel := cellName(col, dataStart)

	g.addConditionalFormat(ws, rng, fmt.Sprintf("%s>0.2", topRel), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD,
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_PCT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	g.addConditionalFormat(ws, rng, fmt.Sprintf("AND(%s>=0.1,%s<=0.2)", topRel, topRel), StyleOptions{
		Bold: true, FontColor: EV_CLR_WARN_TXT, FillColor: EV_CLR_WARN,
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_PCT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalCellFormula(ws, cell, formula, numFmt, fill string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: fill,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalCellInput(ws, cell, numFmt string) {
	_ = g.setValue(ws, cell, 0, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// evalStackRange liefert einen begrenzten, nicht-volatilen Bezug auf eine Stack-Spalte.
func evalStackRange(col string) string {
	return fmt.Sprintf("'%s'!$%s$2:$%s$%d", EVAL_DATEN_SHEET, col, col, EVAL_STACK_MAXROW)
}
