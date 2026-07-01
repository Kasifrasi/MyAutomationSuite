package vorpruefung

import (
	"fmt"
	"strings"
	"shared/constants"
)

type ValidationList []string

// Zentral definierte Dropdown-Werte, damit die API sie nutzen kann
var (
	ListJaNein           ValidationList = []string{"Ja", "Nein"}
	ListAbzug            ValidationList = []string{"Abzug", "Kein Abzug"}
	ListWaehrung         ValidationList = []string{"EUR", "USD", "CHF"}
	ListMonate           ValidationList = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	ListKostenkategorien ValidationList = []string{"Bauausgaben", "Investitionen", "Personalkosten", "Projektaktivitaeten", "Projektverwaltung", "Evaluierung", "Audit", "Reserve"}
)

// --- Table Definitionen ---
type TableColumn struct {
	Header string
	Width  float64
	Format string
}

type TableField struct {
	tableName    string
	sheet        string
	Columns      []TableColumn
	HasTotalsRow bool
}

func (t TableField) TableName() string   { return t.tableName }
func (t TableField) Sheet() string       { return t.sheet }
func (t TableField) Cols() []TableColumn { return t.Columns }

func NewTableField(sheet string, baseName string, hasTotals bool, cols []TableColumn) TableField {
	return TableField{
		tableName:    "Tbl_" + baseName,
		sheet:        sheet,
		HasTotalsRow: hasTotals,
		Columns:      cols,
	}
}

// Factory für dynamische Tabellen (z.B. Ausgaben pro Periode im FB)
type TableFactory struct {
	Sheet        string
	Format       string
	Columns      []TableColumn
	HasTotalsRow bool
}

func (f TableFactory) Get(args ...any) TableField {
	tableName := fmt.Sprintf(f.Format, args...)
	return NewTableField(f.Sheet, tableName, f.HasTotalsRow, f.Columns)
}

// 1. Structs
type InputField struct {
	namedRange string
	sheet      string 
	Validation ValidationList
}

func (i InputField) NamedRange() string { return i.namedRange }
func (i InputField) Sheet() string      { return i.sheet }

type OutputField struct {
	namedRange string 
	sheet      string
}

func (o OutputField) NamedRange() string { return o.namedRange }
func (o OutputField) Sheet() string      { return o.sheet }

// 2. Constructors
func NewInputField(sheet string, baseName string, val ValidationList) InputField {
	namedRange := "Inp_" + baseName
	if strings.HasPrefix(baseName, "Inp_") {
		panic(fmt.Sprintf("[Developer Error] Bitte kein 'Inp_' mehr angeben. Das wird automatisch hinzugefügt für: %s", namedRange))
	}
	return InputField{namedRange: namedRange, sheet: sheet, Validation: val}
}

func NewOutputField(sheet string, baseName string) OutputField {
	namedRange := "Out_" + baseName
	if strings.HasPrefix(baseName, "Out_") {
		panic(fmt.Sprintf("[Developer Error] Bitte kein 'Out_' mehr angeben. Das wird automatisch hinzugefügt für: %s", namedRange))
	}
	return OutputField{namedRange: namedRange, sheet: sheet}
}

// 3. Factories
type InputFactory struct {
	Sheet  string
	Format string
	Val    ValidationList
}

func (f InputFactory) Get(args ...any) InputField {
	namedRange := fmt.Sprintf(f.Format, args...)
	return NewInputField(f.Sheet, namedRange, f.Val)
}

type OutputFactory struct {
	Sheet  string
	Format string
}

func (f OutputFactory) Get(args ...any) OutputField {
	namedRange := fmt.Sprintf(f.Format, args...)
	return NewOutputField(f.Sheet, namedRange)
}

// 4. SheetBuilder (Die geniale Erweiterung zur Gruppierung)
type SheetBuilder struct {
	Sheet  string
	Prefix string
}

func (b SheetBuilder) Inp(baseName string, val ValidationList) InputField {
	return NewInputField(b.Sheet, b.Prefix+baseName, val)
}

func (b SheetBuilder) Out(baseName string) OutputField {
	return NewOutputField(b.Sheet, b.Prefix+baseName)
}

