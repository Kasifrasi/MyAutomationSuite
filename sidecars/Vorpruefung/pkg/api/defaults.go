package api

import (
	"shared/constants"
	"vorpruefung/pkg/vorpruefung"

	"github.com/xuri/excelize/v2"
)

// FillDefaults befüllt ein Bare-Metal-Template mit den ursprünglichen Standardwerten.
// Die Einnahmetabellen (in den Finanzberichten) sind hiervon ausgenommen.
func FillDefaults(f *excelize.File) error {
	// 1. Dashboard
	sheetDash := constants.VPSheetDASHBOARD
	_ = setValByNamedRange(f, vorpruefung.FieldDashVorprojekt, vorpruefung.ListJaNein[0]) // Ja
	for row := RowDashChecklistStart; row <= RowDashChecklistEnd; row++ {
		cell, _ := excelize.CoordinatesToCellName(ColDashChecklist, row)
		_ = setValSafe(f, sheetDash, cell, vorpruefung.ListJaNein[1]) // Nein
	}

	// 2. Budget
	_ = setValByNamedRange(f, vorpruefung.FieldBudgetReserveFreigabe, vorpruefung.ListJaNein[1]) // Nein

	// 3. FB Prüfung
	_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAuswahl, "Neuester FB")
	_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
	_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugMehr, vorpruefung.ListAbzug[0])

	// 4. MA Prüfung
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAuswahl, "Neueste MA")
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugMehr, vorpruefung.ListAbzug[0])
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugPrognose, vorpruefung.ListAbzug[0])
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY1, "8")
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY2, "0")
	_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY3, "0")

	return nil
}

// setValSafe ignoriert Fehler, wenn ein Sheet nicht existiert.
func setValSafe(f *excelize.File, sheet, cell, val string) error {
	if _, err := f.GetSheetIndex(sheet); err != nil {
		return nil // Sheet existiert nicht, ignorieren
	}
	return f.SetCellValue(sheet, cell, val)
}
