package main

import (
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Sheet1")
	f.SetCellValue("Sheet1", "A1", 1)
	f.SetCellValue("Sheet1", "A2", 2)
	f.SetCellValue("Sheet1", "B1", 3)
	f.SetCellValue("Sheet1", "B2", 4)
    
	t := "array"
	ref := "C1"
	f.SetCellFormula("Sheet1", "C1", "_xlfn.VSTACK(A1:A2, B1:B2)", excelize.FormulaOpts{Type: &t, Ref: &ref})
	f.SetCellFormula("Sheet1", "D1", "_xlfn.VSTACK(A1:A2, B1:B2)")
	
	f.SaveAs("test_array.xlsx")
}
