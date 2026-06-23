package main

import (
	"fmt"
	"strings"
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

	// 2. VSTACKs der Tabellen

	// Einnahmen 1 (Explizit) - Spalte C
	_ = f.SetCellValue(ws, "C1", "Einnahmen_Explizit_Stack")
	if len(g.rangesEinnahmen1) > 0 {
		_ = g.setDynArrayFormula(ws, "AA2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesEinnahmen1, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "C2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AA2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AA2),3)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AA2),4)<>0), "")`, StyleOptions{})
	}

	// Einnahmen 2 (Durchschnittskurs) - Spalte I
	_ = f.SetCellValue(ws, "I1", "Einnahmen_Durchschnittskurs_Stack")
	if len(g.rangesEinnahmen2) > 0 {
		_ = g.setDynArrayFormula(ws, "AG2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesEinnahmen2, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "I2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AG2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AG2),3)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AG2),4)<>0), "")`, StyleOptions{})
	}

	// Ausgaben (Finanzbericht) - Spalte O
	_ = f.SetCellValue(ws, "O1", "Ausgaben_Finanzbericht_Stack")
	if len(g.rangesAusgaben) > 0 {
		_ = g.setDynArrayFormula(ws, "AM2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesAusgaben, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "O2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AM2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AM2),2)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AM2),3)<>0), "")`, StyleOptions{})
	}

	// Mittelanforderung (MA) - Spalte U
	_ = f.SetCellValue(ws, "U1", "Mittelanforderung_Stack")
	if len(g.rangesMA) > 0 {
		_ = g.setDynArrayFormula(ws, "AS2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesMA, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "U2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AS2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AS2),2)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AS2),3)<>0), "")`, StyleOptions{})
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
