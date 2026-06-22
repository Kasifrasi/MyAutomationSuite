package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("../../2026_0004_001_português_Test_FB_24.xlsx")
	if err != nil {
		panic(err)
	}
	
	// Force loading the CalcChain
	f.CalcCellValue("Sheet1", "A1") 
    // This probably initializes CalcChain internally if it wasn't.
    // Actually, there's no need. Let's see what happens if we just set it.
    
    // In Go, since we don't have access to xlsxCalcChainC, we can't set it directly?
	// Wait, xlsxCalcChainC is not exported either.
}
