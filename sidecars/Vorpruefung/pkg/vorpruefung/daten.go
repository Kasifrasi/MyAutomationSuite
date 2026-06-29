package vorpruefung

import (
	"fmt"
	"shared/constants"
	"strings"
)

// CreateDatenSheet erstellt ein verstecktes Blatt "Daten" für Dropdowns und VSTACK-Zusammenfassungen.
func (g *Generator) CreateDatenSheet() error {
	ws := constants.VPSheetDATEN
	f := g.file

	_, err := f.NewSheet(ws)
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Daten-Blatts: %w", err)
	}

	// Das Blatt verstecken
	_ = f.SetSheetVisible(ws, false)

	// 1. Perioden-Liste für Dropdowns generieren (Spalte A)
	for i := 1; i <= MA_PERIOD_COUNT; i++ {
		_ = f.SetCellValue(ws, fmt.Sprintf("A%d", i), fmt.Sprintf("Periode %d", i))
	}

	// 2. VSTACKs der Tabellen

	// Einnahmen 1 (Explizit) - Spalte C
	_ = f.SetCellValue(ws, "C1", "Einnahmen_Explizit_Stack")
	if len(g.rangesEinnahmen1) > 0 {
		_ = g.setDynArrayFormula(ws, "AA2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesEinnahmen1, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "C2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AA2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AA2),3)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AA2),4)<>0), "")`, StyleOptions{})
	}

	// Einnahmen 2 (Durchschnittskurs) - Spalte I
	_ = f.SetCellValue(ws, "I1", "Einnahmen_Durchschnittskurs_Stack")
	if len(g.rangesEinnahmen2) > 0 {
		_ = g.setDynArrayFormula(ws, "AG2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesEinnahmen2, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "I2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AG2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AG2),3)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AG2),4)<>0), "")`, StyleOptions{})
	}

	// Ausgaben (Finanzbericht) - Spalte O
	_ = f.SetCellValue(ws, "O1", "Ausgaben_Finanzbericht_Stack")
	if len(g.rangesAusgaben) > 0 {
		_ = g.setDynArrayFormula(ws, "AM2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesAusgaben, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "O2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AM2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AM2),2)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AM2),3)<>0), "")`, StyleOptions{})
	}

	// Mittelanforderung (MA) - Spalte U
	_ = f.SetCellValue(ws, "U1", "Mittelanforderung_Stack")
	if len(g.rangesMA) > 0 {
		_ = g.setDynArrayFormula(ws, "AS2", fmt.Sprintf(`_xlfn.VSTACK(%s)`, strings.Join(g.rangesMA, ",")), StyleOptions{})
		_ = g.setDynArrayFormula(ws, "U2", `_xlfn._xlws.FILTER(_xlfn.ANCHORARRAY(AS2), (_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AS2),2)<>0)+(_xlfn.CHOOSECOLS(_xlfn.ANCHORARRAY(AS2),3)<>0), "")`, StyleOptions{})
	}

	// Kopfzeilen formatieren
	headerOpts := StyleOptions{
		Bold:      true,
		FillColor: "D3D3D3", // COLOR_HEADER
	}
	_ = g.setStyle(ws, "C1", "C1", headerOpts)
	_ = g.setStyle(ws, "I1", "I1", headerOpts)
	_ = g.setStyle(ws, "O1", "O1", headerOpts)
	_ = g.setStyle(ws, "U1", "U1", headerOpts)

	// Helfer-Strukturen für die Auswertung (Auswahllisten, MA-Grid)
	g.evalBuildDatenHelfer(ws)

	return nil
}

// evalBuildDatenHelfer baut die für das Auswertungsblatt benötigten, versteckten
// Hilfsstrukturen:
//   - MA-Meta (je MA-Tabelle: Periode, befüllt?, Rang, Label, Summen, Eigen/Dritt)
//   - FB-Meta (je Periode: befüllt?, Label)
//   - FILTER-Auswahllisten (MA_Auswahl_Liste, FB_Auswahl_Liste)
//   - MA-Grid (je MA-Tabelle × Kategorie: Periode, Rang, Kategorie, LC, EUR)
func (g *Generator) evalBuildDatenHelfer(ws string) {
	f := g.file
	maSheet := constants.VPSheetMA
	fbSheet := constants.VPSheetFINANZBERICHTE

	dc := func(col, row int) string { return cellName(col, row) }

	maDataRows := len(MA_CATEGORIES)
	maBlockHeight := 10 + maDataRows + 8 // same calculation as in MA sheet
	maTotalsRow := 9 + maDataRows + 1

	rowKmw := maTotalsRow + 7
	rowManuell := maTotalsRow + 9
	rowEig := maTotalsRow + 3
	rowDritt := maTotalsRow + 4

	// ─── MA-Meta (Zeilen 1..54) ───────────────────────────────────────────────
	for j := 1; j <= MA_TABLE_COUNT; j++ {
		p := ((j - 1) % MA_PERIOD_COUNT) + 1
		level := ((j - 1) / MA_PERIOD_COUNT) + 1
		offsetR := (level - 1) * (maBlockHeight + 2)

		colS := 2 + (p-1)*4
		perCell := fmt.Sprintf("'%s'!%s", maSheet, cellName(colS+1, 4+offsetR)) // Periode-Kopf

		_ = f.SetCellValue(ws, dc(EV_DTN_MA_META_J, j), j)
		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_PER, j),
			fmt.Sprintf(`=IFERROR(VALUE(TRIM(MID(%s,9,5))),0)`, perCell))

		kmwCellLC := FieldMAKmwLC(j).NamedRange
		kmwCellEUR := FieldMAKmwEUR(j).NamedRange
		manCellEUR := FieldMAManBetrag(j).NamedRange

		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_FILL, j),
			fmt.Sprintf(`=IF(OR(IFERROR(%s,0)>0, IFERROR(%s,0)>0, IFERROR(%s,0)>0),1,0)`,
				kmwCellLC, kmwCellEUR, manCellEUR))

		// EV_DTN_MA_META_RANK hält nun fest den "Slot-Index" (1, 2 oder 3),
		// unabängig davon, ob die MA ausgefüllt ist. Das ist essentiell, damit
		// in pruefung_ma.go die Index-Berechnung (Periode P, Rang K -> Index J)
		// fehlerfrei funktioniert, selbst wenn "Ausfüllungslücken" entstehen.
		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_RANK, j),
			fmt.Sprintf(`=%d`, level))
		var labelFormula string
		if j <= MA_PERIOD_COUNT {
			// Für den ersten Slot jeder Periode (j <= MA_PERIOD_COUNT) immer ein Label vergeben,
			// auch wenn die MA noch nicht befüllt ist (dann "Periode X (#1)").
			labelFormula = fmt.Sprintf(`=IF(%s>0, "Periode "&%s&" (#1)", "")`,
				dc(EV_DTN_MA_META_PER, j), dc(EV_DTN_MA_META_PER, j))
		} else {
			// Für weitere Slots (MA #2, MA #3, etc.) nur ein Label vergeben, wenn sie befüllt sind.
			labelFormula = fmt.Sprintf(`=IF(AND(%s=1,%s>0),"Periode "&%s&" (#%d)","")`,
				dc(EV_DTN_MA_META_FILL, j), dc(EV_DTN_MA_META_PER, j), dc(EV_DTN_MA_META_PER, j), level)
		}
		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_LABEL, j), labelFormula)
		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_SUMLC, j),
			fmt.Sprintf(`=IFERROR(ROUND(SUM('%s'!%s:%s),2),0)`, maSheet, cellName(colS+1, 10+offsetR), cellName(colS+1, 9+maDataRows+offsetR)))
		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_SUMEU, j),
			fmt.Sprintf(`=IFERROR(ROUND(SUM('%s'!%s:%s),2),0)`, maSheet, cellName(colS+2, 10+offsetR), cellName(colS+2, 9+maDataRows+offsetR)))
		eigCellEUR := FieldMAEigenmittelEUR(j).NamedRange
		drittCellEUR := FieldMADrittmittelEUR(j).NamedRange

		_ = f.SetCellFormula(ws, dc(EV_DTN_MA_META_EIGDR, j),
			fmt.Sprintf(`=IFERROR(ROUND(%s+%s,2),0)`, eigCellEUR, drittCellEUR))
	}

	// ─── FB-Meta (Zeilen 1..18) ───────────────────────────────────────────────
	// "Befüllt" = es kam bei den Ausgaben ODER den laufenden Einnahmen (Typzeilen
	// 12..15, ohne Vorperiodensaldo) zu Eingaben ODER in den gelben Eingabefeldern
	// der Saldenaufschlüsselung (Bank, Kasse, Sonstiges) wurde etwas eingetragen
	// ODER das "Von" / "Bis" Datum wurde in Zeile 5 / 6 ausgefüllt.
	for p := 1; p <= MA_PERIOD_COUNT; p++ {
		ausgName := fmt.Sprintf("Ausgaben_%d", p)
		incCol := colLetter(3 + (p-1)*7) // laufende Einnahmen (LC) je FB-Periode

		aufschlBank := FieldFBAufschlBank(p).NamedRange
		aufschlKasse := FieldFBAufschlKasse(p).NamedRange
		aufschlSonstiges := FieldFBAufschlSonstiges(p).NamedRange

		_ = f.SetCellValue(ws, dc(EV_DTN_FB_META_PER, p), p)

		fillFormula := fmt.Sprintf(`=IF((IFERROR(SUBTOTAL(109,%s[Ausgaben (LC)]),0)<>0)+(IFERROR(SUM('%s'!%s12:%s15),0)<>0)+(IFERROR(COUNT(%s, %s, %s),0)>0)+('%s'!%s5<>"")+('%s'!%s6<>"")>0,1,0)`,
			ausgName, fbSheet, incCol, incCol, aufschlBank, aufschlKasse, aufschlSonstiges, fbSheet, incCol, fbSheet, incCol)
		_ = f.SetCellFormula(ws, dc(EV_DTN_FB_META_FILL, p), fillFormula)

		_ = f.SetCellFormula(ws, dc(EV_DTN_FB_META_LABEL, p),
			fmt.Sprintf(`=IF(%s=1,"Periode "&%s,"")`, dc(EV_DTN_FB_META_FILL, p), dc(EV_DTN_FB_META_PER, p)))
	}

	// ─── FB-Auswahlliste (FILTER der befüllten Perioden) ──────────────────────
	fbLabelRng := fmt.Sprintf("$%s$1:$%s$%d", colLetter(EV_DTN_FB_META_LABEL), colLetter(EV_DTN_FB_META_LABEL), MA_PERIOD_COUNT)
	fbFillRng := fmt.Sprintf("$%s$1:$%s$%d", colLetter(EV_DTN_FB_META_FILL), colLetter(EV_DTN_FB_META_FILL), MA_PERIOD_COUNT)
	// Reihenfolge wie die Perioden (aufsteigend):
	// befüllten Perioden, und der Auto-Eintrag "Neuester FB" ganz unten (= jüngster).
	_ = g.setDynArrayFormula(ws, dc(EV_DTN_FB_LISTE, 1),
		fmt.Sprintf(`_xlfn.VSTACK(_xlfn._xlws.FILTER(%s,%s=1,""),"Neuester FB")`, fbLabelRng, fbFillRng), StyleOptions{})
	g.upsertNamedFormula(EVAL_NAME_FB_LISTE,
		fmt.Sprintf(`OFFSET('%s'!%s,0,0,COUNTA('%s'!$%s:$%s),1)`,
			ws, absName(EV_DTN_FB_LISTE, 1), ws, colLetter(EV_DTN_FB_LISTE), colLetter(EV_DTN_FB_LISTE)))

	// ─── MA-Auswahlliste (Bis zu Perioden von Max(FB) + 1, auch wenn diese noch leer sind) ───────
	maLabelRng := fmt.Sprintf("$%s$1:$%s$%d", colLetter(EV_DTN_MA_META_LABEL), colLetter(EV_DTN_MA_META_LABEL), MA_TABLE_COUNT)
	maPerRng := fmt.Sprintf("$%s$1:$%s$%d", colLetter(EV_DTN_MA_META_PER), colLetter(EV_DTN_MA_META_PER), MA_TABLE_COUNT)

	maxFb := fmt.Sprintf(`IFERROR(SUMPRODUCT(MAX(($%s$1:$%s$%d=1)*$%s$1:$%s$%d)),0)`,
		colLetter(EV_DTN_FB_META_FILL), colLetter(EV_DTN_FB_META_FILL), MA_PERIOD_COUNT,
		colLetter(EV_DTN_FB_META_PER), colLetter(EV_DTN_FB_META_PER), MA_PERIOD_COUNT)

	maCond := fmt.Sprintf(`(%s<>"")*(%s<=%s+1)>0`, maLabelRng, maPerRng, maxFb)

	// Alle entsprechenden Anforderungen auflisten ("Periode X (#k)"), plus "Neueste MA".
	_ = g.setDynArrayFormula(ws, dc(EV_DTN_MA_LISTE, 1),
		fmt.Sprintf(`_xlfn.VSTACK(_xlfn._xlws.FILTER(%s,%s,""),"Neueste MA")`, maLabelRng, maCond), StyleOptions{})
	g.upsertNamedFormula(EVAL_NAME_MA_LISTE,
		fmt.Sprintf(`OFFSET('%s'!%s,0,0,COUNTA('%s'!$%s:$%s),1)`,
			ws, absName(EV_DTN_MA_LISTE, 1), ws, colLetter(EV_DTN_MA_LISTE), colLetter(EV_DTN_MA_LISTE)))

	// ─── MA-Grid (je MA-Tabelle × Kategorie/Finanzierungsart) ─────────────────
	// Pro MA-Tabelle ein Block aus 8 Kostenkategorien (MA-Zeilen 10..17) PLUS den
	// 3 Finanzierungsarten der Prognose (Eigenmittel, Drittmittel, KMW-Mittel auf
	// den MA-Zeilen 21/22/25). Die CAT-Beschriftung der Finanzierungsarten ist
	// identisch zu den FB-Einnahmentypen (TYPE_NAMES), sodass die Prognose der
	// Finanzierungsanteile dieselbe Grid-SUMIFS nutzt wie die Ausgabenprognose.
	// "Zinsertraege" hat keine MA-Quelle ⇒ keine Grid-Zeile ⇒ SUMIFS ergibt 0.
	type maGridEntry struct {
		cat   string
		maRow int
	}

	gridEntries := make([]maGridEntry, 0, maDataRows+4)
	for c := 0; c < maDataRows; c++ {
		// Category labels are on the MA sheet. For now we will just use indices for categories if needed,
		// or read the actual label from the MA sheet?
		// Since daten.go is on a different sheet, we could just reference the label cell, but daten sheet
		// SUMIFS requires plain strings if we use "KMW-Mittel".
		// We'll leave the cat string empty or placeholder and fill it dynamically or let the API do it?
		// Wait, daten.go populates the daten sheet. The API won't touch the daten sheet!
		// But in pruefung_fb.go and pruefung_ma.go, the SUMIFS matches this value.
		// So we can just use an index string like "Expense_0" for the category in daten.go,
		// and in pruefung_fb.go/pruefung_ma.go we query for "Expense_0".
		gridEntries = append(gridEntries, maGridEntry{fmt.Sprintf("Expense_%d", c), 10 + c})
	}

	// The income types (KMW-Mittel etc.) are still hardcoded in the comparison logic?
	// The user says "ohne vordefinierte Typen ... Einnahme-Typen".
	// But in MA, they are literally "abzueglich Eigenmittel".
	gridEntries = append(gridEntries,
		maGridEntry{"Eigenmittel", rowEig},
		maGridEntry{"Drittmittel", rowDritt},
		maGridEntry{"KMW-Mittel", rowKmw},
		maGridEntry{"Manueller Betrag", rowManuell},
	)
	blockSize := len(gridEntries)
	for j := 1; j <= MA_TABLE_COUNT; j++ {
		p := ((j - 1) % MA_PERIOD_COUNT) + 1
		level := ((j - 1) / MA_PERIOD_COUNT) + 1
		offsetR := (level - 1) * (maBlockHeight + 2)
		colS := 2 + (p-1)*4
		for idx, e := range gridEntries {
			row := (j-1)*blockSize + idx + 1
			_ = f.SetCellFormula(ws, dc(EV_DTN_MAG_PER, row), fmt.Sprintf("=$%s$%d", colLetter(EV_DTN_MA_META_PER), j))
			_ = f.SetCellFormula(ws, dc(EV_DTN_MAG_RANK, row), fmt.Sprintf("=$%s$%d", colLetter(EV_DTN_MA_META_RANK), j))
			_ = f.SetCellValue(ws, dc(EV_DTN_MAG_CAT, row), e.cat)

			var maValLC, maValEUR string
			if e.cat == "Eigenmittel" {
				maValLC = FieldMAEigenmittelLC(j).NamedRange
				maValEUR = FieldMAEigenmittelEUR(j).NamedRange
			} else if e.cat == "Drittmittel" {
				maValLC = FieldMADrittmittelLC(j).NamedRange
				maValEUR = FieldMADrittmittelEUR(j).NamedRange
			} else if e.cat == "KMW-Mittel" {
				maValLC = FieldMAKmwLC(j).NamedRange
				maValEUR = FieldMAKmwEUR(j).NamedRange
			} else if e.cat == "Manueller Betrag" {
				// Manueller Betrag hat kein LC-Feld, wir nehmen 0
				maValLC = "0"
				maValEUR = FieldMAManBetrag(j).NamedRange
			} else if strings.HasPrefix(e.cat, "Expense_") {
				var idx int
				fmt.Sscanf(e.cat, "Expense_%d", &idx)
				maValLC = FieldMAKat(j, idx+1).NamedRange
				maValEUR = fmt.Sprintf("'%s'!%s", maSheet, cellName(colS+2, e.maRow+offsetR)) // EUR field relies on formula
			}

			_ = f.SetCellFormula(ws, dc(EV_DTN_MAG_LC, row), fmt.Sprintf(`=IFERROR(%s,0)`, maValLC))
			_ = f.SetCellFormula(ws, dc(EV_DTN_MAG_EUR, row), fmt.Sprintf(`=IFERROR(%s,0)`, maValEUR))
		}
	}

	// Hilfsspalten ausblenden
	for col := EV_DTN_MA_META_J; col <= EV_DTN_MAG_EUR; col++ {
		_ = f.SetColVisible(ws, colLetter(col), false)
	}
}