func (b SheetBuilder) InpFact(format string, val ValidationList) InputFactory {
	return InputFactory{Sheet: b.Sheet, Format: b.Prefix + format, Val: val}
}

func (b SheetBuilder) OutFact(format string) OutputFactory {
	return OutputFactory{Sheet: b.Sheet, Format: b.Prefix + format}
}

// 4.1 Erweiterte Factory für Mittelanforderungen (Composition)
type MAInputFactory struct {
	InputFactory
	MaxPerioden int
	MaxSlots    int // Anzahl Mittelanforderungen pro Periode
}

// GetMA wandelt (Periode, Anforderung) in die fortlaufende TableID um (1 bis 54)
func (ma MAInputFactory) GetMA(periode int, anforderung int) InputField {
	if periode < 1 || periode > ma.MaxPerioden {
		panic(fmt.Sprintf("[Developer Error] Ungültige Periode %d. Maximum ist %d", periode, ma.MaxPerioden))
	}
	if anforderung < 1 || anforderung > ma.MaxSlots {
		panic(fmt.Sprintf("[Developer Error] Ungültige Anforderung %d. Maximum ist %d", anforderung, ma.MaxSlots))
	}
	
	// Excel-Tabellen-ID berechnen: (Periode - 1) + ((Anforderung - 1) * MaxPerioden) + 1
	tableId := (periode - 1) + ((anforderung - 1) * ma.MaxPerioden) + 1
	return ma.Get(tableId)
}

type MAOutputFactory struct {
	OutputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAOutputFactory) GetMA(periode int, anforderung int) OutputField {
	if periode < 1 || periode > ma.MaxPerioden {
		panic(fmt.Sprintf("[Developer Error] Ungültige Periode %d. Maximum ist %d", periode, ma.MaxPerioden))
	}
	if anforderung < 1 || anforderung > ma.MaxSlots {
		panic(fmt.Sprintf("[Developer Error] Ungültige Anforderung %d. Maximum ist %d", anforderung, ma.MaxSlots))
	}
	tableId := (periode - 1) + ((anforderung - 1) * ma.MaxPerioden) + 1
	return ma.Get(tableId)
}

// Spezielle Funktion für Felder in der MA, die NOCH eine Dimension haben (z.B. Kostenkategorie-Zeile)
type MAInputKatFactory struct {
	InputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAInputKatFactory) GetMA(periode int, anforderung int, rowIdx int) InputField {
    if periode < 1 || periode > ma.MaxPerioden {
		panic(fmt.Sprintf("[Developer Error] Ungültige Periode %d. Maximum ist %d", periode, ma.MaxPerioden))
	}
	if anforderung < 1 || anforderung > ma.MaxSlots {
		panic(fmt.Sprintf("[Developer Error] Ungültige Anforderung %d. Maximum ist %d", anforderung, ma.MaxSlots))
	}
	return ma.Get(periode, anforderung, rowIdx)
}

type MAOutputKatFactory struct {
	OutputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAOutputKatFactory) GetMA(periode int, anforderung int, rowIdx int) OutputField {
    if periode < 1 || periode > ma.MaxPerioden {
		panic(fmt.Sprintf("[Developer Error] Ungültige Periode %d. Maximum ist %d", periode, ma.MaxPerioden))
	}
	if anforderung < 1 || anforderung > ma.MaxSlots {
		panic(fmt.Sprintf("[Developer Error] Ungültige Anforderung %d. Maximum ist %d", anforderung, ma.MaxSlots))
	}
	return ma.Get(periode, anforderung, rowIdx)
}


