package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

const (
	// Sheet
	MA_SHEET_NAME = constants.VPSheetMA
	MA_TAB_COLOR  = "FFFF00" // Gelb

	// Spalten-Geometrie: Jede Perioden-Tabelle belegt 3 Inhaltsspalten
	// (Label | LC | EUR) plus 1 Trennspalte. colS ist die linke Spalte einer
	// Tabelle; der Abstand zur nächsten Tabelle ist MA_COL_STRIDE.
	MA_START_COL   = 2 // B – linke Spalte der ersten Perioden-Tabelle
	MA_TABLE_COLS  = 3 // Label | LC | EUR
	MA_TABLE_SPACE = 1 // Trennspalte (Pfeil)
	MA_COL_STRIDE  = MA_TABLE_COLS + MA_TABLE_SPACE

	// Zeilen-Geometrie: Block 1 beginnt in Zeile MA_START_ROW. Für die
	// Zusatz-Anforderungen (Ausnahme 1/2) wird der komplette Block um jeweils
	// MA_BLOCK_STRIDE Zeilen nach unten versetzt. Dieser Versatz ist Vertrag mit
	// pruefung_ma.go / pruefung_ma_panel.go (dortiges (level-1)*30).
	MA_START_ROW    = 5
	MA_BLOCK_STRIDE = 30

	// Anzahl Perioden und Anforderungs-Slots (MA #1/#2/#3).
	MA_PERIOD_COUNT = 18
	MA_SLOT_COUNT   = EV_MA_SLOTS
	MA_TABLE_COUNT  = MA_PERIOD_COUNT * MA_SLOT_COUNT

	// Zeilen-Offsets relativ zum Block-Start startR. Der Perioden-Kopf sitzt eine
	// Zeile über startR. Die weiteren Zeilen (Summe/Abzüge/Anforderung) werden aus
	// der Anzahl Kostenkategorien in maComputeLayout abgeleitet.
	MA_OFF_PERIODE    = -1
	MA_OFF_VON        = 0
	MA_OFF_BIS        = 1
	MA_OFF_ZEITRAUM   = 2
	MA_OFF_KURS       = 3
	MA_OFF_TABLE_HDR  = 4
	MA_OFF_DATA_START = 5

	// Spaltenbreiten
	MA_W_LABEL = 36.80
	MA_W_LC    = 24.71
	MA_W_EUR   = 24.71
)

// MA_CATEGORIES sind die Kostenkategorie-Zeilen jeder Anforderungstabelle.
var MA_CATEGORIES = ListKostenkategorien

// ─── Teil B: Layout-Dokumentation ────────────────────────────────────────────
/*
  LAYOUT MITTELANFORDERUNG (je Perioden-Tabelle, colS = linke Spalte):
  | Zeile (Block 1) | colS (Label)            | colS+1 (LC)          | colS+2 (EUR)         |
  |-----------------|-------------------------|----------------------|----------------------|
  |        4        | Periode:                | [Periode N (merged)]                        |
  |        5        | Von:                    | [Inp Von (merged)]                          |
  |        6        | Bis:                    | [Inp Bis (merged)]                          |
  |        7        | Zeitraum:               | [Σ Monate (merged)]                         |
  |        8        | OANDA-Kurs:             | [Inp Kurs (merged)]                         |
  |        9        | Kostenkategorie         | Angefordert (LC)     | Angefordert (EUR)    |
  |    10..17       | <Kategorie>             | [Inp Kat LC]         | [= LC/Kurs]          |
  |       18        | SUMME                   | [Σ LC]               | [Σ EUR]              |
  |       20        | Gesamtbedarf an Mitteln:| [= SUMME]            | [= SUMME]            |
  |       21        | abzueglich Eigenmittel: | [Inp LC]             | [= LC/Kurs]          |
  |       22        | abzueglich Drittmittel: | [Inp LC]             | [= LC/Kurs]          |
  |       23        | abzueglich Saldo …:     | [= FB/Saldovortrag]  | [= LC/Kurs]          |
  |       25        | Anforderung:            | [= Bedarf-Abzüge]    | [= Bedarf-Abzüge]    |
  |       27        | Manueller Betrag (EUR): | (merged Label)       | [Inp EUR]            |

  Blöcke: Block 1 (Standard) ab Zeile 5, Block 2/3 (Ausnahme 1/2) um je +30
  Zeilen versetzt. Perioden 2..18 sind spaltenweise gruppiert/eingeklappt, die
  Blöcke 2/3 zeilenweise gruppiert/eingeklappt.
*/

