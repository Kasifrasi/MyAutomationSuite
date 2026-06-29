package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CreateMAPruefungSheet baut das Blatt "VI. Prüfung MA".
func (g *Generator) CreateMAPruefungSheet() error {
	ws := constants.VPSheetMA_PRUEFUNG
	f := g.file

	if _, err := f.NewSheet(ws); err != nil {
		return fmt.Errorf("fehler beim Erstellen des MA-Prüfungs-Blatts: %w", err)
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
	g.evalBanner(ws, r, ws, "Automatische Prüfung von Mittelanforderungen")
	r += 3

	maSectionTop := r
	g.evalMainHeader(ws, r, "MITTELANFORDERUNGSPRÜFUNG", "Basis: ausgewählte Mittelanforderung (Folgeperiode des Finanzberichts)")
	r += 3

	// Auswahl-Panel (zentriert, oben) zuerst – liefert maSelP-/maSelK-Steuerzellen.
	maSelPCell, maSelKCell, r := g.evalDrawMAPanel(ws, r)
	r += EV_TABLE_GAP

	// Assuming FB_Prüfung runs first, we extract the cell value from g.evalFBSelNumAddr
	fbSelNumCell := strings.ReplaceAll(strings.Split(g.evalFBSelNumAddr, "!")[1], "$", "")
	sel := evalSelRefs{maSelP: maSelPCell, maSelK: maSelKCell, fbSelNum: fbSelNumCell}

	// Jetzt maSelP = N+1 setzen (Folgeperiode des gewählten Finanzberichts).
	// ...wurde entfernt, da maSelP nun direkt aus dem MA-Auswahl-Dropdown gelesen wird.

	maKMW := g.evalDrawKMWSektion(ws, r, true, sel)
	r = maKMW.nextRow + EV_TABLE_GAP

	r = g.evalDrawMonatslimit(ws, r, sel) + EV_TABLE_GAP

	resMAInc := g.evalDrawComparisonTable(ws, r, "Prognostizierte Finanzierungsanteile", true, true, sel)
	r = resMAInc.nextRow + EV_TABLE_GAP

	resMAExp := g.evalDrawComparisonTable(ws, r, "Prognoseprüfung (Ausgaben)", false, true, sel)
	r = resMAExp.nextRow + EV_TABLE_GAP

	// ── Nachgelagerte (sektionsübergreifende) Formeln der MA-Abzugsoptionen ──

	realAct, realBud := g.evalFBMehreinnahmenParts(sel)
	fbMehrFormula := fmt.Sprintf(`ROUND(MAX(0, (%s) - (%s)), 2)`, realAct, realBud)

	// Mehreinnahmen der MA-Prüfung = berechnet aus FB-Ist und FB-Budget
	g.evalDeduct(ws, maKMW.mehrCell, fmt.Sprintf("=IFERROR(%s, 0)", fbMehrFormula))

	// Prognostizierte (zusätzliche) Mehreinnahmen:
	if maKMW.prognCell != "" {
		maEig := fmt.Sprintf(
			`SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1")`,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_EIGDR, 1, MA_TABLE_COUNT),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), maSelPCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), maSelKCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT))
		prognFormula := fmt.Sprintf(
			`=IFERROR(ROUND(MAX(0,MAX(0,(%s)+%s-(%s))-%s),2),0)`, realAct, maEig, realBud, fbMehrFormula)
		g.evalDeduct(ws, maKMW.prognCell, prognFormula)
	}

	for _, cell := range []string{
		maKMW.saldoCell, maKMW.mehrCell, maKMW.prognCell,
	} {
		if cell != "" {
			g.reapplyRightBorder(ws, cell, 2, EV_CLR_BORDER)
		}
	}

	// Schreibgeschützte Spiegel-Panels (rechts neben der jeweiligen Prüfung).
	g.evalDrawMAMirrorPanel(ws, maSectionTop, sel)

	return nil
}

