package report

import (
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	ColorWhite       = "#FFFFFF"
	ColorLightGray   = "#F2F2F2"
	ColorLightYellow = "#FFFAE5"
	ColorDarkGray    = "#D8D8D8"
	ColorBlack       = "#000000"
)

// Hilfsfunktion zum Suchen von Zeilen basierend auf Spalte B
func findRowByColB(rows [][]string, match func(string) bool) int {
	colIdx := 1 // Spalte B ist Index 1 (0-basiert)
	for i, row := range rows {
		if len(row) > colIdx && match(strings.TrimSpace(row[colIdx])) {
			return i + 1 // 1-basiert für Excel
		}
	}
	return -1
}

// modStyleHelper nimmt den existierenden Style einer Zelle, ändert nur die Hintergrundfarbe und gibt die neue Style-ID zurück.
// So bleiben Borders, Fonts (wie Bold) etc. erhalten.
func modStyleHelper(f *excelize.File, sheet, cell, hexColor string) int {
	return modStyleHelperEx(f, sheet, cell, hexColor, false, "")
}

func modStyleHelperEx(f *excelize.File, sheet, cell, hexColor string, setBoldFalse bool, fontColor string) int {
	styleID, err := f.GetCellStyle(sheet, cell)
	if err != nil {
		return 0
	}
	style, err := f.GetStyle(styleID)
	if err != nil || style == nil {
		return 0
	}

	// Nur die Hintergrundfarbe ändern
	if hexColor != "" {
		style.Fill = excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{hexColor},
		}
	}

	if setBoldFalse {
		if style.Font != nil {
			style.Font.Bold = false
		} else {
			style.Font = &excelize.Font{Bold: false}
		}
	}

	if fontColor != "" {
		if style.Font != nil {
			style.Font.Color = fontColor
			style.Font.ColorTheme = nil
		} else {
			style.Font = &excelize.Font{Color: fontColor}
		}
	}

	newID, _ := f.NewStyle(style)
	return newID
}

// Hilfsfunktion: Setzt das Zellformat auf "Allgemein" (NumFmt = 0)
func setGeneralFormat(f *excelize.File, sheet string, cell string) {
	styleID, err := f.GetCellStyle(sheet, cell)
	if err == nil {
		style, err := f.GetStyle(styleID)
		if err == nil && style != nil {
			style.NumFmt = 0 // 0 = Allgemein (General)
			newStyleID, err := f.NewStyle(style)
			if err == nil {
				f.SetCellStyle(sheet, cell, cell, newStyleID)
			}
		}
	}
}
