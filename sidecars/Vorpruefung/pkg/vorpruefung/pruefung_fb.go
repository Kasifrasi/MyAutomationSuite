package vorpruefung

import (
	"fmt"
	"reflect"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ==================================================================================
// Blatt "V. Prüfung FB"
//
//	FINANZBERICHTSPRÜFUNG     (Basis: ausgewählter Finanzbericht / kumulativ)
//
// ==================================================================================
//
// Teil A (Grid-Konstanten): Die Spalten der Vergleichstabellen (EV_COL_*), das
// Auswahl-Panel (EV_PB_*) sowie Farben/Formate (EV_CLR_*, EV_FMT_*) sind zentral
// in pruefung_shared.go definiert; die Styles liegen in styles.go (EV*Style).
//
// Teil B (Layout): Der Aufbau ist rein vertikal (r läuft von oben nach unten):
//   Banner → FINANZBERICHTSPRÜFUNG-Kopf → FB-Auswahl-Panel → KMW-Mittelprüfung →
//   Finanzierungsanteile (4 Einnahme-Kategorien) → Soll-Ist-Abweichungsprüfung
//   (8 Kostenkategorien). Rechts daneben spiegelt das FB-Panel (pruefung_fb_panel.go)
//   den gewählten Finanzbericht.
//
// Kategorien (wie in registry.go definiert):
//   Finanzierungsanteile → TYPE_NAMES (Eigenmittel, Drittmittel, KMW-Mittel, Zinsertraege)
//   Soll-Ist-Abweichung  → EXPENSE_CATEGORIES (Bauausgaben … Reserve)
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

	finBind := evalCompBindingFor(Registry, "OutputFBPruefungFin", []string{"EM", "DM", "KMW", "Zins"})
	resFBInc := g.evalDrawComparisonTable(ws, r, "Finanzierungsanteile", true, false, sel, finBind)
	r = resFBInc.nextRow + EV_TABLE_GAP

	mehrFormula := fmt.Sprintf(
		`=IFERROR(ROUND(MAX(0,(SUM(%s)-%s)-(SUM(%s)-%s)),2),0)`,
		resFBInc.actEURRange, resFBInc.kmwActEUR, resFBInc.budEURRange, resFBInc.kmwBudEUR)
	g.evalDeduct(ws, fbKMW.mehrCell, mehrFormula)

	sollIstBind := evalCompBindingFor(Registry, "OutputFBPruefungSollIst", []string{"Bau", "Inv", "Pers", "Aktiv", "Verw", "Eval", "Audit", "Reserve"})
	resFBExp := g.evalDrawComparisonTable(ws, r, "Soll-Ist Abweichungsprüfung", false, false, sel, sollIstBind)
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
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, EVBannerTitleStyle)
	_ = g.file.SetRowHeight(ws, row, 30.0)
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, EVBannerSubStyle)
	_ = g.file.SetRowHeight(ws, row+1, 18.0)
}

func (g *Generator) evalMainHeader(ws string, row int, title, subtitle string) {
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, EVMainHeaderStyle)
	_ = g.file.SetRowHeight(ws, row, 26.0)
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row+1), cellName(EV_COL_ABW_EUR, row+1), subtitle, EVMainSubStyle)
}

