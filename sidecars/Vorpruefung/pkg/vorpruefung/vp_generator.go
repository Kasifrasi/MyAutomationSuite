package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// UIProgress sendet Statusmeldungen im JSON-Format an die Standardausgabe,
// die von deiner Slint/Python-GUI leicht geparst werden können.

func GenerateVorpruefung(outputPath string, cfg GeneratorConfig) error {
	f := excelize.NewFile()

	orderedSheets := []string{
		constants.VPSheetDASHBOARD,
		constants.VPSheetBUDGET,
		constants.VPSheetKMW_MITTEL,
		constants.VPSheetFINANZBERICHTE,
		constants.VPSheetFB_PRUEFUNG,
		constants.VPSheetMA,
		constants.VPSheetMA_PRUEFUNG,
		constants.VPSheetDATEN,
	}

	for _, sheet := range orderedSheets {
		f.NewSheet(sheet)
	}

	g := &Generator{
		file:           f,
		styleCache:     make(map[string]int),
		condStyleCache: make(map[string]int),
		borderCache:    make(map[string]int),
		cfg:            cfg,
	}

	reg := NewTemplateRegistry()

	if err := g.CreateDashboardSheet(reg); err != nil {
		return fmt.Errorf("fehler beim Erstellen des Dashboard-Blatts: %w", err)
	}
	if err := g.CreateBudgetSheet(reg); err != nil {
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
	if err := g.CreateFBPruefungSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des FB-Prüfungs-Blatts: %w", err)
	}
	if err := g.CreateMAPruefungSheet(); err != nil {
		return fmt.Errorf("fehler beim Erstellen des MA-Prüfungs-Blatts: %w", err)
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

// budgetExpenseCount liefert die Anzahl der Ausgaben-Zeilen (Positionen bei Config,
// sonst die Standard-Kategorien). Bestimmt die Zeilenanzahl der FB-Ausgabentabellen.
func (g *Generator) budgetExpenseCount() int {
	if g.cfg.ExpensePositionsCount > 0 {
		return g.cfg.ExpensePositionsCount
	}
	return len(EXPENSE_CATEGORIES)
}

func (g *Generator) maGridRows() int {
	blockSize := g.budgetExpenseCount() + 4
	return MA_TABLE_COUNT * blockSize
}

func (g *Generator) budgetIncomeCount() int {
	if g.cfg.IncomeTypesCount > 0 {
		return g.cfg.IncomeTypesCount
	}
	return len(TYPE_NAMES)
}