// maLayout hält alle absoluten Zell-Koordinaten einer einzelnen
// Anforderungstabelle. Draw- und Bind-Funktionen teilen sich diese Auflösung.
type maLayout struct {
	periode int // 1..MA_PERIOD_COUNT
	level   int // 1..MA_SLOT_COUNT (Standard=1, Ausnahme 1=2, Ausnahme 2=3)
	tableID int // fortlaufende Tabellen-Nummer 1..MA_TABLE_COUNT

	colLbl int
	colLC  int
	colEUR int

	rowPeriode   int
	rowVon       int
	rowBis       int
	rowZeitraum  int
	rowKurs      int
	rowTableHdr  int
	rowDataStart int
	rowDataEnd   int
	rowSum       int
	rowGesamt    int
	rowEigen     int
	rowDritt     int
	rowSaldo     int
	rowAnf       int
	rowManuell   int

	dataRows int
}

// maTableID berechnet die fortlaufende Tabellen-Nummer aus Periode und Block.
// Identisch zur Registry-Formel (calculateTableID), damit die Named Ranges der
// Registry (Von_%d … über tableID) exakt getroffen werden.
func maTableID(periode, level int) int {
	return (periode - 1) + (level-1)*MA_PERIOD_COUNT + 1
}

// maComputeLayout leitet alle absoluten Zeilen/Spalten einer Tabelle ab.
func maComputeLayout(colS, startR, periode, level int) maLayout {
	n := len(MA_CATEGORIES)
	l := maLayout{
		periode: periode,
		level:   level,
		tableID: maTableID(periode, level),

		colLbl: colS,
		colLC:  colS + 1,
		colEUR: colS + 2,

		rowPeriode:   startR + MA_OFF_PERIODE,
		rowVon:       startR + MA_OFF_VON,
		rowBis:       startR + MA_OFF_BIS,
		rowZeitraum:  startR + MA_OFF_ZEITRAUM,
		rowKurs:      startR + MA_OFF_KURS,
		rowTableHdr:  startR + MA_OFF_TABLE_HDR,
		rowDataStart: startR + MA_OFF_DATA_START,

		dataRows: n,
	}
	l.rowDataEnd = l.rowDataStart + n - 1
	l.rowSum = l.rowDataEnd + 1  // Block 1: Zeile 18
	l.rowGesamt = l.rowSum + 2   // Block 1: Zeile 20 (1 Leerzeile darüber)
	l.rowEigen = l.rowGesamt + 1 // Block 1: Zeile 21
	l.rowDritt = l.rowGesamt + 2 // Block 1: Zeile 22
	l.rowSaldo = l.rowGesamt + 3 // Block 1: Zeile 23
	l.rowAnf = l.rowSaldo + 2    // Block 1: Zeile 25 (1 Leerzeile darüber)
	l.rowManuell = l.rowAnf + 2  // Block 1: Zeile 27 (1 Leerzeile darüber)
	return l
}

// ─── Teil C: Orchestrator ─────────────────────────────────────────────────────

// CreateMittelanforderungSheet erstellt das Blatt "IV. MA" mit je 18 Perioden in
// 3 Anforderungs-Blöcken (Standard + 2 Ausnahmen).
func (g *Generator) CreateMittelanforderungSheet(reg *TemplateRegistry) error {
	ws := MA_SHEET_NAME

	if err := g.maInitSheet(ws); err != nil {
		return err
	}
	fbExists := g.maFinanzberichtExists()

	for level := 1; level <= MA_SLOT_COUNT; level++ {
		startR := MA_START_ROW + (level-1)*MA_BLOCK_STRIDE
		if level > 1 {
			g.drawMABlockTitle(ws, startR, level)
		}
		for p := 1; p <= MA_PERIOD_COUNT; p++ {
			colS := MA_START_COL + (p-1)*MA_COL_STRIDE
			l := maComputeLayout(colS, startR, p, level)

			// Spaltenbreiten + Trennpfeil nur einmal (Spalten werden geteilt).
			if level == 1 {
				g.maSetupColumnWidths(ws, colS)
				if p > 1 {
					g.drawMASeparatorArrow(ws, MA_START_ROW-2, colS-1)
				}
			}

			// ── Teil D: Draw ──────────────────────────────────────────────────
			g.drawMATable(ws, l)
			// ── Teil E: Bind ──────────────────────────────────────────────────
			g.bindMATable(ws, reg, l, fbExists)
		}
	}

	g.maCollapse(ws)
	return nil
}

