package main

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

func main() {
	start := time.Now()
	f := excelize.NewFile()
	for r := 1; r <= 1500; r++ {
		f.SetRowVisible("Sheet1", r, true)
	}
	for c := 1; c <= 100; c++ {
		colName, _ := excelize.ColumnNumberToName(c)
		f.SetColVisible("Sheet1", colName, true)
	}
	fmt.Printf("Took %v\n", time.Since(start))
}
