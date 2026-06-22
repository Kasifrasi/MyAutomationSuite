package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// --- Layout-Konstanten ---
const (
	DB_SHEET_NAME = "Dashboard"
	DB_TAB_COLOR  = "D3D3D3" // hellgrau

	DB_C_LBL1 = 2
	DB_C_IN1  = 3
	DB_C_LBL2 = 4
	DB_C_IN2  = 5

	DB_HEADER_ROW = 2 // grüner DASHBOARD-Banner mit Versions-Tag
	DB_TITLE_ROW  = 4 // "Statische Projektinformationen"

	DB_APP_VERSION = "v1.0"

	// --- Benannte Bereiche ---
	DB_NAME_SALDOVORTRAG_LW  = "Saldovortrag_LW"
	DB_NAME_SALDOVORTRAG_EUR = "Saldovortrag_EUR"

	// --- Farb-Konstanten ---
	DB_CLR_HEADER_BG     = "6C7A50" // RGB(108,122,80) – grüner Banner (VBA COL_BG_HEADER)
	DB_CLR_HEADER_ACCENT = "667022" // RGB(102,112,34) – Akzentlinie (VBA COL_ACCENT)
	DB_CLR_TITLE         = "DCDCDC" // RGB(220,220,220)
	DB_CLR_LABEL         = "F5F5F5" // RGB(245,245,245)
	DB_CLR_INPUT         = "FFFAE5" // RGB(255,250,229)
	DB_CLR_DISABLED      = "F0F0F0" // RGB(240,240,240)
	DB_CLR_FONT_GRAY     = "646464" // RGB(100,100,100)
	DB_CLR_BORDER        = "969696" // RGB(150,150,150)

	// --- Zahlenformate ---
	DB_FMT_LC   = "#,##0.00"
	DB_FMT_EUR  = `#,##0.00" €"`
	DB_FMT_DATE = "dd.mm.yyyy"
	DB_FMT_RATE = "0.0000"

	DB_CCY_COL = 27 // Spalte AA (weit rechts, ausgeblendet)

	DB_SALDOVORTRAG_ROW = 13
	DB_DOCS_START_ROW   = 15
)

// --- Dokumenten-Checkliste ---
var DB_DOCS = []string{
	"Vorprojektsaldo (Nachweis)",
	"Vertrag",
	"Budget",
	"Bankbelege",
	"Finanzbericht(e)",
	"Einzelbelege und Beleglisten",
	"Narrativer Bericht",
}