// ─── Teil D: Draw-Funktionen (nur visuell) ───────────────────────────────────

func (g *Generator) maInitSheet(ws string) error {
	if _, err := g.file.NewSheet(ws); err != nil {
		return fmt.Errorf("fehler beim Erstellen des MA-Blatts: %w", err)
	}
	tabColor := MA_TAB_COLOR
	_ = g.file.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = g.file.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})
	return nil
}

func (g *Generator) maFinanzberichtExists() bool {
	idx, _ := g.file.GetSheetIndex(constants.VPSheetFINANZBERICHTE)
	return idx != -1
}

func (g *Generator) maSetupColumnWidths(ws string, colS int) {
	g.setColWidth(ws, colS, MA_W_LABEL)
	g.setColWidth(ws, colS+1, MA_W_LC)
	g.setColWidth(ws, colS+2, MA_W_EUR)
}

func (g *Generator) drawMASeparatorArrow(ws string, row, col int) {
	_ = g.setValue(ws, cellName(col, row), "➤", MAArrowStyle)
}

func (g *Generator) drawMABlockTitle(ws string, startR, level int) {
	row := startR + MA_OFF_PERIODE - 1 // eine Zeile über dem Perioden-Kopf
	_ = g.setValue(ws, cellName(MA_START_COL, row),
		fmt.Sprintf("Zusätzliche Mittelanforderungen (Ausnahme %d)", level-1), MABlockTitleStyle)
	_ = g.setStyle(ws, cellName(MA_START_COL, row), cellName(MA_START_COL+2, row), MABlockTitleStyle)
}

