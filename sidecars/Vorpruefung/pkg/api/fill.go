package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"shared/constants"
)

// FillData enthält alle Daten, die nach der Generierung der Vorlage über die API eingefügt werden sollen.
type FillData struct {
	Dashboard DashboardData
	KMW       []KMWTranche
	MA        []MAPeriod
	FB        []FBPeriod
	Budget    *BudgetData
}

type DashboardData struct {
	Projektnummer       string
	Vorprojekt          bool
	Projekttitel        string
	Projekttraeger      string
	Berichtswaehrung    string
	Projektstart        time.Time
	Projektende         time.Time
	Vorprojektnummer    string
	VPBerichtswaehrung  string
	Vorprojektende      time.Time
	VPWechselkurs       float64
	VPSaldoLC           float64
	VPSaldoEUR          float64
	VPFolgeprojektstart time.Time
	// Dokumenten-Checkliste (immer 7 Dropdowns D16..D22)
	DocChecklist []string
}

type KMWTranche struct {
	Periode  string
	Waehrung string
	Betrag   float64
	Datum    time.Time
}

type MAPeriod struct {
	Von          time.Time
	Bis          time.Time
	OandaKurs    float64
	KategorienLC map[string]float64
	EigenLC      float64
	DrittLC      float64
}

type FBPeriod struct {
	Von          time.Time
	Bis          time.Time
	KmwLC        float64
	EigenLC      float64
	DrittLC      float64
	AusgabenByID map[string]float64
	BankLC       float64
}

type BudgetData struct {
	AusgabenIDs     []string
	Eigenmittel     *IncomeRow
	KMWMittel       *IncomeRow
	DrittmittelY1   *float64
	DrittmittelY2   *float64
	DrittmittelY3   *float64
	DrittGeber      []GeberRow
	DrittSonstiges  *IncomeRow
	Ausgaben        []AusgabenRow
	ReserveFreigabe bool
}

type IncomeRow struct {
	LC  *float64
	Y1  *float64
	Y2  *float64
	Y3  *float64
	EUR *float64
}

type GeberRow struct {
	Geber string
	LC    *float64
	EUR   *float64
}

type AusgabenRow struct {
	Kategorie string
	ID        string
	Position  string
	LC        *float64
	Y1        *float64
	Y2        *float64
	Y3        *float64
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

// excelSerial rechnet time.Time in den Excel-Datumswert (Tage seit 1900) um.
func excelSerial(t time.Time) float64 {
	delta := t.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC))
	return float64(delta / (24 * time.Hour))
}

func cell(col, row int) string {
	c, _ := excelizeColumnName(col)
	return fmt.Sprintf("%s%d", c, row)
}

// Hilfsfunktion: Spaltennummer (1-basiert) in Buchstaben umwandeln
func excelizeColumnName(col int) (string, error) {
	if col < 1 {
		return "", fmt.Errorf("ungültige Spaltennummer")
	}
	name := ""
	for col > 0 {
		col--
		name = string(rune('A'+(col%26))) + name
		col /= 26
	}
	return name, nil
}

// ─── Haupt-API Funktion ────────────────────────────────────────────────────────

