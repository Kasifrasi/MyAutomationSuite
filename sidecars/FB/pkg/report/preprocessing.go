package report

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

var templateCache sync.Map // map[string][]byte

// LanguageToTemplate ordnet die Dropdown-Werte der Sprache den entsprechenden Vorlagen-Dateien zu.
var LanguageToTemplate = map[string]string{
	"deutsch":   "de.xlsx",
	"english":   "en.xlsx",
	"français":  "fr.xlsx",
	"español":   "es.xlsx",
	"português": "po.xlsx",
}

// PreloadAllTemplates lädt alle Vorlagen aus dem RAM (via //go:embed) und wendet globale Einstellungen an.
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

			// Immer zuerst alle Spalten, Zeilen und Gruppierungen blitzschnell per Regex-Postprocess entfernen/einblenden (Clean Slate)
			if cleaned, err := ProcessExcelVisibility(data, true, true, true); err == nil {
				data = cleaned
			}

			// Einmalig die globalen Layout-Einstellungen für diese Vorlage anwenden
			f, err := excelize.OpenReader(bytes.NewReader(data))
			if err == nil {
				sheets := f.GetSheetList()
				if len(sheets) > 0 {
					mainSheet := sheets[0]

					// Spalten Q-V verstecken (falls gewünscht)
					if globalOpts.HideColumns {
						for _, col := range []string{"Q", "R", "S", "T", "U", "V"} {
							f.SetColVisible(mainSheet, col, false)
						}
					}

					// Einmalig alle Formeln "un-sharen", damit sie beim Kopieren/Löschen
					// von Zeilen später nicht korrumpieren (spart ca. 40ms pro Pipeline-Job)
					_ = UnshareAllFormulas(f, mainSheet)

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

// MapScannedToReportData übersetzt das Rust Scanner-Modell in das Go Report-Modell
func MapScannedToReportData(scanned *ScannedBudgetData) ReportData {
	sprache := strings.ToLower(scanned.Language)

	// Falls die Sprache nicht direkt passt, mappen wir sie grob
	switch {
	case strings.Contains(sprache, "de"):
		sprache = "deutsch"
	case strings.Contains(sprache, "en"):
		sprache = "english"
	case strings.Contains(sprache, "fr"):
		sprache = "français"
	case strings.Contains(sprache, "es"):
		sprache = "español"
	case strings.Contains(sprache, "pt") || strings.Contains(sprache, "po"):
		sprache = "português"
	default:
		sprache = "deutsch" // Fallback
	}

	data := ReportData{
		Sprache:       sprache,
		Lokalwaehrung: scanned.LocalCurrency,
		Projektnummer: scanned.ProjectNumber,
		Projekttitel:  scanned.ProjectTitle,
		EmptyRows: EmptyRowsConfig{
			Global:            3, // Standardmäßig 3 leere Zeilen pro Kategorie beibehalten
			CategoryOverrides: make(map[int]int),
		},
		Eigenleistung: FundingRecord{Budget: amount(scanned.Financing.Eigenmittel.LC)},
		Drittmittel:   FundingRecord{Budget: amount(scanned.Financing.Drittmittel.LC)},
		KMWMittel:     FundingRecord{Budget: amount(scanned.Financing.KMWMittel.LC)},
		Categories:    make(map[int][]CostItem),
		HeaderBudgets: make(map[int]interface{}),
	}

	// 1. Zuerst Hauptkategorien-Budgets ("1.", "6." etc.) aufsammeln
	for _, pos := range scanned.Positions {
		if len(pos.Number) == 0 {
			continue
		}
		catID := int(pos.Number[0] - '0')
		if catID < 1 || catID > 8 {
			continue
		}

		if strings.HasSuffix(pos.Number, ".") {
			// Es ist eine Hauptkategorie! Wir speichern uns ihren Wert.
			budget := amount(pos.LC)
			if budget >= 0 {
				data.HeaderBudgets[catID] = budget
			}
		}
	}

	// 2. Jetzt die echten Unterpositionen zuweisen
	for _, pos := range scanned.Positions {
		if len(pos.Number) == 0 {
			continue
		}

		catID := int(pos.Number[0] - '0')
		if catID < 1 || catID > 8 {
			continue
		}

		// Hauptkategorie-Zeilen überspringen (die haben wir oben als Header-Budgets verarbeitet)
		if strings.HasSuffix(pos.Number, ".") {
			continue
		}

		// ACHTUNG: pos.Number nicht mehr dem Namen voranstellen!
		item := CostItem{
			Name:   pos.Label,
			Budget: amount(pos.LC),
		}

		data.Categories[catID] = append(data.Categories[catID], item)
	}

	// 2.5 Wir trimmen NUR abschließende Items mit Budget = 0 und leerem Namen.
	// Führende leere Items MÜSSEN erhalten bleiben, damit die Nummerierung (z.B. 1.1, 1.2, etc.)
	// nicht verschoben wird. Nur was am Ende "leer" dranhängt, wird weggeschnitten.
	for catID := 1; catID <= 8; catID++ {
		lastValid := -1

		// Definiert, was eine "gültige" Kostenposition ausmacht:
		// Entweder das Budget ist nicht 0 ODER der Name der Position ist nicht leer.
		isValidItem := func(item CostItem) bool {
			// Typischerweise ist Name ein string. Falls nicht, prüfen wir das hier.
			nameStr, ok := item.Name.(string)
			if !ok {
				nameStr = ""
			}

			// Budget ist ein interface{}, wir müssen den Typ sicher asserten,
			// da float64(0) != int(0) sonst in Go true ergibt.
			var budget float64
			switch v := item.Budget.(type) {
			case float64:
				budget = v
			case int:
				budget = float64(v)
			}

			return budget != 0 || strings.TrimSpace(nameStr) != ""
		}

		for i, item := range data.Categories[catID] {
			if isValidItem(item) {
				lastValid = i
			}
		}

		if lastValid != -1 {
			// Wir fangen immer bei 0 an und schneiden nur hinten ab!
			data.Categories[catID] = data.Categories[catID][0 : lastValid+1]
		} else {
			// Alles war ungültig (kein Name, kein Budget) -> Kategorie komplett leeren
			data.Categories[catID] = []CostItem{}
		}
	}

	return data
}

// regexOutline sucht nach outlineLevel="X" Attributen im XML
var regexOutline = regexp.MustCompile(` outlineLevel="\d+"`)

// ProcessExcelVisibility entfernt versteckte Zeilen/Spalten und optional Gruppierungen aus den Worksheets per Regex
func ProcessExcelVisibility(excelData []byte, unhideCols bool, unhideRows bool, removeGroupings bool) ([]byte, error) {
	if !unhideCols && !unhideRows && !removeGroupings {
		return excelData, nil
	}

	reader, err := zip.NewReader(bytes.NewReader(excelData), int64(len(excelData)))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		if regexp.MustCompile(`^xl/worksheets/.*\.xml$`).MatchString(file.Name) {
			// Wir können <row ... hidden="1"> und <col ... hidden="1"> gezielt bereinigen.
			// Der Einfachheit halber entfernen wir hidden="..." Attribute global in dem Worksheet,
			// sofern Zeilen/Spalten eingeblendet werden sollen. Da Excelize die Q:V Ausblendung
			// erst nach dem Preload (wo wir es theoretisch tun könnten) setzt,
			// müssen wir hier vorsichtig sein, falls wir nur Rows unhiden wollen.
			// Ein simpler Regex-Replace für <row> und <col> Tags:
			if unhideRows {
				// Ersetzt hidden="..." nur innerhalb von <row ...>
				content = regexp.MustCompile(`(<row[^>]*?)\shidden="(?:1|true)"([^>]*>)`).ReplaceAll(content, []byte("$1$2"))
			}
			if unhideCols {
				// Ersetzt hidden="..." nur innerhalb von <col ...>
				content = regexp.MustCompile(`(<col[^>]*?)\shidden="(?:1|true)"([^>]*>)`).ReplaceAll(content, []byte("$1$2"))
			}
			if removeGroupings {
				// Entfernt outlineLevel Attribute komplett aus den Zeilen/Spalten (Gruppierungen entfernen)
				content = regexOutline.ReplaceAll(content, []byte(""))
			}
		}

		fWriter, err := writer.Create(file.Name)
		if err != nil {
			return nil, err
		}
		if _, err := fWriter.Write(content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
