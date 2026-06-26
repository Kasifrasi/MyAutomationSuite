package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"shared/models"
	"shared/runner"

	"excelize-test/pkg/report"
)

var (
	initOnce sync.Once
	initErr  error
)

func main() {
	runner.Run(func(data models.ScannedBudgetData, outputPath string, optionsJSON string) error {
		var globalOptions report.ReportOptions
		if optionsJSON != "" {
			if err := json.Unmarshal([]byte(optionsJSON), &globalOptions); err != nil {
				return fmt.Errorf("konnte -options JSON nicht parsen: %v", err)
			}
		}

		initOnce.Do(func() {
			initErr = report.PreloadAllTemplates(globalOptions)
		})
		if initErr != nil {
			return fmt.Errorf("fehler beim Preload: %v", initErr)
		}

		reportData := report.MapScannedToReportData(&data, globalOptions.IsTemplate)
		reportData.Options = globalOptions
		reportData.EmptyRows.Global = globalOptions.EmptyRows

		// Die JobID interessiert in diesem neuen Flow nicht mehr wirklich so stark für Fehler,
		// aber sie ist logischerweise die Projektnummer:
		jobID := data.ProjectNumber
		if jobID == "" {
			jobID = "PROJ-FB"
		}

		return report.ProcessSingleJob(report.ReportJob{
			JobID:      jobID,
			OutputPath: outputPath,
			Data:       reportData,
		})
	})
}