// 4.2 Hilfsmethoden im SheetBuilder für die MA Factories
func (b SheetBuilder) MAInpFact(format string, val ValidationList, maxPerioden, maxSlots int) MAInputFactory {
	baseFact := InputFactory{Sheet: b.Sheet, Format: b.Prefix + format, Val: val}
	return MAInputFactory{InputFactory: baseFact, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAOutFact(format string, maxPerioden, maxSlots int) MAOutputFactory {
	baseFact := OutputFactory{Sheet: b.Sheet, Format: b.Prefix + format}
	return MAOutputFactory{OutputFactory: baseFact, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAInpKatFact(format string, val ValidationList, maxPerioden, maxSlots int) MAInputKatFactory {
	baseFact := InputFactory{Sheet: b.Sheet, Format: b.Prefix + format, Val: val}
	return MAInputKatFactory{InputFactory: baseFact, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAOutKatFact(format string, maxPerioden, maxSlots int) MAOutputKatFactory {
	baseFact := OutputFactory{Sheet: b.Sheet, Format: b.Prefix + format}
	return MAOutputKatFactory{OutputFactory: baseFact, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

// 5. Instanziierung der SheetBuilder für jedes Sheet
var (
	dash   = SheetBuilder{Sheet: constants.VPSheetDASHBOARD, Prefix: "Dash_"}
	budget = SheetBuilder{Sheet: constants.VPSheetBUDGET, Prefix: "Budget_"}
	kmw    = SheetBuilder{Sheet: constants.VPSheetKMW_MITTEL, Prefix: "KMW_"}
	fb     = SheetBuilder{Sheet: constants.VPSheetFINANZBERICHTE, Prefix: "FB_"}
	fbPrue = SheetBuilder{Sheet: constants.VPSheetFB_PRUEFUNG, Prefix: "FBPruef_"}
	ma     = SheetBuilder{Sheet: constants.VPSheetMA, Prefix: "MA_"}
	maPrue = SheetBuilder{Sheet: constants.VPSheetMA_PRUEFUNG, Prefix: "MAPruef_"}
)

// Registry aller Felder (statisch und dynamisch)
var (
	// ─────────────────────────────────────────────────────────────
	// Dashboard
	// ─────────────────────────────────────────────────────────────
	InputDashProjektnummer       = dash.Inp("Projektnummer", nil)
	InputDashVorprojekt          = dash.Inp("Vorprojekt", ListJaNein)
	InputDashProjekttitel        = dash.Inp("Projekttitel", nil)
	InputDashProjekttraeger      = dash.Inp("Projekttraeger", nil)
	InputDashBerichtswaehrung    = dash.Inp("Berichtswaehrung", nil)
	InputDashProjektstart        = dash.Inp("Projektstart", nil)
	InputDashProjektende         = dash.Inp("Projektende", nil)
	InputDashVPNummer            = dash.Inp("VPNummer", nil)
	InputDashVPBerichtswaehrung  = dash.Inp("VPBerichtswaehrung", nil)
	InputDashVPEnde              = dash.Inp("VPEnde", nil)
	InputDashVPWechselkurs       = dash.Inp("VPWechselkurs", nil)
	InputDashVPSaldoLC           = dash.Inp("VPSaldoLC", nil)
	InputDashVPSaldoEUR          = dash.Inp("VPSaldoEUR", nil)
	InputDashVPFolgeprojektstart = dash.Inp("VPFolgeprojektstart", nil)
	InputDashVPFolgeWechselkurs  = dash.Inp("VPFolgeWechselkurs", nil)
	InputDashVPFolgeSaldoLC      = dash.Inp("VPFolgeSaldoLC", nil)
	InputDashVPFolgeSaldoEUR     = dash.Inp("VPFolgeSaldoEUR", nil)
	
	// Dynamische Dashboard-Felder (Checkliste)
	InputDashChecklist           = dash.InpFact("Checklist_%d", ListJaNein)

	// ─────────────────────────────────────────────────────────────
	// I. Budget
	// ─────────────────────────────────────────────────────────────
	InputBudgetReserveFreigabe = budget.Inp("ReserveFreigabe", ListJaNein)
	InputBudgetDrittmittelY1   = budget.Inp("DrittmittelY1", nil)
	InputBudgetDrittmittelY2   = budget.Inp("DrittmittelY2", nil)
	InputBudgetDrittmittelY3   = budget.Inp("DrittmittelY3", nil)

	InputBudgetEigenmittelLC  = budget.Inp("EigenmittelLC", nil)
	InputBudgetEigenmittelY1  = budget.Inp("EigenmittelY1", nil)
	InputBudgetEigenmittelY2  = budget.Inp("EigenmittelY2", nil)
	InputBudgetEigenmittelY3  = budget.Inp("EigenmittelY3", nil)
	InputBudgetEigenmittelEUR = budget.Inp("EigenmittelEUR", nil)

	InputBudgetKMWLC  = budget.Inp("KMWLC", nil)
	InputBudgetKMWY1  = budget.Inp("KMWY1", nil)
	InputBudgetKMWY2  = budget.Inp("KMWY2", nil)
	InputBudgetKMWY3  = budget.Inp("KMWY3", nil)
	InputBudgetKMWEUR = budget.Inp("KMWEUR", nil)

	// ─────────────────────────────────────────────────────────────
	// II. KMW-Mittel
	// ─────────────────────────────────────────────────────────────
	InputKMWPeriode  = kmw.InpFact("Periode_%d", nil)
	InputKMWWaehrung = kmw.InpFact("Waehrung_%d", nil)
	InputKMWBetrag   = kmw.InpFact("Betrag_%d", nil)
	InputKMWDatum    = kmw.InpFact("Datum_%d", nil)

	// ─────────────────────────────────────────────────────────────
	// III. Finanzberichte
	// ─────────────────────────────────────────────────────────────
	InputFBVon              = fb.InpFact("Von_%d", nil)
	InputFBBis              = fb.InpFact("Bis_%d", nil)
	InputFBAufschlBank      = fb.InpFact("aufschl_Bank_%d", nil)
	InputFBAufschlKasse     = fb.InpFact("aufschl_Kasse_%d", nil)
	InputFBAufschlSonstiges = fb.InpFact("aufschl_Sonstiges_%d", nil)

	// ─────────────────────────────────────────────────────────────
	// Pruefung FB
	// ─────────────────────────────────────────────────────────────
	InputFBPruefungAuswahl    = fbPrue.Inp("Auswahl", nil)
	InputFBPruefungAbzugSaldo = fbPrue.Inp("AbzugSaldo", ListAbzug)
	InputFBPruefungAbzugMehr  = fbPrue.Inp("AbzugMehr", ListAbzug)

	// ─────────────────────────────────────────────────────────────
	// IV. MA
	// ─────────────────────────────────────────────────────────────
	// Input Inputs
	InputMAVon              = ma.InpFact("Von_%d", nil)
	InputMABis              = ma.InpFact("Bis_%d", nil)
	InputMAKurs             = ma.InpFact("Kurs_%d", nil)
	InputMAEigenmittelLC    = ma.InpFact("EigenmittelLC_%d", nil)
	InputMADrittmittelLC    = ma.InpFact("DrittmittelLC_%d", nil)
	InputMASaldoLC          = ma.InpFact("SaldoLC_%d", nil)
	InputMAManBetrag        = ma.InpFact("ManBetrag_%d", nil)
	InputMAKat              = ma.InpFact("Kat_%d_%d_%d", nil)
	InputMAKmwLC            = ma.InpFact("KmwLC_%d_%d", nil)

	// Output Inputs
	OutputMAPeriode          = ma.OutFact("Periode_%d")
	OutputMAZeitraum         = ma.OutFact("Zeitraum_%d")
	OutputMASumLC            = ma.OutFact("SumLC_%d")
	OutputMASumEUR           = ma.OutFact("SumEUR_%d")
	OutputMAEigenmittelEUR   = ma.OutFact("EigenmittelEUR_%d")
	OutputMADrittmittelEUR   = ma.OutFact("DrittmittelEUR_%d")
	OutputMAKatEUR           = ma.OutFact("KatEUR_%d_%d_%d")
	OutputMAKmwEUR           = ma.OutFact("KmwEUR_%d_%d")

	// ─────────────────────────────────────────────────────────────
	// Pruefung MA
	// ─────────────────────────────────────────────────────────────
	InputMAPruefungAuswahl       = maPrue.Inp("Auswahl", nil)
	InputMAPruefungAbzugSaldo    = maPrue.Inp("AbzugSaldo", ListAbzug)
	InputMAPruefungAbzugMehr     = maPrue.Inp("AbzugMehr", ListAbzug)
	InputMAPruefungAbzugPrognose = maPrue.Inp("AbzugPrognose", ListAbzug)
	InputMAPruefungMonateY1      = maPrue.Inp("MonateY1", ListMonate)
	InputMAPruefungMonateY2      = maPrue.Inp("MonateY2", ListMonate)
	InputMAPruefungMonateY3      = maPrue.Inp("MonateY3", ListMonate)
)