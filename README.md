<div align="center">

# MyAutomationSuite

**Desktop-Automatisierungstool für Budget-Verarbeitung, Ordnergenerierung und Excel-Workflows**

[![CI](https://github.com/<OWNER>/MyAutomationSuite/actions/workflows/ci.yml/badge.svg)](https://github.com/<OWNER>/MyAutomationSuite/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE-MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE-APACHE)
![Version](https://img.shields.io/badge/version-0.4.5-blue)
![Rust](https://img.shields.io/badge/rust-1.87%2B-orange)
![Go](https://img.shields.io/badge/go-1.26%2B-cyan)

</div>

---

## Überblick

MyAutomationSuite ist eine native Windows-Desktop-Anwendung zur Automatisierung wiederkehrender Aufgaben im Bereich Budget-Verarbeitung, Ordnerstruktur-Generierung und Excel-Dateimanipulation. Die Anwendung kombiniert eine performante Rust-Codebasis mit einer modernen [Slint](https://slint.dev/)-GUI und Go-basierten Sidecar-Prozessen für spezialisierte Report-Generierung.

## Features

| Feature | Beschreibung |
|---|---|
| **FB-Generator** | Erstellt Fachberichte/Reports über einen Go-Sidecar-Prozess |
| **Budget → FB** | Konvertiert Budgetdaten in das FB-Ausgabeformat |
| **Budget → VP** | Überführt Budgetdaten in das Vorprüfungs-Format |
| **Budget Scanner** | Durchsucht Excel-Dateien nach Budget-relevanten Einträgen (parallelisiert mit Rayon) |
| **Ordner-Generator** | Generiert validierte Projektordnerstrukturen inkl. CSV-Batch-Import |
| **Excel Protection** | Versieht Excel-Dateien programmgesteuert mit Blattschutz |
| **Auto-Update** | Integrierte Selbstaktualisierung über GitHub Releases |
| **Dark Mode** | Automatische Erkennung und Umschaltung zwischen Hell-/Dunkelmodus |

## Architektur

```
MyAutomationSuite/
├── app/                          # Hauptanwendung (Slint GUI)
│   ├── src/
│   │   ├── main.rs               # Einstiegspunkt & Feature-Setup
│   │   ├── features/             # Feature-spezifische UI-Logik
│   │   ├── shared/               # Gemeinsame Hilfsfunktionen
│   │   └── shell/                # Shell-UI, Theme & Updater
│   └── ui/                       # Slint UI-Definitionen (.slint)
├── core/                         # Rust-Kernbibliotheken
│   ├── budget_scanner/           # Excel-Budget-Scanner (calamine, rayon)
│   ├── excel_protection/         # Excel-Blattschutz (quick-xml, zip)
│   └── folder_generator/         # Ordnerstruktur-Generator
├── sidecars/                     # Go-basierte Sidecar-Prozesse
│   ├── FB/                       # FB-Report-Generator (→ fb_generator.exe)
│   ├── Vorpruefung/              # Vorprüfungs-Generator (→ vp_generator.exe)
│   └── shared/                   # Geteilte Go-Bibliotheken
├── testdata/                     # Testdaten & Beispiel-Excel-Dateien
└── .github/workflows/            # CI/CD-Pipelines
```

Das Projekt ist als **Rust Workspace** organisiert mit drei internen Crates unter `core/` und einer Desktop-App unter `app/`. Die Go-Sidecars werden separat über eine `go.work`-Workspace-Datei verwaltet und zur Build-Zeit als ausführbare Dateien eingebunden.

## Voraussetzungen

| Abhängigkeit | Version | Zweck |
|---|---|---|
| [Rust](https://rustup.rs/) | ≥ 1.87 | Hauptcompiler & Toolchain |
| [Go](https://go.dev/dl/) | ≥ 1.26 | Sidecar-Prozesse |
| Windows SDK | x86_64-pc-windows-msvc | Zielplattform |

> **Hinweis:** Die Anwendung ist primär für Windows konzipiert. Unter anderen Betriebssystemen können Anpassungen notwendig sein.

## Installation & Build

### 1. Repository klonen

```bash
git clone https://github.com/<OWNER>/MyAutomationSuite.git
cd MyAutomationSuite
```

### 2. Go-Sidecars bauen

```bash
# FB-Generator
cd sidecars/FB
go build -o fb_generator.exe ./cmd/report_generator

# Vorprüfungs-Generator
cd ../Vorpruefung
go build -o vp_generator.exe ./cmd/vp_generator
cd ../..
```

### 3. Rust-Anwendung bauen

```bash
# Debug-Build
cargo build -p app

# Release-Build
cargo build --release -p app
```

Die fertige Anwendung befindet sich unter `target/release/app.exe`.

### Entwicklung mit Live-UI-Preview

```bash
cargo run -p app --features dev-ui
```

## Nutzung

Nach dem Start öffnet sich die Anwendung im Vollbildmodus (Windows). Über die Seitenleiste navigieren Sie zwischen den einzelnen Features:

1. **FB Generator** – Starten Sie den Go-Sidecar zur Report-Erstellung.
2. **Budget → FB** – Laden Sie eine Budget-Excel-Datei und generieren Sie den zugehörigen FB.
3. **Budget → VP** – Konvertieren Sie Budgetdaten in das Vorprüfungs-Format.
4. **Ordner Generator** – Erstellen Sie Projektordnerstrukturen manuell oder per CSV-Batch-Import.

Die Anwendung erkennt automatisch das System-Theme (Hell/Dunkel) und passt die Oberfläche entsprechend an.

### Auto-Update

Die Anwendung prüft beim Start automatisch auf neue Versionen über GitHub Releases. Bei verfügbaren Updates wird ein Dialog zur Aktualisierung angezeigt.

## Tests

```bash
# Alle Workspace-Tests ausführen
cargo test

# Spezifische Crate-Tests
cargo test -p budget_scanner
cargo test -p folder_generator
cargo test -p excel_protection
```

## Lizenz

Dieses Projekt ist unter einer dualen Lizenz verfügbar:

- **[MIT License](./LICENSE-MIT)**
- **[Apache License 2.0](./LICENSE-APACHE)**

Sie können nach den Bedingungen einer der beiden Lizenzen verwenden.

Siehe auch: [THIRDPARTY_LICENSES.md](./THIRDPARTY_LICENSES.md) für Abhängigkeiten von Drittanbietern.

---

<div align="center">

**MyAutomationSuite** · Version 0.4.5 · Gebaut mit Rust & Slint

</div>