func (g *Generator) evalSectionTitle(ws string, row int, title string) {
	_ = g.mergeCells(ws, cellName(EV_COL_LABEL, row), cellName(EV_COL_ABW_EUR, row), title, EVSectionTitleStyle)
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
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", Registry.InputBudgetKMWEUR.NamedRange), false)
	r++
	rRes := r
	g.evalKmwLabel(ws, r, lblL1, lblL2, "Davon Reserve", false)
	g.evalKmwCalc(ws, cellName(valL, r), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", bgKostenName("Reserve", "EUR")), false)
	r++
	rOp := r
	g.evalMergedFormula(ws, cellName(lblL1, r), cellName(lblL2, r),
		fmt.Sprintf(`=IF(%s="Ja","Operatives Budget (Reserve freigegeben)","Operatives Budget (abzgl. Reserve)")`, BudgetNameReserve),
		EVKmwLabelStyle)
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

	// Registry-Bindung linker Block (Basis-KMW-Mittel).
	if isMA {
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWBewilligt.NamedRange, valL, rBew)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWReserve.NamedRange, valL, rRes)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWOperativ.NamedRange, valL, rOp)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWBereitgestellt.NamedRange, valL, rBer)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWVerfuegbar.NamedRange, valL, rVerf)
	} else {
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWBewilligt.NamedRange, valL, rBew)
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWReserve.NamedRange, valL, rRes)
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWOperativ.NamedRange, valL, rOp)
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWBereitgestellt.NamedRange, valL, rBer)
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWVerfuegbar.NamedRange, valL, rVerf)
	}

	// --- RECHTER BLOCK (Abzugsoptionen, mit Abzug/Kein-Abzug-Schaltern) ---
	rr := startRow
	_ = g.mergeCells(ws, cellName(tog, rr), cellName(valR, rr), "Abzugsoptionen KMW-Mittel", EVAbzugHeaderStyle)
	rr++

	type dedRow struct{ togCell, valCell string }
	var deds []dedRow

	// Saldovortrag (berechnet)
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), Registry.InputMAPruefungAbzugSaldo)
	} else {
		g.evalToggle(ws, cellName(tog, rr), Registry.InputFBPruefungAbzugSaldo)
	}
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Saldovortrag", false)
	g.evalDeduct(ws, cellName(valR, rr), fmt.Sprintf("=IFERROR(ROUND(%s,2),0)", Registry.OutputDashSaldovortragEUR.NamedRange))
	saldoCell := cellName(valR, rr)
	if isMA {
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungSaldovortrag.NamedRange, valR, rr)
	} else {
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungSaldovortrag.NamedRange, valR, rr)
	}
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	// Mehreinnahmen (Formel wird nachgelagert gesetzt)
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), Registry.InputMAPruefungAbzugMehr)
	} else {
		g.evalToggle(ws, cellName(tog, rr), Registry.InputFBPruefungAbzugMehr)
	}
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Mehreinnahmen", false)
	g.evalDeductPlaceholder(ws, cellName(valR, rr))
	mehrCell := cellName(valR, rr)
	if isMA {
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungMehreinnahmen.NamedRange, valR, rr)
	} else {
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungMehreinnahmen.NamedRange, valR, rr)
	}
	deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
	rr++

	prognCell := ""
	if isMA {
		g.evalToggle(ws, cellName(tog, rr), Registry.InputMAPruefungAbzugPrognose)
		g.evalKmwLabel(ws, rr, lblR1, lblR2, "Prognostizierte Mehreinnahmen", false)
		g.evalDeductPlaceholder(ws, cellName(valR, rr))
		prognCell = cellName(valR, rr)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungPrognostizierteMehreinnahmen.NamedRange, valR, rr)
		deds = append(deds, dedRow{absName(tog, rr), absName(valR, rr)})
		rr++
	}

	// Bedingte Formatierung: Wertfelder ausgegraut wenn "Kein Abzug" gewählt.
	// Zwei Regeln pro Zelle, da StyleOptions nur eine BorderColor kennt:
	// Regel 1 – Füllung/Schrift + drei dünne Gitterlinien (links/oben/unten)
	// Regel 2 – dicker rechter Außenrahmen (passend zu styleOuterBorder)
	for _, d := range deds {
		cond := fmt.Sprintf(`%s="Kein Abzug"`, d.togCell)
		g.addConditionalFormat(ws, d.valCell, cond, EVDeductOffCFStyle)
		g.addConditionalFormat(ws, d.valCell, cond, EVRightBorderCFStyle)
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
	if isMA {
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungAbzugGesamt.NamedRange, valR, rIns)
	} else {
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungAbzugGesamt.NamedRange, valR, rIns)
	}
	rr++

	// Verfügbare KMW-Mittel (bereinigt)
	g.evalKmwLabel(ws, rr, lblR1, lblR2, "Verfügbare KMW-Mittel (bereinigt)", true)
	g.evalKmwCalc(ws, cellName(valR, rr), fmt.Sprintf("=ROUND(%s-%s,2)", addrVerf, addrInsgesamt), true)
	addrBereinigt := absName(valR, rr)
	if isMA {
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungKMWVerfuegbarBereinigt.NamedRange, valR, rr)
	} else {
		g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungKMWVerfuegbarBereinigt.NamedRange, valR, rr)
	}
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
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungVerbleibendKMW.NamedRange, valL, p1Bottom)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungVerbleibendKMWBereinigt.NamedRange, valR, p1Bottom)
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
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungVerbleibendKMWManuell.NamedRange, valL, p2Bottom)
		g.dbUpsertNamedRange(ws, Registry.OutputMAPruefungVerbleibendKMWManuellBereinigt.NamedRange, valR, p2Bottom)
		g.styleOuterBorder(ws, p2Top, lblL1, p2Bottom, valL, 2, EV_CLR_BORDER)
		g.styleOuterBorder(ws, p2Top, tog, p2Bottom, valR, 2, EV_CLR_BORDER)
		r++

		condReqGrey := fmt.Sprintf(`%s<>0`, addrManL)
		condManGrey := fmt.Sprintf(`%s=0`, addrManL)

		applyGreyCF := func(rStart, rEnd int, cond string) {
			for row := rStart; row <= rEnd; row++ {
				g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(lblL1, row), cellName(lblL2, row)), cond, EVGreyCFStyle)
				g.addConditionalFormat(ws, cellName(valL, row), cond, EVGreyCFStyle)
				g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(tog, row), cellName(lblR2, row)), cond, EVGreyCFStyle)
				g.addConditionalFormat(ws, cellName(valR, row), cond, EVGreyCFStyle)
			}
		}

		applyGreyCF(p1Top, p1Bottom, condReqGrey)
		applyGreyCF(p2Top, p2Bottom, condManGrey)

		// Negativ-Formatierung (rot), nur wenn nicht ausgegraut
		redStyle := EVNegativeStyle
		g.addConditionalFormat(ws, cellName(valL, p1Bottom), fmt.Sprintf(`AND(%s<0, %s=0)`, addrVerblL1, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valR, p1Bottom), fmt.Sprintf(`AND(%s<0, %s=0)`, addrVerblR1, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valL, p2Bottom), fmt.Sprintf(`AND(%s<0, %s<>0)`, addrVerblL2, addrManL), redStyle)
		g.addConditionalFormat(ws, cellName(valR, p2Bottom), fmt.Sprintf(`AND(%s<0, %s<>0)`, addrVerblR2, addrManL), redStyle)
	}

	return evalKMWResult{nextRow: r, mehrCell: mehrCell, prognCell: prognCell, saldoCell: saldoCell}
}

