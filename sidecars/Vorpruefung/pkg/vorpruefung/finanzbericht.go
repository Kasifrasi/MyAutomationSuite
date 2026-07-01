package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

const (
	// Sheet
	FBSheetName = constants.VPSheetFINANZBERICHTE
	FBTabColor  = "FFFF00" // Gelb

	// Ein Bericht (Periode) belegt FBTableCols Spalten; zwischen zwei Perioden
	// liegen FBTableSpacing Leerspalten (Platz für den ➤-Trennpfeil).
	FBTableCols    = 5
	FBTableSpacing = 2

	// Ursprung der ersten Periode (Spalte B / Zeile 5).
	FBStartCol = 2
	FBStartRow = 5

	// Dynamische Auswahllisten (aus dem Budget) für die Detail-Einnahmentabellen.
	FBNameGeberList = "Geber_Liste"
	FBNameIDList    = "Budget_ID_Liste"

	// Cross-Perioden-/Sheet-Named-Ranges. Diese speisen Folgeperioden sowie die
	// Prüf- und MA-Blätter und sind deshalb bewusst als feste Namen zentralisiert.
	FBNameKursFmt     = "FB_Kurs_%d"
	FBNameSaldoLCFmt  = "FB_SaldoLC_%d"
	FBNameSaldoEURFmt = "FB_SaldoEUR_%d"

	// Spalten-Offsets innerhalb einer Periode (relativ zu colStart).
	// Hauptblock (Einnahmen-/Ausgaben-Übersicht):
	FBOffLabel  = 0 // Label / ID
	FBOffLC     = 1 // Betrag (LC)
	FBOffEUR    = 2 // Betrag (EUR)
	FBOffKumLC  = 3 // Kum. (LC)
	FBOffKumEUR = 4 // Kum. (EUR)

	// Detail-Einnahmentabellen (andere Spaltenbelegung als der Hauptblock):
	FBDetOffTyp   = 0
	FBDetOffGeber = 1
	FBDetOffLC    = 2
	FBDetOffEUR   = 3
	FBDetOffKurs  = 4

	// Feste Zeilen-Offsets des Kopf-/Einnahmenblocks (relativ zu FBStartRow).
	// Der Rest (Ausgaben, Saldo, Aufschlüsselung, Detailtabellen) hängt von der
	// Anzahl der Einnahme-Typen und Ausgabe-Positionen ab und wird berechnet.
	FBRowPeriode      = FBStartRow - 1 // 4
	FBRowVon          = FBStartRow     // 5
	FBRowBis          = FBStartRow + 1 // 6
	FBRowZeitraum     = FBStartRow + 2 // 7
	FBRowKurs         = FBStartRow + 3 // 8
	FBRowEinnahmenHdr = FBStartRow + 4 // 9
	FBRowEinnColHdr   = FBStartRow + 5 // 10
	FBRowSaldoVortrag = FBStartRow + 6 // 11
	FBRowIncomeStart  = FBStartRow + 7 // 12

	// Detailtabellen: Anzahl der Datenzeilen.
	FBDetailRowsExplizit = 5 // explizite Kurseingabe (Saldozeile + 4 Eingaben)
	FBDetailRowsWK       = 6 // Durchschnittskurs

	// FB_AUSG_FIRST_ROW / FB_INCOME_FIRST_ROW dokumentieren die erste Datenzeile
	// der Ausgaben- bzw. Einnahmen-Typzeilen bei Standardbelegung (4 Typen).
	FB_AUSG_FIRST_ROW   = 19
	FB_INCOME_FIRST_ROW = FBRowIncomeStart
)

// ─── Domänen-Kategorien (auch von den Prüf-Blättern genutzt) ──────────────────

var (
	TYPE_NAMES = []string{
		"Eigenmittel",
		"Drittmittel",
		"KMW-Mittel",
		"Zinsertraege",
	}

	EXPENSE_CATEGORIES = []string{
		"Bauausgaben",
		"Investitionen",
		"Personalkosten",
		"Projektaktivitaeten",
		"Projektverwaltung",
		"Evaluierung",
		"Audit",
		"Reserve",
	}

	INFO_CATEGORIES = []string{
		"Bank",
		"Kasse",
		"Sonstiges",
	}
)

// ─── Teil B: Layout-Dokumentation ─────────────────────────────────────────────
/*
  LAYOUT FINANZBERICHT (eine Periode, Spalten colStart .. colStart+4):

  | Zeile              | B/Label (colStart) | C (LC/+1)        | D (EUR/+2)       | E (KumLC/+3)     | F (KumEUR/+4) |
  |--------------------|--------------------|------------------|------------------|------------------|---------------|
  | 4  Periode         | Periode:           | [Periode N          (merged C:D)] |
  | 5  Von             | Von:               | [Inp Von            (merged C:D, Datum)] |
  | 6  Bis             | Bis:               | [Inp Bis            (merged C:D, Datum)] |
  | 7  Zeitraum        | Zeitraum:          | [= Monate           (merged C:D)] |
  | 8  Kurs            | Durchschnittskurs: | [= Kurs             (merged C:D)] |
  | 9  Einnahmen       | «Einnahmen» (Sektionskopf, merged B:F)                                       |
  | 10 Spaltenköpfe    | Typ / ID           | Einn. (LC)       | Einn. (EUR)      | Kum. (LC)        | Kum. (EUR)    |
  | 11 Vorperiodensaldo| Vor(projekt|perioden)saldo | …       | …                | …                | …             |
  | 12.. Typ-Zeilen    | (leer/API)         | =SUMIF…          | =SUMIF…          | Kum.             | Kum.          |
  | +  Gesamteinnahmen | Gesamteinnahmen    | =SUM             | =SUM             | =SUM             | =SUM          |
  | +  Ausgaben        | «Ausgaben» (Sektionskopf)                                                    |
  | +  Ausg.-Kopf      | ID                 | Ausg. (LC)       | Ausg. (EUR)      | Kum. (LC)        | Kum. (EUR)    |
  | +  Ausg.-Daten     | ID/Formel          | [Inp LC]         | =LC/Kurs         | Kum.             | Kum.          |
  | +  Gesamtausgaben  | Gesamtausgaben     | =SUBTOTAL        | =SUBTOTAL        | =SUBTOTAL        | =SUBTOTAL     |
  | +  Saldo           | Saldo des Finanzberichts | =Einn-Ausg | …               | …                | …             |
  | +  Aufschlüsselung | Aufschluesselung:  |                  |                  |                  |               |
  | +  Bank/Kasse/Sonst| Kategorie          | [Inp LC]         | =Verteilung EUR  |                  |               |
  | +  Differenz       | Differenz (Pruefung): | =Kontrolle    | =Kontrolle       |                  |               |
  |--------------------------------------------------------------------------------------------------------------|
  |  Danach zwei Detail-Einnahmentabellen (explizite Kurseingabe / Durchschnittskurs).                          |

  Perioden 2..18 werden gruppiert und eingeklappt (ausgeblendet).
*/

