package report

import (
	"bytes"
	"fmt"
	"shared/constants"
	"strings"

	"github.com/xuri/excelize/v2"
)

// NewExcelReportByLanguage wählt automatisch die korrekte Vorlagendatei aus dem eingebetteten Vorlagen-Ordner.
func NewExcelReportByLanguage(language string) (*ExcelReport, error) {
	filename, ok := LanguageToTemplate[language]
	if !ok {
		return nil, fmt.Errorf("nicht unterstützte Sprache: %s", language)
	}

	path := "templates/" + filename
	return NewExcelReport(path)
}

func NewExcelReport(path string) (*ExcelReport, error) {
	data, err := getTemplateBytes(path)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Vorlage %s: %w", path, err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Wir suchen nach dem richtigen Finanzberichts-Sheet anhand der Konstanten.
	// Fallback: Das erste Arbeitsblatt, falls es abweicht.
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("keine Arbeitsblätter in der Datei gefunden")
	}

	sheetName := sheets[0] // Fallback
	for _, s := range sheets {
		for _, validName := range constants.FBSheetNames {
			if s == validName {
				sheetName = s
				break
			}
		}
	}

	report := &ExcelReport{
		file:         f,
		sheet:        sheetName,
		CatStartRows: make(map[int]int),
		CatEndRows:   make(map[int]int),
	}

	sheetRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Lesen der Zeilen: %w", err)
	}

	// Zeilen vorberechnen (bevor irgendwelche Formeln die Textwerte löschen)
	for catID := 1; catID <= 8; catID++ {
		startPrefix := fmt.Sprintf("%d.", catID)
		startRow := findRowByColB(sheetRows, func(val string) bool { return val == startPrefix })
		if startRow != -1 {
			report.CatStartRows[catID] = startRow
		}

		var endRow int
		if catID == 8 {
			endRow = findRowByColB(sheetRows, func(val string) bool {
				v := strings.ToUpper(val)
				return strings.Contains(v, "SUMME") || strings.Contains(v, "TOTAL")
			}) - 1
			// Wenn TOTAL vor Kategorie 8 gefunden wird (z.B. Zwischensummen),
			// suchen wir explizit nach der globalen Summe weiter unten
			if endRow != -2 && endRow < startRow {
				for i := startRow; i < len(sheetRows); i++ {
					if len(sheetRows[i]) > 1 && (strings.Contains(strings.ToUpper(sheetRows[i][1]), "SUMME") || strings.Contains(strings.ToUpper(sheetRows[i][1]), "TOTAL")) {
						endRow = i
						break
					}
				}
			}
		} else {
			nextPrefix := fmt.Sprintf("%d.", catID+1)
			endRow = findRowByColB(sheetRows, func(val string) bool { return val == nextPrefix }) - 1
		}
		if endRow != -2 {
			report.CatEndRows[catID] = endRow
		}
	}

	report.GlobalSumRow = findRowByColB(sheetRows, func(val string) bool {
		v := strings.ToUpper(val)
		return strings.Contains(v, "SUMME") || strings.Contains(v, "TOTAL")
	})
	// Auch hier sicherstellen, dass wir die echte globale Summe ganz unten finden
	if report.GlobalSumRow != -1 && report.GlobalSumRow < report.CatStartRows[8] {
		for i := report.CatStartRows[8]; i < len(sheetRows); i++ {
			if len(sheetRows[i]) > 1 && (strings.Contains(strings.ToUpper(sheetRows[i][1]), "SUMME") || strings.Contains(strings.ToUpper(sheetRows[i][1]), "TOTAL")) {
				report.GlobalSumRow = i + 1
				break
			}
		}
	}

	report.BankRow = findRowByColB(sheetRows, func(val string) bool { return strings.Contains(strings.ToUpper(val), "BANK") })
	report.KasseRow = findRowByColB(sheetRows, func(val string) bool {
		return strings.Contains(strings.ToUpper(val), "KASSE") || strings.Contains(strings.ToUpper(val), "CASH")
	})
	report.SonstRow = findRowByColB(sheetRows, func(val string) bool {
		return strings.Contains(strings.ToUpper(val), "SONSTIGES") || strings.Contains(strings.ToUpper(val), "OTHER")
	})

	return report, nil
}