// drawMATable zeichnet Beschriftungen, Rahmen und leere/formatierte Zellen einer
// einzelnen Anforderungstabelle – ohne Formeln, Named Ranges oder Validierungen.
func (g *Generator) drawMATable(ws string, l maLayout) {
	// Kopfbereich (Periode / Von / Bis / Zeitraum / Kurs)
	g.drawMAHeaderRow(ws, l.colLbl, l.colLC, l.colEUR, l.rowPeriode, "Periode:",
		fmt.Sprintf("Periode %d", l.periode), MAPeriodeValueStyle)
	g.drawMAHeaderRow(ws, l.colLbl, l.colLC, l.colEUR, l.rowVon, "Von:", "", MADateInputStyle)
	g.drawMAHeaderRow(ws, l.colLbl, l.colLC, l.colEUR, l.rowBis, "Bis:", "", MADateInputStyle)
	g.drawMAHeaderRow(ws, l.colLbl, l.colLC, l.colEUR, l.rowZeitraum, "Zeitraum:", "", MAZeitraumStyle)
	g.drawMAHeaderRow(ws, l.colLbl, l.colLC, l.colEUR, l.rowKurs, "OANDA-Kurs:", "", MAKursStyle)

	// Kostenkategorie-Tabelle: Kopf
	_ = g.setValue(ws, cellName(l.colLbl, l.rowTableHdr), "Kostenkategorie", MATableHdrStyle)
	_ = g.setValue(ws, cellName(l.colLC, l.rowTableHdr), "Angefordert (LC)", MATableHdrStyle)
	_ = g.setValue(ws, cellName(l.colEUR, l.rowTableHdr), "Angefordert (EUR)", MATableHdrStyle)

	// Kostenkategorie-Tabelle: Datenzeilen
	for i := 0; i < l.dataRows; i++ {
		row := l.rowDataStart + i
		_ = g.setValue(ws, cellName(l.colLbl, row), MA_CATEGORIES[i], MACatCellStyle)
		_ = g.setStyle(ws, cellName(l.colLC, row), cellName(l.colLC, row), MAInputLCStyle)
		_ = g.setStyle(ws, cellName(l.colEUR, row), cellName(l.colEUR, row), MAEURCellStyle)
	}

	// SUMME
	_ = g.setValue(ws, cellName(l.colLbl, l.rowSum), "SUMME", MATotalLabelStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowSum), cellName(l.colLC, l.rowSum), MATotalLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowSum), cellName(l.colEUR, l.rowSum), MATotalEURStyle)

	// Gesamtbedarf an Mitteln
	_ = g.setValue(ws, cellName(l.colLbl, l.rowGesamt), "Gesamtbedarf an Mitteln:", StyleOptions{VAlign: "center"})
	_ = g.setStyle(ws, cellName(l.colLC, l.rowGesamt), cellName(l.colLC, l.rowGesamt), MAGesamtLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowGesamt), cellName(l.colEUR, l.rowGesamt), MAGesamtEURStyle)

	// abzueglich Eigenmittel / Drittmittel
	_ = g.setValue(ws, cellName(l.colLbl, l.rowEigen), "abzueglich Eigenmittel:", StyleOptions{VAlign: "center"})
	_ = g.setStyle(ws, cellName(l.colLC, l.rowEigen), cellName(l.colLC, l.rowEigen), MAAbzugInputStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowEigen), cellName(l.colEUR, l.rowEigen), MAAbzugEURStyle)

	_ = g.setValue(ws, cellName(l.colLbl, l.rowDritt), "abzueglich Drittmittel:", StyleOptions{VAlign: "center"})
	_ = g.setStyle(ws, cellName(l.colLC, l.rowDritt), cellName(l.colLC, l.rowDritt), MAAbzugInputStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowDritt), cellName(l.colEUR, l.rowDritt), MAAbzugEURStyle)

	// abzueglich Saldo (Vorprojekt / Vorperiode) – Wert ist stets berechnet.
	_ = g.setValue(ws, cellName(l.colLbl, l.rowSaldo), maSaldoLabel(l), StyleOptions{VAlign: "center"})
	_ = g.setStyle(ws, cellName(l.colLC, l.rowSaldo), cellName(l.colLC, l.rowSaldo), MAAbzugFormulaStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowSaldo), cellName(l.colEUR, l.rowSaldo), MAAbzugEURStyle)

	// Anforderung (ehemals "KMW-Mittel Anforderung")
	_ = g.setValue(ws, cellName(l.colLbl, l.rowAnf), "Anforderung:", MAAnforderungLabelStyle)
	_ = g.setStyle(ws, cellName(l.colLC, l.rowAnf), cellName(l.colLC, l.rowAnf), MAAnforderungLCStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowAnf), cellName(l.colEUR, l.rowAnf), MAAnforderungEURStyle)

	// Manueller Betrag (EUR)
	_ = g.mergeCells(ws, cellName(l.colLbl, l.rowManuell), cellName(l.colLC, l.rowManuell),
		"Manueller Betrag (EUR):", MAManBetragLabelStyle)
	_ = g.setStyle(ws, cellName(l.colEUR, l.rowManuell), cellName(l.colEUR, l.rowManuell), MAManBetragEURStyle)
}

// drawMAHeaderRow zeichnet eine Kopfzeile (Label + merged Wertzelle colLC:colEUR).
func (g *Generator) drawMAHeaderRow(ws string, colLbl, colLC, colEUR, row int, label string, value interface{}, valStyle StyleOptions) {
	_ = g.setValue(ws, cellName(colLbl, row), label, MALabelStyle)
	_ = g.mergeCells(ws, cellName(colLC, row), cellName(colEUR, row), value, valStyle)
}

// ─── Teil E: Bind-Funktionen (Formeln, Registry, Validierungen) ───────────────

