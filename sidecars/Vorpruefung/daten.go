package main

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// CreateDatenSheet erstellt ein verstecktes Blatt "Daten" für Dropdowns und VSTACK-Zusammenfassungen.
func (g *Generator) CreateDatenSheet() error {
	ws := "Daten"
	f := g.file

	_, err := f.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Daten-Blatts: %w", err)
	}

	// Das Blatt verstecken
	_ = f.SetSheetVisible(ws, false)

	// 1. Perioden-Liste für Dropdowns generieren (Spalte A)
	for i := 1; i <= MA_PERIOD_COUNT; i++ {
		_ = f.SetCellValue(ws, fmt.Sprintf("A%d", i), fmt.Sprintf("Periode %d", i))
	}

	// Helper to write array formula
	setArray := func(cell string, formula string) {
		t := "array"
		ref := cell
		_ = f.SetCellFormula(ws, cell, formula, excelize.FormulaOpts{Type: &t, Ref: &ref})
	}

	// 2. VSTACKs der Tabellen

	// Einnahmen 1 (Explizit) - Spalte C
	_ = f.SetCellValue(ws, "C1", "Einnahmen_Explizit_Stack")
	if len(g.rangesEinnahmen1) > 0 {
		formula := fmt.Sprintf(`_xlfn.LET(v, _xlfn.VSTACK(%s), _xlfn.FILTER(v, (INDEX(v,0,3)<>0)+(INDEX(v,0,4)<>0), ""))`, strings.Join(g.rangesEinnahmen1, ","))
		setArray("C2", formula)
	}

	// Einnahmen 2 (Durchschnittskurs) - Spalte I
	_ = f.SetCellValue(ws, "I1", "Einnahmen_Durchschnittskurs_Stack")
	if len(g.rangesEinnahmen2) > 0 {
		formula := fmt.Sprintf(`_xlfn.LET(v, _xlfn.VSTACK(%s), _xlfn.FILTER(v, (INDEX(v,0,3)<>0)+(INDEX(v,0,4)<>0), ""))`, strings.Join(g.rangesEinnahmen2, ","))
		setArray("I2", formula)
	}

	// Ausgaben (Finanzbericht) - Spalte O
	_ = f.SetCellValue(ws, "O1", "Ausgaben_Finanzbericht_Stack")
	if len(g.rangesAusgaben) > 0 {
		formula := fmt.Sprintf(`_xlfn.LET(v, _xlfn.VSTACK(%s), _xlfn.FILTER(v, (INDEX(v,0,2)<>0)+(INDEX(v,0,3)<>0), ""))`, strings.Join(g.rangesAusgaben, ","))
		setArray("O2", formula)
	}

	// Mittelanforderung (MA) - Spalte U
	_ = f.SetCellValue(ws, "U1", "Mittelanforderung_Stack")
	if len(g.rangesMA) > 0 {
		formula := fmt.Sprintf(`_xlfn.LET(v, _xlfn.VSTACK(%s), _xlfn.FILTER(v, (INDEX(v,0,2)<>0)+(INDEX(v,0,3)<>0), ""))`, strings.Join(g.rangesMA, ","))
		setArray("U2", formula)
	}

	// Kopfzeilen formatieren
	headerOpts := StyleOptions{
		Bold:      true,
		FillColor: "D3D3D3", // COLOR_HEADER
	}
	_ = g.setStyle(ws, "C1", "C1", headerOpts)
	_ = g.setStyle(ws, "I1", "I1", headerOpts)
	_ = g.setStyle(ws, "O1", "O1", headerOpts)
	_ = g.setStyle(ws, "U1", "U1", headerOpts)

	return nil
}