func (g *Generator) evalKmwLabel(ws string, row, c1, c2 int, text string, bold bool) {
	style := EVKmwLabelStyle
	if bold {
		style = EVKmwLabelBoldStyle
	}
	_ = g.mergeCells(ws, cellName(c1, row), cellName(c2, row), text, style)
}

func (g *Generator) evalKmwCalc(ws, cell, formula string, bold bool) {
	style := EVKmwCalcStyle
	if bold {
		style = EVKmwCalcBoldStyle
	}
	_ = g.setFormula(ws, cell, formula, style)
}

func (g *Generator) evalKmwInput(ws, cell string) {
	_ = g.setStyle(ws, cell, cell, EVKmwInputStyle)
}

// evalKmwInputEmpty entspricht evalKmwInput (gelbe Eingabezelle, EUR-Format),
// lässt die Zelle aber leer statt mit einer 0 vorzubelegen.
func (g *Generator) evalKmwInputEmpty(ws, cell string) {
	_ = g.setStyle(ws, cell, cell, EVKmwInputStyle)
}

func (g *Generator) evalToggle(ws, cell string, field InputField) {
	_ = g.setStyle(ws, cell, cell, EVToggleStyle)
	col, row, _ := excelize.CellNameToCoordinates(cell)
	_ = g.bindInputField(ws, row, col, field)
}

func (g *Generator) evalDeduct(ws, cell, formula string) {
	_ = g.setFormula(ws, cell, formula, EVDeductStyle)
}

func (g *Generator) evalDeductPlaceholder(ws, cell string) {
	_ = g.setStyle(ws, cell, cell, EVDeductStyle)
}

// ==================================================================================
// AUSWAHL-PANELS (zentriert, oben in der jeweiligen Sektion)
// ==================================================================================

