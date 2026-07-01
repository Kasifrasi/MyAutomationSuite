# Architektur-Richtlinien für Excel-Sheets

Dieses Dokument definiert den standardisierten Aufbau für Go-Dateien, die Excel-Sheets generieren (z. B. `dashboard.go`, `budget.go`, `ma.go`). 
Ziel dieser Architektur ist **maximale Wartbarkeit, Lesbarkeit und eine saubere Trennung zwischen Layout (Design) und Logik (Datenbindung).**

---

## 1. Verwaltung von Styles (`styles.go`)

Alle Zell-Formatierungen (`StyleOptions`) werden zentral in einer `styles.go` Datei gesammelt. 

**Regel:** Styles werden strikt **pro Sheet gruppiert**. 
* Es gibt z. B. `DashLabelStyle`, `DashInputStyle`, `BudgetLabelStyle`, `BudgetInputStyle`.
* **Mehrfachdefinitionen sind ausdrücklich erlaubt und erwünscht!** Wenn das Dashboard und das Budget exakt gleich aussehende Eingabefelder haben, bekommt trotzdem jedes Sheet seinen eigenen Style in der Definition. Das verhindert, dass eine nachträgliche Design-Änderung im Budget versehentlich das Dashboard zerschießt.

---

## 2. Struktur einer Sheet-Datei (z. B. `dashboard.go`)

Jede Datei, die ein Sheet zeichnet, muss einem festen Aufbau folgen:

### Teil A: Layout-Konstanten (Grid-System)
Ganz oben in der Datei werden die Zeilen (`Row...`) und Spalten (`Col...`) als Konstanten definiert. Es dürfen **keine** Magic Numbers (wie `r := 15` oder `col := 3`) mitten im Code auftauchen.

```go
const (
    // Spalten-Layout
    DashColLabelLeft  = 2 // B
    DashColInputLeft  = 3 // C
    DashColLabelRight = 4 // D
    DashColInputRight = 5 // E

    // Zeilen-Layout
    DashRowHeader         = 2
    DashRowTitle          = 4
    DashRowProjektNummer  = 5
    // ...
)
```

### Teil B: Layout-Dokumentation (Kommentar oder Markdown)
Für komplexe Sheets sollte über den Konstanten eine kleine visuelle Tabelle als Kommentar stehen, damit Entwickler sofort verstehen, wie das Sheet aufgebaut ist, ohne den Code lesen zu müssen:

```go
/*
  LAYOUT DASHBOARD:
  | Zeile | Spalte B (Label)   | Spalte C (Input 1)  | Spalte D (Label)   | Spalte E (Input 2)  |
  |-------|--------------------|---------------------|--------------------|---------------------|
  | 5     | Projektnummer      | [Inp_Projektnummer] | Vorprojekt?        | [Inp_Vorprojekt]    |
  | 6     | Projekttitel       | [Inp_Projekttitel (merged C-E)]                              |
*/
```

### Teil C: Die Orchestrator-Funktion
Jedes Sheet hat eine Hauptfunktion (z. B. `CreateDashboardSheet`), die den Ablauf steuert. Hier wird **nicht** direkt gezeichnet.

```go
func (g *Generator) CreateDashboardSheet(reg *TemplateRegistry) error {
    ws := constants.VPSheetDASHBOARD
    
    // 1. Initialisierung & Spaltenbreiten
    g.initSheet(ws)
    g.setupDashColumns(ws)

    // 2. Zeichnen (Layout, Farben, Ränder) - rein visuell!
    g.drawDashStaticInfo(ws)
    
    // 3. Binden (Daten, Validierungen, Named Ranges aus Registry)
    g.bindDashFields(ws, reg)

    return nil
}
```

### Teil D: Draw-Funktionen (Nur visuell)
Hier werden die Zellen eingefärbt und Texte (Labels) gesetzt. Es werden ausschließlich die ausgelagerten Styles aus `styles.go` und die Grid-Konstanten von oben verwendet.
**Wichtig:** Hier werden noch keine `Named Ranges` (InputFields) verknüpft!

```go
func (g *Generator) drawDashStaticInfo(ws string) {
    g.setValue(ws, cellName(DashColLabelLeft, DashRowProjektNummer), "Projektnummer", DashLabelStyle)
    g.setValue(ws, cellName(DashColInputLeft, DashRowProjektNummer), "", DashInputStyle)
}
```

### Teil E: Bind-Funktionen (Logik & Registry)
Hier wird das visuelle Konstrukt "lebendig" gemacht. Wir verbinden die gezeichneten Felder über die `TemplateRegistry` mit dem Namensmanager von Excel.
Hier passieren auch Dinge wie Formeln und Datenüberprüfungen (Dropdowns).

```go
func (g *Generator) bindDashFields(ws string, reg *TemplateRegistry) {
    // Projektnummer
    g.bindInputField(ws, DashRowProjektNummer, DashColInputLeft, reg.InputDashProjektnummer.Name)
    
    // Vorprojekt (Dropdown)
    g.applyDataValidation(ws, DashRowProjektNummer, DashColInputRight, reg.InputDashVorprojekt.Validation)
    g.bindInputField(ws, DashRowProjektNummer, DashColInputRight, reg.InputDashVorprojekt.Name)
}
```

---

## 3. Goldene Regeln für den Workflow

1. **Änderungen am Layout:**
   Will ein Nutzer z.B. das Feld "Projektstart" eine Zeile tiefer setzen, ändert der Entwickler **nur** die Konstante `DashRowProjektStart = 8` zu `DashRowProjektStart = 9`. Der restliche Code bleibt unangetastet.
2. **Registry First:**
   Bevor ein neues Eingabe- oder Ausgabefeld in einem Sheet gezeichnet wird, muss es zwingend zuerst in `registry.go` (im `TemplateRegistry` Struct) definiert werden.
3. **Keine harten Strings für Named Ranges:**
   Strings wie `"Inp_Projektnummer"` dürfen in den Sheet-Dateien (`dashboard.go` etc.) nicht existieren. Sie kommen ausschließlich über `reg.InputDashProjektnummer.Name`.