// Close schließt die Datei und gibt Ressourcen frei.
func (r *ExcelReport) Close() error {
	return r.file.Close()
}

// SaveAs speichert die Datei unter einem neuen Pfad ab (Template bleibt unberührt).
func (r *ExcelReport) SaveAs(path string) error {
	return r.file.SaveAs(path)
}

// --- Layout-Konfiguration (Magic Numbers) ---
const (
	RatesStartRow   = 14
	RatesCountBlock = 18
	RatesEndRow     = 31 // 14 + 18 - 1
	ColRates1Start  = "L"
	ColRates2Start  = "S"

	RowSaldovortrag  = 15
	RowEigenleistung = 16
	RowDrittmittel   = 17
	RowKMWMittel     = 18
	RowZinsertraege  = 19

	VLookupBaseIdx = 28
	VLookupMult    = 2
)

// Hilfsfunktion: Verschiebt alle gemerkten Zeilen-Indizes, wenn physikalisch Zeilen gelöscht/hinzugefügt werden.
func (r *ExcelReport) shiftRows(delta int, afterRow int) {
	for k, v := range r.CatStartRows {
		if v > afterRow {
			r.CatStartRows[k] = v + delta
		}
	}
	for k, v := range r.CatEndRows {
		if v > afterRow {
			r.CatEndRows[k] = v + delta
		}
	}
	if r.GlobalSumRow > afterRow {
		r.GlobalSumRow += delta
	}
	if r.BankRow > afterRow {
		r.BankRow += delta
	}
	if r.KasseRow > afterRow {
		r.KasseRow += delta
	}
	if r.SonstRow > afterRow {
		r.SonstRow += delta
	}
}

func (r *ExcelReport) setIfObj(cell string, val interface{}) error {
	if val != nil && val != "" {
		if err := r.file.SetCellValue(r.sheet, cell, val); err != nil {
			return fmt.Errorf("fehler beim Setzen von Zelle %s: %w", cell, err)
		}
	}
	return nil
}

func (r *ExcelReport) UpdateData(data ReportData) error {
	if err := r.updateStaticHeaders(data); err != nil {
		return err
	}
	if err := r.updateRates(data); err != nil {
		return err
	}
	if err := r.updateFundings(data); err != nil {
		return err
	}
	if err := r.updateAllCategories(data); err != nil {
		return err
	}
	if err := r.updateGlobalSum(); err != nil {
		return err
	}
	if err := r.updateSaldoBreakdown(data); err != nil {
		return err
	}

	// Excel Formeln neu berechnen
	r.file.UpdateLinkedValue()

	if err := r.ApplyOptions(data.Options); err != nil {
		return err
	}

	return nil
}

func (r *ExcelReport) updateStaticHeaders(data ReportData) error {
	var err error
	set := func(cell string, val interface{}) {
		if err == nil {
			err = r.setIfObj(cell, val)
		}
	}
	set("E2", data.Sprache)
	set("E3", data.Lokalwaehrung)
	set("D5", data.Projektnummer)
	set("D6", data.Projekttitel)
	set("E8", data.ProjektlaufzeitBeginn)
	set("G8", data.ProjektlaufzeitEnde)
	set("E9", data.AktuellerBerichtszeitraumBeginn)
	set("G9", data.AktuellerBerichtszeitraumEnde)
	return err
}