// bindMATable verknüpft die gezeichnete Tabelle mit Formeln und den Named Ranges
// der TemplateRegistry. Alle Named Ranges kommen ausschließlich aus reg.
func (g *Generator) bindMATable(ws string, reg *TemplateRegistry, l maLayout, fbExists bool) {
	bindOut := func(name string, col, row int) {
		g.upsertNamedFormula(name, fmt.Sprintf("'%s'!%s", ws, absName(col, row)))
	}

	kursName := reg.InputMAKurs.Get(l.periode, l.level).NamedRange
	vonName := reg.InputMAVon.Get(l.periode, l.level).NamedRange
	bisName := reg.InputMABis.Get(l.periode, l.level).NamedRange

	// Kopf: Periode/Zeitraum (Outputs), Von/Bis/Kurs (Inputs)
	bindOut(reg.OutputMAPeriode.Get(l.periode, l.level).NamedRange, l.colLC, l.rowPeriode)
	_ = g.bindInputField(ws, l.rowVon, l.colLC, reg.InputMAVon.Get(l.periode, l.level))
	_ = g.bindInputField(ws, l.rowBis, l.colLC, reg.InputMABis.Get(l.periode, l.level))
	g.file.SetCellFormula(ws, cellName(l.colLC, l.rowZeitraum), fmt.Sprintf(
		`=IF(OR(%s="",%s=""),"",DATEDIF(%s,%s,"m")+1)`,
		vonName, bisName, vonName, bisName))
	bindOut(reg.OutputMAZeitraum.Get(l.periode, l.level).NamedRange, l.colLC, l.rowZeitraum)
	_ = g.bindInputField(ws, l.rowKurs, l.colLC, reg.InputMAKurs.Get(l.periode, l.level))

	// Kostenkategorien: LC = Input, EUR = LC/Kurs (Output)
	dataRange := fmt.Sprintf("'%s'!%s:%s", ws,
		absName(l.colLbl, l.rowDataStart), absName(l.colEUR, l.rowDataEnd))
	g.rangesMA = append(g.rangesMA, dataRange)

	for i := 0; i < l.dataRows; i++ {
		row := l.rowDataStart + i
		_ = g.bindInputField(ws, row, l.colLC, reg.InputMAKat.Get(l.periode, l.level, i+1))
		g.file.SetCellFormula(ws, cellName(l.colEUR, row),
			fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, reg.InputMAKat.Get(l.periode, l.level, i+1).NamedRange, kursName))
		bindOut(reg.OutputMAKatEUR.Get(l.periode, l.level, i+1).NamedRange, l.colEUR, row)
	}

	// SUMME (Outputs)
	g.file.SetCellFormula(ws, cellName(l.colLC, l.rowSum),
		fmt.Sprintf(`=ROUND(SUM(%s:%s),2)`, cellName(l.colLC, l.rowDataStart), cellName(l.colLC, l.rowDataEnd)))
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowSum),
		fmt.Sprintf(`=ROUND(SUM(%s:%s),2)`, cellName(l.colEUR, l.rowDataStart), cellName(l.colEUR, l.rowDataEnd)))
	bindOut(reg.OutputMASumLC.Get(l.periode, l.level).NamedRange, l.colLC, l.rowSum)
	bindOut(reg.OutputMASumEUR.Get(l.periode, l.level).NamedRange, l.colEUR, l.rowSum)

	// Gesamtbedarf an Mitteln = SUMME (über die benannten SUMME-Bereiche)
	g.file.SetCellFormula(ws, cellName(l.colLC, l.rowGesamt), fmt.Sprintf(`=ROUND(%s,2)`, reg.OutputMASumLC.Get(l.periode, l.level).NamedRange))
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowGesamt), fmt.Sprintf(`=ROUND(%s,2)`, reg.OutputMASumEUR.Get(l.periode, l.level).NamedRange))

	// abzueglich Eigenmittel: LC = Input, EUR = LC/Kurs (Output)
	_ = g.bindInputField(ws, l.rowEigen, l.colLC, reg.InputMAEigenmittelLC.Get(l.periode, l.level))
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowEigen),
		fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, reg.InputMAEigenmittelLC.Get(l.periode, l.level).NamedRange, kursName))
	bindOut(reg.OutputMAEigenmittelEUR.Get(l.periode, l.level).NamedRange, l.colEUR, l.rowEigen)

	// abzueglich Drittmittel: LC = Input, EUR = LC/Kurs (Output)
	_ = g.bindInputField(ws, l.rowDritt, l.colLC, reg.InputMADrittmittelLC.Get(l.periode, l.level))
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowDritt),
		fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, reg.InputMADrittmittelLC.Get(l.periode, l.level).NamedRange, kursName))
	bindOut(reg.OutputMADrittmittelEUR.Get(l.periode, l.level).NamedRange, l.colEUR, l.rowDritt)

	// abzueglich Saldo: LC berechnet (FB/Saldovortrag), EUR = LC/Kurs (beides Output)
	// Die Formel für LC wird später zentral in applyMASaldoFormulas gesetzt
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowSaldo),
		fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, reg.OutputMASaldoLC.Get(l.periode, l.level).NamedRange, kursName))
	bindOut(reg.OutputMASaldoLC.Get(l.periode, l.level).NamedRange, l.colLC, l.rowSaldo)
	bindOut(reg.OutputMASaldoEUR.Get(l.periode, l.level).NamedRange, l.colEUR, l.rowSaldo)

	// Anforderung = Gesamtbedarf - Eigenmittel - Drittmittel - Saldo.
	// Da der Saldo positiv "als Bestand" zu interpretieren ist, wird er wie Eigen-/Drittmittel abgezogen.
	g.file.SetCellFormula(ws, cellName(l.colLC, l.rowAnf), fmt.Sprintf(
		`=IFERROR(ROUND(%s-%s-%s-%s,2),0)`,
		reg.OutputMASumLC.Get(l.periode, l.level).NamedRange, reg.InputMAEigenmittelLC.Get(l.periode, l.level).NamedRange,
		reg.InputMADrittmittelLC.Get(l.periode, l.level).NamedRange, reg.OutputMASaldoLC.Get(l.periode, l.level).NamedRange))
	_ = g.bindInputField(ws, l.rowAnf, l.colLC, reg.InputMAAnforderungLC.Get(l.periode, l.level))
	g.file.SetCellFormula(ws, cellName(l.colEUR, l.rowAnf), fmt.Sprintf(
		`=IFERROR(ROUND(%s-%s-%s-%s,2),0)`,
		reg.OutputMASumEUR.Get(l.periode, l.level).NamedRange, reg.OutputMAEigenmittelEUR.Get(l.periode, l.level).NamedRange,
		reg.OutputMADrittmittelEUR.Get(l.periode, l.level).NamedRange, reg.OutputMASaldoEUR.Get(l.periode, l.level).NamedRange))
	bindOut(reg.OutputMAAnforderungEUR.Get(l.periode, l.level).NamedRange, l.colEUR, l.rowAnf)

	// Manueller Betrag (EUR) – Input
	_ = g.bindInputField(ws, l.rowManuell, l.colEUR, reg.InputMAManBetragEUR.Get(l.periode, l.level))
}

