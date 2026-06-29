package main

import (
    "fmt"
    "github.com/xuri/excelize/v2"
    "log"
)

func main() {
    f := excelize.NewFile()
    
    // Simulate setting named range
    err := f.SetDefinedName(&excelize.DefinedName{
        Name:     "Inp_Dash_Projektnummer",
        RefersTo: "'Dashboard'!$C$5",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    err = f.SaveAs("test_dash.xlsx")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Success")
}