func (r *ExcelReport) updateRates(data ReportData) error {
	for i := 0; i < RatesCountBlock; i++ {
		row := RatesStartRow + i
		rate := data.Rates[i]
		// Anstatt SetSheetRow nutzen wir SetCellValue für jede Zelle einzeln.
		// SetSheetRow löscht nämlich heimlich die Formeln in den umliegenden Zellen der gleichen Zeile
		// oder überschreibt sie mit leeren Werten, da es ein Array variabler Länge schreibt.
		if err := r.setIfObj(fmt.Sprintf("%s%d", ColRates1Start, row), rate.Datum); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("M%d", row), rate.EUR); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("N%d", row), rate.LW); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("O%d", row), rate.WK); err != nil {
			return err
		}
	}

	for i := 0; i < RatesCountBlock; i++ {
		idx := RatesCountBlock + i
		row := RatesStartRow + i
		rate := data.Rates[idx]
		if err := r.setIfObj(fmt.Sprintf("%s%d", ColRates2Start, row), rate.Datum); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("T%d", row), rate.EUR); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("U%d", row), rate.LW); err != nil {
			return err
		}
		if err := r.setIfObj(fmt.Sprintf("V%d", row), rate.WK); err != nil {
			return err
		}
	}
	return nil
}

func (r *ExcelReport) updateFundings(data ReportData) error {
	var err error
	setFunding := func(row int, rec FundingRecord) {
		if err != nil {
			return
		}
		if errRow := r.file.SetSheetRow(r.sheet, fmt.Sprintf("D%d", row), &[]interface{}{rec.Budget, rec.EinnahmenBZ, rec.EinnahmenGS}); errRow != nil {
			err = errRow
			return
		}
		err = r.setIfObj(fmt.Sprintf("H%d", row), rec.Begruendung)
	}
	setFunding(RowSaldovortrag, data.Saldovortrag)
	setFunding(RowEigenleistung, data.Eigenleistung)
	setFunding(RowDrittmittel, data.Drittmittel)
	setFunding(RowKMWMittel, data.KMWMittel)
	setFunding(RowZinsertraege, data.Zinsertraege)
	return err
}

func (r *ExcelReport) generateCategoryFormula(startRow, catID int) string {
	return fmt.Sprintf(`IF(ROW()<ROW($B$%d),"",IF(ROW()=ROW($B$%d),"%d.","%d."&(ROW()-ROW($B$%d))))`, startRow, startRow, catID, catID, startRow)
}

func (r *ExcelReport) updateAllCategories(data ReportData) error {
	// Dynamische Kategorien von Unten nach Oben (Bottom-Up) verarbeiten
	for catID := 8; catID >= 1; catID-- {
		startRow, okStart := r.CatStartRows[catID]
		endRow, okEnd := r.CatEndRows[catID]
		if !okStart || !okEnd {
			continue // Kategorie nicht gefunden, überspringen
		}

		if catID >= 6 && catID <= 8 {
			// Kategorien 6-8 sind IMMER im Modus 0 (Kompakt)
			// Wir ignorieren Unterpositionen und Pufferzeilen.

			// Falls im Template doch Zwischenzeilen waren (Modus N), löschen wir sie alle weg
			if startRow != endRow {
				for rIdx := endRow; rIdx > startRow; rIdx-- {
					if err := r.file.RemoveRow(r.sheet, rIdx); err != nil {
						return err
					}
					r.shiftRows(-1, rIdx-1)
				}
				r.CatEndRows[catID] = startRow
			}

			// Budget einfach in den Header eintragen
			if budget, ok := data.HeaderBudgets[catID]; ok {
				r.setIfObj(fmt.Sprintf("D%d", startRow), budget)
			} else {
				r.setIfObj(fmt.Sprintf("D%d", startRow), 0)
			}
		} else {
			// Kategorien 1-5 sind IMMER im Modus N (Erweitert)
			items := data.Categories[catID]

			emptyCount := data.EmptyRows.Global
			if override, ok := data.EmptyRows.CategoryOverrides[catID]; ok {
				emptyCount = override
			}
			if emptyCount < 0 {
				emptyCount = 0
			}

			targetPos := len(items) + emptyCount

			// Schutz der Raten-Tabelle: Kategorien, die mit der Raten-Tabelle überlappen,
			// dürfen nicht so weit verkleinert werden, dass wir physikalisch Zeilen <= RatesEndRow löschen.
			if startRow <= RatesEndRow {
				minTarget := RatesEndRow - startRow
				if targetPos < minTarget {
					targetPos = minTarget
				}
			}

			if err := r.processCategory(catID, items, targetPos); err != nil {
				return fmt.Errorf("fehler in Kategorie %d: %w", catID, err)
			}
		}
	}
	return nil
}

