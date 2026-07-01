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
	g.fbBindEinnahmen(ws, reg, l)
	if err := g.fbBindAusgaben(ws, reg, l); err != nil {
		return err
	}
	g.fbBindSaldo(ws, reg, l)
	g.fbBindAufschluesselung(ws, reg, l)
	g.fbBindDifferenz(ws, reg, l)
	return g.fbBindDetailTables(ws, reg, l)
}

// fbBindPeriodHeader bindet Von/Bis-Eingaben und die Zeitraum-Formel.
func (g *Generator) fbBindPeriodHeader(ws string, reg *TemplateRegistry, l fbLayout) {
	// Periode (statischer, gemergter Wert "Periode N").
	g.dbUpsertNamedRange(ws, reg.OutputFBPeriode.Get(l.periode).NamedRange, l.colLC, l.rowPeriode)

	_ = g.bindInputField(ws, l.rowVon, l.colLC, reg.InputFBVon.Get(l.periode))
	_ = g.bindInputField(ws, l.rowBis, l.colLC, reg.InputFBBis.Get(l.periode))

	vonCell := reg.InputFBVon.Get(l.periode).NamedRange
	bisCell := reg.InputFBBis.Get(l.periode).NamedRange
	_ = g.file.SetCellFormula(ws, cellName(l.colLC, l.rowZeitraum), fmt.Sprintf(
		`=IF(OR(%s="",%s=""),"",DATEDIF(%s,%s,"m")+1)`, vonCell, bisCell, vonCell, bisCell))
	g.dbUpsertNamedRange(ws, reg.OutputFBZeitraum.Get(l.periode).NamedRange, l.colLC, l.rowZeitraum)
}

// fbBindEinnahmen bindet die Vorperiodensaldo- und die Gesamteinnahmen-Formeln.
// (Die Typ-Zeilen-Formeln benötigen die Detailtabellen und folgen dort.)
func (g *Generator) fbBindEinnahmen(ws string, reg *TemplateRegistry, l fbLayout) {
	f := g.file
	r := l.rowSaldoVortrag

	saldovortragLWName := reg.InputDashVPFolgeSaldoLC.NamedRange
	saldovortragEURName := reg.OutputDashSaldovortragEUR.NamedRange
	saldoVortragLC := fmt.Sprintf(`=ROUND(IF(%s="",0,%s),2)`, saldovortragLWName, saldovortragLWName)
	saldoVortragEUR := fmt.Sprintf(`=ROUND(IF(%s="",0,%s),2)`, saldovortragEURName, saldovortragEURName)

	if l.isFollowUp {
		_ = f.SetCellFormula(ws, cellName(l.colLC, r), fmt.Sprintf(`=ROUND(%s,2)`, reg.OutputFBSaldoLC.Get(l.periode-1).NamedRange))
		_ = f.SetCellFormula(ws, cellName(l.colEUR, r), fmt.Sprintf(`=ROUND(%s,2)`, reg.OutputFBSaldoEUR.Get(l.periode-1).NamedRange))
	} else {
		_ = f.SetCellFormula(ws, cellName(l.colLC, r), saldoVortragLC)
		_ = f.SetCellFormula(ws, cellName(l.colEUR, r), saldoVortragEUR)
	}
	// Kumulierte Spalten der Saldozeile speisen sich immer aus dem Dashboard.
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, r), saldoVortragLC)
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, r), saldoVortragEUR)

	// Vorperiodensaldo-Zeile benennen.
	g.dbUpsertNamedRange(ws, reg.OutputFBVSaldoLC.Get(l.periode).NamedRange, l.colLC, r)
	g.dbUpsertNamedRange(ws, reg.OutputFBVSaldoEUR.Get(l.periode).NamedRange, l.colEUR, r)
	g.dbUpsertNamedRange(ws, reg.OutputFBVSaldoKumLC.Get(l.periode).NamedRange, l.colKumLC, r)
	g.dbUpsertNamedRange(ws, reg.OutputFBVSaldoKumEUR.Get(l.periode).NamedRange, l.colKumEUR, r)

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

	// Gesamteinnahmen-Zeile benennen.
	g.dbUpsertNamedRange(ws, reg.OutputFBGEinnahmenLC.Get(l.periode).NamedRange, l.colLC, gr)
	g.dbUpsertNamedRange(ws, reg.OutputFBGEinnahmenEUR.Get(l.periode).NamedRange, l.colEUR, gr)
	g.dbUpsertNamedRange(ws, reg.OutputFBKumGEinnahmenLC.Get(l.periode).NamedRange, l.colKumLC, gr)
	g.dbUpsertNamedRange(ws, reg.OutputFBKumGEinnahmenEUR.Get(l.periode).NamedRange, l.colKumEUR, gr)
}

