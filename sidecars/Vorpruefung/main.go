package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

type StyleOptions struct {
	Bold         bool
	Italic       bool
	Size         float64
	FontColor    string
	FillColor    string
	HAlign       string
	VAlign       string
	NumFormat    string
	BorderTop    int // 0: none, 1: thin, 2: medium
	BorderBottom int
	BorderLeft   int
	BorderRight  int
	BorderColor  string
	WrapText     bool
	Strike       bool
}

type Generator struct {
	file           *excelize.File
	styleCache     map[string]int
	condStyleCache map[string]int

	// Ranges for VSTACK in "Daten"
	rangesAusgaben   []string
	rangesEinnahmen1 []string
	rangesEinnahmen2 []string
	rangesMA         []string

	// Zellen mit dynamischen Array-Formeln (Spill); werden nach dem Speichern
	// mit den Dynamic-Array-Metadaten versehen (siehe dynarray.go).
	dynArrayCells []dynArrayCell

	// Sheet-qualifizierte Absolutadresse der FB-Auswahl-Periodennummer auf dem
	// Auswertungsblatt (z. B. "'V. AUSWERTUNG'!$N$120"). Wird von daten.go genutzt,
	// um die Mittelanforderungs-Auswahlliste auf "Periode FB+1" zu filtern.
	evalFBSelNumAddr string
}

type dynArrayCell struct {
	sheet string
	cell  string
}

func colLetter(col int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return name
}

func cellName(col, row int) string {
	return fmt.Sprintf("%s%d", colLetter(col), row)
}

func absName(col, row int) string {
	return fmt.Sprintf("$%s$%d", colLetter(col), row)
}

func (g *Generator) getOrCreateStyle(opts StyleOptions) (int, error) {
	key := fmt.Sprintf("%t-%t-%f-%s-%s-%s-%s-%s-%d-%d-%d-%d-%s-%t-%t",
		opts.Bold, opts.Italic, opts.Size, opts.FontColor, opts.FillColor,
		opts.HAlign, opts.VAlign, opts.NumFormat,
		opts.BorderTop, opts.BorderBottom, opts.BorderLeft, opts.BorderRight,
		opts.BorderColor, opts.WrapText, opts.Strike)

	if id, exists := g.styleCache[key]; exists {
		return id, nil
	}

	style := &excelize.Style{}

	// Font
	font := &excelize.Font{
		Family: "Segoe UI",
		Size:   opts.Size,
		Bold:   opts.Bold,
		Italic: opts.Italic,
		Strike: opts.Strike,
	}
	if opts.FontColor != "" {
		font.Color = strings.TrimPrefix(opts.FontColor, "#")
	} else {
		font.Color = "000000"
	}
	style.Font = font

	// Fill
	if opts.FillColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Color:   []string{strings.TrimPrefix(opts.FillColor, "#")},
			Pattern: 1,
		}
	}

	// Alignment
	alignment := &excelize.Alignment{
		WrapText: opts.WrapText,
	}
	if opts.HAlign != "" {
		alignment.Horizontal = opts.HAlign
	}
	if opts.VAlign != "" {
		alignment.Vertical = opts.VAlign
	}
	style.Alignment = alignment

	// Number Format
	if opts.NumFormat != "" {
		style.CustomNumFmt = &opts.NumFormat
	}

	// Borders
	var borders []excelize.Border
	borderColor := "808080"
	if opts.BorderColor != "" {
		borderColor = strings.TrimPrefix(opts.BorderColor, "#")
	}

	if opts.BorderTop > 0 {
		borders = append(borders, excelize.Border{Type: "top", Color: borderColor, Style: opts.BorderTop})
	}
	if opts.BorderBottom > 0 {
		borders = append(borders, excelize.Border{Type: "bottom", Color: borderColor, Style: opts.BorderBottom})
	}
	if opts.BorderLeft > 0 {
		borders = append(borders, excelize.Border{Type: "left", Color: borderColor, Style: opts.BorderLeft})
	}
	if opts.BorderRight > 0 {
		borders = append(borders, excelize.Border{Type: "right", Color: borderColor, Style: opts.BorderRight})
	}
	if len(borders) > 0 {
		style.Border = borders
	}

	id, err := g.file.NewStyle(style)
	if err != nil {
		return 0, err
	}
	g.styleCache[key] = id
	return id, nil
}