// evalMAExpenseActual summiert die ausgewählten Mittelanforderungen (#1..#k) der
// Periode P je Kategorie über das MA-Grid auf dem Daten-Blatt.
func evalMAExpenseActual(g *Generator, sel evalSelRefs, idx int, valCol int, isIncome bool) string {
	cat := ""
	if isIncome {
		switch idx {
		case 0:
			cat = "Eigenmittel"
		case 1:
			cat = "Drittmittel"
		case 2:
			cat = "KMW-Mittel"
		case 3:
			cat = "Manueller Betrag" // Unused but mapping holds
		}
	} else {
		cat = fmt.Sprintf("Expense_%d", idx)
	}

	// Weil der Rank (sel.maSelK) jetzt fest der Level-Ebene (1, 2, 3) entspricht,
	// liefert eine SUMIFS über Rang <= k genau alle MAs dieser Periode bis zu diesem Level.
	return fmt.Sprintf(
		`=IFERROR(ROUND(SUMIFS('%s'!%s,'%s'!%s,"%s",'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1"),2),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(valCol, 1, g.maGridRows()),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, g.maGridRows()), cat,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, g.maGridRows()), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()), sel.maSelK,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()))
}

// evalMAChooseManBetrag liefert einen CHOOSE-Ausdruck über die MA_ManBetrag_<n>-Namen
// aller 18 Perioden, gewählt per maSelP – analog zu evalMAChooseKurs.
func evalMAChooseManBetrag(sel evalSelRefs) string {
	j := fmt.Sprintf(`IFERROR(SUMPRODUCT('%s'!%s,('%s'!%s=%s)*('%s'!%s=%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_J, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), sel.maSelK)
	parts := make([]string, MA_TABLE_COUNT)
	for i := range parts {
		parts[i] = fmt.Sprintf("MA_ManBetrag_%d", i+1)
	}
	return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),0)`, j, strings.Join(parts, ","))
}

// evalMAChooseKurs liefert einen nicht-volatilen CHOOSE-Ausdruck über alle 18
// MA_Kurs_<n>-Namen. Ersatz für INDIRECT("MA_Kurs_"&maSelP), das Excel 365 mit
// dem @-Operator (implizite Schnittmenge) versieht und dadurch falsch rendert.
func evalMAChooseKurs(sel evalSelRefs) string {
	j := fmt.Sprintf(`IFERROR(SUMPRODUCT('%s'!%s,('%s'!%s=%s)*('%s'!%s=%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_J, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), sel.maSelK)
	parts := make([]string, MA_TABLE_COUNT)
	for i := range parts {
		parts[i] = FieldMAKurs(i + 1).NamedRange
	}
	return fmt.Sprintf(`IFERROR(CHOOSE(%s,%s),0)`, j, strings.Join(parts, ","))
}

// evalMACurrentKMWRequest summiert die KMW-Mittel-Anforderung der EINEN aktuell
// gewählten Mittelanforderung (Periode P, Rang exakt = k) aus dem MA-Grid. Anders
// als die Prognose (#1..#k) bewusst nicht zusammengesetzt – frühere Anforderungen
// einer Periode sind bereits über die bereitgestellten KMW-Mittel erfasst.
func evalMACurrentKMWRequest(g *Generator, sel evalSelRefs) string {
	return fmt.Sprintf(
		`=IFERROR(ROUND(SUMIFS('%s'!%s,'%s'!%s,"KMW-Mittel",'%s'!%s,%s,'%s'!%s,%s),2),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_EUR, 1, g.maGridRows()),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, g.maGridRows()),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, g.maGridRows()), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()), sel.maSelK)
}

// evalMASelectedZeitraum liefert den Zeitraum (Monate) der aktuell gewählten
// Mittelanforderung (Periode P, Rang k). Es wird – wie im Spiegel-Panel – der
// Tabellenindex j der gewählten MA bestimmt und der Zeitraum (MA-Quellzeile 7,
// Wertspalte = colS+1) per CHOOSE über alle 18 MA-Tabellen ausgelesen.
func evalMASelectedZeitraum(sel evalSelRefs) string {
	j := fmt.Sprintf(`IFERROR(SUMPRODUCT('%s'!%s,('%s'!%s=%s)*('%s'!%s=%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_J, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), sel.maSelK)
	parts := make([]string, 0, MA_TABLE_COUNT)
	for t := 1; t <= MA_TABLE_COUNT; t++ {
		p := ((t - 1) % MA_PERIOD_COUNT) + 1
		level := ((t - 1) / MA_PERIOD_COUNT) + 1
		offsetR := (level - 1) * 30
		colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
		parts = append(parts, fmt.Sprintf("'%s'!%s", MA_SHEET_NAME, absName(colS+1, 7+offsetR)))
	}
	return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),0)`, j, strings.Join(parts, ","))
}

// evalFBMehreinnahmenParts liefert die Summe der Ist- und Budget-Werte für Einnahmen
// (in EUR) aus dem aktuell gewählten Finanzbericht, exklusive KMW-Mittel.
func (g *Generator) evalFBMehreinnahmenParts(sel evalSelRefs) (string, string) {
	var fbIstParts []string
	var budParts []string
	for i := 0; i < g.budgetIncomeCount(); i++ {
		// Wait, KMW is typically index 2. We skip KMW-Mittel
		if i == 2 {
			continue
		}
		fbIstParts = append(fbIstParts, evalFBChooseRef(sel.fbSelNum, 12+i, 4)[1:])
		_, budName := g.evalBudgetNames(true, i)
		budParts = append(budParts, fmt.Sprintf("IFERROR(%s,0)", budName))
	}
	return strings.Join(fbIstParts, "+"), strings.Join(budParts, "+")
}
