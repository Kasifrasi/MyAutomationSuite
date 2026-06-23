package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// setDynArrayFormula schreibt eine dynamische Array-Formel (Spill), z.B. VSTACK/FILTER.
// Der Array-Typ (t="array") sorgt dafür, dass Excel die Zelle als Array-Formel behandelt
// und beim Öffnen NICHT den impliziten Schnittmengen-Operator "@" einfügt. Die endgültige
// Dynamic-Array-Markierung (aca/cm + metadata.xml) wird nach dem Speichern in
// applyDynamicArrayMetadata gesetzt.
func setDynArrayFormula(f *excelize.File, ws, cell, formula string) {
	ref := cell
	formulaType := excelize.STCellFormulaTypeArray
	_ = f.SetCellFormula(ws, cell, formula, excelize.FormulaOpts{Ref: &ref, Type: &formulaType})
}

// dynamicArrayMetadataXML ist exakt die Struktur, die Excel selbst für dynamische
// Array-Formeln in xl/metadata.xml ablegt (ein einzelner XLDAPR-Cell-Metadaten-Block).
const dynamicArrayMetadataXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<metadata xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:xlrd="http://schemas.microsoft.com/office/spreadsheetml/2017/richdata" xmlns:xda="http://schemas.microsoft.com/office/spreadsheetml/2017/dynamicarray"><metadataTypes count="1"><metadataType name="XLDAPR" minSupportedVersion="120000" copy="1" pasteAll="1" pasteValues="1" merge="1" splitFirst="1" rowColShift="1" clearFormats="1" clearComments="1" assign="1" coerce="1" cellMeta="1"/></metadataTypes><futureMetadata name="XLDAPR" count="1"><bk><extLst><ext uri="{bdbb8cdc-fa1e-496e-a857-3c3f30c029c3}"><xda:dynamicArrayProperties fDynamic="1" fCollapsed="0"/></ext></extLst></bk></futureMetadata><cellMetadata count="1"><bk><rc t="1" v="0"/></bk></cellMetadata></metadata>`

const (
	dynArrayMetadataPart        = "xl/metadata.xml"
	dynArrayMetadataContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheetMetadata+xml"
	dynArrayMetadataRelType     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/sheetMetadata"
)

// applyDynamicArrayMetadata öffnet die fertige .xlsx, markiert die übergebenen Zellen
// als echte dynamische Array-Formeln (cm + aca/ca) und ergänzt xl/metadata.xml samt
// Content-Type und Workbook-Relationship. So spillen VSTACK/FILTER beim Öffnen sauber,
// ohne dass Excel den "@"-Operator einfügt.
func applyDynamicArrayMetadata(filePath string, cells []dynArrayCell) error {
	if len(cells) == 0 {
		return nil
	}

	r, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("öffnen der xlsx fehlgeschlagen: %w", err)
	}

	parts := map[string][]byte{}
	var order []string
	for _, zf := range r.File {
		rc, err := zf.Open()
		if err != nil {
			r.Close()
			return err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			r.Close()
			return err
		}
		parts[zf.Name] = data
		order = append(order, zf.Name)
	}
	r.Close()

	// Anker-Zellen je Worksheet-Part bestimmen (Anzeigename -> sheetN.xml).
	sheetFileByName, err := mapSheetNamesToParts(parts)
	if err != nil {
		return err
	}
	cellsByPart := map[string][]string{}
	for _, c := range cells {
		part, ok := sheetFileByName[c.sheet]
		if !ok {
			return fmt.Errorf("worksheet %q nicht in workbook.xml gefunden", c.sheet)
		}
		cellsByPart[part] = append(cellsByPart[part], c.cell)
	}

	// 1. Zellen in den Worksheet-Parts als dynamische Array-Formeln markieren.
	for part, anchors := range cellsByPart {
		data, ok := parts[part]
		if !ok {
			return fmt.Errorf("worksheet-part %q fehlt im Archiv", part)
		}
		patched, err := markDynamicArrayCells(string(data), anchors)
		if err != nil {
			return fmt.Errorf("%s: %w", part, err)
		}
		parts[part] = []byte(patched)
	}

	// 2. metadata.xml ergänzen.
	if _, exists := parts[dynArrayMetadataPart]; !exists {
		parts[dynArrayMetadataPart] = []byte(dynamicArrayMetadataXML)
		order = append(order, dynArrayMetadataPart)
	}

	// 3. Content-Type registrieren.
	ct, ok := parts["[Content_Types].xml"]
	if !ok {
		return fmt.Errorf("[Content_Types].xml fehlt im Archiv")
	}
	parts["[Content_Types].xml"] = []byte(ensureContentTypeOverride(string(ct)))

	// 4. Workbook-Relationship ergänzen.
	const relsPart = "xl/_rels/workbook.xml.rels"
	rels, ok := parts[relsPart]
	if !ok {
		return fmt.Errorf("%s fehlt im Archiv", relsPart)
	}
	parts[relsPart] = []byte(ensureMetadataRelationship(string(rels)))

	return rewriteZip(filePath, order, parts)
}

// markDynamicArrayCells markiert die angegebenen Anker-Zellen exakt so, wie es Excel
// (und rust_xlsxwriter) für echte dynamische Array-Formeln tut:
//
//	<c r="C2" s="3" cm="1"><f t="array" ref="C2">FORMEL</f><v>0</v></c>
//
// Entscheidend: NUR cm="1" (Verweis auf XLDAPR in metadata.xml) markiert die Formel als
// dynamisch. Die Attribute aca/ca dürfen NICHT gesetzt werden - sie machen daraus eine
// Legacy-CSE-Array-Formel, woraufhin Excel beim Öffnen den "@"-Operator wieder einfügt.
// Der skalare Zelltyp t="str" wird entfernt und ein <v>-Cache ergänzt; per
// fullCalcOnLoad rechnet Excel beim Öffnen ohnehin neu.
func markDynamicArrayCells(sheetXML string, anchors []string) (string, error) {
	for _, anchor := range anchors {
		cStart := strings.Index(sheetXML, fmt.Sprintf(`<c r="%s"`, anchor))
		if cStart == -1 {
			return "", fmt.Errorf("zelle %s nicht gefunden", anchor)
		}
		rel := strings.Index(sheetXML[cStart:], "</c>")
		if rel == -1 {
			return "", fmt.Errorf("schließendes </c> für %s nicht gefunden", anchor)
		}
		cEnd := cStart + rel + len("</c>")
		cell := sheetXML[cStart:cEnd]

		if !strings.Contains(cell, "<f t=\"array\"") {
			return "", fmt.Errorf("zelle %s enthält keine Array-Formel", anchor)
		}

		// Legacy-CSE-Marker entfernen und skalaren Zelltyp löschen.
		cell = strings.Replace(cell, ` aca="1"`, "", 1)
		cell = strings.Replace(cell, ` ca="1"`, "", 1)
		cell = strings.Replace(cell, ` t="str"`, "", 1)

		// cm="1" am Zellen-Element ergänzen (Verweis auf cellMetadata-Block 1).
		if !strings.Contains(cell, ` cm="1"`) {
			cell = strings.Replace(cell, fmt.Sprintf(`<c r="%s"`, anchor), fmt.Sprintf(`<c r="%s" cm="1"`, anchor), 1)
		}

		// <v>-Cache ergänzen, falls excelize keinen geschrieben hat.
		if !strings.Contains(cell, "<v>") {
			cell = strings.Replace(cell, "</c>", "<v>0</v></c>", 1)
		}

		sheetXML = sheetXML[:cStart] + cell + sheetXML[cEnd:]
	}
	return sheetXML, nil
}

func ensureContentTypeOverride(contentTypes string) string {
	if strings.Contains(contentTypes, dynArrayMetadataContentType) {
		return contentTypes
	}
	override := fmt.Sprintf(`<Override PartName="/%s" ContentType="%s"/>`, dynArrayMetadataPart, dynArrayMetadataContentType)
	return strings.Replace(contentTypes, "</Types>", override+"</Types>", 1)
}

func ensureMetadataRelationship(rels string) string {
	if strings.Contains(rels, dynArrayMetadataRelType) {
		return rels
	}
	rel := fmt.Sprintf(`<Relationship Id="%s" Type="%s" Target="metadata.xml"/>`, nextRelID(rels), dynArrayMetadataRelType)
	return strings.Replace(rels, "</Relationships>", rel+"</Relationships>", 1)
}

// nextRelID liefert eine freie rId, indem die höchste vorhandene Nummer +1 verwendet wird.
func nextRelID(rels string) string {
	max := 0
	for _, m := range relIDRegexpFindAll(rels) {
		if n, err := strconv.Atoi(m); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("rId%d", max+1)
}

// relIDRegexpFindAll extrahiert die Nummern aller Id="rIdN"-Vorkommen ohne regexp-Paket.
func relIDRegexpFindAll(rels string) []string {
	var out []string
	const marker = `Id="rId`
	rest := rels
	for {
		i := strings.Index(rest, marker)
		if i == -1 {
			break
		}
		rest = rest[i+len(marker):]
		j := strings.IndexByte(rest, '"')
		if j == -1 {
			break
		}
		out = append(out, rest[:j])
		rest = rest[j:]
	}
	return out
}

