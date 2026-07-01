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
	VAlign:      "center",
	HAlign:      "left",
	BorderTop:   1,
	BorderBottom: 1,
	BorderLeft:  1,
	BorderRight: 1,
	BorderColor: DashClrBorder,
}

// ─── Budget-Farben & Formate ──────────────────────────────────────────────────
const (
	BudgetClrHeader    = "D3D3D3"
	BudgetClrSubhead   = "F0F0F0"
	BudgetClrInput     = "FFFAE5"
	BudgetClrBorder    = "808080"
	BudgetClrGrid      = "D3D3D3"
	BudgetClrFont      = "3C3C3C"
	BudgetClrBlack     = "000000"
	BudgetClrResOff    = "F2F2F2"
	BudgetClrResTxt    = "595959"
	BudgetClrResOn     = "C6EFCE"
	BudgetClrResOnTxt  = "006100"
	BudgetClrBad       = "FFC7CE"
	BudgetClrBadTxt    = "9C0006"

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
