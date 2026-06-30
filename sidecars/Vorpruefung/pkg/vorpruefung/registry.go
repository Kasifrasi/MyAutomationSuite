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
	FieldDashProjektnummer       = dash.Inp("Projektnummer", nil)
	FieldDashVorprojekt          = dash.Inp("Vorprojekt", ListJaNein)
	FieldDashProjekttitel        = dash.Inp("Projekttitel", nil)
	FieldDashProjekttraeger      = dash.Inp("Projekttraeger", nil)
	FieldDashBerichtswaehrung    = dash.Inp("Berichtswaehrung", nil)
	FieldDashProjektstart        = dash.Inp("Projektstart", nil)
	FieldDashProjektende         = dash.Inp("Projektende", nil)
	FieldDashVPNummer            = dash.Inp("VPNummer", nil)
	FieldDashVPBerichtswaehrung  = dash.Inp("VPBerichtswaehrung", nil)
	FieldDashVPEnde              = dash.Inp("VPEnde", nil)
	FieldDashVPWechselkurs       = dash.Inp("VPWechselkurs", nil)
	FieldDashVPSaldoLC           = dash.Inp("VPSaldoLC", nil)
	FieldDashVPSaldoEUR          = dash.Inp("VPSaldoEUR", nil)
	FieldDashVPFolgeprojektstart = dash.Inp("VPFolgeprojektstart", nil)
	FieldDashVPFolgeWechselkurs  = dash.Inp("VPFolgeWechselkurs", nil)
	FieldDashVPFolgeSaldoLC      = dash.Inp("VPFolgeSaldoLC", nil)
	FieldDashVPFolgeSaldoEUR     = dash.Inp("VPFolgeSaldoEUR", nil)
	
	// Dynamische Dashboard-Felder (Checkliste)
	FieldDashChecklist           = dash.InpFact("Checklist_%d", ListJaNein)

	// ─────────────────────────────────────────────────────────────
	// I. Budget
	// ─────────────────────────────────────────────────────────────
	FieldBudgetReserveFreigabe = budget.Inp("ReserveFreigabe", ListJaNein)
	FieldBudgetDrittmittelY1   = budget.Inp("DrittmittelY1", nil)
	FieldBudgetDrittmittelY2   = budget.Inp("DrittmittelY2", nil)
	FieldBudgetDrittmittelY3   = budget.Inp("DrittmittelY3", nil)

	FieldBudgetEigenmittelLC  = budget.Inp("EigenmittelLC", nil)
	FieldBudgetEigenmittelY1  = budget.Inp("EigenmittelY1", nil)
	FieldBudgetEigenmittelY2  = budget.Inp("EigenmittelY2", nil)
	FieldBudgetEigenmittelY3  = budget.Inp("EigenmittelY3", nil)
	FieldBudgetEigenmittelEUR = budget.Inp("EigenmittelEUR", nil)

	FieldBudgetKMWLC  = budget.Inp("KMWLC", nil)
	FieldBudgetKMWY1  = budget.Inp("KMWY1", nil)
	FieldBudgetKMWY2  = budget.Inp("KMWY2", nil)
	FieldBudgetKMWY3  = budget.Inp("KMWY3", nil)
	FieldBudgetKMWEUR = budget.Inp("KMWEUR", nil)

	// ─────────────────────────────────────────────────────────────
	// II. KMW-Mittel
	// ─────────────────────────────────────────────────────────────
	FieldKMWPeriode  = kmw.InpFact("Periode_%d", nil)
	FieldKMWWaehrung = kmw.InpFact("Waehrung_%d", nil)
	FieldKMWBetrag   = kmw.InpFact("Betrag_%d", nil)
	FieldKMWDatum    = kmw.InpFact("Datum_%d", nil)

	// ─────────────────────────────────────────────────────────────
	// III. Finanzberichte
	// ─────────────────────────────────────────────────────────────
	FieldFBVon              = fb.InpFact("Von_%d", nil)
	FieldFBBis              = fb.InpFact("Bis_%d", nil)
	FieldFBAufschlBank      = fb.InpFact("aufschl_Bank_%d", nil)
	FieldFBAufschlKasse     = fb.InpFact("aufschl_Kasse_%d", nil)
	FieldFBAufschlSonstiges = fb.InpFact("aufschl_Sonstiges_%d", nil)

	// ─────────────────────────────────────────────────────────────
	// Pruefung FB
	// ─────────────────────────────────────────────────────────────
	FieldFBPruefungAuswahl    = fbPrue.Inp("Auswahl", nil)
	FieldFBPruefungAbzugSaldo = fbPrue.Inp("AbzugSaldo", ListAbzug)
	FieldFBPruefungAbzugMehr  = fbPrue.Inp("AbzugMehr", ListAbzug)

	// ─────────────────────────────────────────────────────────────
	// IV. MA
	// ─────────────────────────────────────────────────────────────
	// Input Fields
	FieldMAVon              = ma.InpFact("Von_%d", nil)
	FieldMABis              = ma.InpFact("Bis_%d", nil)
	FieldMAKurs             = ma.InpFact("Kurs_%d", nil)
	FieldMAEigenmittelLC    = ma.InpFact("EigenmittelLC_%d", nil)
	FieldMADrittmittelLC    = ma.InpFact("DrittmittelLC_%d", nil)
	FieldMASaldoLC          = ma.InpFact("SaldoLC_%d", nil)
	FieldMAManBetrag        = ma.InpFact("ManBetrag_%d", nil)
	FieldMAKat              = ma.InpFact("Kat_%d_%d_%d", nil)
	FieldMAKmwLC            = ma.InpFact("KmwLC_%d_%d", nil)

	// Output Fields
	FieldMAPeriode          = ma.OutFact("Periode_%d")
	FieldMAZeitraum         = ma.OutFact("Zeitraum_%d")
	FieldMASumLC            = ma.OutFact("SumLC_%d")
	FieldMASumEUR           = ma.OutFact("SumEUR_%d")
	FieldMAEigenmittelEUR   = ma.OutFact("EigenmittelEUR_%d")
	FieldMADrittmittelEUR   = ma.OutFact("DrittmittelEUR_%d")
	FieldMAKatEUR           = ma.OutFact("KatEUR_%d_%d_%d")
	FieldMAKmwEUR           = ma.OutFact("KmwEUR_%d_%d")

	// ─────────────────────────────────────────────────────────────
	// Pruefung MA
	// ─────────────────────────────────────────────────────────────
	FieldMAPruefungAuswahl       = maPrue.Inp("Auswahl", nil)
	FieldMAPruefungAbzugSaldo    = maPrue.Inp("AbzugSaldo", ListAbzug)
	FieldMAPruefungAbzugMehr     = maPrue.Inp("AbzugMehr", ListAbzug)
	FieldMAPruefungAbzugPrognose = maPrue.Inp("AbzugPrognose", ListAbzug)
	FieldMAPruefungMonateY1      = maPrue.Inp("MonateY1", ListMonate)
	FieldMAPruefungMonateY2      = maPrue.Inp("MonateY2", ListMonate)
	FieldMAPruefungMonateY3      = maPrue.Inp("MonateY3", ListMonate)
)