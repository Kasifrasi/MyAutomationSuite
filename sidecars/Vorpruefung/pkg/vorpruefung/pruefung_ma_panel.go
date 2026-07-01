package vorpruefung

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
	g.mergeCells(ws, labelCell, cellName(EV_PB_C2, r), "", inputCtr)
	col, row, _ := excelize.CellNameToCoordinates(labelCell)
	_ = g.bindInputField(ws, row, col, Registry.InputMAPruefungAuswahl)

	dv := excelize.NewDataValidation(true)
	dv.Sqref = labelCell
	dv.Type = "list"
	dv.Formula1 = "=" + EVAL_NAME_MA_LISTE
	_ = g.file.AddDataValidation(ws, dv)
	r++

	g.evalSelLabel(ws, r, "Ausgewählte Periode")
	pCell := cellName(EV_PB_V1, r)
	maxFbPer := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(('%s'!%s=1)*'%s'!%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_FB_META_FILL, 1, MA_PERIOD_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_FB_META_PER, 1, MA_PERIOD_COUNT))
	maxMaPer := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(('%s'!%s=1)*'%s'!%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_FILL, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT))
	maxMAP := fmt.Sprintf(`%s+1`, maxFbPer)
	pFormula := fmt.Sprintf(
		`=IF(%s="Neueste MA",%s,IFERROR(VALUE(MID(%s,FIND("Periode ",%s)+8,FIND(" ",%s,FIND("Periode ",%s)+8)-(FIND("Periode ",%s)+8))),0))`,
		labelCell, maxMAP, labelCell, labelCell, labelCell, labelCell, labelCell)
	g.evalMergedFormula(ws, pCell, cellName(EV_PB_C2, r), pFormula, num0)
	r++

	g.evalSelLabel(ws, r, "Ausgewählte Anforderung (#)")
	kCell := cellName(EV_PB_V1, r)
	// Höchster Rang der befüllten MAs in der Folgeperiode (=pCell). MAXIFS wäre
	// als Future-Function (_xlfn., implizite Schnittmenge "@") hier unzuverlässig;
	// das SUMPRODUCT(MAX(...))-Idiom nutzt nur Legacy-Funktionen und erzwingt die
	// Array-Auswertung – identisch zum Vorgehen der Spiegel-Panels.
	maxMAK := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(('%s'!%s=%s)*('%s'!%s=1)*'%s'!%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), pCell,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_FILL, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT))
	kFormula := fmt.Sprintf(
		`=IF(%s="Neueste MA",MAX(1,%s),IFERROR(VALUE(MID(%s,FIND("(#",%s)+2,FIND(")",%s)-FIND("(#",%s)-2)),0))`,
		labelCell, maxMAK, labelCell, labelCell, labelCell, labelCell)
	g.evalMergedFormula(ws, kCell, cellName(EV_PB_C2, r), kFormula, num0)
	r++

	warnCol1 := EV_PB_C2 + 1 // Spalte H (8)
	warnCol2 := EV_PB_C2 + 3 // Spalte J (10)

	// Zeile 9 (top + 1, falls top = 8 ist, was meist der Fall ist; wir nehmen konkret "top + 1" an,
	// um flexibel zu bleiben, bzw. hart 9, wenn wir die Sektion kennen. Da "r" bei top=8 startet, ist top+1 = 9)
	warnTitleRow := top + 1
	warnTextRow1 := top + 2
	warnTextRow2 := top + 6

	// Überschrift (H9:J9), nur anzeigen, wenn Warnbedingung zutrifft
	warnTitleFormula := fmt.Sprintf(`=IF(%s > %s + 1, "Hinweis:", "")`, maxMaPer, maxFbPer)
	g.evalMergedFormula(ws, cellName(warnCol1, warnTitleRow), cellName(warnCol2, warnTitleRow), warnTitleFormula, StyleOptions{
		FontColor: "FF0000", Bold: true, HAlign: "center", VAlign: "center",
	})

	// Textkörper (H10:J14), schwarz, zentriert, etwas kleiner (Size 10 ist Standard, wir nehmen 9)
	warnFormula := fmt.Sprintf(`=IF(%s > %s + 1, "Eine oder mehrere ausgefüllte Mittelanforderungen werden nicht zur Auswahl angezeigt, da diese 2 oder mehr Perioden nach dem letzten Finanzbericht liegen. Es bedarf zunächst einer Finanzberichtsprüfung.", "")`, maxMaPer, maxFbPer)
	g.evalMergedFormula(ws, cellName(warnCol1, warnTextRow1), cellName(warnCol2, warnTextRow2), warnFormula, StyleOptions{
		FontColor: EV_CLR_BLACK, Size: 9.0, WrapText: true, HAlign: "center", VAlign: "top",
	})

	_ = g.mergeCells(ws, cellName(EV_PB_C1, r), cellName(EV_PB_L2, r), "Einbezogene Anforderungen", StyleOptions{
		Bold: true, Size: 9.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	_ = g.setValue(ws, cellName(EV_PB_V1, r), "Angefragt (LC)", StyleOptions{
		Bold: true, Size: 9.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	_ = g.setValue(ws, cellName(EV_PB_SLC2, r), "Angefragt (EUR)", StyleOptions{
		Bold: true, Size: 9.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	_ = g.setValue(ws, cellName(EV_PB_SEU1, r), "Manuell (EUR)", StyleOptions{
		Bold: true, Size: 9.0, FontColor: EV_CLR_BLACK, FillColor: EV_CLR_HEADER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
	})
	r++

	// Slots: bei Bedarf (s <= k) eingeblendet, sonst leer (Struktur im Hintergrund).
	firstSlot := r
	lblSt := StyleOptions{Size: 9.0, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	valSt := StyleOptions{HAlign: "right", VAlign: "center", NumFormat: EV_FMT_LC,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	euSt := StyleOptions{HAlign: "right", VAlign: "center", NumFormat: EV_FMT_EUR,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID}
	for s := 1; s <= EV_MA_SLOTS; s++ {
		row := firstSlot + s - 1
		lbl := fmt.Sprintf(`=IF(%d<=%s,"Periode "&%s&" (#%d)","")`, s, kCell, pCell, s)
		g.evalMergedFormula(ws, cellName(EV_PB_C1, row), cellName(EV_PB_L2, row), lbl, lblSt)

		lcF := fmt.Sprintf(`=IF(%d<=%s,IFERROR(SUMIFS('%s'!%s,'%s'!%s,"KMW-Mittel",'%s'!%s,%s,'%s'!%s,%d),0),"")`,
			s, kCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_LC, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, g.maGridRows()), pCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()), s)
		_ = g.setFormula(ws, cellName(EV_PB_V1, row), lcF, valSt)

		euF := fmt.Sprintf(`=IF(%d<=%s,IFERROR(SUMIFS('%s'!%s,'%s'!%s,"KMW-Mittel",'%s'!%s,%s,'%s'!%s,%d),0),"")`,
			s, kCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_EUR, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, g.maGridRows()), pCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()), s)
		_ = g.setFormula(ws, cellName(EV_PB_SLC2, row), euF, euSt)

		manF := fmt.Sprintf(`=IF(%d<=%s,IFERROR(SUMIFS('%s'!%s,'%s'!%s,"Manueller Betrag",'%s'!%s,%s,'%s'!%s,%d),0),"")`,
			s, kCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_EUR, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_CAT, 1, g.maGridRows()),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_PER, 1, g.maGridRows()), pCell,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MAG_RANK, 1, g.maGridRows()), s)
		_ = g.setFormula(ws, cellName(EV_PB_SEU1, row), manF, euSt)

		labelAddr := absName(EV_PB_C1, row)
		cond := fmt.Sprintf(`%s<>""`, labelAddr)

		bBot := 1
		if s == EV_MA_SLOTS {
			bBot = 2
		}

		g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(EV_PB_C1, row), cellName(EV_PB_L2, row)),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: bBot, BorderLeft: 2, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
		g.addConditionalFormat(ws, cellName(EV_PB_V1, row),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: bBot, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
		g.addConditionalFormat(ws, cellName(EV_PB_SLC2, row),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: bBot, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
			})
		g.addConditionalFormat(ws, cellName(EV_PB_SEU1, row),
			cond, StyleOptions{
				FillColor: EV_CLR_PANEL_REV, BorderTop: 1, BorderBottom: bBot, BorderLeft: 1, BorderRight: 2, BorderColor: EV_CLR_BORDER,
			})
	}
	bottom := firstSlot + EV_MA_SLOTS - 1
	g.styleOuterBorder(ws, top, EV_PB_C1, bottom, EV_PB_C2, 2, EV_CLR_BORDER)
	return pCell, kCell, bottom
}

