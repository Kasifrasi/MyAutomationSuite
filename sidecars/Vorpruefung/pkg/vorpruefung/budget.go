package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

const (
	// Spalten (Financing + Ausgaben)
	BudgetColLabel  = 2  // B
	BudgetColID     = 3  // C
	BudgetColPos    = 4  // D
	BudgetColLC     = 5  // E
	BudgetColY1     = 6  // F
	BudgetColY2     = 7  // G
	BudgetColY3     = 8  // H
	BudgetColEUR    = 9  // I
	BudgetColGap    = 10 // J (Leer-Trenner)
	BudgetColStatus = 11 // K
	BudgetColCheck  = 12 // L
	BudgetColBegr1  = 11 // K  (Begründung links = Status-Spalte)
	BudgetColBegr2  = 13 // M  (Begründung rechts)

	BudgetColHelpLC  = 15 // O (ausgeblendet)
	BudgetColHelpEUR = 16 // P (ausgeblendet)

	BudgetColListGeber = 18 // R (ausgeblendet)
	BudgetColListID    = 19 // S (ausgeblendet)

	// Feste Zeilen
	BudgetRowTitle     = 2
	BudgetRowSection1  = 4 // Abschnittsheader + Kurs-Zeile (gleiche Zeile)
	BudgetRowFinanzHdr = 5
	BudgetRowEigenHead = 6
	BudgetRowEigen     = 7
	BudgetRowDrittHead = 9
	BudgetRowDritt     = 10
	BudgetRowKMWHead   = 12
	BudgetRowKMW       = 13
	BudgetRowGesamt    = 15
	BudgetRowSection2  = 17
	BudgetRowAusgHdr   = 18
	BudgetRowAusgStart = 19

	// Reserve-Box (Spalte K, Zeilen 2–6)
	BudgetReserveRowTitle  = 2
	BudgetReserveRowAmount = 3
	BudgetReserveRowLabel  = 4
	BudgetReserveRowCheck  = 5
	BudgetReserveRowStatus = 6

	// Begründungsfeld (Spalten K–M, Zeilen 8–12)
	BudgetBegrHdrRow   = 8
	BudgetBegrAreaTop  = 9
	BudgetBegrAreaRows = 4

	// Hilfszeilen (immer: 4 + Index in ListKostenkategorien; Reserve = Index 7)
	// BudgetHelpReserveRow = 11 (4 + 7)
	// BudgetHelpTotalRow   = 12 (4 + 8 = 4 + len(ListKostenkategorien))
	BudgetHelpReserveRow = 11
	BudgetHelpTotalRow   = 12

	// Spaltenbreiten
	BudgetWLabel  = 25.0
	BudgetWID     = 8.0
	BudgetWPos    = 100.0
	BudgetWLC     = 18.0
	BudgetWYear   = 14.0
	BudgetWEUR    = 18.0
	BudgetWGap    = 3.0
	BudgetWStatus = 37.0
	BudgetWCheck  = 24.0
	BudgetWBegr   = 24.0

	// Sheet
	BudgetSheetName = constants.VPSheetBUDGET
	BudgetTabColor  = "5B9BD5"

	// Benannte Bereiche für Listenbereiche (kein Zellbezug, kein Inp_/Out_-Präfix)
	BudgetNameGeberList = "Geber_Liste"
	BudgetNameIDList    = "Budget_ID_Liste"
	BudgetNameReserve   = "Inp_Budget_ReserveFreigabe"

	// Tabellennamen (intern und von anderen Sheets verwendet)
	BudgetTableAusg = "TblBudgetAusgaben"
	BudgetTableDrit = "TblDrittmittel"
)

var BudgetYears = []string{"Jahr 1", "Jahr 2", "Jahr 3"}

// ─── Teil B: Layout-Dokumentation ────────────────────────────────────────────
/*
  LAYOUT BUDGET:
  | Zeile | B (Label)              | C (ID)  | D (Pos)  | E (LC)   | F–H (J1–3) | I (EUR)  | K (Status) | M (Begr.) |
  |-------|------------------------|---------|----------|----------|------------|----------|------------|-----------|
  |   2   | KERNDATEN BUDGET (Titel, merged B:I)                                  | Reserve Freigabe (K)            |
  |   4   | 1. GEPLANTE EINNAHMEN (Section Hdr, B:I)  | Budget-Kurs (G4:H4)       | Reserve (K3–K6)                 |
  |   5   | Finanzierungsquelle | B. (LC) | J1 | J2 | J3 | B. (EUR)                Begründung (K8–M12)             |
  |   6   | 1.1 Eigenmittel (SubHdr B:I)                                                                              |
  |   7   | Eigenmittel            |         |          | [Inp] | [Inp] | [Inp] | [Inp] | [Inp]                       |
  |   9   | 1.2 Drittmittel (SubHdr)                                              | Drittmittel-Tabelle (K17:M28)   |
  |  10   | Drittmittel (Summe)   |         |          |[Σ LC]|[Inp]|[Inp]|[Inp]|[Σ EUR]                           |
  |  12   | 1.3 KMW-Mittel (SubHdr)                                                                                   |
  |  13   | KMW-Mittel            |         |          | [Inp] | [Inp] | [Inp] | [Inp] | [Inp]                       |
  |  15   | GESAMTPROJEKTMITTEL (Total)                                                                               |
  |  17   | 2. GEPLANTE AUSGABEN (Section Hdr)                                                                        |
  |  18   | Kostenkategorie | ID | Kostenposition | LC | J1 | J2 | J3 | EUR  (TblBudgetAusgaben, dyn. Zeilen)         |
  | 19..N | [Ausgaben-Daten (kategorie-gesteuert)]                                                                    |
  | N+1   | Geplante Gesamtausgaben (Total)                                                                           |
  | N+3   | Budgetprüfung (Checks-Box, merged B:E)                                                                    |
*/

