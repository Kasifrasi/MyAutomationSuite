package vorpruefung

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

// UIProgress sendet Statusmeldungen im JSON-Format an die Standardausgabe,
// die von deiner Slint/Python-GUI leicht geparst werden können.

func GenerateVorpruefung(outputPath string, budgetCfg *BudgetConfig) error {
	f := excelize.NewFile()

	g := &Generator{
		file:           f,
		styleCache:     make(map[string]int),
		condStyleCache: make(map[string]int),
		budget:         budgetCfg,
	}

	if err := g.CreateDashboardSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Dashboard-Blatts: %w", err)
	}
	if err := g.CreateBudgetSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Budget-Blatts: %w", err)
	}
	if err := g.CreateKMWMittelSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des KMW-Mittel-Blatts: %w", err)
	}
	if err := g.CreateFinanzberichteSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Finanzberichte-Blatts: %w", err)
	}
	if err := g.CreateMittelanforderungSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Mittelanforderung-Blatts: %w", err)
	}
	if err := g.CreateAuswertungSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Auswertungs-Blatts: %w", err)
	}
	if err := g.CreateDatenSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Daten-Blatts: %w", err)
	}

	_ = f.DeleteSheet("Sheet1")

	fullCalc := true
	if err := f.SetCalcProps(&excelize.CalcPropsOptions{FullCalcOnLoad: &fullCalc}); err != nil {
		return fmt.Errorf("fehler beim Setzen der Berechnungsoptionen: %w", err)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("fehler beim Speichern des Dokuments: %w", err)
	}

	if err := applyDynamicArrayMetadata(outputPath, g.dynArrayCells); err != nil {
		return fmt.Errorf("fehler beim Setzen der Dynamic-Array-Metadaten: %w", err)
	}

	return nil
}
