package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"excelize-test/pkg/report"
)

// UIProgress sendet Statusmeldungen im JSON-Format an die Standardausgabe,
// die von deiner Slint/Python-GUI leicht geparst werden können.
type UIProgress struct {
	Status  string `json:"status"`            // "start", "progress", "success", "error", "done"
	File    string `json:"file,omitempty"`    // Verarbeitete Datei (falls vorhanden)
	Current int    `json:"current,omitempty"` // Aktuelle Datei (Nummer)
	Total   int    `json:"total,omitempty"`   // Gesamtanzahl der Dateien
	Message string `json:"message,omitempty"` // Optionale Nachricht
}

func printUI(status, file, msg string, current, total int) {
	data, _ := json.Marshal(UIProgress{
		Status:  status,
		File:    file,
		Current: current,
		Total:   total,
		Message: msg,
	})
	fmt.Println(string(data))
}

func main() {
	// 1. Kommandozeilen-Parameter (Flags) definieren
	inputFlag := flag.String("input", "", "Pfad zu einer JSON-Datei mit den verarbeiteten Scanner-Daten (ScannedBudgetData Array)")
	outputFlag := flag.String("output", "test/output", "Ordner, in dem die finalen Dateien gespeichert werden")
	filenameFlag := flag.String("filename", "Report_{pn}_{la}.xlsx", "Namensmuster (z.B. 'Report_{pn}_{la}_{i}.xlsx'). {pn}=Projektnummer, {la}=Sprache, {i}=Zähler")
	optionsFlag := flag.String("options", "", "JSON-String mit den globalen ReportOptions (Blattschutz, etc.)")
	flag.Parse()

	if *inputFlag == "" || !strings.HasSuffix(strings.ToLower(*inputFlag), ".json") {
		fmt.Println("Fehler: Bitte gib eine gültige JSON-Datei mit -input an (z.B. -input daten.json)")
		os.Exit(1)
	}

	outputOrdner := *outputFlag
	// Sicherstellen, dass der Output-Ordner existiert
	if err := os.MkdirAll(outputOrdner, 0755); err != nil {
		printUI("error", "", fmt.Sprintf("Konnte Output-Ordner nicht erstellen: %v", err), 0, 0)
		os.Exit(1)
	}

	printUI("start", "", "App gestartet", 0, 0)

	var globalOptions report.ReportOptions
	if *optionsFlag != "" {
		if err := json.Unmarshal([]byte(*optionsFlag), &globalOptions); err != nil {
			printUI("error", "", fmt.Sprintf("Konnte -options JSON nicht parsen: %v", err), 0, 0)
			os.Exit(1)
		}
	}

	// 2. Templates aus dem RAM laden (via //go:embed) und globale Einstellungen anwenden
	if err := report.PreloadAllTemplates(globalOptions); err != nil {
		printUI("error", "", fmt.Sprintf("Fehler beim Preload: %v", err), 0, 0)
		os.Exit(1)
	}

	// 3. JSON-Daten einlesen und parsen
	bytes, err := os.ReadFile(*inputFlag)
	if err != nil {
		printUI("error", "", fmt.Sprintf("Konnte JSON-Datei nicht lesen: %v", err), 0, 0)
		os.Exit(1)
	}

	var scannedDaten []report.ScannedBudgetData
	if err := json.Unmarshal(bytes, &scannedDaten); err != nil {
		printUI("error", "", fmt.Sprintf("Konnte JSON-Daten nicht parsen (sollte ein Array von ScannedBudgetData sein): %v", err), 0, 0)
		os.Exit(1)
	}

	totalFiles := len(scannedDaten)
	if totalFiles == 0 {
		printUI("done", "", "Keine Daten zum Verarbeiten gefunden (leeres JSON-Array)", 0, 0)
		return
	}

	// 3.5 Pre-Pass: Dateinamen-Duplikate ermitteln (nur für diesen Lauf)
	nameCounts := make(map[string]int)
	baseNames := make([]string, len(scannedDaten))

	for i, scanned := range scannedDaten {
		jobID := scanned.ProjectNumber
		if jobID == "" {
			jobID = fmt.Sprintf("PROJ-%03d", i+1)
		}
		sprache := strings.ToLower(scanned.Language)

		// Basis-Name berechnen (ohne {i}), um echte Namens-Kollisionen zu finden
		base := *filenameFlag
		base = strings.ReplaceAll(base, "{i}", "")
		base = strings.ReplaceAll(base, "{I}", "")
		base = strings.ReplaceAll(base, "{pn}", jobID)
		base = strings.ReplaceAll(base, "{PN}", jobID)
		base = strings.ReplaceAll(base, "{la}", sprache)
		base = strings.ReplaceAll(base, "{LA}", sprache)
		base = strings.ReplaceAll(base, "{version}", scanned.Version)
		base = strings.ReplaceAll(base, "{VERSION}", scanned.Version)

		baseNames[i] = base
		nameCounts[base]++
	}

	// 4. Stream & Pipeline einrichten (dynamisch skalierend anhand der CPU-Kerne)
	jobs := make(chan report.ReportJob, totalFiles)

	// Sweet Spot: 2x bis 3x die Anzahl der CPU-Kerne für gemischte (CPU + I/O) Lasten.
	// Damit verhindern wir, dass auf schwachen Systemen der RAM (durch 50 gleichzeitige Excelize-Instanzen)
	// vollläuft, nutzen aber auf starken Systemen die volle Leistung aus.
	numWorkers := runtime.NumCPU() * 2
	if numWorkers > 32 {
		numWorkers = 32 // Cap, um exzessiven RAM-Verbrauch und I/O-Stau zu vermeiden
	}

	results := report.StartPipeline(numWorkers, jobs)

	// Gescannten Daten direkt in Jobs umwandeln und in die Pipeline schicken
	go func() {
		duplicatesCounter := make(map[string]int)

		for i, scanned := range scannedDaten {
			// Konvertiere die gescannten Daten in ReportData
			daten := report.MapScannedToReportData(&scanned)
			daten.Options = globalOptions
			jobID := scanned.ProjectNumber
			if jobID == "" {
				jobID = fmt.Sprintf("PROJ-%03d", i+1)
			}

			// Setze die globale EmptyRows Konfiguration
			daten.EmptyRows.Global = globalOptions.EmptyRows
			sprache := strings.ToLower(scanned.Language)

			base := baseNames[i]
			duplicatesCounter[base]++

			dateiname := *filenameFlag
			dateiname = strings.ReplaceAll(dateiname, "{pn}", jobID)
			dateiname = strings.ReplaceAll(dateiname, "{PN}", jobID)
			dateiname = strings.ReplaceAll(dateiname, "{la}", sprache)
			dateiname = strings.ReplaceAll(dateiname, "{LA}", sprache)
			dateiname = strings.ReplaceAll(dateiname, "{version}", scanned.Version)
			dateiname = strings.ReplaceAll(dateiname, "{VERSION}", scanned.Version)

			if nameCounts[base] > 1 {
				// Es gibt mehrere Dateien, die (ohne Zähler) gleich heißen würden
				countStr := strconv.Itoa(duplicatesCounter[base])
				if strings.Contains(dateiname, "{i}") {
					dateiname = strings.ReplaceAll(dateiname, "{i}", countStr)
				} else if strings.Contains(dateiname, "{I}") {
					dateiname = strings.ReplaceAll(dateiname, "{I}", countStr)
				} else {
					// Wenn der Nutzer {i} nicht im Muster angegeben hat, fügen wir es vor .xlsx ein
					ext := filepath.Ext(dateiname)
					if ext == "" {
						ext = ".xlsx"
						dateiname += ext
					}
					nameOhneExt := strings.TrimSuffix(dateiname, ext)
					dateiname = fmt.Sprintf("%s_%s%s", nameOhneExt, countStr, ext)
				}
			} else {
				// Die Datei ist einzigartig in diesem Durchlauf
				dateiname = strings.ReplaceAll(dateiname, "{i}", "")
				dateiname = strings.ReplaceAll(dateiname, "{I}", "")
				// Kleine Schönheitskorrektur, falls durch das Entfernen von {i} Reste bleiben
				dateiname = strings.ReplaceAll(dateiname, "__", "_")
				dateiname = strings.ReplaceAll(dateiname, "_.xlsx", ".xlsx")
				dateiname = strings.ReplaceAll(dateiname, "-.xlsx", ".xlsx")
			}

			// Fallback, falls die Endung vergessen wurde
			if filepath.Ext(dateiname) == "" {
				dateiname += ".xlsx"
			}

			outputPath := filepath.Join(outputOrdner, dateiname)
			// uniqueOutputPath als letzter Schutz gegen Überschreiben auf der Festplatte (bei mehrfachen App-Starts)
			outputPath = uniqueOutputPath(outputPath)

			// Job in den Stream schieben
			jobs <- report.ReportJob{
				JobID:      jobID,
				OutputPath: outputPath,
				Data:       daten,
			}
		}
		close(jobs)
	}()

	// 5. Resultate einsammeln
	successCount := 0
	errorCount := 0
	currentDone := 0

	for res := range results {
		currentDone++
		if res.Err != nil {
			printUI("error", res.JobID, fmt.Sprintf("Fehler beim Verarbeiten: %v", res.Err), currentDone, totalFiles)
			errorCount++
		} else {
			printUI("progress", res.JobID, "Bericht generiert", currentDone, totalFiles)
			successCount++
		}
	}

	printUI("done", "", fmt.Sprintf("Fertig! %d erfolgreich, %d fehlerhaft.", successCount, errorCount), totalFiles, totalFiles)
}

// uniqueOutputPath prüft, ob eine Datei existiert, und hängt (1), (2), etc. an den Namen an
func uniqueOutputPath(basePath string) string {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath // Datei existiert noch nicht
	}

	ext := filepath.Ext(basePath)
	nameWithoutExt := strings.TrimSuffix(basePath, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s(%d)%s", nameWithoutExt, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