// --- Währungs-Auswahlliste ---
var DB_WAEHRUNGEN = []string{
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

// CreateDashboardSheet legt das Blatt "Dashboard" an und zeichnet
// die statischen Projektinformationen, Checklist und Bedingte Formatierungen.
func (g *Generator) CreateDashboardSheet() error {
	ws := DB_SHEET_NAME
	f := g.file

	// Sheet initialisieren
	_, err := f.NewSheet(ws)
	if err != nil {
		return err
	}

	tabColor := DB_TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	// Spaltenbreiten einstellen
	g.dbSetupColumnWidths(ws)

	// Dashboard-Header zeichnen
	err = g.drawDashboardHeader(ws)
	if err != nil {
		return err
	}

	// Statische Projektinformationen zeichnen
	err = g.drawStaticProjectInfo(ws)
	if err != nil {
		return err
	}

	// Benannte Bereiche für Saldovortrag (LW)/(EUR) anlegen bzw. aktualisieren.
	// Sie zeigen auf die Eingabezellen der Saldovortrag-Zeile.
	g.dbUpsertNamedRange(ws, DB_NAME_SALDOVORTRAG_LW, DB_C_IN1, DB_SALDOVORTRAG_ROW)
	g.dbUpsertNamedRange(ws, DB_NAME_SALDOVORTRAG_EUR, DB_C_IN2, DB_SALDOVORTRAG_ROW)

	return nil
}

func (g *Generator) drawDashboardHeader(ws string) error {
	headerOpts := StyleOptions{
		Bold:         true,
		Size:         18.0,
		FontColor:    "FFFFFF",
		FillColor:    DB_CLR_HEADER_BG,
		VAlign:       "center",
		HAlign:       "left",
		BorderBottom: 5, // Thick bottom border
		BorderColor:  DB_CLR_HEADER_ACCENT,
	}
	err := g.mergeCells(ws, cellName(DB_C_LBL1, DB_HEADER_ROW), cellName(DB_C_IN2, DB_HEADER_ROW), "  DASHBOARD ("+DB_APP_VERSION+")", headerOpts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DB_HEADER_ROW, 40.0)
	return nil
}

func (g *Generator) drawStaticProjectInfo(ws string) error {
	// --- Titel ---
	titleOpts := StyleOptions{
		Bold:         true,
		Size:         13.0,
		FillColor:    DB_CLR_TITLE,
		HAlign:       "center",
		VAlign:       "center",
		BorderBottom: 1,
		BorderColor:  DB_CLR_BORDER,
	}
	err := g.mergeCells(ws, cellName(DB_C_LBL1, DB_TITLE_ROW), cellName(DB_C_IN2, DB_TITLE_ROW), "Statische Projektinformationen", titleOpts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DB_TITLE_ROW, 24.0)

	// Ausgeblendete Währungs-Auswahlliste bereitstellen
	err = g.dbEnsureCurrencyList(ws)
	if err != nil {
		return err
	}

	r := DB_TITLE_ROW + 1

	// --- Zeile: Projektnummer | Vorprojekt vorhanden (Dropdown Ja/Nein) ---
	err = g.dbLabel(ws, r, DB_C_LBL1, "Projektnummer")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, "")
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Vorprojekt vorhanden")
	if err != nil {
		return err
	}
	err = g.dbDropdownJaNein(ws, r, DB_C_IN2, "Ja", DB_CLR_INPUT)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// --- Zeile: Projekttitel (C:E zusammengefasst) ---
	err = g.dbLabel(ws, r, DB_C_LBL1, "Projekttitel")
	if err != nil {
		return err
	}
	titleInputOpts := StyleOptions{
		FillColor:    DB_CLR_INPUT,
		VAlign:       "center",
		HAlign:       "left",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	}
	err = g.mergeCells(ws, cellName(DB_C_IN1, r), cellName(DB_C_IN2, r), "", titleInputOpts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// --- Zeile: Projektträger | Berichtswährung ---
	err = g.dbLabel(ws, r, DB_C_LBL1, "Projekttraeger")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, "")
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Berichtswaehrung")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN2, "")
	if err != nil {
		return err
	}
	err = g.dbCurrencyValidation(ws, r, DB_C_IN2)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// --- Zeile: Projektlaufzeit (geplant) | In Monate (Formel) ---
	err = g.dbLabel(ws, r, DB_C_LBL1, "Projektlaufzeit (geplant)")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, "")
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "In Monate")
	if err != nil {
		return err
	}
	monateCellOpts := StyleOptions{
		FillColor:    DB_CLR_DISABLED, // berechnet -> grau
		VAlign:       "center",
		HAlign:       "left",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	}
	formula := dbMonateFormula(cellName(DB_C_IN1, r))
	err = g.setFormula(ws, cellName(DB_C_IN2, r), formula, monateCellOpts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// ── VORPROJEKT-BLOCK (Doppellinie als Oberkante) ──────────────────────────
	vpStart := r

	// Row 9: Vorprojektnummer | VP-Berichtswaehrung (Double Top Border)
	err = g.setValue(ws, cellName(DB_C_LBL1, r), "Vorprojektnummer", StyleOptions{
		Bold:         true,
		Size:         10.0,
		FillColor:    DB_CLR_LABEL,
		VAlign:       "center",
		HAlign:       "left",
		WrapText:     true,
		BorderTop:    6, // Doppellinie
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	})
	if err != nil {
		return err
	}
	err = g.setValue(ws, cellName(DB_C_IN1, r), "", StyleOptions{
		FillColor:    DB_CLR_INPUT,
		VAlign:       "center",
		HAlign:       "left",
		BorderTop:    6, // Doppellinie
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	})
	if err != nil {
		return err
	}
	err = g.setValue(ws, cellName(DB_C_LBL2, r), "VP-Berichtswaehrung", StyleOptions{
		Bold:         true,
		Size:         10.0,
		FillColor:    DB_CLR_LABEL,
		VAlign:       "center",
		HAlign:       "left",
		WrapText:     true,
		BorderTop:    6, // Doppellinie
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	})
	if err != nil {
		return err
	}
	err = g.setValue(ws, cellName(DB_C_IN2, r), "", StyleOptions{
		FillColor:    DB_CLR_INPUT,
		VAlign:       "center",
		HAlign:       "left",
		BorderTop:    6, // Doppellinie
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	})
	if err != nil {
		return err
	}
	err = g.dbCurrencyValidation(ws, r, DB_C_IN2)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// Row 10: Vorprojektende | Wechselkurs
	err = g.dbLabel(ws, r, DB_C_LBL1, "Vorprojektende")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, DB_FMT_DATE)
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Wechselkurs")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN2, DB_FMT_RATE)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// Row 11: Saldo (LW) | Saldo (EUR)
	err = g.dbLabel(ws, r, DB_C_LBL1, "Saldo (LW)")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, DB_FMT_LC)
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Saldo (EUR)")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN2, DB_FMT_EUR)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// Row 12: Folgeprojektstart | Wechselkurs
	err = g.dbLabel(ws, r, DB_C_LBL1, "Folgeprojektstart")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, DB_FMT_DATE)
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Wechselkurs")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN2, DB_FMT_RATE)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	r++

	// Row 13: Saldovortrag (LW) | Saldovortrag (EUR)
	err = g.dbLabel(ws, r, DB_C_LBL1, "Saldovortrag (LW)")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN1, DB_FMT_LC)
	if err != nil {
		return err
	}
	err = g.dbLabel(ws, r, DB_C_LBL2, "Saldovortrag (EUR)")
	if err != nil {
		return err
	}
	err = g.dbInput(ws, r, DB_C_IN2, DB_FMT_EUR)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, r, 22.0)
	vpEnd := r
	r++ // Skip Row 14 (r=14)
	r++ // Next block starts at row 15 (r=15)

	// ── DOKUMENTEN-CHECKLISTE ──────────────────────────────────────────────────
	docStart := DB_DOCS_START_ROW
	docEnd := docStart + len(DB_DOCS) - 1

	// Beschriftung links merged B15:C21
	docLblOpts := StyleOptions{
		Bold:         true,
		Size:         10.0,
		VAlign:       "center",
		HAlign:       "left",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
		WrapText:     true,
	}
	err = g.mergeCells(ws, cellName(DB_C_LBL1, docStart), cellName(DB_C_IN1, docEnd), "Folgende Dokumente liegen vor:", docLblOpts)
	if err != nil {
		return err
	}

	// Dropdowns (D15:D21) und Texte (E15:E21)
	for i, docName := range DB_DOCS {
		row := docStart + i
		_ = g.file.SetRowHeight(ws, row, 22.0)

		// Ja/Nein Dropdown (Spalte D)
		err = g.dbDropdownJaNein(ws, row, DB_C_LBL2, "Nein", "")
		if err != nil {
			return err
		}

		// Text (Spalte E)
		txtOpts := StyleOptions{
			VAlign:       "center",
			HAlign:       "left",
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  DB_CLR_BORDER,
		}
		err = g.setValue(ws, cellName(DB_C_IN2, row), docName, txtOpts)
		if err != nil {
			return err
		}
	}

	// ── BEDINGTE FORMATIERUNG ──────────────────────────────────────────────────
	err = g.applyConditionalFormatting(ws, vpStart, vpEnd, docStart, docEnd)
	if err != nil {
		return err
	}

	return nil
}

