package report

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
)

// regexOutline sucht nach outlineLevel="X" Attributen im XML
var regexOutline = regexp.MustCompile(` outlineLevel="\d+"`)

// PostProcessExcelFile entfernt Gruppierungen aus den Worksheets per Regex,
// da Excelize hierfür keine Lösch-API bietet.
func PostProcessExcelFile(excelData []byte, removeGroupings bool) ([]byte, error) {
	if !removeGroupings {
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

		// Gruppierungen aus den Worksheets entfernen
		if regexp.MustCompile(`^xl/worksheets/.*\.xml$`).MatchString(file.Name) {
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
