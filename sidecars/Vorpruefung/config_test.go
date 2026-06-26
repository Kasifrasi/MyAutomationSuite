package main

import "testing"

func f(v float64) *float64 { return &v }

func pos(number, label, kategorie string, lc, y1, eur *float64) scannedPosition {
	return scannedPosition{Number: number, Label: label, Kategorie: kategorie, LC: lc, Y1: y1, EUR: eur}
}

func ids(cfg *BudgetConfig) []string {
	out := make([]string, len(cfg.Ausgaben))
	for i, a := range cfg.Ausgaben {
		out[i] = a.ID
	}
	return out
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Kopfzeile ohne Wert raus, Sonderkategorien (Wert auf der Kategoriezeile) bleiben.
func TestMapSkipsEmptyHeaderKeepsSpecial(t *testing.T) {
	s := &scannedBudgetData{Positions: []scannedPosition{
		pos("1.", "Bauausgaben", "Bauausgaben", nil, nil, nil),               // Kopfzeile ohne Wert -> raus
		pos("1.1", "", "Bauausgaben", f(10000), f(10000), f(5000)),           // normale Position
		pos("6.", "Evaluierung", "Evaluierung", f(10000), f(10000), f(5000)), // Wert auf Kategoriezeile
		pos("7.", "Audit", "Audit", f(10000), f(10000), f(5000)),
		pos("8.", "Reserve", "Reserve", f(79000), f(79000), f(39500)),
	}}
	cfg := mapScannedToBudget(s)

	if got, want := ids(cfg), []string{"1.1", "6.", "7.", "8."}; !eqStrings(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	// 1.1 hat leeres Label -> Position = Kategoriename.
	if cfg.Ausgaben[0].Position != "Bauausgaben" {
		t.Errorf("Position[1.1] = %q, want \"Bauausgaben\"", cfg.Ausgaben[0].Position)
	}
	eval := cfg.Ausgaben[1]
	if eval.Kategorie != "Evaluierung" || eval.Position != "Evaluierung" || eval.LC == nil || *eval.LC != 10000 {
		t.Errorf("Evaluierung falsch gemappt: %+v", eval)
	}
}

// Leere Platzhalter raus, ohne die übrigen IDs zu verschieben; benannte 0-Position bleibt.
func TestMapFiltersPlaceholdersWithoutRenumbering(t *testing.T) {
	s := &scannedBudgetData{Positions: []scannedPosition{
		pos("1.", "", "Bauausgaben", nil, nil, nil),                  // Kopfzeile -> raus
		pos("1.1", "", "Bauausgaben", f(10000), f(10000), f(5000)),   // Wert -> bleibt
		pos("1.2", "", "Bauausgaben", f(0), nil, f(0)),               // leer + 0 -> raus
		pos("1.3", "", "Bauausgaben", f(11000), f(11000), f(5500)),   // Wert -> bleibt
		pos("1.4", "Büromaterial", "Bauausgaben", f(0), nil, f(0)),   // Name -> bleibt (auch bei 0)
	}}
	cfg := mapScannedToBudget(s)

	if got, want := ids(cfg), []string{"1.1", "1.3", "1.4"}; !eqStrings(got, want) {
		t.Fatalf("ids = %v, want %v", got, want)
	}
	bm := cfg.Ausgaben[2]
	if bm.Position != "Büromaterial" || bm.LC == nil || *bm.LC != 0 {
		t.Errorf("Büromaterial falsch gemappt: %+v", bm)
	}
}

// Drittmittel ohne Geber landen als Summe unter "Sonstiges".
func TestMapDrittmittelSonstiges(t *testing.T) {
	s := &scannedBudgetData{Financing: scannedFinancing{
		Drittmittel: scannedFinancingRow{LC: f(3750000), Y1: f(1500000), Y2: f(1250000), Y3: f(1000000), EUR: f(30000)},
	}}
	cfg := mapScannedToBudget(s)

	if cfg.Drittmittel.Sonstiges == nil || cfg.Drittmittel.Sonstiges.LC == nil || *cfg.Drittmittel.Sonstiges.LC != 3750000 {
		t.Fatalf("Sonstiges.LC falsch: %+v", cfg.Drittmittel.Sonstiges)
	}
	if cfg.Drittmittel.Y1 == nil || *cfg.Drittmittel.Y1 != 1500000 {
		t.Errorf("Drittmittel.Y1 falsch: %+v", cfg.Drittmittel.Y1)
	}
	if len(cfg.Drittmittel.Geber) != 0 {
		t.Errorf("Geber sollte leer sein, ist %v", cfg.Drittmittel.Geber)
	}
}