// ─── Hilfs-Struct für dynamische Ausgaben-Zeilen ─────────────────────────────

type budgetDynRows struct {
	AusgDataRows int
	AusgTotal    int // erste Zeile nach den Daten = Summen-Zeile
}

func (g *Generator) budgetComputeDynRows() budgetDynRows {
	n := g.budgetExpenseCount()
	return budgetDynRows{
		AusgDataRows: n,
		AusgTotal:    BudgetRowAusgStart + n,
	}
}

// ─── Teil C: Orchestrator ─────────────────────────────────────────────────────

func (g *Generator) CreateBudgetSheet(reg *TemplateRegistry) error {
	ws := BudgetSheetName
	dyn := g.budgetComputeDynRows()

	_, _ = g.file.NewSheet(ws)
	tabColor := BudgetTabColor
	_ = g.file.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = g.file.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	g.budgetSetupColumns(ws)

	// ── Teil D: Draw ──────────────────────────────────────────────────────────
	if err := g.drawBudgetTitle(ws); err != nil {
		return err
	}
	if err := g.drawBudgetFinancing(ws); err != nil {
		return err
	}
	if err := g.drawBudgetAusgaben(ws, dyn); err != nil {
		return err
	}
	if err := g.drawBudgetDrittmittelTable(ws); err != nil {
		return err
	}
	// Reserve-Betrag und Reserve-Check referenzieren ausschließlich benannte
	// Bereiche (statt fester Zellbezüge auf Hilfsspalte/Status-Spalte).
	reserveEurAddr := bgKostenName("Reserve", "EUR")
	reserveCheckAddr := reg.InputBudgetReserveFreigabe.NamedRange
	g.drawBudgetReserveBox(ws, reserveEurAddr)
	g.drawBudgetBegruendung(ws, reserveCheckAddr)

	// ── Teil E: Bind ──────────────────────────────────────────────────────────
	if err := g.bindBudgetFinancing(ws, reg, dyn); err != nil {
		return err
	}
	if err := g.bindBudgetAusgaben(ws, reg, dyn); err != nil {
		return err
	}
	_ = g.bindInputField(ws, BudgetReserveRowCheck, BudgetColStatus, reg.InputBudgetReserveFreigabe)
	g.upsertNamedRange(reg.OutputBudgetReserveEUR.NamedRange, BudgetColStatus, BudgetReserveRowAmount)

	// ── Abschluß (braucht bindete Adressen) ──────────────────────────────────
	g.styleOuterBorder(ws, BudgetRowTitle, BudgetColLabel, dyn.AusgTotal, BudgetColEUR, 2, BudgetClrBorder)

	// Prüfungs-Formeln referenzieren die benannten Gesamt-Bereiche bzw. – wo kein
	// benannter Bereich existiert (Ausgaben-Summen) – strukturierte Tabellenbezüge.
	incYearsAddr := fmt.Sprintf("%s+%s+%s",
		reg.OutputBudgetGesamtY1.NamedRange,
		reg.OutputBudgetGesamtY2.NamedRange,
		reg.OutputBudgetGesamtY3.NamedRange,
	)
	expYearsAddr := fmt.Sprintf("SUBTOTAL(109,%s[%s])+SUBTOTAL(109,%s[%s])+SUBTOTAL(109,%s[%s])",
		BudgetTableAusg, BudgetYears[0],
		BudgetTableAusg, BudgetYears[1],
		BudgetTableAusg, BudgetYears[2],
	)
	g.drawBudgetChecks(ws,
		dyn.AusgTotal+2,
		reg.OutputBudgetGesamtLC.NamedRange,
		reg.OutputBudgetGesamtEUR.NamedRange,
		incYearsAddr,
		fmt.Sprintf("SUBTOTAL(109,%s[Betrag (LC)])", BudgetTableAusg),
		fmt.Sprintf("SUBTOTAL(109,%s[Betrag (EUR)])", BudgetTableAusg),
		expYearsAddr,
		reg.OutputBudgetWK.NamedRange,
	)

	return nil
}

// ─── Teil D: Draw-Funktionen (nur visuell) ───────────────────────────────────

func (g *Generator) budgetSetupColumns(ws string) {
	g.setColWidth(ws, BudgetColLabel, BudgetWLabel)
	g.setColWidth(ws, BudgetColID, BudgetWID)
	g.setColWidth(ws, BudgetColPos, BudgetWPos)
	g.setColWidth(ws, BudgetColLC, BudgetWLC)
	g.setColWidth(ws, BudgetColY1, BudgetWYear)
	g.setColWidth(ws, BudgetColY2, BudgetWYear)
	g.setColWidth(ws, BudgetColY3, BudgetWYear)
	g.setColWidth(ws, BudgetColEUR, BudgetWEUR)
	g.setColWidth(ws, BudgetColGap, BudgetWGap)
	g.setColWidth(ws, BudgetColStatus, BudgetWStatus)
	g.setColWidth(ws, BudgetColCheck, BudgetWCheck)
	g.setColWidth(ws, BudgetColBegr2, BudgetWBegr)
}

