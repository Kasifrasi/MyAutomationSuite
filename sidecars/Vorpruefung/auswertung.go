package main

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ==================================================================================
// Blatt "V. AUSWERTUNG"
//
//	A) MITTELANFORDERUNGSPRÜFUNG (Basis: ausgewählte Mittelanforderung / Prognose)
//	B) FINANZBERICHTSPRÜFUNG     (Basis: ausgewählter Finanzbericht / kumulativ)
//
// Über zwei Auswahllisten (rechts in den Sektionen) wird ein befüllter Finanzbericht
// (Periode N) sowie eine Mittelanforderung der Folgeperiode (N+1) gewählt. Die
// Listen werden in daten.go dynamisch (FILTER) aus den befüllten Perioden gebildet.
// Gibt es mehrere Mittelanforderungen einer Periode, erscheinen "Periode X (#k)";
// die Auswahl von "(#k)" addiert in der Prognoseprüfung alle Anforderungen #1..#k.
// ==================================================================================

const (
	EVAL_SHEET_NAME   = "V. AUSWERTUNG"
	EVAL_TAB_COLOR    = "FFFF00" // Gelb
	EVAL_DATEN_SHEET  = "Daten"
	EVAL_FB_SHEET     = "III. Finanzberichte"
	EVAL_STACK_MAXROW = 500

	// Spalten der Vergleichstabellen (B … J)
	EV_COL_LABEL   = 2
	EV_COL_ACT_LC  = 3
	EV_COL_BUD_LC  = 4
	EV_COL_DIF_LC  = 5
	EV_COL_ABW_LC  = 6
	EV_COL_ACT_EUR = 7
	EV_COL_BUD_EUR = 8
	EV_COL_DIF_EUR = 9
	EV_COL_ABW_EUR = 10

	// Auswahl-Panel (zentriert, oben in der Sektion; Spalten C … G)
	EV_PB_C1   = 3 // Box-/Label-Start (C)
	EV_PB_L2   = 4 // Label-Ende (D)
	EV_PB_V1   = 5 // Wert-/Slot-LC-Start (E)
	EV_PB_SLC2 = 6 // Slot-LC-Ende (F)
	EV_PB_SEU1 = 7 // Slot-EUR-Start (G)
	EV_PB_C2   = 7 // Box-/Wert-Ende (G)

	EV_TABLE_GAP = 2
	EV_MA_SLOTS  = 6 // max. gleichzeitig anzeigbare Anforderungen je Periode

	// Daten-Helfer (Spaltennummern auf dem Blatt "Daten")
	EV_DTN_MA_META_J     = 53 // BA  Tabellenindex j
	EV_DTN_MA_META_PER   = 54 // BB  Periode von MA_j
	EV_DTN_MA_META_FILL  = 55 // BC  befüllt?
	EV_DTN_MA_META_RANK  = 56 // BD  Rang innerhalb der Periode
	EV_DTN_MA_META_LABEL = 57 // BE  "Periode X (#k)"
	EV_DTN_MA_META_SUMLC = 58 // BF  Summe Angefordert (LC)
	EV_DTN_MA_META_SUMEU = 59 // BG  Summe Angefordert (EUR)
	EV_DTN_MA_META_EIGDR = 60 // BH  Eigen+Dritt (EUR)
	EV_DTN_FB_META_PER   = 62 // BJ  Periode
	EV_DTN_FB_META_FILL  = 63 // BK  befüllt?
	EV_DTN_FB_META_LABEL = 64 // BL  "Periode X"
	EV_DTN_MA_LISTE      = 66 // BN  FILTER-Spill (Auswahlliste MA)
	EV_DTN_FB_LISTE      = 68 // BP  FILTER-Spill (Auswahlliste FB)
	EV_DTN_MAG_PER       = 70 // BR  Grid: Periode
	EV_DTN_MAG_RANK      = 71 // BS  Grid: Rang
	EV_DTN_MAG_CAT       = 72 // BT  Grid: Kategorie
	EV_DTN_MAG_LC        = 73 // BU  Grid: LC
	EV_DTN_MAG_EUR       = 74 // BV  Grid: EUR
	// Grid-Block je MA-Tabelle: 8 Kostenkategorien (len(MA_CATEGORIES)) + 3
	// Finanzierungsarten (Eigenmittel/Drittmittel/KMW-Mittel) für die Prognose
	// der Finanzierungsanteile. Muss zu gridEntries in daten.go passen.
	EV_DTN_MAG_BLOCK = 8 + 3
	EV_DTN_MAG_ROWS  = MA_PERIOD_COUNT * EV_DTN_MAG_BLOCK

	EVAL_NAME_MA_LISTE = "MA_Auswahl_Liste"
	EVAL_NAME_FB_LISTE = "FB_Auswahl_Liste"

	// Farben
	EV_CLR_BANNER     = "212F3D"
	EV_CLR_BANNER_TXT = "FFFFFF"
	EV_CLR_BANNER_SUB = "B4BEC8"
	EV_CLR_HEADER     = "D3D3D3"
	EV_CLR_TOTAL      = "212F3D"
	EV_CLR_TOTAL_TXT  = "FFFFFF"
	EV_CLR_INPUT      = "FFFAE5" // nur für bearbeitbare Felder
	EV_CLR_DEDUCT     = "EAF2F8" // abzuziehende (berechnete) Beträge
	EV_CLR_CALC       = "F2F2F2" // sonstige berechnete Felder
	EV_CLR_BORDER     = "808080"
	EV_CLR_GRID       = "D3D3D3"
	EV_CLR_BLACK      = "000000"
	EV_CLR_GOOD       = "C6EFCE"
	EV_CLR_GOOD_TXT   = "006100"
	EV_CLR_BAD        = "FFC7CE"
	EV_CLR_BAD_TXT    = "9C0006"
	EV_CLR_WARN       = "FCF3CF"
	EV_CLR_WARN_TXT   = "9C640C"
	EV_CLR_PANEL_REV  = "D6EAF8" // Hintergrund aktiver (eingeblendeter) Slots

	EV_FMT_LC  = "#,##0.00"
	EV_FMT_EUR = `#,##0.00" €"`
	EV_FMT_PCT = "0.0%"
)

// evalSelRefs bündelt die Adressen der Auswahl-Steuerzellen.
type evalSelRefs struct {
	fbSelNum string // Periodennummer des gewählten Finanzberichts (N)
	maSelP   string // Periode der gewählten Mittelanforderung (N+1)
	maSelK   string // Rang (#k) der gewählten Mittelanforderung
}

type evalCompResult struct {
	nextRow     int
	actEURRange string
	budEURRange string
	kmwActEUR   string
	kmwBudEUR   string
}

