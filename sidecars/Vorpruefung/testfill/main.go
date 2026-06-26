// Command testfill befüllt eine bereits (mit -budget) erzeugte Vorpruefung-Vorlage
// mit kohärenten Beispieldaten, damit sich das Gesamtwerk – inkl. der berechneten
// Auswertung – ohne manuelle Eingaben in Excel testen lässt.
//
// Ablauf:
//
//  1. Vorlage erzeugen (aus sidecars/Vorpruefung):
//     go run . -budget budget.example.json -o vorpruefung_output.xlsx
//  2. Vorlage befüllen:
//     go run ./testfill -in vorpruefung_output.xlsx -budget budget.example.json \
//     -o vorpruefung_befuellt.xlsx
//
// Wichtig: Die Eingabewerte werden NICHT über einen excelize-Round-Trip geschrieben,
// sondern direkt in den Worksheet-XML-Parts des .xlsx-Zips gepatcht. Dadurch bleiben
// die nachträglich gesetzten Dynamic-Array-Metadaten (cm="1" + xl/metadata.xml) der
// VSTACK/FILTER-Spills erhalten – sonst würden die Spills (und damit die Auswertung)
// beim Öffnen in Excel brechen.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// ─── Layout-Konstanten (Spiegel der Generator-Konstanten) ─────────────────────
const (
	dbSheet  = "Dashboard"
	kmwSheet = "II. KMW-Mittel"
	fbSheet  = "III. Finanzberichte"
	maSheet  = "IV. MA"

	// Dashboard-Spalten (B=Label1, C=Eingabe1, D=Label2, E=Eingabe2).
	dbColIn1 = 3 // C
	dbColIn2 = 5 // E

	// Finanzberichte: 5 Spalten + 2 Abstand = 7 Spalten Versatz je Periode,
	// erste Periode beginnt in Spalte B (=2).
	fbStartCol = 2
	fbStride   = 7

	// Erste Ausgaben-Datenzeile einer FB-Periode (Layout für alle Perioden gleich).
	fbAusgFirstRow = 19

	// Mittelanforderung: 3 Spalten + 1 Abstand = 4 Spalten Versatz je Periode.
	maStartCol = 2
	maStride   = 4
)

// maCategories entspricht MA_CATEGORIES im Generator (Reihenfolge = Tabellenzeilen 10..17).
var maCategories = []string{
	"Bauausgaben", "Investitionen", "Personalkosten",
	"Projektaktivitaeten", "Projektverwaltung",
	"Evaluierung", "Audit", "Reserve",
}

// ─── Beispiel-Szenario ────────────────────────────────────────────────────────
//
// Durchgehender Kurs: 125 LC = 1 EUR (= Budget-Kurs der budget.example.json).
// Zwei Berichts-/Anforderungsperioden mit sauberem Saldo-Übertrag von P1 nach P2.
const exRate = 125.0

// Vorprojektsaldo (Saldovortrag) aus dem Dashboard. Fließt über die benannten
// Bereiche Saldovortrag_LW/_EUR in den FB-Vorprojektsaldo der Periode 1 und in den
// MA-Abzug "Saldo Vorprojekt". Die Bank-Aufschlüsselung der FB-Perioden ist darauf
// abgestimmt (siehe bankLC), damit die Differenzprüfung 0 bleibt.
const (
	exSaldovortragLC  = 200_000.0
	exSaldovortragEUR = exSaldovortragLC / exRate // 1.600
)

// fbPeriod beschreibt die Beispieleingaben einer FB-Periode.
type fbPeriod struct {
	von, bis time.Time
	// Einnahmen (LC) je Typ.
	kmwLC, eigenLC, drittLC float64
	// Ausgaben (LC) je Budget-ID.
	ausgabenByID map[string]float64
	// Aufschlüsselung des Saldos auf "Bank" (LC). Muss dem berechneten FB-Saldo
	// entsprechen, damit die Differenz-Prüfung 0 ergibt.
	bankLC float64
}

// maPeriod beschreibt die Beispieleingaben einer MA-Periode.
type maPeriod struct {
	von, bis     time.Time
	rate         float64
	kategorienLC map[string]float64
	eigenLC      float64
	drittLC      float64
}

