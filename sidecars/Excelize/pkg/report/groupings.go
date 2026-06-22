package report

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
)

// regexOutline sucht nach outlineLevel="X" Attributen im XML
var regexOutline = regexp.MustCompile(` outlineLevel="\d+"`)

// RemoveGroupingsFromBytes liest die Excel-ZIP-Struktur im Speicher,
// entfernt alle outlineLevel (Gruppierungen) aus allen Arbeitsblättern
// und gibt die bereinigte Excel-Datei als Byte-Slice zurück.
func RemoveGroupingsFromBytes(excelData []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(excelData), int64(len(excelData)))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for _, file := range reader.File {
		// Datei innerhalb der ZIP öffnen
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}

		// Inhalt lesen
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		// Handelt es sich um ein Arbeitsblatt (z.B. xl/worksheets/sheet1.xml)?
		// Dann entfernen wir die outlineLevel Attribute per Regex.
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
