package main

import (
	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("../../2026_0004_001_português_Test_FB_24.xlsx")
	if err != nil {
		panic(err)
	}
	f.CalcChain = nil
	f.SaveAs("../../test_calcchain.xlsx")
}