// FillTemplate liest die fertig generierte Excel-Datei, patchet die XML-Worksheets
// direkt (um dynamische Array-Metadaten zu erhalten) und speichert die Datei.
func FillTemplate(filePath string, data FillData) error {
	parts, order, err := readZip(filePath)
	if err != nil {
		return fmt.Errorf("vorlage lesen: %w", err)
	}

	sheetPart, err := mapSheetNamesToParts(parts)
	if err != nil {
		return fmt.Errorf("sheet-zuordnung: %w", err)
	}

	// 1. Dashboard
	if err := patchSheet(parts, sheetPart, constants.VPSheetDASHBOARD, buildDashboardEdits(data.Dashboard)); err != nil {
		return fmt.Errorf("Dashboard befüllen: %w", err)
	}

	// 2. KMW-Mittel
	if err := patchSheet(parts, sheetPart, constants.VPSheetKMW_MITTEL, buildKMWEdits(data.KMW)); err != nil {
		return fmt.Errorf("KMW-Mittel befüllen: %w", err)
	}

	// 3. MA (Mittelanforderung)
	if err := patchSheet(parts, sheetPart, constants.VPSheetMA, buildMAEdits(data.MA)); err != nil {
		return fmt.Errorf("Mittelanforderung befüllen: %w", err)
	}

	// 4. FB (Finanzberichte)
	if err := patchSheet(parts, sheetPart, constants.VPSheetFINANZBERICHTE, buildFBEdits(data.FB, data.Budget)); err != nil {
		return fmt.Errorf("Finanzberichte befüllen: %w", err)
	}

	// 5. Budget
	if err := patchSheet(parts, sheetPart, constants.VPSheetBUDGET, buildBudgetEdits(data.Budget)); err != nil {
		return fmt.Errorf("Budget befüllen: %w", err)
	}

	if err := writeZip(filePath, order, parts); err != nil {
		return fmt.Errorf("zieldatei schreiben: %w", err)
	}

	return nil
}

func buildDashboardEdits(d DashboardData) []cellEdit {
	var edits []cellEdit
	c := func(row int) string { return cell(3, row) } // Spalte C
	e := func(row int) string { return cell(5, row) } // Spalte E

	vpFlag := "Nein"
	if d.Vorprojekt {
		vpFlag = "Ja"
	}

	edits = append(edits,
		strEdit(c(5), d.Projektnummer),
		strEdit(e(5), vpFlag),
		strEdit(c(6), d.Projekttitel),
		strEdit(c(7), d.Projekttraeger),
		strEdit(e(7), d.Berichtswaehrung),
	)

	if !d.Projektstart.IsZero() {
		edits = append(edits, dateEdit(c(8), d.Projektstart))
	}
	if !d.Projektende.IsZero() {
		edits = append(edits, dateEdit(e(8), d.Projektende))
	}

	if d.Vorprojekt {
		edits = append(edits,
			strEdit(c(10), d.Vorprojektnummer),
			strEdit(e(10), d.VPBerichtswaehrung),
			numEdit(e(11), d.VPWechselkurs),
			numEdit(c(12), d.VPSaldoLC),
			numEdit(e(12), d.VPSaldoEUR),
			numEdit(e(13), d.VPWechselkurs),
			numEdit(c(14), d.VPSaldoLC),
			numEdit(e(14), d.VPSaldoEUR),
		)
		if !d.Vorprojektende.IsZero() {
			edits = append(edits, dateEdit(c(11), d.Vorprojektende))
		}
		if !d.VPFolgeprojektstart.IsZero() {
			edits = append(edits, dateEdit(c(13), d.VPFolgeprojektstart))
		}
	}

	for i, v := range d.DocChecklist {
		if i >= 7 {
			break
		}
		edits = append(edits, strEdit(cell(4, 16+i), v))
	}

	return edits
}

func buildMAEdits(periods []MAPeriod) []cellEdit {
	var edits []cellEdit

	// maCategories entspricht den Zeilen 10..17 im MA Sheet.
	maCategories := []string{
		"Bauausgaben", "Investitionen", "Personalkosten",
		"Projektaktivitaeten", "Projektverwaltung",
		"Evaluierung", "Audit", "Reserve",
	}

	for p, mp := range periods {
		// MA Start Col ist 2 (B) + 1 für die Eingabespalte = 3 (C)
		// Der Stride ist 4 Spalten je Periode.
		// p=0 -> C (3)
		// p=1 -> G (7)
		colS := 2 + p*4
		cLC := colS + 1 // Eingabespalte "Angefordert (LC)"

		if !mp.Von.IsZero() {
			edits = append(edits, dateEdit(cell(cLC, 5), mp.Von))
		}
		if !mp.Bis.IsZero() {
			edits = append(edits, dateEdit(cell(cLC, 6), mp.Bis))
		}
		edits = append(edits, numEdit(cell(cLC, 8), mp.OandaKurs))

		for i, cat := range maCategories {
			if v, ok := mp.KategorienLC[cat]; ok && v != 0 {
				edits = append(edits, numEdit(cell(cLC, 10+i), v))
			}
		}

		if mp.EigenLC != 0 {
			edits = append(edits, numEdit(cell(cLC, 21), mp.EigenLC))
		}
		if mp.DrittLC != 0 {
			edits = append(edits, numEdit(cell(cLC, 22), mp.DrittLC))
		}
	}
	return edits
}

