package report

import (
	"fmt"
	"os"
	"sync"
)

// ReportJob kapselt alle Daten, die für die Erstellung eines Berichts in der Pipeline nötig sind.
type ReportJob struct {
	JobID      string
	OutputPath string
	Data       ReportData
}

// ReportResult liefert den Status eines verarbeiteten Jobs zurück.
type ReportResult struct {
	JobID string
	Err   error
}

// StartPipeline startet einen asynchronen Worker-Pool.
// Er konsumiert Jobs aus dem `jobs` Channel, erstellt die Excel-Berichte
// und sendet das Ergebnis an den zurückgegebenen Results-Channel.
func StartPipeline(numWorkers int, jobs <-chan ReportJob) <-chan ReportResult {
	results := make(chan ReportResult, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				err := processSingleJob(job)
				results <- ReportResult{JobID: job.JobID, Err: err}
			}
		}(i)
	}

	// Schließe den results-Channel, wenn alle Worker fertig sind
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func processSingleJob(job ReportJob) error {
	sprache, ok := job.Data.Sprache.(string)
	if !ok || sprache == "" {
		return fmt.Errorf("sprache in ReportData ist nicht definiert oder kein string")
	}

	rep, err := NewExcelReportByLanguage(sprache)
	if err != nil {
		return fmt.Errorf("laden der Vorlage fehlgeschlagen: %w", err)
	}
	defer rep.Close()

	// Formeln säubern
	if err := rep.CleanUpNumberingFormulas(); err != nil {
		// Logikfehler beim Säubern ignorieren wir für den Stream,
		// um den Prozess nicht komplett abzubrechen.
	}

	// Dynamische Daten injizieren (Der Performance-intensive Teil)
	if err := rep.UpdateData(job.Data); err != nil {
		return fmt.Errorf("update der Daten fehlgeschlagen: %w", err)
	}

	// Bugfix: Shared Formulas un-sharen, da Excelize RemoveRow diese korrumpiert!
	rep.UnshareAllFormulas()

	// Sollen die Gruppierungen restlos entfernt werden?
	if job.Data.RemoveGroupings {
		// Excelize bietet keine API, um Gruppierungen zu LÖSCHEN (SetRowOutlineLevel(0) wirft einen Fehler).
		// Daher schreiben wir den Report in den RAM, filtern das XML und speichern dann!
		buf, err := rep.file.WriteToBuffer()
		if err != nil {
			return fmt.Errorf("fehler beim Schreiben in Puffer: %w", err)
		}

		cleanBytes, err := RemoveGroupingsFromBytes(buf.Bytes())
		if err != nil {
			return fmt.Errorf("fehler beim Entfernen der Gruppierungen: %w", err)
		}

		// Bereinigte Datei auf Festplatte schreiben
		if err := os.WriteFile(job.OutputPath, cleanBytes, 0644); err != nil {
			return fmt.Errorf("speichern unter %s fehlgeschlagen: %w", job.OutputPath, err)
		}
	} else {
		// Report normal auf die Festplatte schreiben
		if err := rep.SaveAs(job.OutputPath); err != nil {
			return fmt.Errorf("speichern unter %s fehlgeschlagen: %w", job.OutputPath, err)
		}
	}

	return nil
}
