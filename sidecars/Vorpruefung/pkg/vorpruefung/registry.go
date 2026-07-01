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
	ListWaehrung         ValidationList = []string{"EUR", "USD", "CHF"}
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
	Name       string
	Sheet      string
	Validation ValidationList
}

func (i InputField) GetName() string  { return i.Name }
func (i InputField) GetSheet() string { return i.Sheet }

type OutputField struct {
	Name  string
	Sheet string
}

func (o OutputField) GetName() string  { return o.Name }
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
	return InputField{Name: namedRange, Sheet: sheet, Validation: val}
}

func NewOutputField(sheet string, baseName string) OutputField {
	namedRange := "Out_" + baseName
	if strings.HasPrefix(baseName, "Out_") {
		panic(fmt.Sprintf("[Developer Error] Bitte kein 'Out_' mehr angeben. Das wird automatisch hinzugefügt für: %s", namedRange))
	}
	return OutputField{Name: namedRange, Sheet: sheet}
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
	InputBudgetReserveFreigabe InputField
	InputBudgetDrittmittelY1   InputField
	InputBudgetDrittmittelY2   InputField
	InputBudgetDrittmittelY3   InputField
	InputBudgetEigenmittelLC   InputField
	InputBudgetEigenmittelY1   InputField
	InputBudgetEigenmittelY2   InputField
	InputBudgetEigenmittelY3   InputField
	InputBudgetEigenmittelEUR  InputField
	InputBudgetKMWLC           InputField
	InputBudgetKMWY1           InputField
	InputBudgetKMWY2           InputField
	InputBudgetKMWY3           InputField
	InputBudgetKMWEUR          InputField

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

func NewTemplateRegistry() *TemplateRegistry {
	dash := SheetBuilder{Sheet: constants.VPSheetDASHBOARD, Prefix: "Dash_"}
	budget := SheetBuilder{Sheet: constants.VPSheetBUDGET, Prefix: "Budget_"}
	kmw := SheetBuilder{Sheet: constants.VPSheetKMW_MITTEL, Prefix: "KMW_"}
	fb := SheetBuilder{Sheet: constants.VPSheetFINANZBERICHTE, Prefix: "FB_"}
	fbPrue := SheetBuilder{Sheet: constants.VPSheetFB_PRUEFUNG, Prefix: "FBPruef_"}
	ma := SheetBuilder{Sheet: constants.VPSheetMA, Prefix: "MA_"}
	maPrue := SheetBuilder{Sheet: constants.VPSheetMA_PRUEFUNG, Prefix: "MAPruef_"}

	return &TemplateRegistry{
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
		InputDashVPFolgeSaldoEUR:     dash.Inp("VPFolgeSaldoEUR", nil),
		InputDashVPSaldoCheck:        dash.Inp("VPSaldoCheck", nil),
		InputDashVertragCheck:        dash.Inp("VertragCheck", nil),
		InputDashBudgetCheck:         dash.Inp("BudgetCheck", nil),
		InputDashBankBelegeCheck:     dash.Inp("BankBelegeCheck", nil),
		InputDashFBCheck:             dash.Inp("FBCheck", nil),
		InputDashMACheck:             dash.Inp("MACheck", nil),

		OutputDashProjektlaufzeit:     dash.Out("Projektlaufzeit"),
		OutputDashMonate:              dash.Out("Monate"),
		OutputDashSaldoEUR:            dash.Out("SaldoEUR"),
		OutputDashSaldovortragEUR:     dash.Out("SaldovortragEUR"),

		// Budget
		InputBudgetReserveFreigabe: budget.Inp("ReserveFreigabe", ListJaNein),
		InputBudgetDrittmittelY1:   budget.Inp("DrittmittelY1", nil),
		InputBudgetDrittmittelY2:   budget.Inp("DrittmittelY2", nil),
		InputBudgetDrittmittelY3:   budget.Inp("DrittmittelY3", nil),
		InputBudgetEigenmittelLC:   budget.Inp("EigenmittelLC", nil),
		InputBudgetEigenmittelY1:   budget.Inp("EigenmittelY1", nil),
		InputBudgetEigenmittelY2:   budget.Inp("EigenmittelY2", nil),
		InputBudgetEigenmittelY3:   budget.Inp("EigenmittelY3", nil),
		InputBudgetEigenmittelEUR:  budget.Inp("EigenmittelEUR", nil),
		InputBudgetKMWLC:           budget.Inp("KMWLC", nil),
		InputBudgetKMWY1:           budget.Inp("KMWY1", nil),
		InputBudgetKMWY2:           budget.Inp("KMWY2", nil),
		InputBudgetKMWY3:           budget.Inp("KMWY3", nil),
		InputBudgetKMWEUR:          budget.Inp("KMWEUR", nil),

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
