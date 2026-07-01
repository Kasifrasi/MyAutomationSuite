package vorpruefung

// ─── Dashboard-Farben & Formate ───────────────────────────────────────────────
const (
	DashClrHeaderBg     = "6C7A50"
	DashClrHeaderAccent = "667022"
	DashClrTitle        = "DCDCDC"
	DashClrLabel        = "F5F5F5"
	DashClrInput        = "FFFAE5"
	DashClrDisabled     = "F0F0F0"
	DashClrFontGray     = "646464"
	DashClrBorder       = "969696"

	DashFmtLC   = "#,##0.00"
	DashFmtEUR  = `#,##0.00" €"`
	DashFmtDate = "dd.mm.yyyy"
	DashFmtRate = "0.0000"
)

// ─── Dashboard Styles ─────────────────────────────────────────────────────────

var DashHeaderStyle = StyleOptions{
	Bold:         true,
	Size:         18.0,
	FontColor:    "FFFFFF",
	FillColor:    DashClrHeaderBg,
	VAlign:       "center",
	HAlign:       "left",
	BorderBottom: 5,
	BorderColor:  DashClrHeaderAccent,
}

var DashTitleStyle = StyleOptions{
	Bold:         true,
	Size:         13.0,
	FillColor:    DashClrTitle,
	HAlign:       "center",
	VAlign:       "center",
	BorderBottom: 1,
	BorderColor:  DashClrBorder,
}

