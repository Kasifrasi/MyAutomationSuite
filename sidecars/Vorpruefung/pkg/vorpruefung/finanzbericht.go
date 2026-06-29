package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	SHEET_NAME = constants.VPSheetFINANZBERICHTE
	TAB_COLOR  = "FFFF00" // Gelb

	TABLE_SPACING = 2
	TABLE_COLS    = 5

	// Dynamische Auswahllisten aus dem Budget
	FB_NAME_GEBER_LIST = "Geber_Liste"
	FB_NAME_ID_LIST    = "Budget_ID_Liste"

	START_COL = 2 // Spalte B
	START_ROW = 5 // Zeile 5

	// FB_AUSG_FIRST_ROW ist die erste Datenzeile der Ausgaben-Tabelle einer Periode.
	// Das Zeilen-Layout ist für alle 18 Perioden identisch (nur die Spalte variiert):
	// Periode(4) Von(5) Bis(6) Zeitraum(7) Kurs(8) | Einnahmen-Kopf(9) Spaltenköpfe(10)
	// Saldo(11) Typen(12–15) Gesamteinnahmen(16) | Ausgaben-Kopf(17) Tab-Kopf(18) Daten ab(19).
	FB_AUSG_FIRST_ROW   = 19
	FB_INCOME_FIRST_ROW = 12 // erste Einnahmen-Typzeile (Eigenmittel)

	// ─── Farb-Konstanten ──────────────────────────────────────────────────────────
	COLOR_HEADER = "D3D3D3" // Titel/Kopf/Sektionen
	COLOR_TOTAL  = "F0F0F0" // Summenzeilen
	COLOR_WHITE  = "FFFFFF" // Standard-Zellen
	COLOR_INPUT  = "FFFAE5" // Eingabe-Zellen
)

// ─── Kategorien ──────────────────────────────────────────────────────────────
var (
	TYPE_NAMES = []string{
		"Eigenmittel",
		"Drittmittel",
		"KMW-Mittel",
		"Zinsertraege",
	}

	EXPENSE_CATEGORIES = []string{
		"Bauausgaben",
		"Investitionen",
		"Personalkosten",
		"Projektaktivitaeten",
		"Projektverwaltung",
		"Evaluierung",
		"Audit",
		"Reserve",
	}

	INFO_CATEGORIES = []string{
		"Bank",
		"Kasse",
		"Sonstiges",
	}
)

// CreateFinanzberichteSheet initialisiert das Blatt "III. Finanzberichte" und zeichnet 18 Berichtsperioden.
func (g *Generator) CreateFinanzberichteSheet() error {
	ws := SHEET_NAME
	f := g.file

	// Sheet initialisieren
	_, err := f.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Finanzberichte-Blatts: %w", err)
	}

	tabColor := TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	// Erstelle 18 Perioden nebeneinander
	for p := 1; p <= 18; p++ {
		colStart := START_COL + (p-1)*(TABLE_COLS+TABLE_SPACING)
		isFollowUp := p > 1
		prevStartCol := 0
		if isFollowUp {
			prevStartCol = colStart - (TABLE_COLS + TABLE_SPACING)
		}

		err = g.drawReportTable(ws, colStart, isFollowUp, prevStartCol)
		if err != nil {
			return fmt.Errorf("fehler beim Zeichnen von Periode %d: %w", p, err)
		}
	}

	// Perioden 2–18 gruppieren und zugeklappt ausblenden (nur Inhaltsspalten der Tabelle).
	for p := 2; p <= 18; p++ {
		colStart := START_COL + (p-1)*(TABLE_COLS+TABLE_SPACING)
		groupFirst := colStart
		groupLast := colStart + TABLE_COLS - 1
		for c := groupFirst; c <= groupLast; c++ {
			_ = f.SetColOutlineLevel(ws, colLetter(c), 1)
			_ = f.SetColVisible(ws, colLetter(c), false)
		}
	}

	return nil
}

