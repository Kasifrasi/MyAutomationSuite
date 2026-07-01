package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"
)

type ValidationList []string

// Zentral definierte Dropdown-Werte, damit die API sie nutzen kann
var (
	ListJaNein           ValidationList = []string{"Ja", "Nein"}
	ListAbzug            ValidationList = []string{"Abzug", "Kein Abzug"}
	ListWaehrung         ValidationList = []string{
		"AED", "AFN", "ALL", "AMD", "AOA", "ARS", "AUD", "AWG", "AZN", "BAM", "BBD",
		"BDT", "BHD", "BIF", "BMD", "BND", "BOB", "BRL", "BSD", "BTN", "BWP", "BYN",
		"BZD", "CAD", "CDF", "CHF", "CLP", "CNY", "COP", "CRC", "CUP", "CVE", "CZK",
		"DJF", "DKK", "DOP", "DZD", "EGP", "ERN", "ETB", "EUR", "FJD", "FKP", "GBP",
		"GEL", "GHS", "GIP", "GMD", "GNF", "GTQ", "GYD", "HKD", "HNL", "HTG", "HUF",
		"IDR", "ILS", "INR", "IQD", "IRR", "ISK", "JMD", "JOD", "JPY", "KES", "KGS",
		"KHR", "KMF", "KPW", "KRW", "KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD",
		"LSL", "LYD", "MAD", "MDL", "MGA", "MKD", "MMK", "MNT", "MOP", "MRU", "MUR",
		"MVR", "MWK", "MXN", "MYR", "MZN", "NAD", "NGN", "NIO", "NOK", "NPR", "NZD",
		"OMR", "PAB", "PEN", "PGK", "PHP", "PKR", "PLN", "PYG", "QAR", "RON", "RSD",
		"RUB", "RWF", "SAR", "SBD", "SCR", "SDG", "SEK", "SGD", "SHP", "SLE", "SOS",
		"SRD", "SSP", "STN", "SVC", "SYP", "SZL", "THB", "TJS", "TMT", "TND", "TOP",
		"TRY", "TTD", "TWD", "TZS", "UAH", "UGX", "USD", "UYU", "UZS", "VED", "VES",
		"VND", "VUV", "WST", "XAF", "XCD", "XCG", "XOF", "XPF", "YER", "ZAR", "ZMW",
		"ZWG",
	}   
	ListMonate           ValidationList = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	ListKostenkategorien ValidationList = []string{"Bauausgaben", "Investitionen", "Personalkosten", "Projektaktivitaeten", "Projektverwaltung", "Evaluierung", "Audit", "Reserve"}
)

// ─────────────────────────────────────────────────────────────
// Gemeinsames Interface
// ─────────────────────────────────────────────────────────────

type ExcelElement interface {
	GetName() string
	GetSheet() string
}

// ─────────────────────────────────────────────────────────────
// Table Definitionen
// ─────────────────────────────────────────────────────────────

type TableColumn struct {
	Header string
	Width  float64
	Format string
}

type TableField struct {
	Name         string
	Sheet        string
	Columns      []TableColumn
	HasTotalsRow bool
}

func (t TableField) GetName() string  { return t.Name }
func (t TableField) GetSheet() string { return t.Sheet }

