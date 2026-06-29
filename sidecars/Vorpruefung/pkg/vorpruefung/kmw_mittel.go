package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

const (
	KMW_SHEET_NAME = constants.VPSheetKMW_MITTEL
	KMW_TAB_COLOR  = "FFFF00" // Gelb

	KMW_COL_PERIODE  = 2  // B
	KMW_COL_WAEHRUNG = 3  // C
	KMW_COL_BETRAG   = 4  // D
	KMW_COL_DATUM    = 5  // E
	KMW_COL_VAL_LIST = 26 // Z  (ausgeblendete Hilfsliste für die Periode-Auswahl)

	KMW_TABLE_NAME = "TblKMWMittel"

	// ─── Farb-Konstanten (RGB → Hex) ──────────────────────────────────────────────
	KMW_CLR_HEADER = "D3D3D3" // 211,211,211 – Titel/Kopf/Summe
	KMW_CLR_INPUT  = "FFFAE5" // 255,250,229 – Eingabezeilen
	KMW_CLR_BORDER = "808080" // 128,128,128 – kräftige Rahmen
	KMW_CLR_GRID   = "D3D3D3" // 211,211,211 – dünne Innenrahmen
	KMW_CLR_FONT   = "3C3C3C" // 60,60,60    – Kopf-/Summen-Schrift
)

