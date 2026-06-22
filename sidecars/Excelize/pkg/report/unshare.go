package report

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// UnshareAllFormulas entfernt das problematische 't="shared"' Attribut
// von allen Formeln im Arbeitsblatt.
// Da Excelize beim Ändern einer Master-Zelle sofort den Cache für die Slave-Zellen löscht,
// muss dies in einem 2-Pass-Verfahren passieren: Erst alle Formeln in den RAM laden,
// dann alle Zellen nacheinander überschreiben.
func (r *ExcelReport) UnshareAllFormulas() error {
	f := r.file
	s := r.sheet

	formulas := make(map[string]string)

	// Pass 1: Alle existierenden Formeln in einem großzügigen Grid auslesen (1000 Zeilen x 100 Spalten).
	// GetRows() schneidet nämlich leere End-Zellen ab und findet sie nicht,
	// daher iterieren wir sicherheitshalber stur über die Koordinaten.
	for rNum := 1; rNum <= 1000; rNum++ {
		for colIdx := 1; colIdx <= 100; colIdx++ {
			colName, err := excelize.ColumnNumberToName(colIdx)
			if err != nil {
				continue
			}
			cellName := fmt.Sprintf("%s%d", colName, rNum)

			form, err := f.GetCellFormula(s, cellName)
			if err == nil && form != "" {
				formulas[cellName] = form
			}
		}
	}

	// Pass 2: Formel löschen und als reguläre (nicht-shared) Formel wieder einfügen
	for cell, form := range formulas {
		f.SetCellFormula(s, cell, "")
		f.SetCellFormula(s, cell, form)
	}

	return nil
}