// evalSelTitle zeichnet die zentrierte Titelzeile einer Auswahlbox (Spalten C..I).
func (g *Generator) evalSelTitle(ws string, row int, text string) {
	_ = g.mergeCells(ws, cellName(EV_PB_C1, row), cellName(EV_PB_C2, row), text, EVSelTitleStyle)
	_ = g.file.SetRowHeight(ws, row, 22.0)
}

// evalSelLabel zeichnet die Beschriftungsspalte (C..E) einer Auswahlzeile.
func (g *Generator) evalSelLabel(ws string, row int, text string) {
	_ = g.mergeCells(ws, cellName(EV_PB_C1, row), cellName(EV_PB_L2, row), text, EVSelLabelStyle)
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
	num0 := EVPanelNumStyle
	inputCtr := EVPanelInputStyle
	r := top
	g.evalSelTitle(ws, r, "Finanzbericht auswählen")
	r++

	g.evalSelLabel(ws, r, "Auswahl:")
	labelCell := cellName(EV_PB_V1, r)
	g.mergeCells(ws, labelCell, cellName(EV_PB_C2, r), "", inputCtr)
	col, row, _ := excelize.CellNameToCoordinates(labelCell)
	_ = g.bindInputField(ws, row, col, Registry.InputFBPruefungAuswahl)

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
	g.dbUpsertNamedRange(ws, Registry.OutputFBPruefungAusgewaehltePeriode.NamedRange, EV_PB_V1, r)
	r++

	bottom := r - 1

	g.styleOuterBorder(ws, top, EV_PB_C1, bottom, EV_PB_C2, 2, EV_CLR_BORDER)
	return numCell, bottom
}

// ==================================================================================
// MONATSLIMIT-PRÜFUNG
// ==================================================================================

// evalCompFields hält die acht Registry-Ausgabefelder einer Vergleichstabellen-Zeile
// (Kategorie × Metrik × Währung) in Spaltenreihenfolge (Act/Bud/Dif/Abw je LC/EUR).
type evalCompFields struct {
	actLC, budLC, difLC, abwLC     OutputField
	actEUR, budEUR, difEUR, abwEUR OutputField
}

// evalCompBinding bündelt die Ausgabefelder je Datenzeile (in Label-Reihenfolge)
// sowie der GESAMT-Zeile einer Vergleichstabelle.
type evalCompBinding struct {
	rows  []evalCompFields
	total evalCompFields
}

// evalCompBindingFor baut die Registry-Bindung einer Vergleichstabelle aus dem
// Feld-Präfix (z. B. "OutputFBPruefungFin") und den Kategorie-Kürzeln (in
// Label-Reihenfolge). Die Namen stammen ausschließlich aus der Registry; das
// Kürzelschema entspricht den Feldnamen in registry.go (…<Cat><Metrik><Währung>).
func evalCompBindingFor(reg *TemplateRegistry, prefix string, cats []string) evalCompBinding {
	rv := reflect.ValueOf(*reg)
	out := func(cat, metric, cur string) OutputField {
		fv := rv.FieldByName(prefix + cat + metric + cur)
		if !fv.IsValid() {
			panic(fmt.Sprintf("[Developer Error] Registry-Feld fehlt: %s%s%s%s", prefix, cat, metric, cur))
		}
		return fv.Interface().(OutputField)
	}
	mk := func(cat string) evalCompFields {
		return evalCompFields{
			actLC: out(cat, "Act", "LC"), budLC: out(cat, "Bud", "LC"), difLC: out(cat, "Dif", "LC"), abwLC: out(cat, "Abw", "LC"),
			actEUR: out(cat, "Act", "EUR"), budEUR: out(cat, "Bud", "EUR"), difEUR: out(cat, "Dif", "EUR"), abwEUR: out(cat, "Abw", "EUR"),
		}
	}
	rows := make([]evalCompFields, len(cats))
	for i, c := range cats {
		rows[i] = mk(c)
	}
	return evalCompBinding{rows: rows, total: mk("Gesamt")}
}