// fbLayout hält alle absoluten Zeilen-/Spaltennummern einer Periode. Es wird
// einmal pro Periode berechnet (computeFBLayout) und sowohl von den Draw- als
// auch den Bind-Funktionen als einzige Quelle für Koordinaten genutzt.
type fbLayout struct {
	periode      int
	isFollowUp   bool
	colStart     int
	prevColStart int

	// Spalten (absolut) – Hauptblock
	colLabel  int
	colLC     int
	colEUR    int
	colKumLC  int
	colKumEUR int

	// Zeilen – Hauptblock
	rowPeriode         int
	rowVon             int
	rowBis             int
	rowZeitraum        int
	rowKurs            int
	rowEinnahmenHdr    int
	rowEinnColHdr      int
	rowSaldoVortrag    int
	rowIncomeStart     int
	incomeCount        int
	rowGesamtEinnahmen int
	rowAusgabenHdr     int
	rowAusgTblHdr      int
	ausgCount          int
	rowAusgTotals      int
	rowSaldoFB         int
	rowAufschlLabel    int
	rowAufschlStart    int
	rowAufschlEnd      int
	rowDifferenz       int
	rowBlockEnd        int // untere Kante des Außenrahmens

	// Zeilen – Detailtabellen
	rowDetail1Label  int
	rowDetail1Hdr    int
	rowDetail1Totals int
	rowDetail2Label  int
	rowDetail2Hdr    int
	rowDetail2Totals int
}

func computeFBLayout(periode, incomeCount, expenseCount int) fbLayout {
	colStart := FBStartCol + (periode-1)*(FBTableCols+FBTableSpacing)

	l := fbLayout{
		periode:      periode,
		isFollowUp:   periode > 1,
		colStart:     colStart,
		prevColStart: colStart - (FBTableCols + FBTableSpacing),

		colLabel:  colStart + FBOffLabel,
		colLC:     colStart + FBOffLC,
		colEUR:    colStart + FBOffEUR,
		colKumLC:  colStart + FBOffKumLC,
		colKumEUR: colStart + FBOffKumEUR,

		rowPeriode:      FBRowPeriode,
		rowVon:          FBRowVon,
		rowBis:          FBRowBis,
		rowZeitraum:     FBRowZeitraum,
		rowKurs:         FBRowKurs,
		rowEinnahmenHdr: FBRowEinnahmenHdr,
		rowEinnColHdr:   FBRowEinnColHdr,
		rowSaldoVortrag: FBRowSaldoVortrag,
		rowIncomeStart:  FBRowIncomeStart,
		incomeCount:     incomeCount,
		ausgCount:       expenseCount,
	}

	l.rowGesamtEinnahmen = l.rowIncomeStart + incomeCount
	l.rowAusgabenHdr = l.rowGesamtEinnahmen + 1
	l.rowAusgTblHdr = l.rowAusgabenHdr + 1
	l.rowAusgTotals = l.rowAusgTblHdr + expenseCount + 1
	l.rowSaldoFB = l.rowAusgTotals + 2 // eine Leerzeile dazwischen
	l.rowAufschlLabel = l.rowSaldoFB + 2
	l.rowAufschlStart = l.rowAufschlLabel + 1
	l.rowAufschlEnd = l.rowAufschlStart + len(INFO_CATEGORIES) - 1
	l.rowDifferenz = l.rowAufschlEnd + 2
	l.rowBlockEnd = l.rowDifferenz

	// Detailtabellen (zwei Leerzeilen Abstand nach dem Hauptblock).
	l.rowDetail1Label = l.rowDifferenz + 3
	l.rowDetail1Hdr = l.rowDetail1Label + 1
	l.rowDetail1Totals = l.rowDetail1Hdr + FBDetailRowsExplizit + 1
	l.rowDetail2Label = l.rowDetail1Totals + 2
	l.rowDetail2Hdr = l.rowDetail2Label + 1
	l.rowDetail2Totals = l.rowDetail2Hdr + FBDetailRowsWK + 1

	return l
}

// ─── Teil C: Orchestrator ─────────────────────────────────────────────────────

// CreateFinanzberichteSheet initialisiert das Blatt "III. Finanzberichte" und
// zeichnet FBPeriodenAnzahl Berichtsperioden nebeneinander.
func (g *Generator) CreateFinanzberichteSheet() error {
	ws := FBSheetName
	f := g.file

	if _, err := f.NewSheet(ws); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Finanzberichte-Blatts: %w", err)
	}
	tabColor := FBTabColor
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	incomeCount := g.budgetIncomeCount()
	expenseCount := g.budgetExpenseCount()

	for p := 1; p <= FBPeriodenAnzahl; p++ {
		lay := computeFBLayout(p, incomeCount, expenseCount)

		// Teil D: Zeichnen (Layout, Farben, statische Texte)
		g.fbDrawPeriod(ws, lay)

		// Teil E: Binden (Formeln, Named Ranges, Tabellen, Validierungen)
		if err := g.fbBindPeriod(ws, Registry, lay); err != nil {
			return fmt.Errorf("fehler beim Binden von Periode %d: %w", p, err)
		}
	}

	g.fbCollapseFollowUpPeriods(ws)
	return nil
}

