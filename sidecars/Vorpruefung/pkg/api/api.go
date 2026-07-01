package api

import (
	"fmt"
	"strings"
	"time"

	"shared/constants"
	"vorpruefung/pkg/vorpruefung"

	"github.com/xuri/excelize/v2"
)

type FBPruefungData struct {
	Auswahl            *string
	AbzugSaldovortrag  *string
	AbzugMehreinnahmen *string
	AbzugPrognoseMehr  *string // Nur in MA-Prüfung relevant, aber im Layout vorgesehen
}

type MAPruefungData struct {
	Auswahl  *string
	MonateY1 *int
	MonateY2 *int
	MonateY3 *int
}

type FillData struct {
	Dashboard DashboardData
	KMW       []KMWTranche
	MA        []MAPeriod
	FB        []FBPeriod
	Budget    *BudgetData

	// Optionale Werte für die Prüfungs-Seiten
	FBPruefung *FBPruefungData
	MAPruefung *MAPruefungData
}

type DashboardData struct {
	Projektnummer       string
	Vorprojekt          *bool
	Projekttitel        string
	Projekttraeger      string
	Berichtswaehrung    string
	Projektstart        time.Time
	Projektende         time.Time
	Vorprojektnummer    string
	VPBerichtswaehrung  string
	Vorprojektende      time.Time
	VPWechselkurs       float64
	VPSaldoLC           float64
	VPSaldoEUR          float64
	VPFolgeprojektstart time.Time
	DocChecklist        []string
}

type KMWTranche struct {
	Periode  string
	Waehrung string
	Betrag   float64
	Datum    time.Time
}

type MAPeriod struct {
	Von          time.Time
	Bis          time.Time
	OandaKurs    float64
	KategorienLC map[string]float64
	EigenLC      float64
	DrittLC      float64
}

type FBEinnahme struct {
	Typ   string
	Geber string
	LC    *float64
	EUR   *float64
}

type FBPeriod struct {
	Von          time.Time
	Bis          time.Time
	Einnahmen1   []FBEinnahme
	EinnahmenWK  []FBEinnahme
	AusgabenByID map[string]float64
	BankLC       *float64
	KasseLC      *float64
	SonstigesLC  *float64
}

type BudgetData struct {
	AusgabenIDs     []string
	Eigenmittel     *IncomeRow
	KMWMittel       *IncomeRow
	DrittmittelY1   *float64
	DrittmittelY2   *float64
	DrittmittelY3   *float64
	DrittGeber      []GeberRow
	DrittSonstiges  *IncomeRow
	Ausgaben        []AusgabenRow
	ReserveFreigabe *bool
}

type IncomeRow struct {
	LC  *float64
	Y1  *float64
	Y2  *float64
	Y3  *float64
	EUR *float64
}

type GeberRow struct {
	Geber string
	LC    *float64
	EUR   *float64
}

type AusgabenRow struct {
	Kategorie string
	ID        string
	Position  string
	LC        *float64
	Y1        *float64
	Y2        *float64
	Y3        *float64
	EUR       *float64
}

func FillTemplate(filePath string, data FillData) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("konnte Datei nicht öffnen: %w", err)
	}
	defer f.Close()

	fillDashboard(f, data.Dashboard)
	fillKMW(f, data.KMW)
	fillMA(f, data.MA, data.Budget)
	fillFB(f, data.FB, data.Budget)
	fillBudget(f, data.Budget)

	// Optionale Werte für die Prüfungs-Auswahl
	// 6. FB Prüfung defaults/data
	if data.FBPruefung != nil {
		if data.FBPruefung.Auswahl != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAuswahl, data.FBPruefung.Auswahl)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAuswahl, "Neuester FB")
		}
		if data.FBPruefung.AbzugSaldovortrag != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugSaldo, data.FBPruefung.AbzugSaldovortrag)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
		}
		if data.FBPruefung.AbzugMehreinnahmen != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugMehr, data.FBPruefung.AbzugMehreinnahmen)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugMehr, vorpruefung.ListAbzug[0])
		}
	} else {
		_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAuswahl, "Neuester FB")
		_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
		_ = setValByNamedRange(f, vorpruefung.FieldFBPruefungAbzugMehr, vorpruefung.ListAbzug[0])
	}

	// 7. MA Prüfung defaults/data
	if data.MAPruefung != nil {
		if data.MAPruefung.Auswahl != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAuswahl, data.MAPruefung.Auswahl)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAuswahl, "Neueste MA")
		}
		if data.MAPruefung.MonateY1 != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY1, data.MAPruefung.MonateY1)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY1, 8)
		}
		if data.MAPruefung.MonateY2 != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY2, data.MAPruefung.MonateY2)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY2, 0)
		}
		if data.MAPruefung.MonateY3 != nil {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY3, data.MAPruefung.MonateY3)
		} else {
			_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY3, 0)
		}
		// NOTE: Toggles are not yet exposed in MAPruefungData in fill.go but let's default them
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugMehr, vorpruefung.ListAbzug[0])
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugPrognose, vorpruefung.ListAbzug[0])
	} else {
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAuswahl, "Neueste MA")
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY1, 8)
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY2, 0)
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungMonateY3, 0)
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugSaldo, vorpruefung.ListAbzug[0])
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugMehr, vorpruefung.ListAbzug[0])
		_ = setValByNamedRange(f, vorpruefung.FieldMAPruefungAbzugPrognose, vorpruefung.ListAbzug[0])
	}

	return f.Save()
}

