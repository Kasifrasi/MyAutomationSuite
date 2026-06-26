package main

import (
	"fmt"
	"shared/models"
	"shared/runner"
	"vorpruefung/pkg/vorpruefung"
)

func main() {
	runner.Run(func(data models.ScannedBudgetData, outputPath string, optionsJSON string) error {
		cfg := vorpruefung.MapScannedToBudget(&data)
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("budget-daten sind ungültig: %w", err)
		}
		return vorpruefung.GenerateVorpruefung(outputPath, cfg)
	})
}
