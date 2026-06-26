package vorpruefung

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"shared/models"

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

// ─── Mapping & Laden ───────────────────────────────────────────────────────────

func incomeFromRow(r models.FinancingRow) IncomeRow {
	return IncomeRow{LC: r.LC, Y1: r.Y1, Y2: r.Y2, Y3: r.Y3, EUR: r.EUR}
}

// isZeroAmount: nil oder 0 gilt als "wertlos".
func isZeroAmount(v *float64) bool {
	return v == nil || *v == 0
}

// MapScannedToBudget bildet die kanonische BudgetData auf das Generator-Modell ab.
// Hier lebt die VP-spezifische Formung (vormals in Rust): leere Kategorie-Kopfzeilen
// und namenlose 0-Platzhalter werden ausgelassen; Drittmittel ohne Geber-Aufstellung
// landen gesammelt unter "Sonstige".
func MapScannedToBudget(s *models.ScannedBudgetData) *BudgetConfig {
	cfg := &BudgetConfig{
		Eigenmittel: incomeFromRow(s.Financing.Eigenmittel),
		KMWMittel:   incomeFromRow(s.Financing.KMWMittel),
		Drittmittel: DrittmittelBlock{
			Y1:    s.Financing.Drittmittel.Y1,
			Y2:    s.Financing.Drittmittel.Y2,
			Y3:    s.Financing.Drittmittel.Y3,
			Geber: nil,
			Sonstiges: &DrittmittelSonstiges{
				LC:  s.Financing.Drittmittel.LC,
				EUR: s.Financing.Drittmittel.EUR,
			},
		},
		ReserveFreigabe: false,
	}

	for _, p := range s.Positions {
		// Kategorie wird vom Scanner aus der Nummer (1..8) abgeleitet; leer ⇒ keine
		// gültige Budget-Kategorie ⇒ überspringen.
		if p.Kategorie == "" {
			continue
		}

		sub := ""
		if idx := strings.IndexByte(p.Number, '.'); idx >= 0 {
			sub = strings.TrimSpace(p.Number[idx+1:])
		}
		labelEmpty := strings.TrimSpace(p.Label) == ""
		isHeader := sub == ""
		valueless := isZeroAmount(p.LC) && isZeroAmount(p.Y1) && isZeroAmount(p.Y2) &&
			isZeroAmount(p.Y3) && isZeroAmount(p.EUR)

		// Wertlose reine Kopfzeilen oder namenlose Platzhalter auslassen. Benannte
		// Positionen und die Sonderkategorien (Wert direkt auf der Kategoriezeile)
		// bleiben erhalten; Lücken in der Nummerierung sind gewollt.
		if valueless && (isHeader || labelEmpty) {
			continue
		}

		position := p.Label
		if labelEmpty {
			position = p.Kategorie
		}

		cfg.Ausgaben = append(cfg.Ausgaben, ExpensePos{
			Kategorie: p.Kategorie,
			ID:        p.Number,
			Position:  position,
			LC:        p.LC,
			Y1:        p.Y1,
			Y2:        p.Y2,
			Y3:        p.Y3,
			EUR:       p.EUR,
		})
	}

	return cfg
}

// loadBudgetConfig liest die kanonische BudgetData-JSON und formt sie für den
// Generator auf. Pfad leer ⇒ (nil, nil) ⇒ leeres Eingabe-Template.
func loadBudgetConfig(path string) (*BudgetConfig, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("budget-datei konnte nicht gelesen werden: %w", err)
	}
	// Bewusst NICHT DisallowUnknownFields: das Sidecar erhält die volle BudgetData
	// und nutzt nur sein Subset; weitere Felder werden ignoriert.
	var scanned models.ScannedBudgetData
	if err := json.Unmarshal(data, &scanned); err != nil {
		return nil, fmt.Errorf("budget-datei (%s) ist kein gültiges JSON: %w", path, err)
	}
	cfg := MapScannedToBudget(&scanned)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("budget-datei (%s) ist ungültig: %w", path, err)
	}
	return cfg, nil
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