type evalKMWResult struct {
	nextRow   int
	mehrCell  string // Wertzelle "Mehreinnahmen" (Formel wird nachgelagert gesetzt)
	prognCell string // Wertzelle "Prognostizierte Mehreinnahmen" (nur MA)
	saldoCell string // Wertzelle "Saldovortrag"
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

	g.setColWidth(ws, 1, 3.0)
	g.setColWidth(ws, EV_COL_LABEL, 42.0)
	for c := EV_COL_ACT_LC; c <= EV_COL_ABW_EUR; c++ {
		g.setColWidth(ws, c, 20.0)
	}

	r := 2
	g.evalBanner(ws, r, "V. AUSWERTUNG", "Automatische Prüfung von Mittelanforderung und Finanzberichten")
	r += 3

	// ========================================================================
	// A. MITTELANFORDERUNGSPRÜFUNG
	// ========================================================================
	maSectionTop := r
	g.evalMainHeader(ws, r, "MITTELANFORDERUNGSPRÜFUNG", "Basis: ausgewählte Mittelanforderung (Folgeperiode des Finanzberichts)")
	r += 3

	// Auswahl-Panel (zentriert, oben) zuerst – liefert maSelP-/maSelK-Steuerzellen.
	maSelPCell, maSelKCell, r := g.evalDrawMAPanel(ws, r)
	r += EV_TABLE_GAP
	sel := evalSelRefs{maSelP: maSelPCell, maSelK: maSelKCell}

	maKMW := g.evalDrawKMWSektion(ws, r, true, sel)
	r = maKMW.nextRow + EV_TABLE_GAP

	r = g.evalDrawMonatslimit(ws, r, sel) + EV_TABLE_GAP

	resMAInc := g.evalDrawComparisonTable(ws, r, "Prognostizierte Finanzierungsanteile", true, true, sel)
	r = resMAInc.nextRow + EV_TABLE_GAP

	resMAExp := g.evalDrawComparisonTable(ws, r, "Prognoseprüfung (Ausgaben)", false, true, sel)
	r = resMAExp.nextRow + EV_TABLE_GAP

	// ========================================================================
	// B. FINANZBERICHTSPRÜFUNG
	// ========================================================================
	r += 4 // etwas mehr Abstand zur Mittelanforderungsprüfung
	fbSectionTop := r
	g.evalMainHeader(ws, r, "FINANZBERICHTSPRÜFUNG", "Basis: ausgewählter Finanzbericht (kumulativ bis zur Periode)")
	r += 3

	// FB-Auswahl-Panel (zentriert, oben) – liefert die Periodennummer N.
	fbSelNumCell, fbNext := g.evalDrawFBPanel(ws, r)
	r = fbNext + EV_TABLE_GAP
	sel.fbSelNum = fbSelNumCell
	g.evalFBSelNumAddr = fmt.Sprintf("'%s'!%s", ws, fbSelNumCell)

	// Jetzt maSelP = N+1 setzen (Folgeperiode des gewählten Finanzberichts).
	_ = g.setFormula(ws, maSelPCell, fmt.Sprintf("=%s+1", fbSelNumCell), StyleOptions{
		Bold: true, HAlign: "center", VAlign: "center", NumFormat: "0", FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})

	fbKMW := g.evalDrawKMWSektion(ws, r, false, sel)
	r = fbKMW.nextRow + EV_TABLE_GAP

	resFBInc := g.evalDrawComparisonTable(ws, r, "Finanzierungsanteile", true, false, sel)
	r = resFBInc.nextRow + EV_TABLE_GAP

	// FB-Mehreinnahmen = MAX(0, Ist ohne KMW − Budget ohne KMW)
	mehrFormula := fmt.Sprintf(
		`=IFERROR(ROUND(MAX(0,(SUM(%s)-%s)-(SUM(%s)-%s)),2),0)`,
		resFBInc.actEURRange, resFBInc.kmwActEUR, resFBInc.budEURRange, resFBInc.kmwBudEUR)
	g.evalDeduct(ws, fbKMW.mehrCell, mehrFormula)

	resFBExp := g.evalDrawComparisonTable(ws, r, "Soll-Ist Abweichungsprüfung", false, false, sel)
	r = resFBExp.nextRow

	// ── Nachgelagerte (sektionsübergreifende) Formeln der MA-Abzugsoptionen ──
	// Mehreinnahmen der MA-Prüfung = identischer Wert wie in der FB-Prüfung.
	g.evalDeduct(ws, maKMW.mehrCell, fmt.Sprintf("=%s", fbKMW.mehrCell))

	// Prognostizierte (zusätzliche) Mehreinnahmen:
	// MAX(0, MAX(0, IstOhneKMW + MA-Eigen/Dritt − BudgetOhneKMW) − bereits realisierte Mehreinnahmen)
	if maKMW.prognCell != "" {
		realAct := fmt.Sprintf("(SUM(%s)-%s)", resFBInc.actEURRange, resFBInc.kmwActEUR)
		realBud := fmt.Sprintf("(SUM(%s)-%s)", resFBInc.budEURRange, resFBInc.kmwBudEUR)
		maEig := fmt.Sprintf(
			`SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1")`,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_EIGDR, 1, MA_PERIOD_COUNT),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_PERIOD_COUNT), maSelPCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT), maSelKCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT))
		prognFormula := fmt.Sprintf(
			`=ROUND(MAX(0,MAX(0,%s+%s-%s)-%s),2)`, realAct, maEig, realBud, fbKMW.mehrCell)
		g.evalDeduct(ws, maKMW.prognCell, prognFormula)
	}

	// Die nachgelagerten evalDeduct-/evalDeductPlaceholder-Aufrufe überschreiben die
	// kräftige Box-Außenkante (styleOuterBorder) der rechten Abzugsspalte. Rechte Kante
	// daher gezielt wiederherstellen (gilt auch für Saldovortrag, der am Box-Rand sitzt).
	for _, cell := range []string{
		maKMW.saldoCell, maKMW.mehrCell, maKMW.prognCell,
		fbKMW.saldoCell, fbKMW.mehrCell,
	} {
		if cell != "" {
			g.reapplyRightBorder(ws, cell, 2, EV_CLR_BORDER)
		}
	}

	// Schreibgeschützte Spiegel-Panels (rechts neben der jeweiligen Prüfung).
	g.evalDrawMAMirrorPanel(ws, maSectionTop, sel)
	g.evalDrawFBMirrorPanel(ws, fbSectionTop, sel)

	return nil
}

// ==================================================================================
// BANNER & TITEL
// ==================================================================================

func (g *Generator) evalBanner(ws string, row int, title, subtitle string) {
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, StyleOptions{
		Bold: true, Size: 18.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "left", VAlign: "center",
	})
	_ = g.file.SetRowHeight(ws, row, 30.0)
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, StyleOptions{
		Italic: true, Size: 9.0, FontColor: EV_CLR_BANNER_SUB, FillColor: EV_CLR_BANNER, HAlign: "left", VAlign: "center",
	})
	_ = g.file.SetRowHeight(ws, row+1, 18.0)
}

