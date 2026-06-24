package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// BudgetConfig hält die optional per JSON (-budget) übergebenen Budgetwerte. Ist
// keine Datei angegeben, bleibt das Blatt "I. Budget" ein leeres Eingabe-Template
// (g.budget == nil) und alle bisherigen Standardstrukturen greifen unverändert.
//
// Beträge sind durchweg *float64: nil ⇒ Zelle bleibt leer (Eingabefeld), 0 ⇒ es
// wird ausdrücklich eine 0 eingetragen.
type BudgetConfig struct {
	// Kurs ist optional/informativ. Der Budget-Kurs der Tabelle wird weiterhin aus
	// Gesamt-LC / Gesamt-EUR abgeleitet (Summe der Einnahmen), daher hier nicht
	// zwingend nötig.
	Kurs *float64 `json:"kurs"`

	Eigenmittel     IncomeRow        `json:"eigenmittel"`
	Drittmittel     DrittmittelBlock `json:"drittmittel"`
	KMWMittel       IncomeRow        `json:"kmwMittel"`
	Ausgaben        []ExpensePos     `json:"ausgaben"`
	ReserveFreigabe bool             `json:"reserveFreigabe"`
}

// IncomeRow ist eine Finanzierungszeile mit Lokalwährung, drei Jahreswerten und EUR.
type IncomeRow struct {
	LC  *float64 `json:"lc"`
	Y1  *float64 `json:"y1"`
	Y2  *float64 `json:"y2"`
	Y3  *float64 `json:"y3"`
	EUR *float64 `json:"eur"`
}

// DrittmittelBlock bündelt die Jahreswerte (Summenzeile) und die variable
// Geber-Aufstellung (Tabelle rechts im Budget).
type DrittmittelBlock struct {
	Y1        *float64              `json:"y1"`
	Y2        *float64              `json:"y2"`
	Y3        *float64              `json:"y3"`
	Geber     []DrittmittelGeber    `json:"geber"`
	Sonstiges *DrittmittelSonstiges `json:"sonstiges,omitempty"`
}

// DrittmittelGeber ist eine Zeile der Geber-Aufstellung (Name, LC, EUR).
type DrittmittelGeber struct {
	Geber string   `json:"geber"`
	LC    *float64 `json:"lc"`
	EUR   *float64 `json:"eur"`
}

// DrittmittelSonstiges hält die Beträge für die feste "Sonstiges"-Zeile, die immer
// als letzte Zeile der Geber-Aufstellung erscheint. Wird in der JSON-Übergabe nicht
// angegeben, bleiben LC und EUR leer (Eingabefeld). Die Allokationslogik — also wie
// viel der Gesamtbetrag auf "Sonstiges" entfällt — liegt beim Aufrufer.
type DrittmittelSonstiges struct {
	LC  *float64 `json:"lc"`
	EUR *float64 `json:"eur"`
}

// ExpensePos ist eine Kostenposition des Budgets. Kategorie muss einem Eintrag aus
// BG_CATEGORIES entsprechen; ID wird als fester Wert (kein Formel) übernommen.
type ExpensePos struct {
	Kategorie string   `json:"kategorie"`
	ID        string   `json:"id"`
	Position  string   `json:"position"`
	LC        *float64 `json:"lc"`
	Y1        *float64 `json:"y1"`
	Y2        *float64 `json:"y2"`
	Y3        *float64 `json:"y3"`
	EUR       *float64 `json:"eur"`
}

// loadBudgetConfig liest und validiert die Budget-JSON. Pfad leer ⇒ (nil, nil).
func loadBudgetConfig(path string) (*BudgetConfig, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("budget-datei konnte nicht gelesen werden: %w", err)
	}
	var cfg BudgetConfig
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("budget-datei (%s) ist kein gültiges JSON: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("budget-datei (%s) ist ungültig: %w", path, err)
	}
	return &cfg, nil
}

func (c *BudgetConfig) validate() error {
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
