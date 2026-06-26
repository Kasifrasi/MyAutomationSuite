package vorpruefung

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"shared/models"
)

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