func (g *Generator) evalMainHeader(ws string, row int, title, subtitle string) {
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, StyleOptions{
		Bold: true, Size: 13.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "center", VAlign: "center",
	})
	_ = g.file.SetRowHeight(ws, row, 26.0)
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, StyleOptions{
		Italic: true, Size: 9.0, FontColor: "595959", HAlign: "center", VAlign: "center",
	})
}

func (g *Generator) evalSectionTitle(ws string, row int, title string) {
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, StyleOptions{
		Bold: true, Size: 11.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER,
		HAlign: "left", VAlign: "center", BorderTop: 2, BorderBottom: 1, BorderColor: EV_CLR_BORDER,
	})
	_ = g.file.SetRowHeight(ws, row, 22.0)
}

// ==================================================================================
// KMW-MITTELPRÜFUNG
// ==================================================================================

func (g *Generator) evalDrawKMWSektion(ws string, r int, isMA bool, sel evalSelRefs) evalKMWResult {
	title := "KMW-Mittelprüfung"
	if isMA {
		title = "KMW-Mittelprüfung (Mittelanforderung)"
	}
	g.evalSectionTitle(ws, r, title)
	r += 2

	const lblL1, lblL2, valL = EV_COL_LABEL, EV_COL_LABEL + 1, EV_COL_LABEL + 2 // B C | D
	const tog, lblR1, lblR2, valR = EV_COL_LABEL + 4, EV_COL_LABEL + 5, EV_COL_LABEL + 6, EV_COL_LABEL + 7

	startRow := r

	// --- LINKER BLOCK (Basis) ---
	rBew := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Bewilligte KMW-Mittel", false)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", BG_NAME_KMW_EUR), false)
	r++
	rRes := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Davon Reserve", false)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", bgKostenName("Reserve", "EUR")), false)
	r++
	rOp := r
	g.evalMergedFormula(ws, cellName(lblL1, r), cellName(lblL2, r),
		fmt.Sprintf(`=IF(%s="Ja","Operatives Budget (Reserve freigegeben)","Operatives Budget (abzgl. Reserve)")`, BG_NAME_RESERVE),
		StyleOptions{Size: 10.0, HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID})
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf(
		`=IF(%s="Ja",ROUND(%s,2),ROUND(%s-%s,2))`,
		BG_NAME_RESERVE, absName(valL, rBew), absName(valL, rBew), absName(valL, rRes)), false)
	r++
	rBer := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Bereitgestellte KMW-Mittel", false)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(SUBTOTAL(109,%s[Betrag]),2),0)", KMW_TABLE_NAME), false)
	r++
	rVerf := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Verfügbare KMW-Mittel", true)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-%s,2)", absName(valL, rOp), absName(valL, rBer)), true)
	addrVerf := absName(valL, rVerf)
	leftBottom := r

	// --- RECHTER BLOCK (Abzugsoptionen, mit Abzug/Kein-Abzug-Schaltern) ---
	rr := startRow
	_ = g.mergeCells(ws, cellName(tog, rr), cellName(valR, rr), "Abzugsoptionen KMW-Mittel", StyleOptions{
		Bold: true, Size: 10.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
	})
	rr++

	type dedRow struct{ togCell, valCell string }
	var deds []dedRow

	// Saldovortrag (berechnet)
	g.evalToggle(ws, cellName(tog, rr))
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Saldovortrag", false)
	g.evalDeduct(ws, cellName(valR, rr), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", DB_NAME_SALDOVORTRAG_EUR))
	saldoCell := cellName(valR, rr)
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	// Mehreinnahmen (Formel wird nachgelagert gesetzt)
	g.evalToggle(ws, cellName(tog, rr))
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Mehreinnahmen", false)
	g.evalDeductPlaceholder(ws, cellName(valR, rr))
	mehrCell := cellName(valR, rr)
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	prognCell := ""
	if isMA {
		g.evalToggle(ws, cellName(tog, rr))
		g.evalKmwLabel(ws, rr, lblR1, lblR2, "Prognostizierte Mehreinnahmen", false)
		g.evalDeductPlaceholder(ws, cellName(valR, rr))
		prognCell = cellName(valR, rr)
		deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
		rr++
	}

	// Abzugsoptionen KMW insgesamt
	rIns := rr
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Abzugsoptionen KMW insgesamt", true)
	var terms []string
	for _, d := range deds {
		terms = append(terms, fmt.Sprintf(`IF(%s="Abzug",MAX(0,%s),0)`, d.togCell, d.valCell))
	}
	g.evalKmwCalc(ws, cellName(valR, rr), fmt.Sprintf("=ROUND(%s,2)", strings.Join(terms, "+")), true)
	addrInsgesamt := absName(valR, rIns)
	rr++

	// Verfügbare KMW-Mittel (bereinigt)
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Verfügbare KMW-Mittel (bereinigt)", true)
	g.evalKmwCalc(ws, cellName(valR, rr), fmt.Sprintf("=ROUND(%s-%s,2)", addrVerf, addrInsgesamt), true)
	addrBereinigt := absName(valR, rr)
	rightBottom := rr

	g.styleOuterBorder(ws, startRow, lblL1, leftBottom, valL, 2, EV_CLR_BORDER)
	g.styleOuterBorder(ws, startRow, tog, rightBottom, valR, 2, EV_CLR_BORDER)

	r = leftBottom
	if rightBottom > r {
		r = rightBottom
	}
	r += 2

	if isMA {
		// Rechte Paar-Beschriftungen spannen F..H (Wert I) – bündig mit der
		// Abzugsoptionen-Box (tog..valR = F..I).
		// "Abzüglich aktuelle Anforderung" = KMW-Mittel-Anforderung der EINEN aktuell
		// gewählten Mittelanforderung (Periode P, Rang exakt = k). Bewusst NICHT
		// zusammengesetzt (#1..#k): frühere Anforderungen einer Periode sind bereits
		// über die bereitgestellten KMW-Mittel (KMW-Tabelle) erfasst. Beide Seiten
		// (Basis / bereinigt) ziehen denselben berechneten Wert ab (grauer Hintergrund).
		p1Top := r
		reqFormula := evalMACurrentKMWRequest(sel)
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwCalc(ws, cellName(valL, r), reqFormula, false)
		addrReqL := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwCalc(ws, cellName(valR, r), reqFormula, false)
		addrReqR := absName(valR, r)
		r++
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrVerf, addrReqL), true)
		g.evalKmwLabel(ws, r, tog, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwCalc(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrBereinigt, addrReqR), true)
		p1Bottom := r
		g.styleOuterBorder(ws, p1Top, lblL1, p1Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p1Top, tog, p1Bottom, valR, 2, EV_CLR_BORDER)
		r += 2

		// Paar 2: Abzüglich manueller Betrag → aus dem "Manueller Betrag"-Feld der gewählten MA
		p2Top := r
		manFormula := evalMAChooseManBetrag(sel.maSelP)
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich manueller Betrag", false)
		g.evalDeduct(ws, cellName(valL, r), manFormula)
		addrManL := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Abzüglich manueller Betrag", false)
		g.evalDeduct(ws, cellName(valR, r), manFormula)
		addrManR := absName(valR, r)
		r++
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s)-MAX(0,%s),2)", addrVerf, addrReqL, addrManL), true)
		g.evalKmwLabel(ws, r, tog, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwCalc(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s)-MAX(0,%s),2)", addrBereinigt, addrReqR, addrManR), true)
		p2Bottom := r
		g.styleOuterBorder(ws, p2Top, lblL1, p2Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p2Top, tog, p2Bottom, valR, 2, EV_CLR_BORDER)
		r++
	}

	return evalKMWResult{nextRow: r, mehrCell: mehrCell, prognCell: prognCell, saldoCell: saldoCell}
}