func (g *Generator) applyConditionalFormatting(
	ws string,
	vpStart, vpEnd int,
	docStart, docEnd int,
) error {
	// Adresse der "Vorprojekt vorhanden"-Dropdown (E5 -> $E$5)
	vpAddr := absName(DB_C_IN2, DB_TITLE_ROW+1)

	// 1) Vorprojekt-Block ausgrauen, wenn "Vorprojekt vorhanden" auf "Nein" steht.
	vpCfOpts := StyleOptions{
		FillColor: DB_CLR_DISABLED,
		FontColor: DB_CLR_FONT_GRAY,
	}
	err := g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(DB_C_LBL1, vpStart), cellName(DB_C_IN2, vpEnd)), fmt.Sprintf("=%s=\"Nein\"", vpAddr), vpCfOpts)
	if err != nil {
		return err
	}

	// 2) Dokument-Text durchstreichen + ausgrauen, solange die zugehörige Dropdown (Spalte D) auf "Nein" steht.
	//    Bezug relativ ($D15="Nein") -> gilt zeilenweise. Range: E15:E21
	docCfOpts := StyleOptions{
		FontColor: DB_CLR_FONT_GRAY,
		Strike:    true,
	}
	err = g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(DB_C_IN2, docStart), cellName(DB_C_IN2, docEnd)), fmt.Sprintf("=$%s%d=\"Nein\"", colLetter(DB_C_LBL2), docStart), docCfOpts)
	if err != nil {
		return err
	}

	// 3) Ohne Vorprojekt kann es keinen Vorprojektsaldo-Nachweis geben:
	//    Erste Dokumentzeile (Dropdown + Text) D15:E15 ausgrauen, wenn kein Vorprojekt ("Nein").
	nachweisCfOpts := StyleOptions{
		FillColor: DB_CLR_DISABLED,
		FontColor: DB_CLR_FONT_GRAY,
	}
	err = g.addConditionalFormat(ws, fmt.Sprintf("%s:%s", cellName(DB_C_LBL2, docStart), cellName(DB_C_IN2, docStart)), fmt.Sprintf("=%s=\"Nein\"", vpAddr), nachweisCfOpts)
	if err != nil {
		return err
	}

	return nil
}