func (r *ExcelReport) processCategory(catID int, items []CostItem, targetPos int) error {
	startRow := r.CatStartRows[catID]
	endRow := r.CatEndRows[catID]
	f := r.file
	s := r.sheet

	// Falls die Vorlage versehentlich Modus 0 ist (sollte bei 1-5 nie der Fall sein),
	// duplizieren wir sicherheitshalber die Header-Zeile, um eine Subtotal-Zeile zu schaffen.
	if startRow == endRow {
		if err := f.DuplicateRow(s, startRow); err != nil {
			return err
		}
		r.shiftRows(1, startRow)
		endRow = startRow + 1
		r.CatEndRows[catID] = endRow
	}

	currentPos := endRow - startRow - 1
	physicalEndRow := endRow

	if currentPos < targetPos {
		lastPos := endRow - 1
		for i := 0; i < targetPos-currentPos; i++ {
			if err := f.DuplicateRow(s, lastPos); err != nil {
				return err
			}
			r.shiftRows(1, lastPos)
			physicalEndRow++
		}
		endRow += (targetPos - currentPos)
		r.CatEndRows[catID] = physicalEndRow
	} else if currentPos > targetPos {
		for i := 0; i < currentPos-targetPos; i++ {
			rowIdx := endRow - 1
			if err := f.RemoveRow(s, rowIdx); err != nil {
				return err
			}
			r.shiftRows(-1, rowIdx-1)
			endRow--
			physicalEndRow--
		}
		r.CatEndRows[catID] = physicalEndRow
	}

	// Daten in die Zielzeilen eintragen (Batch-Insert für C bis F)
	for i := 0; i < targetPos; i++ {
		rowIdx := startRow + 1 + i
		var rowValues []interface{}
		var begruendung interface{}

		if i < len(items) {
			item := items[i]
			rowValues = []interface{}{item.Name, item.Budget, item.AusgabenBZ, item.AusgabenGS}
			begruendung = item.Begruendung
		} else {
			// Leere Zeilen: wir schreiben nur eine 0 in die Budget-Spalte (Index 1)
			rowValues = []interface{}{"", 0, "", ""}
			begruendung = ""
		}

		if err := f.SetSheetRow(s, fmt.Sprintf("C%d", rowIdx), &rowValues); err != nil {
			return err
		}

		// H separat einfügen, um die Formel in G intakt zu lassen
		if err := r.setIfObj(fmt.Sprintf("H%d", rowIdx), begruendung); err != nil {
			return err
		}

		// Nummerierungs-Formel (1.1, 1.2) für Spalte B injizieren
		formulaB := r.generateCategoryFormula(startRow, catID)
		f.SetCellDefault(s, fmt.Sprintf("B%d", rowIdx), "")
		f.SetCellFormula(s, fmt.Sprintf("B%d", rowIdx), formulaB)
	}

	// SUM Formeln der Zwischensumme aktualisieren
	sumStart := startRow + 1
	sumEnd := endRow - 1
	if sumEnd < sumStart {
		// Fallback falls targetPos == 0 (Keine Zwischenzeilen, summiere leeren Bereich)
		sumEnd = sumStart
	}

	f.SetCellValue(s, fmt.Sprintf("D%d", endRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("D%d", endRow), fmt.Sprintf("SUM(D%d:D%d)", sumStart, sumEnd)); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("E%d", endRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("E%d", endRow), fmt.Sprintf("SUM(E%d:E%d)", sumStart, sumEnd)); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("F%d", endRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("F%d", endRow), fmt.Sprintf("SUM(F%d:F%d)", sumStart, sumEnd)); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("G%d", endRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("G%d", endRow), fmt.Sprintf("IFERROR(F%d/D%d,0)", endRow, endRow)); err != nil {
		return err
	}

	return nil
}

func (r *ExcelReport) updateGlobalSum() error {
	if r.GlobalSumRow == -1 {
		return nil
	}

	var dSum, eSum, fSum string
	for i := 1; i <= 8; i++ {
		if _, ok := r.CatStartRows[i]; !ok {
			continue
		}

		var catSumRow int
		if r.CatStartRows[i] == r.CatEndRows[i] {
			catSumRow = r.CatStartRows[i]
		} else {
			catSumRow = r.CatEndRows[i]
		}

		if dSum != "" {
			dSum += "+"
			eSum += "+"
			fSum += "+"
		}
		dSum += fmt.Sprintf("D%d", catSumRow)
		eSum += fmt.Sprintf("E%d", catSumRow)
		fSum += fmt.Sprintf("F%d", catSumRow)
	}
	if dSum == "" {
		dSum = "0"
		eSum = "0"
		fSum = "0"
	}

	f := r.file
	s := r.sheet

	f.SetCellValue(s, fmt.Sprintf("D%d", r.GlobalSumRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("D%d", r.GlobalSumRow), dSum); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("E%d", r.GlobalSumRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("E%d", r.GlobalSumRow), eSum); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("F%d", r.GlobalSumRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("F%d", r.GlobalSumRow), fSum); err != nil {
		return err
	}
	f.SetCellValue(s, fmt.Sprintf("G%d", r.GlobalSumRow), 0)
	if err := f.SetCellFormula(s, fmt.Sprintf("G%d", r.GlobalSumRow), fmt.Sprintf("IFERROR(F%d/D%d,0)", r.GlobalSumRow, r.GlobalSumRow)); err != nil {
		return err
	}

	return nil
}

func (r *ExcelReport) updateSaldoBreakdown(data ReportData) error {
	var err error
	set := func(row int, val interface{}) {
		if row != -1 && err == nil {
			err = r.setIfObj(fmt.Sprintf("E%d", row), val)
		}
	}

	set(r.BankRow, data.Saldo.Bank)
	set(r.KasseRow, data.Saldo.Kasse)
	set(r.SonstRow, data.Saldo.Sonstiges)

	return err
}

func (r *ExcelReport) ApplyOptions(opts ReportOptions) error {
	// Blattschutz (Worksheet)
	if opts.ProtectSheet {
		err := r.file.ProtectSheet(r.sheet, &excelize.SheetProtectionOptions{
			AlgorithmName:       "SHA-512",
			Password:            opts.SheetPassword,
			SelectLockedCells:   opts.SelectLocked,
			SelectUnlockedCells: opts.SelectUnlocked,
			FormatCells:         opts.FormatCells,
			FormatColumns:       opts.FormatColumns,
			FormatRows:          opts.FormatRows,
			InsertColumns:       opts.InsertColumns,
			InsertRows:          opts.InsertRows,
			InsertHyperlinks:    opts.InsertHyperlinks,
			DeleteColumns:       opts.DeleteColumns,
			DeleteRows:          opts.DeleteRows,
			Sort:                opts.Sort,
			AutoFilter:          opts.Autofilter,
			PivotTables:         opts.PivotTables,
			EditObjects:         opts.EditObjects,
			EditScenarios:       opts.EditScenarios,
		})
		if err != nil {
			return err
		}
	} else {
		// Explizit entfernen, falls das Template geschützt war
		err := r.file.UnprotectSheet(r.sheet)
		if err != nil {
			return err
		}
	}

	// Mappenschutz (Workbook)
	if opts.ProtectWorkbook {
		err := r.file.ProtectWorkbook(&excelize.WorkbookProtectionOptions{
			AlgorithmName: "SHA-512",
			Password:      opts.WorkbookPassword,
			LockStructure: true,
		})
		if err != nil {
			return err
		}
	} else {
		// Explizit entfernen, falls das Template geschützt war
		err := r.file.UnprotectWorkbook()
		if err != nil {
			return err
		}
	}

	return nil
}