func (g *Generator) getOrCreateConditionalStyle(opts StyleOptions) (int, error) {
	key := fmt.Sprintf("cond-%t-%t-%f-%s-%s-%s-%s-%s-%d-%d-%d-%d-%s-%t-%t",
		opts.Bold, opts.Italic, opts.Size, opts.FontColor, opts.FillColor,
		opts.HAlign, opts.VAlign, opts.NumFormat,
		opts.BorderTop, opts.BorderBottom, opts.BorderLeft, opts.BorderRight,
		opts.BorderColor, opts.WrapText, opts.Strike)

	if id, exists := g.condStyleCache[key]; exists {
		return id, nil
	}

	style := &excelize.Style{}

	// Font
	font := &excelize.Font{
		Family: "Segoe UI",
		Size:   opts.Size,
		Bold:   opts.Bold,
		Italic: opts.Italic,
		Strike: opts.Strike,
	}
	if opts.FontColor != "" {
		font.Color = strings.TrimPrefix(opts.FontColor, "#")
	} else {
		font.Color = "000000"
	}
	style.Font = font

	// Fill
	if opts.FillColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Color:   []string{strings.TrimPrefix(opts.FillColor, "#")},
			Pattern: 1,
		}
	}

	// Alignment
	alignment := &excelize.Alignment{
		WrapText: opts.WrapText,
	}
	if opts.HAlign != "" {
		alignment.Horizontal = opts.HAlign
	}
	if opts.VAlign != "" {
		alignment.Vertical = opts.VAlign
	}
	style.Alignment = alignment

	// Number Format
	if opts.NumFormat != "" {
		style.CustomNumFmt = &opts.NumFormat
	}

	// Borders
	var borders []excelize.Border
	borderColor := "808080"
	if opts.BorderColor != "" {
		borderColor = strings.TrimPrefix(opts.BorderColor, "#")
	}

	if opts.BorderTop > 0 {
		borders = append(borders, excelize.Border{Type: "top", Color: borderColor, Style: opts.BorderTop})
	}
	if opts.BorderBottom > 0 {
		borders = append(borders, excelize.Border{Type: "bottom", Color: borderColor, Style: opts.BorderBottom})
	}
	if opts.BorderLeft > 0 {
		borders = append(borders, excelize.Border{Type: "left", Color: borderColor, Style: opts.BorderLeft})
	}
	if opts.BorderRight > 0 {
		borders = append(borders, excelize.Border{Type: "right", Color: borderColor, Style: opts.BorderRight})
	}
	if len(borders) > 0 {
		style.Border = borders
	}

	id, err := g.file.NewConditionalStyle(style)
	if err != nil {
		return 0, err
	}
	g.condStyleCache[key] = id
	return id, nil
}

func (g *Generator) setStyle(sheet, hCell, vCell string, opts StyleOptions) error {
	id, err := g.getOrCreateStyle(opts)
	if err != nil {
		return err
	}
	return g.file.SetCellStyle(sheet, hCell, vCell, id)
}

func (g *Generator) setValue(sheet, cell string, val interface{}, opts StyleOptions) error {
	err := g.file.SetCellValue(sheet, cell, val)
	if err != nil {
		return err
	}
	return g.setStyle(sheet, cell, cell, opts)
}

func (g *Generator) setFormula(sheet, cell, formula string, opts StyleOptions) error {
	err := g.file.SetCellFormula(sheet, cell, formula)
	if err != nil {
		return err
	}
	return g.setStyle(sheet, cell, cell, opts)
}

