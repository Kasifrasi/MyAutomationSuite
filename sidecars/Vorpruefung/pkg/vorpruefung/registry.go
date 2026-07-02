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
	ListWaehrungKMW      ValidationList = []string{"EUR", "USD"}
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
	// Validation ist die statische Dropdown-Liste dieser Spalte (analog zu
	// InputField.Validation) und wird über applyColumnValidation angewendet.
	// Nur für Spalten mit fester Werteauswahl gesetzt.
	Validation ValidationList
	// DynamicValidation ist die dynamische Dropdown-Quelle (Zell- oder Named-Range)
	// für Spalten, deren zulässige Werte erst zur Laufzeit feststehen und sich
	// daher nicht als statische Liste ausdrücken lassen. Wird über
	// applyColumnDynamicValidation angewendet.
	DynamicValidation *DynamicValidation
}

// DynamicValidation beschreibt eine dynamische Dropdown-Validierung, deren
// zulässige Werte aus einer Formel bzw. einem Zell- oder Named-Range stammen.
// Dadurch sind auch dynamische Validierungen zentral in der Registry registriert
// und auffindbar – Sheet-Code setzt Dropdowns ausschließlich über die
// Registry-Spalte (applyColumnValidation / applyColumnDynamicValidation), nie
// mit hartkodierten Formeln.
type DynamicValidation struct {
	// Formula ist die Excel-Dropdown-Quelle (Zellbereich oder Named Range),
	// z. B. "'KMW-Mittel'!$Z$1:$Z$36" oder "Geber_Liste".
	Formula string
	// Note beschreibt den Wertebereich kurz für die zentrale Übersicht.
	Note string
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
	Base         string
	Columns      []TableColumn
	HasTotalsRow bool
}

func (f TableFactory) Get(periode int) TableField {
	return NewTableField(f.Sheet, buildNamedRange(f.Base, periode), f.HasTotalsRow, f.Columns)
}

// ─────────────────────────────────────────────────────────────
// Tabellen-Spalten-Validierungen – zentrale Übersicht
// ─────────────────────────────────────────────────────────────
//
// Damit zentral nachvollziehbar bleibt, welche Werte je Tabellenspalte zulässig
// sind, werden ALLE Dropdown-Validierungen an der Spalte selbst registriert und
// ausschließlich über die sanktionierten Applier angewendet:
//   - statische Liste       → TableColumn.Validation        (applyColumnValidation)
//   - dynamische Quelle      → TableColumn.DynamicValidation (applyColumnDynamicValidation)
// So lassen sich alle Validierungen per grep an einer Stelle finden. Reine
// Typ-Constraints (Zahl/Datum) sind keine Dropdowns und werden über
// TableColumn.Format bzw. NumFmtID abgebildet – hier nur dokumentiert.
//
//   Tabelle               | Spalte           | Art       | Registrierung / Quelle
//   ----------------------|------------------|-----------|----------------------------------------
//   TblBudgetAusgaben     | Kostenkategorie  | Liste     | TableColumn.Validation  = ListKostenkategorien
//   TblKMWMittel          | Waehrung         | Liste     | TableColumn.Validation  = ListWaehrung
//   ----------------------|------------------|-----------|----------------------------------------
//   TblKMWMittel          | Periode          | Dynamisch | TableColumn.DynamicValidation (Hilfsliste Spalte Z)
//   Einnahmen_%d / _WK_%d | Typ              | Dynamisch | TODO migrieren – noch Sheet-Code (Saldo-Label* + feste Typen)
//   Einnahmen_%d / _WK_%d | Geber            | Dynamisch | TODO migrieren – noch Sheet-Code (Named Range "Geber_Liste")
//   Ausgaben_%d           | ID               | Dynamisch | TODO migrieren – noch Sheet-Code (Named Range "Budget_ID_Liste")
//   TblDrittmittel        | Name des Gebers  | Freitext  | keine Validierung; speist Named Range "Geber_Liste"
//   ----------------------|------------------|-----------|----------------------------------------
//   TblKMWMittel          | Betrag / Datum   | Typ       | Zahl (#,##0.00) / Datum (NumFmtID 14)
//   Tbl* (diverse)        | Betrag/EUR/Kurs  | Typ       | Zahl – via TableColumn.Format
//
// Regel: Neue Spalte mit Dropdown → an TableColumn registrieren (Validation ODER
// DynamicValidation) und über den passenden Applier setzen; NIE eine Formel/Liste
// direkt im Sheet-Code hartkodieren. Diese Übersicht entsprechend ergänzen.

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

// buildNamedRange hängt an einen Basisnamen je Koordinate ein "_<n>" an. Die
// Anzahl der Koordinaten – und damit die Dimensionalität – bestimmt der konkrete
// Factory-Typ über seine feste Get-Signatur, NICHT der übergebene String. Dadurch
// ist die Stelligkeit compilerseitig garantiert; ein Basisname trägt nie selbst
// %d-Platzhalter.
func buildNamedRange(base string, coords ...int) string {
	var b strings.Builder
	b.WriteString(base)
	for _, c := range coords {
		fmt.Fprintf(&b, "_%d", c)
	}
	return b.String()
}

// InputFactory erzeugt periodenindizierte Input-Felder (1 Dimension).
type InputFactory struct {
	Sheet string
	Base  string
	Val   ValidationList
}

func (f InputFactory) Get(periode int) InputField {
	return NewInputField(f.Sheet, buildNamedRange(f.Base, periode), f.Val)
}

// OutputFactory erzeugt periodenindizierte Output-Felder (1 Dimension).
type OutputFactory struct {
	Sheet string
	Base  string
}

