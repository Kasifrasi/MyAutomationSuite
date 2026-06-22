package report

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

var templateCache sync.Map // map[string][]byte

// func PreloadAllTemplates() error
// Da wir hier einmalig HideColumns anwenden wollen, müssen wir die Options übergeben
func PreloadAllTemplates(globalOpts ReportOptions) error {
	var wg sync.WaitGroup
	var preloadErr error
	var errMu sync.Mutex

	for _, filename := range LanguageToTemplate {
		wg.Add(1)
		go func(fname string) {
			defer wg.Done()

			path := "templates/" + fname
			data, err := templateFiles.ReadFile(path)
			if err != nil {
				errMu.Lock()
				preloadErr = fmt.Errorf("fehler beim Preload von %s: %w", path, err)
				errMu.Unlock()
				return
			}

			// Einmalig die globalen Layout-Einstellungen für diese Vorlage anwenden
			f, err := excelize.OpenReader(bytes.NewReader(data))
			if err == nil {
				sheets := f.GetSheetList()
				if len(sheets) > 0 {
					mainSheet := sheets[0]

					// Spalten Q-V verstecken (Nur EINMAL pro Vorlage im RAM!)
					if globalOpts.HideColumns {
						for _, col := range []string{"Q", "R", "S", "T", "U", "V"} {
							f.SetColVisible(mainSheet, col, false)
						}
					}

					// Einmalig alle Formeln "un-sharen", damit sie beim Kopieren/Löschen
					// von Zeilen später nicht korrumpieren (spart ca. 40ms pro Pipeline-Job)
					_ = UnshareAllFormulas(f, mainSheet)

					// Alle eventuell versteckten Zeilen in der Vorlage wieder sichtbar machen
					for rNum := 1; rNum <= 1000; rNum++ {
						_ = f.SetRowVisible(mainSheet, rNum, true)
					}

					// Wenn wir schon dabei sind: Wir können hier auch direkt Unprotect aufrufen,
					// falls das Template geschützt war, aber der User keinen Schutz möchte.
					if !globalOpts.ProtectSheet {
						_ = f.UnprotectSheet(mainSheet)
					}
					if !globalOpts.ProtectWorkbook {
						_ = f.UnprotectWorkbook()
					}
				}

				var buf bytes.Buffer
				if err := f.Write(&buf); err == nil {
					data = buf.Bytes()
				}
			}

			templateCache.Store(path, data)
		}(filename)
	}

	wg.Wait()
	return preloadErr
}

func getTemplateBytes(path string) ([]byte, error) {
	if val, ok := templateCache.Load(path); ok {
		return val.([]byte), nil
	}
	data, err := templateFiles.ReadFile(path)
	if err != nil {
		return nil, err
	}
	templateCache.Store(path, data)
	return data, nil
}

// LanguageToTemplate ordnet die Dropdown-Werte der Sprache den entsprechenden Vorlagen-Dateien zu.
var LanguageToTemplate = map[string]string{
	"deutsch":   "de.xlsx",
	"english":   "en.xlsx",
	"français":  "fr.xlsx",
	"español":   "es.xlsx",
	"português": "po.xlsx",
}

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

	// Das erste Blatt ist immer der Hauptbericht (z.B. "Finanzbericht", "Financial Report", etc.)
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("keine Arbeitsblätter in der Datei gefunden")
	}
	sheetName := sheets[0]

	report := &ExcelReport{
		file:         f,
		sheet:        sheetName,
		CatStartRows: make(map[int]int),
		CatEndRows:   make(map[int]int),
		styleCache:   make(map[string]int),
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

		wasMode0 := (startRow == endRow)
		items := data.Categories[catID]

		// WICHTIGER FIX:
		// Wenn die Kategorie aktuell im Template im Modus 0 ist (startRow == endRow)
		// und es kommt keine oder nur EINE Kostenposition rein (len(items) <= 1),
		// dann bleiben wir in Modus 0 und wechseln nicht grundlos in Modus N.
		var emptyCount int
		if wasMode0 && len(items) <= 1 {
			emptyCount = 0

			// Falls es genau eine Unterposition gab, retten wir ihr Budget in den
			// Header der Kategorie, damit es nicht verloren geht.
			if len(items) == 1 {
				currentHeader := data.HeaderBudgets[catID]
				var currentVal float64
				switch v := currentHeader.(type) {
				case float64:
					currentVal = v
				case int:
					currentVal = float64(v)
				}
				if currentVal == 0 {
					if data.HeaderBudgets == nil {
						data.HeaderBudgets = make(map[int]interface{})
					}
					data.HeaderBudgets[catID] = items[0].Budget
				}
			}
			// Liste explizit leeren, damit targetPos = 0 bleibt (Modus 0 Definition)
			items = []CostItem{}
		} else {
			emptyCount = data.EmptyRows.Global
			if override, ok := data.EmptyRows.CategoryOverrides[catID]; ok {
				emptyCount = override
			}
			// Sicherstellen, dass niemals automatisch in den Modus 0 gewechselt wird,
			// indem immer mindestens 2 leere Puffer-Zeilen am Ende verbleiben.
			if emptyCount < 2 {
				emptyCount = 2
			}
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

		if err := r.processCategory(catID, items, targetPos, data); err != nil {
			return fmt.Errorf("fehler in Kategorie %d: %w", catID, err)
		}
	}
	return nil
}

