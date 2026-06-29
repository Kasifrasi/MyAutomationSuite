package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"vorpruefung/pkg/api"
)

const exRate = 125.0
const (
	exSaldovortragLC  = 200_000.0
	exSaldovortragEUR = exSaldovortragLC / exRate // 1.600
)

var fbPeriods = []api.FBPeriod{
	{ // Periode 1
		Von: date(2025, 1, 1), Bis: date(2025, 6, 30),
		KmwLC: 1_250_000, EigenLC: 250_000, DrittLC: 125_000,
		AusgabenByID: map[string]float64{
			"1.1": 600_000, "2.1": 300_000, "3.1": 200_000,
			"3.2": 150_000, "4.1": 120_000, "5.1": 80_000, "7.1": 50_000,
		},
		BankLC: 325_000,
	},
	{ // Periode 2
		Von: date(2025, 7, 1), Bis: date(2025, 12, 31),
		KmwLC: 1_000_000, EigenLC: 200_000, DrittLC: 100_000,
		AusgabenByID: map[string]float64{
			"1.1": 400_000, "1.2": 300_000, "3.1": 200_000,
			"3.2": 150_000, "4.1": 100_000, "5.1": 70_000, "7.1": 30_000,
		},
		BankLC: 375_000,
	},
}

var maPeriods = []api.MAPeriod{
	{ // MA Periode 1
		Von: date(2025, 1, 1), Bis: date(2025, 6, 30), OandaKurs: exRate,
		KategorienLC: map[string]float64{
			"Bauausgaben": 700_000, "Investitionen": 300_000, "Personalkosten": 350_000,
			"Projektaktivitaeten": 130_000, "Projektverwaltung": 80_000, "Audit": 50_000,
		},
		EigenLC: 250_000, DrittLC: 125_000,
	},
	{ // MA Periode 2
		Von: date(2025, 7, 1), Bis: date(2025, 12, 31), OandaKurs: exRate,
		KategorienLC: map[string]float64{
			"Bauausgaben": 400_000, "Personalkosten": 350_000,
			"Projektaktivitaeten": 100_000, "Projektverwaltung": 70_000, "Audit": 30_000,
		},
		EigenLC: 200_000, DrittLC: 100_000,
	},
}

var kmwRows = []api.KMWTranche{
	{"Periode 1", "EUR", 10_000, date(2025, 1, 15)},
	{"Periode 1", "EUR", 8_000, date(2025, 4, 15)},
	{"Periode 2", "EUR", 9_000, date(2025, 7, 15)},
}

func main() {
	var inPath, budgetPath, outPath string
	flag.StringVar(&inPath, "in", "vorpruefung_output.xlsx", "mit -budget erzeugte Eingabe-Vorlage (.xlsx)")
	flag.StringVar(&budgetPath, "budget", "", "Budget-JSON (für Ausgaben-IDs/Anzahl); Standard: testdata/fixtures/budget.example.json im Repo-Root")
	flag.StringVar(&outPath, "o", "vorpruefung_befuellt.xlsx", "Zieldatei (.xlsx)")
	flag.Parse()

	if budgetPath == "" {
		root, err := repoRoot()
		if err != nil {
			log.Fatalf("repo-root nicht gefunden: %v", err)
		}
		budgetPath = filepath.Join(root, "testdata", "fixtures", "budget.example.json")
	}

	ausgIDs, err := loadAusgabenIDs(budgetPath)
	if err != nil {
		log.Fatalf("budget laden: %v", err)
	}
	if len(ausgIDs) == 0 {
		log.Fatalf("budget %q enthält keine Ausgaben – bitte eine -budget-Datei mit Positionen angeben", budgetPath)
	}

	data := api.FillData{
		Dashboard: api.DashboardData{
			Projektnummer:       "PRJ-2025-042",
			Vorprojekt:          true,
			Projekttitel:        "Aufbau Gemeindezentrum Beispielstadt",
			Projekttraeger:      "Beispiel Hilfswerk e.V.",
			Berichtswaehrung:    "USD",
			Projektstart:        date(2025, 1, 1),
			Projektende:         date(2027, 12, 31),
			Vorprojektnummer:    "PRJ-2022-017",
			VPBerichtswaehrung:  "USD",
			Vorprojektende:      date(2024, 12, 31),
			VPWechselkurs:       exRate,
			VPSaldoLC:           exSaldovortragLC,
			VPSaldoEUR:          exSaldovortragEUR,
			VPFolgeprojektstart: date(2025, 1, 1),
			DocChecklist:        []string{"Ja", "Ja", "Ja", "Ja", "Ja", "Ja", "Ja"},
		},
		KMW: kmwRows,
		MA:  maPeriods,
		FB:  fbPeriods,
		Budget: &api.BudgetData{
			AusgabenIDs: ausgIDs,
		},
	}

	if err := copyFile(inPath, outPath); err != nil {
		log.Fatalf("Konnte template nicht nach %s kopieren: %v", outPath, err)
	}

	if err := api.FillTemplate(outPath, data); err != nil {
		log.Fatalf("Fehler beim Befüllen: %v", err)
	}

	fmt.Printf("Vorlage befüllt: %s\n", outPath)
	fmt.Printf("  • Dashboard:       Projektstammdaten + Vorprojektsaldo %g LC / %g EUR\n", exSaldovortragLC, exSaldovortragEUR)
	fmt.Printf("  • KMW-Mittel:      %d Tranchen\n", len(kmwRows))
	fmt.Printf("  • Finanzberichte:  %d Perioden (Kurs %g LC/EUR, Saldo-Übertrag)\n", len(fbPeriods), exRate)
	fmt.Printf("  • Mittelanforderung: %d Perioden\n", len(maPeriods))
	fmt.Println("In Excel öffnen – dank FullCalcOnLoad rechnen alle Blätter inkl. Auswertung automatisch.")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func loadAusgabenIDs(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Positions []struct {
			Number    string `json:"number"`
			Kategorie string `json:"kategorie"`
		} `json:"positions"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s ist kein gültiges JSON: %w", path, err)
	}
	var ids []string
	for _, p := range cfg.Positions {
		if p.Kategorie == "" {
			continue
		}
		ids = append(ids, p.Number)
	}
	return ids, nil
}

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func repoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