func (g *Generator) drawBudgetTitle(ws string) error {
	for c := BudgetColLabel; c <= BudgetColEUR; c++ {
		g.setStyle(ws, cellName(c, BudgetRowTitle), cellName(c, BudgetRowTitle), BudgetTitleStyle)
	}
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowTitle), "I. KERNDATEN BUDGET", BudgetTitleStyle)
	_ = g.file.SetRowHeight(ws, BudgetRowTitle, 24)
	return nil
}

func (g *Generator) drawBudgetFinancing(ws string) error {
	f := g.file

	// Section 1 header (= row mit Kurs-Zelle)
	g.bgSectionHeader(ws, BudgetRowSection1, "1. GEPLANTE EINNAHMEN / FINANZIERUNG")
	g.setValue(ws, cellName(BudgetColY2, BudgetRowSection1), "€ Budget-Kurs:", StyleOptions{
		Size: 9, HAlign: "right", VAlign: "center", FillColor: BudgetClrHeader,
		BorderTop: 2, BorderBottom: 1, BorderColor: BudgetClrBorder,
	})
	g.setStyle(ws, cellName(BudgetColY3, BudgetRowSection1), cellName(BudgetColY3, BudgetRowSection1), StyleOptions{
		NumFormat: BudgetFmtRate, Italic: true,
		BorderTop: 2, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BudgetClrBorder,
	})

	// Spaltenüberschriften
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowFinanzHdr), "Finanzierungsquelle", StyleOptions{})
	g.budgetValueHeaderCells(ws, BudgetRowFinanzHdr)
	g.bgTableHeader(ws, BudgetRowFinanzHdr, BudgetColLabel, BudgetColEUR)

	// 1.1 Eigenmittel
	g.bgSubHeader(ws, BudgetRowEigenHead, "1.1 Eigenmittel")
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowEigen), "Eigenmittel", StyleOptions{Size: 10})
	g.drawBudgetYearInputRow(ws, BudgetRowEigen)
	_ = f.SetRowHeight(ws, BudgetRowEigen, 22)

	// 1.2 Drittmittel
	g.bgSubHeader(ws, BudgetRowDrittHead, "1.2 Drittmittel")
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowDritt), "Drittmittel (Summe):", StyleOptions{Size: 10})
	g.setValue(ws, cellName(BudgetColPos, BudgetRowDritt), "Aufstellung je Geber → Tabelle rechts",
		StyleOptions{Size: 8, Italic: true, FontColor: BudgetClrResTxt})
	g.bgInput(ws, cellName(BudgetColY1, BudgetRowDritt), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColY2, BudgetRowDritt), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColY3, BudgetRowDritt), BudgetFmtLC)
	_ = f.SetRowHeight(ws, BudgetRowDritt, 22)

	// 1.3 KMW-Mittel
	g.bgSubHeader(ws, BudgetRowKMWHead, "1.3 KMW-Mittel")
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowKMW), "KMW-Mittel", StyleOptions{Size: 10})
	g.drawBudgetYearInputRow(ws, BudgetRowKMW)
	_ = f.SetRowHeight(ws, BudgetRowKMW, 22)

	// GESAMTPROJEKTMITTEL
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowGesamt), "GESAMTPROJEKTMITTEL", StyleOptions{})
	g.bgTotalRow(ws, BudgetRowGesamt, BudgetColLabel, BudgetColEUR)

	return nil
}

func (g *Generator) drawBudgetAusgaben(ws string, dyn budgetDynRows) error {
	f := g.file

	// Section 2 header
	g.bgSectionHeader(ws, BudgetRowSection2, "2. GEPLANTE AUSGABEN")

	// Spaltenüberschriften
	g.setValue(ws, cellName(BudgetColLabel, BudgetRowAusgHdr), "Kostenkategorie", StyleOptions{})
	g.setValue(ws, cellName(BudgetColID, BudgetRowAusgHdr), "ID", StyleOptions{})
	g.setValue(ws, cellName(BudgetColPos, BudgetRowAusgHdr), "Kostenposition", StyleOptions{})
	g.budgetValueHeaderCells(ws, BudgetRowAusgHdr)

	catArrayStr := `{"` + strings.Join(ListKostenkategorien, `","`) + `"}`
	for i := 0; i < dyn.AusgDataRows; i++ {
		row := BudgetRowAusgStart + i
		_ = f.SetRowHeight(ws, row, 30)

		if g.cfg.ExpensePositionsCount > 0 {
			g.setStyle(ws, cellName(BudgetColLabel, row), cellName(BudgetColLabel, row), BudgetCatCellStyle)
			g.setStyle(ws, cellName(BudgetColID, row), cellName(BudgetColID, row), BudgetIDCellStyle)
			g.setStyle(ws, cellName(BudgetColPos, row), cellName(BudgetColPos, row), BudgetPosCellStyle)
		} else {
			g.setValue(ws, cellName(BudgetColLabel, row), ListKostenkategorien[i], BudgetCatCellStyle)
			formulaID := fmt.Sprintf(
				`=IF(B%d="","",MATCH(B%d,%s,0)&"."&COUNTIF(B$%d:B%d,B%d))`,
				row, row, catArrayStr, BudgetRowAusgStart, row, row)
			g.setFormula(ws, cellName(BudgetColID, row), formulaID, BudgetIDCellStyle)
			g.setValue(ws, cellName(BudgetColPos, row), "", BudgetPosCellStyle)
		}
		g.bgInput(ws, cellName(BudgetColLC, row), BudgetFmtLC)
		g.bgInput(ws, cellName(BudgetColY1, row), BudgetFmtLC)
		g.bgInput(ws, cellName(BudgetColY2, row), BudgetFmtLC)
		g.bgInput(ws, cellName(BudgetColY3, row), BudgetFmtLC)
		g.bgInput(ws, cellName(BudgetColEUR, row), BudgetFmtEUR)
	}

	g.bgTableHeader(ws, BudgetRowAusgHdr, BudgetColLabel, BudgetColEUR)

	// Summen-Zeile
	_ = f.SetRowHeight(ws, dyn.AusgTotal, 30)
	g.setValue(ws, cellName(BudgetColLabel, dyn.AusgTotal), "Geplante Gesamtausgaben", StyleOptions{})
	g.setFormula(ws, cellName(BudgetColLC, dyn.AusgTotal),
		fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (LC)])`, BudgetTableAusg), StyleOptions{NumFormat: BudgetFmtLC})
	g.setFormula(ws, cellName(BudgetColY1, dyn.AusgTotal),
		fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BudgetTableAusg, BudgetYears[0]), StyleOptions{NumFormat: BudgetFmtLC})
	g.setFormula(ws, cellName(BudgetColY2, dyn.AusgTotal),
		fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BudgetTableAusg, BudgetYears[1]), StyleOptions{NumFormat: BudgetFmtLC})
	g.setFormula(ws, cellName(BudgetColY3, dyn.AusgTotal),
		fmt.Sprintf(`=SUBTOTAL(109,%s[%s])`, BudgetTableAusg, BudgetYears[2]), StyleOptions{NumFormat: BudgetFmtLC})
	g.setFormula(ws, cellName(BudgetColEUR, dyn.AusgTotal),
		fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (EUR)])`, BudgetTableAusg), StyleOptions{NumFormat: BudgetFmtEUR})
	g.bgTotalRow(ws, dyn.AusgTotal, BudgetColLabel, BudgetColEUR)

	return nil
}