// evalDrawFBPanel zeichnet die zentrierte Finanzbericht-Auswahlbox und liefert die
// Periodennummer-Zelle (N) sowie die letzte belegte Zeile.
func (g *Generator) evalDrawMonatslimit(ws string, r int, sel evalSelRefs) int {
	g.evalSectionTitle(ws, r, "Monatslimit-Prüfung")
	r += 2

	// Spalten: Lokalwährung eine nach links gerückt (C), gemeinsame Eingabefelder
	// (Monatsanteile) mittig (D), Euro am bewährten Platz (E).
	const lbl, vLC, vMID, vEUR = EV_COL_LABEL, EV_COL_LABEL + 1, EV_COL_LABEL + 2, EV_COL_LABEL + 3
	top := r

	// Jahresbudget-EUR wird mit dem Budget-Kurs (Gesamtprojekt) umgerechnet.
	rate := Registry.OutputBudgetWK.NamedRange

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
	monthsInput := func(row, val int, field InputField) string {
		cell := cellName(vMID, row)
		_ = g.setStyle(ws, cell, cell, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: "0", FillColor: EV_CLR_INPUT,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_GRID,
		})
		c, r, _ := excelize.CellNameToCoordinates(cell)
		_ = g.bindInputField(ws, r, c, field)
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
			EVAL_DATEN_SHEET, evalAbsCol(sumCol, 1, MA_TABLE_COUNT),
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), sel.maSelP,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), sel.maSelK,
			EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT))
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
	yearMonthAddrs := make([]string, len(BudgetYears))
	yearBudLCAddrs := make([]string, len(BudgetYears))
	yearBudEURAddrs := make([]string, len(BudgetYears))
	defaultMonths := []int{8, 0, 0} // bisheriges 8-Monats-Verhalten als Ausgangswert
	fields := []InputField{Registry.InputMAPruefungMonateY1, Registry.InputMAPruefungMonateY2, Registry.InputMAPruefungMonateY3}
	for i, year := range BudgetYears {
		label(r, "Jahresbudget "+year, false)
		budF := fmt.Sprintf("=IFERROR(ROUND(SUBTOTAL(109,%s[%s]),2),0)", BudgetTableAusg, year)
		g.evalLimitCalc(ws, cellName(vLC, r), budF, EV_FMT_LC, false)
		yearBudLCAddrs[i] = absName(vLC, r)
		dm := 0
		if i < len(defaultMonths) {
			dm = defaultMonths[i]
		}
		yearMonthAddrs[i] = monthsInput(r, dm, fields[i])
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
	limTermsLC := make([]string, len(BudgetYears))
	limTermsEUR := make([]string, len(BudgetYears))
	for i := range BudgetYears {
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

// evalDrawMAMirrorPanel zeichnet rechts neben der Mittelanforderungsprüfung eine
// Spiegelung der ausgewählten Mittelanforderung im Format des Blatts "IV. MA".
func (g *Generator) evalDrawMAMirrorPanel(ws string, top int, sel evalSelRefs) {
	const pLbl, pLC, pEUR = 12, 13, 14 // L | M | N
	g.setColWidth(ws, pLbl-1, 3.0)     // Spalte K als Abstand zur Tabelle
	g.setColWidth(ws, pLbl, 36.0)
	g.setColWidth(ws, pLC, 21.0)
	g.setColWidth(ws, pEUR, 21.0)

	// Index j der gewählten MA-Tabelle (Periode = maSelP, Rang = maSelK).
	jAddr := absName(EV_HELP_COL, top)
	_ = g.file.SetCellFormula(ws, cellName(EV_HELP_COL, top), fmt.Sprintf(
		`=IFERROR(SUMPRODUCT('%s'!%s,('%s'!%s=%s)*('%s'!%s=%s)),0)`,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_J, 1, MA_TABLE_COUNT),
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_PER, 1, MA_TABLE_COUNT), sel.maSelP,
		EVAL_DATEN_SHEET, evalAbsCol(EV_DTN_MA_META_RANK, 1, MA_TABLE_COUNT), sel.maSelK))
	_ = g.file.SetColVisible(ws, colLetter(EV_HELP_COL), false)

	// mirror liefert eine CHOOSE-Formel über alle 18 MA-Tabellen, gewählt per j.
	mirror := func(colOffset, srcRow int) string {
		parts := make([]string, 0, MA_TABLE_COUNT)
		for t := 1; t <= MA_TABLE_COUNT; t++ {
			p := ((t - 1) % MA_PERIOD_COUNT) + 1
			level := ((t - 1) / MA_PERIOD_COUNT) + 1
			offsetR := (level - 1) * 30
			colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
			parts = append(parts, fmt.Sprintf("'%s'!%s", MA_SHEET_NAME, absName(colS+colOffset, srcRow+offsetR)))
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, jAddr, strings.Join(parts, ","))
	}

	r := top

	// Titel folgt der Auswahl: "… – Periode X (#k)" bzw. Hinweis ohne Auswahl.
	titleFormula := fmt.Sprintf(
		`=IF(%s=0,"Aktuelle Mittelanforderung (keine gewählt)","Aktuelle Mittelanforderung – Periode "&%s&" (#"&%s&")")`,
		sel.maSelK, sel.maSelP, sel.maSelK)
	_ = g.file.MergeCell(ws, cellName(pLbl, r), cellName(pEUR, r))
	_ = g.setStyle(ws, cellName(pLbl, r), cellName(pEUR, r), StyleOptions{
		Bold: true, Size: 11.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
	})
	_ = g.file.SetCellFormula(ws, cellName(pLbl, r), titleFormula)
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// Kopf-Infozeile: Beschriftung links, Wert über M:N zusammengeführt.
	infoRow := func(label, formula, numFmt string) {
		_ = g.setValue(ws, cellName(pLbl, r), label, StyleOptions{
			Bold: true, HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
		c1, c2 := cellName(pLC, r), cellName(pEUR, r)
		_ = g.file.MergeCell(ws, c1, c2)
		_ = g.setStyle(ws, c1, c2, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: numFmt, FillColor: MAClrGray,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
		_ = g.file.SetCellFormula(ws, c1, formula)
	}
	infoRow("Periode:", mirror(1, 4), "")
	r++
	infoRow("Von:", mirror(1, 5), "DD.MM.YYYY")
	r++
	infoRow("Bis:", mirror(1, 6), "DD.MM.YYYY")
	r++
	infoRow("Zeitraum:", mirror(1, 7), `0" Monate"`)
	r++
	infoRow("OANDA-Kurs:", mirror(1, 8), "0.0000")
	r += 2 // Leerzeile

	// Tabellenkopf
	hdr := StyleOptions{Bold: true, FillColor: MAClrGray, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080"}
	_ = g.setValue(ws, cellName(pLbl, r), "Kostenkategorie", hdr)
	_ = g.setValue(ws, cellName(pLC, r), "Angefordert (LC)", hdr)
	_ = g.setValue(ws, cellName(pEUR, r), "Angefordert (EUR)", hdr)
	r++

	labCell := func(text string, bold bool, fill string) {
		_ = g.setValue(ws, cellName(pLbl, r), text, StyleOptions{
			Bold: bold, HAlign: "left", VAlign: "center", FillColor: fill,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
	}
	valCell := func(col int, formula, numFmt, fill string) {
		_ = g.setFormula(ws, cellName(col, r), formula, StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: fill,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
	}
	bold := func(col int, formula, numFmt, fill string) {
		_ = g.setFormula(ws, cellName(col, r), formula, StyleOptions{
			Bold: true, HAlign: "right", VAlign: "center", NumFormat: numFmt, FillColor: fill,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
	}

	for i, cat := range MA_CATEGORIES {
		src := 10 + i
		labCell(cat, false, "FFFFFF")
		valCell(pLC, mirror(1, src), "#,##0.00", "FFFFFF")
		valCell(pEUR, mirror(2, src), `#,##0.00" €"`, "FFFFFF")
		r++
	}

	labCell("SUMME", true, MAClrGray)
	bold(pLC, mirror(1, 18), "#,##0.00", MAClrGray)
	bold(pEUR, mirror(2, 18), `#,##0.00" €"`, MAClrGray)
	r += 2 // Leerzeile

	labCell("Gesamtbedarf an Mitteln:", false, "FFFFFF")
	valCell(pLC, mirror(1, 20), "#,##0.00", "FFFFFF")
	valCell(pEUR, mirror(2, 20), `#,##0.00" €"`, "FFFFFF")
	r++
	labCell("abzüglich Eigenmittel:", false, "FFFFFF")
	valCell(pLC, mirror(1, 21), "#,##0.00", "FFFFFF")
	valCell(pEUR, mirror(2, 21), `#,##0.00" €"`, "FFFFFF")
	r++
	labCell("abzüglich Drittmittel:", false, "FFFFFF")
	valCell(pLC, mirror(1, 22), "#,##0.00", "FFFFFF")
	valCell(pEUR, mirror(2, 22), `#,##0.00" €"`, "FFFFFF")
	r++
	// Saldo-Beschriftung dynamisch aus der Quelle spiegeln (Vorprojekt/Vorperiode).
	_ = g.setFormula(ws, cellName(pLbl, r), mirror(0, 23), StyleOptions{
		HAlign: "left", VAlign: "center", FillColor: "FFFFFF",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
	})
	valCell(pLC, mirror(1, 23), "#,##0.00", "FFFFFF")
	valCell(pEUR, mirror(2, 23), `#,##0.00" €"`, "FFFFFF")
	r += 2 // Leerzeile

	labCell("KMW-Mittel Anforderung:", true, MAClrKMW)
	bold(pLC, mirror(1, 25), "#,##0.00", MAClrKMW)
	bold(pEUR, mirror(2, 25), `#,##0.00" €"`, MAClrKMW)
	bottom := r

	g.styleOuterBorder(ws, top, pLbl, bottom, pEUR, 2, EV_CLR_BORDER)
}