// evalBindCompRow benennt die acht Wertzellen einer Vergleichstabellen-Zeile.
func (g *Generator) evalBindCompRow(ws string, row int, cf evalCompFields) {
	g.dbUpsertNamedRange(ws, cf.actLC.NamedRange, EV_COL_ACT_LC, row)
	g.dbUpsertNamedRange(ws, cf.budLC.NamedRange, EV_COL_BUD_LC, row)
	g.dbUpsertNamedRange(ws, cf.difLC.NamedRange, EV_COL_DIF_LC, row)
	g.dbUpsertNamedRange(ws, cf.abwLC.NamedRange, EV_COL_ABW_LC, row)
	g.dbUpsertNamedRange(ws, cf.actEUR.NamedRange, EV_COL_ACT_EUR, row)
	g.dbUpsertNamedRange(ws, cf.budEUR.NamedRange, EV_COL_BUD_EUR, row)
	g.dbUpsertNamedRange(ws, cf.difEUR.NamedRange, EV_COL_DIF_EUR, row)
	g.dbUpsertNamedRange(ws, cf.abwEUR.NamedRange, EV_COL_ABW_EUR, row)
}

func (g *Generator) evalDrawComparisonTable(ws string, r int, title string, isIncome, isMA bool, sel evalSelRefs, bind evalCompBinding) evalCompResult {
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
		_ = g.setValue(ws, cellName(EV_COL_LABEL+i, hdrRow), h, EVCompHeaderStyle)
	}
	_ = g.file.SetRowHeight(ws, hdrRow, 28.0)
	r++

	// Zeilen sind kategoriebasiert:
	//   Finanzierungsanteile  → 4 Einnahme-Kategorien (TYPE_NAMES)
	//   Soll-Ist-Abweichung   → 8 Kostenkategorien (EXPENSE_CATEGORIES)
	// Nur die MA-Ausgabenprognose bleibt positionsbasiert (eigene Grid-Quelle).
	labels := g.evalComparisonLabels(isIncome, isMA)
	nRows := len(labels)
	dataStart := r
	kmwActEUR, kmwBudEUR := "", ""

	for i := 0; i < nRows; i++ {
		row := dataStart + i
		name := labels[i]

		_ = g.setValue(ws, cellName(EV_COL_LABEL, row), name, EVCompLabelStyle)

		actLC := cellName(EV_COL_ACT_LC, row)
		actEUR := cellName(EV_COL_ACT_EUR, row)
		lcF, eurF := g.evalActualFormulas(isIncome, isMA, name, i, sel)
		g.evalCellFormula(ws, actLC, lcF, EV_FMT_LC, EV_CLR_CALC)
		g.evalCellFormula(ws, actEUR, eurF, EV_FMT_EUR, EV_CLR_CALC)

		budLCName, budEURName := g.evalComparisonBudgetNames(isIncome, isMA, name, i)
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

		// Alle Wertzellen der Zeile an die Registry-Ausgabefelder binden.
		if i < len(bind.rows) {
			g.evalBindCompRow(ws, row, bind.rows[i])
		}

		if isIncome && name == "KMW-Mittel" {
			kmwActEUR = absName(EV_COL_ACT_EUR, row)
			kmwBudEUR = absName(EV_COL_BUD_EUR, row)
		}
	}

	dataEnd := dataStart + nRows - 1
	totalRow := dataEnd + 1
	g.evalTotalRow(ws, totalRow, dataStart, dataEnd)
	g.evalBindCompRow(ws, totalRow, bind.total)

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

// evalComparisonLabels liefert die Zeilenbeschriftungen der Vergleichstabelle.
// Einnahmen (Finanzierungsanteile) sind die 4 festen Einnahmentypen (TYPE_NAMES),
// Ausgaben (FB-Soll-Ist wie MA-Prognose) die 8 festen Kostenkategorien
// (EXPENSE_CATEGORIES – Bauausgaben, Investitionen …). Beide Seiten sind damit
// fest gesetzt und kategoriebasiert – nicht positions-/budgetabhängig.
func (g *Generator) evalComparisonLabels(isIncome, isMA bool) []string {
	if isIncome {
		return TYPE_NAMES
	}
	return EXPENSE_CATEGORIES
}