func (g *Generator) drawBudgetDrittmittelTable(ws string) error {
	cName, cLc, cEur := BudgetColStatus, BudgetColCheck, BudgetColBegr2
	titleRow := BudgetRowSection2 // = 17
	headerRow := BudgetRowAusgHdr // = 18
	geberRows := 10
	dataRows := geberRows + 1

	g.mergeCells(ws, cellName(cName, titleRow), cellName(cEur, titleRow),
		"Drittmittel – Aufstellung je Geber", StyleOptions{
			Bold: true, FontColor: BudgetClrBlack, FillColor: BudgetClrHeader,
			HAlign: "center", VAlign: "center",
		})
	g.setValue(ws, cellName(cName, headerRow), "Name des Gebers", StyleOptions{})
	g.setValue(ws, cellName(cLc, headerRow), "Betrag (LC)", StyleOptions{})
	g.setValue(ws, cellName(cEur, headerRow), "Betrag (EUR)", StyleOptions{})

	for i := 0; i < geberRows; i++ {
		row := headerRow + 1 + i
		g.setStyle(ws, cellName(cName, row), cellName(cName, row), BudgetNameCellStyle)
		g.bgInput(ws, cellName(cLc, row), BudgetFmtLC)
		g.bgInput(ws, cellName(cEur, row), BudgetFmtEUR)
	}
	sonstigesRow := headerRow + 1 + geberRows
	g.setValue(ws, cellName(cName, sonstigesRow), "Sonstige", BudgetNameCellStyle)
	g.bgInput(ws, cellName(cLc, sonstigesRow), BudgetFmtLC)
	g.bgInput(ws, cellName(cEur, sonstigesRow), BudgetFmtEUR)

	g.bgTableHeader(ws, headerRow, cName, cEur)
	_ = g.file.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(cName, headerRow), cellName(cEur, headerRow+dataRows)),
		Name:           BudgetTableDrit,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})

	firstDataRow := headerRow + 1
	lastDataRow := headerRow + dataRows
	g.upsertNamedFormula(BudgetNameGeberList, fmt.Sprintf(
		"OFFSET('%s'!%s,0,0,MAX(1,COUNTA('%s'!$%s$%d:$%s$%d)),1)",
		ws, absName(cName, firstDataRow),
		ws, colLetter(cName), firstDataRow, colLetter(cName), lastDataRow))

	g.styleOuterBorder(ws, titleRow, cName, headerRow+dataRows, cEur, 2, BudgetClrBorder)
	return nil
}