// --- Zellen-Hilfsfunktionen ---

func (g *Generator) dbLabel(sheet string, row, col int, text string) error {
	opts := StyleOptions{
		Bold:         true,
		Size:         10.0,
		FillColor:    DB_CLR_LABEL,
		VAlign:       "center",
		HAlign:       "left",
		WrapText:     true,
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	}
	return g.setValue(sheet, cellName(col, row), text, opts)
}

func (g *Generator) dbInput(sheet string, row, col int, numFmt string) error {
	opts := StyleOptions{
		FillColor:    DB_CLR_INPUT,
		VAlign:       "center",
		HAlign:       "left",
		NumFormat:    numFmt,
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	}
	return g.setValue(sheet, cellName(col, row), "", opts)
}

func (g *Generator) dbDropdownJaNein(sheet string, row, col int, defaultValue string, fillColor string) error {
	opts := StyleOptions{
		FillColor:    fillColor,
		VAlign:       "center",
		HAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  DB_CLR_BORDER,
	}
	err := g.setValue(sheet, cellName(col, row), defaultValue, opts)
	if err != nil {
		return err
	}
	dv := excelize.NewDataValidation(true)
	dv.Sqref = cellName(col, row)
	dv.SetDropList([]string{"Ja", "Nein"})
	return g.file.AddDataValidation(sheet, dv)
}

func (g *Generator) dbCurrencyValidation(sheet string, row, col int) error {
	dv := excelize.NewDataValidation(true)
	dv.Sqref = cellName(col, row)
	dv.Type = "list"
	dv.Formula1 = fmt.Sprintf("'%s'!$%s$1:$%s$%d", DB_SHEET_NAME, colLetter(DB_CCY_COL), colLetter(DB_CCY_COL), len(DB_WAEHRUNGEN))
	return g.file.AddDataValidation(sheet, dv)
}

func (g *Generator) dbEnsureCurrencyList(sheet string) error {
	for i, ccy := range DB_WAEHRUNGEN {
		cell := cellName(DB_CCY_COL, i+1)
		err := g.setValue(sheet, cell, ccy, StyleOptions{})
		if err != nil {
			return err
		}
	}
	return g.file.SetColVisible(sheet, colLetter(DB_CCY_COL), false)
}

func (g *Generator) dbSetupColumnWidths(sheet string) {
	// Spalte A: leerer Rand
	// Spalte B: Label 1 (32.0, passend für längere Beschriftungen wie "Projektlaufzeit (geplant)")
	// Spalte C: Eingabe 1 (25.0)
	// Spalte D: Label 2 / Ja/Nein Dropdowns in der Checkliste (24.0, passend für "VP-Berichtswaehrung" und komfortable Ja/Nein-Auswahl)
	// Spalte E: Eingabe 2 / Dokumentenname (35.0, passend für längere Dokumentnamen der Checkliste)
	g.setColWidth(sheet, 1, 3.0)
	g.setColWidth(sheet, 2, 32.0)
	g.setColWidth(sheet, 3, 25.0)
	g.setColWidth(sheet, 4, 24.0)
	g.setColWidth(sheet, 5, 35.0)
}

func (g *Generator) dbUpsertNamedRange(sheet string, name string, col, row int) {
	_ = g.file.DeleteDefinedName(&excelize.DefinedName{Name: name})
	_ = g.file.SetDefinedName(&excelize.DefinedName{
		Name:     name,
		RefersTo: fmt.Sprintf("'%s'!%s", sheet, absName(col, row)),
	})
}

func dbMonateFormula(srcAddr string) string {
	return fmt.Sprintf(`=IFERROR(LET(AllNums, TEXTJOIN("", TRUE, IFERROR(MID(%s, SEQUENCE(LEN(%s)), 1) * 1, "")), StartDate, DATE(MID(AllNums, 5, 4), MID(AllNums, 3, 2), LEFT(AllNums, 2)), EndDate, DATE(MID(AllNums, 13, 4), MID(AllNums, 11, 2), MID(AllNums, 9, 2)), DATEDIF(StartDate, EndDate + 1, "M")), "")`, srcAddr, srcAddr)
}
