package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"shared/constants"
)

func main() {
	f, err := excelize.OpenFile("../../tmp/3_full.xlsx")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	for r := 18; r <= 28; r++ {
		cell := fmt.Sprintf("B%d", r)
		val, _ := f.GetCellValue(constants.VPSheetFINANZBERICHTE, cell)
		fmt.Printf("FB %s: %s\n", cell, val)
	}
}