func NewTableField(sheet string, baseName string, hasTotals bool, cols []TableColumn) TableField {
	return TableField{
		Name:         "Tbl_" + baseName,
		Sheet:        sheet,
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

func (f TableFactory) Get(args ...int) TableField {
	anyArgs := make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}
	tableName := fmt.Sprintf(f.Format, anyArgs...)
	return NewTableField(f.Sheet, tableName, f.HasTotalsRow, f.Columns)
}

// ─────────────────────────────────────────────────────────────
// 1. Structs
// ─────────────────────────────────────────────────────────────

type InputField struct {
	NamedRange string
	Sheet      string
	Validation ValidationList
}

func (i InputField) GetName() string  { return i.NamedRange }
func (i InputField) GetSheet() string { return i.Sheet }

type OutputField struct {
	NamedRange string
	Sheet      string
}

func (o OutputField) GetName() string  { return o.NamedRange }
func (o OutputField) GetSheet() string { return o.Sheet }

// ─────────────────────────────────────────────────────────────
// 2. Constructors
// ─────────────────────────────────────────────────────────────

func NewInputField(sheet string, baseName string, val ValidationList) InputField {
	namedRange := "Inp_" + baseName
	if strings.HasPrefix(baseName, "Inp_") {
		// Panic beim Start ist ok, da es ein statischer Entwicklerfehler bei der Konfiguration ist
		panic(fmt.Sprintf("[Developer Error] Bitte kein 'Inp_' mehr angeben. Das wird automatisch hinzugefügt für: %s", namedRange))
	}
	return InputField{NamedRange: namedRange, Sheet: sheet, Validation: val}
}

func NewOutputField(sheet string, baseName string) OutputField {
	namedRange := "Out_" + baseName
	if strings.HasPrefix(baseName, "Out_") {
		panic(fmt.Sprintf("[Developer Error] Bitte kein 'Out_' mehr angeben. Das wird automatisch hinzugefügt für: %s", namedRange))
	}
	return OutputField{NamedRange: namedRange, Sheet: sheet}
}

// ─────────────────────────────────────────────────────────────
// 3. Factories
// ─────────────────────────────────────────────────────────────

type InputFactory struct {
	Sheet  string
	Format string
	Val    ValidationList
}

func (f InputFactory) Get(args ...int) InputField {
	anyArgs := make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}
	namedRange := fmt.Sprintf(f.Format, anyArgs...)
	return NewInputField(f.Sheet, namedRange, f.Val)
}

type OutputFactory struct {
	Sheet  string
	Format string
}

func (f OutputFactory) Get(args ...int) OutputField {
	anyArgs := make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}
	namedRange := fmt.Sprintf(f.Format, anyArgs...)
	return NewOutputField(f.Sheet, namedRange)
}

// ─────────────────────────────────────────────────────────────
// 4. SheetBuilder
// ─────────────────────────────────────────────────────────────

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

// ─────────────────────────────────────────────────────────────
// 4.1 Erweiterte Factory für Mittelanforderungen
// ─────────────────────────────────────────────────────────────

func validatePeriodeAnforderung(periode, anforderung, maxPerioden, maxSlots int) error {
	if periode < 1 || periode > maxPerioden {
		return fmt.Errorf("ungültige Periode %d. Maximum ist %d", periode, maxPerioden)
	}
	if anforderung < 1 || anforderung > maxSlots {
		return fmt.Errorf("ungültige Anforderung %d. Maximum ist %d", anforderung, maxSlots)
	}
	return nil
}

func calculateTableID(periode, anforderung, maxPerioden, maxSlots int) (int, error) {
	if err := validatePeriodeAnforderung(periode, anforderung, maxPerioden, maxSlots); err != nil {
		return 0, err
	}
	return (periode - 1) + ((anforderung - 1) * maxPerioden) + 1, nil
}

type MAInputFactory struct {
	InputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAInputFactory) GetMA(periode int, anforderung int) (InputField, error) {
	tableId, err := calculateTableID(periode, anforderung, ma.MaxPerioden, ma.MaxSlots)
	if err != nil {
		return InputField{}, err
	}
	return ma.Get(tableId), nil
}

type MAOutputFactory struct {
	OutputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAOutputFactory) GetMA(periode int, anforderung int) (OutputField, error) {
	tableId, err := calculateTableID(periode, anforderung, ma.MaxPerioden, ma.MaxSlots)
	if err != nil {
		return OutputField{}, err
	}
	return ma.Get(tableId), nil
}

type MAInputKatFactory struct {
	InputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAInputKatFactory) GetMA(periode int, anforderung int, rowIdx int) (InputField, error) {
	if err := validatePeriodeAnforderung(periode, anforderung, ma.MaxPerioden, ma.MaxSlots); err != nil {
		return InputField{}, err
	}
	return ma.Get(periode, anforderung, rowIdx), nil
}

type MAOutputKatFactory struct {
	OutputFactory
	MaxPerioden int
	MaxSlots    int
}

