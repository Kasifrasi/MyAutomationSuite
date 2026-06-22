package report

import (
	"github.com/xuri/excelize/v2"
)

// Rate repräsentiert eine Rate (Zahlung) mit Datum, EUR, Lokalwährung (LW) und Wechselkurs (WK).
type Rate struct {
	Datum interface{}
	EUR   interface{}
	LW    interface{}
	WK    interface{}
}

// FundingRecord repräsentiert eine Finanzierungszeile für Saldovortrag, Eigenleistung etc.
type FundingRecord struct {
	Budget      interface{} // Spalte D
	EinnahmenBZ interface{} // Spalte E (Berichtszeitraum)
	EinnahmenGS interface{} // Spalte F (Gesamtlaufzeit)
	Begruendung interface{} // Spalte H
}

// CostItem repräsentiert eine dynamische Kostenposition.
type CostItem struct {
	Name        interface{} // Spalte C
	Budget      interface{} // Spalte D
	AusgabenBZ  interface{} // Spalte E
	AusgabenGS  interface{} // Spalte F
	Begruendung interface{} // Spalte H
}

// SaldoBreakdown für die Saldoaufschlüsselung am Ende der Datei.
type SaldoBreakdown struct {
	Bank      interface{} // Spalte E
	Kasse     interface{} // Spalte E
	Sonstiges interface{} // Spalte E
}

// EmptyRowsConfig definiert, wie viele leere Kostenpositionen am Ende einer Kategorie verbleiben sollen.
type EmptyRowsConfig struct {
	Global            int         // Gilt für alle Kategorien, sofern nicht überschrieben (z.B. 3)
	CategoryOverrides map[int]int // Spezifische Überschreibungen, z.B. map[1]0 für Kategorie 1 -> 0 leere Zeilen
}

// ReportData enthält alle statischen und dynamischen Felder der API
type ReportData struct {
	Sprache                         interface{}
	Lokalwaehrung                   interface{}
	Projektnummer                   interface{}
	Projekttitel                    interface{}
	ProjektlaufzeitBeginn           interface{}
	ProjektlaufzeitEnde             interface{}
	AktuellerBerichtszeitraumBeginn interface{}
	AktuellerBerichtszeitraumEnde   interface{}
	RemoveGroupings                 bool

	// Konfiguration für das dynamische Einfügen/Löschen
	EmptyRows EmptyRowsConfig

	Rates [36]Rate

	Saldovortrag  FundingRecord
	Eigenleistung FundingRecord
	Drittmittel   FundingRecord
	KMWMittel     FundingRecord
	Zinsertraege  FundingRecord

	// --- Dynamische Kostenkategorien ---
	Categories    map[int][]CostItem
	HeaderBudgets map[int]interface{} // Speichert Budgets für Hauptkategorien (Modus 0)

	// --- Saldoaufschlüsselung ---
	Saldo SaldoBreakdown

	// --- Optionen (Schutz, Verbergen) ---
	Options ReportOptions
}

// ExcelReport ist unsere API-Hülle um die Excel-Datei.
type ExcelReport struct {
	file         *excelize.File
	sheet        string
	CatStartRows map[int]int
	CatEndRows   map[int]int
	GlobalSumRow int
	BankRow      int
	KasseRow     int
	SonstRow     int

	// Cache für wiederverwendbare Styles, um das Excelize Style-Limit nicht zu sprengen
	// Map-Key ist eine Kombination aus Basis-Style-ID und gewünschter Hex-Farbe (z.B. "15_#FFFFFF")
	styleCache map[string]int
}

// ReportOptions enthält Einstellungen für den Blattschutz, Mappenschutz und das Ausblenden von Spalten
type ReportOptions struct {
	ProtectSheet     bool   `json:"protect_sheet"`
	ProtectWorkbook  bool   `json:"protect_workbook"`
	SheetPassword    string `json:"sheet_password"`
	WorkbookPassword string `json:"workbook_password"`
	HideColumns      bool   `json:"hide_columns"`
	SelectLocked     bool   `json:"select_locked"`
	SelectUnlocked   bool   `json:"select_unlocked"`
	FormatCells      bool   `json:"format_cells"`
	FormatColumns    bool   `json:"format_columns"`
	FormatRows       bool   `json:"format_rows"`
	InsertColumns    bool   `json:"insert_columns"`
	InsertRows       bool   `json:"insert_rows"`
	InsertHyperlinks bool   `json:"insert_hyperlinks"`
	DeleteColumns    bool   `json:"delete_columns"`
	DeleteRows       bool   `json:"delete_rows"`
	Sort             bool   `json:"sort"`
	Autofilter       bool   `json:"autofilter"`
	PivotTables      bool   `json:"pivot_tables"`
	EditObjects      bool   `json:"edit_objects"`
	EditScenarios    bool   `json:"edit_scenarios"`
}