func (g *Generator) evalKmwLabel(ws string, row, c1, c2 int, text string, bold bool) {
	_ = g.mergeCells(ws, cellName(c1, row), cellName(c2, row), text, StyleOptions{
		Bold: bold, Size: 10.0, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalKmwCalc(ws, cell, formula string, bold bool) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		Bold: bold, HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalKmwInput(ws, cell string) {
	_ = g.setValue(ws, cell, 0, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// evalKmwInputEmpty entspricht evalKmwInput (gelbe Eingabezelle, EUR-Format),
// lässt die Zelle aber leer statt mit einer 0 vorzubelegen.
func (g *Generator) evalKmwInputEmpty(ws, cell string) {
	_ = g.setStyle(ws, cell, cell, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalToggle(ws, cell string) {
	_ = g.setValue(ws, cell, "Abzug", StyleOptions{
		Size: 9.0, HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	dv := excelize.NewDataValidation(true)
	dv.Sqref = cell
	dv.SetDropList([]string{"Abzug", "Kein Abzug"})
	_ = g.file.AddDataValidation(ws, dv)
}

func (g *Generator) evalDeduct(ws, cell, formula string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_DEDUCT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalDeductPlaceholder(ws, cell string) {
	_ = g.setValue(ws, cell, 0, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_DEDUCT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// ==================================================================================
// AUSWAHL-PANELS (zentriert, oben in der jeweiligen Sektion)
// ==================================================================================

// evalSelTitle zeichnet die zentrierte Titelzeile einer Auswahlbox (Spalten C..I).
func (g *Generator) evalSelTitle(ws string, row int, text string) {
	_ = g.mergeCells(ws, cellName(EV_PB_C1, row), cellName(EV_PB_C2, row), text, StyleOptions{
		Bold: true, Size: 11.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
	})
	_ = g.file.SetRowHeight(ws, row, 22.0)
}

// evalSelLabel zeichnet die Beschriftungsspalte (C..E) einer Auswahlzeile.
func (g *Generator) evalSelLabel(ws string, row int, text string) {
	_ = g.mergeCells(ws, cellName(EV_PB_C1, row), cellName(EV_PB_L2, row), text, StyleOptions{
		Size: 10.0, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// evalMergedFormula/evalMergedValue: zusammengeführte Zelle, die über den GESAMTEN
// Bereich formatiert wird (verhindert fehlende Rahmenkanten an Folgezellen).
func (g *Generator) evalMergedFormula(ws, c1, c2, formula string, st StyleOptions) {
	_ = g.file.MergeCell(ws, c1, c2)
	_ = g.setStyle(ws, c1, c2, st)
	_ = g.file.SetCellFormula(ws, c1, formula)
}

func (g *Generator) evalMergedValue(ws, c1, c2 string, val interface{}, st StyleOptions) {
	_ = g.file.MergeCell(ws, c1, c2)
	_ = g.setStyle(ws, c1, c2, st)
	_ = g.file.SetCellValue(ws, c1, val)
}

// evalDrawMAPanel zeichnet die zentrierte Mittelanforderungs-Auswahlbox und liefert
// die P-/k-Steuerzellen sowie die letzte belegte Zeile.
func (g *Generator) evalDrawMAPanel(ws string, top int) (string, string, int) {
	num0 := StyleOptions{Bold: true, HAlign: "center", VAlign: "center", NumFormat: "0", FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	inputCtr := StyleOptions{HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}

	r := top
	g.evalSelTitle(ws, r, "Mittelanforderung auswählen")
	r++

	g.evalSelLabel(ws, r, "Auswahl:")
	labelCell := cellName(EV_PB_V1, r)
	g.evalMergedValue(ws, labelCell, cellName(EV_PB_C2, r), "Neuste MA", inputCtr)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = labelCell
	dv.Type = "list"
	dv.Formula1 = "=" + EVAL_NAME_MA_LISTE
	_ = g.file.AddDataValidation(ws, dv)
	r++

	g.evalSelLabel(ws, r, "Geprüfte Periode")
	pCell := cellName(EV_PB_V1, r)
	g.evalMergedValue(ws, pCell, cellName(EV_PB_C2, r), 0, num0) // Formel (=N+1) wird nachgelagert gesetzt
	r++

	g.evalSelLabel(ws, r, "Ausgewählte Anforderung (#)")
	kCell := cellName(EV_PB_V1, r)
	// Höchster Rang der befüllten MAs in der Folgeperiode (=pCell). MAXIFS wäre
	// als Future-Function (_xlfn., implizite Schnittmenge "@") hier unzuverlässig;
	// das SUMPRODUCT(MAX(...))-Idiom nutzt nur Legacy-Funktionen und erzwingt die
	// Array-Auswertung – identisch zum Vorgehen der Spiegel-Panels.
	maxMAK := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(('%s'!%s=%s)*('%s'!%s=1)*'%s'!%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_PERIOD_COUNT), pCell,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_FILL, 1, MA_PERIOD_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT))
	kFormula := fmt.Sprintf(
		`=IF(%s="Neuste MA",%s,IFERROR(VALUE(MID(%s,FIND("(#",%s)+2,FIND(")",%s)-FIND("(#",%s)-2)),0))`,
		labelCell, maxMAK, labelCell, labelCell, labelCell, labelCell)
	g.evalMergedFormula(ws, kCell, cellName(EV_PB_C2, r), kFormula, num0)
	r++

	_ = g.mergeCells(ws, cellName(EV_PB_C1, r), cellName(EV_PB_C2, r), "Einbezogene Anforderungen", StyleOptions{
		Bold: true, Size: 9.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	r++

	// Slots: bei Bedarf (s ≤ k) eingeblendet, sonst leer (Struktur im Hintergrund).
	metaPer := evalAbsCol(EV_DTN_MA_META_PER, 1, MA_PERIOD_COUNT)
	metaRank := evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT)
	metaSumLC := evalAbsCol(EV_DTN_MA_META_SUMLC, 1, MA_PERIOD_COUNT)
	metaSumEU := evalAbsCol(EV_DTN_MA_META_SUMEU, 1, MA_PERIOD_COUNT)
	firstSlot := r
	lblSt := StyleOptions{Size: 9.0, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	lcSt := StyleOptions{HAlign: "right", VAlign: "center", NumFormat: EV_FMT_LC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	euSt := StyleOptions{HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	for s := 1; s <= EV_MA_SLOTS; s++ {
		row := firstSlot + s - 1
		lbl := fmt.Sprintf(`=IF(%d<=%s,"Periode "&%s&" (#%d)","")`, s, kCell, pCell, s)
		g.evalMergedFormula(ws, cellName(EV_PB_C1, row), cellName(EV_PB_L2, row), lbl, lblSt)
		lcF := fmt.Sprintf(`=IF(%d<=%s,IFERROR(SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,%d),0),"")`,
			s, kCell, EVAL_DATEN_SHEET, metaSumLC, EVAL_DATEN_SHEET, metaPer, pCell, EVAL_DATEN_SHEET, metaRank, s)
		g.evalMergedFormula(ws, cellName(EV_PB_V1, row), cellName(EV_PB_SLC2, row), lcF, lcSt)
		euF := fmt.Sprintf(`=IF(%d<=%s,IFERROR(SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,%d),0),"")`,
			s, kCell, EVAL_DATEN_SHEET, metaSumEU, EVAL_DATEN_SHEET, metaPer, pCell, EVAL_DATEN_SHEET, metaRank, s)
		g.evalMergedFormula(ws, cellName(EV_PB_SEU1, row), cellName(EV_PB_C2, row), euF, euSt)

		labelAddr := absName(EV_PB_C1, row)
		cond := fmt.Sprintf(`%s<>""`, labelAddr)
		g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(EV_PB_C1, row), cellName(EV_PB_L2, row)),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: 1, BorderLeft: 2, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
		g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(EV_PB_V1, row), cellName(EV_PB_SLC2, row)),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
		g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(EV_PB_SEU1, row), cellName(EV_PB_C2, row)),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 2, BorderColor: EV_CLR_BORDER,
			})
	}
	bottom := firstSlot + EV_MA_SLOTS - 1
	g.styleOuterBorder(ws, top, EV_PB_C1, bottom, EV_PB_C2, 2, EV_CLR_BORDER)
	return pCell, kCell, bottom
}

// evalDrawFBPanel zeichnet die zentrierte Finanzbericht-Auswahlbox und liefert die
// Periodennummer-Zelle (N) sowie die letzte belegte Zeile.
func (g *Generator) evalDrawFBPanel(ws string, top int) (string, int) {
	num0 := StyleOptions{Bold: true, HAlign: "center", VAlign: "center", NumFormat: "0", FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	inputCtr := StyleOptions{HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	r := top
	g.evalSelTitle(ws, r, "Finanzbericht auswählen")
	r++

	g.evalSelLabel(ws, r, "Auswahl:")
	labelCell := cellName(EV_PB_V1, r)
	g.evalMergedValue(ws, labelCell, cellName(EV_PB_C2, r), "Neuester FB", inputCtr)
	dv := excelize.NewDataValidation(true)
	dv.Sqref = labelCell
	dv.Type = "list"
	dv.Formula1 = "=" + EVAL_NAME_FB_LISTE
	_ = g.file.AddDataValidation(ws, dv)
	r++

	g.evalSelLabel(ws, r, "Geprüfte Periode")
	numCell := cellName(EV_PB_V1, r)
	// Höchste befüllte FB-Periode. SUMPRODUCT(MAX(...)) statt MAXIFS (siehe maxMAK).
	maxFBPer := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(('%s'!%s=1)*'%s'!%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_FB_META_FILL, 1, MA_PERIOD_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_FB_META_PER, 1, MA_PERIOD_COUNT))
	numF := fmt.Sprintf(`=IF(%s="Neuester FB",%s,IF(LEFT(%s,8)="Periode ",IFERROR(VALUE(TRIM(MID(%s,9,5))),0),0))`,
		labelCell, maxFBPer, labelCell, labelCell)
	g.evalMergedFormula(ws, numCell, cellName(EV_PB_C2, r), numF, num0)
	r++

	bottom := r - 1

	g.styleOuterBorder(ws, top, EV_PB_C1, bottom, EV_PB_C2, 2, EV_CLR_BORDER)
	return numCell, bottom
}

// ==================================================================================
// MONATSLIMIT-PRÜFUNG
// ==================================================================================

func (g *Generator) evalDrawMonatslimit(ws string, r int, sel evalSelRefs) int {
	g.evalSectionTitle(ws, r, "Monatslimit-Prüfung")
	r += 2

	// Spalten: Lokalwährung eine nach links gerückt (C), gemeinsame Eingabefelder
	// (Monatsanteile) mittig (D), Euro am bewährten Platz (E).
	const lbl, vLC, vMID, vEUR = EV_COL_LABEL, EV_COL_LABEL + 1, EV_COL_LABEL + 2, EV_COL_LABEL + 3
	top := r

	// Jahresbudget-EUR wird mit dem Budget-Kurs (Gesamtprojekt) umgerechnet.
	rate := BG_NAME_KURS

	gridBorder := func(fill string) StyleOptions {
		return StyleOptions{FillColor: fill, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	}

	// Beschriftungszelle (nur Spalte B).
	label := func(row int, text string, bold bool) {
		_ = g.setValue(ws, cellName(lbl, row), text, StyleOptions{
			Bold: bold, Size: 10.0, HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})
	}
	// Über alle drei Wertspalten (C:E) zusammengeführte, währungsunabhängige Zeile.
	sharedCalc := func(row int, formula, numFmt string) {
		c1, c2 := cellName(vLC, row), cellName(vEUR, row)
		_ = g.file.MergeCell(ws, c1, c2)
		_ = g.setStyle(ws, c1, c2, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: numFmt, FillColor: EV_CLR_CALC,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})
		_ = g.file.SetCellFormula(ws, c1, formula)
	}
	// Mittiges Eingabefeld (Monatsanteil je Jahr, 0..12).
	monthsInput := func(row, val int) string {
		cell := cellName(vMID, row)
		_ = g.setValue(ws, cell, val, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: "0", FillColor: EV_CLR_INPUT,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})
		dv := excelize.NewDataValidation(true)
		dv.Sqref = cell
		_ = dv.SetRange(0, 12, excelize.DataValidationTypeWhole, excelize.DataValidationOperatorBetween)
		_ = g.file.AddDataValidation(ws, dv)
		return absName(vMID, row)
	}
	// Leere (aber gerahmte) Mittelzelle für reine Währungs-Paarzeilen.
	blankMid := func(row int) {
		_ = g.setStyle(ws, cellName(vMID, row), cellName(vMID, row), gridBorder(EV_CLR_CALC))
	}

	// Anforderungssumme (#1..#k der geprüften Periode) – LC bzw. EUR aus der MA-Meta.
	anfSum := func(sumCol int) string {
		return fmt.Sprintf(
			`=IFERROR(ROUND(SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1"),2),0)`,
			EVAL_DATEN_SHEET, evalAbsCol(sumCol, 1, MA_PERIOD_COUNT),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_PERIOD_COUNT), sel.maSelP,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT), sel.maSelK,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT))
	}

	// ─── Spalten-Unterüberschrift: LC links, Monatsanteil mittig, Euro rechts ──
	subHdr := StyleOptions{
		Italic: true, Size: 9.0, FontColor: "595959", FillColor: EV_CLR_HEADER,
		HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	}
	_ = g.setValue(ws, cellName(lbl, r), "", subHdr)
	_ = g.setValue(ws, cellName(vLC, r), "Lokalwährung", subHdr)
	_ = g.setValue(ws, cellName(vMID, r), "Monate (von 12)", subHdr)
	_ = g.setValue(ws, cellName(vEUR, r), "Euro", subHdr)
	r++

	// Anforderungshöhe = Summe der ausgewählten Mittelanforderungen (#1..#k) der
	// geprüften Periode (identisch zur GESAMT-Zeile der Prognoseprüfung).
	rAnf := r
	label(r, "Anforderungshöhe", false)
	g.evalLimitCalc(ws, cellName(vLC, r), anfSum(EV_DTN_MA_META_SUMLC), EV_FMT_LC, false)
	blankMid(r)
	g.evalLimitCalc(ws, cellName(vEUR, r), anfSum(EV_DTN_MA_META_SUMEU), EV_FMT_EUR, false)
	anfLCAddr := absName(vLC, rAnf)
	anfEURAddr := absName(vEUR, rAnf)
	r++

	// ─── Jahresbudget je Haushaltsjahr inkl. einfließendem Monatsanteil ──────
	// Eine Mittelanforderung kann sich über zwei (theoretisch drei) Haushaltsjahre
	// erstrecken. Pro Jahr wird angegeben, wie viele Monate (von 12) in die
	// Berechnung einfließen; das anteilige Limit summiert je Jahr
	// Jahresbudget × Monate / 12.
	yearMonthAddrs := make([]string, len(BG_YEARS))
	yearBudLCAddrs := make([]string, len(BG_YEARS))
	yearBudEURAddrs := make([]string, len(BG_YEARS))
	defaultMonths := []int{8, 0, 0} // bisheriges 8-Monats-Verhalten als Ausgangswert
	for i, year := range BG_YEARS {
		label(r, "Jahresbudget "+year, false)
		budF := fmt.Sprintf("=IFERROR(ROUND(SUBTOTAL(109,%s[%s]),2),0)", BG_TABLE_AUSG, year)
		g.evalLimitCalc(ws, cellName(vLC, r), budF, EV_FMT_LC, false)
		yearBudLCAddrs[i] = absName(vLC, r)
		dm := 0
		if i < len(defaultMonths) {
			dm = defaultMonths[i]
		}
		yearMonthAddrs[i] = monthsInput(r, dm)
		g.evalLimitCalc(ws, cellName(vEUR, r), fmt.Sprintf("=IFERROR(ROUND(%s/%s,2),0)", yearBudLCAddrs[i], rate), EV_FMT_EUR, false)
		yearBudEURAddrs[i] = absName(vEUR, r)
		r++
	}

	// Limit-Monate = Summe der eingerechneten Monatsanteile über alle Jahre.
	label(r, "Limit-Monate (Summe Anteile)", false)
	sharedCalc(r, fmt.Sprintf("=%s", strings.Join(yearMonthAddrs, "+")), "0")
	r++

	// Bezugszeitraum = Zeitraum (Monate) der gewählten Mittelanforderung (übernommen,
	// nicht editierbar). Dient als Kontrolle gegen die Summe der Monatsanteile.
	label(r, "Bezugszeitraum (Monate, aus MA)", false)
	sharedCalc(r, evalMASelectedZeitraum(sel), "0")
	r++

	// Anteiliges Monatslimit = Σ Jahresbudget_i × Monate_i / 12.
	rLimit := r
	limTermsLC := make([]string, len(BG_YEARS))
	limTermsEUR := make([]string, len(BG_YEARS))
	for i := range BG_YEARS {
		limTermsLC[i] = fmt.Sprintf("%s*%s/12", yearBudLCAddrs[i], yearMonthAddrs[i])
		limTermsEUR[i] = fmt.Sprintf("%s*%s/12", yearBudEURAddrs[i], yearMonthAddrs[i])
	}
	label(r, "Monatslimit (anteilig)", true)
	g.evalLimitCalc(ws, cellName(vLC, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", strings.Join(limTermsLC, "+")), EV_FMT_LC, true)
	blankMid(r)
	g.evalLimitCalc(ws, cellName(vEUR, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", strings.Join(limTermsEUR, "+")), EV_FMT_EUR, true)
	limitLCAddr := absName(vLC, rLimit)
	limitEURAddr := absName(vEUR, rLimit)
	r++

	label(r, "Status", true)
	g.evalLimitStatus(ws, vLC, r, anfLCAddr, limitLCAddr)
	blankMid(r)
	g.evalLimitStatus(ws, vEUR, r, anfEURAddr, limitEURAddr)
	r++

	label(r, "Überschreitung", false)
	g.evalLimitUeber(ws, vLC, r, anfLCAddr, limitLCAddr, EV_FMT_LC)
	blankMid(r)
	g.evalLimitUeber(ws, vEUR, r, anfEURAddr, limitEURAddr, EV_FMT_EUR)
	r++

	label(r, "Auslastung %", false)
	g.evalLimitCalc(ws, cellName(vLC, r), fmt.Sprintf("=IFERROR(%s/%s,0)", anfLCAddr, limitLCAddr), EV_FMT_PCT, false)
	blankMid(r)
	g.evalLimitCalc(ws, cellName(vEUR, r), fmt.Sprintf("=IFERROR(%s/%s,0)", anfEURAddr, limitEURAddr), EV_FMT_PCT, false)
	bottom := r

	g.styleOuterBorder(ws, top, lbl, bottom, vEUR, 2, EV_CLR_BORDER)
	return bottom
}

// evalLimitStatus zeichnet eine Status-Zelle (OK/ÜBERSCHRITTEN) mit bedingter
// Formatierung. Die bedingte Formatierung setzt bewusst KEINE eigenen Rahmen –
// dadurch bleibt die korrekte Zellkante erhalten (innen dünn, am Box-Rand kräftig
// durch styleOuterBorder) und passt sich der Umgebung an.
func (g *Generator) evalLimitStatus(ws string, col, row int, anfAddr, limitAddr string) {
	cell := cellName(col, row)
	self := absName(col, row)
	_ = g.setFormula(ws, cell, fmt.Sprintf(`=IF(%s<=%s,"OK","ÜBERSCHRITTEN")`, anfAddr, limitAddr), StyleOptions{
		Bold: true, HAlign: "center", VAlign: "center", FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	g.addConditionalFormat(ws, cell, fmt.Sprintf(`%s="OK"`, self), StyleOptions{
		Bold: true, FontColor: EV_CLR_GOOD_TXT, FillColor: EV_CLR_GOOD, HAlign: "center", VAlign: "center",
	})
	g.addConditionalFormat(ws, cell, fmt.Sprintf(`%s<>"OK"`, self), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD, HAlign: "center", VAlign: "center",
	})
}

// evalLimitUeber zeichnet eine Überschreitungs-Zelle (rot ab >0). Wie evalLimitStatus
// ohne eigene Rahmen in der bedingten Formatierung.
func (g *Generator) evalLimitUeber(ws string, col, row int, anfAddr, limitAddr, numFmt string) {
	cell := cellName(col, row)
	self := absName(col, row)
	g.evalLimitCalc(ws, cell, fmt.Sprintf("=ROUND(MAX(0,%s-%s),2)", anfAddr, limitAddr), numFmt, false)
	g.addConditionalFormat(ws, cell, fmt.Sprintf(`%s>0`, self), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD, HAlign: "right", VAlign: "center", NumFormat: numFmt,
	})
}

func (g *Generator) evalLimitCalc(ws, cell, formula, numFmt string, bold bool) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		Bold: bold, HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: EV_CLR_CALC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// ==================================================================================
// VERGLEICHSTABELLEN
// ==================================================================================

func (g *Generator) evalDrawComparisonTable(ws string, r int, title string, isIncome, isMA bool, sel evalSelRefs) evalCompResult {
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

		_ = g.setValue(ws, cellName(EV_COL_LABEL, row), name, StyleOptions{
			HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})

		actLC := cellName(EV_COL_ACT_LC, row)
		actEUR := cellName(EV_COL_ACT_EUR, row)
		lcF, eurF := g.evalActualFormulas(isIncome, isMA, name, i, sel)
		g.evalCellFormula(ws, actLC, lcF, EV_FMT_LC, EV_CLR_CALC)
		g.evalCellFormula(ws, actEUR, eurF, EV_FMT_EUR, EV_CLR_CALC)

		budLCName, budEURName := g.evalBudgetNames(isIncome, name)
		g.evalCellFormula(ws, cellName(EV_COL_BUD_LC, row), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", budLCName), EV_FMT_LC, EV_CLR_CALC)
		g.evalCellFormula(ws, cellName(EV_COL_BUD_EUR, row), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", budEURName), EV_FMT_EUR, EV_CLR_CALC)

		g.evalCellFormula(ws, cellName(EV_COL_DIF_LC, row),
			fmt.Sprintf("=ROUND(IFERROR(%s-%s,0),2)", absName(EV_COL_ACT_LC, row), absName(EV_COL_BUD_LC, row)), EV_FMT_LC, "")
		g.evalCellFormula(ws, cellName(EV_COL_ABW_LC, row),
			fmt.Sprintf("=ROUND(IFERROR((%s/%s)-1,0),4)", absName(EV_COL_ACT_LC, row), absName(EV_COL_BUD_LC, row)), EV_FMT_PCT, "")
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
	totalRow := dataEnd + 1
	g.evalTotalRow(ws, totalRow, dataStart, dataEnd)

	// Abweichungs-Ampel auf beiden Währungsspalten (Einnahmen wie Ausgaben):
	// ≥ 20 % rot, 10–20 % gelb.
	g.evalDeviationConditional(ws, EV_COL_ABW_LC, dataStart, dataEnd)
	g.evalDeviationConditional(ws, EV_COL_ABW_EUR, dataStart, dataEnd)

	g.styleOuterBorder(ws, hdrRow, EV_COL_LABEL, totalRow, EV_COL_ABW_EUR, 2, EV_CLR_BORDER)

	return evalCompResult{
		nextRow:     totalRow,
		actEURRange: fmt.Sprintf("%s:%s", absName(EV_COL_ACT_EUR, dataStart), absName(EV_COL_ACT_EUR, dataEnd)),
		budEURRange: fmt.Sprintf("%s:%s", absName(EV_COL_BUD_EUR, dataStart), absName(EV_COL_BUD_EUR, dataEnd)),
		kmwActEUR:   kmwActEUR,
		kmwBudEUR:   kmwBudEUR,
	}
}

// evalActualFormulas: Ist-/Prognose-Formeln je Zeile.
func (g *Generator) evalActualFormulas(isIncome, isMA bool, name string, idx int, sel evalSelRefs) (string, string) {
	if isIncome && isMA {
		// Prognose der Finanzierungsanteile = Summe der ausgewählten Mittelanforderungen
		// (#1..#k) der Periode P je Finanzierungsart (Eigenmittel/Drittmittel/KMW-Mittel)
		// aus dem MA-Grid. "Zinsertraege" hat keine MA-Quelle ⇒ SUMIFS ergibt 0.
		return evalMAExpenseActual(sel, name, EV_DTN_MAG_LC), evalMAExpenseActual(sel, name, EV_DTN_MAG_EUR)
	}
	if isIncome {
		// FB: kumulative Einnahmen je Typ bis zur gewählten Periode N (CHOOSE über die
		// Kum-Spalten der Finanzberichte). Einnahmen-Typzeilen liegen bei 12..15.
		return evalFBChooseRef(sel.fbSelNum, 12+idx, 3), evalFBChooseRef(sel.fbSelNum, 12+idx, 4)
	}
	if isMA {
		// Prognose der Ausgaben = Summe der ausgewählten Mittelanforderungen (#1..#k)
		// der Periode P, je Kategorie (MA-Grid auf dem Daten-Blatt).
		return evalMAExpenseActual(sel, name, EV_DTN_MAG_LC), evalMAExpenseActual(sel, name, EV_DTN_MAG_EUR)
	}
	// FB: kumulative Ausgaben je Kategorie bis Periode N. Die FB-Ausgabentabellen
	// sind positionsbasiert (eine Zeile je Kostenposition); für die kategorienweise
	// Auswertung werden alle zur Kategorie gehörenden Positionszeilen aufsummiert.
	rows := g.fbExpenseRowsForCategory(name)
	return evalFBChooseRefRows(sel.fbSelNum, rows, 3), evalFBChooseRefRows(sel.fbSelNum, rows, 4)
}

func (g *Generator) evalBudgetNames(isIncome bool, name string) (string, string) {
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

func (g *Generator) evalDeviationConditional(ws string, col, dataStart, dataEnd int) {
	rng := fmt.Sprintf("%s:%s", cellName(col, dataStart), cellName(col, dataEnd))
	topRel := cellName(col, dataStart)
	// Bewusst OHNE eigene Rahmen: die bedingte Formatierung überschreibt sonst die
	// vorhandene Zellkante (innen dünnes Gitter, an der Box rechts kräftig durch
	// styleOuterBorder). Nur Schrift/Füllung/Format setzen – wie evalLimitUeber.
	g.addConditionalFormat(ws, rng, fmt.Sprintf("%s>=0.2", topRel), StyleOptions{
		Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD, HAlign: "right", VAlign: "center", NumFormat: EV_FMT_PCT,
	})
	g.addConditionalFormat(ws, rng, fmt.Sprintf("AND(%s>=0.1,%s<0.2)", topRel, topRel), StyleOptions{
		Bold: true, FontColor: EV_CLR_WARN_TXT, FillColor: EV_CLR_WARN, HAlign: "right", VAlign: "center", NumFormat: EV_FMT_PCT,
	})
}

func (g *Generator) evalCellFormula(ws, cell, formula, numFmt, fill string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: fill,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

// ==================================================================================
// FORMEL-BAUSTEINE
// ==================================================================================

// evalFBChooseRef referenziert die kumulative Kum-Zelle des gewählten Finanzberichts
// (Periode N) per CHOOSE über alle 18 Perioden (nicht-volatil, robust).
func evalFBChooseRef(selNum string, baseRow, colOffset int) string {
	parts := make([]string, 0, 18)
	for p := 1; p <= 18; p++ {
		col := 2 + (p-1)*7 + colOffset
		parts = append(parts, fmt.Sprintf("'%s'!%s", EVAL_FB_SHEET, absName(col, baseRow)))
	}
	return fmt.Sprintf(`=IFERROR(ROUND(CHOOSE(%s,%s),2),0)`, selNum, strings.Join(parts, ","))
}

// evalFBChooseRefRows summiert je Finanzbericht-Periode die übergebenen Zeilen
// (eine je Kostenposition einer Kategorie) der Kum-Spalte (colOffset 3=LC, 4=EUR)
// und wählt die geprüfte Periode N per CHOOSE. So bleibt die Auswertung
// kategorienbasiert, obwohl die FB-Tabellen positionsbasiert sind.
func evalFBChooseRefRows(selNum string, rows []int, colOffset int) string {
	if len(rows) == 0 {
		return "=0"
	}
	parts := make([]string, 0, 18)
	for p := 1; p <= 18; p++ {
		colBase := 2 + (p-1)*7 + colOffset
		cellRefs := make([]string, 0, len(rows))
		for _, br := range rows {
			cellRefs = append(cellRefs, fmt.Sprintf("'%s'!%s", EVAL_FB_SHEET, absName(colBase, br)))
		}
		parts = append(parts, strings.Join(cellRefs, "+"))
	}
	return fmt.Sprintf(`=IFERROR(ROUND(CHOOSE(%s,%s),2),0)`, selNum, strings.Join(parts, ","))
}

// evalMAExpenseActual summiert die ausgewählten Mittelanforderungen (#1..#k) der
// Periode P je Kategorie über das MA-Grid auf dem Daten-Blatt.
func evalMAExpenseActual(sel evalSelRefs, cat string, valCol int) string {
	return fmt.Sprintf(
		`=IFERROR(ROUND(SUMIFS('%s'!%s,'%s'!%s,"%s",'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1"),2),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(valCol, 1, EV_DTN_MAG_ROWS),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, EV_DTN_MAG_ROWS), cat,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, EV_DTN_MAG_ROWS), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, EV_DTN_MAG_ROWS), sel.maSelK,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, EV_DTN_MAG_ROWS))
}

// evalMAChooseManBetrag liefert einen CHOOSE-Ausdruck über die MA_ManBetrag_<n>-Namen
// aller 18 Perioden, gewählt per maSelP – analog zu evalMAChooseKurs.
func evalMAChooseManBetrag(maSelP string) string {
	parts := make([]string, MA_PERIOD_COUNT)
	for i := range parts {
		parts[i] = fmt.Sprintf("MA_ManBetrag_%d", i+1)
	}
	return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),0)`, maSelP, strings.Join(parts, ","))
}

// evalMAChooseKurs liefert einen nicht-volatilen CHOOSE-Ausdruck über alle 18
// MA_Kurs_<n>-Namen. Ersatz für INDIRECT("MA_Kurs_"&maSelP), das Excel 365 mit
// dem @-Operator (implizite Schnittmenge) versieht und dadurch falsch rendert.
func evalMAChooseKurs(maSelP string) string {
	parts := make([]string, MA_PERIOD_COUNT)
	for i := range parts {
		parts[i] = fmt.Sprintf("MA_Kurs_%d", i+1)
	}
	return fmt.Sprintf(`IFERROR(CHOOSE(%s,%s),0)`, maSelP, strings.Join(parts, ","))
}

// evalMACurrentKMWRequest summiert die KMW-Mittel-Anforderung der EINEN aktuell
// gewählten Mittelanforderung (Periode P, Rang exakt = k) aus dem MA-Grid. Anders
// als die Prognose (#1..#k) bewusst nicht zusammengesetzt – frühere Anforderungen
// einer Periode sind bereits über die bereitgestellten KMW-Mittel erfasst.
func evalMACurrentKMWRequest(sel evalSelRefs) string {
	return fmt.Sprintf(
		`=IFERROR(ROUND(SUMIFS('%s'!%s,'%s'!%s,"KMW-Mittel",'%s'!%s,%s,'%s'!%s,%s),2),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_EUR, 1, EV_DTN_MAG_ROWS),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, EV_DTN_MAG_ROWS),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, EV_DTN_MAG_ROWS), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, EV_DTN_MAG_ROWS), sel.maSelK)
}

// evalMASelectedZeitraum liefert den Zeitraum (Monate) der aktuell gewählten
// Mittelanforderung (Periode P, Rang k). Es wird – wie im Spiegel-Panel – der
// Tabellenindex j der gewählten MA bestimmt und der Zeitraum (MA-Quellzeile 7,
// Wertspalte = colS+1) per CHOOSE über alle 18 MA-Tabellen ausgelesen.
func evalMASelectedZeitraum(sel evalSelRefs) string {
	j := fmt.Sprintf(`IFERROR(SUMPRODUCT('%s'!%s,('%s'!%s=%s)*('%s'!%s=%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_J, 1, MA_PERIOD_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_PERIOD_COUNT), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_PERIOD_COUNT), sel.maSelK)
	parts := make([]string, 0, MA_PERIOD_COUNT)
	for t := 1; t <= MA_PERIOD_COUNT; t++ {
		colS := MA_START_COL + (t-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
		parts = append(parts, fmt.Sprintf("'%s'!%s", MA_SHEET_NAME, absName(colS+1, 7)))
	}
	return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),0)`, j, strings.Join(parts, ","))
}

// evalAbsCol liefert einen absoluten Spaltenbereich, z. B. "$BU$1:$BU$144".
func evalAbsCol(col, r1, r2 int) string {
	return fmt.Sprintf("$%s$%d:$%s$%d", colLetter(col), r1, colLetter(col), r2)
}