var DashLabelStyle = StyleOptions{
	Bold:         true,
	Size:         10.0,
	FillColor:    DashClrLabel,
	VAlign:       "center",
	HAlign:       "left",
	WrapText:     true,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashInputStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashInputDateStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "left",
	NumFmtID:     14,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashInputRateStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "left",
	NumFormat:    DashFmtRate,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashInputLCStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "left",
	NumFormat:    DashFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashOutputStyle = StyleOptions{
	FillColor:    DashClrDisabled,
	VAlign:       "center",
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashOutputEURStyle = StyleOptions{
	FillColor:    DashClrDisabled,
	VAlign:       "center",
	HAlign:       "left",
	NumFormat:    DashFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashDropdownStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

// VP-Block: Doppellinie oben
var DashVPLabelStyle = StyleOptions{
	Bold:         true,
	Size:         10.0,
	FillColor:    DashClrLabel,
	VAlign:       "center",
	HAlign:       "left",
	WrapText:     true,
	BorderTop:    6,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

var DashVPInputStyle = StyleOptions{
	FillColor:    DashClrInput,
	VAlign:       "center",
	HAlign:       "left",
	BorderTop:    6,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

// Checkliste
var DashChecklistLabelStyle = StyleOptions{
	Bold:         true,
	Size:         10.0,
	VAlign:       "center",
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
	WrapText:     true,
}

var DashChecklistTextStyle = StyleOptions{
	VAlign:       "center",
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  DashClrBorder,
}

// ─── Budget-Farben & Formate ──────────────────────────────────────────────────
const (
	BudgetClrHeader   = "D3D3D3"
	BudgetClrSubhead  = "F0F0F0"
	BudgetClrInput    = "FFFAE5"
	BudgetClrBorder   = "808080"
	BudgetClrGrid     = "D3D3D3"
	BudgetClrFont     = "3C3C3C"
	BudgetClrBlack    = "000000"
	BudgetClrResOff   = "F2F2F2"
	BudgetClrResTxt   = "595959"
	BudgetClrResOn    = "C6EFCE"
	BudgetClrResOnTxt = "006100"
	BudgetClrBad      = "FFC7CE"
	BudgetClrBadTxt   = "9C0006"

	BudgetFmtLC   = "#,##0.00"
	BudgetFmtEUR  = `#,##0.00" €"`
	BudgetFmtRate = "0.0000"
)

// ─── Budget Styles ────────────────────────────────────────────────────────────

var BudgetTitleStyle = StyleOptions{
	Size:         14,
	Bold:         true,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrHeader,
	VAlign:       "center",
	BorderTop:    2,
	BorderBottom: 2,
	BorderColor:  BudgetClrBorder,
}

var BudgetSectionHdrStyle = StyleOptions{
	Bold:         true,
	Size:         11,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrHeader,
	HAlign:       "left",
	VAlign:       "center",
	BorderTop:    2,
	BorderBottom: 1,
	BorderColor:  BudgetClrBorder,
}

var BudgetSubHdrStyle = StyleOptions{
	Bold:         true,
	Size:         10,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrSubhead,
	HAlign:       "left",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrBorder,
}

var BudgetTableHdrStyle = StyleOptions{
	Bold:         true,
	Size:         9,
	FontColor:    BudgetClrFont,
	FillColor:    BudgetClrHeader,
	HAlign:       "center",
	VAlign:       "center",
	BorderBottom: 2,
	BorderColor:  BudgetClrBorder,
}

var BudgetInputLCStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    BudgetFmtLC,
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

var BudgetInputEURStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    BudgetFmtEUR,
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

var BudgetCatCellStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "left",
	VAlign:       "center",
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

var BudgetIDCellStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "center",
	VAlign:       "center",
	NumFormat:    "@",
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

var BudgetPosCellStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "left",
	VAlign:       "center",
	WrapText:     true,
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

var BudgetTotalRowStyle = StyleOptions{
	Bold:         true,
	Size:         10,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrSubhead,
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  BudgetClrBorder,
}

var BudgetTotalRowLCStyle = StyleOptions{
	Bold:         true,
	Size:         10,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrSubhead,
	VAlign:       "center",
	NumFormat:    BudgetFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  BudgetClrBorder,
}

var BudgetTotalRowEURStyle = StyleOptions{
	Bold:         true,
	Size:         10,
	FontColor:    BudgetClrBlack,
	FillColor:    BudgetClrSubhead,
	VAlign:       "center",
	NumFormat:    BudgetFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  BudgetClrBorder,
}

var BudgetNameCellStyle = StyleOptions{
	FillColor:    BudgetClrInput,
	HAlign:       "left",
	VAlign:       "center",
	BorderLeft:   1,
	BorderRight:  1,
	BorderTop:    1,
	BorderBottom: 1,
	BorderColor:  BudgetClrGrid,
}

// ─── KMW-Mittel-Farben & Formate ──────────────────────────────────────────────
const (
	KMWClrHeader = "D3D3D3" // 211,211,211 – Titel/Kopf/Summe
	KMWClrInput  = "FFFAE5" // 255,250,229 – Eingabezeilen
	KMWClrBorder = "808080" // 128,128,128 – kräftige Rahmen
	KMWClrGrid   = "D3D3D3" // 211,211,211 – dünne Innenrahmen
	KMWClrFont   = "3C3C3C" // 60,60,60    – Kopf-/Summen-Schrift

	KMWFmtBetrag = "#,##0.00"
)

// ─── KMW-Mittel Styles ────────────────────────────────────────────────────────

var KMWTitleStyle = StyleOptions{
	Size:      14.0,
	Bold:      true,
	FillColor: KMWClrHeader,
	HAlign:    "center",
	VAlign:    "center",
}

var KMWHeaderStyle = StyleOptions{
	Bold:         true,
	Size:         9.0,
	FontColor:    KMWClrFont,
	FillColor:    KMWClrHeader,
	HAlign:       "center",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

var KMWInputStyle = StyleOptions{
	FillColor:    KMWClrInput,
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

var KMWBetragStyle = StyleOptions{
	FillColor:    KMWClrInput,
	NumFormat:    KMWFmtBetrag,
	HAlign:       "right",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

var KMWDatumStyle = StyleOptions{
	FillColor:    KMWClrInput,
	NumFmtID:     14, // Excel built-in kurzes Datum
	HAlign:       "center",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

var KMWTotalStyle = StyleOptions{
	Bold:         true,
	Size:         9.0,
	FontColor:    KMWClrFont,
	FillColor:    KMWClrHeader,
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

var KMWTotalBetragStyle = StyleOptions{
	Bold:         true,
	Size:         9.0,
	FontColor:    KMWClrFont,
	FillColor:    KMWClrHeader,
	VAlign:       "center",
	NumFormat:    KMWFmtBetrag,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  KMWClrGrid,
}

// ─── Finanzbericht-Farben & Formate ───────────────────────────────────────────
const (
	FBClrHeader = "D3D3D3" // Titel/Kopf/Sektionen
	FBClrTotal  = "F0F0F0" // Summenzeilen
	FBClrWhite  = "FFFFFF" // Standard-/Ausgabezellen
	FBClrInput  = "FFFAE5" // Eingabezellen
	FBClrGrid   = "D3D3D3" // dünne Innenrahmen
	FBClrBorder = "808080" // kräftige Rahmen
	FBClrMuted  = "808080" // Pfeil / Prüf-Hinweise

	FBFmtLC     = "#,##0.00"
	FBFmtEUR    = `#,##0.00" €"`
	FBFmtKurs   = "0.000000"
	FBFmtMonate = `0" Monate"`
)

// ─── Finanzbericht Styles ─────────────────────────────────────────────────────

// Kopfzeilen-Labels der Periode ("Periode:", "Von:", "Saldo des Finanzberichts" …)
var FBLabelBoldStyle = StyleOptions{
	Bold:   true,
	HAlign: "left",
	VAlign: "center",
}

// Nicht-fettes Label ("Aufschluesselung:")
var FBLabelPlainStyle = StyleOptions{
	HAlign: "left",
	VAlign: "center",
}

var FBPeriodValueStyle = StyleOptions{
	HAlign:       "center",
	VAlign:       "center",
	FillColor:    FBClrTotal,
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

var FBPeriodDatumStyle = StyleOptions{
	HAlign:       "center",
	VAlign:       "center",
	FillColor:    FBClrInput,
	NumFmtID:     14, // Excel built-in kurzes Datum
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

var FBZeitraumStyle = StyleOptions{
	HAlign:       "center",
	VAlign:       "center",
	FillColor:    FBClrTotal,
	NumFormat:    FBFmtMonate,
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

var FBKursStyle = StyleOptions{
	HAlign:       "center",
	VAlign:       "center",
	NumFormat:    FBFmtKurs,
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

// Sektionsköpfe ("Einnahmen", "Ausgaben")
var FBSectionHdrStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrHeader,
	HAlign:       "left",
	VAlign:       "center",
	BorderTop:    2,
	BorderBottom: 1,
	BorderColor:  FBClrBorder,
}

// Einnahmen-Spaltenköpfe
var FBColHdrLabelStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrHeader,
	HAlign:       "left",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

var FBColHdrValStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrHeader,
	HAlign:       "center",
	VAlign:       "center",
	WrapText:     true,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

// Einnahmen-Übersichtszeilen (Vorperiodensaldo, Typ-Zeilen)
var FBIncomeLabelStyle = StyleOptions{
	HAlign:       "left",
	VAlign:       "center",
	FillColor:    FBClrWhite,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBIncomeLCStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrWhite,
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBIncomeEURStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrWhite,
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// Gesamteinnahmen-Summenzeile
var FBIncomeTotalLabelStyle = StyleOptions{
	Bold:         true,
	HAlign:       "left",
	VAlign:       "center",
	FillColor:    FBClrTotal,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

var FBIncomeTotalLCStyle = StyleOptions{
	Bold:         true,
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrTotal,
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

var FBIncomeTotalEURStyle = StyleOptions{
	Bold:         true,
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrTotal,
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

// Ausgaben-Tabelle
var FBAusgHdrStyle = StyleOptions{
	Bold:         true,
	FontColor:    "000000",
	FillColor:    FBClrHeader,
	HAlign:       "center",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

var FBAusgIDStyle = StyleOptions{
	HAlign:       "center",
	VAlign:       "center",
	NumFormat:    "@",
	FillColor:    FBClrWhite,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgLCStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    FBFmtLC,
	FillColor:    FBClrInput,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgEURStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    FBFmtEUR,
	FillColor:    FBClrWhite,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgKumLCStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    FBFmtLC,
	FillColor:    FBClrWhite,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgKumEURStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	NumFormat:    FBFmtEUR,
	FillColor:    FBClrWhite,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// Gesamtausgaben-Summenzeile
var FBAusgTotalLabelStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrTotal,
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgTotalLCStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrTotal,
	HAlign:       "right",
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBAusgTotalEURStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrTotal,
	HAlign:       "right",
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// Saldo des Finanzberichts (Doppel-Unterstreichung)
var FBSaldoLCStyle = StyleOptions{
	Bold:         true,
	NumFormat:    FBFmtLC,
	BorderTop:    6,
	BorderBottom: 6,
}

var FBSaldoEURStyle = StyleOptions{
	Bold:         true,
	NumFormat:    FBFmtEUR,
	BorderTop:    6,
	BorderBottom: 6,
}

// Aufschlüsselung (Bank / Kasse / Sonstiges)
var FBInfoLabelStyle = StyleOptions{
	HAlign:       "left",
	VAlign:       "center",
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

var FBInfoLCStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrInput,
	NumFormat:    FBFmtLC,
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

var FBInfoEURStyle = StyleOptions{
	HAlign:       "right",
	VAlign:       "center",
	FillColor:    FBClrWhite,
	NumFormat:    FBFmtEUR,
	BorderBottom: 1,
	BorderColor:  FBClrGrid,
}

// Differenz (Prüfzeile)
var FBDiffLabelStyle = StyleOptions{
	Size:      8.0,
	FontColor: FBClrMuted,
	HAlign:    "left",
	VAlign:    "center",
}

var FBDiffLCStyle = StyleOptions{
	Size:      8.0,
	FontColor: FBClrMuted,
	HAlign:    "right",
	VAlign:    "center",
	NumFormat: FBFmtLC,
}

var FBDiffEURStyle = StyleOptions{
	Size:      8.0,
	FontColor: FBClrMuted,
	HAlign:    "right",
	VAlign:    "center",
	NumFormat: FBFmtEUR,
}

// Separator-Pfeil zwischen den Perioden
var FBArrowStyle = StyleOptions{
	Size:      24.0,
	Bold:      true,
	FontColor: FBClrMuted,
	HAlign:    "center",
	VAlign:    "center",
}

// ─── Detail-Einnahmentabellen (explizit / Durchschnittskurs) ──────────────────

var FBDetailHdrStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrHeader,
	HAlign:       "center",
	VAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrBorder,
}

// Saldo-Zeile (kursiv, keine Eingabe)
var FBDetailSaldoTypStyle = StyleOptions{
	Italic:       true,
	FillColor:    FBClrWhite,
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailSaldoGeberStyle = StyleOptions{
	Italic:       true,
	FillColor:    FBClrWhite,
	HAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailSaldoLCStyle = StyleOptions{
	Italic:       true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailSaldoEURStyle = StyleOptions{
	Italic:       true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailSaldoKursStyle = StyleOptions{
	Italic:       true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtKurs,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// Eingabezeilen
var FBDetailTypStyle = StyleOptions{
	FillColor:    FBClrInput,
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailGeberStyle = StyleOptions{
	FillColor:    FBClrInput,
	HAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailLCStyle = StyleOptions{
	FillColor:    FBClrInput,
	HAlign:       "right",
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// EUR-Eingabe (explizite Kurstabelle)
var FBDetailEURInputStyle = StyleOptions{
	FillColor:    FBClrInput,
	HAlign:       "right",
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// EUR-Berechnung (Durchschnittskurstabelle)
var FBDetailEURCalcStyle = StyleOptions{
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailKursStyle = StyleOptions{
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtKurs,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

// Summenzeile der Detailtabellen
var FBDetailTotalLabelStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrWhite,
	HAlign:       "left",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailTotalGeberStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrWhite,
	HAlign:       "center",
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailTotalLCStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtLC,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailTotalEURStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtEUR,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}

var FBDetailTotalKursStyle = StyleOptions{
	Bold:         true,
	FillColor:    FBClrWhite,
	HAlign:       "right",
	NumFormat:    FBFmtKurs,
	BorderTop:    1,
	BorderBottom: 1,
	BorderLeft:   1,
	BorderRight:  1,
	BorderColor:  FBClrGrid,
}
