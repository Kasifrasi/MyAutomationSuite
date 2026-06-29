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
	// Mehreinnahmen der MA-Prüfung = identischer Wert wie in der FB-Prüfung.
	g.evalDeduct(ws, maKMW.mehrCell, fmt.Sprintf("=%s", g.evalFBMehrCell))

	// Prognostizierte (zusätzliche) Mehreinnahmen:
	if maKMW.prognCell != "" {
		realAct := fmt.Sprintf("(SUM(%s)-%s)", g.evalFBResIncActEUR, g.evalFBResIncKmwActEUR)
		realBud := fmt.Sprintf("(SUM(%s)-%s)", g.evalFBResIncBudEUR, g.evalFBResIncKmwBudEUR)
		maEig := fmt.Sprintf(
			`SUMIFS('%s'!%s,'%s'!%s,%s,'%s'!%s,"<="&%s,'%s'!%s,">=1")`,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_EIGDR, 1, MA_TABLE_COUNT),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), maSelPCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), maSelKCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT))
		prognFormula := fmt.Sprintf(
			`=ROUND(MAX(0,MAX(0,%s+%s-%s)-%s),2)`, realAct, maEig, realBud, g.evalFBMehrCell)
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