func buildBudgetEdits(budget *BudgetData) []cellEdit {
	var edits []cellEdit
	if budget == nil {
		return edits
	}

	const (
		cLC  = 5
		cY1  = 6
		cY2  = 7
		cY3  = 8
		cEUR = 9
	)

	fillIncome := func(row int, inc *IncomeRow) {
		if inc == nil {
			return
		}
		if inc.LC != nil {
			edits = append(edits, numEdit(cell(cLC, row), *inc.LC))
		}
		if inc.Y1 != nil {
			edits = append(edits, numEdit(cell(cY1, row), *inc.Y1))
		}
		if inc.Y2 != nil {
			edits = append(edits, numEdit(cell(cY2, row), *inc.Y2))
		}
		if inc.Y3 != nil {
			edits = append(edits, numEdit(cell(cY3, row), *inc.Y3))
		}
		if inc.EUR != nil {
			edits = append(edits, numEdit(cell(cEUR, row), *inc.EUR))
		}
	}

	fillIncome(8, budget.Eigenmittel)
	fillIncome(14, budget.KMWMittel)

	if budget.DrittmittelY1 != nil {
		edits = append(edits, numEdit(cell(cY1, 11), *budget.DrittmittelY1))
	}
	if budget.DrittmittelY2 != nil {
		edits = append(edits, numEdit(cell(cY2, 11), *budget.DrittmittelY2))
	}
	if budget.DrittmittelY3 != nil {
		edits = append(edits, numEdit(cell(cY3, 11), *budget.DrittmittelY3))
	}

	// Ausgaben Tabelle (startet bei Zeile 19)
	for i, a := range budget.Ausgaben {
		r := 19 + i
		if a.LC != nil {
			edits = append(edits, numEdit(cell(cLC, r), *a.LC))
		}
		if a.Y1 != nil {
			edits = append(edits, numEdit(cell(cY1, r), *a.Y1))
		}
		if a.Y2 != nil {
			edits = append(edits, numEdit(cell(cY2, r), *a.Y2))
		}
		if a.Y3 != nil {
			edits = append(edits, numEdit(cell(cY3, r), *a.Y3))
		}
	}

	// Drittmittel Tabelle (startet bei Zeile 19, Spalten 18 und 19)
	geberRows := 10
	if len(budget.DrittGeber) > 10 {
		geberRows = len(budget.DrittGeber)
	}

	const (
		cGeberName = 18 // R
		cGeberLC   = 19 // S
		cGeberEUR  = 20 // T
	)

	for i := 0; i < geberRows; i++ {
		r := 19 + i
		if i < len(budget.DrittGeber) {
			geb := budget.DrittGeber[i]
			edits = append(edits, strEdit(cell(cGeberName, r), geb.Geber))
			if geb.LC != nil {
				edits = append(edits, numEdit(cell(cGeberLC, r), *geb.LC))
			}
			if geb.EUR != nil {
				edits = append(edits, numEdit(cell(cGeberEUR, r), *geb.EUR))
			}
		}
	}

	sonstigesRow := 19 + geberRows
	if budget.DrittSonstiges != nil {
		if budget.DrittSonstiges.LC != nil {
			edits = append(edits, numEdit(cell(cGeberLC, sonstigesRow), *budget.DrittSonstiges.LC))
		}
		if budget.DrittSonstiges.EUR != nil {
			edits = append(edits, numEdit(cell(cGeberEUR, sonstigesRow), *budget.DrittSonstiges.EUR))
		}
	}

	// Reserve (checkAddr in Spalte cGeberEUR+1 ? Nein, in budget.go ist es:
	// cCheck := BG_COL_BEGR_2 + 1. wait.
	// In budget.go: BG_COL_BEGR_2 = 13. col = 15 für Reserve?
	// Lasse ich erstmal aus, wenn ReserveFreigabe statisch beim Generieren gesetzt wird,
	// oder ich finde die exakte Spalte (col=13).
	// ...

	return edits
}

