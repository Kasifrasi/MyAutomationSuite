package vorpruefung

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
	key := fmt.Sprintf("%t-%t-%f-%s-%s-%s-%s-%s-%d-%d-%d-%d-%d-%s-%t-%t-%t",
		opts.Bold, opts.Italic, opts.Size, opts.FontColor, opts.FillColor,
		opts.HAlign, opts.VAlign, opts.NumFormat, opts.NumFmtID,
		opts.BorderTop, opts.BorderBottom, opts.BorderLeft, opts.BorderRight,
		opts.BorderColor, opts.WrapText, opts.Strike, opts.Unlocked)

	if id, exists := g.styleCache[key]; exists {
		return id, nil
	}

	style := &excelize.Style{}

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

	if opts.FillColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Color:   []string{strings.TrimPrefix(opts.FillColor, "#")},
			Pattern: 1,
		}
	}

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

	if opts.NumFmtID > 0 {
		style.NumFmt = opts.NumFmtID
	} else if opts.NumFormat != "" {
		style.CustomNumFmt = &opts.NumFormat
	}

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

	if opts.Unlocked {
		style.Protection = &excelize.Protection{
			Locked: false,
		}
	}

	id, err := g.file.NewStyle(style)
	if err != nil {
		return 0, err
	}
	g.styleCache[key] = id
	return id, nil
}

func (g *Generator) getOrCreateConditionalStyle(opts StyleOptions) (int, error) {
	key := fmt.Sprintf("cond-%t-%t-%f-%s-%s-%s-%s-%s-%d-%d-%d-%d-%d-%s-%t-%t-%t",
		opts.Bold, opts.Italic, opts.Size, opts.FontColor, opts.FillColor,
		opts.HAlign, opts.VAlign, opts.NumFormat, opts.NumFmtID,
		opts.BorderTop, opts.BorderBottom, opts.BorderLeft, opts.BorderRight,
		opts.BorderColor, opts.WrapText, opts.Strike, opts.Unlocked)

	if id, exists := g.condStyleCache[key]; exists {
		return id, nil
	}

	style := &excelize.Style{}

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

	if opts.FillColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Color:   []string{strings.TrimPrefix(opts.FillColor, "#")},
			Pattern: 1,
		}
	}

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

	if opts.NumFmtID > 0 {
		style.NumFmt = opts.NumFmtID
	} else if opts.NumFormat != "" {
		style.CustomNumFmt = &opts.NumFormat
	}

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
	colorClean := strings.TrimPrefix(color, "#")

	for r := r1; r <= r2; r++ {
		for c := c1; c <= c2; c++ {
			if r == r1 || r == r2 || c == c1 || c == c2 {
				cell := cellName(c, r)
				styleID, err := g.file.GetCellStyle(sheet, cell)
				if err != nil {
					continue
				}

				isTop := r == r1
				isBottom := r == r2
				isLeft := c == c1
				isRight := c == c2
				cacheKey := fmt.Sprintf("border-%d-%t-%t-%t-%t-%d-%s", styleID, isTop, isBottom, isLeft, isRight, weight, colorClean)

				if cachedStyleID, exists := g.borderCache[cacheKey]; exists {
					_ = g.file.SetCellStyle(sheet, cell, cell, cachedStyleID)
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
							borders[i].Color = colorClean
							found = true
							break
						}
					}
					if !found {
						borders = append(borders, excelize.Border{
							Type:  borderType,
							Color: colorClean,
							Style: styleVal,
						})
					}
				}

				if isTop {
					upsertBorder("top", weight)
				}
				if isBottom {
					upsertBorder("bottom", weight)
				}
				if isLeft {
					upsertBorder("left", weight)
				}
				if isRight {
					upsertBorder("right", weight)
				}

				style.Border = borders
				newStyleID, err := g.file.NewStyle(style)
				if err == nil {
					g.borderCache[cacheKey] = newStyleID
					_ = g.file.SetCellStyle(sheet, cell, cell, newStyleID)
				}
			}
		}
	}
	return nil
}

func (g *Generator) reapplyRightBorder(sheet, cell string, weight int, color string) {
	styleID, err := g.file.GetCellStyle(sheet, cell)
	if err != nil {
		return
	}

	colorClean := strings.TrimPrefix(color, "#")
	cacheKey := fmt.Sprintf("rightborder-%d-%d-%s", styleID, weight, colorClean)

	if cachedStyleID, exists := g.borderCache[cacheKey]; exists {
		_ = g.file.SetCellStyle(sheet, cell, cell, cachedStyleID)
		return
	}

	style, err := g.file.GetStyle(styleID)
	if err != nil || style == nil {
		return
	}

	found := false
	for i, b := range style.Border {
		if b.Type == "right" {
			style.Border[i].Style = weight
			style.Border[i].Color = colorClean
			found = true
			break
		}
	}
	if !found {
		style.Border = append(style.Border, excelize.Border{Type: "right", Color: colorClean, Style: weight})
	}

	if newID, err := g.file.NewStyle(style); err == nil {
		g.borderCache[cacheKey] = newID
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
	formula = strings.TrimPrefix(formula, "=")
	return g.file.SetConditionalFormat(sheet, cell, []excelize.ConditionalFormatOptions{
		{
			Type:     "formula",
			Criteria: formula,
			Format:   &styleID,
		},
	})
}

func (g *Generator) bindInputField(sheet string, row, col int, field InputField) error {
	// Name setzen
	err := g.file.SetDefinedName(&excelize.DefinedName{
		Name:     field.NamedRange,
		RefersTo: fmt.Sprintf("'%s'!%s", sheet, absName(col, row)),
	})
	if err != nil {
		return err
	}

	// Validierung setzen, falls vorhanden
	if len(field.Validation) > 0 {
		dv := excelize.NewDataValidation(true)
		dv.Sqref = cellName(col, row)
		dv.SetDropList(field.Validation)
		return g.file.AddDataValidation(sheet, dv)
	}
	return nil
}