var fbPeriods = []fbPeriod{
	{ // Periode 1
		von: date(2025, 1, 1), bis: date(2025, 6, 30),
		kmwLC: 1_250_000, eigenLC: 250_000, drittLC: 125_000,
		ausgabenByID: map[string]float64{
			"1.1": 600_000, "2.1": 300_000, "3.1": 200_000,
			"3.2": 150_000, "4.1": 120_000, "5.1": 80_000, "7.1": 50_000,
		},
		// (200.000 Vorprojektsaldo + 1.625.000 Einnahmen) − 1.500.000 Ausgaben = 325.000
		bankLC: 325_000,
	},
	{ // Periode 2 (Vorperiodensaldo 125.000 wird automatisch übernommen)
		von: date(2025, 7, 1), bis: date(2025, 12, 31),
		kmwLC: 1_000_000, eigenLC: 200_000, drittLC: 100_000,
		ausgabenByID: map[string]float64{
			"1.1": 400_000, "1.2": 300_000, "3.1": 200_000,
			"3.2": 150_000, "4.1": 100_000, "5.1": 70_000, "7.1": 30_000,
		},
		// (325.000 Vorperiodensaldo + 1.300.000 neue Einnahmen) − 1.250.000 Ausgaben = 375.000
		bankLC: 375_000,
	},
}

var maPeriods = []maPeriod{
	{ // MA Periode 1
		von: date(2025, 1, 1), bis: date(2025, 6, 30), rate: exRate,
		kategorienLC: map[string]float64{
			"Bauausgaben": 700_000, "Investitionen": 300_000, "Personalkosten": 350_000,
			"Projektaktivitaeten": 130_000, "Projektverwaltung": 80_000, "Audit": 50_000,
		},
		eigenLC: 250_000, drittLC: 125_000,
	},
	{ // MA Periode 2
		von: date(2025, 7, 1), bis: date(2025, 12, 31), rate: exRate,
		kategorienLC: map[string]float64{
			"Bauausgaben": 400_000, "Personalkosten": 350_000,
			"Projektaktivitaeten": 100_000, "Projektverwaltung": 70_000, "Audit": 30_000,
		},
		eigenLC: 200_000, drittLC: 100_000,
	},
}

// kmwRow beschreibt eine bereitgestellte KMW-Mittel-Tranche (Blatt II).
type kmwRow struct {
	periode  string
	waehrung string
	betrag   float64
	datum    time.Time
}

var kmwRows = []kmwRow{
	{"Periode 1", "EUR", 10_000, date(2025, 1, 15)},
	{"Periode 1", "EUR", 8_000, date(2025, 4, 15)},
	{"Periode 2", "EUR", 9_000, date(2025, 7, 15)},
}

// ─── Budget-Teilstruktur (kanonische BudgetData, nur Positionen zum Befüllen) ──
type budgetCfg struct {
	Positions []struct {
		Number    string   `json:"number"`
		Label     string   `json:"label"`
		Kategorie string   `json:"kategorie"`
		LC        *float64 `json:"lc"`
		Y1        *float64 `json:"y1"`
		Y2        *float64 `json:"y2"`
		Y3        *float64 `json:"y3"`
		EUR       *float64 `json:"eur"`
	} `json:"positions"`
}

