package vorpruefung

import (
	"fmt"
	"strings"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

// EV_HELP_COL ist eine weit rechts liegende, ausgeblendete Helferspalte des
// Auswertungsblatts (hält z. B. den Index der aktuell gewählten MA-Tabelle).
const EV_HELP_COL = 30

const EV_GRID_LIGHT = "D3D3D3" // helle Gitterlinie wie auf den Quellblättern

// Spalten des FB-Spiegel-Panels (rechts neben der Finanzberichtsprüfung).
const (
	EVFBMirrorColLabel  = 12 // L
	EVFBMirrorColLC     = 13 // M
	EVFBMirrorColEUR    = 14 // N
	EVFBMirrorColKumLC  = 15 // O
	EVFBMirrorColKumEUR = 16 // P
)

// ─── Teil B: Layout-Dokumentation ─────────────────────────────────────────────
/*
  SPIEGEL-PANEL (schreibgeschützte Ansicht des aktuell gewählten Finanzberichts)

  Das Panel steht rechts neben der Finanzberichtsprüfung und gibt – im Layout/Format
  des Blatts "III. Finanzberichte" – exakt den ausgewählten Finanzbericht wieder.
  Jede Wertzelle verweist per CHOOSE(N, …) auf die zugehörige Periodenspalte des
  Quellblatts (N = sel.fbSelNum), sodass die Ansicht der Auswahl dynamisch folgt.

  | Spalte L (Label)      | M (LC) | N (EUR) | O (Kum LC) | P (Kum EUR) |
  |-----------------------|--------|---------|------------|-------------|
  | Kopf (Periode/Von/…)  | gespiegelt (merged M:P)                    |
  | Einnahmen (Typen)     | =CHOOSE… je Offset 1..4                    |
  | Ausgaben (Positionen) | =CHOOSE… je Offset 1..4                    |
  | Saldo des FB          | =CHOOSE… je Offset 1..4                    |
*/

// ==================================================================================
// FB-Spiegel-Panel
// ==================================================================================

// evalDrawFBMirrorPanel zeichnet rechts neben der Finanzberichtsprüfung eine
// Spiegelung des ausgewählten Finanzberichts (Hauptblock). N = sel.fbSelNum wählt
// die Periode direkt per CHOOSE.
func (g *Generator) evalDrawFBMirrorPanel(ws string, top int, sel evalSelRefs) {
	cLbl, cLC, cKumEUR := EVFBMirrorColLabel, EVFBMirrorColLC, EVFBMirrorColKumEUR

	g.setColWidth(ws, cLbl-1, 3.0)
	g.setColWidth(ws, cLbl, 30.0)
	for c := cLC; c <= cKumEUR; c++ {
		g.setColWidth(ws, c, 21.0)
	}

	nAddr := sel.fbSelNum
	// mirror spiegelt eine Quellzelle (colOffset innerhalb der Periode, srcRow) über
	// CHOOSE(N, …) je Finanzbericht-Periode.
	mirror := func(colOffset, srcRow int) string {
		parts := make([]string, 0, FBPeriodenAnzahl)
		for p := 1; p <= FBPeriodenAnzahl; p++ {
			colStart := FBStartCol + (p-1)*(FBTableCols+FBTableSpacing)
			parts = append(parts, fmt.Sprintf("'%s'!%s", FBSheetName, absName(colStart+colOffset, srcRow)))
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, nAddr, strings.Join(parts, ","))
	}

	r := top

	// Titel folgt der Auswahl: "… – Periode N" bzw. Hinweis ohne Auswahl.
	titleFormula := fmt.Sprintf(
		`=IF(%s=0,"Aktueller Finanzbericht (keiner gewählt)","Aktueller Finanzbericht – Periode "&%s)`,
		nAddr, nAddr)
	_ = g.file.MergeCell(ws, cellName(cLbl, r), cellName(cKumEUR, r))
	_ = g.setStyle(ws, cellName(cLbl, r), cellName(cKumEUR, r), EVMirrorTitleStyle)
	_ = g.file.SetCellFormula(ws, cellName(cLbl, r), titleFormula)
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	infoRow := func(label, formula, numFmt string) {
		_ = g.setValue(ws, cellName(cLbl, r), label, EVMirrorInfoLabelStyle)
		c1, c2 := cellName(cLC, r), cellName(cKumEUR, r)
		_ = g.file.MergeCell(ws, c1, c2)
		_ = g.setStyle(ws, c1, c2, StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: numFmt, FillColor: FBClrTotal,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
		_ = g.file.SetCellFormula(ws, c1, formula)
	}
	infoRow("Periode:", mirror(FBOffLC, FBRowPeriode), "")
	r++
	infoRow("Von:", mirror(FBOffLC, FBRowVon), "DD.MM.YYYY")
	r++
	infoRow("Bis:", mirror(FBOffLC, FBRowBis), "DD.MM.YYYY")
	r++
	infoRow("Zeitraum:", mirror(FBOffLC, FBRowZeitraum), `0" Monate"`)
	r++
	infoRow("Durchschnittskurs:", mirror(FBOffLC, FBRowKurs), "0.000000")
	r += 2 // Leerzeile

	section := func(title string) {
		_ = g.mergeCells(ws, cellName(cLbl, r), cellName(cKumEUR, r), title, EVMirrorSectionStyle)
		r++
	}
	colHeaders := func(h0 string, hs []string) {
		_ = g.setValue(ws, cellName(cLbl, r), h0, EVMirrorColHeaderStyle)
		for i, h := range hs {
			_ = g.setValue(ws, cellName(cLC+i, r), h, EVMirrorColHeaderStyle)
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
	dataRow(mirror(FBOffLabel, FBRowSaldoVortrag), true, FBRowSaldoVortrag, FBClrWhite, false, "left") // Vorperiodensaldo (Label gespiegelt)
	for i, t := range TYPE_NAMES {
		dataRow(t, false, FBRowIncomeStart+i, FBClrWhite, false, "left")
	}
	gesamtEinnahmenRow := FBRowIncomeStart + len(TYPE_NAMES)
	dataRow("Gesamteinnahmen", false, gesamtEinnahmenRow, FBClrTotal, true, "left")

	// ─── AUSGABEN ───
	// Positionsbasiert: eine Zeile je Kostenposition (Anzahl folgt dem Budget).
	section("Ausgaben")
	colHeaders("ID", []string{"Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"})
	nPos := g.budgetExpenseCount()
	for i := 0; i < nPos; i++ {
		dataRow(mirror(FBOffLabel, FB_AUSG_FIRST_ROW+i), true, FB_AUSG_FIRST_ROW+i, FBClrWhite, false, "center") // ID gespiegelt
	}
	gesamtAusgRow := FB_AUSG_FIRST_ROW + nPos // = ausgTotalsRow auf dem FB-Blatt
	dataRow("Gesamtausgaben", false, gesamtAusgRow, FBClrTotal, true, "left")
	r++ // Leerzeile

	// ─── SALDO DES FINANZBERICHTS ───
	_ = g.setValue(ws, cellName(cLbl, r), "Saldo des Finanzberichts", EVMirrorInfoLabelStyle)
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
