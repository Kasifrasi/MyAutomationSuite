package excel

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

// UnlockSheetsArea sets the locked flag to false for the first 1000 columns (A to ALL)
// for the provided list of sheet names.
func UnlockSheetsArea(f *excelize.File, sheets []string) error {
	unlockedStyle, err := f.NewStyle(&excelize.Style{
		Protection: &excelize.Protection{
			Locked: false,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create unlocked style: %w", err)
	}

	colALL, err := excelize.ColumnNumberToName(1000)
	if err != nil {
		return fmt.Errorf("failed to get column name for 1000: %w", err)
	}
	
	colRange := fmt.Sprintf("A:%s", colALL)

	for _, sheet := range sheets {
		err = f.SetColStyle(sheet, colRange, unlockedStyle)
		if err != nil {
			return fmt.Errorf("failed to set col style for sheet %s: %w", sheet, err)
		}
	}

	return nil
}