func (r *ExcelReport) processCategory(catID int, items []CostItem, targetPos int, data ReportData) error {
	startRow := r.CatStartRows[catID]
	endRow := r.CatEndRows[catID]
	f := r.file
	s := r.sheet

	wasMode0 := (startRow == endRow)
	isMode0 := (targetPos == 0)

	// 1. FAST PATH: Kein Modus-Wechsel und Ziel ist Modus 0.
	if wasMode0 && isMode0 {
		// Update den Budget-Wert für die Hauptkategorie
		if budget, ok := data.HeaderBudgets[catID]; ok {
			r.setIfObj(fmt.Sprintf("D%d", startRow), budget)
		} else {
			r.setIfObj(fmt.Sprintf("D%d", startRow), 0)
		}
		return nil
	}

	// 2. MODUS N -> MODUS 0 Wechsel
	if isMode0 {
		if endRow > startRow {
			for rIdx := endRow; rIdx > startRow; rIdx-- {
				if err := f.RemoveRow(s, rIdx); err != nil {
					return err
				}
				r.shiftRows(-1, rIdx-1)
			}
		}
		// Styling für Modus 0 anwenden (Borders und Bold beibehalten!)
		styleCWhite := r.getCachedStyle(fmt.Sprintf("C%d", startRow), ColorWhite, false)
		styleDF2 := r.getCachedStyle(fmt.Sprintf("D%d", startRow), ColorLightGray, false)
		styleEFFF := r.getCachedStyle(fmt.Sprintf("E%d", startRow), ColorLightYellow, false)
		styleFFFF := r.getCachedStyle(fmt.Sprintf("F%d", startRow), ColorLightYellow, false)
		styleHFFF := r.getCachedStyle(fmt.Sprintf("H%d", startRow), ColorLightYellow, false)

		f.SetCellStyle(s, fmt.Sprintf("C%d", startRow), fmt.Sprintf("C%d", startRow), styleCWhite)
		f.SetCellStyle(s, fmt.Sprintf("D%d", startRow), fmt.Sprintf("D%d", startRow), styleDF2)
		f.SetCellStyle(s, fmt.Sprintf("E%d", startRow), fmt.Sprintf("E%d", startRow), styleEFFF)
		f.SetCellStyle(s, fmt.Sprintf("F%d", startRow), fmt.Sprintf("F%d", startRow), styleFFFF)
		f.SetCellStyle(s, fmt.Sprintf("H%d", startRow), fmt.Sprintf("H%d", startRow), styleHFFF)

		f.SetCellValue(s, fmt.Sprintf("G%d", startRow), 0)
		f.SetCellFormula(s, fmt.Sprintf("G%d", startRow), fmt.Sprintf("IFERROR(F%d/D%d,0)", startRow, startRow))

		formulaB := r.generateCategoryFormula(startRow, catID)
		f.SetCellDefault(s, fmt.Sprintf("B%d", startRow), "")
		setGeneralFormat(f, s, fmt.Sprintf("B%d", startRow))
		f.SetCellFormula(s, fmt.Sprintf("B%d", startRow), formulaB)

		// Wert eintragen
		if budget, ok := data.HeaderBudgets[catID]; ok {
			r.setIfObj(fmt.Sprintf("D%d", startRow), budget)
		} else {
			r.setIfObj(fmt.Sprintf("D%d", startRow), 0)
		}

		r.CatEndRows[catID] = startRow // Subtotal existiert nicht mehr, wir referenzieren den Header
		return nil
	}

	// 3. Ziel ist MODUS N (Zielzeilen > 0)
	modeChanged := false
	rowsChanged := false

	if wasMode0 {
		// Wechsel von Modus 0 auf Modus N
		modeChanged = true
		rowsChanged = true

		for colName := 'D'; colName <= 'H'; colName++ {
			cell := fmt.Sprintf("%c%d", colName, startRow)
			whiteStyle := r.getCachedStyle(cell, ColorWhite, false)
			f.SetCellStyle(s, cell, cell, whiteStyle)
		}

		f.SetCellValue(s, fmt.Sprintf("G%d", startRow), "") // Formel entfernen

		for i := 0; i < targetPos+1; i++ {
			if err := f.DuplicateRow(s, startRow); err != nil {
				return err
			}
			r.shiftRows(1, startRow)
		}

		subtotalRow := startRow + targetPos + 1

		formulaB := r.generateCategoryFormula(startRow, catID)
		f.SetCellDefault(s, fmt.Sprintf("B%d", startRow), "")
		setGeneralFormat(f, s, fmt.Sprintf("B%d", startRow))
		f.SetCellFormula(s, fmt.Sprintf("B%d", startRow), formulaB)

		for rowIdx := startRow + 1; rowIdx <= startRow+targetPos; rowIdx++ {
			for colName := 'C'; colName <= 'H'; colName++ {
				cell := fmt.Sprintf("%c%d", colName, rowIdx)
				f.SetCellValue(s, cell, "")
			}
			f.SetCellDefault(s, fmt.Sprintf("B%d", rowIdx), "")
			f.SetCellFormula(s, fmt.Sprintf("B%d", rowIdx), formulaB)
			setGeneralFormat(f, s, fmt.Sprintf("B%d", rowIdx))
		}
		endRow = subtotalRow
		r.CatEndRows[catID] = endRow
	} else {
		// Bereits in Modus N
		currentPos := endRow - startRow - 1
		physicalEndRow := endRow
		if currentPos != targetPos {
			rowsChanged = true
		}

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
	}

	// Wenn sich das Layout nicht geändert hat (modeChanged=false && rowsChanged=false),
	// ist der komplette Style inkl. Formeln aus dem Template bereits perfekt.
	// Wir können styleCategory drastisch abkürzen!
	return r.styleCategory(startRow, endRow, catID, modeChanged, rowsChanged)
}

