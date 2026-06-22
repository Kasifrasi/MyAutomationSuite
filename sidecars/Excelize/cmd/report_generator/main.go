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
		base = strings.ReplaceAll(base, "{pn}", jobID)
		base = strings.ReplaceAll(base, "{la}", sprache)

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
			daten := mapScannedToReportData(&scanned)
			daten.Options = globalOptions
			jobID := scanned.ProjectNumber
			if jobID == "" {
				jobID = fmt.Sprintf("PROJ-%03d", i+1)
			}

			// Setze das Flag, um Excel-Gruppierungen auf Wunsch zu entfernen
			daten.RemoveGroupings = true
			sprache := strings.ToLower(scanned.Language)

			base := baseNames[i]
			duplicatesCounter[base]++

			dateiname := *filenameFlag
			dateiname = strings.ReplaceAll(dateiname, "{pn}", jobID)
			dateiname = strings.ReplaceAll(dateiname, "{la}", sprache)

			if nameCounts[base] > 1 {
				// Es gibt mehrere Dateien, die (ohne Zähler) gleich heißen würden
				countStr := strconv.Itoa(duplicatesCounter[base])
				if strings.Contains(dateiname, "{i}") {
					dateiname = strings.ReplaceAll(dateiname, "{i}", countStr)
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

// Hilfsfunktion zum Parsen von Geldbeträgen aus Strings (z.B. "12.000,50" oder "12000.50")
func parseAmount(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	// Wenn es Tausender-Punkte und ein Komma gibt (deutsche Formatierung)
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// mapScannedToReportData übersetzt das Rust Scanner-Modell in das Go Report-Modell
func mapScannedToReportData(scanned *report.ScannedBudgetData) report.ReportData {
	sprache := strings.ToLower(scanned.Language)

	// Falls die Sprache nicht direkt passt, mappen wir sie grob
	switch {
	case strings.Contains(sprache, "de"):
		sprache = "deutsch"
	case strings.Contains(sprache, "en"):
		sprache = "english"
	case strings.Contains(sprache, "fr"):
		sprache = "français"
	case strings.Contains(sprache, "es"):
		sprache = "español"
	case strings.Contains(sprache, "pt") || strings.Contains(sprache, "po"):
		sprache = "português"
	default:
		sprache = "deutsch" // Fallback
	}

	data := report.ReportData{
		Sprache:       sprache,
		Lokalwaehrung: scanned.LocalCurrency,
		Projektnummer: scanned.ProjectNumber,
		Projekttitel:  scanned.ProjectTitle,
		EmptyRows: report.EmptyRowsConfig{
			Global:            3, // Standardmäßig 3 leere Zeilen pro Kategorie beibehalten
			CategoryOverrides: make(map[int]int),
		},
		Eigenleistung: report.FundingRecord{Budget: parseAmount(scanned.Eigenleistung)},
		Drittmittel:   report.FundingRecord{Budget: parseAmount(scanned.Drittmittel)},
		KMWMittel:     report.FundingRecord{Budget: parseAmount(scanned.KmwMittel)},
		Categories:    make(map[int][]report.CostItem),
		HeaderBudgets: make(map[int]interface{}),
	}

	// 1. Zuerst Hauptkategorien-Budgets ("1.", "6." etc.) aufsammeln
	for _, pos := range scanned.Positions {
		if len(pos.Number) == 0 {
			continue
		}
		catID := int(pos.Number[0] - '0')
		if catID < 1 || catID > 8 {
			continue
		}

		if strings.HasSuffix(pos.Number, ".") {
			// Es ist eine Hauptkategorie! Wir speichern uns ihren Wert.
			budget := parseAmount(pos.CostCol1)
			if budget >= 0 {
				data.HeaderBudgets[catID] = budget
			}
		}
	}

	// 2. Jetzt die echten Unterpositionen zuweisen
	for _, pos := range scanned.Positions {
		if len(pos.Number) == 0 {
			continue
		}

		catID := int(pos.Number[0] - '0')
		if catID < 1 || catID > 8 {
			continue
		}

		// Hauptkategorie-Zeilen überspringen (die haben wir oben als Header-Budgets verarbeitet)
		if strings.HasSuffix(pos.Number, ".") {
			continue
		}

		// ACHTUNG: pos.Number nicht mehr dem Namen voranstellen!
		item := report.CostItem{
			Name:   pos.Label,
			Budget: parseAmount(pos.CostCol1),
		}

		data.Categories[catID] = append(data.Categories[catID], item)
	}

	// 2.5 Wir trimmen NUR abschließende Items mit Budget = 0 und leerem Namen.
	// Führende leere Items MÜSSEN erhalten bleiben, damit die Nummerierung (z.B. 1.1, 1.2, etc.)
	// nicht verschoben wird. Nur was am Ende "leer" dranhängt, wird weggeschnitten.
	for catID := 1; catID <= 8; catID++ {
		lastValid := -1

		// Definiert, was eine "gültige" Kostenposition ausmacht:
		// Entweder das Budget ist nicht 0 ODER der Name der Position ist nicht leer.
		isValidItem := func(item report.CostItem) bool {
			// Typischerweise ist Name ein string. Falls nicht, prüfen wir das hier.
			nameStr, ok := item.Name.(string)
			if !ok {
				nameStr = ""
			}

			// Budget ist ein interface{}, wir müssen den Typ sicher asserten,
			// da float64(0) != int(0) sonst in Go true ergibt.
			var budget float64
			switch v := item.Budget.(type) {
			case float64:
				budget = v
			case int:
				budget = float64(v)
			}

			return budget != 0 || strings.TrimSpace(nameStr) != ""
		}

		for i, item := range data.Categories[catID] {
			if isValidItem(item) {
				lastValid = i
			}
		}

		if lastValid != -1 {
			// Wir fangen immer bei 0 an und schneiden nur hinten ab!
			data.Categories[catID] = data.Categories[catID][0 : lastValid+1]
		} else {
			// Alles war ungültig (kein Name, kein Budget) -> Kategorie komplett leeren
			data.Categories[catID] = []report.CostItem{}
		}
	}

	return data
}