func main() {
	var inPath, budgetPath, outPath string
	flag.StringVar(&inPath, "in", "vorpruefung_output.xlsx", "mit -budget erzeugte Eingabe-Vorlage (.xlsx)")
	flag.StringVar(&budgetPath, "budget", "budget.example.json", "Budget-JSON (für Ausgaben-IDs/Anzahl)")
	flag.StringVar(&outPath, "o", "vorpruefung_befuellt.xlsx", "Zieldatei (.xlsx)")
	flag.Parse()

	ausgIDs, err := loadAusgabenIDs(budgetPath)
	if err != nil {
		log.Fatalf("budget laden: %v", err)
	}
	if len(ausgIDs) == 0 {
		log.Fatalf("budget %q enthält keine Ausgaben – bitte eine -budget-Datei mit Positionen angeben", budgetPath)
	}

	parts, order, err := readZip(inPath)
	if err != nil {
		log.Fatalf("vorlage lesen: %v", err)
	}

	sheetPart, err := mapSheetNamesToParts(parts)
	if err != nil {
		log.Fatalf("sheet-zuordnung: %v", err)
	}

	// Eingaben je Blatt sammeln und in einem Rutsch in das jeweilige XML patchen.
	if err := patchSheet(parts, sheetPart, dbSheet, dashboardEdits()); err != nil {
		log.Fatalf("Dashboard befüllen: %v", err)
	}
	if err := patchSheet(parts, sheetPart, kmwSheet, kmwEdits()); err != nil {
		log.Fatalf("KMW-Mittel befüllen: %v", err)
	}
	if err := patchSheet(parts, sheetPart, fbSheet, fbEdits(ausgIDs)); err != nil {
		log.Fatalf("Finanzberichte befüllen: %v", err)
	}
	if err := patchSheet(parts, sheetPart, maSheet, maEdits()); err != nil {
		log.Fatalf("Mittelanforderung befüllen: %v", err)
	}

	if err := writeZip(outPath, order, parts); err != nil {
		log.Fatalf("zieldatei schreiben: %v", err)
	}

	fmt.Printf("Vorlage befüllt: %s\n", outPath)
	fmt.Printf("  • Dashboard:       Projektstammdaten + Vorprojektsaldo %g LC / %g EUR\n", exSaldovortragLC, exSaldovortragEUR)
	fmt.Printf("  • KMW-Mittel:      %d Tranchen\n", len(kmwRows))
	fmt.Printf("  • Finanzberichte:  %d Perioden (Kurs %g LC/EUR, Saldo-Übertrag)\n", len(fbPeriods), exRate)
	fmt.Printf("  • Mittelanforderung: %d Perioden\n", len(maPeriods))
	fmt.Println("In Excel öffnen – dank FullCalcOnLoad rechnen alle Blätter inkl. Auswertung automatisch.")
}

// ─── Edit-Modell ──────────────────────────────────────────────────────────────

type editKind int

const (
	editNum editKind = iota
	editStr
	editDate
)

type cellEdit struct {
	ref  string
	kind editKind
	num  float64
	str  string
}

func numEdit(ref string, v float64) cellEdit { return cellEdit{ref: ref, kind: editNum, num: v} }
func strEdit(ref, s string) cellEdit         { return cellEdit{ref: ref, kind: editStr, str: s} }
func dateEdit(ref string, t time.Time) cellEdit {
	return cellEdit{ref: ref, kind: editDate, num: excelSerial(t)}
}

// ─── Dashboard (Blatt I/Statische Projektinformationen) ───────────────────────
// Eingabespalten: C (=dbColIn1) und E (=dbColIn2). Zeilen siehe drawStaticProjectInfo.
func dashboardEdits() []cellEdit {
	c := func(row int) string { return cell(dbColIn1, row) } // Spalte C
	e := func(row int) string { return cell(dbColIn2, row) } // Spalte E
	var edits []cellEdit
	edits = append(edits,
		// Zeile 5: Projektnummer | Vorprojekt vorhanden (Ja → Vorprojekt-Block aktiv)
		strEdit(c(5), "PRJ-2025-042"),
		strEdit(e(5), "Ja"),
		// Zeile 6: Projekttitel (C6:E6 verbunden, Anker C6)
		strEdit(c(6), "Aufbau Gemeindezentrum Beispielstadt"),
		// Zeile 7: Projekttraeger | Berichtswaehrung (aus Währungsliste)
		strEdit(c(7), "Beispiel Hilfswerk e.V."),
		strEdit(e(7), "USD"),
		// Zeile 8: Projektstart | Projektende (Datum-Eingaben; Projektlaufzeit/In-Monate sind Formeln)
		dateEdit(c(8), date(2025, 1, 1)),
		dateEdit(e(8), date(2027, 12, 31)),

		// ── Vorprojekt-Block ──
		// Zeile 10: Vorprojektnummer | VP-Berichtswaehrung
		strEdit(c(10), "PRJ-2022-017"),
		strEdit(e(10), "USD"),
		// Zeile 11: Vorprojektende | Wechselkurs
		dateEdit(c(11), date(2024, 12, 31)),
		numEdit(e(11), exRate),
		// Zeile 12: Saldo (LW) | Saldo (EUR) bei Vorprojektende
		numEdit(c(12), exSaldovortragLC),
		numEdit(e(12), exSaldovortragEUR),
		// Zeile 13: Folgeprojektstart | Wechselkurs
		dateEdit(c(13), date(2025, 1, 1)),
		numEdit(e(13), exRate),
		// Zeile 14: Saldovortrag (LW)/(EUR) → benannte Bereiche (FB-/MA-Vorprojektsaldo)
		numEdit(c(14), exSaldovortragLC),
		numEdit(e(14), exSaldovortragEUR),
	)

	// Dokumenten-Checkliste: Dropdowns D16..D22 auf "Ja" (alle Belege liegen vor).
	const docFirstRow, docCount, docDropdownCol = 16, 7, 4 // Spalte D
	for i := 0; i < docCount; i++ {
		edits = append(edits, strEdit(cell(docDropdownCol, docFirstRow+i), "Ja"))
	}
	return edits
}

