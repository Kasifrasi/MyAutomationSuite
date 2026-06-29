package api

import (
	"strings"

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
	if err := fillValidationDefaults(f, sheetFBPruefung); err != nil {
		return err
	}

	// 4. MA Prüfung
	sheetMAPruefung := constants.VPSheetMA_PRUEFUNG
	_ = setValSafe(f, sheetMAPruefung, CellMAPruefungAuswahl, "Neueste MA")
	if err := fillValidationDefaults(f, sheetMAPruefung); err != nil {
		return err
	}

	return nil
}

// setValSafe ignoriert Fehler, wenn ein Sheet nicht existiert.
func setValSafe(f *excelize.File, sheet, cell, val string) error {
	if _, err := f.GetSheetIndex(sheet); err != nil {
		return nil // Sheet existiert nicht, ignorieren
	}
	return f.SetCellValue(sheet, cell, val)
}

// fillValidationDefaults durchsucht das Blatt nach speziellen DataValidations
// und füllt diese mit Standardwerten ab.
func fillValidationDefaults(f *excelize.File, sheet string) error {
	if _, err := f.GetSheetIndex(sheet); err != nil {
		return nil // Sheet existiert nicht, ignorieren
	}

	dvs, err := f.GetDataValidations(sheet)
	if err != nil {
		return nil // Wenn keine Validierungen vorhanden sind, nichts tun
	}

	monthLimitIdx := 0
	monthDefaults := []int{8, 0, 0}

	for _, dv := range dvs {
		sqref := dv.Sqref
		cells := strings.Split(sqref, " ") // manchmal durch Leerzeichen getrennt? Normal ist ":"
		if strings.Contains(sqref, ":") {
			cells = strings.Split(sqref, ":")
		}

		for _, cell := range cells {
			// Abzug-Toggles
			if dv.Type == "list" && strings.Contains(dv.Formula1, "Abzug") {
				_ = f.SetCellValue(sheet, cell, "Abzug")
			}
			// Limit-Monate
			if dv.Type == "whole" && dv.Formula1 == "0" && dv.Formula2 == "12" {
				val := 0
				if monthLimitIdx < len(monthDefaults) {
					val = monthDefaults[monthLimitIdx]
				}
				_ = f.SetCellValue(sheet, cell, val)
				monthLimitIdx++
			}
		}
	}
	return nil
}
