package vorpruefung

import (
	"fmt"
	"strings"
)

// EV_HELP_COL ist eine weit rechts liegende, ausgeblendete Helferspalte des
// Auswertungsblatts (hält z. B. den Index der aktuell gewählten MA-Tabelle).
const EV_HELP_COL = 30

const EV_GRID_LIGHT = "D3D3D3" // helle Gitterlinie wie auf den Quellblättern

// ==================================================================================
// SPIEGEL-PANELS (schreibgeschützte Ansicht der aktuell ausgewählten Belege)
//
// Beide Panels stehen rechts neben der jeweiligen Prüfung und geben – im Layout/Format
// des Quellblatts – exakt den ausgewählten Beleg wieder. Jede Wertzelle verweist per
// CHOOSE auf die zugehörige Tabelle des Quellblatts (MA: Index j über die Daten-Meta,
// FB: direkt die Periodennummer N), sodass die Ansicht der Auswahl dynamisch folgt.
// ==================================================================================

// evalDrawMAMirrorPanel zeichnet rechts neben der Mittelanforderungsprüfung eine
// Spiegelung der ausgewählten Mittelanforderung im Format des Blatts "IV. MA".
func (g *Generator) evalDrawMAMirrorPanel(ws string, top int, sel evalSelRefs) {
	const pLbl, pLC, pEUR = 12, 13, 14 // L | M | N
	g.setColWidth(ws, pLbl-1, 3.0)     // Spalte K als Abstand zur Tabelle
	g.setColWidth(ws, pLbl, 25.0)
	g.setColWidth(ws, pLC, 18.0)
	g.setColWidth(ws, pEUR, 18.0)

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
			HAlign: "center", VAlign: "center", NumFormat: numFmt, FillColor: MA_CLR_GRAY,
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
	hdr := StyleOptions{Bold: true, FillColor: MA_CLR_GRAY, HAlign: "center", VAlign: "center",
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
		labCell(cat, false, COLOR_WHITE)
		valCell(pLC, mirror(1, src), "#,##0.00", COLOR_WHITE)
		valCell(pEUR, mirror(2, src), `#,##0.00" €"`, COLOR_WHITE)
		r++
	}

	labCell("SUMME", true, MA_CLR_GRAY)
	bold(pLC, mirror(1, 18), "#,##0.00", MA_CLR_GRAY)
	bold(pEUR, mirror(2, 18), `#,##0.00" €"`, MA_CLR_GRAY)
	r += 2 // Leerzeile

	labCell("Gesamtbedarf an Mitteln:", false, COLOR_WHITE)
	valCell(pLC, mirror(1, 20), "#,##0.00", COLOR_WHITE)
	valCell(pEUR, mirror(2, 20), `#,##0.00" €"`, COLOR_WHITE)
	r++
	labCell("abzüglich Eigenmittel:", false, COLOR_WHITE)
	valCell(pLC, mirror(1, 21), "#,##0.00", COLOR_WHITE)
	valCell(pEUR, mirror(2, 21), `#,##0.00" €"`, COLOR_WHITE)
	r++
	labCell("abzüglich Drittmittel:", false, COLOR_WHITE)
	valCell(pLC, mirror(1, 22), "#,##0.00", COLOR_WHITE)
	valCell(pEUR, mirror(2, 22), `#,##0.00" €"`, COLOR_WHITE)
	r++
	// Saldo-Beschriftung dynamisch aus der Quelle spiegeln (Vorprojekt/Vorperiode).
	_ = g.setFormula(ws, cellName(pLbl, r), mirror(0, 23), StyleOptions{
		HAlign: "left", VAlign: "center", FillColor: COLOR_WHITE,
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
	})
	valCell(pLC, mirror(1, 23), "#,##0.00", COLOR_WHITE)
	valCell(pEUR, mirror(2, 23), `#,##0.00" €"`, COLOR_WHITE)
	r += 2 // Leerzeile

	labCell("KMW-Mittel Anforderung:", true, MA_CLR_KMW)
	bold(pLC, mirror(1, 25), "#,##0.00", MA_CLR_KMW)
	bold(pEUR, mirror(2, 25), `#,##0.00" €"`, MA_CLR_KMW)
	bottom := r

	g.styleOuterBorder(ws, top, pLbl, bottom, pEUR, 2, EV_CLR_BORDER)
}

