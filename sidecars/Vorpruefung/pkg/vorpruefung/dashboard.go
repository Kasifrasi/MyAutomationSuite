package vorpruefung

import (
	"fmt"
	"shared/constants"

	"github.com/xuri/excelize/v2"
)

// ─── Teil A: Grid-Konstanten ──────────────────────────────────────────────────

const (
	// Spalten
	DashColLabelLeft  = 2 // B
	DashColInputLeft  = 3 // C
	DashColLabelRight = 4 // D
	DashColInputRight = 5 // E

	// Zeilen
	DashRowHeader          = 2
	DashRowTitle           = 4
	DashRowProjektNummer   = 5
	DashRowProjekttitel    = 6
	DashRowProjekttraeger  = 7
	DashRowProjektstart    = 8
	DashRowProjektlaufzeit = 9
	DashRowVPNummer        = 10 // VP-Block Beginn (Doppellinie)
	DashRowVPEnde          = 11
	DashRowVPSaldo         = 12
	DashRowVPFolgestart    = 13
	DashRowVPSaldovortrag  = 14 // VP-Block Ende; Inp_Dash_VPFolgeSaldoLC / Out_Dash_SaldovortragEUR
	DashRowChecklistStart  = 16

	DashCCYCol = 27 // Hilfsspalte für Währungsliste (ausgeblendet)

	// Sheet
	DashSheetName = constants.VPSheetDASHBOARD
	DashTabColor  = "D3D3D3"
)

// ─── Teil B: Layout-Dokumentation ────────────────────────────────────────────
/*
  LAYOUT DASHBOARD:
  | Zeile | Spalte B (Label)              | Spalte C (Input 1)             | Spalte D (Label)            | Spalte E (Input 2)              |
  |-------|-------------------------------|--------------------------------|-----------------------------|----------------------------------|
  |   2   | DASHBOARD-Header (merged B:E, grün)                                                                                            |
  |   4   | "Statische Projektinformationen" (Titel, merged B:E)                                                                           |
  |   5   | Projektnummer                 | [Inp_Dash_Projektnummer]       | Vorprojekt vorhanden        | [Inp_Dash_Vorprojekt]           |
  |   6   | Projekttitel                  | [Inp_Dash_Projekttitel (merged C:E)]                                                          |
  |   7   | Projekttraeger                | [Inp_Dash_Projekttraeger]      | Berichtswaehrung            | [Inp_Dash_Berichtswaehrung]     |
  |   8   | Projektstart                  | [Inp_Dash_Projektstart]        | Projektende                 | [Inp_Dash_Projektende]          |
  |   9   | Projektlaufzeit (geplant)     | [Out_Dash_Projektlaufzeit]     | In Monate                   | [Out_Dash_Monate]               |
  |-------|-------------------------------|--------------------------------|-----------------------------|----------------------------------|
  |  10   | Vorprojektnummer (=VPStart ══)| [Inp_Dash_VPNummer]            | VP-Berichtswaehrung         | [Inp_Dash_VPBerichtswaehrung]   |
  |  11   | Vorprojektende                | [Inp_Dash_VPEnde]              | Wechselkurs                 | [Inp_Dash_VPWechselkurs]        |
  |  12   | Saldo (LW)                    | [Inp_Dash_VPSaldoLC]           | Saldo (EUR)                 | [Out_Dash_SaldoEUR]             |
  |  13   | Folgeprojektstart             | [Inp_Dash_VPFolgeprojektstart] | Wechselkurs                 | [Inp_Dash_VPFolgeWechselkurs]   |
  |  14   | Saldovortrag (LW)             | [Inp_Dash_VPFolgeSaldoLC]      | Saldovortrag (EUR)          | [Out_Dash_SaldovortragEUR]      |
  |-------|-------------------------------|--------------------------------|-----------------------------|----------------------------------|
  |  16   | "Folgende Dokumente liegen vor:" (merged B:C, Zeilen 16:21)        | [D16] Ja/Nein               | Vorprojektsaldo (Nachweis)  |
  |  17   |                               |                                | [D17] Ja/Nein               | Vertrag                         |
  |  18   |                               |                                | [D18] Ja/Nein               | Budget                          |
  |  19   |                               |                                | [D19] Ja/Nein               | Bankbelege                      |
  |  20   |                               |                                | [D20] Ja/Nein               | Finanzbericht(e)                |
  |  21   |                               |                                | [D21] Ja/Nein               | Mittelanforderung(en)           |
*/