func setVal(f *excelize.File, sheet, cell string, val interface{}) {
	switch v := val.(type) {
	case time.Time:
		if !v.IsZero() {
			delta := v.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC))
			excelDate := float64(delta / (24 * time.Hour))
			_ = f.SetCellValue(sheet, cell, excelDate)
		}
	case float64:
		if v != 0 {
			_ = f.SetCellValue(sheet, cell, v)
		}
	case string:
		if v != "" {
			_ = f.SetCellValue(sheet, cell, v)
		}
	case *string:
		if v != nil && *v != "" {
			_ = f.SetCellValue(sheet, cell, *v)
		}
	case *int:
		if v != nil {
			_ = f.SetCellValue(sheet, cell, *v)
		}
	case *bool:
		if v != nil {
			if *v {
				_ = f.SetCellValue(sheet, cell, "Ja")
			} else {
				_ = f.SetCellValue(sheet, cell, "Nein")
			}
		}
	}
}

func setValByNamedRange(f *excelize.File, field vorpruefung.InputField, val interface{}) error {
	names := f.GetDefinedName()
	for _, n := range names {
		if n.Name == field.NamedRange {
			// n.RefersTo looks like "'I. Dashboard'!$C$5" or "I. Dashboard!$C$5"
			parts := strings.Split(n.RefersTo, "!")
			if len(parts) == 2 {
				sheet := strings.Trim(parts[0], "'")
				cell := strings.ReplaceAll(parts[1], "$", "")
				setVal(f, sheet, cell, val)
				return nil
			}
		}
	}
	return fmt.Errorf("NamedRange %s nicht gefunden", field.NamedRange)
}

func fillDashboard(f *excelize.File, d DashboardData) {
	sheet := constants.VPSheetDASHBOARD
	_ = setValByNamedRange(f, vorpruefung.FieldDashProjektnummer, d.Projektnummer)
	_ = setValByNamedRange(f, vorpruefung.FieldDashVorprojekt, d.Vorprojekt)
	_ = setValByNamedRange(f, vorpruefung.FieldDashProjekttitel, d.Projekttitel)
	_ = setValByNamedRange(f, vorpruefung.FieldDashProjekttraeger, d.Projekttraeger)
	_ = setValByNamedRange(f, vorpruefung.FieldDashBerichtswaehrung, d.Berichtswaehrung)
	_ = setValByNamedRange(f, vorpruefung.FieldDashProjektstart, d.Projektstart)
	_ = setValByNamedRange(f, vorpruefung.FieldDashProjektende, d.Projektende)

	if d.Vorprojekt != nil && *d.Vorprojekt {
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPNummer, d.Vorprojektnummer)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPBerichtswaehrung, d.VPBerichtswaehrung)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPEnde, d.Vorprojektende)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPWechselkurs, d.VPWechselkurs)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPSaldoLC, d.VPSaldoLC)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPSaldoEUR, d.VPSaldoEUR)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPFolgeprojektstart, d.VPFolgeprojektstart)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPFolgeWechselkurs, d.VPWechselkurs)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPFolgeSaldoLC, d.VPSaldoLC)
		_ = setValByNamedRange(f, vorpruefung.FieldDashVPFolgeSaldoEUR, d.VPSaldoEUR)
	}

	for i, v := range d.DocChecklist {
		if i > (RowDashChecklistEnd - RowDashChecklistStart) {
			break
		}
		cell, _ := excelize.CoordinatesToCellName(ColDashChecklist, RowDashChecklistStart+i)
		setVal(f, sheet, cell, v)
	}
}