// evalDrawFBMirrorPanel zeichnet rechts neben der Finanzberichtsprüfung eine
// Spiegelung des ausgewählten Finanzberichts (Hauptblock) im Format des Blatts
// "III. Finanzberichte". N = sel.fbSelNum wählt die Periode direkt per CHOOSE.
func (g *Generator) evalDrawFBMirrorPanel(ws string, top int, sel evalSelRefs) {
	const cLbl, cLC, cEUR, cKumLC, cKumEUR = 12, 13, 14, 15, 16 // L | M | N | O | P
	g.setColWidth(ws, cLbl-1, 3.0)
	g.setColWidth(ws, cLbl, 25.0)
	for c := cLC; c <= cKumEUR; c++ {
		g.setColWidth(ws, c, 18.0)
	}

	nAddr := sel.fbSelNum
	mirror := func(colOffset, srcRow int) string {
		parts := make([]string, 0, MA_PERIOD_COUNT)
		for p := 1; p <= MA_PERIOD_COUNT; p++ {
			colStart := START_COL + (p-1)*(TABLE_COLS+TABLE_SPACING)
			parts = append(parts, fmt.Sprintf("'%s'!%s", SHEET_NAME, absName(colStart+colOffset, srcRow)))
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, nAddr, strings.Join(parts, ","))
	}

	r := top

	// Titel folgt der Auswahl: "… – Periode N" bzw. Hinweis ohne Auswahl.
	titleFormula := fmt.Sprintf(
		`=IF(%s=0,"Aktueller Finanzbericht (keiner gewählt)","Aktueller Finanzbericht – Periode "&%s)`,
		nAddr, nAddr)
	_ = g.file.MergeCell(ws, cellName(cLbl, r), cellName(cKumEUR, r))
	_ = g.setStyle(ws, cellName(cLbl, r), cellName(cKumEUR, r), StyleOptions{
		Bold: true, Size: 11.0, FontColor: EV_CLR_BANNER_TXT, FillColor: EV_CLR_BANNER, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_CLR_BORDER,
	})
	_ = g.file.SetCellFormula(ws, cellName(cLbl, r), titleFormula)
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	infoRow := func(label, formula, numFmt string) {
		_ = g.setValue(ws, cellName(cLbl, r), label, StyleOptions{
			Bold: true, HAlign: "left", VAlign: "center",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
		c1, c2 := cellName(cLC, r), cellName(cKumEUR, r)
		_ = g.file.MergeCell(ws, c1, c2)
		_ = g.setStyle(ws, c1, c2, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: numFmt, FillColor: COLOR_TOTAL,
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
	infoRow("Durchschnittskurs:", mirror(1, 8), "0.000000")
	r += 2 // Leerzeile

	section := func(title string) {
		_ = g.mergeCells(ws, cellName(cLbl, r), cellName(cKumEUR, r), title, StyleOptions{
			Bold: true, FillColor: COLOR_HEADER, HAlign: "left", VAlign: "center",
			BorderTop: 2, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080",
		})
		r++
	}
	colHeaders := func(h0 string, hs []string) {
		hdr := StyleOptions{Bold: true, Size: 9.0, FillColor: COLOR_HEADER, HAlign: "center", VAlign: "center", WrapText: true,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080"}
		_ = g.setValue(ws, cellName(cLbl, r), h0, hdr)
		for i, h := range hs {
			_ = g.setValue(ws, cellName(cLC+i, r), h, hdr)
		}
		_ = g.file.SetRowHeight(ws, r, 26.0)
		r++
	}
	// dataRow spiegelt LC/EUR/Kum-LC/Kum-EUR (Offsets 1..4) einer Quellzeile.
	fmts := []string{"#,##0.00", `#,##0.00" €"`, "#,##0.00", `#,##0.00" €"`}
	dataRow := func(label string, labelIsFormula bool, srcRow int, fill string, bld bool) {
		lblOpts := StyleOptions{Bold: bld, HAlign: "left", VAlign: "center", FillColor: fill,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT}
		if labelIsFormula {
			_ = g.setFormula(ws, cellName(cLbl, r), label, lblOpts)
		} else {
			_ = g.setValue(ws, cellName(cLbl, r), label, lblOpts)
		}
		for i := 0; i < 4; i++ {
			_ = g.setFormula(ws, cellName(cLC+i, r), mirror(i+1, srcRow), StyleOptions{
				Bold: bld, HAlign: "right", VAlign: "center", NumFormat: fmts[i], FillColor: fill,
				BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
			})
		}
		r++
	}

	// ─── EINNAHMEN ───
	section("Einnahmen")
	colHeaders("Typ / ID", []string{"Einnahmen (LC)", "Einnahmen (EUR)", "Kum. Einnahmen (LC)", "Kum. Einnahmen (EUR)"})
	dataRow(mirror(0, 11), true, 11, COLOR_WHITE, false) // Vorperiodensaldo (Label gespiegelt)
	for i, t := range TYPE_NAMES {
		dataRow(t, false, 12+i, COLOR_WHITE, false)
	}
	dataRow("Gesamteinnahmen", false, 16, COLOR_TOTAL, true)

	// ─── AUSGABEN ───
	// Positionsbasiert: eine Zeile je Kostenposition (Anzahl folgt dem Budget).
	section("Ausgaben")
	colHeaders("ID", []string{"Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"})
	nPos := g.budgetExpenseCount()
	for i := 0; i < nPos; i++ {
		dataRow(mirror(0, FB_AUSG_FIRST_ROW+i), true, FB_AUSG_FIRST_ROW+i, COLOR_WHITE, false) // ID gespiegelt
	}
	gesamtAusgRow := FB_AUSG_FIRST_ROW + nPos // = ausgTotalsRow auf dem FB-Blatt
	dataRow("Gesamtausgaben", false, gesamtAusgRow, COLOR_TOTAL, true)
	r++ // Leerzeile

	// ─── SALDO DES FINANZBERICHTS ───
	_ = g.setValue(ws, cellName(cLbl, r), "Saldo des Finanzberichts", StyleOptions{
		Bold: true, HAlign: "left", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
	})
	saldoSrcRow := gesamtAusgRow + 2 // FB-Saldo liegt zwei Zeilen unter Gesamtausgaben
	for i := 0; i < 4; i++ {
		_ = g.setFormula(ws, cellName(cLC+i, r), mirror(i+1, saldoSrcRow), StyleOptions{
			Bold: true, HAlign: "right", VAlign: "center", NumFormat: fmts[i],
			BorderTop: 6, BorderBottom: 6, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
	}
	bottom := r

	g.styleOuterBorder(ws, top, cLbl, bottom, cKumEUR, 2, EV_CLR_BORDER)
}
