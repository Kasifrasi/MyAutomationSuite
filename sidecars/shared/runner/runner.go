package runner

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"shared/models"
)

type UIProgress struct {
	Status  string `json:"status"`
	File    string `json:"file,omitempty"`
	Current int    `json:"current,omitempty"`
	Total   int    `json:"total,omitempty"`
	Message string `json:"message,omitempty"`
}

func PrintUI(status, file, msg string, current, total int) {
	data, _ := json.Marshal(UIProgress{
		Status:  status,
		File:    file,
		Current: current,
		Total:   total,
		Message: msg,
	})
	fmt.Println(string(data))
}

type Job struct {
	JobID      string
	OutputPath string
	Data       models.ScannedBudgetData
}

type Result struct {
	JobID string
	Err   error
}

// GenerateFunc is implemented by the specific sidecar to generate the file.
type GenerateFunc func(data models.ScannedBudgetData, outputPath string, optionsJSON string) error

func Run(generate GenerateFunc) {
	inputFlag := flag.String("input", "", "Pfad zu einer JSON-Datei mit den verarbeiteten Scanner-Daten")
	outputFlag := flag.String("output", "test/output", "Ordner, in dem die finalen Dateien gespeichert werden")
	filenameFlag := flag.String("filename", "Output_{pn}_{la}.xlsx", "Namensmuster")
	optionsFlag := flag.String("options", "", "JSON-String mit globalen Optionen (z.B. Blattschutz)")

	// Backwards compatibility for single-file mode (VP legacy)
	var legacyOutputPath string
	var legacyBudgetPath string
	flag.StringVar(&legacyOutputPath, "o", "", "Legacy: output file path")
	flag.StringVar(&legacyBudgetPath, "budget", "", "Legacy: optionale Budget-JSON")

	flag.Parse()

	if legacyBudgetPath != "" || legacyOutputPath != "" {
		// Legacy Modus (Einzelfile)
		dataBytes, err := os.ReadFile(legacyBudgetPath)
		var scanned models.ScannedBudgetData
		if err == nil {
			_ = json.Unmarshal(dataBytes, &scanned)
		}

		out := legacyOutputPath
		if out == "" {
			out = "output.xlsx"
		}

		if err := generate(scanned, out, *optionsFlag); err != nil {
			log.Fatalf("%v", err)
		}
		fmt.Printf("Erfolgreich generiert: %s\n", out)
		return
	}

	if *inputFlag == "" || !strings.HasSuffix(strings.ToLower(*inputFlag), ".json") {
		fmt.Println("Fehler: Bitte gib eine gültige JSON-Datei mit -input an")
		os.Exit(1)
	}

	outputOrdner := *outputFlag
	if err := os.MkdirAll(outputOrdner, 0755); err != nil {
		PrintUI("error", "", fmt.Sprintf("Konnte Output-Ordner nicht erstellen: %v", err), 0, 0)
		os.Exit(1)
	}

	PrintUI("start", "", "App gestartet", 0, 0)

	bytesData, err := os.ReadFile(*inputFlag)
	if err != nil {
		PrintUI("error", "", fmt.Sprintf("Konnte JSON-Datei nicht lesen: %v", err), 0, 0)
		os.Exit(1)
	}

	var scannedDaten []models.ScannedBudgetData
	if err := json.Unmarshal(bytesData, &scannedDaten); err != nil {
		PrintUI("error", "", fmt.Sprintf("Konnte JSON-Daten nicht parsen: %v", err), 0, 0)
		os.Exit(1)
	}

	totalFiles := len(scannedDaten)
	if totalFiles == 0 {
		PrintUI("done", "", "Keine Daten zum Verarbeiten gefunden", 0, 0)
		return
	}

	nameCounts := make(map[string]int)
	baseNames := make([]string, totalFiles)

	for i, scanned := range scannedDaten {
		jobID := scanned.ProjectNumber
		if jobID == "" {
			jobID = fmt.Sprintf("PROJ-%03d", i+1)
		}
		sprache := strings.ToLower(scanned.Language)

		base := *filenameFlag
		base = strings.ReplaceAll(base, "{i}", "")
		base = strings.ReplaceAll(base, "{I}", "")
		base = strings.ReplaceAll(base, "{pn}", jobID)
		base = strings.ReplaceAll(base, "{PN}", jobID)
		base = strings.ReplaceAll(base, "{la}", sprache)
		base = strings.ReplaceAll(base, "{LA}", sprache)
		base = strings.ReplaceAll(base, "{version}", scanned.Version)
		base = strings.ReplaceAll(base, "{VERSION}", scanned.Version)
		base = strings.ReplaceAll(base, "{pt}", scanned.ProjectTitle)
		base = strings.ReplaceAll(base, "{PT}", scanned.ProjectTitle)

		baseNames[i] = base
		nameCounts[base]++
	}

	jobs := make(chan Job, totalFiles)
	results := make(chan Result, totalFiles)

	numWorkers := runtime.NumCPU() * 2
	if numWorkers > 32 {
		numWorkers = 32
	}

	for w := 0; w < numWorkers; w++ {
		go func() {
			for job := range jobs {
				err := generate(job.Data, job.OutputPath, *optionsFlag)
				results <- Result{JobID: job.JobID, Err: err}
			}
		}()
	}

	go func() {
		duplicatesCounter := make(map[string]int)
		for i, scanned := range scannedDaten {
			jobID := scanned.ProjectNumber
			if jobID == "" {
				jobID = fmt.Sprintf("PROJ-%03d", i+1)
			}
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
			dateiname = strings.ReplaceAll(dateiname, "{pt}", scanned.ProjectTitle)
			dateiname = strings.ReplaceAll(dateiname, "{PT}", scanned.ProjectTitle)

			if nameCounts[base] > 1 {
				countStr := strconv.Itoa(duplicatesCounter[base])
				if strings.Contains(dateiname, "{i}") {
					dateiname = strings.ReplaceAll(dateiname, "{i}", countStr)
				} else if strings.Contains(dateiname, "{I}") {
					dateiname = strings.ReplaceAll(dateiname, "{I}", countStr)
				} else {
					ext := filepath.Ext(dateiname)
					if ext == "" {
						ext = ".xlsx"
						dateiname += ext
					}
					nameOhneExt := strings.TrimSuffix(dateiname, ext)
					dateiname = fmt.Sprintf("%s_%s%s", nameOhneExt, countStr, ext)
				}
			} else {
				dateiname = strings.ReplaceAll(dateiname, "{i}", "")
				dateiname = strings.ReplaceAll(dateiname, "{I}", "")
				dateiname = strings.ReplaceAll(dateiname, "__", "_")
				dateiname = strings.ReplaceAll(dateiname, "_.xlsx", ".xlsx")
				dateiname = strings.ReplaceAll(dateiname, "-.xlsx", ".xlsx")
			}

			if filepath.Ext(dateiname) == "" {
				dateiname += ".xlsx"
			}

			outputPath := filepath.Join(outputOrdner, dateiname)
			outputPath = uniqueOutputPath(outputPath)

			jobs <- Job{
				JobID:      jobID,
				OutputPath: outputPath,
				Data:       scanned,
			}
		}
		close(jobs)
	}()

	successCount := 0
	errorCount := 0
	currentDone := 0

	for i := 0; i < totalFiles; i++ {
		res := <-results
		currentDone++
		if res.Err != nil {
			PrintUI("error", res.JobID, fmt.Sprintf("Fehler beim Verarbeiten: %v", res.Err), currentDone, totalFiles)
			errorCount++
		} else {
			PrintUI("progress", res.JobID, "Erfolgreich generiert", currentDone, totalFiles)
			successCount++
		}
	}

	PrintUI("done", "", fmt.Sprintf("Fertig! %d erfolgreich, %d fehlerhaft.", successCount, errorCount), totalFiles, totalFiles)
}

func uniqueOutputPath(basePath string) string {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
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