func fillKMW(f *excelize.File, tranchen []KMWTranche) {
	for i, kr := range tranchen {
		// table row index starts at 1, so index is i+1
		idx := i + 1
		_ = setValByNamedRange(f, vorpruefung.FieldKMWPeriode(idx), kr.Periode)
		_ = setValByNamedRange(f, vorpruefung.FieldKMWWaehrung(idx), kr.Waehrung)
		_ = setValByNamedRange(f, vorpruefung.FieldKMWBetrag(idx), kr.Betrag)
		_ = setValByNamedRange(f, vorpruefung.FieldKMWDatum(idx), kr.Datum)
	}
}

func fillMA(f *excelize.File, periods []MAPeriod, budget *BudgetData) {
	if budget == nil {
		return
	}
	sheet := constants.VPSheetMA

	maDataRows := len(vorpruefung.ListKostenkategorien)
	maBlockHeight := 10 + maDataRows + 8 // header+data+footer
	rowOffsetEbene := maBlockHeight + 2

	for p, mp := range periods {
		col := ColMAStart + p*ColMAStep

		cVON, _ := excelize.CoordinatesToCellName(col, RowMAVon)
		cBIS, _ := excelize.CoordinatesToCellName(col, RowMABis)
		cKurs, _ := excelize.CoordinatesToCellName(col, RowMAKurs)
		setVal(f, sheet, cVON, mp.Von)
		setVal(f, sheet, cBIS, mp.Bis)
		setVal(f, sheet, cKurs, mp.OandaKurs)

		// Wir iterieren über die 3 Ebenen (jeweils mit dem rowOffsetEbene verschoben)
		for level := 0; level < 3; level++ {
			offsetR := level * rowOffsetEbene
			colT := col - 1
			row := 9 + offsetR // Header is on row 9 + offsetR

			for i, cat := range vorpruefung.ListKostenkategorien {
				// Schreibe Kategorie in die linke Spalte (colT)
				cLabel, _ := excelize.CoordinatesToCellName(colT, row+1+i)
				setVal(f, sheet, cLabel, cat)

				if level == 0 { // Werte nur in Ebene 1 setzen
					if v, exists := mp.KategorienLC[cat]; exists && v != 0 {
						cVal, _ := excelize.CoordinatesToCellName(colT+1, row+1+i)
						setVal(f, sheet, cVal, v)
					}
				}
			}

			if level == 0 {
				rEigen := row + len(vorpruefung.ListKostenkategorien) + RowMAOffsetEigen
				rDritt := row + len(vorpruefung.ListKostenkategorien) + RowMAOffsetDritt

				cEigen, _ := excelize.CoordinatesToCellName(col, rEigen)
				cDritt, _ := excelize.CoordinatesToCellName(col, rDritt)
				setVal(f, sheet, cEigen, mp.EigenLC)
				setVal(f, sheet, cDritt, mp.DrittLC)
			}
		}
	}
}

func addOrUpdateEinnahme(f *excelize.File, sheet, tableName, typ, geber string, lc, eur *float64) error {
	tables, err := f.GetTables(sheet)
	if err != nil {
		return err
	}
	var tRange string
	for _, t := range tables {
		if t.Name == tableName {
			tRange = t.Range
			break
		}
	}
	if tRange == "" {
		return fmt.Errorf("Tabelle %s nicht gefunden", tableName)
	}

	coords := strings.Split(tRange, ":")
	colStart, rowStart, _ := excelize.CellNameToCoordinates(coords[0])
	_, rowEnd, _ := excelize.CellNameToCoordinates(coords[1])

	targetRow := -1
	emptyRow := -1

	for r := rowStart + 1; r <= rowEnd; r++ {
		cTyp, _ := excelize.CoordinatesToCellName(colStart, r)
		cGeber, _ := excelize.CoordinatesToCellName(colStart+1, r)

		valTyp, _ := f.GetCellValue(sheet, cTyp)
		valGeber, _ := f.GetCellValue(sheet, cGeber)

		if valTyp == typ && valGeber == geber {
			targetRow = r
			break
		}
		if valTyp == "" && emptyRow == -1 {
			emptyRow = r
		}
	}

	if targetRow == -1 {
		if emptyRow != -1 {
			targetRow = emptyRow
		} else {
			targetRow = rowEnd
			if err := f.InsertRows(sheet, targetRow, 1); err != nil {
				return err
			}
		}
	}

	cTyp, _ := excelize.CoordinatesToCellName(colStart, targetRow)
	cGeber, _ := excelize.CoordinatesToCellName(colStart+1, targetRow)
	cLC, _ := excelize.CoordinatesToCellName(colStart+2, targetRow)
	cEUR, _ := excelize.CoordinatesToCellName(colStart+3, targetRow)

	setVal(f, sheet, cTyp, typ)
	if geber != "" {
		setVal(f, sheet, cGeber, geber)
	}
	if lc != nil {
		setVal(f, sheet, cLC, *lc)
	}
	if eur != nil {
		setVal(f, sheet, cEUR, *eur)
	}

	return nil
}

