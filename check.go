package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("tmp/3_full.xlsx")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	fmt.Println("Dashboard C5 (Projektnummer):", getVal(f, "I. Dashboard", "C5"))
	fmt.Println("Budget B24 (Kategorie 1):", getVal(f, "II. Budget", "B24"))
	fmt.Println("FB B12 (Einnahme 1):", getVal(f, "IV. Finanzberichte", "B12"))
}

func getVal(f *excelize.File, sheet, cell string) string {
	v, _ := f.GetCellValue(sheet, cell)
	return v
}
