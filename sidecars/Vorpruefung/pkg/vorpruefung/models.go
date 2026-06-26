package vorpruefung

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ─── Eingabe-Schema (Wire-Format) ──────────────────────────────────────────────
//
// Die folgenden scanned*-Typen spiegeln die kanonische `budget_scanner::BudgetData`
// (Rust). Das Sidecar bekommt IMMER die vollständige BudgetData übergeben und
// deklariert nur die Felder, die es nutzt — `encoding/json` ignoriert den Rest.
// Ein neues Scanner-Feld wird hier mit einer einzigen zusätzlichen Zeile nutzbar.

// ─── Generator-Modell (intern) ─────────────────────────────────────────────────
//
// BudgetConfig hält die für das Blatt "I. Budget" aufbereiteten Werte. Ist keine
// Budget-Datei angegeben, bleibt das Blatt ein leeres Eingabe-Template
// (g.budget == nil). Diese Typen sind NICHT das JSON-Schema mehr — sie werden von
// MapScannedToBudget aus der kanonischen BudgetData befüllt.
//
// Beträge sind durchweg *float64: nil ⇒ Zelle bleibt leer (Eingabefeld), 0 ⇒ es
// wird ausdrücklich eine 0 eingetragen.
type BudgetConfig struct {
	Kurs            *float64
	Eigenmittel     IncomeRow
	Drittmittel     DrittmittelBlock
	KMWMittel       IncomeRow
	Ausgaben        []ExpensePos
	ReserveFreigabe bool // Checkbox "Reserve freigeben" (falls im Scanner-Output vorhanden)
}

type StyleOptions struct {
	Bold         bool
	Italic       bool
	Size         float64
	FontColor    string
	FillColor    string
	HAlign       string
	VAlign       string
	NumFormat    string
	NumFmtID     int
	BorderTop    int
	BorderBottom int
	BorderLeft   int
	BorderRight  int
	BorderColor  string
	WrapText     bool
	Strike       bool
}

type Generator struct {
	file           *excelize.File
	styleCache     map[string]int
	condStyleCache map[string]int

	rangesAusgaben   []string
	rangesEinnahmen1 []string
	rangesEinnahmen2 []string
	rangesMA         []string

	dynArrayCells    []dynArrayCell
	evalFBSelNumAddr string

	budget *BudgetConfig
}

type dynArrayCell struct {
	sheet string
	cell  string
}

// IncomeRow ist eine Finanzierungszeile mit Lokalwährung, drei Jahreswerten und EUR.
type IncomeRow struct {
	LC  *float64
	Y1  *float64
	Y2  *float64
	Y3  *float64
	EUR *float64
}

// DrittmittelBlock bündelt die Jahreswerte (Summenzeile) und die variable
// Geber-Aufstellung (Tabelle rechts im Budget).
type DrittmittelBlock struct {
	Y1        *float64
	Y2        *float64
	Y3        *float64
	Geber     []DrittmittelGeber
	Sonstiges *DrittmittelSonstiges
}

// DrittmittelGeber ist eine Zeile der Geber-Aufstellung (Name, LC, EUR).
type DrittmittelGeber struct {
	Geber string
	LC    *float64
	EUR   *float64
}

// DrittmittelSonstiges hält die Beträge für die feste "Sonstiges"-Zeile, die immer
// als letzte Zeile der Geber-Aufstellung erscheint.
type DrittmittelSonstiges struct {
	LC  *float64
	EUR *float64
}

// ExpensePos ist eine Kostenposition des Budgets. Kategorie muss einem Eintrag aus
// BG_CATEGORIES entsprechen; ID wird als fester Wert (keine Formel) übernommen.
type ExpensePos struct {
	Kategorie string
	ID        string
	Position  string
	LC        *float64
	Y1        *float64
	Y2        *float64
	Y3        *float64
	EUR       *float64
}

func (c *BudgetConfig) Validate() error {
	valid := make(map[string]bool, len(BG_CATEGORIES))
	for _, cat := range BG_CATEGORIES {
		valid[cat] = true
	}
	seenID := make(map[string]bool, len(c.Ausgaben))
	for i, p := range c.Ausgaben {
		if !valid[p.Kategorie] {
			return fmt.Errorf("ausgaben[%d]: unbekannte kategorie %q (erlaubt: %v)", i, p.Kategorie, BG_CATEGORIES)
		}
		if p.ID == "" {
			return fmt.Errorf("ausgaben[%d] (%s): id fehlt", i, p.Position)
		}
		if seenID[p.ID] {
			return fmt.Errorf("ausgaben[%d]: doppelte id %q", i, p.ID)
		}
		seenID[p.ID] = true
	}
	return nil
}
