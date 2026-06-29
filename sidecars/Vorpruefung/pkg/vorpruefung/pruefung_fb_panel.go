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

// evalDrawFBMirrorPanel zeichnet rechts neben der Finanzberichtsprüfung eine
// Spiegelung des ausgewählten Finanzberichts (Hauptblock) im Format des Blatts
// "III. Finanzberichte". N = sel.fbSelNum wählt die Periode direkt per CHOOSE.
func (g *Generator) evalDrawFBMirrorPanel(ws string, top int, sel evalSelRefs) {
	const cLbl, cLC, cEUR, cKumLC, cKumEUR = 12, 13, 14, 15, 16 // L | M | N | O | P
	g.setColWidth(ws, cLbl-1, 3.0)
	g.setColWidth(ws, cLbl, 30.0)
	for c := cLC; c <= cKumEUR; c++ {
		g.setColWidth(ws, c, 21.0)
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
	dataRow := func(label string, labelIsFormula bool, srcRow int, fill string, bld bool, align string) {
		lblOpts := StyleOptions{Bold: bld, HAlign: align, VAlign: "center", FillColor: fill,
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
	dataRow(mirror(0, 11), true, 11, COLOR_WHITE, false, "left") // Vorperiodensaldo (Label gespiegelt)
	for i, t := range TYPE_NAMES {
		dataRow(t, false, 12+i, COLOR_WHITE, false, "left")
	}
	dataRow("Gesamteinnahmen", false, 16, COLOR_TOTAL, true, "left")

	// ─── AUSGABEN ───
	// Positionsbasiert: eine Zeile je Kostenposition (Anzahl folgt dem Budget).
	section("Ausgaben")
	colHeaders("ID", []string{"Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"})
	nPos := g.budgetExpenseCount()
	for i := 0; i < nPos; i++ {
		dataRow(mirror(0, FB_AUSG_FIRST_ROW+i), true, FB_AUSG_FIRST_ROW+i, COLOR_WHITE, false, "center") // ID gespiegelt
	}
	gesamtAusgRow := FB_AUSG_FIRST_ROW + nPos // = ausgTotalsRow auf dem FB-Blatt
	dataRow("Gesamtausgaben", false, gesamtAusgRow, COLOR_TOTAL, true, "left")
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
