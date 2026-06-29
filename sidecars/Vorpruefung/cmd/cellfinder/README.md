# CellFinder Tool

Dieses Hilfsprogramm generiert automatisch eine frische, leere Excel-Vorlage (basierend auf dem aktuellsten Quellcode im Sidecar) und durchsucht diese gezielt nach allen Zellen, die als "Eingabefelder" für die API vorgesehen sind. 

Solche Felder sind in den Templates konsequent durch die Hintergrundfarbe `#FFFAE5` markiert. Das Tool ist extrem hilfreich, um die statischen Zell-Koordinaten für die Fill-API auszulesen oder zu aktualisieren, falls sich das Layout der Excel-Vorlage verschoben hat.

## Installation / Ausführung

Navigiere in den Ordner des CellFinders im Vorpruefung-Sidecar und führe das Tool über Go aus:

```bash
cd sidecars/Vorpruefung/cmd/cellfinder
go run main.go
```

## Parameter

### Standardlauf (Alle statischen Sheets)
Wird kein Parameter übergeben, durchsucht das Tool automatisch alle vordefinierten, statischen Sheets:
- `Dashboard`
- `II. KMW-Mittel`
- `IV. MA`
- `Pruefung FB`
- `Pruefung MA`

```bash
go run main.go
```

### Bestimmte Sheets durchsuchen (`-sheets`)
Mit dem Parameter `-sheets` kannst du gezielt angeben, welche Blätter analysiert werden sollen. Nutze dafür eine kommaseparierte Liste und verwende exakt die Namen der Sheets (bzw. die Namen, die auch in den `constants` definiert sind).

**Beispiele:**

Nur das Dashboard und die Mittelanforderung analysieren:
```bash
go run main.go -sheets "Dashboard,IV. MA"
```

Nur die Finanzbericht-Prüfung analysieren:
```bash
go run main.go -sheets "Pruefung FB"
```

## Funktionsweise
1. Es wird über `vorpruefung.GenerateVorpruefung(tmpFile, nil)` eine temporäre Datei (`temp_template_scan.xlsx`) generiert.
2. Das Programm iteriert über ein festes Raster (Spalten A-AX, Zeilen 1-200) der festgelegten Blätter.
3. Jeder Zell-Style wird ausgelesen, und wenn die Füllfarbe exakt `FFFAE5` ist, wird die Koordinate in die Konsole geprintet.
4. Abschließend wird die temporäre Datei wieder gelöscht.