func fillFB(f *excelize.File, periods []FBPeriod, budget *BudgetData) {
	if budget == nil {
		return
	}
	sheet := constants.VPSheetFINANZBERICHTE
	tables, err := f.GetTables(sheet)
	if err != nil {
		return
	}

	// Map TableName -> Range (e.g. "TblFB_Ausgaben_1" -> "B18:F25")
	tableMap := make(map[string]string)
	for _, t := range tables {
		tableMap[t.Name] = t.Range
	}

	for p, fp := range periods {
		colStart := ColFBStart + p*ColFBStep
		cLabel, _ := excelize.ColumnNumberToName(colStart)
		cInput, _ := excelize.ColumnNumberToName(colStart + 1)

		setVal(f, sheet, fmt.Sprintf("%s%d", cInput, RowFBVon), fp.Von)
		setVal(f, sheet, fmt.Sprintf("%s%d", cInput, RowFBBis), fp.Bis)

		// Einnahmetypen aus dem Generator importieren und schreiben (ab Row 12)
		for i, tn := range vorpruefung.TYPE_NAMES {
			setVal(f, sheet, fmt.Sprintf("%s%d", cLabel, RowFBEinnahmen+i), tn)
		}

		// 1. Ausgaben Tabelle
		tNameAusg := fmt.Sprintf("Ausgaben_%d", p+1)
		if rng, ok := tableMap[tNameAusg]; ok {
			coords := strings.Split(rng, ":")
			col, row, _ := excelize.CellNameToCoordinates(coords[0])
			// header = row, data starts at row + 1
			for i, id := range budget.AusgabenIDs {
				// ID eintragen in die Label-Spalte
				cellID, _ := excelize.CoordinatesToCellName(col, row+1+i)
				setVal(f, sheet, cellID, id)

				if v, exists := fp.AusgabenByID[id]; exists && v != 0 {
					cell, _ := excelize.CoordinatesToCellName(col+1, row+1+i)
					setVal(f, sheet, cell, v)
				}
			}
			// Bank-Aufschlüsselung (liegt einige Zeilen unter der Ausgaben-Tabelle)
			// Wir berechnen den Offset ab dem Tabellen-Ende
			_, rowEnd, _ := excelize.CellNameToCoordinates(coords[1])
			cellBank, _ := excelize.CoordinatesToCellName(col+1, rowEnd+OffsetFBBank)
			cellKasse, _ := excelize.CoordinatesToCellName(col+1, rowEnd+OffsetFBKasse)
			cellSonstiges, _ := excelize.CoordinatesToCellName(col+1, rowEnd+OffsetFBSonstiges)

			if fp.BankLC != nil {
				setVal(f, sheet, cellBank, *fp.BankLC)
			}
			if fp.KasseLC != nil {
				setVal(f, sheet, cellKasse, *fp.KasseLC)
			}
			if fp.SonstigesLC != nil {
				setVal(f, sheet, cellSonstiges, *fp.SonstigesLC)
			}
		}

		// 2. Einnahmen Tabelle 1 (KMW)
		tNameT1 := fmt.Sprintf("Einnahmen_%d", p+1)
		for _, e := range fp.Einnahmen1 {
			_ = addOrUpdateEinnahme(f, sheet, tNameT1, e.Typ, e.Geber, e.LC, e.EUR)
		}

		// 3. Einnahmen Tabelle 2 (Eigen/Dritt)
		tNameT2 := fmt.Sprintf("Einnahmen_WK_%d", p+1)
		for _, e := range fp.EinnahmenWK {
			_ = addOrUpdateEinnahme(f, sheet, t[@registry.go (406:455)](file:///home/ardit/repos/MyAutomationSuite/sidecars/Vorpruefung/pkg/vorpruefung/registry.go#L406:455) Ich raffe es nicht was soll dieser Scheiß?NameT2, e.Typ, e.Geber, e.LC, e.EUR)
		}
	}
}

func fillBudget(f *excelize.File, budget *BudgetData) {
	if budget == nil {
		return
	}
	sheet := constants.VPSheetBUDGET

	fillIncFields := func(inc *IncomeRow, lc, y1, y2, y3, eur vorpruefung.InputField) {
		if inc == nil {
			return
		}
		if inc.LC != nil {
			_ = setValByNamedRange(f, lc, *inc.LC)
		}
		if inc.Y1 != nil {
			_ = setValByNamedRange(f, y1, *inc.Y1)
		}
		if inc.Y2 != nil {
			_ = setValByNamedRange(f, y2, *inc.Y2)
		}
		if inc.Y3 != nil {
			_ = setValByNamedRange(f, y3, *inc.Y3)
		}
		if inc.EUR != nil {
			_ = setValByNamedRange(f, eur, *inc.EUR)
		}
	}

	fillIncFields(budget.Eigenmittel, vorpruefung.FieldBudgetEigenmittelLC, vorpruefung.FieldBudgetEigenmittelY1, vorpruefung.FieldBudgetEigenmittelY2, vorpruefung.FieldBudgetEigenmittelY3, vorpruefung.FieldBudgetEigenmittelEUR)
	fillIncFields(budget.KMWMittel, vorpruefung.FieldBudgetKMWLC, vorpruefung.FieldBudgetKMWY1, vorpruefung.FieldBudgetKMWY2, vorpruefung.FieldBudgetKMWY3, vorpruefung.FieldBudgetKMWEUR)

	if budget.DrittmittelY1 != nil {
		_ = setValByNamedRange(f, vorpruefung.FieldBudgetDrittmittelY1, *budget.DrittmittelY1)
	}
	if budget.DrittmittelY2 != nil {
		_ = setValByNamedRange(f, vorpruefung.FieldBudgetDrittmittelY2, *budget.DrittmittelY2)
	}
	if budget.DrittmittelY3 != nil {
		_ = setValByNamedRange(f, vorpruefung.FieldBudgetDrittmittelY3, *budget.DrittmittelY3)
	}

	_ = setValByNamedRange(f, vorpruefung.FieldBudgetReserveFreigabe, budget.ReserveFreigabe)

	tables, _ := f.GetTables(sheet)
	tableMap := make(map[string]string)
	for _, t := range tables {
		tableMap[t.Name] = t.Range
	}

	// TblBudgetAusgaben
	if rng, ok := tableMap["TblBudgetAusgaben"]; ok {
		coords := strings.Split(rng, ":")
		col, row, _ := excelize.CellNameToCoordinates(coords[0])
		for i, a := range budget.Ausgaben {
			r := row + 1 + i

			if a.Kategorie != "" {
				c, _ := excelize.CoordinatesToCellName(col, r)
				setVal(f, sheet, c, a.Kategorie)
			}
			if a.ID != "" {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetID, r)
				setVal(f, sheet, c, a.ID)
			}
			if a.Position != "" {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetPosition, r)
				setVal(f, sheet, c, a.Position)
			}

			if a.LC != nil {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetLC, r)
				setVal(f, sheet, c, *a.LC)
			}
			if a.Y1 != nil {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetY1, r)
				setVal(f, sheet, c, *a.Y1)
			}
			if a.Y2 != nil {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetY2, r)
				setVal(f, sheet, c, *a.Y2)
			}
			if a.Y3 != nil {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetY3, r)
				setVal(f, sheet, c, *a.Y3)
			}
			if a.EUR != nil {
				c, _ := excelize.CoordinatesToCellName(col+ColBudgetOffsetEUR, r)
				setVal(f, sheet, c, *a.EUR)
			}
		}
	}

	// TblDrittmittel
	if rng, ok := tableMap["TblDrittmittel"]; ok {
		coords := strings.Split(rng, ":")
		col, row, _ := excelize.CellNameToCoordinates(coords[0])
		for i := 0; i < len(budget.DrittGeber); i++ {
			geb := budget.DrittGeber[i]
			r := row + 1 + i
			cName, _ := excelize.CoordinatesToCellName(col, r)
			cLC, _ := excelize.CoordinatesToCellName(col+1, r)
			cEUR, _ := excelize.CoordinatesToCellName(col+2, r)
			setVal(f, sheet, cName, geb.Geber)
			if geb.LC != nil {
				setVal(f, sheet, cLC, *geb.LC)
			}
			if geb.EUR != nil {
				setVal(f, sheet, cEUR, *geb.EUR)
			}
		}
		if budget.DrittSonstiges != nil {
			r := row + 1 + len(budget.DrittGeber)
			cLC, _ := excelize.CoordinatesToCellName(col+1, r)
			cEUR, _ := excelize.CoordinatesToCellName(col+2, r)
			if budget.DrittSonstiges.LC != nil {
				setVal(f, sheet, cLC, *budget.DrittSonstiges.LC)
			}
			if budget.DrittSonstiges.EUR != nil {
				setVal(f, sheet, cEUR, *budget.DrittSonstiges.EUR)
			}
		}
	}
}
