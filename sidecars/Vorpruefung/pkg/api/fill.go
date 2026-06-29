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
	if data.FBPruefung != nil {
		sheet := constants.VPSheetFB_PRUEFUNG
		setVal(f, sheet, "C6", data.FBPruefung.Auswahl)
		// Die Toggle-Zellen befinden sich in F18, F19, F20 (im Template oft dynamisch,
		// aber sie sind relativ zur KMW-Tabelle). Aktuell setzen wir sie über die defaults,
		// wenn nicht explizit gesetzt.
	}
	if data.MAPruefung != nil {
		sheet := constants.VPSheetMA_PRUEFUNG
		setVal(f, sheet, "C6", data.MAPruefung.Auswahl)
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

func fillDashboard(f *excelize.File, d DashboardData) {
	sheet := constants.VPSheetDASHBOARD
	setVal(f, sheet, "C5", d.Projektnummer)
	setVal(f, sheet, "E5", d.Vorprojekt)
	setVal(f, sheet, "C6", d.Projekttitel)
	setVal(f, sheet, "C7", d.Projekttraeger)
	setVal(f, sheet, "E7", d.Berichtswaehrung)
	setVal(f, sheet, "C8", d.Projektstart)
	setVal(f, sheet, "E8", d.Projektende)

	if d.Vorprojekt != nil && *d.Vorprojekt {
		setVal(f, sheet, "C10", d.Vorprojektnummer)
		setVal(f, sheet, "E10", d.VPBerichtswaehrung)
		setVal(f, sheet, "C11", d.Vorprojektende)
		setVal(f, sheet, "E11", d.VPWechselkurs)
		setVal(f, sheet, "C12", d.VPSaldoLC)
		setVal(f, sheet, "E12", d.VPSaldoEUR)
		setVal(f, sheet, "C13", d.VPFolgeprojektstart)
		setVal(f, sheet, "E13", d.VPWechselkurs)
		setVal(f, sheet, "C14", d.VPSaldoLC)
		setVal(f, sheet, "E14", d.VPSaldoEUR)
	}

	for i, v := range d.DocChecklist {
		if i >= 7 {
			break
		}
		cell, _ := excelize.CoordinatesToCellName(4, 16+i)
		setVal(f, sheet, cell, v)
	}
}

func fillKMW(f *excelize.File, tranchen []KMWTranche) {
	sheet := constants.VPSheetKMW_MITTEL
	for i, kr := range tranchen {
		row := 5 + i
		if row > 22 {
			break
		}
		cPeriode, _ := excelize.CoordinatesToCellName(2, row)
		cWaehrung, _ := excelize.CoordinatesToCellName(3, row)
		cBetrag, _ := excelize.CoordinatesToCellName(4, row)
		cDatum, _ := excelize.CoordinatesToCellName(5, row)

		setVal(f, sheet, cPeriode, kr.Periode)
		setVal(f, sheet, cWaehrung, kr.Waehrung)
		setVal(f, sheet, cBetrag, kr.Betrag)
		setVal(f, sheet, cDatum, kr.Datum)
	}
}

func fillMA(f *excelize.File, periods []MAPeriod, budget *BudgetData) {
	if budget == nil {
		return
	}
	sheet := constants.VPSheetMA
	tables, err := f.GetTables(sheet)
	if err != nil {
		return
	}

	tableMap := make(map[string]string)
	for _, t := range tables {
		tableMap[t.Name] = t.Range
	}

	for p, mp := range periods {
		col := 3 + p*4 // C, G, K...

		cVON, _ := excelize.CoordinatesToCellName(col, 5)
		cBIS, _ := excelize.CoordinatesToCellName(col, 6)
		cKurs, _ := excelize.CoordinatesToCellName(col, 8)
		setVal(f, sheet, cVON, mp.Von)
		setVal(f, sheet, cBIS, mp.Bis)
		setVal(f, sheet, cKurs, mp.OandaKurs)

		// Ebene 1
		tNameL1 := fmt.Sprintf("TblMA_L1_%d", p+1)
		if rng, ok := tableMap[tNameL1]; ok {
			coords := strings.Split(rng, ":")
			colT, row, _ := excelize.CellNameToCoordinates(coords[0])
			for i, a := range budget.Ausgaben {
				// Schreibe Kategorie in die linke Spalte (colT)
				cLabel, _ := excelize.CoordinatesToCellName(colT, row+1+i)
				setVal(f, sheet, cLabel, a.Kategorie)

				if v, exists := mp.KategorienLC[a.Kategorie]; exists && v != 0 {
					cVal, _ := excelize.CoordinatesToCellName(colT+1, row+1+i)
					setVal(f, sheet, cVal, v)
				}
			}

			// We can also find the total row and offsets for Eigen, Dritt, KMW.
			// The data rows take len(budget.Ausgaben).
			// The offsets are: Eigen = row + len(budget.Ausgaben) + 4
			// Dritt = row + len(...) + 5
			rEigen := row + len(budget.Ausgaben) + 4
			rDritt := row + len(budget.Ausgaben) + 5

			cEigen, _ := excelize.CoordinatesToCellName(col, rEigen)
			cDritt, _ := excelize.CoordinatesToCellName(col, rDritt)
			setVal(f, sheet, cEigen, mp.EigenLC)
			setVal(f, sheet, cDritt, mp.DrittLC)
		}

		// 3 Finanzierungsarten (Tabelle offsets depend on row calculation, easier is to just address them dynamically if possible)
		// but since we only have fixed row references for EigenLC, we might need to find them or compute them.
		// Actually, mp.EigenLC and mp.DrittLC are mapped in MAPeriod struct. Let's see what is there.
		tNameL2 := fmt.Sprintf("TblMA_L2_%d", p+1)
		if rng, ok := tableMap[tNameL2]; ok {
			coords := strings.Split(rng, ":")
			colT, row, _ := excelize.CellNameToCoordinates(coords[0])
			for i, a := range budget.Ausgaben {
				cLabel, _ := excelize.CoordinatesToCellName(colT, row+1+i)
				setVal(f, sheet, cLabel, a.Kategorie)
			}
		}

		// Ebene 3
		tNameL3 := fmt.Sprintf("TblMA_L3_%d", p+1)
		if rng, ok := tableMap[tNameL3]; ok {
			coords := strings.Split(rng, ":")
			colT, row, _ := excelize.CellNameToCoordinates(coords[0])
			for i, a := range budget.Ausgaben {
				cLabel, _ := excelize.CoordinatesToCellName(colT, row+1+i)
				setVal(f, sheet, cLabel, a.Kategorie)
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
		colStart := 2 + p*7
		cLabel, _ := excelize.ColumnNumberToName(colStart)     // B, I, P...
		cInput, _ := excelize.ColumnNumberToName(colStart + 1) // C, J, Q...

		setVal(f, sheet, fmt.Sprintf("%s5", cInput), fp.Von)
		setVal(f, sheet, fmt.Sprintf("%s6", cInput), fp.Bis)

		// Einnahmetypen aus dem Generator importieren und schreiben (ab Row 12)
		for i, tn := range vorpruefung.TYPE_NAMES {
			setVal(f, sheet, fmt.Sprintf("%s%d", cLabel, 12+i), tn)
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
			cellBank, _ := excelize.CoordinatesToCellName(col+1, rowEnd+6)
			cellKasse, _ := excelize.CoordinatesToCellName(col+1, rowEnd+7)
			cellSonstiges, _ := excelize.CoordinatesToCellName(col+1, rowEnd+8)

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
			_ = addOrUpdateEinnahme(f, sheet, tNameT2, e.Typ, e.Geber, e.LC, e.EUR)
		}
	}
}

func fillBudget(f *excelize.File, budget *BudgetData) {
	if budget == nil {
		return
	}
	sheet := constants.VPSheetBUDGET

	fillInc := func(row int, inc *IncomeRow) {
		if inc == nil {
			return
		}
		if inc.LC != nil {
			c, _ := excelize.CoordinatesToCellName(5, row)
			setVal(f, sheet, c, *inc.LC)
		}
		if inc.Y1 != nil {
			c, _ := excelize.CoordinatesToCellName(6, row)
			setVal(f, sheet, c, *inc.Y1)
		}
		if inc.Y2 != nil {
			c, _ := excelize.CoordinatesToCellName(7, row)
			setVal(f, sheet, c, *inc.Y2)
		}
		if inc.Y3 != nil {
			c, _ := excelize.CoordinatesToCellName(8, row)
			setVal(f, sheet, c, *inc.Y3)
		}
		if inc.EUR != nil {
			c, _ := excelize.CoordinatesToCellName(9, row)
			setVal(f, sheet, c, *inc.EUR)
		}
	}

	fillInc(7, budget.Eigenmittel)
	fillInc(13, budget.KMWMittel)

	if budget.DrittmittelY1 != nil {
		setVal(f, sheet, "F10", *budget.DrittmittelY1)
	}
	if budget.DrittmittelY2 != nil {
		setVal(f, sheet, "G10", *budget.DrittmittelY2)
	}
	if budget.DrittmittelY3 != nil {
		setVal(f, sheet, "H10", *budget.DrittmittelY3)
	}

	setVal(f, sheet, "K5", budget.ReserveFreigabe)

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

			// col = Label/Kategorie, col+1 = ID, col+2 = Position
			if a.Kategorie != "" {
				c, _ := excelize.CoordinatesToCellName(col, r)
				setVal(f, sheet, c, a.Kategorie)
			}
			if a.ID != "" {
				c, _ := excelize.CoordinatesToCellName(col+1, r)
				setVal(f, sheet, c, a.ID)
			}
			if a.Position != "" {
				c, _ := excelize.CoordinatesToCellName(col+2, r)
				setVal(f, sheet, c, a.Position)
			}

			// Betrag (LC) ist col + 3 (Spalte E = 5, wenn col = 2 (B))
			if a.LC != nil {
				c, _ := excelize.CoordinatesToCellName(col+3, r)
				setVal(f, sheet, c, *a.LC)
			}
			if a.Y1 != nil {
				c, _ := excelize.CoordinatesToCellName(col+4, r)
				setVal(f, sheet, c, *a.Y1)
			}
			if a.Y2 != nil {
				c, _ := excelize.CoordinatesToCellName(col+5, r)
				setVal(f, sheet, c, *a.Y2)
			}
			if a.Y3 != nil {
				c, _ := excelize.CoordinatesToCellName(col+6, r)
				setVal(f, sheet, c, *a.Y3)
			}
			if a.EUR != nil {
				c, _ := excelize.CoordinatesToCellName(col+7, r)
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
