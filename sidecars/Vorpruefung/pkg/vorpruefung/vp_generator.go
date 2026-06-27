package vorpruefung

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"shared/constants"
	"shared/excel"
)

// UIProgress sendet Statusmeldungen im JSON-Format an die Standardausgabe,
// die von deiner Slint/Python-GUI leicht geparst werden können.

func GenerateVorpruefung(outputPath string, budgetCfg *BudgetConfig) error {
	f := excelize.NewFile()

	orderedSheets := []string{
		constants.VPSheetDASHBOARD,
		constants.VPSheetBUDGET,
		constants.VPSheetKMW_MITTEL,
		constants.VPSheetFINANZBERICHTE,
		constants.VPSheetMA,
		constants.VPSheetAUSWERTUNG,
		constants.VPSheetDATEN,
	}

	for _, sheet := range orderedSheets {
		f.NewSheet(sheet)
	}

	if err := excel.UnlockSheetsArea(f, orderedSheets); err != nil {
		return fmt.Errorf("fehler beim Entsperren der Arbeitsbereiche: %w", err)
	}

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

// budgetExpenseCount liefert die Anzahl der Ausgaben-Zeilen (Positionen bei Config,
// sonst die Standard-Kategorien). Bestimmt die Zeilenanzahl der FB-Ausgabentabellen.
func (g *Generator) budgetExpenseCount() int {
	if g.budget != nil {
		return len(g.budget.Ausgaben)
	}
	return len(EXPENSE_CATEGORIES)
}

// fbExpenseRowsForCategory liefert die FB-Ausgaben-Zeilennummern (auf dem Blatt
// "III. Finanzberichte"), die zu einer Kostenkategorie gehören. Die erste
// Ausgaben-Datenzeile liegt bei FB_AUSG_FIRST_ROW; Position i ⇒ Zeile +i.
func (g *Generator) fbExpenseRowsForCategory(cat string) []int {
	var rows []int
	if g.budget != nil {
		for i, p := range g.budget.Ausgaben {
			if p.Kategorie == cat {
				rows = append(rows, FB_AUSG_FIRST_ROW+i)
			}
		}
		return rows
	}
	for i, c := range EXPENSE_CATEGORIES {
		if c == cat {
			rows = append(rows, FB_AUSG_FIRST_ROW+i)
		}
	}
	return rows
}
