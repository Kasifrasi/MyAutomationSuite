package report

// Die folgenden Typen spiegeln die kanonische `budget_scanner::BudgetData` (Rust).
// Es werden bewusst NUR die Felder deklariert, die das FB-Sidecar verwendet —
// `encoding/json` ignoriert alle weiteren Felder. Ein neues Scanner-Feld lässt sich
// hier mit einer einzigen zusätzlichen Struct-Zeile nutzbar machen.

// FinancingRow ist eine Finanzierungszeile (typisierte Beträge, nil = leere Zelle).
type FinancingRow struct {
	LC  *float64 `json:"lc"`
	Y1  *float64 `json:"y1"`
	Y2  *float64 `json:"y2"`
	Y3  *float64 `json:"y3"`
	EUR *float64 `json:"eur"`
}

// Financing bündelt die drei Finanzierungsquellen.
type Financing struct {
	Eigenmittel FinancingRow `json:"eigenmittel"`
	Drittmittel FinancingRow `json:"drittmittel"`
	KMWMittel   FinancingRow `json:"kmw_mittel"`
}

// BudgetPosition repräsentiert eine einzelne Kostenposition (z.B. "1.1 Personal").
type BudgetPosition struct {
	Number    string   `json:"number"`
	Label     string   `json:"label"`
	Kategorie string   `json:"kategorie"`
	LC        *float64 `json:"lc"`
	Y1        *float64 `json:"y1"`
	Y2        *float64 `json:"y2"`
	Y3        *float64 `json:"y3"`
	EUR       *float64 `json:"eur"`
}

// ScannedBudgetData entspricht dem extrahierten Rust `BudgetData` Struct.
type ScannedBudgetData struct {
	Version       string           `json:"version"`
	ProjectTitle  string           `json:"project_title"`
	ProjectNumber string           `json:"project_number"`
	Language      string           `json:"language"`
	LocalCurrency string           `json:"local_currency"`
	Financing     Financing        `json:"financing"`
	Positions     []BudgetPosition `json:"positions"`
}

// amount liefert den Betrag oder 0, falls die Zelle leer war (nil).
func amount(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