func (r *ExcelReport) getCachedStyle(cell, hexColor string, setBoldFalse bool) int {
	f := r.file
	s := r.sheet

	// 1. Ursprüngliche Style-ID der Zelle auslesen
	baseStyleID, err := f.GetCellStyle(s, cell)
	if err != nil {
		return 0
	}

	// 2. Cache-Key bilden
	cacheKey := fmt.Sprintf("%d_%s_%v", baseStyleID, hexColor, setBoldFalse)

	// 3. Im Cache suchen
	if cachedID, ok := r.styleCache[cacheKey]; ok {
		return cachedID
	}

	// 4. Style generieren und cachen, falls nicht vorhanden
	newID := modStyleHelperEx(f, s, cell, hexColor, setBoldFalse, "")
	r.styleCache[cacheKey] = newID
	return newID
}

func (r *ExcelReport) styleCategory(startRow, endRow, catID int, modeChanged, rowsChanged bool) error {
	f := r.file
	s := r.sheet

	// Nur wenn wir von Modus 0 auf Modus N gewechselt sind, müssen wir die neu erstellten Zeilen
	// komplett formatieren. Wenn wir Modus N beibehalten haben, hat `DuplicateRow` die Formate
	// automatisch und perfekt für uns kopiert!
	if modeChanged {
		// Styling für die gesamte Kategorie sicherstellen
		for rowIdx := startRow + 1; rowIdx < endRow; rowIdx++ {
			styleB := r.getCachedStyle(fmt.Sprintf("B%d", rowIdx), ColorWhite, true)
			f.SetCellStyle(s, fmt.Sprintf("B%d", rowIdx), fmt.Sprintf("B%d", rowIdx), styleB)

			styleC := r.getCachedStyle(fmt.Sprintf("C%d", rowIdx), ColorLightYellow, true)
			f.SetCellStyle(s, fmt.Sprintf("C%d", rowIdx), fmt.Sprintf("C%d", rowIdx), styleC)

			styleD := r.getCachedStyle(fmt.Sprintf("D%d", rowIdx), ColorLightGray, true)
			f.SetCellStyle(s, fmt.Sprintf("D%d", rowIdx), fmt.Sprintf("D%d", rowIdx), styleD)

			styleEF := r.getCachedStyle(fmt.Sprintf("E%d", rowIdx), ColorLightYellow, true)
			f.SetCellStyle(s, fmt.Sprintf("E%d", rowIdx), fmt.Sprintf("E%d", rowIdx), styleEF)
			f.SetCellStyle(s, fmt.Sprintf("F%d", rowIdx), fmt.Sprintf("F%d", rowIdx), styleEF)

			styleH := r.getCachedStyle(fmt.Sprintf("H%d", rowIdx), ColorLightYellow, true)
			f.SetCellStyle(s, fmt.Sprintf("H%d", rowIdx), fmt.Sprintf("H%d", rowIdx), styleH)

			styleG := r.getCachedStyle(fmt.Sprintf("G%d", rowIdx), ColorWhite, true)
			f.SetCellStyle(s, fmt.Sprintf("G%d", rowIdx), fmt.Sprintf("G%d", rowIdx), styleG)
		}

		// Zwischensummen-Zeile formatieren
		f.UnmergeCell(s, fmt.Sprintf("B%d", endRow), fmt.Sprintf("C%d", endRow))

		f.SetCellDefault(s, fmt.Sprintf("B%d", endRow), "")
		f.SetCellDefault(s, fmt.Sprintf("C%d", endRow), "")
		f.SetCellFormula(s, fmt.Sprintf("B%d", endRow), "")
		f.SetCellFormula(s, fmt.Sprintf("C%d", endRow), "")
		f.SetCellValue(s, fmt.Sprintf("C%d", endRow), "")

		// Dynamische Formel für Zwischensumme
		vlookupIdx := VLookupBaseIdx + (catID * VLookupMult)
		f.SetCellDefault(s, fmt.Sprintf("B%d", endRow), "")
		if err := f.SetCellFormula(s, fmt.Sprintf("B%d", endRow), fmt.Sprintf(`IF($E$2="","",VLOOKUP($E$2,Sprachversionen!$B:$BN,%d,FALSE))`, vlookupIdx)); err != nil {
			return err
		}

		f.MergeCell(s, fmt.Sprintf("B%d", endRow), fmt.Sprintf("C%d", endRow))

		for colName := 'B'; colName <= 'H'; colName++ {
			cell := fmt.Sprintf("%c%d", colName, endRow)
			fontColor := ""
			if colName >= 'D' && colName <= 'H' {
				fontColor = ColorBlack
			}

			// Den ModHelper können wir hier nicht direkt mit getCachedStyle abdecken, weil fontColor noch dazu kommt.
			grayStyle := modStyleHelperEx(f, s, cell, ColorDarkGray, false, fontColor)
			f.SetCellStyle(s, cell, cell, grayStyle)
		}
	}

	// SUM Formeln müssen wir aktualisieren, wenn sich die ANZAHL der Zeilen geändert hat
	// oder wenn wir neu in Modus N gekommen sind.
	if modeChanged || rowsChanged {
		f.SetCellValue(s, fmt.Sprintf("D%d", endRow), 0)
		if err := f.SetCellFormula(s, fmt.Sprintf("D%d", endRow), fmt.Sprintf("SUM(D%d:D%d)", startRow+1, endRow-1)); err != nil {
			return err
		}
		f.SetCellValue(s, fmt.Sprintf("E%d", endRow), 0)
		if err := f.SetCellFormula(s, fmt.Sprintf("E%d", endRow), fmt.Sprintf("SUM(E%d:E%d)", startRow+1, endRow-1)); err != nil {
			return err
		}
		f.SetCellValue(s, fmt.Sprintf("F%d", endRow), 0)
		if err := f.SetCellFormula(s, fmt.Sprintf("F%d", endRow), fmt.Sprintf("SUM(F%d:F%d)", startRow+1, endRow-1)); err != nil {
			return err
		}
		f.SetCellValue(s, fmt.Sprintf("G%d", endRow), 0)
		if err := f.SetCellFormula(s, fmt.Sprintf("G%d", endRow), fmt.Sprintf("IFERROR(F%d/D%d,0)", endRow, endRow)); err != nil {
			return err
		}
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
