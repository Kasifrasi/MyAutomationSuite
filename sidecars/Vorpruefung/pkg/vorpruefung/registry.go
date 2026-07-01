package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"
)

type ValidationList []string

// Zentral definierte Dropdown-Werte, damit die API sie nutzen kann
var (
	ListJaNein   ValidationList = []string{"Ja", "Nein"}
	ListAbzug    ValidationList = []string{"Abzug", "Kein Abzug"}
	ListWaehrung ValidationList = []string{
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

// FBPeriodenAnzahl ist die Anzahl der Berichtsperioden im Finanzbericht.
// Für jede Periode wird pro Tabellen-Factory (Ausgaben/Einnahmen) eine
// dynamische Tabelle 1..FBPeriodenAnzahl erzeugt.
const FBPeriodenAnzahl = 18

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

	// KMW-Mittel-Tabelle
	TableKMWMittel TableField

	// Finanzberichte-Tabellen (dynamisch pro Periode)
	TableFBAusgaben    TableFactory
	TableFBEinnahmen   TableFactory
	TableFBEinnahmenWK TableFactory

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

	OutputDashProjektlaufzeit OutputField
	OutputDashMonate          OutputField
	OutputDashSaldoEUR        OutputField
	OutputDashSaldovortragEUR OutputField

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

	// Finanzberichte
	OutputFBPeriode          OutputFactory
	InputFBVon               InputFactory
	InputFBBis               InputFactory
	OutputFBZeitraum         OutputFactory
	OutputFBKurs             OutputFactory
	OutputFBVSaldoLC         OutputFactory
	OutputFBVSaldoEUR        OutputFactory
	OutputFBVSaldoKumLC      OutputFactory
	OutputFBVSaldoKumEUR     OutputFactory
	OutputFBEMlLC            OutputFactory
	OutputFBEMEUR            OutputFactory
	OutputFBKumEMLC          OutputFactory
	OutputFBKumEMEUR         OutputFactory
	OutputFBDMLC             OutputFactory
	OutputFBDMEUR            OutputFactory
	OutputFBKumDMLC          OutputFactory
	OutputFBKumDMEUR         OutputFactory
	OutputFBKMWLC            OutputFactory
	OutputFBKMWEUR           OutputFactory
	OutputFBKumKMWLC         OutputFactory
	OutputFBKumKMWEUR        OutputFactory
	OutputFBZinsLC           OutputFactory
	OutputFBZinsEUR          OutputFactory
	OutputFBKumZinsLC        OutputFactory
	OutputFBKumZinsEUR       OutputFactory
	OutputFBGEinnahmenLC     OutputFactory
	OutputFBGEinnahmenEUR    OutputFactory
	OutputFBKumGEinnahmenLC  OutputFactory
	OutputFBKumGEinnahmenEUR OutputFactory
	OutputFBSaldoLC          OutputFactory
	OutputFBSaldoEUR         OutputFactory

	InputFBAufschlBankLC        InputFactory
	OutputFBAufschlBankEUR      OutputFactory
	InputFBAufschlKasseLC       InputFactory
	OutputFBAufschlKasseEUR     OutputFactory
	InputFBAufschlSonstigesLC   InputFactory
	OutputFBAufschlSonstigesEUR OutputFactory

	OutputFBDifferenzLC  OutputFactory
	OutputFBDifferenzEUR OutputFactory

	// Pruefung FB
	InputFBPruefungAuswahl    InputField
	InputFBPruefungAbzugSaldo InputField
	InputFBPruefungAbzugMehr  InputField

	// Pruefung FB – Auswahl (berechnete Periodennummer aus der Auswahl)
	OutputFBPruefungAusgewaehltePeriode OutputField

	// Pruefung FB – KMW-Mittelpruefung (berechnete Ergebnisfelder)
	OutputFBPruefungKMWBewilligt           OutputField
	OutputFBPruefungKMWReserve             OutputField
	OutputFBPruefungKMWOperativ            OutputField
	OutputFBPruefungKMWBereitgestellt      OutputField
	OutputFBPruefungKMWVerfuegbar          OutputField
	OutputFBPruefungSaldovortrag           OutputField
	OutputFBPruefungMehreinnahmen          OutputField
	OutputFBPruefungAbzugGesamt            OutputField
	OutputFBPruefungKMWVerfuegbarBereinigt OutputField

	// Pruefung FB – Finanzierungsanteile (Einnahmen-Vergleich: 4 Kategorien + Gesamt, je 8 Spalten = 40)
	OutputFBPruefungFinEMActLC      OutputField
	OutputFBPruefungFinEMBudLC      OutputField
	OutputFBPruefungFinEMDifLC      OutputField
	OutputFBPruefungFinEMAbwLC      OutputField
	OutputFBPruefungFinEMActEUR     OutputField
	OutputFBPruefungFinEMBudEUR     OutputField
	OutputFBPruefungFinEMDifEUR     OutputField
	OutputFBPruefungFinEMAbwEUR     OutputField
	OutputFBPruefungFinDMActLC      OutputField
	OutputFBPruefungFinDMBudLC      OutputField
	OutputFBPruefungFinDMDifLC      OutputField
	OutputFBPruefungFinDMAbwLC      OutputField
	OutputFBPruefungFinDMActEUR     OutputField
	OutputFBPruefungFinDMBudEUR     OutputField
	OutputFBPruefungFinDMDifEUR     OutputField
	OutputFBPruefungFinDMAbwEUR     OutputField
	OutputFBPruefungFinKMWActLC     OutputField
	OutputFBPruefungFinKMWBudLC     OutputField
	OutputFBPruefungFinKMWDifLC     OutputField
	OutputFBPruefungFinKMWAbwLC     OutputField
	OutputFBPruefungFinKMWActEUR    OutputField
	OutputFBPruefungFinKMWBudEUR    OutputField
	OutputFBPruefungFinKMWDifEUR    OutputField
	OutputFBPruefungFinKMWAbwEUR    OutputField
	OutputFBPruefungFinZinsActLC    OutputField
	OutputFBPruefungFinZinsBudLC    OutputField
	OutputFBPruefungFinZinsDifLC    OutputField
	OutputFBPruefungFinZinsAbwLC    OutputField
	OutputFBPruefungFinZinsActEUR   OutputField
	OutputFBPruefungFinZinsBudEUR   OutputField
	OutputFBPruefungFinZinsDifEUR   OutputField
	OutputFBPruefungFinZinsAbwEUR   OutputField
	OutputFBPruefungFinGesamtActLC  OutputField
	OutputFBPruefungFinGesamtBudLC  OutputField
	OutputFBPruefungFinGesamtDifLC  OutputField
	OutputFBPruefungFinGesamtAbwLC  OutputField
	OutputFBPruefungFinGesamtActEUR OutputField
	OutputFBPruefungFinGesamtBudEUR OutputField
	OutputFBPruefungFinGesamtDifEUR OutputField
	OutputFBPruefungFinGesamtAbwEUR OutputField

	// Pruefung FB – Soll-Ist-Abweichungspruefung (8 Kostenkategorien + Gesamt, je 8 Spalten = 72)
	OutputFBPruefungSollIstBauActLC      OutputField
	OutputFBPruefungSollIstBauBudLC      OutputField
	OutputFBPruefungSollIstBauDifLC      OutputField
	OutputFBPruefungSollIstBauAbwLC      OutputField
	OutputFBPruefungSollIstBauActEUR     OutputField
	OutputFBPruefungSollIstBauBudEUR     OutputField
	OutputFBPruefungSollIstBauDifEUR     OutputField
	OutputFBPruefungSollIstBauAbwEUR     OutputField
	OutputFBPruefungSollIstInvActLC      OutputField
	OutputFBPruefungSollIstInvBudLC      OutputField
	OutputFBPruefungSollIstInvDifLC      OutputField
	OutputFBPruefungSollIstInvAbwLC      OutputField
	OutputFBPruefungSollIstInvActEUR     OutputField
	OutputFBPruefungSollIstInvBudEUR     OutputField
	OutputFBPruefungSollIstInvDifEUR     OutputField
	OutputFBPruefungSollIstInvAbwEUR     OutputField
	OutputFBPruefungSollIstPersActLC     OutputField
	OutputFBPruefungSollIstPersBudLC     OutputField
	OutputFBPruefungSollIstPersDifLC     OutputField
	OutputFBPruefungSollIstPersAbwLC     OutputField
	OutputFBPruefungSollIstPersActEUR    OutputField
	OutputFBPruefungSollIstPersBudEUR    OutputField
	OutputFBPruefungSollIstPersDifEUR    OutputField
	OutputFBPruefungSollIstPersAbwEUR    OutputField
	OutputFBPruefungSollIstAktivActLC    OutputField
	OutputFBPruefungSollIstAktivBudLC    OutputField
	OutputFBPruefungSollIstAktivDifLC    OutputField
	OutputFBPruefungSollIstAktivAbwLC    OutputField
	OutputFBPruefungSollIstAktivActEUR   OutputField
	OutputFBPruefungSollIstAktivBudEUR   OutputField
	OutputFBPruefungSollIstAktivDifEUR   OutputField
	OutputFBPruefungSollIstAktivAbwEUR   OutputField
	OutputFBPruefungSollIstVerwActLC     OutputField
	OutputFBPruefungSollIstVerwBudLC     OutputField
	OutputFBPruefungSollIstVerwDifLC     OutputField
	OutputFBPruefungSollIstVerwAbwLC     OutputField
	OutputFBPruefungSollIstVerwActEUR    OutputField
	OutputFBPruefungSollIstVerwBudEUR    OutputField
	OutputFBPruefungSollIstVerwDifEUR    OutputField
	OutputFBPruefungSollIstVerwAbwEUR    OutputField
	OutputFBPruefungSollIstEvalActLC     OutputField
	OutputFBPruefungSollIstEvalBudLC     OutputField
	OutputFBPruefungSollIstEvalDifLC     OutputField
	OutputFBPruefungSollIstEvalAbwLC     OutputField
	OutputFBPruefungSollIstEvalActEUR    OutputField
	OutputFBPruefungSollIstEvalBudEUR    OutputField
	OutputFBPruefungSollIstEvalDifEUR    OutputField
	OutputFBPruefungSollIstEvalAbwEUR    OutputField
	OutputFBPruefungSollIstAuditActLC    OutputField
	OutputFBPruefungSollIstAuditBudLC    OutputField
	OutputFBPruefungSollIstAuditDifLC    OutputField
	OutputFBPruefungSollIstAuditAbwLC    OutputField
	OutputFBPruefungSollIstAuditActEUR   OutputField
	OutputFBPruefungSollIstAuditBudEUR   OutputField
	OutputFBPruefungSollIstAuditDifEUR   OutputField
	OutputFBPruefungSollIstAuditAbwEUR   OutputField
	OutputFBPruefungSollIstReserveActLC  OutputField
	OutputFBPruefungSollIstReserveBudLC  OutputField
	OutputFBPruefungSollIstReserveDifLC  OutputField
	OutputFBPruefungSollIstReserveAbwLC  OutputField
	OutputFBPruefungSollIstReserveActEUR OutputField
	OutputFBPruefungSollIstReserveBudEUR OutputField
	OutputFBPruefungSollIstReserveDifEUR OutputField
	OutputFBPruefungSollIstReserveAbwEUR OutputField
	OutputFBPruefungSollIstGesamtActLC   OutputField
	OutputFBPruefungSollIstGesamtBudLC   OutputField
	OutputFBPruefungSollIstGesamtDifLC   OutputField
	OutputFBPruefungSollIstGesamtAbwLC   OutputField
	OutputFBPruefungSollIstGesamtActEUR  OutputField
	OutputFBPruefungSollIstGesamtBudEUR  OutputField
	OutputFBPruefungSollIstGesamtDifEUR  OutputField
	OutputFBPruefungSollIstGesamtAbwEUR  OutputField

	// MA
	OutputMAPeriode      MAOutputFactory
	InputMAVon           MAInputFactory
	InputMABis           MAInputFactory
	OutputMAZeitraum     MAOutputFactory
	InputMAKurs          MAInputFactory
	
	InputMAKat           MAInputKatFactory
	OutputMAKatEUR         MAOutputKatFactory
	OutputMASumLC          MAOutputFactory
	OutputMASumEUR         MAOutputFactory
	
	InputMAEigenmittelLC MAInputFactory
	OutputMAEigenmittelEUR MAOutputFactory
	InputMADrittmittelLC MAInputFactory
	OutputMADrittmittelEUR MAOutputFactory
	OutputMASaldoLC       MAOutputFactory
	OutputMASaldoEUR     MAOutputFactory
	InputMAAnforderungLC         MAInputFactory
	OutputMAAnforderungEUR         MAOutputFactory
	
	InputMAManBetragEUR     MAInputFactory

	// Pruefung MA
	InputMAPruefungAuswahl       InputField
	InputMAPruefungAbzugSaldo    InputField
	InputMAPruefungAbzugMehr     InputField
	InputMAPruefungAbzugPrognose InputField
	InputMAPruefungMonateY1      InputField
	InputMAPruefungMonateY2      InputField
	InputMAPruefungMonateY3      InputField
}

var Registry = NewTemplateRegistry()

func NewTemplateRegistry() *TemplateRegistry {
	dash := SheetBuilder{Sheet: constants.VPSheetDASHBOARD, Prefix: "Dash_"}
	budget := SheetBuilder{Sheet: constants.VPSheetBUDGET, Prefix: "Budget_"}
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

		// KMW-Mittel-Tabelle
		TableKMWMittel: TableField{
			Name:         "TblKMWMittel",
			Sheet:        constants.VPSheetKMW_MITTEL,
			HasTotalsRow: true,
			Columns: []TableColumn{
				{Header: "Periode"},
				{Header: "Waehrung"},
				{Header: "Betrag", Format: "#,##0.00"},
				{Header: "Datum"},
			},
		},

		// Finanzberichte-Tabellen (dynamisch pro Periode 1..FBPeriodenAnzahl)
		TableFBAusgaben: TableFactory{
			Sheet:        constants.VPSheetFINANZBERICHTE,
			Format:       "Ausgaben_%d",
			HasTotalsRow: true,
			Columns: []TableColumn{
				{Header: "ID"},
				{Header: "Ausgaben (LC)", Format: "#,##0.00"},
				{Header: "Ausgaben (EUR)", Format: `#,##0.00" €"`},
				{Header: "Kum. Ausgaben (LC)", Format: "#,##0.00"},
				{Header: "Kum. Ausgaben (EUR)", Format: `#,##0.00" €"`},
			},
		},
		TableFBEinnahmen: TableFactory{
			Sheet:        constants.VPSheetFINANZBERICHTE,
			Format:       "Einnahmen_%d",
			HasTotalsRow: true,
			Columns: []TableColumn{
				{Header: "Typ"},
				{Header: "Geber"},
				{Header: "Einnahmen (LC)", Format: "#,##0.00"},
				{Header: "Einnahmen (EUR)", Format: `#,##0.00" €"`},
				{Header: "Kurs", Format: "0.000000"},
			},
		},
		TableFBEinnahmenWK: TableFactory{
			Sheet:        constants.VPSheetFINANZBERICHTE,
			Format:       "Einnahmen_WK_%d",
			HasTotalsRow: true,
			Columns: []TableColumn{
				{Header: "Typ"},
				{Header: "Geber"},
				{Header: "Einnahmen (LC)", Format: "#,##0.00"},
				{Header: "Einnahmen (EUR)", Format: `#,##0.00" €"`},
				{Header: "Kurs", Format: "0.000000"},
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

		OutputDashProjektlaufzeit: dash.Out("Projektlaufzeit"),
		OutputDashMonate:          dash.Out("Monate"),
		OutputDashSaldoEUR:        dash.Out("SaldoEUR"),
		OutputDashSaldovortragEUR: dash.Out("SaldovortragEUR"),

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

		// Finanzberichte
		OutputFBPeriode:           fb.OutFact("Periode"),
		InputFBVon:                fb.InpFact("Von_%d", nil),
		InputFBBis:                fb.InpFact("Bis_%d", nil),
		InputFBAufschlBankLC:      fb.InpFact("aufschl_Bank_%d", nil),
		InputFBAufschlKasseLC:     fb.InpFact("aufschl_Kasse_%d", nil),
		InputFBAufschlSonstigesLC: fb.InpFact("aufschl_Sonstiges_%d", nil),

		// Pruefung FB
		InputFBPruefungAuswahl:    fbPrue.Inp("Auswahl", nil),
		InputFBPruefungAbzugSaldo: fbPrue.Inp("AbzugSaldo", ListAbzug),
		InputFBPruefungAbzugMehr:  fbPrue.Inp("AbzugMehr", ListAbzug),

		// Pruefung FB – Auswahl (berechnete Periodennummer aus der Auswahl)
		OutputFBPruefungAusgewaehltePeriode: fbPrue.Out("AusgewaehltePeriode"),

		// Pruefung FB – KMW-Mittelpruefung (berechnete Ergebnisfelder)
		OutputFBPruefungKMWBewilligt:           fbPrue.Out("KMWBewilligt"),
		OutputFBPruefungKMWReserve:             fbPrue.Out("KMWReserve"),
		OutputFBPruefungKMWOperativ:            fbPrue.Out("KMWOperativ"),
		OutputFBPruefungKMWBereitgestellt:      fbPrue.Out("KMWBereitgestellt"),
		OutputFBPruefungKMWVerfuegbar:          fbPrue.Out("KMWVerfuegbar"),
		OutputFBPruefungSaldovortrag:           fbPrue.Out("Saldovortrag"),
		OutputFBPruefungMehreinnahmen:          fbPrue.Out("Mehreinnahmen"),
		OutputFBPruefungAbzugGesamt:            fbPrue.Out("AbzugGesamt"),
		OutputFBPruefungKMWVerfuegbarBereinigt: fbPrue.Out("KMWVerfuegbarBereinigt"),

		// Pruefung FB – Finanzierungsanteile (Einnahmen-Vergleich: 4 Kategorien + Gesamt, je 8 Spalten = 40)
		OutputFBPruefungFinEMActLC:      fbPrue.Out("FinEMActLC"),
		OutputFBPruefungFinEMBudLC:      fbPrue.Out("FinEMBudLC"),
		OutputFBPruefungFinEMDifLC:      fbPrue.Out("FinEMDifLC"),
		OutputFBPruefungFinEMAbwLC:      fbPrue.Out("FinEMAbwLC"),
		OutputFBPruefungFinEMActEUR:     fbPrue.Out("FinEMActEUR"),
		OutputFBPruefungFinEMBudEUR:     fbPrue.Out("FinEMBudEUR"),
		OutputFBPruefungFinEMDifEUR:     fbPrue.Out("FinEMDifEUR"),
		OutputFBPruefungFinEMAbwEUR:     fbPrue.Out("FinEMAbwEUR"),
		OutputFBPruefungFinDMActLC:      fbPrue.Out("FinDMActLC"),
		OutputFBPruefungFinDMBudLC:      fbPrue.Out("FinDMBudLC"),
		OutputFBPruefungFinDMDifLC:      fbPrue.Out("FinDMDifLC"),
		OutputFBPruefungFinDMAbwLC:      fbPrue.Out("FinDMAbwLC"),
		OutputFBPruefungFinDMActEUR:     fbPrue.Out("FinDMActEUR"),
		OutputFBPruefungFinDMBudEUR:     fbPrue.Out("FinDMBudEUR"),
		OutputFBPruefungFinDMDifEUR:     fbPrue.Out("FinDMDifEUR"),
		OutputFBPruefungFinDMAbwEUR:     fbPrue.Out("FinDMAbwEUR"),
		OutputFBPruefungFinKMWActLC:     fbPrue.Out("FinKMWActLC"),
		OutputFBPruefungFinKMWBudLC:     fbPrue.Out("FinKMWBudLC"),
		OutputFBPruefungFinKMWDifLC:     fbPrue.Out("FinKMWDifLC"),
		OutputFBPruefungFinKMWAbwLC:     fbPrue.Out("FinKMWAbwLC"),
		OutputFBPruefungFinKMWActEUR:    fbPrue.Out("FinKMWActEUR"),
		OutputFBPruefungFinKMWBudEUR:    fbPrue.Out("FinKMWBudEUR"),
		OutputFBPruefungFinKMWDifEUR:    fbPrue.Out("FinKMWDifEUR"),
		OutputFBPruefungFinKMWAbwEUR:    fbPrue.Out("FinKMWAbwEUR"),
		OutputFBPruefungFinZinsActLC:    fbPrue.Out("FinZinsActLC"),
		OutputFBPruefungFinZinsBudLC:    fbPrue.Out("FinZinsBudLC"),
		OutputFBPruefungFinZinsDifLC:    fbPrue.Out("FinZinsDifLC"),
		OutputFBPruefungFinZinsAbwLC:    fbPrue.Out("FinZinsAbwLC"),
		OutputFBPruefungFinZinsActEUR:   fbPrue.Out("FinZinsActEUR"),
		OutputFBPruefungFinZinsBudEUR:   fbPrue.Out("FinZinsBudEUR"),
		OutputFBPruefungFinZinsDifEUR:   fbPrue.Out("FinZinsDifEUR"),
		OutputFBPruefungFinZinsAbwEUR:   fbPrue.Out("FinZinsAbwEUR"),
		OutputFBPruefungFinGesamtActLC:  fbPrue.Out("FinGesamtActLC"),
		OutputFBPruefungFinGesamtBudLC:  fbPrue.Out("FinGesamtBudLC"),
		OutputFBPruefungFinGesamtDifLC:  fbPrue.Out("FinGesamtDifLC"),
		OutputFBPruefungFinGesamtAbwLC:  fbPrue.Out("FinGesamtAbwLC"),
		OutputFBPruefungFinGesamtActEUR: fbPrue.Out("FinGesamtActEUR"),
		OutputFBPruefungFinGesamtBudEUR: fbPrue.Out("FinGesamtBudEUR"),
		OutputFBPruefungFinGesamtDifEUR: fbPrue.Out("FinGesamtDifEUR"),
		OutputFBPruefungFinGesamtAbwEUR: fbPrue.Out("FinGesamtAbwEUR"),

		// Pruefung FB – Soll-Ist-Abweichungspruefung (8 Kostenkategorien + Gesamt, je 8 Spalten = 72)
		OutputFBPruefungSollIstBauActLC:      fbPrue.Out("SollIstBauActLC"),
		OutputFBPruefungSollIstBauBudLC:      fbPrue.Out("SollIstBauBudLC"),
		OutputFBPruefungSollIstBauDifLC:      fbPrue.Out("SollIstBauDifLC"),
		OutputFBPruefungSollIstBauAbwLC:      fbPrue.Out("SollIstBauAbwLC"),
		OutputFBPruefungSollIstBauActEUR:     fbPrue.Out("SollIstBauActEUR"),
		OutputFBPruefungSollIstBauBudEUR:     fbPrue.Out("SollIstBauBudEUR"),
		OutputFBPruefungSollIstBauDifEUR:     fbPrue.Out("SollIstBauDifEUR"),
		OutputFBPruefungSollIstBauAbwEUR:     fbPrue.Out("SollIstBauAbwEUR"),
		OutputFBPruefungSollIstInvActLC:      fbPrue.Out("SollIstInvActLC"),
		OutputFBPruefungSollIstInvBudLC:      fbPrue.Out("SollIstInvBudLC"),
		OutputFBPruefungSollIstInvDifLC:      fbPrue.Out("SollIstInvDifLC"),
		OutputFBPruefungSollIstInvAbwLC:      fbPrue.Out("SollIstInvAbwLC"),
		OutputFBPruefungSollIstInvActEUR:     fbPrue.Out("SollIstInvActEUR"),
		OutputFBPruefungSollIstInvBudEUR:     fbPrue.Out("SollIstInvBudEUR"),
		OutputFBPruefungSollIstInvDifEUR:     fbPrue.Out("SollIstInvDifEUR"),
		OutputFBPruefungSollIstInvAbwEUR:     fbPrue.Out("SollIstInvAbwEUR"),
		OutputFBPruefungSollIstPersActLC:     fbPrue.Out("SollIstPersActLC"),
		OutputFBPruefungSollIstPersBudLC:     fbPrue.Out("SollIstPersBudLC"),
		OutputFBPruefungSollIstPersDifLC:     fbPrue.Out("SollIstPersDifLC"),
		OutputFBPruefungSollIstPersAbwLC:     fbPrue.Out("SollIstPersAbwLC"),
		OutputFBPruefungSollIstPersActEUR:    fbPrue.Out("SollIstPersActEUR"),
		OutputFBPruefungSollIstPersBudEUR:    fbPrue.Out("SollIstPersBudEUR"),
		OutputFBPruefungSollIstPersDifEUR:    fbPrue.Out("SollIstPersDifEUR"),
		OutputFBPruefungSollIstPersAbwEUR:    fbPrue.Out("SollIstPersAbwEUR"),
		OutputFBPruefungSollIstAktivActLC:    fbPrue.Out("SollIstAktivActLC"),
		OutputFBPruefungSollIstAktivBudLC:    fbPrue.Out("SollIstAktivBudLC"),
		OutputFBPruefungSollIstAktivDifLC:    fbPrue.Out("SollIstAktivDifLC"),
		OutputFBPruefungSollIstAktivAbwLC:    fbPrue.Out("SollIstAktivAbwLC"),
		OutputFBPruefungSollIstAktivActEUR:   fbPrue.Out("SollIstAktivActEUR"),
		OutputFBPruefungSollIstAktivBudEUR:   fbPrue.Out("SollIstAktivBudEUR"),
		OutputFBPruefungSollIstAktivDifEUR:   fbPrue.Out("SollIstAktivDifEUR"),
		OutputFBPruefungSollIstAktivAbwEUR:   fbPrue.Out("SollIstAktivAbwEUR"),
		OutputFBPruefungSollIstVerwActLC:     fbPrue.Out("SollIstVerwActLC"),
		OutputFBPruefungSollIstVerwBudLC:     fbPrue.Out("SollIstVerwBudLC"),
		OutputFBPruefungSollIstVerwDifLC:     fbPrue.Out("SollIstVerwDifLC"),
		OutputFBPruefungSollIstVerwAbwLC:     fbPrue.Out("SollIstVerwAbwLC"),
		OutputFBPruefungSollIstVerwActEUR:    fbPrue.Out("SollIstVerwActEUR"),
		OutputFBPruefungSollIstVerwBudEUR:    fbPrue.Out("SollIstVerwBudEUR"),
		OutputFBPruefungSollIstVerwDifEUR:    fbPrue.Out("SollIstVerwDifEUR"),
		OutputFBPruefungSollIstVerwAbwEUR:    fbPrue.Out("SollIstVerwAbwEUR"),
		OutputFBPruefungSollIstEvalActLC:     fbPrue.Out("SollIstEvalActLC"),
		OutputFBPruefungSollIstEvalBudLC:     fbPrue.Out("SollIstEvalBudLC"),
		OutputFBPruefungSollIstEvalDifLC:     fbPrue.Out("SollIstEvalDifLC"),
		OutputFBPruefungSollIstEvalAbwLC:     fbPrue.Out("SollIstEvalAbwLC"),
		OutputFBPruefungSollIstEvalActEUR:    fbPrue.Out("SollIstEvalActEUR"),
		OutputFBPruefungSollIstEvalBudEUR:    fbPrue.Out("SollIstEvalBudEUR"),
		OutputFBPruefungSollIstEvalDifEUR:    fbPrue.Out("SollIstEvalDifEUR"),
		OutputFBPruefungSollIstEvalAbwEUR:    fbPrue.Out("SollIstEvalAbwEUR"),
		OutputFBPruefungSollIstAuditActLC:    fbPrue.Out("SollIstAuditActLC"),
		OutputFBPruefungSollIstAuditBudLC:    fbPrue.Out("SollIstAuditBudLC"),
		OutputFBPruefungSollIstAuditDifLC:    fbPrue.Out("SollIstAuditDifLC"),
		OutputFBPruefungSollIstAuditAbwLC:    fbPrue.Out("SollIstAuditAbwLC"),
		OutputFBPruefungSollIstAuditActEUR:   fbPrue.Out("SollIstAuditActEUR"),
		OutputFBPruefungSollIstAuditBudEUR:   fbPrue.Out("SollIstAuditBudEUR"),
		OutputFBPruefungSollIstAuditDifEUR:   fbPrue.Out("SollIstAuditDifEUR"),
		OutputFBPruefungSollIstAuditAbwEUR:   fbPrue.Out("SollIstAuditAbwEUR"),
		OutputFBPruefungSollIstReserveActLC:  fbPrue.Out("SollIstReserveActLC"),
		OutputFBPruefungSollIstReserveBudLC:  fbPrue.Out("SollIstReserveBudLC"),
		OutputFBPruefungSollIstReserveDifLC:  fbPrue.Out("SollIstReserveDifLC"),
		OutputFBPruefungSollIstReserveAbwLC:  fbPrue.Out("SollIstReserveAbwLC"),
		OutputFBPruefungSollIstReserveActEUR: fbPrue.Out("SollIstReserveActEUR"),
		OutputFBPruefungSollIstReserveBudEUR: fbPrue.Out("SollIstReserveBudEUR"),
		OutputFBPruefungSollIstReserveDifEUR: fbPrue.Out("SollIstReserveDifEUR"),
		OutputFBPruefungSollIstReserveAbwEUR: fbPrue.Out("SollIstReserveAbwEUR"),
		OutputFBPruefungSollIstGesamtActLC:   fbPrue.Out("SollIstGesamtActLC"),
		OutputFBPruefungSollIstGesamtBudLC:   fbPrue.Out("SollIstGesamtBudLC"),
		OutputFBPruefungSollIstGesamtDifLC:   fbPrue.Out("SollIstGesamtDifLC"),
		OutputFBPruefungSollIstGesamtAbwLC:   fbPrue.Out("SollIstGesamtAbwLC"),
		OutputFBPruefungSollIstGesamtActEUR:  fbPrue.Out("SollIstGesamtActEUR"),
		OutputFBPruefungSollIstGesamtBudEUR:  fbPrue.Out("SollIstGesamtBudEUR"),
		OutputFBPruefungSollIstGesamtDifEUR:  fbPrue.Out("SollIstGesamtDifEUR"),
		OutputFBPruefungSollIstGesamtAbwEUR:  fbPrue.Out("SollIstGesamtAbwEUR"),

		// MA
		OutputMAPeriode:  ma.MAOutFact("Periode_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAVon:       ma.MAInpFact("Von_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMABis:       ma.MAInpFact("Bis_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAZeitraum: ma.MAOutFact("Zeitraum_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAKurs:      ma.MAInpFact("Kurs_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAKat:     ma.MAInpKatFact("Kat_%d_%d_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAKatEUR: ma.MAOutKatFact("KatEUR_%d_%d_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASumLC:  ma.MAOutFact("SumLC_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASumEUR: ma.MAOutFact("SumEUR_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAEigenmittelLC:   ma.MAInpFact("EigenmittelLC_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAEigenmittelEUR: ma.MAOutFact("EigenmittelEUR_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMADrittmittelLC:   ma.MAInpFact("DrittmittelLC_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMADrittmittelEUR: ma.MAOutFact("DrittmittelEUR_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASaldoLC:        ma.MAOutFact("SaldoLC_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASaldoEUR:       ma.MAOutFact("SaldoEUR_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAAnforderungLC:   ma.MAInpFact("AnforderungLC_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAAnforderungEUR: ma.MAOutFact("AnforderungEUR_%d", MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAManBetragEUR: ma.MAInpFact("ManBetragEUR_%d", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),

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