func (ma MAOutputKatFactory) GetMA(periode int, anforderung int, rowIdx int) (OutputField, error) {
	if err := validatePeriodeAnforderung(periode, anforderung, ma.MaxPerioden, ma.MaxSlots); err != nil {
		return OutputField{}, err
	}
	return ma.Get(periode, anforderung, rowIdx), nil
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

// ─────────────────────────────────────────────────────────────
// 5. Instanz-Registry aller Felder
// ─────────────────────────────────────────────────────────────

type TemplateRegistry struct {
	// Budget-Tabellen
	TableBudgetAusgaben    TableField
	TableBudgetDrittmittel TableField

	// Dashboard
	InputDashProjektnummer       InputField
	InputDashVorprojekt          InputField
	InputDashProjekttitel        InputField
	InputDashProjekttraeger      InputField
	InputDashBerichtswaehrung    InputField
	InputDashProjektstart        InputField
	InputDashProjektende         InputField
	InputDashVPNummer            InputField
	InputDashVPBerichtswaehrung  InputField
	InputDashVPEnde              InputField
	InputDashVPWechselkurs       InputField
	InputDashVPSaldoLC           InputField
	InputDashVPSaldoEUR          InputField
	InputDashVPFolgeprojektstart InputField
	InputDashVPFolgeWechselkurs  InputField
	InputDashVPFolgeSaldoLC      InputField
	InputDashVPFolgeSaldoEUR     InputField
	InputDashVPSaldoCheck        InputField
	InputDashVertragCheck        InputField
	InputDashBudgetCheck         InputField
	InputDashBankBelegeCheck     InputField
	InputDashFBCheck             InputField
	InputDashMACheck             InputField
	

	OutputDashProjektlaufzeit    OutputField
	OutputDashMonate             OutputField
	OutputDashSaldoEUR           OutputField
	OutputDashSaldovortragEUR    OutputField

	// Budget
	OutputBudgetWK             OutputField
	OutputBudgetReserveEUR     OutputField
	InputBudgetReserveFreigabe InputField
	InputBudgetBegruendung     InputField
	InputBudgetEigenmittelLC   InputField
	InputBudgetEigenmittelY1   InputField
	InputBudgetEigenmittelY2   InputField
	InputBudgetEigenmittelY3   InputField
	InputBudgetEigenmittelEUR  InputField
	OutputBudgetDrittmittelLC  OutputField
	InputBudgetDrittmittelY1   InputField
	InputBudgetDrittmittelY2   InputField
	InputBudgetDrittmittelY3   InputField
	OutputBudgetDrittmittelEUR OutputField
	InputBudgetKMWLC           InputField
	InputBudgetKMWY1           InputField
	InputBudgetKMWY2           InputField
	InputBudgetKMWY3           InputField
	InputBudgetKMWEUR          InputField
	OutputBudgetGesamtLC       OutputField
	OutputBudgetGesamtY1       OutputField
	OutputBudgetGesamtY2       OutputField
	OutputBudgetGesamtY3       OutputField
	OutputBudgetGesamtEUR      OutputField


	// KMW-Mittel
	InputKMWPeriode  InputFactory
	InputKMWWaehrung InputFactory
	InputKMWBetrag   InputFactory
	InputKMWDatum    InputFactory

	// Finanzberichte
	InputFBVon              InputFactory
	InputFBBis              InputFactory
	InputFBAufschlBank      InputFactory
	InputFBAufschlKasse     InputFactory
	InputFBAufschlSonstiges InputFactory

	// Pruefung FB
	InputFBPruefungAuswahl    InputField
	InputFBPruefungAbzugSaldo InputField
	InputFBPruefungAbzugMehr  InputField

	// MA
	InputMAVon           InputFactory
	InputMABis           InputFactory
	InputMAKurs          InputFactory
	InputMAEigenmittelLC InputFactory
	InputMADrittmittelLC InputFactory
	InputMASaldoLC       InputFactory
	InputMAManBetrag     InputFactory
	InputMAKat           InputFactory
	InputMAKmwLC         InputFactory

	OutputMAPeriode        OutputFactory
	OutputMAZeitraum       OutputFactory
	OutputMASumLC          OutputFactory
	OutputMASumEUR         OutputFactory
	OutputMAEigenmittelEUR OutputFactory
	OutputMADrittmittelEUR OutputFactory
	OutputMAKatEUR         OutputFactory
	OutputMAKmwEUR         OutputFactory

	// Pruefung MA
	InputMAPruefungAuswahl       InputField
	InputMAPruefungAbzugSaldo    InputField
	InputMAPruefungAbzugMehr     InputField
	InputMAPruefungAbzugPrognose InputField
	InputMAPruefungMonateY1      InputField
	InputMAPruefungMonateY2      InputField
	InputMAPruefungMonateY3      InputField
}

// ─────────────────────────────────────────────────────────────
// 6. Paket-weite Field-Accessoren (für Generator & API)
// ─────────────────────────────────────────────────────────────

var Registry = NewTemplateRegistry()

func FieldDashChecklist(i int) InputField {
	checks := [6]InputField{
		Registry.InputDashVPSaldoCheck,
		Registry.InputDashVertragCheck,
		Registry.InputDashBudgetCheck,
		Registry.InputDashBankBelegeCheck,
		Registry.InputDashFBCheck,
		Registry.InputDashMACheck,
	}
	if i < 1 || i > len(checks) {
		panic(fmt.Sprintf("FieldDashChecklist: ungültiger Index %d (1-%d)", i, len(checks)))
	}
	return checks[i-1]
}

// KMW
func FieldKMWPeriode(i int) InputField  { return Registry.InputKMWPeriode.Get(i) }
func FieldKMWWaehrung(i int) InputField { return Registry.InputKMWWaehrung.Get(i) }
func FieldKMWBetrag(i int) InputField   { return Registry.InputKMWBetrag.Get(i) }
func FieldKMWDatum(i int) InputField    { return Registry.InputKMWDatum.Get(i) }

// Finanzberichte
func FieldFBVon(i int) InputField              { return Registry.InputFBVon.Get(i) }
func FieldFBBis(i int) InputField              { return Registry.InputFBBis.Get(i) }
func FieldFBAufschlBank(i int) InputField      { return Registry.InputFBAufschlBank.Get(i) }
func FieldFBAufschlKasse(i int) InputField     { return Registry.InputFBAufschlKasse.Get(i) }
func FieldFBAufschlSonstiges(i int) InputField { return Registry.InputFBAufschlSonstiges.Get(i) }

// MA (Input)
func FieldMAVon(i int) InputField           { return Registry.InputMAVon.Get(i) }
func FieldMABis(i int) InputField           { return Registry.InputMABis.Get(i) }
func FieldMAKurs(i int) InputField          { return Registry.InputMAKurs.Get(i) }
func FieldMAEigenmittelLC(i int) InputField { return Registry.InputMAEigenmittelLC.Get(i) }
func FieldMADrittmittelLC(i int) InputField { return Registry.InputMADrittmittelLC.Get(i) }
func FieldMASaldoLC(i int) InputField       { return Registry.InputMASaldoLC.Get(i) }
func FieldMAManBetrag(i int) InputField     { return Registry.InputMAManBetrag.Get(i) }
func FieldMAKat(j, k int) InputField        { return Registry.InputMAKat.Get(j, k) }
func FieldMAKmwLC(i int) InputField         { return Registry.InputMAKmwLC.Get(i) }

// MA (Output → liefert Named-Range-String für Formeln)
func FieldMAPeriode(i int) string        { return Registry.OutputMAPeriode.Get(i).NamedRange }
func FieldMAZeitraum(i int) string       { return Registry.OutputMAZeitraum.Get(i).NamedRange }
func FieldMASumLC(i int) string          { return Registry.OutputMASumLC.Get(i).NamedRange }
func FieldMASumEUR(i int) string         { return Registry.OutputMASumEUR.Get(i).NamedRange }
func FieldMAEigenmittelEUR(i int) string { return Registry.OutputMAEigenmittelEUR.Get(i).NamedRange }
func FieldMADrittmittelEUR(i int) string { return Registry.OutputMADrittmittelEUR.Get(i).NamedRange }
func FieldMAKatEUR(j, k int) string      { return Registry.OutputMAKatEUR.Get(j, k).NamedRange }
func FieldMAKmwEUR(i int) string         { return Registry.OutputMAKmwEUR.Get(i).NamedRange }

// Paket-weite Variablen für api.go (Flat-Accessor-Aliases → Registry)

// Dashboard
var (
	FieldDashProjektnummer       = Registry.InputDashProjektnummer
	FieldDashVorprojekt          = Registry.InputDashVorprojekt
	FieldDashProjekttitel        = Registry.InputDashProjekttitel
	FieldDashProjekttraeger      = Registry.InputDashProjekttraeger
	FieldDashBerichtswaehrung    = Registry.InputDashBerichtswaehrung
	FieldDashProjektstart        = Registry.InputDashProjektstart
	FieldDashProjektende         = Registry.InputDashProjektende
	FieldDashVPNummer            = Registry.InputDashVPNummer
	FieldDashVPBerichtswaehrung  = Registry.InputDashVPBerichtswaehrung
	FieldDashVPEnde              = Registry.InputDashVPEnde
	FieldDashVPWechselkurs       = Registry.InputDashVPWechselkurs
	FieldDashVPSaldoLC           = Registry.InputDashVPSaldoLC
	FieldDashVPSaldoEUR          = Registry.InputDashVPSaldoEUR
	FieldDashVPFolgeprojektstart = Registry.InputDashVPFolgeprojektstart
	FieldDashVPFolgeWechselkurs  = Registry.InputDashVPFolgeWechselkurs
	FieldDashVPFolgeSaldoLC      = Registry.InputDashVPFolgeSaldoLC
	FieldDashVPFolgeSaldoEUR     = Registry.InputDashVPFolgeSaldoEUR
)

// Budget
var (
	FieldBudgetReserveFreigabe = Registry.InputBudgetReserveFreigabe
	FieldBudgetEigenmittelLC   = Registry.InputBudgetEigenmittelLC
	FieldBudgetEigenmittelY1   = Registry.InputBudgetEigenmittelY1
	FieldBudgetEigenmittelY2   = Registry.InputBudgetEigenmittelY2
	FieldBudgetEigenmittelY3   = Registry.InputBudgetEigenmittelY3
	FieldBudgetEigenmittelEUR  = Registry.InputBudgetEigenmittelEUR
	FieldBudgetDrittmittelY1   = Registry.InputBudgetDrittmittelY1
	FieldBudgetDrittmittelY2   = Registry.InputBudgetDrittmittelY2
	FieldBudgetDrittmittelY3   = Registry.InputBudgetDrittmittelY3
	FieldBudgetKMWLC           = Registry.InputBudgetKMWLC
	FieldBudgetKMWY1           = Registry.InputBudgetKMWY1
	FieldBudgetKMWY2           = Registry.InputBudgetKMWY2
	FieldBudgetKMWY3           = Registry.InputBudgetKMWY3
	FieldBudgetKMWEUR          = Registry.InputBudgetKMWEUR
)

// Prüfung FB
var (
	FieldFBPruefungAuswahl    = Registry.InputFBPruefungAuswahl
	FieldFBPruefungAbzugSaldo = Registry.InputFBPruefungAbzugSaldo
	FieldFBPruefungAbzugMehr  = Registry.InputFBPruefungAbzugMehr
)

// Prüfung MA
var (
	FieldMAPruefungAuswahl       = Registry.InputMAPruefungAuswahl
	FieldMAPruefungAbzugSaldo    = Registry.InputMAPruefungAbzugSaldo
	FieldMAPruefungAbzugMehr     = Registry.InputMAPruefungAbzugMehr
	FieldMAPruefungAbzugPrognose = Registry.InputMAPruefungAbzugPrognose
	FieldMAPruefungMonateY1      = Registry.InputMAPruefungMonateY1
	FieldMAPruefungMonateY2      = Registry.InputMAPruefungMonateY2
	FieldMAPruefungMonateY3      = Registry.InputMAPruefungMonateY3
)

func NewTemplateRegistry() *TemplateRegistry {
	dash := SheetBuilder{Sheet: constants.VPSheetDASHBOARD, Prefix: "Dash_"}
	budget := SheetBuilder{Sheet: constants.VPSheetBUDGET, Prefix: "Budget_"}
	kmw := SheetBuilder{Sheet: constants.VPSheetKMW_MITTEL, Prefix: "KMW_"}
	fb := SheetBuilder{Sheet: constants.VPSheetFINANZBERICHTE, Prefix: "FB_"}
	fbPrue := SheetBuilder{Sheet: constants.VPSheetFB_PRUEFUNG, Prefix: "FBPruef_"}
	ma := SheetBuilder{Sheet: constants.VPSheetMA, Prefix: "MA_"}
	maPrue := SheetBuilder{Sheet: constants.VPSheetMA_PRUEFUNG, Prefix: "MAPruef_"}

	return &TemplateRegistry{
		// Budget-Tabellen
		TableBudgetAusgaben: TableField{
			Name:         "TblBudgetAusgaben",
			Sheet:        constants.VPSheetBUDGET,
			HasTotalsRow: true,
			Columns: []TableColumn{
				{Header: "Kostenkategorie"},
				{Header: "ID"},
				{Header: "Kostenposition"},
				{Header: "Betrag (LC)", Format: "#,##0.00"},
				{Header: "Jahr 1", Format: "#,##0.00"},
				{Header: "Jahr 2", Format: "#,##0.00"},
				{Header: "Jahr 3", Format: "#,##0.00"},
				{Header: "Betrag (EUR)", Format: `#,##0.00" €"`},
			},
		},
		TableBudgetDrittmittel: TableField{
			Name:         "TblDrittmittel",
			Sheet:        constants.VPSheetBUDGET,
			HasTotalsRow: false,
			Columns: []TableColumn{
				{Header: "Name des Gebers"},
				{Header: "Betrag (LC)", Format: "#,##0.00"},
				{Header: "Betrag (EUR)", Format: `#,##0.00" €"`},
			},
		},

		// Dashboard
		InputDashProjektnummer:       dash.Inp("Projektnummer", nil),
		InputDashVorprojekt:          dash.Inp("Vorprojekt", ListJaNein),
		InputDashProjekttitel:        dash.Inp("Projekttitel", nil),
		InputDashProjekttraeger:      dash.Inp("Projekttraeger", nil),
		InputDashBerichtswaehrung:    dash.Inp("Berichtswaehrung", nil),
		InputDashProjektstart:        dash.Inp("Projektstart", nil),
		InputDashProjektende:         dash.Inp("Projektende", nil),
		InputDashVPNummer:            dash.Inp("VPNummer", nil),
		InputDashVPBerichtswaehrung:  dash.Inp("VPBerichtswaehrung", nil),
		InputDashVPEnde:              dash.Inp("VPEnde", nil),
		InputDashVPWechselkurs:       dash.Inp("VPWechselkurs", nil),
		InputDashVPSaldoLC:           dash.Inp("VPSaldoLC", nil),
		InputDashVPSaldoEUR:          dash.Inp("VPSaldoEUR", nil),
		InputDashVPFolgeprojektstart: dash.Inp("VPFolgeprojektstart", nil),
		InputDashVPFolgeWechselkurs:  dash.Inp("VPFolgeWechselkurs", nil),
		InputDashVPFolgeSaldoLC:      dash.Inp("VPFolgeSaldoLC", nil),
		InputDashVPFolgeSaldoEUR:     dash.Inp("VPFolgeSaldoEUR", ListJaNein),
		InputDashVPSaldoCheck:        dash.Inp("VPSaldoCheck", ListJaNein),
		InputDashVertragCheck:        dash.Inp("VertragCheck", ListJaNein),
		InputDashBudgetCheck:         dash.Inp("BudgetCheck", ListJaNein),
		InputDashBankBelegeCheck:     dash.Inp("BankBelegeCheck", ListJaNein),
		InputDashFBCheck:             dash.Inp("FBCheck", ListJaNein),
		InputDashMACheck:             dash.Inp("MACheck", ListJaNein),

		OutputDashProjektlaufzeit:     dash.Out("Projektlaufzeit"),
		OutputDashMonate:              dash.Out("Monate"),
		OutputDashSaldoEUR:            dash.Out("SaldoEUR"),
		OutputDashSaldovortragEUR:     dash.Out("SaldovortragEUR"),

		// Budget
		OutputBudgetWK:             budget.Out("WK"),
		OutputBudgetReserveEUR:     budget.Out("ReserveEUR"),
		InputBudgetReserveFreigabe: budget.Inp("ReserveFreigabe", ListJaNein),
		InputBudgetBegruendung:     budget.Inp("Begruendung", nil),
		InputBudgetEigenmittelLC:   budget.Inp("EigenmittelLC", nil),
		InputBudgetEigenmittelY1:   budget.Inp("EigenmittelY1", nil),
		InputBudgetEigenmittelY2:   budget.Inp("EigenmittelY2", nil),
		InputBudgetEigenmittelY3:   budget.Inp("EigenmittelY3", nil),
		InputBudgetEigenmittelEUR:  budget.Inp("EigenmittelEUR", nil),
		OutputBudgetDrittmittelLC:  budget.Out("DrittmittelLC"),
		InputBudgetDrittmittelY1:   budget.Inp("DrittmittelY1", nil),
		InputBudgetDrittmittelY2:   budget.Inp("DrittmittelY2", nil),
		InputBudgetDrittmittelY3:   budget.Inp("DrittmittelY3", nil),
		OutputBudgetDrittmittelEUR: budget.Out("DrittmittelEUR"),
		InputBudgetKMWLC:           budget.Inp("KMWLC", nil),
		InputBudgetKMWY1:           budget.Inp("KMWY1", nil),
		InputBudgetKMWY2:           budget.Inp("KMWY2", nil),
		InputBudgetKMWY3:           budget.Inp("KMWY3", nil),
		InputBudgetKMWEUR:          budget.Inp("KMWEUR", nil),
		OutputBudgetGesamtLC:       budget.Out("GesamtLC"),
		OutputBudgetGesamtY1:       budget.Out("GesamtY1"),
		OutputBudgetGesamtY2:       budget.Out("GesamtY2"),
		OutputBudgetGesamtY3:       budget.Out("GesamtY3"),
		OutputBudgetGesamtEUR:      budget.Out("GesamtEUR"),

		// KMW-Mittel
		InputKMWPeriode:  kmw.InpFact("Periode_%d", nil),
		InputKMWWaehrung: kmw.InpFact("Waehrung_%d", nil),
		InputKMWBetrag:   kmw.InpFact("Betrag_%d", nil),
		InputKMWDatum:    kmw.InpFact("Datum_%d", nil),

		// Finanzberichte
		InputFBVon:              fb.InpFact("Von_%d", nil),
		InputFBBis:              fb.InpFact("Bis_%d", nil),
		InputFBAufschlBank:      fb.InpFact("aufschl_Bank_%d", nil),
		InputFBAufschlKasse:     fb.InpFact("aufschl_Kasse_%d", nil),
		InputFBAufschlSonstiges: fb.InpFact("aufschl_Sonstiges_%d", nil),

		// Pruefung FB
		InputFBPruefungAuswahl:    fbPrue.Inp("Auswahl", nil),
		InputFBPruefungAbzugSaldo: fbPrue.Inp("AbzugSaldo", ListAbzug),
		InputFBPruefungAbzugMehr:  fbPrue.Inp("AbzugMehr", ListAbzug),

		// MA
		InputMAVon:           ma.InpFact("Von_%d", nil),
		InputMABis:           ma.InpFact("Bis_%d", nil),
		InputMAKurs:          ma.InpFact("Kurs_%d", nil),
		InputMAEigenmittelLC: ma.InpFact("EigenmittelLC_%d", nil),
		InputMADrittmittelLC: ma.InpFact("DrittmittelLC_%d", nil),
		InputMASaldoLC:       ma.InpFact("SaldoLC_%d", nil),
		InputMAManBetrag:     ma.InpFact("ManBetrag_%d", nil),
		InputMAKat:           ma.InpFact("Kat_%d_%d_%d", nil),
		InputMAKmwLC:         ma.InpFact("KmwLC_%d_%d", nil),

		OutputMAPeriode:        ma.OutFact("Periode_%d"),
		OutputMAZeitraum:       ma.OutFact("Zeitraum_%d"),
		OutputMASumLC:          ma.OutFact("SumLC_%d"),
		OutputMASumEUR:         ma.OutFact("SumEUR_%d"),
		OutputMAEigenmittelEUR: ma.OutFact("EigenmittelEUR_%d"),
		OutputMADrittmittelEUR: ma.OutFact("DrittmittelEUR_%d"),
		OutputMAKatEUR:         ma.OutFact("KatEUR_%d_%d_%d"),
		OutputMAKmwEUR:         ma.OutFact("KmwEUR_%d_%d"),

		// Pruefung MA
		InputMAPruefungAuswahl:       maPrue.Inp("Auswahl", nil),
		InputMAPruefungAbzugSaldo:    maPrue.Inp("AbzugSaldo", ListAbzug),
		InputMAPruefungAbzugMehr:     maPrue.Inp("AbzugMehr", ListAbzug),
		InputMAPruefungAbzugPrognose: maPrue.Inp("AbzugPrognose", ListAbzug),
		InputMAPruefungMonateY1:      maPrue.Inp("MonateY1", ListMonate),
		InputMAPruefungMonateY2:      maPrue.Inp("MonateY2", ListMonate),
		InputMAPruefungMonateY3:      maPrue.Inp("MonateY3", ListMonate),
	}
}