// fbCollapseFollowUpPeriods gruppiert und blendet die Perioden 2..N ein.
func (g *Generator) fbCollapseFollowUpPeriods(ws string) {
	f := g.file
	for p := 2; p <= FBPeriodenAnzahl; p++ {
		colStart := FBStartCol + (p-1)*(FBTableCols+FBTableSpacing)
		for c := colStart; c <= colStart+FBTableCols-1; c++ {
			_ = f.SetColOutlineLevel(ws, colLetter(c), 1)
			_ = f.SetColVisible(ws, colLetter(c), false)
		}
	}
}

// ─── Teil D: Draw-Funktionen (nur visuell) ────────────────────────────────────

// fbDrawPeriod zeichnet das komplette Layout einer Periode: Spaltenbreiten,
// statische Labels/Überschriften und alle Zell-Styles. Es werden keine Formeln,
// Named Ranges oder Validierungen gesetzt (siehe fbBindPeriod).
func (g *Generator) fbDrawPeriod(ws string, l fbLayout) {
	g.fbSetupColumnWidths(ws, l.colStart)

	if l.isFollowUp {
		g.fbDrawSeparatorArrow(ws, FBStartRow-2, l.colStart-1)
	}

	g.fbDrawPeriodHeader(ws, l)
	g.fbDrawEinnahmenSection(ws, l)
	g.fbDrawAusgabenSection(ws, l)
	g.fbDrawSaldo(ws, l)
	g.fbDrawAufschluesselung(ws, l)
	g.fbDrawDifferenz(ws, l)
	g.fbDrawDetailLabels(ws, l)

	// Äußerer Rahmen um den Hauptblock (ab Periodenkopf).
	_ = g.styleOuterBorder(ws, l.rowPeriode, l.colStart, l.rowBlockEnd, l.colStart+FBTableCols-1, 2, FBClrBorder)
}

func (g *Generator) fbSetupColumnWidths(ws string, colStart int) {
	g.setColWidth(ws, colStart+FBOffLabel, 30.43)
	g.setColWidth(ws, colStart+FBOffLC, 24.71)
	g.setColWidth(ws, colStart+FBOffEUR, 24.71)
	g.setColWidth(ws, colStart+FBOffKumLC, 24.71)
	g.setColWidth(ws, colStart+FBOffKumEUR, 24.71)
}

func (g *Generator) fbDrawSeparatorArrow(ws string, row, col int) {
	_ = g.setValue(ws, cellName(col, row), "➤", FBArrowStyle)
}

// fbDrawPeriodHeader zeichnet Periode/Von/Bis/Zeitraum/Kurs (Labels + leere,
// formatierte Wertzellen). Werte-Bindungen und Formeln folgen im Bind.
func (g *Generator) fbDrawPeriodHeader(ws string, l fbLayout) {
	// Periode
	_ = g.setValue(ws, cellName(l.colLabel, l.rowPeriode), "Periode:", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLC, l.rowPeriode), cellName(l.colEUR, l.rowPeriode),
		fmt.Sprintf("Periode %d", l.periode), FBPeriodValueStyle)

	// Von / Bis (Datums-Eingaben)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowVon), "Von:", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLC, l.rowVon), cellName(l.colEUR, l.rowVon), "", FBPeriodDatumStyle)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowBis), "Bis:", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLC, l.rowBis), cellName(l.colEUR, l.rowBis), "", FBPeriodDatumStyle)

	// Zeitraum (berechnete Monate)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowZeitraum), "Zeitraum:", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLC, l.rowZeitraum), cellName(l.colEUR, l.rowZeitraum), "", FBZeitraumStyle)

	// Durchschnittskurs (berechnet)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowKurs), "Durchschnittskurs:", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLC, l.rowKurs), cellName(l.colEUR, l.rowKurs), "", FBKursStyle)
}

// fbDrawEinnahmenSection zeichnet Sektionskopf, Spaltenköpfe, die
// Vorperiodensaldo-Zeile, die Typ-Zeilen und die Gesamteinnahmen-Zeile.
func (g *Generator) fbDrawEinnahmenSection(ws string, l fbLayout) {
	g.fbDrawSectionHeader(ws, l.rowEinnahmenHdr, l.colLabel, l.colKumEUR, "Einnahmen")

	g.fbDrawColumnHeaders(ws, l.rowEinnColHdr, l.colLabel, []string{
		"Einnahmen (LC)",
		"Einnahmen (EUR)",
		"Kum. Einnahmen (LC)",
		"Kum. Einnahmen (EUR)",
	})

	saldoLabel := "Vorprojektsaldo"
	if l.isFollowUp {
		saldoLabel = "Vorperiodensaldo"
	}
	g.fbDrawIncomeRow(ws, l.rowSaldoVortrag, l.colLabel, saldoLabel)

	for i := 0; i < l.incomeCount; i++ {
		// Label bleibt leer, damit die API es befüllt; nur ohne Config-Vorgaben
		// werden die Standard-Typennamen gezeichnet.
		label := ""
		if g.cfg.IncomeTypesCount == 0 && i < len(TYPE_NAMES) {
			label = TYPE_NAMES[i]
		}
		g.fbDrawIncomeRow(ws, l.rowIncomeStart+i, l.colLabel, label)
	}

	// Gesamteinnahmen (Label + Styles; Formeln im Bind)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowGesamtEinnahmen), "Gesamteinnahmen", FBIncomeTotalLabelStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowGesamtEinnahmen), cellName(l.colLC, l.rowGesamtEinnahmen), FBIncomeTotalLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowGesamtEinnahmen), cellName(l.colEUR, l.rowGesamtEinnahmen), FBIncomeTotalEURStyle)
	_ = g.setStyle(ws, cellName(l.colKumLC, l.rowGesamtEinnahmen), cellName(l.colKumLC, l.rowGesamtEinnahmen), FBIncomeTotalLCStyle)
	_ = g.setStyle(ws, cellName(l.colKumEUR, l.rowGesamtEinnahmen), cellName(l.colKumEUR, l.rowGesamtEinnahmen), FBIncomeTotalEURStyle)
}

