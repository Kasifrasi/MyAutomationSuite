package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"shared/constants"
	"vorpruefung/pkg/vorpruefung"

	"github.com/xuri/excelize/v2"
)

func main() {
	// 1. Parameter definieren
	sheetsFlag := flag.String("sheets", "", "Kommaseparierte Liste von Sheets (z.B. 'Dashboard,IV. MA'). Leer lassen für alle statischen Ziel-Sheets.")
	flag.Parse()

	// 2. Temporäre Datei frisch generieren
	tmpFile := "temp_template_scan.xlsx"
	fmt.Println("Generiere temporäre Vorlage auf Basis des aktuellen Codes...")

	// leere config generiert ein Standard-Template ohne spezifische Budget-Daten
	if err := vorpruefung.GenerateVorpruefung(tmpFile, vorpruefung.GeneratorConfig{}); err != nil {
		log.Fatalf("Fehler bei der Generierung: %v", err)
	}
	// Räumt die temporäre Datei am Ende auf
	defer os.Remove(tmpFile)

	// 3. Excel öffnen
	f, err := excelize.OpenFile(tmpFile)
	if err != nil {
		log.Fatalf("Fehler beim Öffnen der generierten Datei: %v", err)
	}
	defer f.Close()

	// 4. Zu durchsuchende Sheets ermitteln
	var targetSheets []string
	if *sheetsFlag != "" {
		for _, s := range strings.Split(*sheetsFlag, ",") {
			targetSheets = append(targetSheets, strings.TrimSpace(s))
		}
	} else {
		// Standardmäßig diese Sheets durchsuchen
		targetSheets = []string{
			constants.VPSheetDASHBOARD,
			constants.VPSheetKMW_MITTEL,
			constants.VPSheetMA,
			constants.VPSheetFB_PRUEFUNG,
			constants.VPSheetMA_PRUEFUNG,
		}
	}

	// 5. Zellen analysieren
	fmt.Printf("Suche nach Eingabefeldern (Farbe FFFAE5) in: %v\n\n", targetSheets)

	for _, sheet := range targetSheets {
		fmt.Printf("--- Sheet: %s ---\n", sheet)
		foundAny := false

		// Wir scannen ein Raster (1-50 Spalten, 1-200 Zeilen),
		// damit wir auch komplett leere, aber formatierte Zellen sicher erwischen.
		for r := 1; r <= 200; r++ {
			for c := 1; c <= 50; c++ {
				colName, _ := excelize.ColumnNumberToName(c)
				cellName := fmt.Sprintf("%s%d", colName, r)

				styleIdx, err := f.GetCellStyle(sheet, cellName)
				if err != nil || styleIdx == 0 {
					continue
				}

				style, err := f.GetStyle(styleIdx)
				if err != nil || style == nil || style.Fill.Pattern != 1 {
					continue
				}

				// Hintergrundfarbe auf FFFAE5 prüfen
				if len(style.Fill.Color) > 0 {
					colorStr := strings.ToUpper(strings.TrimPrefix(style.Fill.Color[0], "#"))
					if colorStr == "FFFAE5" {
						fmt.Printf("Gefunden: %s\n", cellName)
						foundAny = true
					}
				}
			}
		}
		if !foundAny {
			fmt.Println("  Keine Eingabefelder gefunden.")
		}
		fmt.Println()
	}
}