// ─── Checklist-Einträge ───────────────────────────────────────────────────────

var AppVersion = "v1.0.1"

var DashDocs = []string{
	"Vorprojektsaldo (Nachweis)",
	"Vertrag",
	"Budget",
	"Bankbelege",
	"Finanzbericht(e)",
	"Mittelanforderung(en)",
}

// ─── Teil C: Orchestrator ─────────────────────────────────────────────────────

func (g *Generator) CreateDashboardSheet(reg *TemplateRegistry) error {
	ws := DashSheetName

	_, err := g.file.NewSheet(ws)
	if err != nil {
		return err
	}
	tabColor := DashTabColor
	_ = g.file.SetSheetProps(ws, &excelize.SheetPropsOptions{TabColorRGB: &tabColor})
	_ = g.file.SetSheetView(ws, 0, &excelize.ViewOptions{ShowGridLines: falsePtr()})

	g.dashSetupColumns(ws)

	if err = g.drawDashHeader(ws); err != nil {
		return err
	}
	if err = g.drawDashStaticInfo(ws); err != nil {
		return err
	}
	if err = g.drawDashChecklist(ws); err != nil {
		return err
	}
	if err = g.bindDashFields(ws, reg); err != nil {
		return err
	}
	if err = g.bindDashChecklist(ws, reg); err != nil {
		return err
	}
	if err = g.applyDashConditionalFormatting(ws, reg); err != nil {
		return err
	}

	return nil
}

// ─── Teil D: Draw-Funktionen (nur visuell) ───────────────────────────────────

func (g *Generator) drawDashHeader(ws string) error {
	err := g.mergeCells(ws,
		cellName(DashColLabelLeft, DashRowHeader),
		cellName(DashColInputRight, DashRowHeader),
		"  DASHBOARD ("+AppVersion+")",
		DashHeaderStyle,
	)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowHeader, 40.0)
	return nil
}