// maSaldoLabel liefert die (gespiegelte) Beschriftung der Saldo-Abzugszeile.
func maSaldoLabel(l maLayout) string {
	if l.level == 1 {
		if l.tableID == 1 {
			return "abzueglich Saldo Vorprojekt:"
		}
		return "abzueglich Saldo Vorperiode (FB):"
	}
	return "abzueglich Saldo Vorperiode (manuell):"
}

// maSaldoFormula liefert die LC-Formel der Saldo-Abzugszeile. Der Standard-Block
// zieht den Saldovortrag (Dashboard) bzw. den FB-Saldo der Vorperiode; die
// Ausnahme-Blöcke haben keine automatische Quelle (0).
func maSaldoFormula(l maLayout, fbExists bool) string {
	if l.level == 1 && fbExists {
		if l.tableID == 1 {
			saldovortragLWName := Registry.InputDashVPFolgeSaldoLC.NamedRange
			return fmt.Sprintf(`=ROUND(IF(%s="",0,%s),2)`, saldovortragLWName, saldovortragLWName)
		}
		return fmt.Sprintf(`=ROUND(IFERROR(%s,0),2)`, Registry.OutputFBSaldoLC.Get(l.tableID-1).NamedRange)
	}
	return "=0"
}

// maCollapse gruppiert und klappt die zusätzlichen Perioden (Spalten) sowie die
// Ausnahme-Blöcke (Zeilen) ein.
func (g *Generator) maCollapse(ws string) {
	f := g.file

	// Perioden 2..18 spaltenweise gruppieren und ausblenden.
	for p := 2; p <= MA_PERIOD_COUNT; p++ {
		colS := MA_START_COL + (p-1)*MA_COL_STRIDE
		for c := colS; c < colS+MA_TABLE_COLS; c++ {
			_ = f.SetColOutlineLevel(ws, colLetter(c), 1)
			_ = f.SetColVisible(ws, colLetter(c), false)
		}
	}

	// Ausnahme-Blöcke (level 2/3) zeilenweise gruppieren und ausblenden.
	for level := 2; level <= MA_SLOT_COUNT; level++ {
		startR := MA_START_ROW + (level-1)*MA_BLOCK_STRIDE
		l := maComputeLayout(MA_START_COL, startR, 1, level)
		for r := startR + MA_OFF_PERIODE - 1; r <= l.rowManuell+1; r++ {
			_ = f.SetRowOutlineLevel(ws, r, 1)
			_ = f.SetRowVisible(ws, r, false)
		}
	}
}
