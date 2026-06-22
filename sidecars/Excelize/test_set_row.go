package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("../../2026_0004_001_português_Test_FB_15.xlsx")
	if err != nil {
		panic(err)
	}
	sheet := f.GetSheetName(0)
    
    fmt.Printf("Before:\n")
    for _, col := range []string{"L", "M", "N", "O"} {
        cell := fmt.Sprintf("%s14", col)
        form, _ := f.GetCellFormula(sheet, cell)
        val, _ := f.GetCellValue(sheet, cell)
        fmt.Printf("%s form: %q, val: %q\n", cell, form, val)
    }

    err = f.SetSheetRow(sheet, "L14", &[]interface{}{"date", 10, 20, 30})
    
    fmt.Printf("\nAfter SetSheetRow:\n")
    for _, col := range []string{"L", "M", "N", "O"} {
        cell := fmt.Sprintf("%s14", col)
        form, _ := f.GetCellFormula(sheet, cell)
        val, _ := f.GetCellValue(sheet, cell)
        fmt.Printf("%s form: %q, val: %q\n", cell, form, val)
    }
}
