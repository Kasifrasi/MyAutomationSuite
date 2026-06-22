package report

import (
	"fmt"
)

// CleanUpNumberingFormulas sucht in den Kategorien 1 bis 5 nach festgeschriebenen
// Texten in Spalte B (z.B. "1.1", "1.2") und ersetzt diese durch die dynamische
// Excel-Formel, falls sie durch harten Text überschrieben wurden.
// So stellen wir sicher, dass bei der späteren Duplizierung Formeln kopiert werden.
func (r *ExcelReport) CleanUpNumberingFormulas() error {
	f := r.file
	s := r.sheet

	// Wir verarbeiten nur Kategorien 1 bis 5, da 6, 7 und 8 keine Unterpositionen haben
	for catID := 1; catID <= 5; catID++ {
		startRow, okStart := r.CatStartRows[catID]
		endRow, okEnd := r.CatEndRows[catID]
		if !okStart || !okEnd {
			continue // Start oder Ende der Kategorie nicht gefunden
		}

		// Falls endRow <= startRow ist, befindet sich die Kategorie ggf. bereits im "Modus 0"
		if endRow <= startRow {
			continue
		}

		// Alle Zeilen zwischen startRow und endRow sind Kostenpositionen (inkl. Header).
		// (StartRow = Header, EndRow = Zwischensumme)
		for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
			cellName := fmt.Sprintf("B%d", rowIdx)

			// Wir prüfen, ob eine Formel vorhanden ist.
			formula, _ := f.GetCellFormula(s, cellName)
			if formula == "" {
				// Es gibt keine Formel (also harter Text oder leere Zelle).
				// Wir überschreiben es mit der dynamischen Excel-Formel.
				formulaB := fmt.Sprintf(`=IF(ROW()<ROW($B$%d),"",IF(ROW()=ROW($B$%d),"%d.","%d."&(ROW()-ROW($B$%d))))`, startRow, startRow, catID, catID, startRow)

				// Zellenwert bereinigen (SetCellDefault verhindert den Excelize-Bug mit veralteten Shared-String-Indizes)
				f.SetCellDefault(s, cellName, "")
				setGeneralFormat(f, s, cellName)
				f.SetCellFormula(s, cellName, formulaB)
			}
		}
	}
	return nil
}
