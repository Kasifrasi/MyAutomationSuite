package report

import (
	"bytes"
	"fmt"
	"strconv"
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

// parseAmount ist eine Hilfsfunktion zum Parsen von Geldbeträgen aus Strings
func parseAmount(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	// Wenn es Tausender-Punkte und ein Komma gibt (deutsche Formatierung)
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
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
		Eigenleistung: FundingRecord{Budget: parseAmount(scanned.Eigenleistung)},
		Drittmittel:   FundingRecord{Budget: parseAmount(scanned.Drittmittel)},
		KMWMittel:     FundingRecord{Budget: parseAmount(scanned.KmwMittel)},
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
			budget := parseAmount(pos.CostCol1)
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
			Budget: parseAmount(pos.CostCol1),
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