func (g *Generator) drawDashStaticInfo(ws string) error {
	// Titel
	err := g.mergeCells(ws,
		cellName(DashColLabelLeft, DashRowTitle),
		cellName(DashColInputRight, DashRowTitle),
		"Statische Projektinformationen",
		DashTitleStyle,
	)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowTitle, 24.0)

	// Zeile 5: Projektnummer | Vorprojekt vorhanden
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowProjektNummer), "Projektnummer", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowProjektNummer), "", DashInputStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowProjektNummer), "Vorprojekt vorhanden", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setStyle(ws, cellName(DashColInputRight, DashRowProjektNummer), cellName(DashColInputRight, DashRowProjektNummer), DashDropdownStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowProjektNummer, 22.0)

	// Zeile 6: Projekttitel (C:E zusammengefasst)
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowProjekttitel), "Projekttitel", DashLabelStyle); err != nil {
		return err
	}
	if err = g.mergeCells(ws,
		cellName(DashColInputLeft, DashRowProjekttitel),
		cellName(DashColInputRight, DashRowProjekttitel),
		"", DashInputStyle,
	); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowProjekttitel, 22.0)

	// Zeile 7: Projekttraeger | Berichtswaehrung
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowProjekttraeger), "Projekttraeger", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowProjekttraeger), "", DashInputStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowProjekttraeger), "Berichtswaehrung", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowProjekttraeger), "", DashInputStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowProjekttraeger, 22.0)

	// Zeile 8: Projektstart | Projektende
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowProjektstart), "Projektstart", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowProjektstart), "", DashInputDateStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowProjektstart), "Projektende", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowProjektstart), "", DashInputDateStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowProjektstart, 22.0)

	// Zeile 9: Projektlaufzeit (Formel) | In Monate (Formel) — wird in bindDashFields befüllt
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowProjektlaufzeit), "Projektlaufzeit (geplant)", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowProjektlaufzeit), "", DashOutputStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowProjektlaufzeit), "In Monate", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowProjektlaufzeit), "", DashOutputStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowProjektlaufzeit, 22.0)

	// ── VP-Block (Doppellinie oben ab Zeile 10) ──────────────────────────────

	// Zeile 10: Vorprojektnummer | VP-Berichtswaehrung
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowVPNummer), "Vorprojektnummer", DashVPLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowVPNummer), "", DashVPInputStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowVPNummer), "VP-Berichtswaehrung", DashVPLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowVPNummer), "", DashVPInputStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowVPNummer, 22.0)

	// Zeile 11: Vorprojektende | Wechselkurs
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowVPEnde), "Vorprojektende", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowVPEnde), "", DashInputDateStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowVPEnde), "Wechselkurs", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowVPEnde), "", DashInputRateStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowVPEnde, 22.0)

	// Zeile 12: Saldo (LW) | Saldo (EUR) — EUR ist Output/Formel
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowVPSaldo), "Saldo (LW)", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowVPSaldo), "", DashInputLCStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowVPSaldo), "Saldo (EUR)", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowVPSaldo), "", DashOutputEURStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowVPSaldo, 22.0)

	// Zeile 13: Folgeprojektstart | Wechselkurs
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowVPFolgestart), "Folgeprojektstart", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowVPFolgestart), "", DashInputDateStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowVPFolgestart), "Wechselkurs", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowVPFolgestart), "", DashInputRateStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowVPFolgestart, 22.0)

	// Zeile 14: Saldovortrag (LW) | Saldovortrag (EUR) — EUR ist Output/Formel
	if err = g.setValue(ws, cellName(DashColLabelLeft, DashRowVPSaldovortrag), "Saldovortrag (LW)", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputLeft, DashRowVPSaldovortrag), "", DashInputLCStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColLabelRight, DashRowVPSaldovortrag), "Saldovortrag (EUR)", DashLabelStyle); err != nil {
		return err
	}
	if err = g.setValue(ws, cellName(DashColInputRight, DashRowVPSaldovortrag), "", DashOutputEURStyle); err != nil {
		return err
	}
	_ = g.file.SetRowHeight(ws, DashRowVPSaldovortrag, 22.0)

	return nil
}

func (g *Generator) drawDashChecklist(ws string) error {
	docEnd := DashRowChecklistStart + len(DashDocs) - 1

	// B16:C21 merged — Label
	if err := g.mergeCells(ws,
		cellName(DashColLabelLeft, DashRowChecklistStart),
		cellName(DashColInputLeft, docEnd),
		"Folgende Dokumente liegen vor:",
		DashChecklistLabelStyle,
	); err != nil {
		return err
	}

	// D16:D21 Dropdowns + E16:E21 Texte
	for i, docName := range DashDocs {
		row := DashRowChecklistStart + i
		_ = g.file.SetRowHeight(ws, row, 22.0)

		if err := g.setStyle(ws, cellName(DashColLabelRight, row), cellName(DashColLabelRight, row), DashDropdownStyle); err != nil {
			return err
		}
		if err := g.setValue(ws, cellName(DashColInputRight, row), docName, DashChecklistTextStyle); err != nil {
			return err
		}
	}
	return nil
}

// ─── Teil E: Bind-Funktionen (Logik & Registry) ───────────────────────────────