// fbDrawAusgabenSection zeichnet Sektionskopf, Tabellenkopf, leere Datenzellen
// und die Gesamtausgaben-Zeile (nur Styles/Labels).
func (g *Generator) fbDrawAusgabenSection(ws string, l fbLayout) {
	g.fbDrawSectionHeader(ws, l.rowAusgabenHdr, l.colLabel, l.colKumEUR, "Ausgaben")

	headers := []string{"ID", "Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"}
	for i, h := range headers {
		_ = g.setValue(ws, cellName(l.colLabel+i, l.rowAusgTblHdr), h, FBAusgHdrStyle)
	}

	for i := 0; i < l.ausgCount; i++ {
		row := l.rowAusgTblHdr + 1 + i
		_ = g.setStyle(ws, cellName(l.colLabel, row), cellName(l.colLabel, row), FBAusgIDStyle)
		_ = g.setStyle(ws, cellName(l.colLC, row), cellName(l.colLC, row), FBAusgLCStyle)
		_ = g.setStyle(ws, cellName(l.colEUR, row), cellName(l.colEUR, row), FBAusgEURStyle)
		_ = g.setStyle(ws, cellName(l.colKumLC, row), cellName(l.colKumLC, row), FBAusgKumLCStyle)
		_ = g.setStyle(ws, cellName(l.colKumEUR, row), cellName(l.colKumEUR, row), FBAusgKumEURStyle)
	}

	// Gesamtausgaben (Label + Styles; Formeln im Bind)
	_ = g.setValue(ws, cellName(l.colLabel, l.rowAusgTotals), "Gesamtausgaben", FBAusgTotalLabelStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowAusgTotals), cellName(l.colLC, l.rowAusgTotals), FBAusgTotalLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowAusgTotals), cellName(l.colEUR, l.rowAusgTotals), FBAusgTotalEURStyle)
	_ = g.setStyle(ws, cellName(l.colKumLC, l.rowAusgTotals), cellName(l.colKumLC, l.rowAusgTotals), FBAusgTotalLCStyle)
	_ = g.setStyle(ws, cellName(l.colKumEUR, l.rowAusgTotals), cellName(l.colKumEUR, l.rowAusgTotals), FBAusgTotalEURStyle)
}

// fbDrawSaldo zeichnet die "Saldo des Finanzberichts"-Zeile (Label + Styles).
func (g *Generator) fbDrawSaldo(ws string, l fbLayout) {
	_ = g.setValue(ws, cellName(l.colLabel, l.rowSaldoFB), "Saldo des Finanzberichts", FBLabelBoldStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowSaldoFB), cellName(l.colLC, l.rowSaldoFB), FBSaldoLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowSaldoFB), cellName(l.colEUR, l.rowSaldoFB), FBSaldoEURStyle)
	_ = g.setStyle(ws, cellName(l.colKumLC, l.rowSaldoFB), cellName(l.colKumLC, l.rowSaldoFB), FBSaldoLCStyle)
	_ = g.setStyle(ws, cellName(l.colKumEUR, l.rowSaldoFB), cellName(l.colKumEUR, l.rowSaldoFB), FBSaldoEURStyle)
}

// fbDrawAufschluesselung zeichnet das Label und die Info-Zeilen (Bank/Kasse/…).
func (g *Generator) fbDrawAufschluesselung(ws string, l fbLayout) {
	_ = g.setValue(ws, cellName(l.colLabel, l.rowAufschlLabel), "Aufschluesselung:", FBLabelPlainStyle)

	for i, cat := range INFO_CATEGORIES {
		row := l.rowAufschlStart + i
		_ = g.setValue(ws, cellName(l.colLabel, row), cat, FBInfoLabelStyle)
		_ = g.setStyle(ws, cellName(l.colLC, row), cellName(l.colLC, row), FBInfoLCStyle)
		_ = g.setStyle(ws, cellName(l.colEUR, row), cellName(l.colEUR, row), FBInfoEURStyle)
	}
}

// fbDrawDifferenz zeichnet die Prüfzeile (Label + Styles).
func (g *Generator) fbDrawDifferenz(ws string, l fbLayout) {
	_ = g.setValue(ws, cellName(l.colLabel, l.rowDifferenz), "Differenz (Pruefung):", FBDiffLabelStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowDifferenz), cellName(l.colLC, l.rowDifferenz), FBDiffLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowDifferenz), cellName(l.colEUR, l.rowDifferenz), FBDiffEURStyle)
}

// fbDrawDetailLabels zeichnet die beiden Überschriften der Detailtabellen.
func (g *Generator) fbDrawDetailLabels(ws string, l fbLayout) {
	_ = g.mergeCells(ws, cellName(l.colLabel, l.rowDetail1Label), cellName(l.colLabel, l.rowDetail1Label),
		"Einnahmen (Explizite Kurseingabe)", FBLabelBoldStyle)
	_ = g.mergeCells(ws, cellName(l.colLabel, l.rowDetail2Label), cellName(l.colLabel, l.rowDetail2Label),
		"Einnahmen (Durchschnittskurs)", FBLabelBoldStyle)
}

// fbDrawSectionHeader zeichnet einen verbundenen Sektionskopf (z. B. "Einnahmen").
func (g *Generator) fbDrawSectionHeader(ws string, row, col1, col2 int, text string) {
	_ = g.mergeCells(ws, cellName(col1, row), cellName(col2, row), text, FBSectionHdrStyle)
}

// fbDrawColumnHeaders zeichnet die Kopfzeile der Einnahmen-Übersicht.
func (g *Generator) fbDrawColumnHeaders(ws string, row, col1 int, headers []string) {
	_ = g.setValue(ws, cellName(col1, row), "Typ / ID", FBColHdrLabelStyle)
	for i, h := range headers {
		_ = g.setValue(ws, cellName(col1+1+i, row), h, FBColHdrValStyle)
	}
}