// ─── KMW-Mittel (Blatt II): Tabelle B5:E22 ────────────────────────────────────
func kmwEdits() []cellEdit {
	var edits []cellEdit
	for i, kr := range kmwRows {
		row := 5 + i // erste Datenzeile = 5
		edits = append(edits,
			strEdit(cell(2, row), kr.periode),  // B: Periode
			strEdit(cell(3, row), kr.waehrung), // C: Waehrung
			numEdit(cell(4, row), kr.betrag),   // D: Betrag
			dateEdit(cell(5, row), kr.datum),   // E: Datum
		)
	}
	return edits
}

// ─── Finanzberichte (Blatt III): 18 Perioden nebeneinander ────────────────────
func fbEdits(ausgIDs []string) []cellEdit {
	var edits []cellEdit
	n := len(ausgIDs)
	for p, fp := range fbPeriods {
		colStart := fbStartCol + p*fbStride
		l := fbLayout(n)

		// Zeitraum Von/Bis (Eingabe in cValLC = colStart+1, Datums-Zellen).
		edits = append(edits,
			dateEdit(cell(colStart+1, 5), fp.von),
			dateEdit(cell(colStart+1, 6), fp.bis),
		)

		// Ausgaben (LC) je Budget-ID: Eingabespalte = colStart+1, Zeilen ab 19.
		for i, id := range ausgIDs {
			if v, ok := fp.ausgabenByID[id]; ok && v != 0 {
				edits = append(edits, numEdit(cell(colStart+1, fbAusgFirstRow+i), v))
			}
		}

		// Aufschlüsselung "Bank" (LC) – Eingabespalte colStart+1.
		if fp.bankLC != 0 {
			edits = append(edits, numEdit(cell(colStart+1, l.aufschStartRow), fp.bankLC))
		}

		// Detail-Tabelle 1 "Einnahmen (Explizite Kurseingabe)":
		//   LC = colStart+2, EUR = colStart+2+1 (Kurs berechnet sich daraus).
		//   Vorbelegte Zeile "KMW-Mittel" liegt auf headerRow+2.
		kmwRowT1 := l.tbl1HeaderRow + 2
		if fp.kmwLC != 0 {
			edits = append(edits,
				numEdit(cell(colStart+2, kmwRowT1), fp.kmwLC),
				numEdit(cell(colStart+3, kmwRowT1), fp.kmwLC/exRate),
			)
		}

		// Detail-Tabelle 2 "Einnahmen (Durchschnittskurs)": nur LC = colStart+2,
		//   EUR berechnet sich über den Durchschnittskurs. Vorbelegte Zeilen:
		//   Eigenmittel = headerRow+1, Drittmittel = headerRow+2, Zinsertraege = headerRow+3.
		if fp.eigenLC != 0 {
			edits = append(edits, numEdit(cell(colStart+2, l.tbl2HeaderRow+1), fp.eigenLC))
		}
		if fp.drittLC != 0 {
			edits = append(edits, numEdit(cell(colStart+2, l.tbl2HeaderRow+2), fp.drittLC))
		}
	}
	return edits
}

