# Excel-Report Generator (Go Sidecar) - Integration Guide

**An die assistierende KI:**
Dieses Dokument beschreibt die Schnittstelle zur `scanner.exe` (einem in Go geschriebenen Headless-Sidecar). Deine Aufgabe ist es, diese `.exe` aus einer Rust-Anwendung (mit Slint GUI) heraus aufzurufen. 

Die Architektur folgt einer strikten Trennung ("Separation of Concerns"):
1. **Rust (GUI & Scanner)** liest Excel-Budgets ein, extrahiert die Daten und speichert sie als JSON-Array.
2. **Go (Sidecar)** nimmt dieses JSON-Array, füllt damit extrem performant (via Worker-Pool und RAM-Template-Injection) die Ziel-Excel-Berichte aus und streamt den Fortschritt als JSON via `stdout` zurück an Rust.

Es müssen **keine** separaten Template-Dateien ausgeliefert werden. Die `scanner.exe` enthält alle Excel-Vorlagen bereits als embedded Files (`//go:embed`).

---

## 1. Kommandozeilen-Schnittstelle (CLI)

Der Aufruf der ausführbaren Datei erfolgt mit drei Parametern:

```bash
scanner.exe -input <pfad_zur_json> -output <ziel_ordner> [-filename <muster>]
```

| Flag | Pflicht | Beschreibung |
| :--- | :---: | :--- |
| `-input` | **Ja** | Pfad zu der von Rust generierten JSON-Datei, die ein Array mit den zu verarbeitenden Daten enthält (siehe Schema unten). |
| `-output` | Nein | Zielordner für die generierten Excel-Berichte. Standard: `test/output`. Fehlende Ordner werden automatisch erstellt. |
| `-filename` | Nein | Namensmuster für die Ausgabedateien. Standard: `Report_{pn}_{la}.xlsx`. |

### Filename-Templating (`-filename`)
Das Namensmuster unterstützt folgende Platzhalter:
* `{pn}` -> Wird durch die Projektnummer ersetzt (z. B. `2026_0004_001`).
* `{la}` -> Wird durch die extrahierte Sprache ersetzt (z. B. `deutsch` oder `english`).
* `{i}` -> Ein intelligenter Zähler (1, 2, 3...). 
  * **WICHTIG:** Der Zähler `{i}` wird *nur* angewendet, wenn es im JSON-Input mehrere Berichte gibt, die ansonsten denselben Namen hätten (Kollisionsschutz). Bei einzigartigen Berichten wird das `{i}` (inklusive eventuell überflüssiger Unterstriche) unsichtbar entfernt.

---

## 2. Input: Das JSON-Datenformat

Die Go-Anwendung erwartet ein JSON-Array. Jedes Objekt im Array repräsentiert einen vollständigen Bericht (basierend auf dem Rust-Struct `ScannedBudgetData`).

**Beispiel `input.json`:**
```json
[
  {
    "file_path": "C:\\temp\\budget_de.xlsx",
    "sheet_name": "Budget",
    "version": "1.0",
    "project_title": "Mein tolles Projekt",
    "project_number": "PROJ-123",
    "language": "de",
    "local_currency": "EUR",
    "cost_col1": 8,
    "cost_col2": 13,
    "eigenleistung": "1000,50",
    "drittmittel": "0",
    "kmw_mittel": "5000",
    "positions": [
      {
        "number": "1.1",
        "label": "Personal",
        "cost_col1": "2000",
        "cost_col2": "1000"
      },
      {
        "number": "1.2",
        "label": "Reisekosten",
        "cost_col1": "500",
        "cost_col2": ""
      }
    ]
  }
]
```

**Datentyp-Hinweise für die Rust-Generierung:**
* Geldbeträge (`eigenleistung`, `cost_col1` in positions, etc.) werden in Go als **String** erwartet. Go parst diese Strings selbst in Floats. Sowohl deutsche `1.000,50` als auch englische `1000.50` Formate sind erlaubt. Wenn ein Betrag leer ist oder `-` enthält, wird er als `0` interpretiert.
* `cost_col1` und `cost_col2` auf Root-Ebene (für die Spalten-Indices) sind Integer. `cost_col2` ist optional (`null` bzw. `Option<i32>` in Rust).

---

## 3. Output: Das Progress-Streaming (Stdout)

Während die Go-Anwendung die Excel-Dateien generiert (was stark parallelisiert abläuft), schreibt sie für jeden Schritt eine JSON-Zeile in die Standardausgabe (`stdout`). Dies ermöglicht der Slint-GUI eine Echtzeit-Aktualisierung einer Progress-Bar.

**Das Schema jeder Output-Zeile:**
```json
{
  "status": "...",
  "file": "...",
  "current": 1,
  "total": 100,
  "message": "..."
}
```

### Die 4 Status-Typen:

1. **Start:** Wird exakt einmal zu Beginn gesendet.
   ```json
   {"status":"start","message":"App gestartet"}
   ```

2. **Progress:** Wird jedes Mal gesendet, wenn ein Bericht erfolgreich generiert wurde.
   ```json
   {"status":"progress","file":"PROJ-123","current":1,"total":3,"message":"Bericht generiert"}
   ```

3. **Error:** Wird gesendet, wenn ein einzelner Bericht fehlgeschlagen ist (die Verarbeitung der anderen läuft weiter!).
   ```json
   {"status":"error","file":"PROJ-999","current":2,"total":3,"message":"Fehler beim Schreiben..."}
   ```

4. **Done:** Wird exakt einmal am Ende gesendet.
   ```json
   {"status":"done","current":3,"total":3,"message":"Fertig! 2 erfolgreich, 1 fehlerhaft."}
   ```

---

## 4. Best Practices für die Rust-Implementierung (Slint)

Wenn du (die KI) den Aufruf in Rust programmierst, nutze folgendes Pattern:

1. **JSON schreiben:** Serialisiere das Rust `Vec<BudgetData>` via `serde_json` in eine temporäre Datei (z. B. im System-Temp-Ordner).
2. **Subprocess starten:** Nutze `std::process::Command` und setze `.stdout(Stdio::piped())`.
3. **Streaming lesen:** Wickle den `stdout` des Child-Prozesses in einen `BufReader` und iteriere über `lines()`.
4. **UI updaten:** Parse jede gelesene JSON-Zeile wieder in ein Rust-Struct (`ProgressMessage`) und sende die Daten über einen Channel (`mpsc`) oder einen Slint-Callback in den UI-Thread, um den Ladebalken (`current` / `total`) und Status-Texte zu aktualisieren.
5. **Aufräumen:** Lösche die temporäre Input-JSON-Datei am Ende des Vorgangs.