// fbDrawIncomeRow zeichnet eine Einnahmen-Übersichtszeile (Label + 4 Wertzellen).
func (g *Generator) fbDrawIncomeRow(ws string, row, col1 int, label string) {
	_ = g.setValue(ws, cellName(col1, row), label, FBIncomeLabelStyle)
	for i := 0; i < 4; i++ {
		style := FBIncomeLCStyle
		if i%2 == 1 {
			style = FBIncomeEURStyle
		}
		_ = g.setStyle(ws, cellName(col1+1+i, row), cellName(col1+1+i, row), style)
	}
}

// ─── Teil E: Bind-Funktionen (Logik & Registry) ───────────────────────────────

// fbBindPeriod verknüpft die gezeichnete Periode mit Formeln, Named Ranges,
// Tabellen und Validierungen. Reihenfolge beachtet die Abhängigkeiten:
// Detailtabellen liefern Kurs und Einnahmen-Summen für den Hauptblock.
func (g *Generator) fbBindPeriod(ws string, reg *TemplateRegistry, l fbLayout) error {
	g.fbBindPeriodHeader(ws, reg, l)
	g.fbBindEinnahmen(ws, l)
	if err := g.fbBindAusgaben(ws, reg, l); err != nil {
		return err
	}
	g.fbBindSaldo(ws, l)
	g.fbBindAufschluesselung(ws, reg, l)
	g.fbBindDifferenz(ws, l)
	return g.fbBindDetailTables(ws, reg, l)
}

// fbBindPeriodHeader bindet Von/Bis-Eingaben und die Zeitraum-Formel.
func (g *Generator) fbBindPeriodHeader(ws string, reg *TemplateRegistry, l fbLayout) {
	_ = g.bindInputField(ws, l.rowVon, l.colLC, reg.InputFBVon.Get(l.periode))
	_ = g.bindInputField(ws, l.rowBis, l.colLC, reg.InputFBBis.Get(l.periode))

	vonCell := cellName(l.colLC, l.rowVon)
	bisCell := cellName(l.colLC, l.rowBis)
	_ = g.file.SetCellFormula(ws, cellName(l.colLC, l.rowZeitraum), fmt.Sprintf(
		`=IF(OR(%s="",%s=""),"",DATEDIF(%s,%s,"m")+1)`, vonCell, bisCell, vonCell, bisCell))
}

// fbBindEinnahmen bindet die Vorperiodensaldo- und die Gesamteinnahmen-Formeln.
// (Die Typ-Zeilen-Formeln benötigen die Detailtabellen und folgen dort.)
func (g *Generator) fbBindEinnahmen(ws string, l fbLayout) {
	f := g.file
	r := l.rowSaldoVortrag

	saldoVortragLC := fmt.Sprintf(`=ROUND(IF(%s="",0,%s),2)`, DB_NAME_SALDOVORTRAG_LW, DB_NAME_SALDOVORTRAG_LW)
	saldoVortragEUR := fmt.Sprintf(`=ROUND(IF(%s="",0,%s),2)`, DB_NAME_SALDOVORTRAG_EUR, DB_NAME_SALDOVORTRAG_EUR)

	if l.isFollowUp {
		_ = f.SetCellFormula(ws, cellName(l.colLC, r), fmt.Sprintf(`=ROUND(%s,2)`, fmt.Sprintf(FBNameSaldoLCFmt, l.periode-1)))
		_ = f.SetCellFormula(ws, cellName(l.colEUR, r), fmt.Sprintf(`=ROUND(%s,2)`, fmt.Sprintf(FBNameSaldoEURFmt, l.periode-1)))
	} else {
		_ = f.SetCellFormula(ws, cellName(l.colLC, r), saldoVortragLC)
		_ = f.SetCellFormula(ws, cellName(l.colEUR, r), saldoVortragEUR)
	}
	// Kumulierte Spalten der Saldozeile speisen sich immer aus dem Dashboard.
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, r), saldoVortragLC)
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, r), saldoVortragEUR)

	// Gesamteinnahmen (Vorperiodensaldo + Typ-Zeilen).
	lastRow := l.rowGesamtEinnahmen - 1
	rng := func(col int) string {
		return fmt.Sprintf("%s:%s", cellName(col, l.rowSaldoVortrag), cellName(col, lastRow))
	}
	gr := l.rowGesamtEinnahmen
	_ = f.SetCellFormula(ws, cellName(l.colLC, gr), fmt.Sprintf(`=ROUND(SUM(%s),2)`, rng(l.colLC)))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, gr), fmt.Sprintf(`=ROUND(SUM(%s),2)`, rng(l.colEUR)))
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, gr), fmt.Sprintf(`=ROUND(SUM(%s),2)`, rng(l.colKumLC)))
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, gr), fmt.Sprintf(`=ROUND(SUM(%s),2)`, rng(l.colKumEUR)))
}

