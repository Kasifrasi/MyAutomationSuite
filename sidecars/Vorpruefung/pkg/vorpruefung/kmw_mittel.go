package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

const (
	// Spalten
	KMWColPeriode  = 2  // B
	KMWColWaehrung = 3  // C
	KMWColBetrag   = 4  // D
	KMWColDatum    = 5  // E
	KMWColValList  = 26 // Z  (ausgeblendete Hilfsliste für die Periode-Auswahl)

	// Zeilen
	KMWRowTitle     = 2
	KMWRowHeader    = 4
	KMWRowDataStart = 5
	KMWRowDataEnd   = 76
	KMWRowTotal     = 77

	// Ab dieser Zeile werden die Datenzeilen gruppiert und ausgeblendet zugeklappt
	KMWRowCollapseStart = 23

	// Anzahl der wählbaren Perioden in der Hilfsliste (Spalte Z)
	KMWPeriodenAnzahl = 36

	// Spaltenbreiten
	KMWWPeriode  = 15.0
	KMWWWaehrung = 10.0
	KMWWBetrag   = 20.0
	KMWWDatum    = 15.0

	// Sheet
	KMWSheetName = constants.VPSheetKMW_MITTEL
	KMWTabColor  = "FFFF00" // Gelb
)

// KMW_TABLE_NAME wird auch von anderen Sheets referenziert (z. B. pruefung_fb.go)
// und stammt – Registry First – aus der zentralen TemplateRegistry.
var KMW_TABLE_NAME = Registry.TableKMWMittel.Name

// ─── Teil B: Layout-Dokumentation ────────────────────────────────────────────
/*
  LAYOUT KMW-MITTEL:
  | Zeile | B (Periode)        | C (Waehrung)      | D (Betrag)          | E (Datum)          |
  |-------|--------------------|-------------------|---------------------|--------------------|
  |   2   | II. KMW-MITTEL BEREITGESTELLT (Titel, merged B:E)                                    |
  |   4   | Periode            | Waehrung          | Betrag              | Datum   (TblKMWMittel-Kopf) |
  | 5..76 | [Dropdown Periode] | [Dropdown Waehr.] | [Inp Betrag]        | [Inp Datum]        |
  |  77   | GESAMT             |                   | [Σ SUBTOTAL Betrag] |                    |

  Zeilen 23–76 sind gruppiert und standardmäßig eingeklappt (ausgeblendet).
  Spalte Z (ausgeblendet): Hilfsliste "Periode 1".."Periode 36" für das Periode-Dropdown.
*/

// ─── Teil C: Orchestrator ─────────────────────────────────────────────────────

// CreateKMWMittelSheet erstellt das Layout des Blatts "II. KMW-Mittel".
func (g *Generator) CreateKMWMittelSheet() error {
	ws := KMWSheetName

	_, err := g.file.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des KMW-Mittel-Blatts: %w", err)
	}
	tabColor := KMWTabColor
	_ = g.file.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = g.file.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	g.kmwSetupColumns(ws)

	// ── Teil D: Draw ──────────────────────────────────────────────────────────
	if err := g.drawKMWTitle(ws); err != nil {
		return err
	}
	if err := g.drawKMWTable(ws, Registry); err != nil {
		return err
	}
	if err := g.drawKMWTotals(ws); err != nil {
		return err
	}

	// ── Teil E: Bind ──────────────────────────────────────────────────────────
	if err := g.bindKMWTable(ws, Registry); err != nil {
		return err
	}

	// ── Abschluß (Außenrahmen) ────────────────────────────────────────────────
	_ = g.styleOuterBorder(ws, KMWRowHeader, KMWColPeriode, KMWRowHeader, KMWColDatum, 2, KMWClrBorder)
	_ = g.styleOuterBorder(ws, KMWRowDataStart, KMWColPeriode, KMWRowDataEnd, KMWColDatum, 2, KMWClrBorder)
	_ = g.styleOuterBorder(ws, KMWRowTotal, KMWColPeriode, KMWRowTotal, KMWColDatum, 2, KMWClrBorder)

	return nil
}

// ─── Teil D: Draw-Funktionen (nur visuell) ───────────────────────────────────

func (g *Generator) kmwSetupColumns(ws string) {
	g.setColWidth(ws, KMWColPeriode, KMWWPeriode)
	g.setColWidth(ws, KMWColWaehrung, KMWWWaehrung)
	g.setColWidth(ws, KMWColBetrag, KMWWBetrag)
	g.setColWidth(ws, KMWColDatum, KMWWDatum)
}

func (g *Generator) drawKMWTitle(ws string) error {
	if err := g.mergeCells(ws,
		cellName(KMWColPeriode, KMWRowTitle),
		cellName(KMWColDatum, KMWRowTitle),
		"II. KMW-MITTEL BEREITGESTELLT",
		KMWTitleStyle,
	); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Titels: %w", err)
	}
	_ = g.file.SetRowHeight(ws, KMWRowTitle, 24.0)
	return nil
}

