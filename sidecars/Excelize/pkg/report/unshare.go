package report

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func (r *ExcelReport) UnshareAllFormulas() error {
	f := r.file
	s := r.sheet

	rows, err := f.GetRows(s)
	if err != nil {
		return err
	}

	for rowIdx, row := range rows {
		rNum := rowIdx + 1
		for colIdx := range row {
			colName, err := excelize.ColumnNumberToName(colIdx + 1)
			if err != nil {
				continue
			}
			cellName := fmt.Sprintf("%s%d", colName, rNum)

			form, err := f.GetCellFormula(s, cellName)
			if err == nil && form != "" {
				f.SetCellFormula(s, cellName, "")
				f.SetCellFormula(s, cellName, form)
			}
		}
	}
	return nil
}