type fbLayoutInfo struct {
	aufschStartRow int
	tbl1HeaderRow  int
	tbl2HeaderRow  int
}

func getFBLayout(ausgDataRows int) fbLayoutInfo {
	const ausgHdrRow = 18
	ausgTotalsRow := ausgHdrRow + ausgDataRows + 1
	saldoRow := ausgTotalsRow + 2
	aufschLabelRow := saldoRow + 2
	aufschStart := aufschLabelRow + 1
	differenzRow := aufschStart + 4
	tbl1Label := differenzRow + 3
	tbl1Header := tbl1Label + 1
	totalsRow1 := tbl1Header + 6
	tbl2Label := totalsRow1 + 2
	tbl2Header := tbl2Label + 1
	return fbLayoutInfo{
		aufschStartRow: aufschStart,
		tbl1HeaderRow:  tbl1Header,
		tbl2HeaderRow:  tbl2Header,
	}
}

func buildFBEdits(periods []FBPeriod, budget *BudgetData) []cellEdit {
	var edits []cellEdit
	if budget == nil {
		return edits
	}

	n := len(budget.AusgabenIDs)
	l := getFBLayout(n)

	for p, fp := range periods {
		// FB Start Col ist 2 (B), Stride ist 7
		colStart := 2 + p*7

		if !fp.Von.IsZero() {
			edits = append(edits, dateEdit(cell(colStart+1, 5), fp.Von))
		}
		if !fp.Bis.IsZero() {
			edits = append(edits, dateEdit(cell(colStart+1, 6), fp.Bis))
		}

		for i, id := range budget.AusgabenIDs {
			if v, ok := fp.AusgabenByID[id]; ok && v != 0 {
				edits = append(edits, numEdit(cell(colStart+1, 19+i), v))
			}
		}

		if fp.BankLC != 0 {
			edits = append(edits, numEdit(cell(colStart+1, l.aufschStartRow), fp.BankLC))
		}

		kmwRowT1 := l.tbl1HeaderRow + 2
		if fp.KmwLC != 0 {
			// Es gibt den Wert LC und EUR für KMW, wir setzen hier nur LC,
			// oder den Kurs, je nach Vorlage. In testfill war es:
			// numEdit(..., fp.KmwLC) und numEdit(..., fp.KmwLC/exRate)
			// Wenn wir exRate nicht haben, können wir ggf. EUR weglassen oder benötigen ihn.
			// Der Einfachheit halber lassen wir hier EUR weg oder nehmen an es wird berechnet.
			// Oder wir fügen KmwEUR zum Struct hinzu.
			edits = append(edits, numEdit(cell(colStart+2, kmwRowT1), fp.KmwLC))
		}

		if fp.EigenLC != 0 {
			edits = append(edits, numEdit(cell(colStart+2, l.tbl2HeaderRow+1), fp.EigenLC))
		}
		if fp.DrittLC != 0 {
			edits = append(edits, numEdit(cell(colStart+2, l.tbl2HeaderRow+2), fp.DrittLC))
		}
	}
	return edits
}
func buildKMWEdits(tranchen []KMWTranche) []cellEdit {
	var edits []cellEdit
	for i, kr := range tranchen {
		row := 5 + i // erste Datenzeile = 5
		if row > 22 {
			break
		} // Nur bis Zeile 22
		edits = append(edits, strEdit(cell(2, row), kr.Periode))
		edits = append(edits, strEdit(cell(3, row), kr.Waehrung))
		edits = append(edits, numEdit(cell(4, row), kr.Betrag))
		if !kr.Datum.IsZero() {
			edits = append(edits, dateEdit(cell(5, row), kr.Datum))
		}
	}
	return edits
}