func (g *Generator) drawKMWTable(ws string, reg *TemplateRegistry) error {
	f := g.file

	// Kopfzeile (Spaltenüberschriften aus der Registry)
	for i, col := range reg.TableKMWMittel.Columns {
		_ = g.setValue(ws, cellName(KMWColPeriode+i, KMWRowHeader), col.Header, KMWHeaderStyle)
	}
	_ = f.SetRowHeight(ws, KMWRowHeader, 20.0)

	// Datenzeilen formatieren
	for row := KMWRowDataStart; row <= KMWRowDataEnd; row++ {
		_ = g.setStyle(ws, cellName(KMWColPeriode, row), cellName(KMWColPeriode, row), KMWInputStyle)
		_ = g.setStyle(ws, cellName(KMWColWaehrung, row), cellName(KMWColWaehrung, row), KMWInputStyle)
		_ = g.setStyle(ws, cellName(KMWColBetrag, row), cellName(KMWColBetrag, row), KMWBetragStyle)
		_ = g.setStyle(ws, cellName(KMWColDatum, row), cellName(KMWColDatum, row), KMWDatumStyle)
	}

	// Zeilen 23 bis 76 gruppieren und ausgeblendet zuklappen
	for r := KMWRowCollapseStart; r <= KMWRowDataEnd; r++ {
		_ = f.SetRowOutlineLevel(ws, r, 1)
		_ = f.SetRowVisible(ws, r, false)
	}

	return nil
}

func (g *Generator) drawKMWTotals(ws string) error {
	row := KMWRowTotal
	_ = g.setValue(ws, cellName(KMWColPeriode, row), "GESAMT", KMWTotalStyle)
	_ = g.setStyle(ws, cellName(KMWColWaehrung, row), cellName(KMWColWaehrung, row), KMWTotalStyle)
	_ = g.setFormula(ws, cellName(KMWColBetrag, row),
		fmt.Sprintf("=SUBTOTAL(109,%s[Betrag])", KMW_TABLE_NAME), KMWTotalBetragStyle)
	_ = g.setStyle(ws, cellName(KMWColDatum, row), cellName(KMWColDatum, row), KMWTotalStyle)
	return nil
}

// ─── Teil E: Bind-Funktionen (Logik & Registry) ───────────────────────────────

func (g *Generator) bindKMWTable(ws string, reg *TemplateRegistry) error {
	f := g.file
	tbl := reg.TableKMWMittel

	// Excel-Tabelle (Kopf in Zeile 4 + Datenzeilen). Die GESAMT-Zeile liegt bewusst
	// AUSSERHALB des Table-Range: excelize kann keine echte Totals-Row erzeugen
	// (totalsRowCount wird nie gesetzt, jede Zeile im Range gilt als Datenzeile).
	// Als Zeile direkt unter der Tabelle referenziert SUBTOTAL(109,…) nur die
	// Datenzeilen und schließt sich selbst aus – kein Zirkelbezug, keine Doppelzählung.
	if err := f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(KMWColPeriode, KMWRowHeader), cellName(KMWColDatum, KMWRowDataEnd)),
		Name:           tbl.Name,
		StyleName:      "TableStyleLight1",
		ShowRowStripes: falsePtr(),
	}); err != nil {
		return fmt.Errorf("fehler beim Erstellen der Tabelle %s: %w", tbl.Name, err)
	}

	// Ausgeblendete Hilfsliste "Periode 1..36" in Spalte Z für das Periode-Dropdown
	for i := 1; i <= KMWPeriodenAnzahl; i++ {
		if err := f.SetCellValue(ws, cellName(KMWColValList, i), fmt.Sprintf("Periode %d", i)); err != nil {
			return fmt.Errorf("fehler beim Schreiben der Periode in Spalte Z: %w", err)
		}
	}
	_ = f.SetColVisible(ws, colLetter(KMWColValList), false)

	// Validierung 'Periode' – dynamische Quelle (Hilfsliste) aus der Registry-Spalte
	periodeSqref := fmt.Sprintf("%s:%s", cellName(KMWColPeriode, KMWRowDataStart), cellName(KMWColPeriode, KMWRowDataEnd))
	if err := g.applyColumnDynamicValidation(ws, periodeSqref, tbl.Columns[KMWColPeriode-KMWColPeriode]); err != nil {
		return fmt.Errorf("fehler beim Hinzufügen der Periode-Validierung: %w", err)
	}

	// Validierung 'Waehrung' – statische Liste aus der Registry-Spalte
	waehrungSqref := fmt.Sprintf("%s:%s", cellName(KMWColWaehrung, KMWRowDataStart), cellName(KMWColWaehrung, KMWRowDataEnd))
	if err := g.applyColumnValidation(ws, waehrungSqref, tbl.Columns[KMWColWaehrung-KMWColPeriode]); err != nil {
		return fmt.Errorf("fehler beim Hinzufügen der Waehrung-Validierung: %w", err)
	}

	return nil
}