// evalActualFormulas: Ist-/Prognose-Formeln je Zeile.
func (g *Generator) evalActualFormulas(isIncome, isMA bool, name string, idx int, sel evalSelRefs) (string, string) {
	if isIncome && isMA {
		maL := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_LC, isIncome)
		maE := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_EUR, isIncome)
		fbL := evalFBChooseRef(sel.fbSelNum, FB_INCOME_FIRST_ROW+idx, FBOffKumLC)
		fbE := evalFBChooseRef(sel.fbSelNum, FB_INCOME_FIRST_ROW+idx, FBOffKumEUR)
		return fmt.Sprintf("=%s + %s", maL[1:], fbL[1:]),
			fmt.Sprintf("=%s + %s", maE[1:], fbE[1:])
	}
	if isIncome {
		return evalFBChooseRef(sel.fbSelNum, FB_INCOME_FIRST_ROW+idx, FBOffKumLC),
			evalFBChooseRef(sel.fbSelNum, FB_INCOME_FIRST_ROW+idx, FBOffKumEUR)
	}
	if isMA {
		// MA-Ausgabenprognose bleibt positionsbasiert (1:1 auf FB-Ausgabenzeile).
		maL := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_LC, isIncome)
		maE := evalMAExpenseActual(g, sel, idx, EV_DTN_MAG_EUR, isIncome)
		row := FB_AUSG_FIRST_ROW + idx
		fbL := evalFBChooseRef(sel.fbSelNum, row, FBOffKumLC)
		fbE := evalFBChooseRef(sel.fbSelNum, row, FBOffKumEUR)
		return fmt.Sprintf("=%s + %s", maL[1:], fbL[1:]),
			fmt.Sprintf("=%s + %s", maE[1:], fbE[1:])
	}
	// FB-Ausgaben (Soll-Ist): kumulierte Ist-Werte je Kostenkategorie.
	return g.evalFBExpenseActualByCategory(sel.fbSelNum, name, FBOffKumLC),
		g.evalFBExpenseActualByCategory(sel.fbSelNum, name, FBOffKumEUR)
}

// evalFBExpenseActualByCategory summiert die kumulierten FB-Ausgaben einer
// Kostenkategorie per SUMPRODUCT: die Kategoriespalte des Budgets (gleiche
// Positionsreihenfolge wie die FB-Ausgabenzeilen) mal die per CHOOSE gewählte
// Periodenspalte des Finanzberichts. colOffset = FBOffKumLC/FBOffKumEUR.
func (g *Generator) evalFBExpenseActualByCategory(selNum, cat string, colOffset int) string {
	nPos := g.budgetExpenseCount()
	catRange := fmt.Sprintf("'%s'!%s:%s", BudgetSheetName,
		absName(BudgetColLabel, BudgetRowAusgStart), absName(BudgetColLabel, BudgetRowAusgStart+nPos-1))
	parts := make([]string, 0, FBPeriodenAnzahl)
	for p := 1; p <= FBPeriodenAnzahl; p++ {
		col := FBStartCol + (p-1)*(FBTableCols+FBTableSpacing) + colOffset
		parts = append(parts, fmt.Sprintf("'%s'!%s:%s", EVAL_FB_SHEET,
			absName(col, FB_AUSG_FIRST_ROW), absName(col, FB_AUSG_FIRST_ROW+nPos-1)))
	}
	return fmt.Sprintf(`=IFERROR(ROUND(SUMPRODUCT((%s="%s")*CHOOSE(%s,%s)),2),0)`,
		catRange, cat, selNum, strings.Join(parts, ","))
}

// evalComparisonBudgetNames liefert die Budget-Named-Ranges je Zeile. FB-Ausgaben
// nutzen die kategorienbasierten Budgetsummen (Kosten_<Kat>_…); Einnahmen und die
// positionsbasierte MA-Prognose greifen auf evalBudgetNames zu.
func (g *Generator) evalComparisonBudgetNames(isIncome, isMA bool, name string, idx int) (string, string) {
	if isIncome || isMA {
		return g.evalBudgetNames(isIncome, idx)
	}
	lc := bgKostenName(name, "LW")
	eur := bgKostenName(name, "EUR")
	if name == "Reserve" {
		eur = Registry.OutputBudgetReserveEUR.NamedRange
	}
	return lc, eur
}