// mapSheetNamesToParts liefert die Zuordnung Worksheet-Anzeigename -> zip-Part-Pfad.
func mapSheetNamesToParts(parts map[string][]byte) (map[string]string, error) {
	type xlsxSheet struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"id,attr"`
	}
	type xlsxWorkbook struct {
		Sheets []xlsxSheet `xml:"sheets>sheet"`
	}
	type relationship struct {
		ID     string `xml:"Id,attr"`
		Target string `xml:"Target,attr"`
	}
	type relationships struct {
		Rels []relationship `xml:"Relationship"`
	}

	wbData, ok := parts["xl/workbook.xml"]
	if !ok {
		return nil, fmt.Errorf("xl/workbook.xml fehlt im Archiv")
	}
	var wb xlsxWorkbook
	if err := xml.Unmarshal(wbData, &wb); err != nil {
		return nil, fmt.Errorf("workbook.xml parsen: %w", err)
	}

	relData, ok := parts["xl/_rels/workbook.xml.rels"]
	if !ok {
		return nil, fmt.Errorf("xl/_rels/workbook.xml.rels fehlt im Archiv")
	}
	var rels relationships
	if err := xml.Unmarshal(relData, &rels); err != nil {
		return nil, fmt.Errorf("workbook.xml.rels parsen: %w", err)
	}
	targetByID := map[string]string{}
	for _, rel := range rels.Rels {
		targetByID[rel.ID] = rel.Target
	}

	result := map[string]string{}
	for _, s := range wb.Sheets {
		target := targetByID[s.RID]
		if target == "" {
			continue
		}
		target = strings.TrimPrefix(target, "/xl/")
		if !strings.HasPrefix(target, "xl/") {
			target = path.Join("xl", target)
		}
		result[s.Name] = target
	}
	return result, nil
}

// rewriteZip schreibt das Archiv mit den (teils geänderten) Parts neu an dieselbe Stelle.
func rewriteZip(filePath string, order []string, parts map[string][]byte) error {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range order {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := w.Write(parts[name]); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(filePath, buf.Bytes(), 0o644)
}