// CreateKMWMittelSheet erstellt das Layout des Blatts "II. KMW-Mittel".
func (g *Generator) CreateKMWMittelSheet() error {
	ws := KMW_SHEET_NAME
	f := g.file

	// Sheet initialisieren
	_, err := f.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des KMW-Mittel-Blatts: %w", err)
	}

	tabColor := KMW_TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	// Spaltenbreiten einstellen
	g.setColWidth(ws, KMW_COL_PERIODE, 15.0)
	g.setColWidth(ws, KMW_COL_WAEHRUNG, 10.0)
	g.setColWidth(ws, KMW_COL_BETRAG, 20.0)
	g.setColWidth(ws, KMW_COL_DATUM, 15.0)

	// ─── Titel (B2:E2) ────────────────────────────────────────────────────────
	titleOpts := StyleOptions{
		Size:      14.0,
		Bold:      true,
		FillColor: KMW_CLR_HEADER,
		HAlign:    "center",
		VAlign:    "center",
	}
	err = g.mergeCells(ws, cellName(KMW_COL_PERIODE, 2), cellName(KMW_COL_DATUM, 2), "II. KMW-MITTEL BEREITGESTELLT", titleOpts)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Titels: %w", err)
	}
	_ = f.SetRowHeight(ws, 2, 24.0)

	// ─── Validierungs-Hilfsliste "Periode 1..36" in Spalte Z (ausgeblendet) ───
	for i := 1; i <= 36; i++ {
		cell := cellName(KMW_COL_VAL_LIST, i)
		err = f.SetCellValue(ws, cell, fmt.Sprintf("Periode %d", i))
		if err != nil {
			return fmt.Errorf("fehler beim Schreiben der Periode in Spalte Z: %w", err)
		}
	}
	_ = f.SetColVisible(ws, colLetter(KMW_COL_VAL_LIST), false)

	// ─── Tabelle (Kopf in Zeile 4 + 18 Datenzeilen) ────────────────────────────
	headers := []string{"Periode", "Waehrung", "Betrag", "Datum"}
	for i, h := range headers {
		cell := cellName(KMW_COL_PERIODE+i, 4)
		_ = f.SetCellValue(ws, cell, h)
	}

	// Tabelle erstellen
	err = f.AddTable(ws, &excelize.Table{
		Range:          "B4:E76",
		Name:           KMW_TABLE_NAME,
		StyleName:      "TableStyleLight1",
		ShowRowStripes: falsePtr(),
	})
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen der Tabelle %s: %w", KMW_TABLE_NAME, err)
	}

	// ─── Validierungen hinzufügen ─────────────────────────────────────────────
	// Validierung 'Periode' aus der Hilfsliste
	dvPeriode := excelize.NewDataValidation(true)
	dvPeriode.Sqref = "B5:B76"
	dvPeriode.SetSqrefDropList(fmt.Sprintf("'%s'!$%s$1:$%s$36", KMW_SHEET_NAME, colLetter(KMW_COL_VAL_LIST), colLetter(KMW_COL_VAL_LIST)))
	err = f.AddDataValidation(ws, dvPeriode)
	if err != nil {
		return fmt.Errorf("fehler beim Hinzufügen der Periode-Validierung: %w", err)
	}

	// Validierung 'Waehrung'
	dvWaehrung := excelize.NewDataValidation(true)
	dvWaehrung.Sqref = "C5:C76"
	dvWaehrung.SetDropList(ListWaehrung)
	err = f.AddDataValidation(ws, dvWaehrung)
	if err != nil {
		return fmt.Errorf("fehler beim Hinzufügen der Waehrung-Validierung: %w", err)
	}

	// ─── Datenbereich formatieren (B5:E76) ────────────────────────────────────
	for row := 5; row <= 76; row++ {
		// B: Periode
		_ = g.setStyle(ws, cellName(KMW_COL_PERIODE, row), cellName(KMW_COL_PERIODE, row), StyleOptions{
			FillColor:    KMW_CLR_INPUT,
			VAlign:       "center",
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  KMW_CLR_GRID,
		})
		_ = g.bindInputField(ws, row, KMW_COL_PERIODE, FieldKMWPeriode(row-4))

		// C: Waehrung
		_ = g.setStyle(ws, cellName(KMW_COL_WAEHRUNG, row), cellName(KMW_COL_WAEHRUNG, row), StyleOptions{
			FillColor:    KMW_CLR_INPUT,
			VAlign:       "center",
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  KMW_CLR_GRID,
		})
		_ = g.bindInputField(ws, row, KMW_COL_WAEHRUNG, FieldKMWWaehrung(row-4))

		// D: Betrag
		_ = g.setStyle(ws, cellName(KMW_COL_BETRAG, row), cellName(KMW_COL_BETRAG, row), StyleOptions{
			FillColor:    KMW_CLR_INPUT,
			NumFormat:    "#,##0.00",
			HAlign:       "right",
			VAlign:       "center",
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  KMW_CLR_GRID,
		})
		_ = g.bindInputField(ws, row, KMW_COL_BETRAG, FieldKMWBetrag(row-4))

		// E: Datum
		_ = g.setStyle(ws, cellName(KMW_COL_DATUM, row), cellName(KMW_COL_DATUM, row), StyleOptions{
			FillColor:    KMW_CLR_INPUT,
			NumFmtID:     14, // Excel built-in kurzes Datum
			HAlign:       "center",
			VAlign:       "center",
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  KMW_CLR_GRID,
		})
		_ = g.bindInputField(ws, row, KMW_COL_DATUM, FieldKMWDatum(row-4))
	}

	// ─── Kopfzeile formatieren (B4:E4) ────────────────────────────────────────
	headerOpts := StyleOptions{
		Bold:         true,
		Size:         9.0,
		FontColor:    KMW_CLR_FONT,
		FillColor:    KMW_CLR_HEADER,
		HAlign:       "center",
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  KMW_CLR_GRID,
	}
	for c := KMW_COL_PERIODE; c <= KMW_COL_DATUM; c++ {
		_ = g.setStyle(ws, cellName(c, 4), cellName(c, 4), headerOpts)
	}
	_ = f.SetRowHeight(ws, 4, 20.0)

	// ─── Ergebniszeile (GESAMT + Summe Betrag) (B77:E77) ──────────────────────
	totalsRow := 77
	_ = f.SetCellValue(ws, cellName(KMW_COL_PERIODE, totalsRow), "GESAMT")
	_ = f.SetCellFormula(ws, cellName(KMW_COL_BETRAG, totalsRow), fmt.Sprintf("=SUBTOTAL(109,%s[Betrag])", KMW_TABLE_NAME))

	totalsOpts := StyleOptions{
		Bold:         true,
		Size:         9.0,
		FontColor:    KMW_CLR_FONT,
		FillColor:    KMW_CLR_HEADER,
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  KMW_CLR_GRID,
	}
	for c := KMW_COL_PERIODE; c <= KMW_COL_DATUM; c++ {
		_ = g.setStyle(ws, cellName(c, totalsRow), cellName(c, totalsRow), totalsOpts)
	}

	// Spezifisches Zahlenformat für Betrag in der Ergebniszeile
	_ = g.setStyle(ws, cellName(KMW_COL_BETRAG, totalsRow), cellName(KMW_COL_BETRAG, totalsRow), StyleOptions{
		Bold:         true,
		Size:         9.0,
		FontColor:    KMW_CLR_FONT,
		FillColor:    KMW_CLR_HEADER,
		VAlign:       "center",
		NumFormat:    "#,##0.00",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  KMW_CLR_GRID,
	})

	// ─── Außenrahmen (kräftig grau) ───────────────────────────────────────────
	_ = g.styleOuterBorder(ws, 4, KMW_COL_PERIODE, 4, KMW_COL_DATUM, 2, KMW_CLR_BORDER)
	_ = g.styleOuterBorder(ws, 5, KMW_COL_PERIODE, 76, KMW_COL_DATUM, 2, KMW_CLR_BORDER)
	_ = g.styleOuterBorder(ws, totalsRow, KMW_COL_PERIODE, totalsRow, KMW_COL_DATUM, 2, KMW_CLR_BORDER)

	// ─── Zeilen 23 bis 76 gruppieren und ausgeblendet zuklappen ─────────────────
	for r := 23; r <= 76; r++ {
		_ = f.SetRowOutlineLevel(ws, r, 1)
		_ = f.SetRowVisible(ws, r, false)
	}

	return nil
}