// fbLayoutInfo hält die von der Ausgaben-Zeilenanzahl abhängigen Zeilennummern
// einer FB-Periode (siehe drawReportTable im Generator).
type fbLayoutInfo struct {
	aufschStartRow int // erste Aufschlüsselungszeile (Bank)
	tbl1HeaderRow  int // Kopf Detail-Tabelle 1
	tbl2HeaderRow  int // Kopf Detail-Tabelle 2
}

func fbLayout(ausgDataRows int) fbLayoutInfo {
	const ausgHdrRow = 18 // Einnahmenblock ist fix; Ausgaben-Kopf immer Zeile 18
	ausgTotalsRow := ausgHdrRow + ausgDataRows + 1
	saldoRow := ausgTotalsRow + 2
	aufschLabelRow := saldoRow + 2
	aufschStart := aufschLabelRow + 1 // Bank, Kasse, Sonstiges (3 Zeilen)
	differenzRow := aufschStart + 4
	tbl1Label := differenzRow + 3
	tbl1Header := tbl1Label + 1
	totalsRow1 := tbl1Header + 6 // 5 Datenzeilen + Summenzeile
	tbl2Label := totalsRow1 + 2
	tbl2Header := tbl2Label + 1
	return fbLayoutInfo{
		aufschStartRow: aufschStart,
		tbl1HeaderRow:  tbl1Header,
		tbl2HeaderRow:  tbl2Header,
	}
}

// ─── Mittelanforderung (Blatt IV): 18 Perioden nebeneinander ──────────────────
func maEdits() []cellEdit {
	var edits []cellEdit
	for p, mp := range maPeriods {
		colS := maStartCol + p*maStride
		cLC := colS + 1 // Eingabespalte "Angefordert (LC)"

		edits = append(edits,
			dateEdit(cell(cLC, 5), mp.von), // Von (Zeile 5)
			dateEdit(cell(cLC, 6), mp.bis), // Bis (Zeile 6)
			numEdit(cell(cLC, 8), mp.rate), // OANDA-Kurs (Zeile 8)
		)

		// Kostenkategorien: Tabellenzeilen 10..17 in Reihenfolge maCategories.
		for i, cat := range maCategories {
			if v, ok := mp.kategorienLC[cat]; ok && v != 0 {
				edits = append(edits, numEdit(cell(cLC, 10+i), v))
			}
		}

		// abzueglich Eigenmittel (Zeile 21) / Drittmittel (Zeile 22).
		if mp.eigenLC != 0 {
			edits = append(edits, numEdit(cell(cLC, 21), mp.eigenLC))
		}
		if mp.drittLC != 0 {
			edits = append(edits, numEdit(cell(cLC, 22), mp.drittLC))
		}
	}
	return edits
}

// ─── XML-Patching ─────────────────────────────────────────────────────────────

func patchSheet(parts map[string][]byte, sheetPart map[string]string, sheetName string, edits []cellEdit) error {
	part, ok := sheetPart[sheetName]
	if !ok {
		return fmt.Errorf("blatt %q nicht in workbook.xml gefunden", sheetName)
	}
	data, ok := parts[part]
	if !ok {
		return fmt.Errorf("worksheet-part %q fehlt im Archiv", part)
	}
	xmlStr := string(data)
	for _, e := range edits {
		var inner, typeAttr string
		switch e.kind {
		case editStr:
			inner = "<is><t>" + xmlEscape(e.str) + "</t></is>"
			typeAttr = "inlineStr"
		case editNum, editDate:
			inner = "<v>" + strconv.FormatFloat(e.num, 'f', -1, 64) + "</v>"
		}
		patched, err := patchCell(xmlStr, e.ref, inner, typeAttr)
		if err != nil {
			return fmt.Errorf("%s!%s: %w", sheetName, e.ref, err)
		}
		xmlStr = patched
	}
	parts[part] = []byte(xmlStr)
	return nil
}

