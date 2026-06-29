package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func main() {
	f := excelize.NewFile()
	dv := excelize.NewDataValidation(true)
	dv.Sqref = "A1"
	dv.SetDropList([]string{"Ja", "Nein"})
	f.AddDataValidation("Sheet1", dv)
	
	dvs, _ := f.GetDataValidations("Sheet1")
	for _, d := range dvs {
		fmt.Printf("Sqref: %s, Formula1: %s\n", d.Sqref, d.Formula1)
	}
}
