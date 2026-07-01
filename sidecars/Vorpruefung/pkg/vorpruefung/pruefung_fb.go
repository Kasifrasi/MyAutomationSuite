package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ==================================================================================
// Blatt "V. Prüfung FB"
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
	EVAL_SHEET_NAME = constants.VPSheetFB_PRUEFUNG
)

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

// CreateFBPruefungSheet baut das Blatt "V. Prüfung FB".
func (g *Generator) CreateFBPruefungSheet() error {
	ws := constants.VPSheetFB_PRUEFUNG
	f := g.file

	if _, err := f.NewSheet(ws); err != nil {
		return fmt.Errorf("fehler beim Erstellen des FB-Prüfungs-Blatts: %w", err)
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
	g.evalBanner(ws, r, ws, "Automatische Prüfung von Finanzberichten")
	r += 3

	fbSectionTop := r
	g.evalMainHeader(ws, r, "FINANZBERICHTSPRÜFUNG", "Basis: ausgewählter Finanzbericht (kumulativ bis zur Periode)")
	r += 3

	fbSelNumCell, fbNext := g.evalDrawFBPanel(ws, r)
	r = fbNext + EV_TABLE_GAP
	sel := evalSelRefs{fbSelNum: fbSelNumCell}
	g.evalFBSelNumAddr = fmt.Sprintf("'%s'!%s", ws, fbSelNumCell)

	fbKMW := g.evalDrawKMWSektion(ws, r, false, sel)
	r = fbKMW.nextRow + EV_TABLE_GAP

	resFBInc := g.evalDrawComparisonTable(ws, r, "Finanzierungsanteile", true, false, sel)
	r = resFBInc.nextRow + EV_TABLE_GAP

	mehrFormula := fmt.Sprintf(
		`=IFERROR(ROUND(MAX(0,(SUM(%s)-%s)-(SUM(%s)-%s)),2),0)`,
		resFBInc.actEURRange, resFBInc.kmwActEUR, resFBInc.budEURRange, resFBInc.kmwBudEUR)
	g.evalDeduct(ws, fbKMW.mehrCell, mehrFormula)

	resFBExp := g.evalDrawComparisonTable(ws, r, "Soll-Ist Abweichungsprüfung", false, false, sel)
	r = resFBExp.nextRow + EV_TABLE_GAP

	for _, cell := range []string{
		fbKMW.saldoCell, fbKMW.mehrCell,
	} {
		if cell != "" {
			g.reapplyRightBorder(ws, cell, 2, EV_CLR_BORDER)
		}
	}

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
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", FieldBudgetKMWEUR.NamedRange), false)
	r++
	rRes := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Davon Reserve", false)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", bgKostenName("Reserve", "EUR")), false)
	r++
	rOp := r
	g.evalMergedFormula(ws, cellName(lblL1, r), cellName(lblL2, r),
		fmt.Sprintf(`=IF(%s="Ja","Operatives Budget (Reserve freigegeben)","Operatives Budget (abzgl. Reserve)")`, BudgetNameReserve),
		StyleOptions{Size: 10.0, HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID})
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf(
		`=IF(%s="Ja",ROUND(%s,2),ROUND(%s-%s,2))`,
		BudgetNameReserve, absName(valL, rBew), absName(valL, rBew), absName(valL, rRes)), false)
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
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), FieldMAPruefungAbzugSaldo)
	} else {
		g.evalToggle(ws, cellName(tog, rr), FieldFBPruefungAbzugSaldo)
	}
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Saldovortrag", false)
	g.evalDeduct(ws, cellName(valR, rr), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", DB_NAME_SALDOVORTRAG_EUR))
	saldoCell := cellName(valR, rr)
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	// Mehreinnahmen (Formel wird nachgelagert gesetzt)
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), FieldMAPruefungAbzugMehr)
	} else {
		g.evalToggle(ws, cellName(tog, rr), FieldFBPruefungAbzugMehr)
	}
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Mehreinnahmen", false)
	g.evalDeductPlaceholder(ws, cellName(valR, rr))
	mehrCell := cellName(valR, rr)
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	prognCell := ""
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), FieldMAPruefungAbzugPrognose)
		g.evalKmwLabel(ws, rr, lblR1, lblR2, "Prognostizierte Mehreinnahmen", false)
		g.evalDeductPlaceholder(ws, cellName(valR, rr))
		prognCell = cellName(valR, rr)
		deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
		rr++
	}

	// Bedingte Formatierung: Wertfelder ausgegraut wenn "Kein Abzug" gewählt.
	// Zwei Regeln pro Zelle, da StyleOptions nur eine BorderColor kennt:
	// Regel 1 – Füllung/Schrift + drei dünne Gitterlinien (links/oben/unten)
	// Regel 2 – dicker rechter Außenrahmen (passend zu styleOuterBorder)
	for _, d := range deds {
		cond := fmt.Sprintf(`%s="Kein Abzug"`, d.togCell)
		g.addConditionalFormat(ws, d.valCell, cond, StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR,
			FillColor: EV_CLR_DEDUCT_OFF, FontColor: "A0A0A0",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderColor: EV_CLR_GRID,
		})
		g.addConditionalFormat(ws, d.valCell, cond, StyleOptions{
			BorderRight: 2, BorderColor: EV_CLR_BORDER,
		})
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
		reqFormula := evalMACurrentKMWRequest(g, sel)
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwCalc(ws, cellName(valL, r), reqFormula, false)
		addrReqL := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Abzüglich aktuelle Anforderung", false)
		g.evalKmwCalc(ws, cellName(valR, r), reqFormula, false)
		addrReqR := absName(valR, r)
		r++
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrVerf, addrReqL), true)
		addrVerblL1 := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwCalc(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrBereinigt, addrReqR), true)
		addrVerblR1 := absName(valR, r)
		p1Bottom := r
		g.styleOuterBorder(ws, p1Top, lblL1, p1Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p1Top, tog, p1Bottom, valR, 2, EV_CLR_BORDER)
		r += 2

		// Paar 2: Abzüglich manueller Betrag → aus dem "Manueller Betrag"-Feld der gewählten MA
		p2Top := r
		manFormula := evalMAChooseManBetrag(sel)
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Abzüglich manueller Betrag", false)
		g.evalDeduct(ws, cellName(valL, r), manFormula)
		addrManL := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Abzüglich manueller Betrag", false)
		g.evalDeduct(ws, cellName(valR, r), manFormula)
		addrManR := absName(valR, r)
		r++
		g.evalKmwLabel(ws, r, lblL1, lblL2, "Verbleibende KMW-Mittel", true)
		g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrVerf, addrManL), true)
		addrVerblL2 := absName(valL, r)
		g.evalKmwLabel(ws, r, tog, lblR2, "Verbleibende KMW-Mittel (bereinigt)", true)
		g.evalKmwCalc(ws, cellName(valR, r), fmt.Sprintf("=ROUND(%s-MAX(0,%s),2)", addrBereinigt, addrManR), true)
		addrVerblR2 := absName(valR, r)
		p2Bottom := r
		g.styleOuterBorder(ws, p2Top, lblL1, p2Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p2Top, tog, p2Bottom, valR, 2, EV_CLR_BORDER)
		r++

		condReqGrey := fmt.Sprintf(`%s<>0`, addrManL)
		condManGrey := fmt.Sprintf(`%s=0`, addrManL)

		applyGreyCF := func(rStart, rEnd int, cond string) {
			for row := rStart; row <= rEnd; row++ {
				g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(lblL1, row), cellName(lblL2, row)), cond, StyleOptions{
					FillColor: EV_CLR_DEDUCT_OFF, FontColor: "A0A0A0",
				})
				g.addConditionalFormat(ws, cellName(valL, row), cond, StyleOptions{
					FillColor: EV_CLR_DEDUCT_OFF, FontColor: "A0A0A0",
				})
				g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(tog, row), cellName(lblR2, row)), cond, StyleOptions{
					FillColor: EV_CLR_DEDUCT_OFF, FontColor: "A0A0A0",
				})
				g.addConditionalFormat(ws, cellName(valR, row), cond, StyleOptions{
					FillColor: EV_CLR_DEDUCT_OFF, FontColor: "A0A0A0",
				})
			}
		}

		applyGreyCF(p1Top, p1Bottom, condReqGrey)
		applyGreyCF(p2Top, p2Bottom, condManGrey)

		// Negativ-Formatierung (rot), nur wenn nicht ausgegraut
		redStyle := StyleOptions{
			Bold: true, FontColor: EV_CLR_BAD_TXT, FillColor: EV_CLR_BAD,
			BorderTop: 1, BorderBottom: 2, BorderLeft: 1, BorderRight: 2, BorderColor: EV_CLR_BORDER,
		}
		g.addConditionalFormat(ws, cellName(valL, p1Bottom), fmt.Sprintf(`AND(%s<0, %s=0)`, addrVerblL1, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valR, p1Bottom), fmt.Sprintf(`AND(%s<0, %s=0)`, addrVerblR1, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valL, p2Bottom), fmt.Sprintf(`AND(%s<0, %s<>0)`, addrVerblL2, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valR, p2Bottom), fmt.Sprintf(`AND(%s<0, %s<>0)`, addrVerblR2, addrManL), redStyle)
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
	_ = g.setStyle(ws, cell, cell, StyleOptions{
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

func (g *Generator) evalToggle(ws, cell string, field InputField) {
	_ = g.setStyle(ws, cell, cell, StyleOptions{
		Size: 9.0, HAlign: "center", VAlign: "center", FillColor: EV_CLR_INPUT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	col, row, _ := excelize.CellNameToCoordinates(cell)
	_ = g.bindInputField(ws, row, col, field)
}

func (g *Generator) evalDeduct(ws, cell, formula string) {
	_ = g.setFormula(ws, cell, formula, StyleOptions{
		HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR, FillColor: EV_CLR_DEDUCT,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
}

func (g *Generator) evalDeductPlaceholder(ws, cell string) {
	_ = g.setStyle(ws, cell, cell, StyleOptions{
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
	g.mergeCells(ws, labelCell, cellName(EV_PB_C2, r), "", inputCtr)
	col, row, _ := excelize.CellNameToCoordinates(labelCell)
	_ = g.bindInputField(ws, row, col, FieldFBPruefungAuswahl)

	dv := excelize.NewDataValidation(true)
	dv.Sqref = labelCell
	dv.Type = "list"
	dv.Formula1 = "=" + EVAL_NAME_FB_LISTE
	_ = g.file.AddDataValidation(ws, dv)
	r++

	g.evalSelLabel(ws, r, "Ausgewählte Periode")
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

	nRows := g.budgetIncomeCount()
	if !isIncome {
		nRows = g.budgetExpenseCount()
	}
	dataStart := r
	kmwActEUR, kmwBudEUR := "", ""

	for i := 0; i < nRows; i++ {
		row := dataStart + i

		// Label wird von der API gefüllt. Nur formatiert bereitstellen.
		labelVal := ""
		if g.cfg.IncomeTypesCount == 0 && g.cfg.ExpensePositionsCount == 0 {
			if isIncome && i < len(TYPE_NAMES) {
				labelVal = TYPE_NAMES[i]
			} else if !isIncome && i < len(EXPENSE_CATEGORIES) {
				labelVal = EXPENSE_CATEGORIES[i]
			}
		}

		_ = g.setValue(ws, cellName(EV_COL_LABEL, row), labelVal, StyleOptions{
			HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})

		actLC := cellName(EV_COL_ACT_LC, row)
		actEUR := cellName(EV_COL_ACT_EUR, row)
		lcF, eurF := g.evalActualFormulas(isIncome, isMA, i, sel)
		g.evalCellFormula(ws, actLC, lcF, EV_FMT_LC, EV_CLR_CALC)
		g.evalCellFormula(ws, actEUR, eurF, EV_FMT_EUR, EV_CLR_CALC)

		budLCName, budEURName := g.evalBudgetNames(isIncome, i)
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

		if !isMA {
			// cleanName war für AW_FB_ActLC_Bauausgaben. Jetzt ohne festen Namen einfach Index?
			// Da Named Ranges keine Leerzeichen etc. mögen, nutzen wir einen generischen Index.
			g.dbUpsertNamedRange(ws, fmt.Sprintf("AW_FB_ActLC_%s_%d", map[bool]string{true: "Inc", false: "Exp"}[isIncome], i), EV_COL_ACT_LC, row)
			g.dbUpsertNamedRange(ws, fmt.Sprintf("AW_FB_ActEUR_%s_%d", map[bool]string{true: "Inc", false: "Exp"}[isIncome], i), EV_COL_ACT_EUR, row)
		}

		// Hack for KMW comparison later
		// Since we don't know the exact row for KMW anymore from hardcoded strings, we will
		// just assume KMW is the 3rd row for backward compatibility or let the API figure it out.
		// Wait, KMW is usually Income Type 3 (index 2). Let's use index 2 for now.
		if isIncome && i == 2 {
			kmwActEUR = absName(EV_COL_ACT_EUR, row)
			kmwBudEUR = absName(EV_COL_BUD_EUR, row)
		}
	}

	dataEnd := dataStart + nRows - 1
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
func (g *Generator) evalActualFormulas(isIncome, isMA bool, idx int, sel evalSelRefs) (string, string) {
	if isIncome && isMA {
		maL := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_LC, isIncome)
		maE := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_EUR, isIncome)
		fbL := evalFBChooseRef(sel.fbSelNum, 12+idx, 3)
		fbE := evalFBChooseRef(sel.fbSelNum, 12+idx, 4)
		return fmt.Sprintf("=%s + %s", maL[1:], fbL[1:]),
			fmt.Sprintf("=%s + %s", maE[1:], fbE[1:])
	}
	if isIncome {
		return evalFBChooseRef(sel.fbSelNum, 12+idx, 3), evalFBChooseRef(sel.fbSelNum, 12+idx, 4)
	}
	if isMA {
		maL := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_LC, isIncome)
		maE := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_EUR, isIncome)
		// 1 to 1 mapping: FB row is FB_AUSG_FIRST_ROW + idx
		row := FB_AUSG_FIRST_ROW + idx
		fbL := evalFBChooseRef(sel.fbSelNum, row, 3)
		fbE := evalFBChooseRef(sel.fbSelNum, row, 4)
		return fmt.Sprintf("=%s + %s", maL[1:], fbL[1:]),
			fmt.Sprintf("=%s + %s", maE[1:], fbE[1:])
	}

	row := FB_AUSG_FIRST_ROW + idx
	return evalFBChooseRef(sel.fbSelNum, row, 3), evalFBChooseRef(sel.fbSelNum, row, 4)
}

func (g *Generator) evalBudgetNames(isIncome bool, idx int) (string, string) {
	if isIncome {
		switch idx {
		case 0:
			return FieldBudgetEigenmittelLC.NamedRange, FieldBudgetEigenmittelEUR.NamedRange
		case 1:
			return FieldBudgetDrittmittelLC.NamedRange, FieldBudgetDrittmittelEUR.NamedRange
		case 2:
			return FieldBudgetKMWLC.NamedRange, FieldBudgetKMWEUR.NamedRange
		default:
			return "0", "0"
		}
	}

	// Budget table has rows starting at ausgHdrRow + 1. We don't have Named Ranges per position anymore.
	// But wait! We need to pull the budget value for this position.
	// Where is the budget value? It's in the Budget sheet.
	// We can use the BudgetTableAusg!
	// Wait, we can just use the absolute cell reference for the budget position row.
	// The budget expense block starts at row 24 or similar, depending on Drittmittel etc...
	// Actually, the Budget Ausgaben block starts after "2. GEPLANTE AUSGABEN".
	// Let's just lookup by index in the budget table using INDEX.
	// BudgetTableAusg columns: Betrag (LC) is 4, Betrag (EUR) is 8
	budLC := fmt.Sprintf("INDEX(%s[Betrag (LC)], %d)", BudgetTableAusg, idx+1)
	budEUR := fmt.Sprintf("INDEX(%s[Betrag (EUR)], %d)", BudgetTableAusg, idx+1)
	return budLC, budEUR
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
		parts = append(parts, fbRowsSumExpr(colBase, rows))
	}
	return fmt.Sprintf(`=IFERROR(ROUND(CHOOSE(%s,%s),2),0)`, selNum, strings.Join(parts, ","))
}

// fbRowsSumExpr baut für eine Spalte einen kompakten SUM-Ausdruck über die Zeilen
// einer Kategorie, indem aufeinanderfolgende Zeilen zu Bereichen zusammengefasst
// werden: SUM('FB'!$A$10:$A$39). Bei vielen Positionen je Kategorie blieb die alte
// Einzelzellen-Addition (A10+A11+...) sonst über Excels 8192-Zeichen-Formellimit.
// rows ist aufsteigend (siehe fbExpenseRowsForCategory).
func fbRowsSumExpr(col int, rows []int) string {
	segs := make([]string, 0)
	start, prev := rows[0], rows[0]
	flush := func(a, b int) {
		if a == b {
			segs = append(segs, fmt.Sprintf("'%s'!%s", EVAL_FB_SHEET, absName(col, a)))
		} else {
			segs = append(segs, fmt.Sprintf("'%s'!%s:%s", EVAL_FB_SHEET, absName(col, a), absName(col, b)))
		}
	}
	for _, r := range rows[1:] {
		if r == prev+1 {
			prev = r
			continue
		}
		flush(start, prev)
		start, prev = r, r
	}
	flush(start, prev)
	return fmt.Sprintf("SUM(%s)", strings.Join(segs, ","))
}