// fbBindAusgaben legt die Ausgaben-Tabelle an und setzt ID-/EUR-/Kum-/Summenformeln.
func (g *Generator) fbBindAusgaben(ws string, reg *TemplateRegistry, l fbLayout) error {
	f := g.file
	ausgName := reg.TableFBAusgaben.Get(l.periode).Name
	rateAddr := reg.OutputFBKurs.Get(l.periode).NamedRange

	// Datenbereich für den VSTACK der Datei-Übersicht registrieren.
	dataRange := fmt.Sprintf("'%s'!%s:%s", ws,
		absName(l.colLabel, l.rowAusgTblHdr+1), absName(l.colKumEUR, l.rowAusgTblHdr+l.ausgCount))
	g.rangesAusgaben = append(g.rangesAusgaben, dataRange)

	if err := f.AddTable(ws, &excelize.Table{
		// Summenzeile (Gesamtausgaben, l.rowAusgTotals) liegt AUSSERHALB des
		// Table-Range: excelize erzeugt keine echte Totals-Row, jede Zeile im Range
		// gilt sonst als Datenzeile. Als Zeile direkt unter der Tabelle summiert
		// SUBTOTAL(109, ausgName[…]) nur die Datenzeilen (kein Zirkelbezug).
		Range:          fmt.Sprintf("%s:%s", cellName(l.colLabel, l.rowAusgTblHdr), cellName(l.colKumEUR, l.rowAusgTblHdr+l.ausgCount)),
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

	// Gesamtausgaben (SUBTOTAL über die strukturierten Tabellenspalten).
	subtotal := func(colHeader string) string {
		return fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s[%s]),2)`, ausgName, colHeader)
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, l.rowAusgTotals), subtotal("Ausgaben (LC)"))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, l.rowAusgTotals), subtotal("Ausgaben (EUR)"))
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, l.rowAusgTotals), subtotal("Kum. Ausgaben (LC)"))
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, l.rowAusgTotals), subtotal("Kum. Ausgaben (EUR)"))

	// Ergebniszeile "Gesamtausgaben" als Named Ranges exponieren.
	g.dbUpsertNamedRange(ws, reg.OutputFBAusgGesamtLC.Get(l.periode).NamedRange, l.colLC, l.rowAusgTotals)
	g.dbUpsertNamedRange(ws, reg.OutputFBAusgGesamtEUR.Get(l.periode).NamedRange, l.colEUR, l.rowAusgTotals)
	g.dbUpsertNamedRange(ws, reg.OutputFBAusgGesamtKumLC.Get(l.periode).NamedRange, l.colKumLC, l.rowAusgTotals)
	g.dbUpsertNamedRange(ws, reg.OutputFBAusgGesamtKumEUR.Get(l.periode).NamedRange, l.colKumEUR, l.rowAusgTotals)

	return nil
}

// fbBindSaldo setzt die Saldo-Formeln und benennt die LC/EUR-Zellen.
func (g *Generator) fbBindSaldo(ws string, reg *TemplateRegistry, l fbLayout) {
	f := g.file
	r := l.rowSaldoFB
	ausgName := reg.TableFBAusgaben.Get(l.periode).Name

	// Saldo = Gesamteinnahmen (benannt) − Gesamtausgaben (strukturierte Tabellen-Summe).
	diff := func(einnahmenRef, ausgHeader string) string {
		return fmt.Sprintf(`=ROUND(IFERROR(%s-SUBTOTAL(109,%s[%s]),""),2)`,
			einnahmenRef, ausgName, ausgHeader)
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, r), diff(reg.OutputFBGEinnahmenLC.Get(l.periode).NamedRange, "Ausgaben (LC)"))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, r), diff(reg.OutputFBGEinnahmenEUR.Get(l.periode).NamedRange, "Ausgaben (EUR)"))
	_ = f.SetCellFormula(ws, cellName(l.colKumLC, r), diff(reg.OutputFBKumGEinnahmenLC.Get(l.periode).NamedRange, "Kum. Ausgaben (LC)"))
	_ = f.SetCellFormula(ws, cellName(l.colKumEUR, r), diff(reg.OutputFBKumGEinnahmenEUR.Get(l.periode).NamedRange, "Kum. Ausgaben (EUR)"))

	g.dbUpsertNamedRange(ws, reg.OutputFBSaldoLC.Get(l.periode).NamedRange, l.colLC, r)
	g.dbUpsertNamedRange(ws, reg.OutputFBSaldoEUR.Get(l.periode).NamedRange, l.colEUR, r)
}

// fbBindAufschluesselung bindet die LC-Eingaben und verteilt den EUR-Saldo.
func (g *Generator) fbBindAufschluesselung(ws string, reg *TemplateRegistry, l fbLayout) {
	f := g.file
	saldoEUR := reg.OutputFBSaldoEUR.Get(l.periode).NamedRange
	sumLC := fmt.Sprintf(`ROUND(SUM(%s:%s), 2)`,
		cellName(l.colLC, l.rowAufschlStart), cellName(l.colLC, l.rowAufschlEnd))

	for i, cat := range INFO_CATEGORIES {
		row := l.rowAufschlStart + i
		isLast := i == len(INFO_CATEGORIES)-1

		var field InputField
		var eurOut OutputField
		switch cat {
		case "Bank":
			field = reg.InputFBAufschlBankLC.Get(l.periode)
			eurOut = reg.OutputFBAufschlBankEUR.Get(l.periode)
		case "Kasse":
			field = reg.InputFBAufschlKasseLC.Get(l.periode)
			eurOut = reg.OutputFBAufschlKasseEUR.Get(l.periode)
		default:
			field = reg.InputFBAufschlSonstigesLC.Get(l.periode)
			eurOut = reg.OutputFBAufschlSonstigesEUR.Get(l.periode)
		}
		_ = g.bindInputField(ws, row, l.colLC, field)
		g.dbUpsertNamedRange(ws, eurOut.NamedRange, l.colEUR, row)
		// Aufschlüsselungsbetrag (LC) dieser Zeile über seinen benannten Bereich.
		currentLC := field.NamedRange

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
func (g *Generator) fbBindDifferenz(ws string, reg *TemplateRegistry, l fbLayout) {
	f := g.file
	r := l.rowDifferenz
	// Kontrolle: Saldo (benannt) − Summe der Aufschlüsselung (Mehrzeilenblock).
	check := func(saldoRef string, col int) string {
		return fmt.Sprintf(`=ROUND(IFERROR(%s-SUM(%s:%s),""),2)`,
			saldoRef, cellName(col, l.rowAufschlStart), cellName(col, l.rowAufschlEnd))
	}
	_ = f.SetCellFormula(ws, cellName(l.colLC, r), check(reg.OutputFBSaldoLC.Get(l.periode).NamedRange, l.colLC))
	_ = f.SetCellFormula(ws, cellName(l.colEUR, r), check(reg.OutputFBSaldoEUR.Get(l.periode).NamedRange, l.colEUR))

	g.dbUpsertNamedRange(ws, reg.OutputFBDifferenzLC.Get(l.periode).NamedRange, l.colLC, r)
	g.dbUpsertNamedRange(ws, reg.OutputFBDifferenzEUR.Get(l.periode).NamedRange, l.colEUR, r)
}

// fbBindDetailTables baut die beiden Detail-Einnahmentabellen, setzt die
// Durchschnittskurs-Formel und die SUMIF-Formeln der Einnahmen-Typzeilen.
func (g *Generator) fbBindDetailTables(ws string, reg *TemplateRegistry, l fbLayout) error {
	f := g.file
	saldoVorLC := reg.OutputFBVSaldoLC.Get(l.periode).NamedRange
	saldoVorEUR := reg.OutputFBVSaldoEUR.Get(l.periode).NamedRange
	rateAddr := reg.OutputFBKurs.Get(l.periode).NamedRange

	// Durchschnittskurs-Named-Range (Wert wird nach Tabelle 1 gesetzt).
	g.dbUpsertNamedRange(ws, reg.OutputFBKurs.Get(l.periode).NamedRange, l.colLC, l.rowKurs)

	// Tabelle 1: explizite Kurseingabe
	if err := g.fbCreateEinnahmenTabelle(ws, reg, l, false, saldoVorLC, saldoVorEUR, rateAddr); err != nil {
		return err
	}
	// Tabelle 2: Durchschnittskurs
	if err := g.fbCreateEinnahmenTabelle(ws, reg, l, true, saldoVorLC, saldoVorEUR, rateAddr); err != nil {
		return err
	}

	tbl1Name := reg.TableFBEinnahmen.Get(l.periode).Name
	tbl2Name := reg.TableFBEinnahmenWK.Get(l.periode).Name

	// Durchschnittskurs (Zeile 8) = Gesamt-LC / Gesamt-EUR der Detailtabelle 1
	// (strukturierte Tabellensummen statt Bezug auf die Summenzeile).
	_ = f.SetCellFormula(ws, cellName(l.colLC, l.rowKurs),
		fmt.Sprintf(`=ROUND(IFERROR(SUBTOTAL(109,%s[Einnahmen (LC)])/SUBTOTAL(109,%s[Einnahmen (EUR)]),0),6)`, tbl1Name, tbl1Name))

	// SUMIF-Spalten der beiden Detailtabellen als strukturierte Tabellenbezüge.
	tbl1Typ := fmt.Sprintf("%s[Typ]", tbl1Name)
	tbl1LC := fmt.Sprintf("%s[Einnahmen (LC)]", tbl1Name)
	tbl1EUR := fmt.Sprintf("%s[Einnahmen (EUR)]", tbl1Name)

	tbl2Typ := fmt.Sprintf("%s[Typ]", tbl2Name)
	tbl2LC := fmt.Sprintf("%s[Einnahmen (LC)]", tbl2Name)
	tbl2EUR := fmt.Sprintf("%s[Einnahmen (EUR)]", tbl2Name)

	// Einnahmen-Typzeilen (feste Reihenfolge EM/DM/KMW/Zins, vgl. TYPE_NAMES) an die
	// Registry-Ausgabefelder binden: je Typ LC/EUR und kumuliert LC/EUR.
	incomeFields := [][4]OutputFactory{
		{reg.OutputFBEMlLC, reg.OutputFBEMEUR, reg.OutputFBKumEMLC, reg.OutputFBKumEMEUR},
		{reg.OutputFBDMLC, reg.OutputFBDMEUR, reg.OutputFBKumDMLC, reg.OutputFBKumDMEUR},
		{reg.OutputFBKMWLC, reg.OutputFBKMWEUR, reg.OutputFBKumKMWLC, reg.OutputFBKumKMWEUR},
		{reg.OutputFBZinsLC, reg.OutputFBZinsEUR, reg.OutputFBKumZinsLC, reg.OutputFBKumZinsEUR},
	}

	for i := 0; i < l.incomeCount; i++ {
		typeRow := l.rowIncomeStart + i
		// SUMIF-Kriterium: der feste Einnahmentyp als Literal (statt Bezug auf die
		// Label-Zelle); Fallback auf den Zellbezug, falls kein Typ definiert ist.
		crit := absName(l.colLabel, typeRow)
		if i < len(TYPE_NAMES) {
			crit = `"` + TYPE_NAMES[i] + `"`
		}
		lcFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,%s,%s)+SUMIF(%s,%s,%s),2)`, tbl1Typ, crit, tbl1LC, tbl2Typ, crit, tbl2LC)
		eurFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,%s,%s)+SUMIF(%s,%s,%s),2)`, tbl1Typ, crit, tbl1EUR, tbl2Typ, crit, tbl2EUR)

		_ = f.SetCellFormula(ws, cellName(l.colLC, typeRow), lcFormula)
		_ = f.SetCellFormula(ws, cellName(l.colEUR, typeRow), eurFormula)

		if i < len(incomeFields) {
			g.dbUpsertNamedRange(ws, incomeFields[i][0].Get(l.periode).NamedRange, l.colLC, typeRow)
			g.dbUpsertNamedRange(ws, incomeFields[i][1].Get(l.periode).NamedRange, l.colEUR, typeRow)
			g.dbUpsertNamedRange(ws, incomeFields[i][2].Get(l.periode).NamedRange, l.colKumLC, typeRow)
			g.dbUpsertNamedRange(ws, incomeFields[i][3].Get(l.periode).NamedRange, l.colKumEUR, typeRow)
		}

		if l.isFollowUp {
			// Vorperioden-Kum und aktueller Wert über ihre benannten Bereiche.
			curLC := cellName(l.colLC, typeRow)
			curEUR := cellName(l.colEUR, typeRow)
			prevKumLC := cellName(l.prevColStart+FBOffKumLC, typeRow)
			prevKumEUR := cellName(l.prevColStart+FBOffKumEUR, typeRow)
			if i < len(incomeFields) {
				curLC = incomeFields[i][0].Get(l.periode).NamedRange
				curEUR = incomeFields[i][1].Get(l.periode).NamedRange
				prevKumLC = incomeFields[i][2].Get(l.periode - 1).NamedRange
				prevKumEUR = incomeFields[i][3].Get(l.periode - 1).NamedRange
			}
			_ = f.SetCellFormula(ws, cellName(l.colKumLC, typeRow),
				fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumLC, curLC, curLC))
			_ = f.SetCellFormula(ws, cellName(l.colKumEUR, typeRow),
				fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumEUR, curEUR, curEUR))
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
		// Summenzeile (totalsRow) liegt AUSSERHALB des Table-Range: excelize erzeugt
		// keine echte Totals-Row, jede Zeile im Range gilt sonst als Datenzeile. Als
		// Zeile direkt unter der Tabelle beziehen sich die SUBTOTAL-Formeln nur auf
		// die Datenzeilen (kein Zirkelbezug, keine Doppelzählung).
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
		fmt.Sprintf("=ROUND(SUBTOTAL(109,%s[Einnahmen (LC)]),2)", tblName))
	_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffEUR, totalsRow),
		fmt.Sprintf("=ROUND(SUBTOTAL(109,%s[Einnahmen (EUR)]),2)", tblName))
	_ = f.SetCellFormula(ws, cellName(colStart+FBDetOffKurs, totalsRow),
		fmt.Sprintf("=ROUND(IFERROR(%s/%s,0),6)", cellName(colStart+FBDetOffLC, totalsRow), cellName(colStart+FBDetOffEUR, totalsRow)))

	_ = g.setStyle(ws, cellName(colStart+FBDetOffLC, totalsRow), cellName(colStart+FBDetOffLC, totalsRow), FBDetailTotalLCStyle)
	_ = g.setStyle(ws, cellName(colStart+FBDetOffEUR, totalsRow), cellName(colStart+FBDetOffEUR, totalsRow), FBDetailTotalEURStyle)
	_ = g.setStyle(ws, cellName(colStart+FBDetOffKurs, totalsRow), cellName(colStart+FBDetOffKurs, totalsRow), FBDetailTotalKursStyle)

	// Ergebniszeile (Gesamteinnahmen) als Named Ranges exponieren.
	if isWK {
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnWKGesamtLC.Get(l.periode).NamedRange, colStart+FBDetOffLC, totalsRow)
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnWKGesamtEUR.Get(l.periode).NamedRange, colStart+FBDetOffEUR, totalsRow)
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnWKGesamtKurs.Get(l.periode).NamedRange, colStart+FBDetOffKurs, totalsRow)
	} else {
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnGesamtLC.Get(l.periode).NamedRange, colStart+FBDetOffLC, totalsRow)
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnGesamtEUR.Get(l.periode).NamedRange, colStart+FBDetOffEUR, totalsRow)
		g.dbUpsertNamedRange(ws, reg.OutputFBEinnGesamtKurs.Get(l.periode).NamedRange, colStart+FBDetOffKurs, totalsRow)
	}

	return nil
}
