package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

const (
	MA_SHEET_NAME  = constants.VPSheetMA
	MA_TAB_COLOR   = "FFFF00" // Gelb
	MA_TABLE_COLS  = 3
	MA_TABLE_SPACE = 1
	MA_START_COL   = 2 // Spalte B
	MA_START_ROW   = 5 // Zeile 5

	MA_PERIOD_COUNT = 18
	MA_TABLE_COUNT  = 54

	MA_CLR_GRAY  = "F2F2F2"
	MA_CLR_INPUT = "FFFAE5"
	MA_CLR_SPENT = "F0F0F0"
	MA_CLR_KMW   = "DCE6F1"
)

var MA_CATEGORIES = ListKostenkategorien

// CreateMittelanforderungSheet initialisiert das Blatt "IV. MA" und zeichnet 18 Perioden.
func (g *Generator) CreateMittelanforderungSheet() error {
	ws := MA_SHEET_NAME
	f := g.file

	_, err := f.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des MA-Blatts: %w", err)
	}

	tabColor := MA_TAB_COLOR
	_ = f.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = f.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	// Auswahlliste "Periode 1..18" in ausgeblendeter Spalte A
	g.maEnsurePeriodList(ws)

	fbExists := true
	if idx, _ := f.GetSheetIndex(constants.VPSheetFINANZBERICHTE); idx == -1 {
		fbExists = false
	}

	for p := 1; p <= MA_PERIOD_COUNT; p++ {
		colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)

		g.maSetupColumnWidths(ws, colS)
		if colS > MA_START_COL {
			_ = g.drawSeparatorArrow(ws, MA_START_ROW-2, colS-1)
		}

		err = g.drawMATable(ws, colS, MA_START_ROW, p, p, fbExists, true)
		if err != nil {
			return fmt.Errorf("fehler beim Zeichnen von MA Periode %d: %w", p, err)
		}
	}

	maDataRows := len(MA_CATEGORIES)
	maBlockHeight := 10 + maDataRows + 8           // header+data+footer
	startRowL2 := MA_START_ROW + maBlockHeight + 2 // +2 for spacing
	startRowL3 := startRowL2 + maBlockHeight + 2

	_ = f.SetCellValue(ws, cellName(MA_START_COL, startRowL2-2), "Zusätzliche Mittelanforderungen (Ausnahme 1)")
	_ = g.setStyle(ws, cellName(MA_START_COL, startRowL2-2), cellName(MA_START_COL+2, startRowL2-2), StyleOptions{Bold: true})

	for p := 1; p <= MA_PERIOD_COUNT; p++ {
		colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
		tableId := p + MA_PERIOD_COUNT

		err = g.drawMATable(ws, colS, startRowL2, tableId, p, fbExists, false)
		if err != nil {
			return fmt.Errorf("fehler beim Zeichnen von MA (Zusatz 1) %d: %w", tableId, err)
		}
	}

	_ = f.SetCellValue(ws, cellName(MA_START_COL, startRowL3-2), "Zusätzliche Mittelanforderungen (Ausnahme 2)")
	_ = g.setStyle(ws, cellName(MA_START_COL, startRowL3-2), cellName(MA_START_COL+2, startRowL3-2), StyleOptions{Bold: true})

	for p := 1; p <= MA_PERIOD_COUNT; p++ {
		colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
		tableId := p + MA_PERIOD_COUNT*2

		err = g.drawMATable(ws, colS, startRowL3, tableId, p, fbExists, false)
		if err != nil {
			return fmt.Errorf("fehler beim Zeichnen von MA (Zusatz 2) %d: %w", tableId, err)
		}
	}

	// Perioden 2–18 gruppieren und zugeklappt ausblenden (nur Inhaltsspalten der Tabelle).
	f = g.file
	for p := 2; p <= MA_PERIOD_COUNT; p++ {
		colS := MA_START_COL + (p-1)*(MA_TABLE_COLS+MA_TABLE_SPACE)
		groupFirst := colS
		groupLast := colS + MA_TABLE_COLS - 1
		for c := groupFirst; c <= groupLast; c++ {
			_ = f.SetColOutlineLevel(ws, colLetter(c), 1)
			_ = f.SetColVisible(ws, colLetter(c), false)
		}
	}

	// Zeilen-Gruppierungen für Ebene 2 und 3 (jeweils mit 1 ungruppierten Abstandzeile dazwischen)
	for r := startRowL2 - 2; r <= startRowL2+maBlockHeight; r++ {
		_ = f.SetRowOutlineLevel(ws, r, 1)
		_ = f.SetRowVisible(ws, r, false)
	}

	for r := startRowL3 - 2; r <= startRowL3+maBlockHeight; r++ {
		_ = f.SetRowOutlineLevel(ws, r, 1)
		_ = f.SetRowVisible(ws, r, false)
	}

	return nil
}