// drawReportTable zeichnet eine einzelne Berichtsperiode.
func (g *Generator) drawReportTable(
	ws string,
	colStart int,
	isFollowUp bool,
	prevStartCol int,
) error {
	f := g.file
	// Periode rückt eine Zeile nach oben (Zeile 4), damit Von/Bis/Zeitraum/Kurs
	// darunter passen und die Einnahmen weiterhin ab Zeile 9 beginnen.
	r := START_ROW - 1
	periodenNr := ((colStart - START_COL) / (TABLE_COLS + TABLE_SPACING)) + 1

	cLabel := colStart
	cValLC := colStart + 1
	cValEUR := colStart + 2
	cKurs := colStart + 4

	// Reset/Set Spaltenbreiten (5 Spalten für die Periode)
	g.fbSetupColumnWidths(ws, colStart)

	// ➤ Pfeil zeichnen (nur für Nachfolgeperioden)
	if colStart > START_COL {
		_ = g.drawSeparatorArrow(ws, START_ROW-2, colStart-1)
	}

	// ─── KOPFZEILE ─── (Eingabefeld nur über LC..EUR – reduziert)
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Periode:", true, "", false)
	_ = g.drawMergedCell(ws, r, cValLC, cValEUR, fmt.Sprintf("Periode %d", periodenNr), false, COLOR_TOTAL, false)

	hdrRngOpts := StyleOptions{
		HAlign:       "center",
		VAlign:       "center",
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
		FillColor:    COLOR_TOTAL,
	}
	_ = g.setStyle(ws, cellName(cValLC, r), cellName(cValEUR, r), hdrRngOpts)
	r++

	// ─── ZEITRAUM (Von / Bis) ───
	vonRow := r
	for _, zlbl := range []string{"Von:", "Bis:"} {
		_ = g.drawMergedCell(ws, r, cLabel, cLabel, zlbl, true, "", false)
		_ = f.MergeCell(ws, cellName(cValLC, r), cellName(cValEUR, r))
		_ = g.setStyle(ws, cellName(cValLC, r), cellName(cValEUR, r), StyleOptions{
			HAlign:       "center",
			VAlign:       "center",
			BorderBottom: 1,
			BorderColor:  "D3D3D3",
			FillColor:    COLOR_INPUT,
			NumFmtID:     14,
		})
		r++
	}

	// ─── ZEITRAUM (Monate, berechnet) ───
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Zeitraum:", true, "", false)
	_ = g.drawMergedCell(ws, r, cValLC, cValEUR, "", false, COLOR_TOTAL, false)
	_ = f.SetCellFormula(ws, cellName(cValLC, r), fmt.Sprintf(
		`=IF(OR(%s="",%s=""),"",DATEDIF(%s,%s,"m")+1)`,
		cellName(cValLC, vonRow), cellName(cValLC, vonRow+1), cellName(cValLC, vonRow), cellName(cValLC, vonRow+1)))
	_ = g.setStyle(ws, cellName(cValLC, r), cellName(cValEUR, r), StyleOptions{
		HAlign:       "center",
		VAlign:       "center",
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
		FillColor:    COLOR_TOTAL,
		NumFormat:    `0" Monate"`,
	})
	r++

	// ─── DURCHSCHNITTSKURS ───
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Durchschnittskurs:", true, "", false)
	_ = g.drawMergedCell(ws, r, cValLC, cValEUR, "", false, "", false)

	kursRngOpts := StyleOptions{
		HAlign:       "center",
		VAlign:       "center",
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
		NumFormat:    "0.000000",
	}
	_ = g.setStyle(ws, cellName(cValLC, r), cellName(cValEUR, r), kursRngOpts)

	rateRow := r
	rateAddr := absName(cValLC, r) // e.g. "$C$7" for Period 1

	fbKursName := fmt.Sprintf("FB_Kurs_%d", periodenNr)
	g.dbUpsertNamedRange(ws, fbKursName, cValLC, rateRow)
	r++ // Periode/Von/Bis/Kurs belegen Zeilen 5–8; Einnahmen folgen ab Zeile 9

	// ─── EINNAHMEN ───
	_ = g.fbDrawSectionHeader(ws, r, cLabel, cKurs, "Einnahmen")
	r++
	_ = g.fbDrawColumnHeaders(ws, r, cLabel, []string{
		"Einnahmen (LC)",
		"Einnahmen (EUR)",
		"Kum. Einnahmen (LC)",
		"Kum. Einnahmen (EUR)",
	})
	r++

	rowSumStart := r

	// Vorperiodensaldo / Vorprojektsaldo
	saldoLabel := "Vorprojektsaldo"
	if isFollowUp {
		saldoLabel = "Vorperiodensaldo"
	}
	_ = g.fbDrawEinnahmenRow(ws, r, cLabel, saldoLabel)
	cellSaldoVorLC := cellName(cValLC, r)
	cellSaldoVorEUR := cellName(cValEUR, r)

	if isFollowUp {
		_ = f.SetCellFormula(ws, cellName(cValLC, r), fmt.Sprintf("=ROUND(FB_SaldoLC_%d,2)", periodenNr-1))
		_ = f.SetCellFormula(ws, cellName(cValEUR, r), fmt.Sprintf("=ROUND(FB_SaldoEUR_%d,2)", periodenNr-1))
		_ = f.SetCellFormula(ws, cellName(cValLC+2, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_LW, DB_NAME_SALDOVORTRAG_LW))
		_ = f.SetCellFormula(ws, cellName(cValEUR+2, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_EUR, DB_NAME_SALDOVORTRAG_EUR))
	} else {
		_ = f.SetCellFormula(ws, cellName(cValLC, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_LW, DB_NAME_SALDOVORTRAG_LW))
		_ = f.SetCellFormula(ws, cellName(cValEUR, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_EUR, DB_NAME_SALDOVORTRAG_EUR))
		_ = f.SetCellFormula(ws, cellName(cValLC+2, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_LW, DB_NAME_SALDOVORTRAG_LW))
		_ = f.SetCellFormula(ws, cellName(cValEUR+2, r), fmt.Sprintf("=ROUND(IF(%s=\"\",0,%s),2)", DB_NAME_SALDOVORTRAG_EUR, DB_NAME_SALDOVORTRAG_EUR))
	}
	r++

	// Einnahmen-Typen-Zeilen
	var typeRows []int
	for _, typeName := range TYPE_NAMES {
		_ = g.fbDrawEinnahmenRow(ws, r, cLabel, typeName)
		typeRows = append(typeRows, r)
		r++
	}

	// Gesamteinnahmen (inkl. Vorperiodensaldo)
	rngSumLC := fmt.Sprintf("%s:%s", cellName(cValLC, rowSumStart), cellName(cValLC, r-1))
	rngSumEUR := fmt.Sprintf("%s:%s", cellName(cValEUR, rowSumStart), cellName(cValEUR, r-1))
	rngSumKumLC := fmt.Sprintf("%s:%s", cellName(cValLC+2, rowSumStart), cellName(cValLC+2, r-1))
	rngSumKumEUR := fmt.Sprintf("%s:%s", cellName(cValEUR+2, rowSumStart), cellName(cValEUR+2, r-1))

	_ = g.fbDrawTotalRow(ws, r, cLabel, "Gesamteinnahmen", []string{
		rngSumLC,
		rngSumEUR,
		rngSumKumLC,
		rngSumKumEUR,
	})
	cellSumEinnahmenLC := cellName(cValLC, r)
	cellSumEinnahmenEUR := cellName(cValEUR, r)
	cellSumEinnahmenKumLC := cellName(cValLC+2, r)
	cellSumEinnahmenKumEUR := cellName(cValEUR+2, r)
	r++

	// ─── AUSGABEN ───
	_ = g.fbDrawSectionHeader(ws, r, cLabel, cKurs, "Ausgaben")
	r++

	ausgName := fmt.Sprintf("Ausgaben_%d", periodenNr)
	ausgHdrRow := r

	// Table Headers schreiben
	headers := []string{"ID", "Ausgaben (LC)", "Ausgaben (EUR)", "Kum. Ausgaben (LC)", "Kum. Ausgaben (EUR)"}
	for i, h := range headers {
		cell := cellName(cLabel+i, ausgHdrRow)
		_ = f.SetCellValue(ws, cell, h)
	}

	// Kopfzeile formatieren
	ausgHdrOpts := StyleOptions{
		Bold:         true,
		FillColor:    COLOR_HEADER,
		FontColor:    "000000",
		HAlign:       "center",
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  "808080",
	}
	for i := 0; i < 5; i++ {
		_ = g.setStyle(ws, cellName(cLabel+i, ausgHdrRow), cellName(cLabel+i, ausgHdrRow), ausgHdrOpts)
	}

	ausgDataRows := g.budgetExpenseCount()
	ausgTotalsRow := ausgHdrRow + ausgDataRows + 1

	// Add data body range to Ausgaben list for VSTACK
	dataRangeAusg := fmt.Sprintf("'%s'!%s:%s", ws, absName(cLabel, ausgHdrRow+1), absName(cLabel+4, ausgHdrRow+ausgDataRows))
	g.rangesAusgaben = append(g.rangesAusgaben, dataRangeAusg)

	// Tabelle erstellen
	err := f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(cLabel, ausgHdrRow), cellName(cLabel+4, ausgTotalsRow-1)),
		Name:           ausgName,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})
	if err != nil {
		return err
	}

	// Formeln für Tabellenzeilen setzen
	for i := 0; i < ausgDataRows; i++ {
		row := ausgHdrRow + 1 + i

		// ID: bei Config fester Wert (sortierbar, klare Zuteilung), sonst per INDEX
		// aus der Budget-ID-Liste.
		if g.budget != nil {
			_ = f.SetCellValue(ws, cellName(cLabel, row), g.budget.Ausgaben[i].ID)
		} else {
			_ = f.SetCellFormula(ws, cellName(cLabel, row), fmt.Sprintf(`=IFERROR(INDEX(%s, ROW() - %d), "")`, FB_NAME_ID_LIST, ausgHdrRow))
		}

		// EUR Formula (direct absolute cell address is used here to avoid named range resolution issues with formulas in Excel)
		_ = f.SetCellFormula(ws, cellName(cLabel+2, row), fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, cellName(cLabel+1, row), rateAddr))

		// Kum LC and EUR formulas
		if periodenNr > 1 {
			_ = f.SetCellFormula(ws, cellName(cLabel+3, row), fmt.Sprintf(`=ROUND(%s+%s,2)`, cellName(cLabel+1, row), cellName(prevStartCol+3, row)))
			_ = f.SetCellFormula(ws, cellName(cLabel+4, row), fmt.Sprintf(`=ROUND(%s+%s,2)`, cellName(cLabel+2, row), cellName(prevStartCol+4, row)))
		} else {
			_ = f.SetCellFormula(ws, cellName(cLabel+3, row), fmt.Sprintf(`=ROUND(%s,2)`, cellName(cLabel+1, row)))
			_ = f.SetCellFormula(ws, cellName(cLabel+4, row), fmt.Sprintf(`=ROUND(%s,2)`, cellName(cLabel+2, row)))
		}
	}

	// Datenzellen formatieren (B19:F26)
	for i := 0; i < ausgDataRows; i++ {
		row := ausgHdrRow + 1 + i

		// ID (cLabel)
		_ = g.setStyle(ws, cellName(cLabel, row), cellName(cLabel, row), StyleOptions{
			HAlign: "center", VAlign: "center", NumFormat: "@", FillColor: COLOR_WHITE,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})

		// LC (cLabel+1)
		_ = g.setStyle(ws, cellName(cLabel+1, row), cellName(cLabel+1, row), StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: "#,##0.00", FillColor: COLOR_INPUT,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})

		// EUR (cLabel+2)
		_ = g.setStyle(ws, cellName(cLabel+2, row), cellName(cLabel+2, row), StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: `#,##0.00" €"`, FillColor: COLOR_WHITE,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})

		// Kum LC (cLabel+3)
		_ = g.setStyle(ws, cellName(cLabel+3, row), cellName(cLabel+3, row), StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: "#,##0.00", FillColor: COLOR_WHITE,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})

		// Kum EUR (cLabel+4)
		_ = g.setStyle(ws, cellName(cLabel+4, row), cellName(cLabel+4, row), StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: `#,##0.00" €"`, FillColor: COLOR_WHITE,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})
	}

	// Summen-Zeile Gesamtausgaben schreiben
	_ = f.SetCellValue(ws, cellName(cLabel, ausgTotalsRow), "Gesamtausgaben")

	lcRange := fmt.Sprintf("%s:%s", absName(cLabel+1, ausgHdrRow+1), absName(cLabel+1, ausgHdrRow+ausgDataRows))
	eurRange := fmt.Sprintf("%s:%s", absName(cLabel+2, ausgHdrRow+1), absName(cLabel+2, ausgHdrRow+ausgDataRows))
	kumLcRange := fmt.Sprintf("%s:%s", absName(cLabel+3, ausgHdrRow+1), absName(cLabel+3, ausgHdrRow+ausgDataRows))
	kumEurRange := fmt.Sprintf("%s:%s", absName(cLabel+4, ausgHdrRow+1), absName(cLabel+4, ausgHdrRow+ausgDataRows))

	_ = f.SetCellFormula(ws, cellName(cLabel+1, ausgTotalsRow), fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s),2)`, lcRange))
	_ = f.SetCellFormula(ws, cellName(cLabel+2, ausgTotalsRow), fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s),2)`, eurRange))
	_ = f.SetCellFormula(ws, cellName(cLabel+3, ausgTotalsRow), fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s),2)`, kumLcRange))
	_ = f.SetCellFormula(ws, cellName(cLabel+4, ausgTotalsRow), fmt.Sprintf(`=ROUND(SUBTOTAL(109,%s),2)`, kumEurRange))

	// Summen-Zeile formatieren
	_ = g.setStyle(ws, cellName(cLabel, ausgTotalsRow), cellName(cLabel, ausgTotalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_TOTAL, HAlign: "left", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(cLabel+1, ausgTotalsRow), cellName(cLabel+1, ausgTotalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_TOTAL, HAlign: "right", NumFormat: "#,##0.00", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(cLabel+2, ausgTotalsRow), cellName(cLabel+2, ausgTotalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_TOTAL, HAlign: "right", NumFormat: `#,##0.00" €"`, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(cLabel+3, ausgTotalsRow), cellName(cLabel+3, ausgTotalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_TOTAL, HAlign: "right", NumFormat: "#,##0.00", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(cLabel+4, ausgTotalsRow), cellName(cLabel+4, ausgTotalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_TOTAL, HAlign: "right", NumFormat: `#,##0.00" €"`, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})

	ausgTotLCAddr := absName(cValLC, ausgTotalsRow)
	ausgTotEURAddr := absName(cValEUR, ausgTotalsRow)

	r = ausgTotalsRow + 1
	r++ // Leerzeile

	// ─── SALDO DES FINANZBERICHTS ───
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Saldo des Finanzberichts", true, "", false)
	cellSaldoTotal := cellName(cValLC, r)
	cellSaldoTotalEUR := cellName(cValEUR, r)

	// Formeln für Saldo Range
	_ = f.SetCellFormula(ws, cellName(cValLC, r), fmt.Sprintf(`=ROUND(IFERROR(%s-%s,""),2)`, cellSumEinnahmenLC, ausgTotLCAddr))
	_ = f.SetCellFormula(ws, cellName(cValEUR, r), fmt.Sprintf(`=ROUND(IFERROR(%s-%s,""),2)`, cellSumEinnahmenEUR, ausgTotEURAddr))
	_ = f.SetCellFormula(ws, cellName(cValLC+2, r), fmt.Sprintf(`=ROUND(IFERROR(%s-%s,""),2)`, cellSumEinnahmenKumLC, cellName(cValLC+2, ausgTotalsRow)))
	_ = f.SetCellFormula(ws, cellName(cValEUR+2, r), fmt.Sprintf(`=ROUND(IFERROR(%s-%s,""),2)`, cellSumEinnahmenKumEUR, cellName(cValEUR+2, ausgTotalsRow)))

	// Saldo-Zellen formatieren (Doppel-Unterstreichung)
	_ = g.setStyle(ws, cellName(cValLC, r), cellName(cValLC, r), StyleOptions{Bold: true, NumFormat: "#,##0.00", BorderTop: 6, BorderBottom: 6})
	_ = g.setStyle(ws, cellName(cValEUR, r), cellName(cValEUR, r), StyleOptions{Bold: true, NumFormat: `#,##0.00" €"`, BorderTop: 6, BorderBottom: 6})
	_ = g.setStyle(ws, cellName(cValLC+2, r), cellName(cValLC+2, r), StyleOptions{Bold: true, NumFormat: "#,##0.00", BorderTop: 6, BorderBottom: 6})
	_ = g.setStyle(ws, cellName(cValEUR+2, r), cellName(cValEUR+2, r), StyleOptions{Bold: true, NumFormat: `#,##0.00" €"`, BorderTop: 6, BorderBottom: 6})

	// „Saldo des Finanzberichts" (LC/EUR) benennen
	g.dbUpsertNamedRange(ws, fmt.Sprintf("FB_SaldoLC_%d", periodenNr), cValLC, r)
	g.dbUpsertNamedRange(ws, fmt.Sprintf("FB_SaldoEUR_%d", periodenNr), cValEUR, r)
	r += 2

	// ─── AUFSCHLÜSSELUNG ───
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Aufschluesselung:", false, "", false)
	r++

	rowStartAufsch := r
	rowEndAufsch := r + len(INFO_CATEGORIES) - 1
	sumLC := fmt.Sprintf(`ROUND(SUM(%s:%s), 2)`, cellName(cValLC, rowStartAufsch), cellName(cValLC, rowEndAufsch))

	for i, cat := range INFO_CATEGORIES {
		isLast := i == len(INFO_CATEGORIES)-1
		currentLC := cellName(cValLC, r)

		// Named Ranges für die gelben Eingabefelder (Bank, Kasse, Sonstiges) anlegen
		var safeName string
		switch cat {
		case "Bank":
			safeName = "Bank"
		case "Kasse":
			safeName = "Kasse"
		default:
			safeName = "Sonstiges"
		}
		g.dbUpsertNamedRange(ws, fmt.Sprintf("aufschl_%s_%d", safeName, periodenNr), cValLC, r)

		var formulaEUR string
		if !isLast {
			remainingLCs := fmt.Sprintf(`SUM(%s:%s)`, cellName(cValLC, r+1), cellName(cValLC, rowEndAufsch))
			var prevSum string
			if i == 0 {
				prevSum = "0"
			} else {
				prevSum = fmt.Sprintf(`SUM(%s:%s)`, cellName(cValEUR, rowStartAufsch), cellName(cValEUR, r-1))
			}
			formulaEUR = fmt.Sprintf(`=IF(%s=0, 0, IF(ROUND(%s, 2)=0, ROUND(%s - %s, 2), ROUND(%s / %s * %s, 2)))`, currentLC, remainingLCs, cellSaldoTotalEUR, prevSum, currentLC, sumLC, cellSaldoTotalEUR)
		} else {
			prevSum := fmt.Sprintf(`SUM(%s:%s)`, cellName(cValEUR, rowStartAufsch), cellName(cValEUR, r-1))
			formulaEUR = fmt.Sprintf(`=IF(%s=0, 0, ROUND(%s - %s, 2))`, currentLC, cellSaldoTotalEUR, prevSum)
		}

		_ = g.fbDrawInfoRow(ws, r, cLabel, cValLC, cValEUR, formulaEUR)
		_ = f.SetCellValue(ws, cellName(cLabel, r), cat)
		r++
	}
	r++ // blank row

	// Differenz (Prüfung)
	_ = g.fbDrawDifferenceRow(ws, r, cLabel, cValLC, cValEUR, cellSaldoTotal, cellSaldoTotalEUR, rowStartAufsch, rowEndAufsch)
	r++

	// Äußerer Rahmen für den gesamten Hauptblock (Periode-Kopf ab Zeile 4)
	_ = g.styleOuterBorder(ws, START_ROW-1, colStart, r-1, colStart+TABLE_COLS-1, 2, "808080")
	r += 2

	// ==================================================================================
	// EINNAHMEN-TABELLEN (Detail)
	// ==================================================================================
	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Einnahmen (Explizite Kurseingabe)", true, "", false)
	r++
	r1_start := r // Merken für Ranges

	_, totalsRow1, err := g.createEinnahmenTabelle(ws, r, colStart, periodenNr, cellSaldoVorLC, cellSaldoVorEUR, false, rateAddr)
	if err != nil {
		return err
	}
	r = totalsRow1 + 2

	_ = g.drawMergedCell(ws, r, cLabel, cLabel, "Einnahmen (Durchschnittskurs)", true, "", false)
	r++
	r2_start := r // Merken für Ranges

	_, totalsRow2, err := g.createEinnahmenTabelle(ws, r, colStart, periodenNr, cellSaldoVorLC, cellSaldoVorEUR, true, rateAddr)
	if err != nil {
		return err
	}
	r = totalsRow2 + 2

	// Set Durchschnittskurs-Formel in rateRow (Ohne strukturelle Referenz auf [#Totals])
	_ = f.SetCellFormula(ws, cellName(cValLC, rateRow), fmt.Sprintf(`=ROUND(IFERROR(%s,0),6)`, cellName(colStart+4, totalsRow1)))

	// Ranges für SUMIF aufbauen
	tbl1Typ := fmt.Sprintf("%s:%s", absName(colStart, r1_start+1), absName(colStart, r1_start+5))
	tbl1LC := fmt.Sprintf("%s:%s", absName(colStart+2, r1_start+1), absName(colStart+2, r1_start+5))
	tbl1EUR := fmt.Sprintf("%s:%s", absName(colStart+3, r1_start+1), absName(colStart+3, r1_start+5))

	tbl2Typ := fmt.Sprintf("%s:%s", absName(colStart, r2_start+1), absName(colStart, r2_start+6))
	tbl2LC := fmt.Sprintf("%s:%s", absName(colStart+2, r2_start+1), absName(colStart+2, r2_start+6))
	tbl2EUR := fmt.Sprintf("%s:%s", absName(colStart+3, r2_start+1), absName(colStart+3, r2_start+6))

	// Formeln für Einnahmen-Typen (LC/EUR)
	for i, tStr := range TYPE_NAMES {
		lcFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s)+SUMIF(%s,"%s",%s),2)`, tbl1Typ, tStr, tbl1LC, tbl2Typ, tStr, tbl2LC)
		eurFormula := fmt.Sprintf(`=ROUND(SUMIF(%s,"%s",%s)+SUMIF(%s,"%s",%s),2)`, tbl1Typ, tStr, tbl1EUR, tbl2Typ, tStr, tbl2EUR)

		_ = f.SetCellFormula(ws, cellName(cValLC, typeRows[i]), lcFormula)
		_ = f.SetCellFormula(ws, cellName(cValEUR, typeRows[i]), eurFormula)

		if isFollowUp {
			prevKumLcCell := cellName(prevStartCol+3, typeRows[i])
			prevKumEurCell := cellName(prevStartCol+4, typeRows[i])
			_ = f.SetCellFormula(ws, cellName(cValLC+2, typeRows[i]), fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumLcCell, cellName(cValLC, typeRows[i]), cellName(cValLC, typeRows[i])))
			_ = f.SetCellFormula(ws, cellName(cValEUR+2, typeRows[i]), fmt.Sprintf(`=IFERROR(ROUND(%s + %s, 2), %s)`, prevKumEurCell, cellName(cValEUR, typeRows[i]), cellName(cValEUR, typeRows[i])))
		} else {
			_ = f.SetCellFormula(ws, cellName(cValLC+2, typeRows[i]), lcFormula)
			_ = f.SetCellFormula(ws, cellName(cValEUR+2, typeRows[i]), eurFormula)
		}
	}

	return nil
}

// createEinnahmenTabelle erstellt eine der beiden Einnahmentabellen.
func (g *Generator) createEinnahmenTabelle(
	ws string,
	startRow int,
	colStart int,
	periodenNr int,
	saldoLCAddr string,
	saldoEURAddr string,
	isWK bool,
	avgRateAddr string,
) (string, int, error) {
	f := g.file
	tblName := fmt.Sprintf("Einnahmen_%d", periodenNr)
	if isWK {
		tblName = fmt.Sprintf("Einnahmen_WK_%d", periodenNr)
	}

	dataRows := 5
	if isWK {
		dataRows = 6
	}

	// Write headers (Typ, Geber, Einnahmen (LC), Einnahmen (EUR), Kurs)
	headers := []string{"Typ", "Geber", "Einnahmen (LC)", "Einnahmen (EUR)", "Kurs"}
	for i, h := range headers {
		cell := cellName(colStart+i, startRow)
		_ = f.SetCellValue(ws, cell, h)
	}

	// Format header row (startRow)
	headerOpts := StyleOptions{
		Bold:         true,
		FillColor:    COLOR_HEADER,
		HAlign:       "center",
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
	}
	for i := 0; i < 5; i++ {
		_ = g.setStyle(ws, cellName(colStart+i, startRow), cellName(colStart+i, startRow), headerOpts)
	}

	// Add data body range to VSTACK lists
	dataRangeEinn := fmt.Sprintf("'%s'!%s:%s", ws, absName(colStart, startRow+1), absName(colStart+4, startRow+dataRows))
	if isWK {
		g.rangesEinnahmen2 = append(g.rangesEinnahmen2, dataRangeEinn)
	} else {
		g.rangesEinnahmen1 = append(g.rangesEinnahmen1, dataRangeEinn)
	}

	// Tabelle anlegen
	err := f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(colStart, startRow), cellName(colStart+4, startRow+dataRows)),
		Name:           tblName,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})
	if err != nil {
		return "", 0, err
	}

	saldoLabel := "Saldo des Vorprojekts"
	if periodenNr > 1 {
		saldoLabel = "Saldo der Vorperiode"
	}

	type RowData struct {
		Typ, Geber string
		LC, EUR    string
	}

	var rows []RowData
	if !isWK {
		rows = []RowData{
			{Typ: saldoLabel, Geber: "", LC: fmt.Sprintf("=ROUND(%s,2)", saldoLCAddr), EUR: fmt.Sprintf("=ROUND(%s,2)", saldoEURAddr)},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
		}
	} else {
		// For EUR formulas in WK table, we reference the LC column directly to avoid [@...] syntax
		// The EUR formula string will be formatted in the loop below where we have the specific row index
		rows = []RowData{
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
			{Typ: "", Geber: "", LC: "", EUR: ""},
		}
	}

	for i, rdata := range rows {
		row := startRow + 1 + i
		_ = f.SetCellValue(ws, cellName(colStart, row), rdata.Typ)
		_ = f.SetCellValue(ws, cellName(colStart+1, row), rdata.Geber)

		if rdata.LC != "" {
			if strings.HasPrefix(rdata.LC, "=") {
				_ = f.SetCellFormula(ws, cellName(colStart+2, row), rdata.LC)
			} else {
				_ = f.SetCellValue(ws, cellName(colStart+2, row), rdata.LC)
			}
		}

		if isWK {
			// EUR = ROUND(LC / AvgRate, 2)
			_ = f.SetCellFormula(ws, cellName(colStart+3, row), fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, cellName(colStart+2, row), avgRateAddr))
		} else if rdata.EUR != "" {
			if strings.HasPrefix(rdata.EUR, "=") {
				_ = f.SetCellFormula(ws, cellName(colStart+3, row), rdata.EUR)
			} else {
				_ = f.SetCellValue(ws, cellName(colStart+3, row), rdata.EUR)
			}
		}

		// Kurs Formula
		_ = f.SetCellFormula(ws, cellName(colStart+4, row), fmt.Sprintf(`=ROUND(IFERROR(%s/%s,0),6)`, cellName(colStart+2, row), cellName(colStart+3, row)))
	}

	// Formatierung für Datenbereich
	for i := 0; i < dataRows; i++ {
		row := startRow + 1 + i
		isSaldo := i == 0 && !isWK

		var baseOpts StyleOptions
		if isSaldo {
			baseOpts = StyleOptions{
				Italic:       true,
				FillColor:    COLOR_WHITE,
				BorderTop:    1,
				BorderBottom: 1,
				BorderLeft:   1,
				BorderRight:  1,
				BorderColor:  "D3D3D3",
			}
		} else {
			baseOpts = StyleOptions{
				BorderTop:    1,
				BorderBottom: 1,
				BorderLeft:   1,
				BorderRight:  1,
				BorderColor:  "D3D3D3",
			}
		}

		// Typ (colStart)
		typOpts := baseOpts
		typOpts.HAlign = "left"
		if !isSaldo {
			typOpts.FillColor = COLOR_INPUT
		}
		_ = g.setStyle(ws, cellName(colStart, row), cellName(colStart, row), typOpts)

		// Geber (colStart+1)
		gebOpts := baseOpts
		gebOpts.HAlign = "center"
		if !isSaldo {
			gebOpts.FillColor = COLOR_INPUT
		}
		_ = g.setStyle(ws, cellName(colStart+1, row), cellName(colStart+1, row), gebOpts)

		// LC (colStart+2)
		lcOpts := baseOpts
		lcOpts.HAlign = "right"
		lcOpts.NumFormat = "#,##0.00"
		if !isSaldo {
			lcOpts.FillColor = COLOR_INPUT
		}
		_ = g.setStyle(ws, cellName(colStart+2, row), cellName(colStart+2, row), lcOpts)

		// EUR (colStart+3)
		eurOpts := baseOpts
		eurOpts.HAlign = "right"
		eurOpts.NumFormat = `#,##0.00" €"`
		if !isSaldo && !isWK {
			eurOpts.FillColor = COLOR_INPUT
		} else {
			eurOpts.FillColor = COLOR_WHITE
		}
		_ = g.setStyle(ws, cellName(colStart+3, row), cellName(colStart+3, row), eurOpts)

		// Kurs (colStart+4)
		kursOpts := baseOpts
		kursOpts.HAlign = "right"
		kursOpts.NumFormat = "0.000000"
		kursOpts.FillColor = COLOR_WHITE
		_ = g.setStyle(ws, cellName(colStart+4, row), cellName(colStart+4, row), kursOpts)
	}

	// Datenvalidierung hinzufügen
	dvTyp := excelize.NewDataValidation(true)
	dvTyp.Sqref = fmt.Sprintf("%s:%s", cellName(colStart, startRow+1), cellName(colStart, startRow+dataRows))
	dvTyp.SetDropList([]string{saldoLabel, "Eigenmittel", "Drittmittel", "KMW-Mittel", "Zinsertraege"})
	_ = f.AddDataValidation(ws, dvTyp)

	dvGeber := excelize.NewDataValidation(true)
	dvGeber.Sqref = fmt.Sprintf("%s:%s", cellName(colStart+1, startRow+1), cellName(colStart+1, startRow+dataRows))
	dvGeber.Type = "list"
	dvGeber.Formula1 = "=" + FB_NAME_GEBER_LIST
	_ = f.AddDataValidation(ws, dvGeber)

	// Totals Row
	totalsRow := startRow + dataRows + 1
	_ = f.SetCellValue(ws, cellName(colStart, totalsRow), "Gesamteinnahmen in Periode")
	if isWK {
		_ = f.SetCellValue(ws, cellName(colStart, totalsRow), "Gesamt (Durchschnittskurs)")
	}
	_ = f.SetCellValue(ws, cellName(colStart+1, totalsRow), "Durchschnittskurs:")

	_ = f.SetCellFormula(ws, cellName(colStart+2, totalsRow), fmt.Sprintf("=ROUND(SUBTOTAL(109,%s:%s),2)", absName(colStart+2, startRow+1), absName(colStart+2, startRow+dataRows)))
	_ = f.SetCellFormula(ws, cellName(colStart+3, totalsRow), fmt.Sprintf("=ROUND(SUBTOTAL(109,%s:%s),2)", absName(colStart+3, startRow+1), absName(colStart+3, startRow+dataRows)))

	_ = f.SetCellFormula(ws, cellName(colStart+4, totalsRow), fmt.Sprintf("=ROUND(IFERROR(%s/%s,0),6)", cellName(colStart+2, totalsRow), cellName(colStart+3, totalsRow)))

	// Style totals row
	_ = g.setStyle(ws, cellName(colStart, totalsRow), cellName(colStart, totalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_WHITE, HAlign: "left", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(colStart+1, totalsRow), cellName(colStart+1, totalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_WHITE, HAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(colStart+2, totalsRow), cellName(colStart+2, totalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_WHITE, HAlign: "right", NumFormat: "#,##0.00", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(colStart+3, totalsRow), cellName(colStart+3, totalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_WHITE, HAlign: "right", NumFormat: `#,##0.00" €"`, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})
	_ = g.setStyle(ws, cellName(colStart+4, totalsRow), cellName(colStart+4, totalsRow), StyleOptions{
		Bold: true, FillColor: COLOR_WHITE, HAlign: "right", NumFormat: "0.000000", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
	})

	return tblName, totalsRow, nil
}

// ─── ZEICHEN-HILFSFUNKTIONEN ──────────────────────────────────────────────────

func (g *Generator) fbSetupColumnWidths(ws string, colStart int) {
	g.setColWidth(ws, colStart, 30.43)
	g.setColWidth(ws, colStart+1, 24.71)
	g.setColWidth(ws, colStart+2, 24.71)
	g.setColWidth(ws, colStart+3, 24.71)
	g.setColWidth(ws, colStart+4, 24.71)
}

func (g *Generator) drawSeparatorArrow(ws string, row int, col int) error {
	cell := cellName(col, row)
	arrowOpts := StyleOptions{
		Size:      24.0,
		Bold:      true,
		FontColor: "808080",
		HAlign:    "center",
		VAlign:    "center",
	}
	return g.setValue(ws, cell, "➤", arrowOpts)
}

func (g *Generator) drawMergedCell(
	ws string,
	row int,
	col1, col2 int,
	text string,
	bold bool,
	bgColor string,
	italic bool,
) error {
	f := g.file
	hCell := cellName(col1, row)
	vCell := cellName(col2, row)

	err := f.MergeCell(ws, hCell, vCell)
	if err != nil {
		return err
	}
	err = f.SetCellValue(ws, hCell, text)
	if err != nil {
		return err
	}

	opts := StyleOptions{
		Bold:      bold,
		Italic:    italic,
		HAlign:    "left",
		VAlign:    "center",
		FillColor: bgColor,
	}
	return g.setStyle(ws, hCell, vCell, opts)
}

func (g *Generator) fbDrawSectionHeader(
	ws string,
	row int,
	col1, col2 int,
	text string,
) error {
	f := g.file
	hCell := cellName(col1, row)
	vCell := cellName(col2, row)

	err := f.MergeCell(ws, hCell, vCell)
	if err != nil {
		return err
	}
	err = f.SetCellValue(ws, hCell, text)
	if err != nil {
		return err
	}

	opts := StyleOptions{
		Bold:         true,
		FillColor:    COLOR_HEADER,
		HAlign:       "left",
		VAlign:       "center",
		BorderTop:    2,
		BorderBottom: 1,
		BorderColor:  "808080",
	}
	return g.setStyle(ws, hCell, vCell, opts)
}

func (g *Generator) fbDrawColumnHeaders(
	ws string,
	row int,
	col1 int,
	headers []string,
) error {
	f := g.file

	// Left header Typ / ID
	lblCell := cellName(col1, row)
	_ = f.SetCellValue(ws, lblCell, "Typ / ID")
	lblOpts := StyleOptions{
		Bold:         true,
		FillColor:    COLOR_HEADER,
		HAlign:       "left",
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  "808080",
	}
	_ = g.setStyle(ws, lblCell, lblCell, lblOpts)

	// Columns headers
	for i, h := range headers {
		cell := cellName(col1+1+i, row)
		_ = f.SetCellValue(ws, cell, h)
		valOpts := StyleOptions{
			Bold:         true,
			FillColor:    COLOR_HEADER,
			HAlign:       "center",
			VAlign:       "center",
			WrapText:     true,
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  "808080",
		}
		_ = g.setStyle(ws, cell, cell, valOpts)
	}

	return nil
}

func (g *Generator) fbDrawEinnahmenRow(
	ws string,
	row int,
	col1 int,
	label string,
) error {
	f := g.file
	lblCell := cellName(col1, row)
	_ = f.SetCellValue(ws, lblCell, label)
	_ = g.setStyle(ws, lblCell, lblCell, StyleOptions{
		HAlign:       "left",
		VAlign:       "center",
		FillColor:    COLOR_WHITE,
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  "D3D3D3",
	})

	for i := 0; i < 4; i++ {
		cell := cellName(col1+1+i, row)
		numFormat := "#,##0.00"
		if i%2 == 1 {
			numFormat = `#,##0.00" €"`
		}
		_ = g.setStyle(ws, cell, cell, StyleOptions{
			HAlign:       "right",
			VAlign:       "center",
			FillColor:    COLOR_WHITE,
			NumFormat:    numFormat,
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  "D3D3D3",
		})
	}
	return nil
}

func (g *Generator) fbDrawTotalRow(
	ws string,
	row int,
	col1 int,
	label string,
	sumRanges []string,
) error {
	f := g.file
	lblCell := cellName(col1, row)
	_ = f.SetCellValue(ws, lblCell, label)
	_ = g.setStyle(ws, lblCell, lblCell, StyleOptions{
		Bold:         true,
		HAlign:       "left",
		VAlign:       "center",
		FillColor:    COLOR_TOTAL,
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  "808080",
	})

	for i, rStr := range sumRanges {
		cell := cellName(col1+1+i, row)
		_ = f.SetCellFormula(ws, cell, fmt.Sprintf("=ROUND(SUM(%s),2)", rStr))

		numFormat := "#,##0.00"
		if i%2 == 1 {
			numFormat = `#,##0.00" €"`
		}
		_ = g.setStyle(ws, cell, cell, StyleOptions{
			Bold:         true,
			HAlign:       "right",
			VAlign:       "center",
			FillColor:    COLOR_TOTAL,
			NumFormat:    numFormat,
			BorderTop:    1,
			BorderBottom: 1,
			BorderLeft:   1,
			BorderRight:  1,
			BorderColor:  "808080",
		})
	}
	return nil
}

func (g *Generator) fbDrawInfoRow(
	ws string,
	row int,
	cLabel1 int,
	cValLC int,
	cValEUR int,
	formulaEUR string,
) error {
	f := g.file
	lblCell := cellName(cLabel1, row)
	_ = g.setStyle(ws, lblCell, lblCell, StyleOptions{
		HAlign:       "left",
		VAlign:       "center",
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
	})

	lcCell := cellName(cValLC, row)
	_ = g.setStyle(ws, lcCell, lcCell, StyleOptions{
		HAlign:       "right",
		VAlign:       "center",
		FillColor:    COLOR_INPUT,
		NumFormat:    "#,##0.00",
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
	})

	eurCell := cellName(cValEUR, row)
	_ = f.SetCellFormula(ws, eurCell, formulaEUR)
	_ = g.setStyle(ws, eurCell, eurCell, StyleOptions{
		HAlign:       "right",
		VAlign:       "center",
		FillColor:    COLOR_WHITE,
		NumFormat:    `#,##0.00" €"`,
		BorderBottom: 1,
		BorderColor:  "D3D3D3",
	})

	return nil
}

func (g *Generator) fbDrawDifferenceRow(
	ws string,
	row int,
	cLabel1 int,
	cValLC int,
	cValEUR int,
	saldoLCAddr string,
	saldoEURAddr string,
	rowStart int,
	rowEnd int,
) error {
	f := g.file
	lblCell := cellName(cLabel1, row)
	_ = f.SetCellValue(ws, lblCell, "Differenz (Pruefung):")
	_ = g.setStyle(ws, lblCell, lblCell, StyleOptions{
		Size:      8.0,
		FontColor: "808080",
		HAlign:    "left",
		VAlign:    "center",
	})

	lcCell := cellName(cValLC, row)
	_ = f.SetCellFormula(ws, lcCell, fmt.Sprintf(`=ROUND(IFERROR(%s-SUM(%s:%s),""),2)`, saldoLCAddr, cellName(cValLC, rowStart), cellName(cValLC, rowEnd)))
	_ = g.setStyle(ws, lcCell, lcCell, StyleOptions{
		Size:      8.0,
		FontColor: "808080",
		HAlign:    "right",
		VAlign:    "center",
		NumFormat: "#,##0.00",
	})

	eurCell := cellName(cValEUR, row)
	_ = f.SetCellFormula(ws, eurCell, fmt.Sprintf(`=ROUND(IFERROR(%s-SUM(%s:%s),""),2)`, saldoEURAddr, cellName(cValEUR, rowStart), cellName(cValEUR, rowEnd)))
	_ = g.setStyle(ws, eurCell, eurCell, StyleOptions{
		Size:      8.0,
		FontColor: "808080",
		HAlign:    "right",
		VAlign:    "center",
		NumFormat: `#,##0.00" €"`,
	})

	return nil
}