func (g *Generator) drawBudgetReserveBox(ws string, reserveEurAddr string) {
	col := BudgetColStatus

	g.setValue(ws, cellName(col, BudgetReserveRowTitle), "Reserve Freigabe", StyleOptions{
		Bold: true, Size: 9, FontColor: BudgetClrBlack, FillColor: BudgetClrHeader,
		HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})
	if reserveEurAddr != "" {
		g.setFormula(ws, cellName(col, BudgetReserveRowAmount), fmt.Sprintf("=%s", reserveEurAddr), StyleOptions{
			Bold: true, Size: 9, FontColor: BudgetClrFont, HAlign: "center", VAlign: "center",
			NumFormat:  BudgetFmtEUR,
			BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
		})
	}
	g.setValue(ws, cellName(col, BudgetReserveRowLabel), "Reserve freigeben:", StyleOptions{
		Size: 9, FontColor: BudgetClrResTxt, Italic: true, HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})
	g.setStyle(ws, cellName(col, BudgetReserveRowCheck), cellName(col, BudgetReserveRowCheck), StyleOptions{
		FillColor: BudgetClrInput, HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})

	// Reserve-Freigabe wird über ihren benannten Bereich referenziert.
	checkAddr := BudgetNameReserve
	statusFormula := fmt.Sprintf(`=IF(%s="Ja","FREIGEGEBEN","NICHT FREIGEGEBEN")`, checkAddr)
	statusStyleID, _ := g.getOrCreateStyle(StyleOptions{
		Bold: true, Size: 9, FontColor: BudgetClrResTxt, FillColor: BudgetClrResOff,
		HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})
	g.file.SetCellFormula(ws, cellName(col, BudgetReserveRowStatus), statusFormula)
	g.file.SetCellStyle(ws, cellName(col, BudgetReserveRowStatus), cellName(col, BudgetReserveRowStatus), statusStyleID)

	g.addConditionalFormat(ws, cellName(col, BudgetReserveRowStatus), fmt.Sprintf(`%s="Ja"`, checkAddr), StyleOptions{
		Bold: true, Size: 9, FontColor: BudgetClrResOnTxt, FillColor: BudgetClrResOn,
		BorderLeft: 2, BorderRight: 2, BorderTop: 1, BorderBottom: 2, BorderColor: BudgetClrBorder,
	})
	g.styleOuterBorder(ws, BudgetReserveRowTitle, col, BudgetReserveRowStatus, col, 2, BudgetClrBorder)
}

func (g *Generator) drawBudgetBegruendung(ws string, reserveCheckAddr string) {
	c1, c2 := BudgetColBegr1, BudgetColBegr2
	g.mergeCells(ws, cellName(c1, BudgetBegrHdrRow), cellName(c2, BudgetBegrHdrRow), "Begruendung", StyleOptions{
		Bold: true, Size: 9, FontColor: BudgetClrBlack, FillColor: BudgetClrHeader,
		HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrBorder,
	})
	g.mergeCells(ws, cellName(c1, BudgetBegrAreaTop), cellName(c2, BudgetBegrAreaTop+BudgetBegrAreaRows-1), "", StyleOptions{
		FillColor: "F2F2F2", HAlign: "left", VAlign: "top", WrapText: true,
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: "D3D3D3", Unlocked: true,
	})
	condFormula := fmt.Sprintf(`%s="Ja"`, reserveCheckAddr)
	g.addConditionalFormat(ws,
		fmt.Sprintf("%s:%s", cellName(c1, BudgetBegrAreaTop), cellName(c2, BudgetBegrAreaTop+BudgetBegrAreaRows-1)),
		condFormula,
		StyleOptions{FillColor: BudgetClrInput, BorderTop: 1, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BudgetClrBorder},
	)
}

func (g *Generator) drawBudgetChecks(ws string, top int, incLocAddr, incEurAddr, incYearsAddr, expLocAddr, expEurAddr, expYearsAddr, rateCellAddr string) {
	cLbl, cVal := BudgetColLabel, BudgetColLC
	g.mergeCells(ws, cellName(cLbl, top), cellName(cVal, top), "Budgetpruefung", StyleOptions{
		Bold: true, Size: 9, FontColor: BudgetClrBlack, FillColor: BudgetClrHeader,
		HAlign: "center", VAlign: "center",
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})

	checks := []struct{ lbl, fml string }{
		{"Einnahmen = Ausgaben (LC)", fmt.Sprintf(`=IF(ROUND(%s-%s,2)=0,"OK","Abweichung")`, incLocAddr, expLocAddr)},
		{"Einnahmen = Ausgaben (EUR)", fmt.Sprintf(`=IF(ROUND(%s-%s,2)=0,"OK","Abweichung")`, incEurAddr, expEurAddr)},
		{"Gleicher Budget-Kurs", fmt.Sprintf(`=IF(ROUND(%s,4)=ROUND(IFERROR(%s/%s,0),4),"OK","Abweichung")`, rateCellAddr, expLocAddr, expEurAddr)},
		{"Einnahmen: Jahre = Gesamt (LC)", fmt.Sprintf(`=IF(ROUND((%s)-%s,2)=0,"OK","Abweichung")`, incYearsAddr, incLocAddr)},
		{"Ausgaben: Jahre = Gesamt (LC)", fmt.Sprintf(`=IF(ROUND((%s)-%s,2)=0,"OK","Abweichung")`, expYearsAddr, expLocAddr)},
	}

	for i, chk := range checks {
		rr := top + 1 + i
		g.mergeCells(ws, cellName(cLbl, rr), cellName(BudgetColPos, rr), chk.lbl, StyleOptions{
			Size: 9, FontColor: BudgetClrResTxt, HAlign: "left", VAlign: "center",
			BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
		})
		valCell := cellName(cVal, rr)
		valStyleID, _ := g.getOrCreateStyle(StyleOptions{
			Bold: true, Size: 9, FontColor: BudgetClrResTxt, FillColor: BudgetClrResOff,
			HAlign: "center", VAlign: "center",
			BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
		})
		g.file.SetCellFormula(ws, valCell, chk.fml)
		g.file.SetCellStyle(ws, valCell, valCell, valStyleID)

		valAddr := absName(cVal, rr)
		bBottom := 1
		if i == len(checks)-1 {
			bBottom = 2
		}
		g.addConditionalFormat(ws, valCell, fmt.Sprintf(`%s="OK"`, valAddr), StyleOptions{
			Bold: true, Size: 9, FontColor: BudgetClrResOnTxt, FillColor: BudgetClrResOn,
			BorderLeft: 1, BorderTop: 1, BorderRight: 2, BorderBottom: bBottom, BorderColor: BudgetClrBorder,
		})
		g.addConditionalFormat(ws, valCell, fmt.Sprintf(`%s<>"OK"`, valAddr), StyleOptions{
			Bold: true, Size: 9, FontColor: BudgetClrBadTxt, FillColor: BudgetClrBad,
			BorderLeft: 1, BorderTop: 1, BorderRight: 2, BorderBottom: bBottom, BorderColor: BudgetClrBorder,
		})
	}
	g.styleOuterBorder(ws, top, cLbl, top+len(checks), cVal, 2, BudgetClrBorder)
}