// setDynArrayFormula schreibt eine dynamische Array-Formel (Spill) und merkt sich die
// Zelle für das nachträgliche Setzen der Dynamic-Array-Metadaten (siehe dynarray.go).
func (g *Generator) setDynArrayFormula(sheet, cell, formula string, opts StyleOptions) error {
	setDynArrayFormula(g.file, sheet, cell, formula)
	g.dynArrayCells = append(g.dynArrayCells, dynArrayCell{sheet: sheet, cell: cell})
	return g.setStyle(sheet, cell, cell, opts)
}

func (g *Generator) mergeCells(sheet, hCell, vCell string, val interface{}, opts StyleOptions) error {
	err := g.file.MergeCell(sheet, hCell, vCell)
	if err != nil {
		return err
	}
	err = g.file.SetCellValue(sheet, hCell, val)
	if err != nil {
		return err
	}
	return g.setStyle(sheet, hCell, vCell, opts)
}

func (g *Generator) setColWidth(sheet string, col int, width float64) {
	_ = g.file.SetColWidth(sheet, colLetter(col), colLetter(col), width)
}

func (g *Generator) upsertNamedRange(name string, col, row int) {
	_ = g.file.DeleteDefinedName(&excelize.DefinedName{Name: name})
	_ = g.file.SetDefinedName(&excelize.DefinedName{
		Name:     name,
		RefersTo: fmt.Sprintf("'%s'!%s", BG_SHEET_NAME, absName(col, row)),
	})
}

func (g *Generator) upsertNamedFormula(name string, refersTo string) {
	_ = g.file.DeleteDefinedName(&excelize.DefinedName{Name: name})
	_ = g.file.SetDefinedName(&excelize.DefinedName{
		Name:     name,
		RefersTo: refersTo,
	})
}