func (g *Generator) bindDashFields(ws string, reg *TemplateRegistry) error {
	// Projektnummer
	_ = g.bindInputField(ws, DashRowProjektNummer, DashColInputLeft, reg.InputDashProjektnummer)

	// Vorprojekt vorhanden (Dropdown)
	_ = g.bindInputField(ws, DashRowProjektNummer, DashColInputRight, reg.InputDashVorprojekt)

	// Projekttitel
	_ = g.bindInputField(ws, DashRowProjekttitel, DashColInputLeft, reg.InputDashProjekttitel)

	// Projekttraeger
	_ = g.bindInputField(ws, DashRowProjekttraeger, DashColInputLeft, reg.InputDashProjekttraeger)

	// Berichtswaehrung (Dropdown)
	_ = g.bindInputField(ws, DashRowProjekttraeger, DashColInputRight, reg.InputDashBerichtswaehrung)
	if err := g.dashCurrencyValidation(ws, DashRowProjekttraeger, DashColInputRight); err != nil {
		return err
	}

	// Projektstart / Projektende
	_ = g.bindInputField(ws, DashRowProjektstart, DashColInputLeft, reg.InputDashProjektstart)
	_ = g.bindInputField(ws, DashRowProjektstart, DashColInputRight, reg.InputDashProjektende)

	// Projektlaufzeit & Monate (Formeln)
	fmtDate := func(addr string) string {
		return fmt.Sprintf(`TEXT(DAY(%s),"00")&"."&TEXT(MONTH(%s),"00")&"."&TEXT(YEAR(%s),"0000")`, addr, addr, addr)
	}
	startAddr := reg.InputDashProjektstart.NamedRange
	endeAddr := reg.InputDashProjektende.NamedRange
	laufzeitFormula := fmt.Sprintf(
		`=IF(AND(ISNUMBER(%s),ISNUMBER(%s)),%s&" - "&%s,"")`,
		startAddr, endeAddr, fmtDate(startAddr), fmtDate(endeAddr),
	)
	if err := g.setFormula(ws, cellName(DashColInputLeft, DashRowProjektlaufzeit), laufzeitFormula, DashOutputStyle); err != nil {
		return err
	}
	g.dbUpsertNamedRange(ws, reg.OutputDashProjektlaufzeit.NamedRange, DashColInputLeft, DashRowProjektlaufzeit)
	monateFormula := fmt.Sprintf(
		`=IF(AND(ISNUMBER(%s),ISNUMBER(%s)),DATEDIF(%s,%s+1,"M"),"")`,
		startAddr, endeAddr, startAddr, endeAddr,
	)
	if err := g.setFormula(ws, cellName(DashColInputRight, DashRowProjektlaufzeit), monateFormula, DashOutputStyle); err != nil {
		return err
	}
	g.dbUpsertNamedRange(ws, reg.OutputDashMonate.NamedRange, DashColInputRight, DashRowProjektlaufzeit)

	// VP-Block
	_ = g.bindInputField(ws, DashRowVPNummer, DashColInputLeft, reg.InputDashVPNummer)
	_ = g.bindInputField(ws, DashRowVPNummer, DashColInputRight, reg.InputDashVPBerichtswaehrung)
	if err := g.dashCurrencyValidation(ws, DashRowVPNummer, DashColInputRight); err != nil {
		return err
	}

	_ = g.bindInputField(ws, DashRowVPEnde, DashColInputLeft, reg.InputDashVPEnde)
	_ = g.bindInputField(ws, DashRowVPEnde, DashColInputRight, reg.InputDashVPWechselkurs)

	// Saldo (LW) + Saldo (EUR) als Formel
	_ = g.bindInputField(ws, DashRowVPSaldo, DashColInputLeft, reg.InputDashVPSaldoLC)
	saldoEURFormula := fmt.Sprintf(
		"=IFERROR(ROUND(%s/%s,2),0)",
		reg.InputDashVPSaldoLC.NamedRange,
		reg.InputDashVPWechselkurs.NamedRange,
	)
	if err := g.setFormula(ws, cellName(DashColInputRight, DashRowVPSaldo), saldoEURFormula, DashOutputEURStyle); err != nil {
		return err
	}
	_ = g.bindInputField(ws, DashRowVPSaldo, DashColInputRight, reg.InputDashVPSaldoEUR)
	g.dbUpsertNamedRange(ws, reg.OutputDashSaldoEUR.NamedRange, DashColInputRight, DashRowVPSaldo)

	// Folgeprojektstart + FolgeWechselkurs
	_ = g.bindInputField(ws, DashRowVPFolgestart, DashColInputLeft, reg.InputDashVPFolgeprojektstart)
	_ = g.bindInputField(ws, DashRowVPFolgestart, DashColInputRight, reg.InputDashVPFolgeWechselkurs)

	// Saldovortrag (LW) + Saldovortrag (EUR) als Formel
	_ = g.bindInputField(ws, DashRowVPSaldovortrag, DashColInputLeft, reg.InputDashVPFolgeSaldoLC)
	saldovortragEURFormula := fmt.Sprintf(
		"=IFERROR(ROUND(%s/%s,2),0)",
		reg.InputDashVPFolgeSaldoLC.NamedRange,
		reg.InputDashVPFolgeWechselkurs.NamedRange,
	)
	if err := g.setFormula(ws, cellName(DashColInputRight, DashRowVPSaldovortrag), saldovortragEURFormula, DashOutputEURStyle); err != nil {
		return err
	}
	_ = g.bindInputField(ws, DashRowVPSaldovortrag, DashColInputRight, reg.InputDashVPFolgeSaldoEUR)
	g.dbUpsertNamedRange(ws, reg.OutputDashSaldovortragEUR.NamedRange, DashColInputRight, DashRowVPSaldovortrag)

	return nil
}

