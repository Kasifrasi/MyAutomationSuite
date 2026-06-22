package report

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
)

// PostProcessExcelFile entfernt versteckte Zeilen/Spalten aus den Worksheets per Regex
func PostProcessExcelFile(excelData []byte, unhideCols bool, unhideRows bool) ([]byte, error) {
	if !unhideCols && !unhideRows {
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
