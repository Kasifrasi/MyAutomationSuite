package report

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
)

// regexCalcChain sucht nach der CalcChain-Definition im Content_Types
var regexCalcChain = regexp.MustCompile(`<Override PartName="/xl/calcChain\.xml"[^>]*>(?:</Override>)?`)

// regexCalcChainRel sucht nach der CalcChain-Beziehung in workbook.xml.rels
var regexCalcChainRel = regexp.MustCompile(`<Relationship Id="[^"]+" Target="calcChain\.xml" Type="[^"]*calcChain"[^>]*>(?:</Relationship>)?`)

// PostProcessExcelFile entfernt auf Wunsch Gruppierungen und IMMER die CalcChain,
// um korrupte Dateien durch verwaiste Formel-Caches zu vermeiden.
func PostProcessExcelFile(excelData []byte, removeGroupings bool) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(excelData), int64(len(excelData)))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, file := range reader.File {
		// CalcChain komplett aus der ZIP weglassen
		if file.Name == "xl/calcChain.xml" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		// CalcChain-Eintrag aus [Content_Types].xml löschen
		if file.Name == "[Content_Types].xml" {
			content = regexCalcChain.ReplaceAll(content, []byte(""))
		}

		// CalcChain-Beziehung aus xl/_rels/workbook.xml.rels löschen
		if file.Name == "xl/_rels/workbook.xml.rels" {
			content = regexCalcChainRel.ReplaceAll(content, []byte(""))
		}

		// Auf Wunsch auch Gruppierungen entfernen
		if removeGroupings && regexp.MustCompile(`^xl/worksheets/.*\.xml$`).MatchString(file.Name) {
			content = regexOutline.ReplaceAll(content, []byte(""))
		}

		// Datei wieder in die neue ZIP schreiben
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

// regexOutline sucht nach outlineLevel="X" Attributen im XML
var regexOutline = regexp.MustCompile(` outlineLevel="\d+"`)