func (g *Generator) drawSectionHeader(sheet string, row int, title string) error {
	opts := StyleOptions{
		Bold:         true,
		Size:         11.0,
		FontColor:    BG_CLR_BLACK,
		FillColor:    BG_CLR_HEADER,
		HAlign:       "left",
		VAlign:       "center",
		BorderTop:    2,
		BorderBottom: 1,
		BorderColor:  BG_CLR_BORDER,
	}
	err := g.mergeCells(sheet, cellName(BG_COL_LABEL, row), cellName(BG_COL_EUR, row), title, opts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(sheet, row, 24.0)
	return nil
}

func (g *Generator) drawSubHeader(sheet string, row int, title string) error {
	opts := StyleOptions{
		Bold:         true,
		Size:         10.0,
		FontColor:    BG_CLR_BLACK,
		FillColor:    BG_CLR_SUBHEAD,
		HAlign:       "left",
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderColor:  BG_CLR_BORDER,
	}
	err := g.mergeCells(sheet, cellName(BG_COL_LABEL, row), cellName(BG_COL_EUR, row), title, opts)
	if err != nil {
		return err
	}
	_ = g.file.SetRowHeight(sheet, row, 20.0)
	return nil
}

func (g *Generator) drawYearRow(sheet string, row int, label string, labelName, lwName, eurName string) error {
	lblOpts := StyleOptions{
		Bold:   false,
		Size:   10.0,
		HAlign: "left",
		VAlign: "center",
	}
	err := g.setValue(sheet, cellName(BG_COL_LABEL, row), label, lblOpts)
	if err != nil {
		return err
	}

	inputLCOpts := StyleOptions{
		FillColor:    BG_CLR_INPUT,
		HAlign:       "right",
		VAlign:       "center",
		NumFormat:    BG_FMT_LC,
		BorderLeft:   1,
		BorderRight:  1,
		BorderTop:    1,
		BorderBottom: 1,
		BorderColor:  BG_CLR_GRID,
	}
	for c := BG_COL_LC; c <= BG_COL_Y3; c++ {
		err = g.setValue(sheet, cellName(c, row), "", inputLCOpts)
		if err != nil {
			return err
		}
	}

	inputEuroOpts := inputLCOpts
	inputEuroOpts.NumFormat = BG_FMT_EUR
	err = g.setValue(sheet, cellName(BG_COL_EUR, row), "", inputEuroOpts)
	if err != nil {
		return err
	}

	if lwName != "" {
		g.upsertNamedRange(lwName, BG_COL_LC, row)
	}
	if eurName != "" {
		g.upsertNamedRange(eurName, BG_COL_EUR, row)
	}

	_ = g.file.SetRowHeight(sheet, row, 22.0)
	return nil
}

func (g *Generator) styleHeader(sheet string, r1, c1, r2, c2 int) error {
	opts := StyleOptions{
		Bold:         true,
		Size:         9.0,
		FontColor:    BG_CLR_FONT,
		FillColor:    BG_CLR_HEADER,
		HAlign:       "center",
		VAlign:       "center",
		BorderBottom: 2,
		BorderColor:  BG_CLR_BORDER,
	}
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			err := g.setStyle(sheet, cellName(c, r), cellName(c, r), opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) styleOuterBorder(sheet string, r1, c1, r2, c2 int, weight int, color string) error {
	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if r == r1 || r == r2 || c == c1 || c == c2 {
				cell := cellName(c, r)
				styleID, err := g.file.GetCellStyle(sheet, cell)
				if err != nil {
					continue
				}
				style, err := g.file.GetStyle(styleID)
				if err != nil || style == nil {
					style = &excelize.Style{}
				}

				var borders []excelize.Border
				if style.Border != nil {
					borders = style.Border
				}

				upsertBorder := func(borderType string, styleVal int) {
					found := false
					for i, b := range borders {
						if b.Type == borderType {
							borders[i].Style = styleVal
							borders[i].Color = strings.TrimPrefix(color, "#")
							found = true
							break
						}
					}
					if !found {
						borders = append(borders, excelize.Border{
							Type:  borderType,
							Color: strings.TrimPrefix(color, "#"),
							Style: styleVal,
						})
					}
				}

				if r == r1 {
					upsertBorder("top", weight)
				}
				if r == r2 {
					upsertBorder("bottom", weight)
				}
				if c == c1 {
					upsertBorder("left", weight)
				}
				if c == c2 {
					upsertBorder("right", weight)
				}

				style.Border = borders
				newStyleID, err := g.file.NewStyle(style)
				if err == nil {
					_ = g.file.SetCellStyle(sheet, cell, cell, newStyleID)
				}
			}
		}
	}
	return nil
}

// reapplyRightBorder hebt die rechte Kante einer einzelnen Zelle auf das angegebene
// Gewicht/Farbe an, ohne die übrigen Stil-Eigenschaften (Füllung, Format, Formel) zu
// verlieren. Nötig, wenn eine Zelle nach styleOuterBorder erneut formatiert wurde
// (z. B. nachgelagerte Abzugsformeln) und dadurch die kräftige Box-Kante einbüßt.
func (g *Generator) reapplyRightBorder(sheet, cell string, weight int, color string) {
	styleID, err := g.file.GetCellStyle(sheet, cell)
	if err != nil {
		return
	}
	style, err := g.file.GetStyle(styleID)
	if err != nil || style == nil {
		return
	}
	color = strings.TrimPrefix(color, "#")
	found := false
	for i, b := range style.Border {
		if b.Type == "right" {
			style.Border[i].Style = weight
			style.Border[i].Color = color
			found = true
			break
		}
	}
	if !found {
		style.Border = append(style.Border, excelize.Border{Type: "right", Color: color, Style: weight})
	}
	if newID, err := g.file.NewStyle(style); err == nil {
		_ = g.file.SetCellStyle(sheet, cell, cell, newID)
	}
}

