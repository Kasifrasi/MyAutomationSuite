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
	// CHOOSE(N, …) je Finanzbericht-Periode. Nur noch für Zellen ohne benannten
	// Bereich (FB-Ausgabenpositionen, Labels/IDs, Kum-Saldo) verwendet.
	mirror := func(colOffset, srcRow int) string {
		parts := make([]string, 0, FBPeriodenAnzahl)
		for p := 1; p <= FBPeriodenAnzahl; p++ {
			colStart := FBStartCol + (p-1)*(FBTableCols+FBTableSpacing)
			parts = append(parts, fmt.Sprintf("'%s'!%s", FBSheetName, absName(colStart+colOffset, srcRow)))
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, nAddr, strings.Join(parts, ","))
	}
	// mirrorOut/mirrorInp spiegeln über die benannten Perioden-Bereiche (Registry
	// First) statt über feste Zellbezüge auf das FB-Blatt.
	mirrorOut := func(fac OutputFactory) string {
		parts := make([]string, 0, FBPeriodenAnzahl)
		for p := 1; p <= FBPeriodenAnzahl; p++ {
			parts = append(parts, fac.Get(p).NamedRange)
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, nAddr, strings.Join(parts, ","))
	}
	mirrorInp := func(fac InputFactory) string {
		parts := make([]string, 0, FBPeriodenAnzahl)
		for p := 1; p <= FBPeriodenAnzahl; p++ {
			parts = append(parts, fac.Get(p).NamedRange)
		}
		return fmt.Sprintf(`=IFERROR(CHOOSE(%s,%s),"")`, nAddr, strings.Join(parts, ","))
	}
	// mirrorOutVals baut die vier Wertformeln (LC, EUR, Kum-LC, Kum-EUR) einer
	// Zeile aus den benannten Perioden-Bereichen.
	mirrorOutVals := func(facs [4]OutputFactory) [4]string {
		var out [4]string
		for i, f := range facs {
			out[i] = mirrorOut(f)
		}
		return out
	}
	// mirrorPosVals baut die vier Wertformeln positionsbasiert (für Zellen ohne
	// benannten Bereich: FB-Ausgaben, Kum-Saldo).
	mirrorPosVals := func(srcRow int) [4]string {
		var out [4]string
		for i := 0; i < 4; i++ {
			out[i] = mirror(i+1, srcRow)
		}
		return out
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
	infoRow("Periode:", mirrorOut(Registry.OutputFBPeriode), "")
	r++
	infoRow("Von:", mirrorInp(Registry.InputFBVon), "DD.MM.YYYY")
	r++
	infoRow("Bis:", mirrorInp(Registry.InputFBBis), "DD.MM.YYYY")
	r++
	infoRow("Zeitraum:", mirrorOut(Registry.OutputFBZeitraum), `0" Monate"`)
	r++
	infoRow("Durchschnittskurs:", mirrorOut(Registry.OutputFBKurs), "0.000000")
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
	// dataRow spiegelt LC/EUR/Kum-LC/Kum-EUR (vals) einer Quellzeile.
	fmts := []string{"#,##0.00", `#,##0.00" €"`, "#,##0.00", `#,##0.00" €"`}
	dataRow := func(label string, labelIsFormula bool, vals [4]string, fill string, bld bool, align string) {
		lblOpts := StyleOptions{Bold: bld, HAlign: align, VAlign: "center", FillColor: fill,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT}
		if labelIsFormula {
			_ = g.setFormula(ws, cellName(cLbl, r), label, lblOpts)
		} else {
			_ = g.setValue(ws, cellName(cLbl, r), label, lblOpts)
		}
		for i := 0; i < 4; i++ {
			_ = g.setFormula(ws, cellName(cLC+i, r), vals[i], StyleOptions{
				Bold: bld, HAlign: "right", VAlign: "center", NumFormat: fmts[i], FillColor: fill,
				BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
			})
		}
		r++
	}

	// ─── EINNAHMEN ───
	section("Einnahmen")
	colHeaders("Typ / ID", []string{"Einnahmen (LC)", "Einnahmen (EUR)", "Kum. Einnahmen (LC)", "Kum. Einnahmen (EUR)"})
	// Vorperiodensaldo (Label positionsbasiert gespiegelt, Werte über Named Ranges)
	dataRow(mirror(FBOffLabel, FBRowSaldoVortrag), true,
		mirrorOutVals([4]OutputFactory{Registry.OutputFBVSaldoLC, Registry.OutputFBVSaldoEUR, Registry.OutputFBVSaldoKumLC, Registry.OutputFBVSaldoKumEUR}),
		FBClrWhite, false, "left")
	for i, t := range TYPE_NAMES {
		dataRow(t, false, mirrorOutVals(evalFBIncomeFactories(i)), FBClrWhite, false, "left")
	}
	dataRow("Gesamteinnahmen", false,
		mirrorOutVals([4]OutputFactory{Registry.OutputFBGEinnahmenLC, Registry.OutputFBGEinnahmenEUR, Registry.OutputFBKumGEinnahmenLC, Registry.OutputFBKumGEinnahmenEUR}),
		FBClrTotal, true, "left")

	// ─── AUSGABEN ───
	// Positionsbasiert: eine Zeile je Kostenposition (Anzahl folgt dem Budget).
	// Die FB-Ausgabenzellen tragen keine benannten Bereiche (Tabellenzellen) und
	// werden daher positionsbasiert gespiegelt.
	section("Ausgaben")
	colHeaders("ID", []string{"Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"})
	nPos := g.budgetExpenseCount()
	for i := 0; i < nPos; i++ {
		dataRow(mirror(FBOffLabel, FB_AUSG_FIRST_ROW+i), true, mirrorPosVals(FB_AUSG_FIRST_ROW+i), FBClrWhite, false, "center") // ID gespiegelt
	}
	gesamtAusgRow := FB_AUSG_FIRST_ROW + nPos // = ausgTotalsRow auf dem FB-Blatt
	dataRow("Gesamtausgaben", false, mirrorPosVals(gesamtAusgRow), FBClrTotal, true, "left")
	r++ // Leerzeile

	// ─── SALDO DES FINANZBERICHTS ───
	_ = g.setValue(ws, cellName(cLbl, r), "Saldo des Finanzberichts", EVMirrorInfoLabelStyle)
	saldoSrcRow := gesamtAusgRow + 2 // FB-Saldo liegt zwei Zeilen unter Gesamtausgaben
	// LC/EUR über benannte Bereiche; Kum-LC/Kum-EUR haben keinen Named Range.
	saldoVals := [4]string{
		mirrorOut(Registry.OutputFBSaldoLC),
		mirrorOut(Registry.OutputFBSaldoEUR),
		mirror(3, saldoSrcRow),
		mirror(4, saldoSrcRow),
	}
	for i := 0; i < 4; i++ {
		_ = g.setFormula(ws, cellName(cLC+i, r), saldoVals[i], StyleOptions{
			Bold: true, HAlign: "right", VAlign: "center", NumFormat: fmts[i],
			BorderTop: 6, BorderBottom: 6, BorderLeft: 1, BorderRight: 1, BorderColor: EV_GRID_LIGHT,
		})
	}
	bottom := r

	g.styleOuterBorder(ws, top, cLbl, bottom, cKumEUR, 2, EV_CLR_BORDER)
}
