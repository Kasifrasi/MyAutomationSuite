package main

import (
	"log"
	"os"
	"path/filepath"
	"vorpruefung/pkg/api"

	"github.com/xuri/excelize/v2"
)

func runThreeOutputs() {
	inPath := "vorpruefung_base.xlsx"

	root, _ := repoRoot()
	tmpDir := filepath.Join(root, "tmp")
	_ = os.MkdirAll(tmpDir, 0755)

	out1 := filepath.Join(tmpDir, "1_bare_no_einnahmen.xlsx")
	out2 := filepath.Join(tmpDir, "2_defaults_no_einnahmen.xlsx")
	out3 := filepath.Join(tmpDir, "3_full.xlsx")

	budgetData, err := loadBudgetData(filepath.Join(root, "testdata/fixtures/budget.example.json"))
	if err != nil {
		log.Fatalf("budget laden: %v", err)
	}

	fbPeriodsOhneEinnahmen := []api.FBPeriod{
		{
			Von: date(2025, 1, 1), Bis: date(2025, 6, 30),
			AusgabenByID: map[string]float64{"1.1": 600_000},
			BankLC:       floatPtr(325_000),
		},
	}

	fbPeriodsMitEinnahmen := []api.FBPeriod{
		{
			Von: date(2025, 1, 1), Bis: date(2025, 6, 30),
			Einnahmen1: []api.FBEinnahme{
				{Typ: "KMW-Mittel", LC: floatPtr(1_250_000)},
			},
			EinnahmenWK: []api.FBEinnahme{
				{Typ: "Eigenmittel", LC: floatPtr(250_000)},
				{Typ: "Drittmittel", Geber: "Beispiel-Geber 1", LC: floatPtr(125_000)},
			},
			AusgabenByID: map[string]float64{"1.1": 600_000},
			BankLC:       floatPtr(325_000),
		},
	}

	// 1. Ohne Standardwerte, ohne Einnahmen
	// We want to test bare metal, so we copy budgetData but set ReserveFreigabe to nil
	bareBudget := *budgetData
	bareBudget.ReserveFreigabe = nil
	bareBudget.Ausgaben = nil
	bareBudget.Eigenmittel = nil
	bareBudget.KMWMittel = nil
	bareBudget.DrittmittelY1 = nil
	bareBudget.DrittmittelY2 = nil
	bareBudget.DrittmittelY3 = nil
	bareBudget.DrittGeber = nil
	bareBudget.DrittSonstiges = nil
	bareDashboard := api.DashboardData{}

	data1 := api.FillData{FB: nil, Budget: &bareBudget, Dashboard: bareDashboard}
	err = copyFile(inPath, out1)
	if err == nil {
		err = api.FillTemplate(out1, data1)
	}
	if err != nil {
		log.Fatalf("Fehler 1: %v", err)
	}

	// 2. Mit Standardwerten, ohne Einnahmen
	data2 := api.FillData{FB: fbPeriodsOhneEinnahmen, Budget: budgetData}
	err = copyFile(inPath, out2)
	if err == nil {
		f, _ := excelize.OpenFile(out2)
		api.FillDefaults(f)
		f.Save()
		f.Close()
		err = api.FillTemplate(out2, data2)
	}
	if err != nil {
		log.Fatalf("Fehler 2: %v", err)
	}

	// 3. Mit Standardwerten, mit Einnahmen
	strFB := "Neuester FB"
	strMA := "Neueste MA"

	data3 := api.FillData{
		FB:     fbPeriodsMitEinnahmen,
		Budget: budgetData,
		FBPruefung: &api.FBPruefungData{
			Auswahl: &strFB,
		},
		MAPruefung: &api.MAPruefungData{
			Auswahl: &strMA,
		},
	}
	err = copyFile(inPath, out3)
	if err == nil {
		f, _ := excelize.OpenFile(out3)
		api.FillDefaults(f)
		f.Save()
		f.Close()
		err = api.FillTemplate(out3, data3)
	}
	if err != nil {
		log.Fatalf("Fehler 3: %v", err)
	}

	log.Printf("Erfolgreich 3 Testdateien erzeugt im Ordner: %s\n - %s\n - %s\n - %s\n", tmpDir, filepath.Base(out1), filepath.Base(out2), filepath.Base(out3))
}
