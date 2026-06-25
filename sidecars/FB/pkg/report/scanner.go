package report

// BudgetPosition repräsentiert eine einzelne Kostenposition (z.B. "1.1 Personal")
type BudgetPosition struct {
	Number   string `json:"number"`
	Label    string `json:"label"`
	CostCol1 string `json:"cost_col1"`
	CostCol2 string `json:"cost_col2"`
}

// ScannedBudgetData entspricht dem extrahierten Rust `BudgetData` Struct
type ScannedBudgetData struct {
	FilePath      string           `json:"file_path"`
	SheetName     string           `json:"sheet_name"`
	Version       string           `json:"version"`
	ProjectTitle  string           `json:"project_title"`
	ProjectNumber string           `json:"project_number"`
	Language      string           `json:"language"`
	LocalCurrency string           `json:"local_currency"`
	CostCol1      int              `json:"cost_col1"`
	CostCol2      *int             `json:"cost_col2"`
	Eigenleistung string           `json:"eigenleistung"`
	Drittmittel   string           `json:"drittmittel"`
	KmwMittel     string           `json:"kmw_mittel"`
	Positions     []BudgetPosition `json:"positions"`
}