func (f OutputFactory) Get(periode int) OutputField {
	return NewOutputField(f.Sheet, buildNamedRange(f.Base, periode))
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

func (b SheetBuilder) InpFact(base string, val ValidationList) InputFactory {
	return InputFactory{Sheet: b.Sheet, Base: b.Prefix + base, Val: val}
}

func (b SheetBuilder) OutFact(base string) OutputFactory {
	return OutputFactory{Sheet: b.Sheet, Base: b.Prefix + base}
}

// ─────────────────────────────────────────────────────────────
// 4.1 Erweiterte Factories für Mittelanforderungen
// ─────────────────────────────────────────────────────────────
//
// Mittelanforderungen kennen genau zwei Indizierungen, die hier bewusst als
// getrennte Typen mit fester Get-Stelligkeit abgebildet sind:
//
//   - (Periode, Slot)              → 2 Dimensionen  (MAInputFactory  / MAOutputFactory)
//   - (Periode, Slot, Kategorie)   → 3 Dimensionen  (MAInputKatFactory / MAOutputKatFactory)
//
// Die Basisnamen tragen KEINE %d-Platzhalter; die Koordinaten-Suffixe erzeugt
// buildNamedRange. Periode und Slot werden gegen die Grid-Grenzen geprüft – ein
// Verstoß ist ein statischer Entwicklerfehler und panict schon beim Aufbau der
// Named Range (konsistent zu NewInputField/NewOutputField).

func mustValidateMASlot(periode, slot, maxPerioden, maxSlots int) {
	if periode < 1 || periode > maxPerioden {
		panic(fmt.Sprintf("[Developer Error] ungültige MA-Periode %d (Maximum %d)", periode, maxPerioden))
	}
	if slot < 1 || slot > maxSlots {
		panic(fmt.Sprintf("[Developer Error] ungültiger MA-Slot %d (Maximum %d)", slot, maxSlots))
	}
}

type MAInputFactory struct {
	Sheet       string
	Base        string
	Val         ValidationList
	MaxPerioden int
	MaxSlots    int
}

func (f MAInputFactory) Get(periode, slot int) InputField {
	mustValidateMASlot(periode, slot, f.MaxPerioden, f.MaxSlots)
	return NewInputField(f.Sheet, buildNamedRange(f.Base, periode, slot), f.Val)
}

type MAOutputFactory struct {
	Sheet       string
	Base        string
	MaxPerioden int
	MaxSlots    int
}

func (f MAOutputFactory) Get(periode, slot int) OutputField {
	mustValidateMASlot(periode, slot, f.MaxPerioden, f.MaxSlots)
	return NewOutputField(f.Sheet, buildNamedRange(f.Base, periode, slot))
}

type MAInputKatFactory struct {
	Sheet       string
	Base        string
	Val         ValidationList
	MaxPerioden int
	MaxSlots    int
}

func (f MAInputKatFactory) Get(periode, slot, kategorie int) InputField {
	mustValidateMASlot(periode, slot, f.MaxPerioden, f.MaxSlots)
	return NewInputField(f.Sheet, buildNamedRange(f.Base, periode, slot, kategorie), f.Val)
}

type MAOutputKatFactory struct {
	Sheet       string
	Base        string
	MaxPerioden int
	MaxSlots    int
}

func (f MAOutputKatFactory) Get(periode, slot, kategorie int) OutputField {
	mustValidateMASlot(periode, slot, f.MaxPerioden, f.MaxSlots)
	return NewOutputField(f.Sheet, buildNamedRange(f.Base, periode, slot, kategorie))
}

// 4.2 Hilfsmethoden im SheetBuilder für die MA Factories
func (b SheetBuilder) MAInpFact(base string, val ValidationList, maxPerioden, maxSlots int) MAInputFactory {
	return MAInputFactory{Sheet: b.Sheet, Base: b.Prefix + base, Val: val, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAOutFact(base string, maxPerioden, maxSlots int) MAOutputFactory {
	return MAOutputFactory{Sheet: b.Sheet, Base: b.Prefix + base, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAInpKatFact(base string, val ValidationList, maxPerioden, maxSlots int) MAInputKatFactory {
	return MAInputKatFactory{Sheet: b.Sheet, Base: b.Prefix + base, Val: val, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
}

func (b SheetBuilder) MAOutKatFact(base string, maxPerioden, maxSlots int) MAOutputKatFactory {
	return MAOutputKatFactory{Sheet: b.Sheet, Base: b.Prefix + base, MaxPerioden: maxPerioden, MaxSlots: maxSlots}
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

	// Ergebniszeile der Tabelle TblBudgetAusgaben ("Geplante Gesamtausgaben")
	OutputBudgetAusgabenGesamtLC  OutputField
	OutputBudgetAusgabenGesamtY1  OutputField
	OutputBudgetAusgabenGesamtY2  OutputField
	OutputBudgetAusgabenGesamtY3  OutputField
	OutputBudgetAusgabenGesamtEUR OutputField

	// Ergebniszeile der Tabelle TblKMWMittel ("GESAMT")
	OutputKMWGesamtBetrag OutputField

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

	// Ergebniszeilen der intelligenten FB-Tabellen (je Periode)
	OutputFBAusgGesamtLC     OutputFactory // Ausgaben_%d "Gesamtausgaben"
	OutputFBAusgGesamtEUR    OutputFactory
	OutputFBAusgGesamtKumLC  OutputFactory
	OutputFBAusgGesamtKumEUR OutputFactory
	OutputFBEinnGesamtLC     OutputFactory // Einnahmen_%d "Gesamteinnahmen in Periode"
	OutputFBEinnGesamtEUR    OutputFactory
	OutputFBEinnGesamtKurs   OutputFactory
	OutputFBEinnWKGesamtLC   OutputFactory // Einnahmen_WK_%d "Gesamt (Durchschnittskurs)"
	OutputFBEinnWKGesamtEUR  OutputFactory
	OutputFBEinnWKGesamtKurs OutputFactory

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
	OutputMAPeriode  MAOutputFactory
	InputMAVon       MAInputFactory
	InputMABis       MAInputFactory
	OutputMAZeitraum MAOutputFactory
	InputMAKurs      MAInputFactory

	InputMAKat     MAInputKatFactory
	OutputMAKatEUR MAOutputKatFactory
	OutputMASumLC  MAOutputFactory
	OutputMASumEUR MAOutputFactory

	InputMAEigenmittelLC   MAInputFactory
	OutputMAEigenmittelEUR MAOutputFactory
	InputMADrittmittelLC   MAInputFactory
	OutputMADrittmittelEUR MAOutputFactory
	OutputMASaldoLC        MAOutputFactory
	OutputMASaldoEUR       MAOutputFactory
	InputMAAnforderungLC   MAInputFactory
	OutputMAAnforderungEUR MAOutputFactory

	InputMAManBetragEUR MAInputFactory

	// Pruefung MA
	InputMAPruefungAuswahl       InputField
	InputMAPruefungAbzugSaldo    InputField
	InputMAPruefungAbzugMehr     InputField
	InputMAPruefungAbzugPrognose InputField
	InputMAPruefungMonateY1      InputField
	InputMAPruefungMonateY2      InputField
	InputMAPruefungMonateY3      InputField

	// Pruefung MA – Auswahl (berechnete Perioden-/Anforderungsnummer)
	OutputMAPruefungAusgewaehltePeriode     OutputField
	OutputMAPruefungAusgewaehlteAnforderung OutputField

	// Pruefung MA – KMW-Mittelpruefung (berechnete Ergebnisfelder, isMA)
	OutputMAPruefungKMWBewilligt                   OutputField
	OutputMAPruefungKMWReserve                     OutputField
	OutputMAPruefungKMWOperativ                    OutputField
	OutputMAPruefungKMWBereitgestellt              OutputField
	OutputMAPruefungKMWVerfuegbar                  OutputField
	OutputMAPruefungSaldovortrag                   OutputField
	OutputMAPruefungMehreinnahmen                  OutputField
	OutputMAPruefungPrognostizierteMehreinnahmen   OutputField
	OutputMAPruefungAbzugGesamt                    OutputField
	OutputMAPruefungKMWVerfuegbarBereinigt         OutputField
	OutputMAPruefungVerbleibendKMW                 OutputField
	OutputMAPruefungVerbleibendKMWBereinigt        OutputField
	OutputMAPruefungVerbleibendKMWManuell          OutputField
	OutputMAPruefungVerbleibendKMWManuellBereinigt OutputField

	// Pruefung MA – Monatslimit-Pruefung
	OutputMAPruefungLimitAnforderungLC      OutputField
	OutputMAPruefungLimitAnforderungEUR     OutputField
	OutputMAPruefungLimitJahresbudgetY1LC   OutputField
	OutputMAPruefungLimitJahresbudgetY1EUR  OutputField
	OutputMAPruefungLimitJahresbudgetY2LC   OutputField
	OutputMAPruefungLimitJahresbudgetY2EUR  OutputField
	OutputMAPruefungLimitJahresbudgetY3LC   OutputField
	OutputMAPruefungLimitJahresbudgetY3EUR  OutputField
	OutputMAPruefungLimitMonate             OutputField
	OutputMAPruefungLimitZeitraum           OutputField
	OutputMAPruefungLimitMonatslimitLC      OutputField
	OutputMAPruefungLimitMonatslimitEUR     OutputField
	OutputMAPruefungLimitStatusLC           OutputField
	OutputMAPruefungLimitStatusEUR          OutputField
	OutputMAPruefungLimitUeberschreitungLC  OutputField
	OutputMAPruefungLimitUeberschreitungEUR OutputField
	OutputMAPruefungLimitAuslastungLC       OutputField
	OutputMAPruefungLimitAuslastungEUR      OutputField

	// Pruefung MA – Prognostizierte Finanzierungsanteile (4 Kategorien + Gesamt, je 8 Spalten = 40)
	OutputMAPruefungProgFinEMActLC      OutputField
	OutputMAPruefungProgFinEMBudLC      OutputField
	OutputMAPruefungProgFinEMDifLC      OutputField
	OutputMAPruefungProgFinEMAbwLC      OutputField
	OutputMAPruefungProgFinEMActEUR     OutputField
	OutputMAPruefungProgFinEMBudEUR     OutputField
	OutputMAPruefungProgFinEMDifEUR     OutputField
	OutputMAPruefungProgFinEMAbwEUR     OutputField
	OutputMAPruefungProgFinDMActLC      OutputField
	OutputMAPruefungProgFinDMBudLC      OutputField
	OutputMAPruefungProgFinDMDifLC      OutputField
	OutputMAPruefungProgFinDMAbwLC      OutputField
	OutputMAPruefungProgFinDMActEUR     OutputField
	OutputMAPruefungProgFinDMBudEUR     OutputField
	OutputMAPruefungProgFinDMDifEUR     OutputField
	OutputMAPruefungProgFinDMAbwEUR     OutputField
	OutputMAPruefungProgFinKMWActLC     OutputField
	OutputMAPruefungProgFinKMWBudLC     OutputField
	OutputMAPruefungProgFinKMWDifLC     OutputField
	OutputMAPruefungProgFinKMWAbwLC     OutputField
	OutputMAPruefungProgFinKMWActEUR    OutputField
	OutputMAPruefungProgFinKMWBudEUR    OutputField
	OutputMAPruefungProgFinKMWDifEUR    OutputField
	OutputMAPruefungProgFinKMWAbwEUR    OutputField
	OutputMAPruefungProgFinZinsActLC    OutputField
	OutputMAPruefungProgFinZinsBudLC    OutputField
	OutputMAPruefungProgFinZinsDifLC    OutputField
	OutputMAPruefungProgFinZinsAbwLC    OutputField
	OutputMAPruefungProgFinZinsActEUR   OutputField
	OutputMAPruefungProgFinZinsBudEUR   OutputField
	OutputMAPruefungProgFinZinsDifEUR   OutputField
	OutputMAPruefungProgFinZinsAbwEUR   OutputField
	OutputMAPruefungProgFinGesamtActLC  OutputField
	OutputMAPruefungProgFinGesamtBudLC  OutputField
	OutputMAPruefungProgFinGesamtDifLC  OutputField
	OutputMAPruefungProgFinGesamtAbwLC  OutputField
	OutputMAPruefungProgFinGesamtActEUR OutputField
	OutputMAPruefungProgFinGesamtBudEUR OutputField
	OutputMAPruefungProgFinGesamtDifEUR OutputField
	OutputMAPruefungProgFinGesamtAbwEUR OutputField

	// Pruefung MA – Prognosepruefung Ausgaben (8 Kostenkategorien + Gesamt, je 8 Spalten = 72)
	OutputMAPruefungProgAusgBauActLC      OutputField
	OutputMAPruefungProgAusgBauBudLC      OutputField
	OutputMAPruefungProgAusgBauDifLC      OutputField
	OutputMAPruefungProgAusgBauAbwLC      OutputField
	OutputMAPruefungProgAusgBauActEUR     OutputField
	OutputMAPruefungProgAusgBauBudEUR     OutputField
	OutputMAPruefungProgAusgBauDifEUR     OutputField
	OutputMAPruefungProgAusgBauAbwEUR     OutputField
	OutputMAPruefungProgAusgInvActLC      OutputField
	OutputMAPruefungProgAusgInvBudLC      OutputField
	OutputMAPruefungProgAusgInvDifLC      OutputField
	OutputMAPruefungProgAusgInvAbwLC      OutputField
	OutputMAPruefungProgAusgInvActEUR     OutputField
	OutputMAPruefungProgAusgInvBudEUR     OutputField
	OutputMAPruefungProgAusgInvDifEUR     OutputField
	OutputMAPruefungProgAusgInvAbwEUR     OutputField
	OutputMAPruefungProgAusgPersActLC     OutputField
	OutputMAPruefungProgAusgPersBudLC     OutputField
	OutputMAPruefungProgAusgPersDifLC     OutputField
	OutputMAPruefungProgAusgPersAbwLC     OutputField
	OutputMAPruefungProgAusgPersActEUR    OutputField
	OutputMAPruefungProgAusgPersBudEUR    OutputField
	OutputMAPruefungProgAusgPersDifEUR    OutputField
	OutputMAPruefungProgAusgPersAbwEUR    OutputField
	OutputMAPruefungProgAusgAktivActLC    OutputField
	OutputMAPruefungProgAusgAktivBudLC    OutputField
	OutputMAPruefungProgAusgAktivDifLC    OutputField
	OutputMAPruefungProgAusgAktivAbwLC    OutputField
	OutputMAPruefungProgAusgAktivActEUR   OutputField
	OutputMAPruefungProgAusgAktivBudEUR   OutputField
	OutputMAPruefungProgAusgAktivDifEUR   OutputField
	OutputMAPruefungProgAusgAktivAbwEUR   OutputField
	OutputMAPruefungProgAusgVerwActLC     OutputField
	OutputMAPruefungProgAusgVerwBudLC     OutputField
	OutputMAPruefungProgAusgVerwDifLC     OutputField
	OutputMAPruefungProgAusgVerwAbwLC     OutputField
	OutputMAPruefungProgAusgVerwActEUR    OutputField
	OutputMAPruefungProgAusgVerwBudEUR    OutputField
	OutputMAPruefungProgAusgVerwDifEUR    OutputField
	OutputMAPruefungProgAusgVerwAbwEUR    OutputField
	OutputMAPruefungProgAusgEvalActLC     OutputField
	OutputMAPruefungProgAusgEvalBudLC     OutputField
	OutputMAPruefungProgAusgEvalDifLC     OutputField
	OutputMAPruefungProgAusgEvalAbwLC     OutputField
	OutputMAPruefungProgAusgEvalActEUR    OutputField
	OutputMAPruefungProgAusgEvalBudEUR    OutputField
	OutputMAPruefungProgAusgEvalDifEUR    OutputField
	OutputMAPruefungProgAusgEvalAbwEUR    OutputField
	OutputMAPruefungProgAusgAuditActLC    OutputField
	OutputMAPruefungProgAusgAuditBudLC    OutputField
	OutputMAPruefungProgAusgAuditDifLC    OutputField
	OutputMAPruefungProgAusgAuditAbwLC    OutputField
	OutputMAPruefungProgAusgAuditActEUR   OutputField
	OutputMAPruefungProgAusgAuditBudEUR   OutputField
	OutputMAPruefungProgAusgAuditDifEUR   OutputField
	OutputMAPruefungProgAusgAuditAbwEUR   OutputField
	OutputMAPruefungProgAusgReserveActLC  OutputField
	OutputMAPruefungProgAusgReserveBudLC  OutputField
	OutputMAPruefungProgAusgReserveDifLC  OutputField
	OutputMAPruefungProgAusgReserveAbwLC  OutputField
	OutputMAPruefungProgAusgReserveActEUR OutputField
	OutputMAPruefungProgAusgReserveBudEUR OutputField
	OutputMAPruefungProgAusgReserveDifEUR OutputField
	OutputMAPruefungProgAusgReserveAbwEUR OutputField
	OutputMAPruefungProgAusgGesamtActLC   OutputField
	OutputMAPruefungProgAusgGesamtBudLC   OutputField
	OutputMAPruefungProgAusgGesamtDifLC   OutputField
	OutputMAPruefungProgAusgGesamtAbwLC   OutputField
	OutputMAPruefungProgAusgGesamtActEUR  OutputField
	OutputMAPruefungProgAusgGesamtBudEUR  OutputField
	OutputMAPruefungProgAusgGesamtDifEUR  OutputField
	OutputMAPruefungProgAusgGesamtAbwEUR  OutputField
}

var Registry = NewTemplateRegistry()

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
				{Header: "Kostenkategorie", Validation: ListKostenkategorien},
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
				{Header: "Periode", DynamicValidation: &DynamicValidation{
					Formula: fmt.Sprintf("'%s'!$%s$1:$%s$%d",
						constants.VPSheetKMW_MITTEL,
						colLetter(KMWColValList), colLetter(KMWColValList), KMWPeriodenAnzahl),
					Note: "Hilfsliste Periode 1..36 (Spalte Z, ausgeblendet)",
				}},
				{Header: "Waehrung", Validation: ListWaehrungKMW},
				{Header: "Betrag", Format: "#,##0.00"},
				{Header: "Datum"},
			},
		},

		// Finanzberichte-Tabellen (dynamisch pro Periode 1..FBPeriodenAnzahl)
		TableFBAusgaben: TableFactory{
			Sheet:        constants.VPSheetFINANZBERICHTE,
			Base:         "Ausgaben",
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
			Base:         "Einnahmen",
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
			Base:         "Einnahmen_WK",
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

		OutputBudgetAusgabenGesamtLC:  budget.Out("AusgabenGesamtLC"),
		OutputBudgetAusgabenGesamtY1:  budget.Out("AusgabenGesamtY1"),
		OutputBudgetAusgabenGesamtY2:  budget.Out("AusgabenGesamtY2"),
		OutputBudgetAusgabenGesamtY3:  budget.Out("AusgabenGesamtY3"),
		OutputBudgetAusgabenGesamtEUR: budget.Out("AusgabenGesamtEUR"),

		// KMW-Mittel
		OutputKMWGesamtBetrag: kmw.Out("GesamtBetrag"),

		// Finanzberichte
		OutputFBPeriode:  fb.OutFact("Periode"),
		InputFBVon:       fb.InpFact("Von", nil),
		InputFBBis:       fb.InpFact("Bis", nil),
		OutputFBZeitraum: fb.OutFact("Zeitraum"),
		OutputFBKurs:     fb.OutFact("Kurs"),

		OutputFBVSaldoLC:     fb.OutFact("VSaldoLC"),
		OutputFBVSaldoEUR:    fb.OutFact("VSaldoEUR"),
		OutputFBVSaldoKumLC:  fb.OutFact("VSaldoKumLC"),
		OutputFBVSaldoKumEUR: fb.OutFact("VSaldoKumEUR"),

		OutputFBEMlLC:      fb.OutFact("EMlLC"),
		OutputFBEMEUR:      fb.OutFact("EMEUR"),
		OutputFBKumEMLC:    fb.OutFact("KumEMLC"),
		OutputFBKumEMEUR:   fb.OutFact("KumEMEUR"),
		OutputFBDMLC:       fb.OutFact("DMLC"),
		OutputFBDMEUR:      fb.OutFact("DMEUR"),
		OutputFBKumDMLC:    fb.OutFact("KumDMLC"),
		OutputFBKumDMEUR:   fb.OutFact("KumDMEUR"),
		OutputFBKMWLC:      fb.OutFact("KMWLC"),
		OutputFBKMWEUR:     fb.OutFact("KMWEUR"),
		OutputFBKumKMWLC:   fb.OutFact("KumKMWLC"),
		OutputFBKumKMWEUR:  fb.OutFact("KumKMWEUR"),
		OutputFBZinsLC:     fb.OutFact("ZinsLC"),
		OutputFBZinsEUR:    fb.OutFact("ZinsEUR"),
		OutputFBKumZinsLC:  fb.OutFact("KumZinsLC"),
		OutputFBKumZinsEUR: fb.OutFact("KumZinsEUR"),

		OutputFBGEinnahmenLC:     fb.OutFact("GEinnahmenLC"),
		OutputFBGEinnahmenEUR:    fb.OutFact("GEinnahmenEUR"),
		OutputFBKumGEinnahmenLC:  fb.OutFact("KumGEinnahmenLC"),
		OutputFBKumGEinnahmenEUR: fb.OutFact("KumGEinnahmenEUR"),

		OutputFBSaldoLC:  fb.OutFact("SaldoLC"),
		OutputFBSaldoEUR: fb.OutFact("SaldoEUR"),

		OutputFBAusgGesamtLC:     fb.OutFact("AusgGesamtLC"),
		OutputFBAusgGesamtEUR:    fb.OutFact("AusgGesamtEUR"),
		OutputFBAusgGesamtKumLC:  fb.OutFact("AusgGesamtKumLC"),
		OutputFBAusgGesamtKumEUR: fb.OutFact("AusgGesamtKumEUR"),
		OutputFBEinnGesamtLC:     fb.OutFact("EinnGesamtLC"),
		OutputFBEinnGesamtEUR:    fb.OutFact("EinnGesamtEUR"),
		OutputFBEinnGesamtKurs:   fb.OutFact("EinnGesamtKurs"),
		OutputFBEinnWKGesamtLC:   fb.OutFact("EinnWKGesamtLC"),
		OutputFBEinnWKGesamtEUR:  fb.OutFact("EinnWKGesamtEUR"),
		OutputFBEinnWKGesamtKurs: fb.OutFact("EinnWKGesamtKurs"),

		InputFBAufschlBankLC:        fb.InpFact("aufschl_Bank", nil),
		OutputFBAufschlBankEUR:      fb.OutFact("AufschlBankEUR"),
		InputFBAufschlKasseLC:       fb.InpFact("aufschl_Kasse", nil),
		OutputFBAufschlKasseEUR:     fb.OutFact("AufschlKasseEUR"),
		InputFBAufschlSonstigesLC:   fb.InpFact("aufschl_Sonstiges", nil),
		OutputFBAufschlSonstigesEUR: fb.OutFact("AufschlSonstigesEUR"),

		OutputFBDifferenzLC:  fb.OutFact("DifferenzLC"),
		OutputFBDifferenzEUR: fb.OutFact("DifferenzEUR"),

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
		OutputMAPeriode:  ma.MAOutFact("Periode", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAVon:       ma.MAInpFact("Von", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMABis:       ma.MAInpFact("Bis", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAZeitraum: ma.MAOutFact("Zeitraum", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAKurs:      ma.MAInpFact("Kurs", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAKat:     ma.MAInpKatFact("Kat", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAKatEUR: ma.MAOutKatFact("KatEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASumLC:  ma.MAOutFact("SumLC", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASumEUR: ma.MAOutFact("SumEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAEigenmittelLC:   ma.MAInpFact("EigenmittelLC", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAEigenmittelEUR: ma.MAOutFact("EigenmittelEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMADrittmittelLC:   ma.MAInpFact("DrittmittelLC", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMADrittmittelEUR: ma.MAOutFact("DrittmittelEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASaldoLC:        ma.MAOutFact("SaldoLC", MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMASaldoEUR:       ma.MAOutFact("SaldoEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),
		InputMAAnforderungLC:   ma.MAInpFact("AnforderungLC", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),
		OutputMAAnforderungEUR: ma.MAOutFact("AnforderungEUR", MA_PERIOD_COUNT, EV_MA_SLOTS),

		InputMAManBetragEUR: ma.MAInpFact("ManBetragEUR", nil, MA_PERIOD_COUNT, EV_MA_SLOTS),

		// Pruefung MA
		InputMAPruefungAuswahl:       maPrue.Inp("Auswahl", nil),
		InputMAPruefungAbzugSaldo:    maPrue.Inp("AbzugSaldo", ListAbzug),
		InputMAPruefungAbzugMehr:     maPrue.Inp("AbzugMehr", ListAbzug),
		InputMAPruefungAbzugPrognose: maPrue.Inp("AbzugPrognose", ListAbzug),
		InputMAPruefungMonateY1:      maPrue.Inp("MonateY1", ListMonate),
		InputMAPruefungMonateY2:      maPrue.Inp("MonateY2", ListMonate),
		InputMAPruefungMonateY3:      maPrue.Inp("MonateY3", ListMonate),

		// Pruefung MA – Auswahl (berechnete Perioden-/Anforderungsnummer)
		OutputMAPruefungAusgewaehltePeriode:     maPrue.Out("AusgewaehltePeriode"),
		OutputMAPruefungAusgewaehlteAnforderung: maPrue.Out("AusgewaehlteAnforderung"),

		// Pruefung MA – KMW-Mittelpruefung (berechnete Ergebnisfelder, isMA)
		OutputMAPruefungKMWBewilligt:                   maPrue.Out("KMWBewilligt"),
		OutputMAPruefungKMWReserve:                     maPrue.Out("KMWReserve"),
		OutputMAPruefungKMWOperativ:                    maPrue.Out("KMWOperativ"),
		OutputMAPruefungKMWBereitgestellt:              maPrue.Out("KMWBereitgestellt"),
		OutputMAPruefungKMWVerfuegbar:                  maPrue.Out("KMWVerfuegbar"),
		OutputMAPruefungSaldovortrag:                   maPrue.Out("Saldovortrag"),
		OutputMAPruefungMehreinnahmen:                  maPrue.Out("Mehreinnahmen"),
		OutputMAPruefungPrognostizierteMehreinnahmen:   maPrue.Out("PrognostizierteMehreinnahmen"),
		OutputMAPruefungAbzugGesamt:                    maPrue.Out("AbzugGesamt"),
		OutputMAPruefungKMWVerfuegbarBereinigt:         maPrue.Out("KMWVerfuegbarBereinigt"),
		OutputMAPruefungVerbleibendKMW:                 maPrue.Out("VerbleibendKMW"),
		OutputMAPruefungVerbleibendKMWBereinigt:        maPrue.Out("VerbleibendKMWBereinigt"),
		OutputMAPruefungVerbleibendKMWManuell:          maPrue.Out("VerbleibendKMWManuell"),
		OutputMAPruefungVerbleibendKMWManuellBereinigt: maPrue.Out("VerbleibendKMWManuellBereinigt"),

		// Pruefung MA – Monatslimit-Pruefung
		OutputMAPruefungLimitAnforderungLC:      maPrue.Out("LimitAnforderungLC"),
		OutputMAPruefungLimitAnforderungEUR:     maPrue.Out("LimitAnforderungEUR"),
		OutputMAPruefungLimitJahresbudgetY1LC:   maPrue.Out("LimitJahresbudgetY1LC"),
		OutputMAPruefungLimitJahresbudgetY1EUR:  maPrue.Out("LimitJahresbudgetY1EUR"),
		OutputMAPruefungLimitJahresbudgetY2LC:   maPrue.Out("LimitJahresbudgetY2LC"),
		OutputMAPruefungLimitJahresbudgetY2EUR:  maPrue.Out("LimitJahresbudgetY2EUR"),
		OutputMAPruefungLimitJahresbudgetY3LC:   maPrue.Out("LimitJahresbudgetY3LC"),
		OutputMAPruefungLimitJahresbudgetY3EUR:  maPrue.Out("LimitJahresbudgetY3EUR"),
		OutputMAPruefungLimitMonate:             maPrue.Out("LimitMonate"),
		OutputMAPruefungLimitZeitraum:           maPrue.Out("LimitZeitraum"),
		OutputMAPruefungLimitMonatslimitLC:      maPrue.Out("LimitMonatslimitLC"),
		OutputMAPruefungLimitMonatslimitEUR:     maPrue.Out("LimitMonatslimitEUR"),
		OutputMAPruefungLimitStatusLC:           maPrue.Out("LimitStatusLC"),
		OutputMAPruefungLimitStatusEUR:          maPrue.Out("LimitStatusEUR"),
		OutputMAPruefungLimitUeberschreitungLC:  maPrue.Out("LimitUeberschreitungLC"),
		OutputMAPruefungLimitUeberschreitungEUR: maPrue.Out("LimitUeberschreitungEUR"),
		OutputMAPruefungLimitAuslastungLC:       maPrue.Out("LimitAuslastungLC"),
		OutputMAPruefungLimitAuslastungEUR:      maPrue.Out("LimitAuslastungEUR"),

		// Pruefung MA – Prognostizierte Finanzierungsanteile (4 Kategorien + Gesamt, je 8 Spalten = 40)
		OutputMAPruefungProgFinEMActLC:      maPrue.Out("ProgFinEMActLC"),
		OutputMAPruefungProgFinEMBudLC:      maPrue.Out("ProgFinEMBudLC"),
		OutputMAPruefungProgFinEMDifLC:      maPrue.Out("ProgFinEMDifLC"),
		OutputMAPruefungProgFinEMAbwLC:      maPrue.Out("ProgFinEMAbwLC"),
		OutputMAPruefungProgFinEMActEUR:     maPrue.Out("ProgFinEMActEUR"),
		OutputMAPruefungProgFinEMBudEUR:     maPrue.Out("ProgFinEMBudEUR"),
		OutputMAPruefungProgFinEMDifEUR:     maPrue.Out("ProgFinEMDifEUR"),
		OutputMAPruefungProgFinEMAbwEUR:     maPrue.Out("ProgFinEMAbwEUR"),
		OutputMAPruefungProgFinDMActLC:      maPrue.Out("ProgFinDMActLC"),
		OutputMAPruefungProgFinDMBudLC:      maPrue.Out("ProgFinDMBudLC"),
		OutputMAPruefungProgFinDMDifLC:      maPrue.Out("ProgFinDMDifLC"),
		OutputMAPruefungProgFinDMAbwLC:      maPrue.Out("ProgFinDMAbwLC"),
		OutputMAPruefungProgFinDMActEUR:     maPrue.Out("ProgFinDMActEUR"),
		OutputMAPruefungProgFinDMBudEUR:     maPrue.Out("ProgFinDMBudEUR"),
		OutputMAPruefungProgFinDMDifEUR:     maPrue.Out("ProgFinDMDifEUR"),
		OutputMAPruefungProgFinDMAbwEUR:     maPrue.Out("ProgFinDMAbwEUR"),
		OutputMAPruefungProgFinKMWActLC:     maPrue.Out("ProgFinKMWActLC"),
		OutputMAPruefungProgFinKMWBudLC:     maPrue.Out("ProgFinKMWBudLC"),
		OutputMAPruefungProgFinKMWDifLC:     maPrue.Out("ProgFinKMWDifLC"),
		OutputMAPruefungProgFinKMWAbwLC:     maPrue.Out("ProgFinKMWAbwLC"),
		OutputMAPruefungProgFinKMWActEUR:    maPrue.Out("ProgFinKMWActEUR"),
		OutputMAPruefungProgFinKMWBudEUR:    maPrue.Out("ProgFinKMWBudEUR"),
		OutputMAPruefungProgFinKMWDifEUR:    maPrue.Out("ProgFinKMWDifEUR"),
		OutputMAPruefungProgFinKMWAbwEUR:    maPrue.Out("ProgFinKMWAbwEUR"),
		OutputMAPruefungProgFinZinsActLC:    maPrue.Out("ProgFinZinsActLC"),
		OutputMAPruefungProgFinZinsBudLC:    maPrue.Out("ProgFinZinsBudLC"),
		OutputMAPruefungProgFinZinsDifLC:    maPrue.Out("ProgFinZinsDifLC"),
		OutputMAPruefungProgFinZinsAbwLC:    maPrue.Out("ProgFinZinsAbwLC"),
		OutputMAPruefungProgFinZinsActEUR:   maPrue.Out("ProgFinZinsActEUR"),
		OutputMAPruefungProgFinZinsBudEUR:   maPrue.Out("ProgFinZinsBudEUR"),
		OutputMAPruefungProgFinZinsDifEUR:   maPrue.Out("ProgFinZinsDifEUR"),
		OutputMAPruefungProgFinZinsAbwEUR:   maPrue.Out("ProgFinZinsAbwEUR"),
		OutputMAPruefungProgFinGesamtActLC:  maPrue.Out("ProgFinGesamtActLC"),
		OutputMAPruefungProgFinGesamtBudLC:  maPrue.Out("ProgFinGesamtBudLC"),
		OutputMAPruefungProgFinGesamtDifLC:  maPrue.Out("ProgFinGesamtDifLC"),
		OutputMAPruefungProgFinGesamtAbwLC:  maPrue.Out("ProgFinGesamtAbwLC"),
		OutputMAPruefungProgFinGesamtActEUR: maPrue.Out("ProgFinGesamtActEUR"),
		OutputMAPruefungProgFinGesamtBudEUR: maPrue.Out("ProgFinGesamtBudEUR"),
		OutputMAPruefungProgFinGesamtDifEUR: maPrue.Out("ProgFinGesamtDifEUR"),
		OutputMAPruefungProgFinGesamtAbwEUR: maPrue.Out("ProgFinGesamtAbwEUR"),

		// Pruefung MA – Prognosepruefung Ausgaben (8 Kostenkategorien + Gesamt, je 8 Spalten = 72)
		OutputMAPruefungProgAusgBauActLC:      maPrue.Out("ProgAusgBauActLC"),
		OutputMAPruefungProgAusgBauBudLC:      maPrue.Out("ProgAusgBauBudLC"),
		OutputMAPruefungProgAusgBauDifLC:      maPrue.Out("ProgAusgBauDifLC"),
		OutputMAPruefungProgAusgBauAbwLC:      maPrue.Out("ProgAusgBauAbwLC"),
		OutputMAPruefungProgAusgBauActEUR:     maPrue.Out("ProgAusgBauActEUR"),
		OutputMAPruefungProgAusgBauBudEUR:     maPrue.Out("ProgAusgBauBudEUR"),
		OutputMAPruefungProgAusgBauDifEUR:     maPrue.Out("ProgAusgBauDifEUR"),
		OutputMAPruefungProgAusgBauAbwEUR:     maPrue.Out("ProgAusgBauAbwEUR"),
		OutputMAPruefungProgAusgInvActLC:      maPrue.Out("ProgAusgInvActLC"),
		OutputMAPruefungProgAusgInvBudLC:      maPrue.Out("ProgAusgInvBudLC"),
		OutputMAPruefungProgAusgInvDifLC:      maPrue.Out("ProgAusgInvDifLC"),
		OutputMAPruefungProgAusgInvAbwLC:      maPrue.Out("ProgAusgInvAbwLC"),
		OutputMAPruefungProgAusgInvActEUR:     maPrue.Out("ProgAusgInvActEUR"),
		OutputMAPruefungProgAusgInvBudEUR:     maPrue.Out("ProgAusgInvBudEUR"),
		OutputMAPruefungProgAusgInvDifEUR:     maPrue.Out("ProgAusgInvDifEUR"),
		OutputMAPruefungProgAusgInvAbwEUR:     maPrue.Out("ProgAusgInvAbwEUR"),
		OutputMAPruefungProgAusgPersActLC:     maPrue.Out("ProgAusgPersActLC"),
		OutputMAPruefungProgAusgPersBudLC:     maPrue.Out("ProgAusgPersBudLC"),
		OutputMAPruefungProgAusgPersDifLC:     maPrue.Out("ProgAusgPersDifLC"),
		OutputMAPruefungProgAusgPersAbwLC:     maPrue.Out("ProgAusgPersAbwLC"),
		OutputMAPruefungProgAusgPersActEUR:    maPrue.Out("ProgAusgPersActEUR"),
		OutputMAPruefungProgAusgPersBudEUR:    maPrue.Out("ProgAusgPersBudEUR"),
		OutputMAPruefungProgAusgPersDifEUR:    maPrue.Out("ProgAusgPersDifEUR"),
		OutputMAPruefungProgAusgPersAbwEUR:    maPrue.Out("ProgAusgPersAbwEUR"),
		OutputMAPruefungProgAusgAktivActLC:    maPrue.Out("ProgAusgAktivActLC"),
		OutputMAPruefungProgAusgAktivBudLC:    maPrue.Out("ProgAusgAktivBudLC"),
		OutputMAPruefungProgAusgAktivDifLC:    maPrue.Out("ProgAusgAktivDifLC"),
		OutputMAPruefungProgAusgAktivAbwLC:    maPrue.Out("ProgAusgAktivAbwLC"),
		OutputMAPruefungProgAusgAktivActEUR:   maPrue.Out("ProgAusgAktivActEUR"),
		OutputMAPruefungProgAusgAktivBudEUR:   maPrue.Out("ProgAusgAktivBudEUR"),
		OutputMAPruefungProgAusgAktivDifEUR:   maPrue.Out("ProgAusgAktivDifEUR"),
		OutputMAPruefungProgAusgAktivAbwEUR:   maPrue.Out("ProgAusgAktivAbwEUR"),
		OutputMAPruefungProgAusgVerwActLC:     maPrue.Out("ProgAusgVerwActLC"),
		OutputMAPruefungProgAusgVerwBudLC:     maPrue.Out("ProgAusgVerwBudLC"),
		OutputMAPruefungProgAusgVerwDifLC:     maPrue.Out("ProgAusgVerwDifLC"),
		OutputMAPruefungProgAusgVerwAbwLC:     maPrue.Out("ProgAusgVerwAbwLC"),
		OutputMAPruefungProgAusgVerwActEUR:    maPrue.Out("ProgAusgVerwActEUR"),
		OutputMAPruefungProgAusgVerwBudEUR:    maPrue.Out("ProgAusgVerwBudEUR"),
		OutputMAPruefungProgAusgVerwDifEUR:    maPrue.Out("ProgAusgVerwDifEUR"),
		OutputMAPruefungProgAusgVerwAbwEUR:    maPrue.Out("ProgAusgVerwAbwEUR"),
		OutputMAPruefungProgAusgEvalActLC:     maPrue.Out("ProgAusgEvalActLC"),
		OutputMAPruefungProgAusgEvalBudLC:     maPrue.Out("ProgAusgEvalBudLC"),
		OutputMAPruefungProgAusgEvalDifLC:     maPrue.Out("ProgAusgEvalDifLC"),
		OutputMAPruefungProgAusgEvalAbwLC:     maPrue.Out("ProgAusgEvalAbwLC"),
		OutputMAPruefungProgAusgEvalActEUR:    maPrue.Out("ProgAusgEvalActEUR"),
		OutputMAPruefungProgAusgEvalBudEUR:    maPrue.Out("ProgAusgEvalBudEUR"),
		OutputMAPruefungProgAusgEvalDifEUR:    maPrue.Out("ProgAusgEvalDifEUR"),
		OutputMAPruefungProgAusgEvalAbwEUR:    maPrue.Out("ProgAusgEvalAbwEUR"),
		OutputMAPruefungProgAusgAuditActLC:    maPrue.Out("ProgAusgAuditActLC"),
		OutputMAPruefungProgAusgAuditBudLC:    maPrue.Out("ProgAusgAuditBudLC"),
		OutputMAPruefungProgAusgAuditDifLC:    maPrue.Out("ProgAusgAuditDifLC"),
		OutputMAPruefungProgAusgAuditAbwLC:    maPrue.Out("ProgAusgAuditAbwLC"),
		OutputMAPruefungProgAusgAuditActEUR:   maPrue.Out("ProgAusgAuditActEUR"),
		OutputMAPruefungProgAusgAuditBudEUR:   maPrue.Out("ProgAusgAuditBudEUR"),
		OutputMAPruefungProgAusgAuditDifEUR:   maPrue.Out("ProgAusgAuditDifEUR"),
		OutputMAPruefungProgAusgAuditAbwEUR:   maPrue.Out("ProgAusgAuditAbwEUR"),
		OutputMAPruefungProgAusgReserveActLC:  maPrue.Out("ProgAusgReserveActLC"),
		OutputMAPruefungProgAusgReserveBudLC:  maPrue.Out("ProgAusgReserveBudLC"),
		OutputMAPruefungProgAusgReserveDifLC:  maPrue.Out("ProgAusgReserveDifLC"),
		OutputMAPruefungProgAusgReserveAbwLC:  maPrue.Out("ProgAusgReserveAbwLC"),
		OutputMAPruefungProgAusgReserveActEUR: maPrue.Out("ProgAusgReserveActEUR"),
		OutputMAPruefungProgAusgReserveBudEUR: maPrue.Out("ProgAusgReserveBudEUR"),
		OutputMAPruefungProgAusgReserveDifEUR: maPrue.Out("ProgAusgReserveDifEUR"),
		OutputMAPruefungProgAusgReserveAbwEUR: maPrue.Out("ProgAusgReserveAbwEUR"),
		OutputMAPruefungProgAusgGesamtActLC:   maPrue.Out("ProgAusgGesamtActLC"),
		OutputMAPruefungProgAusgGesamtBudLC:   maPrue.Out("ProgAusgGesamtBudLC"),
		OutputMAPruefungProgAusgGesamtDifLC:   maPrue.Out("ProgAusgGesamtDifLC"),
		OutputMAPruefungProgAusgGesamtAbwLC:   maPrue.Out("ProgAusgGesamtAbwLC"),
		OutputMAPruefungProgAusgGesamtActEUR:  maPrue.Out("ProgAusgGesamtActEUR"),
		OutputMAPruefungProgAusgGesamtBudEUR:  maPrue.Out("ProgAusgGesamtBudEUR"),
		OutputMAPruefungProgAusgGesamtDifEUR:  maPrue.Out("ProgAusgGesamtDifEUR"),
		OutputMAPruefungProgAusgGesamtAbwEUR:  maPrue.Out("ProgAusgGesamtAbwEUR"),
	}
}