// fbBindAusgaben legt die Ausgaben-Tabelle an und setzt ID-/EUR-/Kum-/Summenformeln.
func (g *Generator) fbBindAusgaben(ws string, reg *TemplateRegistry, l fbLayout) error {
	f := g.file
	ausgName := reg.TableFBAusgaben.Get(l.periode).Name
	rateAddr := absName(l.colLC, l.rowKurs)

	// Datenbereich für den VSTACK der Datei-Übersicht registrieren.
	dataRange := fmt.Sprintf("'%s'!%s:%s", ws,
		absName(l.colLabel, l.rowAusgTblHdr+1), absName(l.colKumEUR, l.rowAusgTblHdr+l.ausgCount))
	g.rangesAusgaben = append(g.rangesAusgaben, dataRange)

	if err := f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(l.colLabel, l.rowAusgTblHdr), cellName(l.colKumEUR, l.rowAusgTotals-1)),
		Name:           ausgName,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	}); err != nil {
		return err
	}

	for i := 0; i < l.ausgCount; i++ {
		row := l.rowAusgTblHdr + 1 + i

		// ID: bei Config leer (API befüllt), sonst per INDEX aus der Budget-ID-Liste.
		if g.cfg.ExpensePositionsCount > 0 {
			_ = f.SetCellValue(ws, cellName(l.colLabel, row), "")
		} else {
			_ = f.SetCellFormula(ws, cellName(l.colLabel, row),
				fmt.Sprintf(`=IFERROR(INDEX(%s, ROW() - %d), "")`, FBNameIDList, l.rowAusgTblHdr))
		}

		// EUR = LC / Kurs
		_ = f.SetCellFormula(ws, cellName(l.colEUR, row),
			fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, cellName(l.colLC, row), rateAddr))

		// Kumulierte Spalten
		if l.isFollowUp {
			_ = f.SetCellFormula(ws, cellName(l.colKumLC, row),
				fmt.Sprintf(`=ROUND(%s+%s,2)`, cellName(l.colLC, row), cellName(l.prevColStart+FBOffKumLC, row)))
			_ = f.SetCellFormula(ws, cellName(l.colKumEUR, row),
				fmt.Sprintf(`=ROUND(%s+%s,2)`, cellName(l.colEUR, row), cellName(l.prevColStart+FBOffKumEUR, row)))
		} else {
			_ = f.SetCellFormula(ws, cellName(l.colKumLC, row), fmt.Sprintf(`=ROUND(%s,2)`, cellName(l.colLC, row)))
			_ = f.SetCellFormula(ws, cellName(l.colKumEUR, row), fmt.Sprintf(`=ROUND(%s,2)`, cellName(l.colEUR, row)))
		}
	}

	// Gesamtausgaben (SUBTOTAL über den Datenbereich).
	subtotal := func(col int) string {
		rng := fmt.Sprintf("%s:%s", absName(col, l.rowAusgTblHdr+1), absName(col, l.rowAusgTblHdr+l.ausgCount))
		return fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s),2)`, rng)
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, l.rowAusgTotals), subtotal(l.colLC))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, l.rowAusgTotals), subtotal(l.colEUR))
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, l.rowAusgTotals), subtotal(l.colKumLC))
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, l.rowAusgTotals), subtotal(l.colKumEUR))

	return nil
}

// fbBindSaldo setzt die Saldo-Formeln und benennt die LC/EUR-Zellen.
func (g *Generator) fbBindSaldo(ws string, l fbLayout) {
	f := g.file
	r := l.rowSaldoFB

	diff := func(einnahmenCol, ausgabenCol int) string {
		return fmt.Sprintf(`=ROUND(IFERROR(%s-%s,""),2)`,
			cellName(einnahmenCol, l.rowGesamtEinnahmen), cellName(ausgabenCol, l.rowAusgTotals))
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, r), diff(l.colLC, l.colLC))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, r), diff(l.colEUR, l.colEUR))
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, r), diff(l.colKumLC, l.colKumLC))
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, r), diff(l.colKumEUR, l.colKumEUR))

	g.dbUpsertNamedRange(ws, fmt.Sprintf(FBNameSaldoLCFmt, l.periode), l.colLC, r)
	g.dbUpsertNamedRange(ws, fmt.Sprintf(FBNameSaldoEURFmt, l.periode), l.colEUR, r)
}

// fbBindAufschluesselung bindet die LC-Eingaben und verteilt den EUR-Saldo.
func (g *Generator) fbBindAufschluesselung(ws string, reg *TemplateRegistry, l fbLayout) {
	f := g.file
	saldoEUR := cellName(l.colEUR, l.rowSaldoFB)
	sumLC := fmt.Sprintf(`ROUND(SUM(%s:%s), 2)`,
		cellName(l.colLC, l.rowAufschlStart), cellName(l.colLC, l.rowAufschlEnd))

	for i, cat := range INFO_CATEGORIES {
		row := l.rowAufschlStart + i
		isLast := i == len(INFO_CATEGORIES)-1
		currentLC := cellName(l.colLC, row)

		var field InputField
		switch cat {
		case "Bank":
			field = reg.InputFBAufschlBankLC.Get(l.periode)
		case "Kasse":
			field = reg.InputFBAufschlKasseLC.Get(l.periode)
		default:
			field = reg.InputFBAufschlSonstigesLC.Get(l.periode)
		}
		_ = g.bindInputField(ws, row, l.colLC, field)

		var formulaEUR string
		if !isLast {
			remainingLCs := fmt.Sprintf(`SUM(%s:%s)`, cellName(l.colLC, row+1), cellName(l.colLC, l.rowAufschlEnd))
			prevSum := "0"
			if i > 0 {
				prevSum = fmt.Sprintf(`SUM(%s:%s)`, cellName(l.colEUR, l.rowAufschlStart), cellName(l.colEUR, row-1))
			}
			formulaEUR = fmt.Sprintf(`=IF(%s=0, 0, IF(ROUND(%s, 2)=0, ROUND(%s - %s, 2), ROUND(%s / %s * %s, 2)))`,
				currentLC, remainingLCs, saldoEUR, prevSum, currentLC, sumLC, saldoEUR)
		} else {
			prevSum := fmt.Sprintf(`SUM(%s:%s)`, cellName(l.colEUR, l.rowAufschlStart), cellName(l.colEUR, row-1))
			formulaEUR = fmt.Sprintf(`=IF(%s=0, 0, ROUND(%s - %s, 2))`, currentLC, saldoEUR, prevSum)
		}
		_ = f.SetCellFormula(ws, cellName(l.colEUR, row), formulaEUR)
	}
}

// fbBindDifferenz setzt die Kontrollformeln (Saldo vs. Aufschlüsselung).
func (g *Generator) fbBindDifferenz(ws string, l fbLayout) {
	f := g.file
	r := l.rowDifferenz
	check := func(col int) string {
		return fmt.Sprintf(`=ROUND(IFERROR(%s-SUM(%s:%s),""),2)`,
			cellName(col, l.rowSaldoFB), cellName(col, l.rowAufschlStart), cellName(col, l.rowAufschlEnd))
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, r), check(l.colLC))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, r), check(l.colEUR))
}

// fbBindDetailTables baut die beiden Detail-Einnahmentabellen, setzt die
// Durchschnittskurs-Formel und die SUMIF-Formeln der Einnahmen-Typzeilen.
func (g *Generator) fbBindDetailTables(ws string, reg *TemplateRegistry, l fbLayout) error {
	f := g.file
	saldoVorLC := cellName(l.colLC, l.rowSaldoVortrag)
	saldoVorEUR := cellName(l.colEUR, l.rowSaldoVortrag)
	rateAddr := absName(l.colLC, l.rowKurs)

	// Durchschnittskurs-Named-Range (Wert wird nach Tabelle 1 gesetzt).
	g.dbUpsertNamedRange(ws, fmt.Sprintf(FBNameKursFmt, l.periode), l.colLC, l.rowKurs)

	// Tabelle 1: explizite Kurseingabe
	if err := g.fbCreateEinnahmenTabelle(ws, reg, l, false, saldoVorLC, saldoVorEUR, rateAddr); err != nil {
		return err
	}
	// Tabelle 2: Durchschnittskurs
	if err := g.fbCreateEinnahmenTabelle(ws, reg, l, true, saldoVorLC, saldoVorEUR, rateAddr); err != nil {
		return err
	}

	// Durchschnittskurs (Zeile 8) = Kurs der Summenzeile von Tabelle 1.
	_ = f.SetCellFormula(ws, cellName(l.colLC, l.rowKurs),
		fmt.Sprintf(`=ROUND(IFERROR(%s,0),6)`, cellName(l.colStart+FBDetOffKurs, l.rowDetail1Totals)))

	// SUMIF-Bereiche der beiden Detailtabellen.
	tbl1Typ := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffTyp, l.rowDetail1Hdr+1), absName(l.colStart+FBDetOffTyp, l.rowDetail1Hdr+FBDetailRowsExplizit))
	tbl1LC := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffLC, l.rowDetail1Hdr+1), absName(l.colStart+FBDetOffLC, l.rowDetail1Hdr+FBDetailRowsExplizit))
	tbl1EUR := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffEUR, l.rowDetail1Hdr+1), absName(l.colStart+FBDetOffEUR, l.rowDetail1Hdr+FBDetailRowsExplizit))

	tbl2Typ := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffTyp, l.rowDetail2Hdr+1), absName(l.colStart+FBDetOffTyp, l.rowDetail2Hdr+FBDetailRowsWK))
	tbl2LC := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffLC, l.rowDetail2Hdr+1), absName(l.colStart+FBDetOffLC, l.rowDetail2Hdr+FBDetailRowsWK))
	tbl2EUR := fmt.Sprintf("%s:%s", absName(l.colStart+FBDetOffEUR, l.rowDetail2Hdr+1), absName(l.colStart+FBDetOffEUR, l.rowDetail2Hdr+FBDetailRowsWK))

	for i := 0; i < l.incomeCount; i++ {
		typeRow := l.rowIncomeStart + i
		labelAddr := absName(l.colLabel, typeRow)
		lcFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,%s,%s)+SUMIF(%s,%s,%s),2)`, tbl1Typ, labelAddr, tbl1LC, tbl2Typ, labelAddr, tbl2LC)
		eurFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,%s,%s)+SUMIF(%s,%s,%s),2)`, tbl1Typ, labelAddr, tbl1EUR, tbl2Typ, labelAddr, tbl2EUR)

		_ = f.SetCellFormula(ws, cellName(l.colLC, typeRow), lcFormula)
		_ = f.SetCellFormula(ws, cellName(l.colEUR, typeRow), eurFormula)

		if l.isFollowUp {
			prevKumLC := cellName(l.prevColStart+FBOffKumLC, typeRow)
			prevKumEUR := cellName(l.prevColStart+FBOffKumEUR, typeRow)
			_ = f.SetCellFormula(ws, cellName(l.colKumLC, typeRow),
				fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumLC, cellName(l.colLC, typeRow), cellName(l.colLC, typeRow)))
			_ = f.SetCellFormula(ws, cellName(l.colKumEUR, typeRow),
				fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumEUR, cellName(l.colEUR, typeRow), cellName(l.colEUR, typeRow)))
		} else {
			_ = f.SetCellFormula(ws, cellName(l.colKumLC, typeRow), lcFormula)
			_ = f.SetCellFormula(ws, cellName(l.colKumEUR, typeRow), eurFormula)
		}
	}

	return nil
}

// fbCreateEinnahmenTabelle baut eine der beiden Detail-Einnahmentabellen
// (Layout, Styles, Formeln, Tabelle und Validierungen). isWK schaltet zwischen
// expliziter Kurseingabe und Durchschnittskurs um.
func (g *Generator) fbCreateEinnahmenTabelle(
	ws string,
	reg *TemplateRegistry,
	l fbLayout,
	isWK bool,
	saldoLCAddr, saldoEURAddr, avgRateAddr string,
) error {
	f := g.file
	colStart := l.colStart

	var (
		tblName  string
		startRow int
		dataRows int
	)
	if isWK {
		tblName = reg.TableFBEinnahmenWK.Get(l.periode).Name
		startRow = l.rowDetail2Hdr
		dataRows = FBDetailRowsWK
	} else {
		tblName = reg.TableFBEinnahmen.Get(l.periode).Name
		startRow = l.rowDetail1Hdr
		dataRows = FBDetailRowsExplizit
	}
	totalsRow := startRow + dataRows + 1

	// Kopfzeile
	headers := []string{"Typ", "Geber", "Einnahmen (LC)", "Einnahmen (EUR)", "Kurs"}
	for i, h := range headers {
		_ = g.setValue(ws, cellName(colStart+i, startRow), h, FBDetailHdrStyle)
	}

	// Datenbereich für den VSTACK registrieren.
	dataRange := fmt.Sprintf("'%s'!%s:%s", ws,
		absName(colStart+FBDetOffTyp, startRow+1), absName(colStart+FBDetOffKurs, startRow+dataRows))
	if isWK {
		g.rangesEinnahmen2 = append(g.rangesEinnahmen2, dataRange)
	} else {
		g.rangesEinnahmen1 = append(g.rangesEinnahmen1, dataRange)
	}

	if err := f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(colStart, startRow), cellName(colStart+FBDetOffKurs, startRow+dataRows)),
		Name:           tblName,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	}); err != nil {
		return err
	}

	saldoLabel := "Saldo des Vorprojekts"
	if l.isFollowUp {
		saldoLabel = "Saldo der Vorperiode"
	}

	// Werte & Formeln je Datenzeile
	for i := 0; i < dataRows; i++ {
		row := startRow + 1 + i
		isSaldo := i == 0 && !isWK

		typ, geber := "", ""
		if isSaldo {
			typ = saldoLabel
		}
		_ = f.SetCellValue(ws, cellName(colStart+FBDetOffTyp, row), typ)
		_ = f.SetCellValue(ws, cellName(colStart+FBDetOffGeber, row), geber)

		if isSaldo {
			_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffLC, row), fmt.Sprintf(`=ROUND(%s,2)`, saldoLCAddr))
			_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffEUR, row), fmt.Sprintf(`=ROUND(%s,2)`, saldoEURAddr))
		} else if isWK {
			// EUR = LC / Durchschnittskurs
			_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffEUR, row),
				fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, cellName(colStart+FBDetOffLC, row), avgRateAddr))
		}

		// Kurs = LC / EUR
		_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffKurs, row),
			fmt.Sprintf(`=ROUND(IFERROR(%s/%s,0),6)`, cellName(colStart+FBDetOffLC, row), cellName(colStart+FBDetOffEUR, row)))
	}

	// Styles je Datenzeile
	for i := 0; i < dataRows; i++ {
		row := startRow + 1 + i
		if i == 0 && !isWK {
			_ = g.setStyle(ws, cellName(colStart+FBDetOffTyp, row), cellName(colStart+FBDetOffTyp, row), FBDetailSaldoTypStyle)
			_ = g.setStyle(ws, cellName(colStart+FBDetOffGeber, row), cellName(colStart+FBDetOffGeber, row), FBDetailSaldoGeberStyle)
			_ = g.setStyle(ws, cellName(colStart+FBDetOffLC, row), cellName(colStart+FBDetOffLC, row), FBDetailSaldoLCStyle)
			_ = g.setStyle(ws, cellName(colStart+FBDetOffEUR, row), cellName(colStart+FBDetOffEUR, row), FBDetailSaldoEURStyle)
			_ = g.setStyle(ws, cellName(colStart+FBDetOffKurs, row), cellName(colStart+FBDetOffKurs, row), FBDetailSaldoKursStyle)
			continue
		}
		eurStyle := FBDetailEURInputStyle
		if isWK {
			eurStyle = FBDetailEURCalcStyle
		}
		_ = g.setStyle(ws, cellName(colStart+FBDetOffTyp, row), cellName(colStart+FBDetOffTyp, row), FBDetailTypStyle)
		_ = g.setStyle(ws, cellName(colStart+FBDetOffGeber, row), cellName(colStart+FBDetOffGeber, row), FBDetailGeberStyle)
		_ = g.setStyle(ws, cellName(colStart+FBDetOffLC, row), cellName(colStart+FBDetOffLC, row), FBDetailLCStyle)
		_ = g.setStyle(ws, cellName(colStart+FBDetOffEUR, row), cellName(colStart+FBDetOffEUR, row), eurStyle)
		_ = g.setStyle(ws, cellName(colStart+FBDetOffKurs, row), cellName(colStart+FBDetOffKurs, row), FBDetailKursStyle)
	}

	// Validierungen: Typ (Saldo-Label + feste Typen), Geber (dynamische Liste).
	dvTyp := excelize.NewDataValidation(true)
	dvTyp.Sqref = fmt.Sprintf("%s:%s", cellName(colStart+FBDetOffTyp, startRow+1), cellName(colStart+FBDetOffTyp, startRow+dataRows))
	dvTyp.SetDropList(append([]string{saldoLabel}, TYPE_NAMES...))
	_ = f.AddDataValidation(ws, dvTyp)

	dvGeber := excelize.NewDataValidation(true)
	dvGeber.Sqref = fmt.Sprintf("%s:%s", cellName(colStart+FBDetOffGeber, startRow+1), cellName(colStart+FBDetOffGeber, startRow+dataRows))
	dvGeber.Type = "list"
	dvGeber.Formula1 = "=" + FBNameGeberList
	_ = f.AddDataValidation(ws, dvGeber)

	// Summenzeile
	totalLabel := "Gesamteinnahmen in Periode"
	if isWK {
		totalLabel = "Gesamt (Durchschnittskurs)"
	}
	_ = g.setValue(ws, cellName(colStart+FBDetOffTyp, totalsRow), totalLabel, FBDetailTotalLabelStyle)
	_ = g.setValue(ws, cellName(colStart+FBDetOffGeber, totalsRow), "Durchschnittskurs:", FBDetailTotalGeberStyle)

	_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffLC, totalsRow),
		fmt.Sprintf("=ROUND(SUBTOTAL(109,%s:%s),2)", absName(colStart+FBDetOffLC, startRow+1), absName(colStart+FBDetOffLC, startRow+dataRows)))
	_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffEUR, totalsRow),
		fmt.Sprintf("=ROUND(SUBTOTAL(109,%s:%s),2)", absName(colStart+FBDetOffEUR, startRow+1), absName(colStart+FBDetOffEUR, startRow+dataRows)))
	_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffKurs, totalsRow),
		fmt.Sprintf("=ROUND(IFERROR(%s/%s,0),6)", cellName(colStart+FBDetOffLC, totalsRow), cellName(colStart+FBDetOffEUR, totalsRow)))

	_ = g.setStyle(ws, cellName(colStart+FBDetOffLC, totalsRow), cellName(colStart+FBDetOffLC, totalsRow), FBDetailTotalLCStyle)
	_ = g.setStyle(ws, cellName(colStart+FBDetOffEUR, totalsRow), cellName(colStart+FBDetOffEUR, totalsRow), FBDetailTotalEURStyle)
	_ = g.setStyle(ws, cellName(colStart+FBDetOffKurs, totalsRow), cellName(colStart+FBDetOffKurs, totalsRow), FBDetailTotalKursStyle)

	return nil
}