func (g *Generator) evalBudgetNames(isIncome bool, idx int) (string, string) {
	if isIncome {
		switch idx {
		case 0:
			return Registry.InputBudgetEigenmittelLC.NamedRange, Registry.InputBudgetEigenmittelEUR.NamedRange
		case 1:
			return Registry.OutputBudgetDrittmittelLC.NamedRange, Registry.OutputBudgetDrittmittelEUR.NamedRange
		case 2:
			return Registry.InputBudgetKMWLC.NamedRange, Registry.InputBudgetKMWEUR.NamedRange
		default:
			return "0", "0"
		}
	}
	// Positionsbasiert (MA-Prognose): Budgetwert der Position per INDEX.
	budLC := fmt.Sprintf("INDEX(%s[Betrag (LC)], %d)", BudgetTableAusg, idx+1)
	budEUR := fmt.Sprintf("INDEX(%s[Betrag (EUR)], %d)", BudgetTableAusg, idx+1)
	return budLC, budEUR
}

func (g *Generator) evalTotalRow(ws string, totalRow, dataStart, dataEnd int) {
	_ = g.setValue(ws, cellName(EV_COL_LABEL, totalRow), "GESAMT", EVTotalLabelStyle)
	sumCol := func(col int, style StyleOptions) {
		rng := fmt.Sprintf("%s:%s", absName(col, dataStart), absName(col, dataEnd))
		_ = g.setFormula(ws, cellName(col, totalRow), fmt.Sprintf("=ROUND(SUM(%s),2)", rng), style)
	}
	pctCol := func(col, actCol, budCol int) {
		_ = g.setFormula(ws, cellName(col, totalRow),
			fmt.Sprintf("=ROUND(IFERROR((%s/%s)-1,0),4)", absName(actCol, totalRow), absName(budCol, totalRow)),
			EVTotalPctStyle)
	}
	sumCol(EV_COL_ACT_LC, EVTotalLCStyle)
	sumCol(EV_COL_BUD_LC, EVTotalLCStyle)
	sumCol(EV_COL_DIF_LC, EVTotalLCStyle)
	pctCol(EV_COL_ABW_LC, EV_COL_ACT_LC, EV_COL_BUD_LC)
	sumCol(EV_COL_ACT_EUR, EVTotalEURStyle)
	sumCol(EV_COL_BUD_EUR, EVTotalEURStyle)
	sumCol(EV_COL_DIF_EUR, EVTotalEURStyle)
	pctCol(EV_COL_ABW_EUR, EV_COL_ACT_EUR, EV_COL_BUD_EUR)
	_ = g.file.SetRowHeight(ws, totalRow, 20.0)
}

func (g *Generator) evalDeviationConditional(ws string, col, dataStart, dataEnd int) {
	rng := fmt.Sprintf("%s:%s", cellName(col, dataStart), cellName(col, dataEnd))
	topRel := cellName(col, dataStart)
	// Bewusst OHNE eigene Rahmen: die bedingte Formatierung überschreibt sonst die
	// vorhandene Zellkante (innen dünnes Gitter, an der Box rechts kräftig durch
	// styleOuterBorder). Nur Schrift/Füllung/Format setzen – wie evalLimitUeber.
	g.addConditionalFormat(ws, rng, fmt.Sprintf("%s>=0.2", topRel), EVDevBadStyle)
	g.addConditionalFormat(ws, rng, fmt.Sprintf("AND(%s>=0.1,%s<0.2)", topRel, topRel), EVDevWarnStyle)
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
// (Periode N) per CHOOSE über alle Perioden (nicht-volatil, robust). colOffset ist
// ein FB-Spalten-Offset (z. B. FBOffKumLC/FBOffKumEUR).
func evalFBChooseRef(selNum string, baseRow, colOffset int) string {
	parts := make([]string, 0, FBPeriodenAnzahl)
	for p := 1; p <= FBPeriodenAnzahl; p++ {
		col := FBStartCol + (p-1)*(FBTableCols+FBTableSpacing) + colOffset
		parts = append(parts, fmt.Sprintf("'%s'!%s", EVAL_FB_SHEET, absName(col, baseRow)))
	}
	return fmt.Sprintf(`=IFERROR(ROUND(CHOOSE(%s,%s),2),0)`, selNum, strings.Join(parts, ","))
}