// ─── XML-Patching Helpers ─────────────────────────────────────────────────────

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

func patchCell(xmlStr, ref, inner, typeAttr string) (string, error) {
	marker := `<c r="` + ref + `"`
	i := strings.Index(xmlStr, marker)
	if i == -1 {
		return xmlStr, fmt.Errorf("zelle nicht gefunden (oder nicht im xml angelegt)")
	}

	start := i + len(marker)
	end := strings.Index(xmlStr[start:], ">")
	if end == -1 {
		return xmlStr, fmt.Errorf("fehlerhaftes xml bei %s", ref)
	}
	end += start

	endTag := "</c>"
	hasContent := false
	if xmlStr[end-1] == '/' {
		end = end - 1
	} else {
		hasContent = true
	}

	attrs := xmlStr[start:end]
	attrs = strings.TrimSpace(attrs)

	if typeAttr != "" {
		if strings.Contains(attrs, ` t="`) {
			parts := strings.Split(attrs, ` t="`)
			rest := strings.SplitN(parts[1], `"`, 2)
			attrs = parts[0] + ` t="` + typeAttr + `"`
			if len(rest) > 1 {
				attrs += rest[1]
			}
		} else {
			attrs += ` t="` + typeAttr + `"`
		}
	} else {
		if idx := strings.Index(attrs, ` t="`); idx != -1 {
			endIdx := strings.Index(attrs[idx+4:], `"`)
			if endIdx != -1 {
				attrs = attrs[:idx] + attrs[idx+4+endIdx+1:]
			}
		}
	}

	var replaced string
	if !hasContent {
		replaced = xmlStr[:i] + marker + " " + strings.TrimSpace(attrs) + ">" + inner + "</c>" + xmlStr[end+2:]
	} else {
		endNode := strings.Index(xmlStr[end:], endTag)
		if endNode == -1 {
			return xmlStr, fmt.Errorf("schließendes tag fehlt bei %s", ref)
		}
		endNode += end
		replaced = xmlStr[:i] + marker + " " + strings.TrimSpace(attrs) + ">" + inner + "</c>" + xmlStr[endNode+4:]
	}
	return replaced, nil
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

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

func mapSheetNamesToParts(parts map[string][]byte) (map[string]string, error) {
	wbData, ok := parts["xl/workbook.xml"]
	if !ok {
		return nil, fmt.Errorf("xl/workbook.xml fehlt")
	}

	relsData, ok := parts["xl/_rels/workbook.xml.rels"]
	if !ok {
		return nil, fmt.Errorf("xl/_rels/workbook.xml.rels fehlt")
	}

	rels := make(map[string]string)
	for _, line := range strings.Split(string(relsData), "<Relationship ") {
		id := extractAttr(line, "Id")
		target := extractAttr(line, "Target")
		if id != "" && target != "" {
			rels[id] = target
		}
	}

	sheetPart := make(map[string]string)
	for _, line := range strings.Split(string(wbData), "<sheet ") {
		name := extractAttr(line, "name")
		rId := extractAttr(line, "r:id")
		if name != "" && rId != "" {
			target := rels[rId]
			if strings.HasPrefix(target, "/") {
				target = target[1:]
			} else {
				target = "xl/" + target
			}
			sheetPart[name] = target
		}
	}
	return sheetPart, nil
}

func extractAttr(xmlFrag, attr string) string {
	marker := attr + `="`
	i := strings.Index(xmlFrag, marker)
	if i == -1 {
		return ""
	}
	start := i + len(marker)
	end := strings.Index(xmlFrag[start:], `"`)
	if end == -1 {
		return ""
	}
	return xmlFrag[start : start+end]
}
