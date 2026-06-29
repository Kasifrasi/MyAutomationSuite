package api

import (
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// FillDefaults befüllt ein Bare-Metal-Template mit den ursprünglichen Standardwerten.
// Die Einnahmetabellen (in den Finanzberichten) sind hiervon ausgenommen.
func FillDefaults(f *excelize.File) error {
	// 1. Dashboard
	sheetDash := constants.VPSheetDASHBOARD
	_ = setValSafe(f, sheetDash, CellDashVorprojekt, "Ja")
	for row := RowDashChecklistStart; row <= RowDashChecklistEnd; row++ {
		cell, _ := excelize.CoordinatesToCellName(ColDashChecklist, row)
		_ = setValSafe(f, sheetDash, cell, "Nein")
	}

	// 2. Budget
	sheetBudget := constants.VPSheetBUDGET
	_ = setValSafe(f, sheetBudget, CellBudgetReserveFreigabe, "Nein")

	// 3. FB Prüfung
	sheetFBPruefung := constants.VPSheetFB_PRUEFUNG
	_ = setValSafe(f, sheetFBPruefung, CellFBPruefungAuswahl, "Neuester FB")
	_ = setValSafe(f, sheetFBPruefung, CellFBPruefungAbzugSaldo, "Abzug")
	_ = setValSafe(f, sheetFBPruefung, CellFBPruefungAbzugMehr, "Abzug")

	// 4. MA Prüfung
	sheetMAPruefung := constants.VPSheetMA_PRUEFUNG
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungAuswahl, "Neueste MA")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungAbzugSaldo, "Abzug")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungAbzugMehr, "Abzug")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungAbzugPrognose, "Abzug")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungMonateY1, "8")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungMonateY2, "0")
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungMonateY3, "0")

	return nil
}

// setValSafe ignoriert Fehler, wenn ein Sheet nicht existiert.
func setValSafe(f *excelize.File, sheet, cell, val string) error {
	if _, err := f.GetSheetIndex(sheet); err != nil {
		return nil // Sheet existiert nicht, ignorieren
	}
	return f.SetCellValue(sheet, cell, val)
}