func (g *Generator) styleTotalRow(sheet string, row int) error {
	opts := StyleOptions{
		Bold:         true,
		Size:         10.0,
		FontColor:    BG_CLR_BLACK,
		FillColor:    BG_CLR_SUBHEAD,
		VAlign:       "center",
		BorderTop:    1,
		BorderBottom: 1,
		BorderLeft:   1,
		BorderRight:  1,
		BorderColor:  BG_CLR_BORDER,
	}
	for c := BG_COL_LABEL; c <= BG_COL_EUR; c++ {
		err := g.setStyle(sheet, cellName(c, row), cellName(c, row), opts)
		if err != nil {
			return err
		}
	}
	_ = g.file.SetRowHeight(sheet, row, 20.0)
	return nil
}

func (g *Generator) addConditionalFormat(sheet, cell, formula string, opts StyleOptions) error {
	styleID, err := g.getOrCreateConditionalStyle(opts)
	if err != nil {
		return err
	}
	// Führendes '=' entfernen, um doppelte Gleichheitszeichen im generierten XML zu vermeiden
	formula = strings.TrimPrefix(formula, "=")
	return g.file.SetConditionalFormat(sheet, cell, []excelize.ConditionalFormatOptions{
		{
			Type:     "formula",
			Criteria: formula,
			Format:   &styleID,
		},
	})
}

func main() {
	var outputPath string
	flag.StringVar(&outputPath, "o", "vorpruefung_output.xlsx", "output file path")
	flag.Parse()

	f := excelize.NewFile()

	g := &Generator{
		file:           f,
		styleCache:     make(map[string]int),
		condStyleCache: make(map[string]int),
	}

	// 1. Erstelle das Dashboard-Blatt
	err := g.CreateDashboardSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Dashboard-Blatts: %v", err)
	}

	// 2. Erstelle das Budget-Blatt
	err = g.CreateBudgetSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Budget-Blatts: %v", err)
	}

	// 3. Erstelle das KMW-Mittel-Blatt
	err = g.CreateKMWMittelSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des KMW-Mittel-Blatts: %v", err)
	}

	// 4. Erstelle das Finanzberichte-Blatt
	err = g.CreateFinanzberichteSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Finanzberichte-Blatts: %v", err)
	}

	// 5. Erstelle das Mittelanforderung-Blatt
	err = g.CreateMittelanforderungSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Mittelanforderung-Blatts: %v", err)
	}

	// 6. Erstelle das Auswertungs-Blatt (Vergleich Budget vs. MA/Finanzberichte)
	err = g.CreateAuswertungSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Auswertungs-Blatts: %v", err)
	}

	// 7. Erstelle das Daten-Blatt für Listen und VSTACKS
	err = g.CreateDatenSheet()
	if err != nil {
		log.Fatalf("fehler beim Erstellen des Daten-Blatts: %v", err)
	}

	// Hier können in Zukunft weitere Blätter hinzugefügt werden:
	// err = g.CreateMittelSheet()
	// err = g.CreateFinanzberichtSheet()

	// 2. Lösche das standardmäßig erstellte "Sheet1", damit nur unsere Blätter übrig bleiben
	_ = f.DeleteSheet("Sheet1")

	// Excel beim Öffnen vollständig neu rechnen lassen, damit die dynamischen
	// Array-Formeln (VSTACK/FILTER) sofort korrekt spillen.
	fullCalc := true
	if err := f.SetCalcProps(&excelize.CalcPropsOptions{FullCalcOnLoad: &fullCalc}); err != nil {
		log.Fatalf("fehler beim Setzen der Berechnungsoptionen: %v", err)
	}

	// 3. Speichere das gesamte Dokument
	err = f.SaveAs(outputPath)
	if err != nil {
		log.Fatalf("fehler beim Speichern des Dokuments: %v", err)
	}

	// 4. Dynamische Array-Formeln (VSTACK/FILTER) nachträglich als echte Spill-Formeln
	//    markieren, damit Excel beim Öffnen keinen "@"-Operator einfügt.
	if err := applyDynamicArrayMetadata(outputPath, g.dynArrayCells); err != nil {
		log.Fatalf("fehler beim Setzen der Dynamic-Array-Metadaten: %v", err)
	}

	fmt.Printf("Vorpruefung erfolgreich generiert: %s\n", outputPath)
}