// drawBudgetYearInputRow zeichnet leere Eingabe-Zellen für LC/Y1/Y2/Y3/EUR.
func (g *Generator) drawBudgetYearInputRow(ws string, r int) {
	g.bgInput(ws, cellName(BudgetColLC, r), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColY1, r), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColY2, r), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColY3, r), BudgetFmtLC)
	g.bgInput(ws, cellName(BudgetColEUR, r), BudgetFmtEUR)
}

// ─── Teil E: Bind-Funktionen (Logik & Registry) ───────────────────────────────

func (g *Generator) bindBudgetFinancing(ws string, reg *TemplateRegistry, dyn budgetDynRows) error {
	f := g.file

	// Kurs
	g.upsertNamedRange(reg.OutputBudgetWK.NamedRange, BudgetColY3, BudgetRowSection1)

	// Eigenmittel — bindInputField setzt "Eigenmittel_LW" / "Eigenmittel_EUR" direkt aus dem Registry-NamedRange
	_ = g.bindInputField(ws, BudgetRowEigen, BudgetColLC, reg.InputBudgetEigenmittelLC)
	_ = g.bindInputField(ws, BudgetRowEigen, BudgetColY1, reg.InputBudgetEigenmittelY1)
	_ = g.bindInputField(ws, BudgetRowEigen, BudgetColY2, reg.InputBudgetEigenmittelY2)
	_ = g.bindInputField(ws, BudgetRowEigen, BudgetColY3, reg.InputBudgetEigenmittelY3)
	_ = g.bindInputField(ws, BudgetRowEigen, BudgetColEUR, reg.InputBudgetEigenmittelEUR)

	// Drittmittel — LC/EUR sind Formel-Outputs, Y1–Y3 Input
	g.bgSummeCell(ws, BudgetRowDritt, BudgetColLC,
		fmt.Sprintf(`=SUM(%s[Betrag (LC)])`, BudgetTableDrit), BudgetFmtLC)
	g.bgSummeCell(ws, BudgetRowDritt, BudgetColEUR,
		fmt.Sprintf(`=SUM(%s[Betrag (EUR)])`, BudgetTableDrit), BudgetFmtEUR)
	_ = g.bindInputField(ws, BudgetRowDritt, BudgetColY1, reg.InputBudgetDrittmittelY1)
	_ = g.bindInputField(ws, BudgetRowDritt, BudgetColY2, reg.InputBudgetDrittmittelY2)
	_ = g.bindInputField(ws, BudgetRowDritt, BudgetColY3, reg.InputBudgetDrittmittelY3)
	g.upsertNamedRange(reg.OutputBudgetDrittmittelLC.NamedRange, BudgetColLC, BudgetRowDritt)
	g.upsertNamedRange(reg.OutputBudgetDrittmittelEUR.NamedRange, BudgetColEUR, BudgetRowDritt)

	// KMW-Mittel — bindInputField setzt "KMW_Mittel_LW" / "KMW_Mittel_EUR"
	_ = g.bindInputField(ws, BudgetRowKMW, BudgetColLC, reg.InputBudgetKMWLC)
	_ = g.bindInputField(ws, BudgetRowKMW, BudgetColY1, reg.InputBudgetKMWY1)
	_ = g.bindInputField(ws, BudgetRowKMW, BudgetColY2, reg.InputBudgetKMWY2)
	_ = g.bindInputField(ws, BudgetRowKMW, BudgetColY3, reg.InputBudgetKMWY3)
	_ = g.bindInputField(ws, BudgetRowKMW, BudgetColEUR, reg.InputBudgetKMWEUR)

	// GESAMTPROJEKTMITTEL – Summen aus Eigenmittel/Drittmittel/KMW-Mittel,
	// ausschließlich über die benannten Bereiche der jeweiligen Zeile.
	gesamt := []struct {
		col               int
		eigen, dritt, kmw string
		out               OutputField
		fmtStr            string
	}{
		{BudgetColLC, reg.InputBudgetEigenmittelLC.NamedRange, reg.OutputBudgetDrittmittelLC.NamedRange, reg.InputBudgetKMWLC.NamedRange, reg.OutputBudgetGesamtLC, BudgetFmtLC},
		{BudgetColY1, reg.InputBudgetEigenmittelY1.NamedRange, reg.InputBudgetDrittmittelY1.NamedRange, reg.InputBudgetKMWY1.NamedRange, reg.OutputBudgetGesamtY1, BudgetFmtLC},
		{BudgetColY2, reg.InputBudgetEigenmittelY2.NamedRange, reg.InputBudgetDrittmittelY2.NamedRange, reg.InputBudgetKMWY2.NamedRange, reg.OutputBudgetGesamtY2, BudgetFmtLC},
		{BudgetColY3, reg.InputBudgetEigenmittelY3.NamedRange, reg.InputBudgetDrittmittelY3.NamedRange, reg.InputBudgetKMWY3.NamedRange, reg.OutputBudgetGesamtY3, BudgetFmtLC},
		{BudgetColEUR, reg.InputBudgetEigenmittelEUR.NamedRange, reg.OutputBudgetDrittmittelEUR.NamedRange, reg.InputBudgetKMWEUR.NamedRange, reg.OutputBudgetGesamtEUR, BudgetFmtEUR},
	}
	for _, gs := range gesamt {
		g.setFormula(ws, cellName(gs.col, BudgetRowGesamt),
			fmt.Sprintf("=%s+%s+%s", gs.eigen, gs.dritt, gs.kmw),
			StyleOptions{NumFormat: gs.fmtStr})
		g.upsertNamedRange(gs.out.NamedRange, gs.col, BudgetRowGesamt)
	}

	// Budget-Kurs-Formel (Gesamt LC / Gesamt EUR über benannte Bereiche)
	rateCellOpts := StyleOptions{
		NumFormat: BudgetFmtRate, Italic: true,
		BorderTop: 2, BorderBottom: 1, BorderLeft: 1, BorderRight: 1, BorderColor: BudgetClrBorder,
	}
	g.setFormula(ws, cellName(BudgetColY3, BudgetRowSection1),
		fmt.Sprintf(`=IFERROR(%s/%s,0)`, reg.OutputBudgetGesamtLC.NamedRange, reg.OutputBudgetGesamtEUR.NamedRange), rateCellOpts)

	// Ausgaben-Hilfszeilen (Kategorien-Summen, per-category = dynamisch)
	for i, cat := range ListKostenkategorien {
		hr := 4 + i
		g.setFormula(ws, cellName(BudgetColHelpLC, hr),
			fmt.Sprintf(`=SUMIFS(%s[Betrag (LC)],%s[Kostenkategorie],"%s")`, BudgetTableAusg, BudgetTableAusg, cat),
			StyleOptions{})
		g.setFormula(ws, cellName(BudgetColHelpEUR, hr),
			fmt.Sprintf(`=SUMIFS(%s[Betrag (EUR)],%s[Kostenkategorie],"%s")`, BudgetTableAusg, BudgetTableAusg, cat),
			StyleOptions{})
		lcName, eurName := bgKostenName(cat, "LW"), bgKostenName(cat, "EUR")
		g.upsertNamedRange(lcName, BudgetColHelpLC, hr)
		g.upsertNamedRange(eurName, BudgetColHelpEUR, hr)
	}

	// Gesamtausgaben-Hilfszeile (für Prüfungs-Formeln — kein Named Range nötig)
	gesHr := 4 + len(ListKostenkategorien)
	g.setFormula(ws, cellName(BudgetColHelpLC, gesHr),
		fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (LC)])`, BudgetTableAusg), StyleOptions{})
	g.setFormula(ws, cellName(BudgetColHelpEUR, gesHr),
		fmt.Sprintf(`=SUBTOTAL(109,%s[Betrag (EUR)])`, BudgetTableAusg), StyleOptions{})

	// Begründungsfeld
	_ = g.bindInputField(ws, BudgetBegrAreaTop, BudgetColBegr1, reg.InputBudgetBegruendung)

	_ = f.SetColVisible(ws, colLetter(BudgetColHelpLC), false)
	_ = f.SetColVisible(ws, colLetter(BudgetColHelpEUR), false)
	_ = f.SetColVisible(ws, colLetter(BudgetColListGeber), false)
	_ = f.SetColVisible(ws, colLetter(BudgetColListID), false)

	return nil
}

func (g *Generator) bindBudgetAusgaben(ws string, reg *TemplateRegistry, dyn budgetDynRows) error {
	f := g.file

	// Excel-Tabelle anlegen. Die Summenzeile (dyn.AusgTotal, "Geplante
	// Gesamtausgaben") liegt AUSSERHALB des Table-Range – excelize kann keine echte
	// Totals-Row erzeugen. Als Zeile direkt unter der Tabelle summiert
	// SUBTOTAL(109, BudgetTableAusg[…]) nur die Datenzeilen (kein Zirkelbezug).
	_ = f.AddTable(ws, &excelize.Table{
		Range:          fmt.Sprintf("%s:%s", cellName(BudgetColLabel, BudgetRowAusgHdr), cellName(BudgetColEUR, dyn.AusgTotal-1)),
		Name:           BudgetTableAusg,
		StyleName:      "",
		ShowRowStripes: falsePtr(),
	})

	// Kategorie-Dropdown über alle Ausgaben-Zeilen – statische Liste aus der Registry-Spalte
	catSqref := fmt.Sprintf("%s:%s",
		cellName(BudgetColLabel, BudgetRowAusgStart),
		cellName(BudgetColLabel, BudgetRowAusgStart+dyn.AusgDataRows-1))
	_ = g.applyColumnValidation(ws, catSqref, reg.TableBudgetAusgaben.Columns[0])

	// ID-Liste für Lookup-Formeln
	g.setDynArrayFormula(ws, cellName(BudgetColListID, 1),
		fmt.Sprintf(`IFERROR(_xlfn._xlws.FILTER(%s[ID],%s[ID]<>""),"")`, BudgetTableAusg, BudgetTableAusg),
		StyleOptions{})
	g.upsertNamedFormula(BudgetNameIDList,
		fmt.Sprintf("OFFSET('%s'!%s, 0, 0, COUNTA('%s'!%s:%s), 1)",
			ws, absName(BudgetColListID, 1),
			ws, colLetter(BudgetColListID), colLetter(BudgetColListID)))

	// Ergebniszeile "Geplante Gesamtausgaben" als Named Ranges exponieren.
	g.upsertNamedRange(reg.OutputBudgetAusgabenGesamtLC.NamedRange, BudgetColLC, dyn.AusgTotal)
	g.upsertNamedRange(reg.OutputBudgetAusgabenGesamtY1.NamedRange, BudgetColY1, dyn.AusgTotal)
	g.upsertNamedRange(reg.OutputBudgetAusgabenGesamtY2.NamedRange, BudgetColY2, dyn.AusgTotal)
	g.upsertNamedRange(reg.OutputBudgetAusgabenGesamtY3.NamedRange, BudgetColY3, dyn.AusgTotal)
	g.upsertNamedRange(reg.OutputBudgetAusgabenGesamtEUR.NamedRange, BudgetColEUR, dyn.AusgTotal)

	return nil
}

// ─── Bestehende Hilfsfunktionen (intern, unverändert) ────────────────────────

func bgKostenName(cat string, cur string) string {
	return fmt.Sprintf("Kosten_%s_%s", cat, cur)
}

func falsePtr() *bool {
	b := false
	return &b
}

func (g *Generator) bgSectionHeader(ws string, r int, title string) {
	for c := BudgetColLabel; c <= BudgetColEUR; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), BudgetSectionHdrStyle)
	}
	g.setValue(ws, cellName(BudgetColLabel, r), title, BudgetSectionHdrStyle)
	_ = g.file.SetRowHeight(ws, r, 24)
}

func (g *Generator) bgSubHeader(ws string, r int, title string) {
	for c := BudgetColLabel; c <= BudgetColEUR; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), BudgetSubHdrStyle)
	}
	g.setValue(ws, cellName(BudgetColLabel, r), title, BudgetSubHdrStyle)
	_ = g.file.SetRowHeight(ws, r, 20)
}

func (g *Generator) budgetValueHeaderCells(ws string, r int) {
	g.setValue(ws, cellName(BudgetColLC, r), "Betrag (LC)", StyleOptions{})
	g.setValue(ws, cellName(BudgetColY1, r), BudgetYears[0], StyleOptions{})
	g.setValue(ws, cellName(BudgetColY2, r), BudgetYears[1], StyleOptions{})
	g.setValue(ws, cellName(BudgetColY3, r), BudgetYears[2], StyleOptions{})
	g.setValue(ws, cellName(BudgetColEUR, r), "Betrag (EUR)", StyleOptions{})
}

func (g *Generator) bgTableHeader(ws string, r int, c1 int, c2 int) {
	for c := c1; c <= c2; c++ {
		g.setStyle(ws, cellName(c, r), cellName(c, r), BudgetTableHdrStyle)
	}
}

func (g *Generator) bgYearRow(ws string, r int, fields []InputField) {
	for i, c := range []int{BudgetColLC, BudgetColY1, BudgetColY2, BudgetColY3} {
		g.bgInput(ws, cellName(c, r), BudgetFmtLC)
		if i < len(fields) {
			_ = g.bindInputField(ws, r, c, fields[i])
		}
	}
	g.bgInput(ws, cellName(BudgetColEUR, r), BudgetFmtEUR)
	if len(fields) > 4 {
		_ = g.bindInputField(ws, r, BudgetColEUR, fields[4])
	}
}

func (g *Generator) bgFillIncomeRow(ws string, r int, inc IncomeRow) {
	g.bgFillInput(ws, cellName(BudgetColLC, r), inc.LC)
	g.bgFillInput(ws, cellName(BudgetColY1, r), inc.Y1)
	g.bgFillInput(ws, cellName(BudgetColY2, r), inc.Y2)
	g.bgFillInput(ws, cellName(BudgetColY3, r), inc.Y3)
	g.bgFillInput(ws, cellName(BudgetColEUR, r), inc.EUR)
}

func (g *Generator) bgSummeCell(ws string, r int, c int, formula string, fmtStr string) {
	g.setFormula(ws, cellName(c, r), formula, StyleOptions{
		Bold: true, HAlign: "right", VAlign: "center", NumFormat: fmtStr,
	})
}

func (g *Generator) bgInput(ws string, cell string, numFmt string) {
	g.setStyle(ws, cell, cell, StyleOptions{
		FillColor: BudgetClrInput, HAlign: "right", VAlign: "center", NumFormat: numFmt,
		BorderLeft: 1, BorderRight: 1, BorderTop: 1, BorderBottom: 1, BorderColor: BudgetClrGrid,
	})
}

func (g *Generator) bgFillInput(ws, cell string, v *float64) {
	if v != nil {
		_ = g.file.SetCellValue(ws, cell, *v)
	}
}

func (g *Generator) bgTotalRow(ws string, r int, c1 int, c2 int) {
	for c := c1; c <= c2; c++ {
		var opts StyleOptions
		switch {
		case c >= BudgetColLC && c <= BudgetColY3:
			opts = BudgetTotalRowLCStyle
		case c == BudgetColEUR:
			opts = BudgetTotalRowEURStyle
		default:
			opts = BudgetTotalRowStyle
		}
		g.setStyle(ws, cellName(c, r), cellName(c, r), opts)
	}
	_ = g.file.SetRowHeight(ws, r, 20)
}