func (g *Generator) bindDashChecklist(ws string, reg *TemplateRegistry) error {
	checkFields := []InputField{
		reg.InputDashVPSaldoCheck,
		reg.InputDashVertragCheck,
		reg.InputDashBudgetCheck,
		reg.InputDashBankBelegeCheck,
		reg.InputDashFBCheck,
		reg.InputDashMACheck,
	}
	for i, field := range checkFields {
		row := DashRowChecklistStart + i
		_ = g.bindInputField(ws, row, DashColLabelRight, field)
	}
	return nil
}

func (g *Generator) applyDashConditionalFormatting(ws string, reg *TemplateRegistry) error {
	// "Vorprojekt vorhanden?"-Auswahl über ihren benannten Bereich.
	vpAddr := reg.InputDashVorprojekt.NamedRange
	docEnd := DashRowChecklistStart + len(DashDocs) - 1

	vpCfOpts := StyleOptions{FillColor: DashClrDisabled, FontColor: DashClrFontGray}
	if err := g.addConditionalFormat(ws,
		fmt.Sprintf("%s:%s", cellName(DashColLabelLeft, DashRowVPNummer), cellName(DashColInputRight, DashRowVPSaldovortrag)),
		fmt.Sprintf("=%s=\"Nein\"", vpAddr),
		vpCfOpts,
	); err != nil {
		return err
	}

	docCfOpts := StyleOptions{FontColor: DashClrFontGray, Strike: true}
	if err := g.addConditionalFormat(ws,
		fmt.Sprintf("%s:%s", cellName(DashColInputRight, DashRowChecklistStart), cellName(DashColInputRight, docEnd)),
		fmt.Sprintf("=$%s%d=\"Nein\"", colLetter(DashColLabelRight), DashRowChecklistStart),
		docCfOpts,
	); err != nil {
		return err
	}

	nachweisCfOpts := StyleOptions{FillColor: DashClrDisabled, FontColor: DashClrFontGray}
	if err := g.addConditionalFormat(ws,
		fmt.Sprintf("%s:%s", cellName(DashColLabelRight, DashRowChecklistStart), cellName(DashColInputRight, DashRowChecklistStart)),
		fmt.Sprintf("=%s=\"Nein\"", vpAddr),
		nachweisCfOpts,
	); err != nil {
		return err
	}

	return nil
}

// ─── Hilfsfunktionen ──────────────────────────────────────────────────────────

func (g *Generator) dashSetupColumns(ws string) {
	g.setColWidth(ws, 1, 3.0)
	g.setColWidth(ws, DashColLabelLeft, 32.0)
	g.setColWidth(ws, DashColInputLeft, 43.0)
	g.setColWidth(ws, DashColLabelRight, 24.0)
	g.setColWidth(ws, DashColInputRight, 35.0)
}

func (g *Generator) dashCurrencyValidation(sheet string, row, col int) error {
	dv := excelize.NewDataValidation(true)
	dv.Sqref = cellName(col, row)
	dv.SetSqrefDropList("Waehrungen_Liste")
	return g.file.AddDataValidation(sheet, dv)
}

func (g *Generator) dbUpsertNamedRange(sheet string, name string, col, row int) {
	_ = g.file.DeleteDefinedName(&excelize.DefinedName{Name: name})
	_ = g.file.SetDefinedName(&excelize.DefinedName{
		Name:     name,
		RefersTo: fmt.Sprintf("'%s'!%s", sheet, absName(col, row)),
	})
}
