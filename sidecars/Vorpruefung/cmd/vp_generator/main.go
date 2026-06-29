package main

import (
	"shared/models"
	"shared/runner"
	"vorpruefung/pkg/vorpruefung"
)

func main() {
	runner.Run(func(data models.ScannedBudgetData, outputPath string, optionsJSON string) error {
		expCount := len(data.Positions)

		cfg := vorpruefung.GeneratorConfig{
			ExpensePositionsCount: expCount,
		}
		return vorpruefung.GenerateVorpruefung(outputPath, cfg)
	})
}
