package vorpruefung

import (
	"fmt"
	"shared/constants"
)

const (
	EVAL_TAB_COLOR    = "FF0000" // Rot
	EVAL_DATEN_SHEET  = constants.VPSheetDATEN
	EVAL_FB_SHEET     = constants.VPSheetFINANZBERICHTE
	EVAL_STACK_MAXROW = 500

	// Spalten der Vergleichstabellen (B … J)
	EV_COL_LABEL   = 2
	EV_COL_ACT_LC  = 3
	EV_COL_BUD_LC  = 4
	EV_COL_DIF_LC  = 5
	EV_COL_ABW_LC  = 6
	EV_COL_ACT_EUR = 7
	EV_COL_BUD_EUR = 8
	EV_COL_DIF_EUR = 9
	EV_COL_ABW_EUR = 10

	// Auswahl-Panel (zentriert, oben in der Sektion; Spalten C … G)
	EV_PB_C1   = 3 // Box-/Label-Start (C)
	EV_PB_L2   = 4 // Label-Ende (D)
	EV_PB_V1   = 5 // Wert-/Slot-LC-Start (E)
	EV_PB_SLC2 = 6 // Slot-LC-Ende (F)
	EV_PB_SEU1 = 7 // Slot-EUR-Start (G)
	EV_PB_C2   = 7 // Box-/Wert-Ende (G)

	EV_TABLE_GAP = 2
	EV_MA_SLOTS  = 6 // max. gleichzeitig anzeigbare Anforderungen je Periode

	// Daten-Helfer (Spaltennummern auf dem Blatt "Daten")
	EV_DTN_MA_META_J     = 53 // BA  Tabellenindex j
	EV_DTN_MA_META_PER   = 54 // BB  Periode von MA_j
	EV_DTN_MA_META_FILL  = 55 // BC  befüllt?
	EV_DTN_MA_META_RANK  = 56 // BD  Rang innerhalb der Periode
	EV_DTN_MA_META_LABEL = 57 // BE  "Periode X (#k)"
	EV_DTN_MA_META_SUMLC = 58 // BF  Summe Angefordert (LC)
	EV_DTN_MA_META_SUMEU = 59 // BG  Summe Angefordert (EUR)
	EV_DTN_MA_META_EIGDR = 60 // BH  Eigen+Dritt (EUR)
	EV_DTN_FB_META_PER   = 62 // BJ  Periode
	EV_DTN_FB_META_FILL  = 63 // BK  befüllt?
	EV_DTN_FB_META_LABEL = 64 // BL  "Periode X"
	EV_DTN_MA_LISTE      = 66 // BN  FILTER-Spill (Auswahlliste MA)
	EV_DTN_FB_LISTE      = 68 // BP  FILTER-Spill (Auswahlliste FB)
	EV_DTN_MAG_PER       = 70 // BR  Grid: Periode
	EV_DTN_MAG_RANK      = 71 // BS  Grid: Rang
	EV_DTN_MAG_CAT       = 72 // BT  Grid: Kategorie
	EV_DTN_MAG_LC        = 73 // BU  Grid: LC
	EV_DTN_MAG_EUR       = 74 // BV  Grid: EUR
	// Grid-Block je MA-Tabelle: 8 Kostenkategorien (len(MA_CATEGORIES)) + 4
	// Finanzierungsarten (Eigenmittel/Drittmittel/KMW-Mittel/Manueller Betrag) für die Prognose
	// der Finanzierungsanteile. Muss zu gridEntries in daten.go passen.
	EV_DTN_MAG_BLOCK = 8 + 4
	EV_DTN_MAG_ROWS  = MA_TABLE_COUNT * EV_DTN_MAG_BLOCK

	EVAL_NAME_MA_LISTE = "MA_Auswahl_Liste"
	EVAL_NAME_FB_LISTE = "FB_Auswahl_Liste"

	// Farben
	EV_CLR_BANNER     = "212F3D"
	EV_CLR_BANNER_TXT = "FFFFFF"
	EV_CLR_BANNER_SUB = "B4BEC8"
	EV_CLR_HEADER     = "D3D3D3"
	EV_CLR_TOTAL      = "212F3D"
	EV_CLR_TOTAL_TXT  = "FFFFFF"
	EV_CLR_INPUT      = "FFFAE5" // nur für bearbeitbare Felder
	EV_CLR_DEDUCT     = "EAF2F8" // abzuziehende (berechnete) Beträge
	EV_CLR_DEDUCT_OFF = "D9D9D9" // deaktivierte Abzugszelle ("Kein Abzug")
	EV_CLR_CALC       = "F2F2F2" // sonstige berechnete Felder
	EV_CLR_BORDER     = "808080"
	EV_CLR_GRID       = "D3D3D3"
	EV_CLR_BLACK      = "000000"
	EV_CLR_GOOD       = "C6EFCE"
	EV_CLR_GOOD_TXT   = "006100"
	EV_CLR_BAD        = "FFC7CE"
	EV_CLR_BAD_TXT    = "9C0006"
	EV_CLR_WARN       = "FCF3CF"
	EV_CLR_WARN_TXT   = "9C640C"
	EV_CLR_PANEL_REV  = "D6EAF8" // Hintergrund aktiver (eingeblendeter) Slots

	EV_FMT_LC  = "#,##0.00"
	EV_FMT_EUR = `#,##0.00" €"`
	EV_FMT_PCT = "0.0%"
)

// evalSelRefs bündelt die Adressen der Auswahl-Steuerzellen.
type evalSelRefs struct {
	fbSelNum string // Periodennummer des gewählten Finanzberichts (N)
	maSelP   string // Periode der gewählten Mittelanforderung (N+1)
	maSelK   string // Rang (#k) der gewählten Mittelanforderung
}

// evalAbsCol liefert einen absoluten Spaltenbereich, z. B. "$BU$1:$BU$144".
func evalAbsCol(col, r1, r2 int) string {
	return fmt.Sprintf("$%s$%d:$%s$%d", colLetter(col), r1, colLetter(col), r2)
}