// patchCell ersetzt den Inhalt der Zelle <c r="REF" ...> durch inner. Vorhandene
// (leere) Eingabezellen liegen als self-closing <c r=".." s="N"/> vor; der Stil
// (s-Attribut) bleibt erhalten. Ein evtl. vorhandenes t="..."-Attribut wird ersetzt.
func patchCell(xmlStr, ref, inner, typeAttr string) (string, error) {
	marker := `<c r="` + ref + `"`
	i := strings.Index(xmlStr, marker)
	if i == -1 {
		return "", fmt.Errorf("zelle nicht gefunden")
	}
	gt := strings.IndexByte(xmlStr[i:], '>')
	if gt == -1 {
		return "", fmt.Errorf("öffnendes tag unvollständig")
	}
	openEnd := i + gt
	open := xmlStr[i : openEnd+1] // inkl. '>' bzw. '/>'

	selfClosing := strings.HasSuffix(open, "/>")
	var attrs string
	if selfClosing {
		attrs = strings.TrimSuffix(open, "/>")
	} else {
		attrs = strings.TrimSuffix(open, ">")
	}
	attrs = strings.TrimRight(attrs, " ")
	attrs = removeTypeAttr(attrs)
	if typeAttr != "" {
		attrs += ` t="` + typeAttr + `"`
	}
	newCell := attrs + ">" + inner + "</c>"

	if selfClosing {
		return xmlStr[:i] + newCell + xmlStr[openEnd+1:], nil
	}
	rel := strings.Index(xmlStr[i:], "</c>")
	if rel == -1 {
		return "", fmt.Errorf("schließendes </c> nicht gefunden")
	}
	cEnd := i + rel + len("</c>")
	return xmlStr[:i] + newCell + xmlStr[cEnd:], nil
}

// removeTypeAttr entfernt ein evtl. vorhandenes ` t="..."` aus den Zellattributen.
func removeTypeAttr(attrs string) string {
	idx := strings.Index(attrs, ` t="`)
	if idx == -1 {
		return attrs
	}
	end := strings.IndexByte(attrs[idx+4:], '"')
	if end == -1 {
		return attrs
	}
	return attrs[:idx] + attrs[idx+4+end+1:]
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// ─── Zip-Hilfen (analog dynarray.go) ──────────────────────────────────────────

func readZip(filePath string) (map[string][]byte, []string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()
	parts := map[string][]byte{}
	var order []string
	for _, zf := range r.File {
		rc, err := zf.Open()
		if err != nil {
			return nil, nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, nil, err
		}
		parts[zf.Name] = data
		order = append(order, zf.Name)
	}
	return parts, order, nil
}

func writeZip(filePath string, order []string, parts map[string][]byte) error {
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

// mapSheetNamesToParts liefert Worksheet-Anzeigename -> zip-Part-Pfad.
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

// ─── kleine Helfer ────────────────────────────────────────────────────────────

func loadAusgabenIDs(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg budgetCfg
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s ist kein gültiges JSON: %w", path, err)
	}
	// Gleiche Formung wie das Sidecar (config.go mapScannedToBudget): leere
	// Kopfzeilen/Platzhalter auslassen, damit die IDs exakt den Sheet-Zeilen entsprechen.
	zero := func(v *float64) bool { return v == nil || *v == 0 }
	var ids []string
	for _, p := range cfg.Positions {
		if p.Kategorie == "" {
			continue
		}
		sub := ""
		if idx := strings.IndexByte(p.Number, '.'); idx >= 0 {
			sub = strings.TrimSpace(p.Number[idx+1:])
		}
		labelEmpty := strings.TrimSpace(p.Label) == ""
		valueless := zero(p.LC) && zero(p.Y1) && zero(p.Y2) && zero(p.Y3) && zero(p.EUR)
		if valueless && (sub == "" || labelEmpty) {
			continue
		}
		ids = append(ids, p.Number)
	}
	return ids, nil
}

// cell wandelt 1-basierte Spalten-/Zeilennummern in eine A1-Referenz um.
func cell(col, row int) string {
	return colLetter(col) + strconv.Itoa(row)
}

func colLetter(col int) string {
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+col%26)) + name
		col /= 26
	}
	return name
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// excelSerial liefert die Excel-Seriennummer (1900-Datumssystem) für ein Datum.
func excelSerial(t time.Time) float64 {
	epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return float64(int(t.Sub(epoch).Hours() / 24))
}