// maEnsurePeriodList is now handled by daten.go
// but we keep a dummy or just remove it.
func (g *Generator) maEnsurePeriodList(ws string) {
	// Das Dropdown verweist nun auf das Daten-Blatt.
	// Die Erstellung der Liste passiert in CreateDatenSheet()
}

func (g *Generator) maSetupColumnWidths(ws string, colS int) {
	g.setColWidth(ws, colS, 36.80) // 15% breiter als vorher (32.00)
	g.setColWidth(ws, colS+1, 24.71)
	g.setColWidth(ws, colS+2, 24.71)
}

func (g *Generator) drawMATable(ws string, colS, startR, tableId, periodNr int, fbExists, isStandard bool) error {
	f := g.file
	cLbl := colS
	cLC := colS + 1
	cEUR := colS + 2

	// Periode rückt eine Zeile nach oben (Zeile 4), damit Von/Bis/Zeitraum/Kurs
	// darunter passen und der Tabellenkopf weiterhin auf Zeile 9 bleibt.
	r := startR - 1

	// ─── Zeile 1: Periode-Kopfzeile (Dropdown 1..18) ──────────────────────────
	lblPer := cellName(cLbl, r)
	_ = f.SetCellValue(ws, lblPer, "Periode:")
	_ = g.setStyle(ws, lblPer, lblPer, StyleOptions{Bold: true, HAlign: "left", VAlign: "center"})

	rngPerStart := cellName(cLC, r)
	rngPerEnd := cellName(cEUR, r)
	_ = f.MergeCell(ws, rngPerStart, rngPerEnd)

	_ = f.SetCellValue(ws, rngPerStart, fmt.Sprintf("Periode %d", periodNr))
	_ = g.setStyle(ws, rngPerStart, rngPerEnd, StyleOptions{
		HAlign: "center", VAlign: "center", FillColor: MA_CLR_GRAY, BorderBottom: 1, BorderColor: "D3D3D3",
	})
	r++

	// ─── Zeile 2/3: Zeitraum (Von / Bis) ──────────────────────────────────────
	vonRow := r
	for i, zlbl := range []string{"Von:", "Bis:"} {
		lblZeit := cellName(cLbl, r)
		_ = f.SetCellValue(ws, lblZeit, zlbl)
		_ = g.setStyle(ws, lblZeit, lblZeit, StyleOptions{Bold: true, HAlign: "left", VAlign: "center"})

		rngZeitStart := cellName(cLC, r)
		rngZeitEnd := cellName(cEUR, r)
		_ = f.MergeCell(ws, rngZeitStart, rngZeitEnd)
		_ = g.setStyle(ws, rngZeitStart, rngZeitEnd, StyleOptions{
			HAlign: "center", VAlign: "center", FillColor: MA_CLR_INPUT, BorderBottom: 1, BorderColor: "D3D3D3", NumFmtID: 14,
		})
		if i == 0 {
			_ = g.bindInputField(ws, r, cLC, FieldMAVon(tableId))
		} else {
			_ = g.bindInputField(ws, r, cLC, FieldMABis(tableId))
		}
		r++
	}

	// ─── Zeile 4: Zeitraum (Monate, berechnet) ────────────────────────────────
	lblZr := cellName(cLbl, r)
	_ = f.SetCellValue(ws, lblZr, "Zeitraum:")
	_ = g.setStyle(ws, lblZr, lblZr, StyleOptions{Bold: true, HAlign: "left", VAlign: "center"})

	zrStart := cellName(cLC, r)
	zrEnd := cellName(cEUR, r)
	_ = f.MergeCell(ws, zrStart, zrEnd)
	_ = f.SetCellFormula(ws, zrStart, fmt.Sprintf(
		`=IF(OR(%s="",%s=""),"",DATEDIF(%s,%s,"m")+1)`,
		cellName(cLC, vonRow), cellName(cLC, vonRow+1), cellName(cLC, vonRow), cellName(cLC, vonRow+1)))
	_ = g.setStyle(ws, zrStart, zrEnd, StyleOptions{
		HAlign: "center", VAlign: "center", FillColor: MA_CLR_GRAY, BorderBottom: 1, BorderColor: "D3D3D3", NumFormat: `0" Monate"`,
	})
	r++

	// ─── Zeile 5: OANDA-Kurs-Eingabe (benannt Inp_MA_Kurs_<p>) ────────────────────
	rateAddr := absName(cLC, r)
	maKursName := FieldMAKurs(tableId).NamedRange

	lblRate := cellName(cLbl, r)
	_ = f.SetCellValue(ws, lblRate, "OANDA-Kurs:")
	_ = g.setStyle(ws, lblRate, lblRate, StyleOptions{Bold: true, HAlign: "left", VAlign: "center"})

	rngRateStart := cellName(cLC, r)
	rngRateEnd := cellName(cEUR, r)
	_ = f.MergeCell(ws, rngRateStart, rngRateEnd)
	_ = g.setStyle(ws, rngRateStart, rngRateEnd, StyleOptions{
		HAlign: "center", VAlign: "center", FillColor: MA_CLR_INPUT, BorderBottom: 1, BorderColor: "D3D3D3", NumFormat: "0.0000",
	})
	_ = g.bindInputField(ws, r, cLC, FieldMAKurs(tableId))
	r++ // Tabellenkopf folgt direkt (Periode/Von/Bis/Kurs belegen Zeilen 5–8)

	// ─── Zeile 9: Tabelle MA_<p> (Kostenkategorie | LC | EUR) ──────────────────
	maHdrRow := r

	_ = f.SetCellValue(ws, cellName(cLbl, maHdrRow), "Kostenkategorie")
	_ = f.SetCellValue(ws, cellName(cLC, maHdrRow), "Angefordert (LC)")
	_ = f.SetCellValue(ws, cellName(cEUR, maHdrRow), "Angefordert (EUR)")

	_ = g.setStyle(ws, cellName(cLbl, maHdrRow), cellName(cEUR, maHdrRow), StyleOptions{
		Bold: true, FillColor: MA_CLR_GRAY, HAlign: "center", VAlign: "center",
		BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080",
	})

	maDataRows := len(MA_CATEGORIES)
	maTotalsRow := maHdrRow + maDataRows + 1

	// Add data body range to MA list for VSTACK
	dataRangeMA := fmt.Sprintf("'%s'!%s:%s", ws, absName(cLbl, maHdrRow+1), absName(cEUR, maHdrRow+maDataRows))
	g.rangesMA = append(g.rangesMA, dataRangeMA)

	for i := 0; i < maDataRows; i++ {
		row := maHdrRow + 1 + i
		labelVal := ""
		if i < len(MA_CATEGORIES) {
			labelVal = MA_CATEGORIES[i]
		}
		_ = f.SetCellValue(ws, cellName(cLbl, row), labelVal)
		_ = f.SetCellFormula(ws, cellName(cEUR, row), fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, cellName(cLC, row), maKursName))
		g.upsertNamedFormula(FieldMAKatEUR(tableId, i+1), fmt.Sprintf("'%s'!%s", ws, absName(cEUR, row)))

		_ = g.setStyle(ws, cellName(cLbl, row), cellName(cLbl, row), StyleOptions{
			HAlign: "left", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})
		_ = g.setStyle(ws, cellName(cLC, row), cellName(cLC, row), StyleOptions{
			HAlign: "right", VAlign: "center", FillColor: MA_CLR_INPUT, NumFormat: "#,##0.00",
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})
		_ = g.bindInputField(ws, row, cLC, FieldMAKat(tableId, i+1))

		_ = g.setStyle(ws, cellName(cEUR, row), cellName(cEUR, row), StyleOptions{
			HAlign: "right", VAlign: "center", NumFormat: `#,##0.00" €"`,
			BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3",
		})
	}

	// Totals
	_ = f.SetCellValue(ws, cellName(cLbl, maTotalsRow), "SUMME")
	_ = f.SetCellFormula(ws, cellName(cLC, maTotalsRow), fmt.Sprintf(`=ROUND(SUM(%s:%s),2)`, cellName(cLC, maHdrRow+1), cellName(cLC, maHdrRow+maDataRows)))
	g.upsertNamedFormula(FieldMASumLC(tableId), fmt.Sprintf("'%s'!%s", ws, absName(cLC, maTotalsRow)))
	_ = f.SetCellFormula(ws, cellName(cEUR, maTotalsRow), fmt.Sprintf(`=ROUND(SUM(%s:%s),2)`, cellName(cEUR, maHdrRow+1), cellName(cEUR, maHdrRow+maDataRows)))
	g.upsertNamedFormula(FieldMASumEUR(tableId), fmt.Sprintf("'%s'!%s", ws, absName(cEUR, maTotalsRow)))

	_ = g.setStyle(ws, cellName(cLbl, maTotalsRow), cellName(cLbl, maTotalsRow), StyleOptions{
		Bold: true, FillColor: MA_CLR_GRAY, HAlign: "left", VAlign: "center",
		BorderTop: 6, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080",
	})
	_ = g.setStyle(ws, cellName(cLC, maTotalsRow), cellName(cLC, maTotalsRow), StyleOptions{
		Bold: true, FillColor: MA_CLR_GRAY, HAlign: "right", VAlign: "center", NumFormat: "#,##0.00",
		BorderTop: 6, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080",
	})
	_ = g.setStyle(ws, cellName(cEUR, maTotalsRow), cellName(cEUR, maTotalsRow), StyleOptions{
		Bold: true, FillColor: MA_CLR_GRAY, HAlign: "right", VAlign: "center", NumFormat: `#,##0.00" €"`,
		BorderTop: 6, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080",
	})

	addrSumGL := absName(cLC, maTotalsRow)
	addrSumGE := absName(cEUR, maTotalsRow)

	r = maTotalsRow + 2 // Summe + Leerzeile

	// ─── Gesamtbedarf an Mitteln ──────────────────────────────────────────────
	_ = f.SetCellValue(ws, cellName(cLbl, r), "Gesamtbedarf an Mitteln:")

	gdLC := cellName(cLC, r)
	_ = f.SetCellFormula(ws, gdLC, fmt.Sprintf(`=ROUND(%s,2)`, addrSumGL))
	_ = g.setStyle(ws, gdLC, gdLC, StyleOptions{Italic: true, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center"})

	gdEUR := cellName(cEUR, r)
	_ = f.SetCellFormula(ws, gdEUR, fmt.Sprintf(`=ROUND(%s,2)`, addrSumGE))
	_ = g.setStyle(ws, gdEUR, gdEUR, StyleOptions{NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center"})
	r++

	// ─── abzueglich Eigenmittel ───────────────────────────────────────────────
	_ = f.SetCellValue(ws, cellName(cLbl, r), "abzueglich Eigenmittel:")

	eigenLC := cellName(cLC, r)
	_ = g.setStyle(ws, eigenLC, eigenLC, StyleOptions{FillColor: MA_CLR_INPUT, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.bindInputField(ws, r, cLC, FieldMAEigenmittelLC(tableId))
	addrEigenLC := absName(cLC, r)

	eigenEUR := cellName(cEUR, r)
	_ = f.SetCellFormula(ws, eigenEUR, fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, addrEigenLC, rateAddr))
	_ = g.setStyle(ws, eigenEUR, eigenEUR, StyleOptions{NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.bindInputField(ws, r, cEUR, FieldMAEigenmittelEUR(tableId))
	addrEigenEUR := absName(cEUR, r)
	r++

	// ─── abzueglich Drittmittel ───────────────────────────────────────────────
	_ = f.SetCellValue(ws, cellName(cLbl, r), "abzueglich Drittmittel:")

	drittLC := cellName(cLC, r)
	_ = g.setStyle(ws, drittLC, drittLC, StyleOptions{FillColor: MA_CLR_INPUT, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.bindInputField(ws, r, cLC, FieldMADrittmittelLC(tableId))
	addrDrittLC := absName(cLC, r)

	drittEUR := cellName(cEUR, r)
	_ = f.SetCellFormula(ws, drittEUR, fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, addrDrittLC, rateAddr))
	_ = g.setStyle(ws, drittEUR, drittEUR, StyleOptions{NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.bindInputField(ws, r, cEUR, FieldMADrittmittelEUR(tableId))
	addrDrittEUR := absName(cEUR, r)
	r++

	// ─── abzueglich Saldo Vorperiode (FB) ─────────────────────────────────────
	addrSaldoLC := absName(cLC, r)
	saldoLblCell := cellName(cLbl, r)
	saldoLCCell := cellName(cLC, r)

	if isStandard {
		if fbExists {
			safeSaldoVortrag := fmt.Sprintf(`IF(%s="",0,%s)`, DB_NAME_SALDOVORTRAG_LW, DB_NAME_SALDOVORTRAG_LW)
			if tableId == 1 {
				_ = f.SetCellValue(ws, saldoLblCell, "abzueglich Saldo Vorprojekt:")
				_ = f.SetCellFormula(ws, saldoLCCell, fmt.Sprintf(`=ROUND(%s,2)`, safeSaldoVortrag))
			} else {
				_ = f.SetCellValue(ws, saldoLblCell, "abzueglich Saldo Vorperiode (FB):")
				_ = f.SetCellFormula(ws, saldoLCCell, fmt.Sprintf(`=ROUND(IFERROR(FB_SaldoLC_%d,0),2)`, tableId-1))
			}
			_ = g.setStyle(ws, saldoLCCell, saldoLCCell, StyleOptions{Italic: true, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
		} else {
			if tableId == 1 {
				_ = f.SetCellValue(ws, saldoLblCell, "abzueglich Saldo Vorprojekt:")
			} else {
				_ = f.SetCellValue(ws, saldoLblCell, "abzueglich Saldo Vorperiode (FB):")
			}
			_ = g.setStyle(ws, saldoLCCell, saldoLCCell, StyleOptions{FillColor: MA_CLR_INPUT, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
			_ = g.bindInputField(ws, r, cLC, FieldMASaldoLC(tableId))
		}
	} else {
		_ = f.SetCellValue(ws, saldoLblCell, "abzueglich Saldo Vorperiode (manuell):")
		_ = g.setStyle(ws, saldoLCCell, saldoLCCell, StyleOptions{FillColor: MA_CLR_INPUT, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
		_ = g.bindInputField(ws, r, cLC, FieldMASaldoLC(tableId))
	}

	addrSaldoEUR := absName(cEUR, r)
	saldoEURCell := cellName(cEUR, r)
	_ = f.SetCellFormula(ws, saldoEURCell, fmt.Sprintf(`=IFERROR(ROUND(%s/%s,2),0)`, addrSaldoLC, rateAddr))
	_ = g.setStyle(ws, saldoEURCell, saldoEURCell, StyleOptions{NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	r += 2

	// ─── KMW-Mittel Anforderung ───────────────────────────────────────────────
	lblKMW := cellName(cLbl, r)
	_ = f.SetCellValue(ws, lblKMW, "KMW-Mittel Anforderung:")
	_ = g.setStyle(ws, lblKMW, lblKMW, StyleOptions{Bold: true, Size: 12.0, HAlign: "left", VAlign: "center", BorderTop: 6, BorderColor: "808080"})

	kmwLC := cellName(cLC, r)
	_ = f.SetCellFormula(ws, kmwLC, fmt.Sprintf(`=IFERROR(ROUND(%s-%s-%s-%s,2),0)`, addrSumGL, addrEigenLC, addrDrittLC, addrSaldoLC))
	_ = g.setStyle(ws, kmwLC, kmwLC, StyleOptions{Bold: true, Size: 12.0, FillColor: MA_CLR_KMW, NumFormat: "#,##0.00", HAlign: "right", VAlign: "center", BorderTop: 6, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080"})
	_ = g.bindInputField(ws, r, cLC, FieldMAKmwLC(tableId))

	kmwEUR := cellName(cEUR, r)
	_ = f.SetCellFormula(ws, kmwEUR, fmt.Sprintf(`=IFERROR(ROUND(%s-%s-%s-%s,2),0)`, addrSumGE, addrEigenEUR, addrDrittEUR, addrSaldoEUR))
	_ = g.setStyle(ws, kmwEUR, kmwEUR, StyleOptions{Bold: true, Size: 12.0, FillColor: MA_CLR_KMW, NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center", BorderTop: 6, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "808080"})
	_ = g.bindInputField(ws, r, cEUR, FieldMAKmwEUR(tableId))
	r += 2

	// ─── Manueller Betrag (EUR) – zwei Zeilen unter KMW-Mittel Anforderung ────
	_ = f.MergeCell(ws, cellName(cLbl, r), cellName(cLC, r))
	_ = f.SetCellValue(ws, cellName(cLbl, r), "Manueller Betrag (EUR):")
	_ = g.setStyle(ws, cellName(cLbl, r), cellName(cLC, r), StyleOptions{HAlign: "left", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.setStyle(ws, cellName(cEUR, r), cellName(cEUR, r), StyleOptions{FillColor: MA_CLR_INPUT, NumFormat: `#,##0.00" €"`, HAlign: "right", VAlign: "center", BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: "D3D3D3"})
	_ = g.bindInputField(ws, r, cEUR, FieldMAManBetrag(tableId))

	return nil
}
