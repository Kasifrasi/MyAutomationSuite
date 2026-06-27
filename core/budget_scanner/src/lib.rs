use calamine::{open_workbook_auto, Data, Range, Reader};
use rayon::prelude::*;
use serde::Serialize;
use shared_constants;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::sync::LazyLock;
use thiserror::Error;
use walkdir::WalkDir;

// ── Constants ────────────────────────────────────────────────────────────────

const STOP_WORDS: &[&str] = &["Summe", "Total", "TOTAL"];
const COST_TERMS: &[&str] = &[
    "Gesamtkosten",
    "Total Costs",
    "Coût total",
    "Gastos total",
    "Custos total",
];

const EIGENLEISTUNG_TERMS: &[&str] = &[
    "Lokale Eigenleistung",
    "Local Contribution",
    "Contribution locale",
    "Aporte propio",
    "Contribuição própria",
];
const DRITTMITTEL_TERMS: &[&str] = &[
    "Drittmittel",
    "Third party contribution",
    "Contributions de tiers",
    "Aportes de terceros",
    "Contribucões de terceiros",
];
const KMW_TERMS: &[&str] = &[
    "Beim KMW beantragt",
    "Requested from KMW",
    "Montant demandé à Kindermissionswerk",
    "Importe solicitado KMW",
    "Subsídio solicitado KMW",
];

/// Kostenkategorien in der Reihenfolge, die der Positions-Präfix (1..8) adressiert.
/// Gemeinsame Domänen-Wahrheit für alle Consumer (FB- und VP-Sidecar).
pub const CATEGORIES: [&str; 8] = [
    "Bauausgaben",
    "Investitionen",
    "Personalkosten",
    "Projektaktivitaeten",
    "Projektverwaltung",
    "Evaluierung",
    "Audit",
    "Reserve",
];

const MAX_EMPTY_ROWS: usize = 100;
const DEFAULT_COL1: usize = 8; // Spalte I
const DEFAULT_COL2: usize = 13; // Spalte N
const FALLBACK_MAX_ROWS: usize = 100;
const FALLBACK_MAX_COLS: usize = 26;

// Regex wird einmalig kompiliert
static POSITION_RE: LazyLock<regex::Regex> =
    LazyLock::new(|| regex::Regex::new(r"^[1-8]\.\d*").unwrap());

// Projektnummer-Erkennung im Dateinamen (Fallback wenn Excel-Feld leer ist)
// Formate: nnnn_nnnn_nnn (numerisch) | b nn nnnn nnn (alpha)
static PROJECT_NUM_RE: LazyLock<regex::Regex> =
    LazyLock::new(|| regex::Regex::new(r"\d{4}_\d{4}_\d{3}|[A-Za-z] \d{2} \d{4} \d{3}").unwrap());

// ── Public Types ─────────────────────────────────────────────────────────────

#[derive(Debug, Clone, Serialize)]
pub struct BudgetData {
    pub file_path: PathBuf,
    pub sheet_name: String,
    pub version: String,
    pub project_title: String,
    pub project_number: String,
    pub language: String,
    pub local_currency: String,
    pub cost_col1: usize,
    pub cost_col2: Option<usize>,
    /// Volle Aufschlüsselung der Finanzierungsquellen (LC-Gesamt, Jahre, EUR).
    /// Kanonische Repräsentation der Finanzierung – seitenspezifische Consumer
    /// (z. B. das FB-Sidecar) leiten ihre flachen Felder hieraus ab.
    pub financing: FinancingDetail,
    pub positions: Vec<BudgetPosition>,
}

#[derive(Debug, Clone, Serialize)]
pub struct BudgetPosition {
    pub number: String,
    pub label: String,
    /// Aus `number` (Präfix 1..8) abgeleitete Kostenkategorie; "" falls keine Zuordnung.
    /// Wird direkt beim Scan befüllt, sodass jeder Consumer die Kategorie frei Haus erhält.
    pub kategorie: String,
    /// Betrag in Lokalwährung (Spalte cost_col1). None ⇒ leere Zelle.
    pub lc: Option<f64>,
    /// Kosten je Jahr/Phase (Spalten cost_col1+1..3, Lokalwährung). None ⇒ leere Zelle.
    pub y1: Option<f64>,
    pub y2: Option<f64>,
    pub y3: Option<f64>,
    /// Betrag in EUR (Spalte cost_col2). None ⇒ leere Zelle bzw. keine EUR-Spalte.
    pub eur: Option<f64>,
}

/// FinancingRow bündelt eine Finanzierungszeile (Eigenmittel/Drittmittel/KMW) mit
/// Gesamt-LC, den drei Jahres-/Phasenwerten (LC) und dem EUR-Gesamtwert.
/// Beträge sind typisiert: None ⇒ leere Zelle.
#[derive(Debug, Clone, Default, Serialize)]
pub struct FinancingRow {
    pub lc: Option<f64>,
    pub y1: Option<f64>,
    pub y2: Option<f64>,
    pub y3: Option<f64>,
    pub eur: Option<f64>,
}

/// FinancingDetail enthält die drei Finanzierungsquellen in voller Aufschlüsselung
/// (LC-Gesamt, Jahre, EUR) und ist die kanonische Repräsentation der Finanzierung.
#[derive(Debug, Clone, Default, Serialize)]
pub struct FinancingDetail {
    pub eigenmittel: FinancingRow,
    pub drittmittel: FinancingRow,
    pub kmw_mittel: FinancingRow,
}

#[derive(Debug, Serialize)]
pub struct ScanFailure {
    pub file_path: PathBuf,
    pub file_name: String,
    pub reason: ScanError,
}

#[derive(Debug, Error)]
pub enum ScanError {
    #[error("Datei konnte nicht geöffnet werden: {0}")]
    OpenFailed(String),

    #[error("Kein passendes Sheet gefunden (vorhanden: {available})")]
    NoMatchingSheet { available: String },

    #[error("Version in A2 ungültig: \"{found}\" (erwartet: enthält \"V2\")")]
    InvalidVersion { found: String },

    #[error("Keine Kostenspalte gefunden")]
    CostColumnNotFound,

    #[error("Fehler beim Lesen des Sheets: {0}")]
    ReadError(String),
}

// Damit der ScanError in JSON als normaler Fehlertext auftaucht
impl Serialize for ScanError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}

#[derive(Debug, Serialize)]
pub struct ScanResult {
    pub successes: Vec<BudgetData>,
    pub failures: Vec<ScanFailure>,
}

// ── Public API ───────────────────────────────────────────────────────────────

/// Scannt eine einzelne xlsx/xlsm Datei.
pub fn scan_file(path: &Path) -> Result<BudgetData, ScanFailure> {
    let file_name = path
        .file_name()
        .map(|n| n.to_string_lossy().into_owned())
        .unwrap_or_default();

    scan_file_inner(path).map_err(|reason| ScanFailure {
        file_path: path.to_path_buf(),
        file_name,
        reason,
    })
}

/// Scannt einen Ordner rekursiv nach xlsx/xlsm Dateien — parallelisiert mit rayon.
pub fn scan_directory(path: &Path) -> ScanResult {
    let entries: Vec<PathBuf> = WalkDir::new(path)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| {
            e.file_type().is_file()
                && matches!(
                    e.path().extension().and_then(|s| s.to_str()),
                    Some("xlsx") | Some("xlsm")
                )
        })
        .map(|e| e.into_path())
        .collect();

    let results: Vec<Result<BudgetData, ScanFailure>> =
        entries.par_iter().map(|p| scan_file(p)).collect();

    let mut successes = Vec::with_capacity(results.len());
    let mut failures = Vec::new();

    for result in results {
        match result {
            Ok(data) => successes.push(data),
            Err(failure) => failures.push(failure),
        }
    }

    ScanResult {
        successes,
        failures,
    }
}

/// Schreibt einen Fehler-Report als CSV-Datei.
pub fn write_failure_report(failures: &[ScanFailure], output: &Path) -> std::io::Result<()> {
    let mut buf = std::io::BufWriter::new(std::fs::File::create(output)?);

    writeln!(buf, "Dateiname;Grund;Pfad")?;

    for f in failures {
        let reason = f.reason.to_string().replace(';', ",");
        writeln!(buf, "{};{};{}", f.file_name, reason, f.file_path.display())?;
    }

    Ok(())
}

pub fn col_to_letter(col: usize) -> char {
    (b'A' + col as u8) as char
}

// ── Internal ─────────────────────────────────────────────────────────────────

#[inline]
fn cell_text_owned(cell: &Data) -> Option<String> {
    match cell {
        Data::String(s) => Some(s.clone()),
        Data::Float(f) => Some(f.to_string()),
        Data::Int(i) => Some(i.to_string()),
        _ => None,
    }
}

/// Wandelt eine Zelle in einen typisierten Betrag um. Numerische Zellen direkt,
/// Text-Zellen über `parse_amount` (lokalisierte Tausender-/Dezimaltrenner).
/// None ⇒ leere/unlesbare Zelle.
#[inline]
fn cell_amount(cell: &Data) -> Option<f64> {
    match cell {
        Data::Float(f) => Some(*f),
        Data::Int(i) => Some(*i as f64),
        Data::String(s) => parse_amount(s),
        _ => None,
    }
}

/// Parst einen Geldbetrag aus einem String und erkennt dabei robust deutsche
/// ("1.234,56") wie englische ("1,234.56") Tausender-/Dezimaltrenner. None bei leer.
fn parse_amount(raw: &str) -> Option<f64> {
    let mut s: String = raw
        .chars()
        .filter(|c| c.is_ascii_digit() || *c == ',' || *c == '.' || *c == '-')
        .collect();
    if s.is_empty() || s == "-" {
        return None;
    }

    let has_dot = s.contains('.');
    let has_comma = s.contains(',');

    if has_dot && has_comma {
        // Der zuletzt auftretende Trenner ist der Dezimaltrenner.
        let last_dot = s.rfind('.').unwrap();
        let last_comma = s.rfind(',').unwrap();
        if last_comma > last_dot {
            s = s.replace('.', "").replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    } else if has_comma {
        let after = s.rsplit(',').next().map(|p| p.len()).unwrap_or(0);
        let commas = s.matches(',').count();
        if commas == 1 && after != 3 {
            // Einzelnes Komma, nicht im Tausenderformat ⇒ Dezimaltrenner.
            s = s.replace(',', ".");
        } else {
            s = s.replace(',', "");
        }
    }

    s.parse::<f64>().ok()
}

/// Leitet aus einer Positionsnummer ("n" / "n.m") die Kostenkategorie ab.
/// Liefert "" für Nummern außerhalb 1..8.
fn category_for(number: &str) -> String {
    let head = number.split('.').next().unwrap_or("").trim();
    match head.parse::<usize>() {
        Ok(n) if (1..=8).contains(&n) => CATEGORIES[n - 1].to_string(),
        _ => String::new(),
    }
}

/// Schneller Vergleich: Ist der Zellinhalt exakt einer der Kostenbegriffe?
#[inline]
fn is_exact_cost_term(cell: &Data) -> bool {
    match cell {
        Data::String(s) => COST_TERMS.iter().any(|&t| s.trim() == t),
        _ => false,
    }
}

/// Prüft ob eine bestimmte Spalte in irgendeiner Zeile einen Kostenbegriff enthält.
/// Bricht beim ersten Fund ab.
fn col_has_cost_term(range: &Range<Data>, col: usize) -> bool {
    range
        .rows()
        .any(|row| row.get(col).is_some_and(is_exact_cost_term))
}

fn find_cost_columns(range: &Range<Data>) -> Result<(usize, Option<usize>), ScanError> {
    let first_col = if col_has_cost_term(range, DEFAULT_COL1) {
        Some(DEFAULT_COL1)
    } else {
        None
    };
    let second_col = if col_has_cost_term(range, DEFAULT_COL2) {
        Some(DEFAULT_COL2)
    } else {
        None
    };

    if first_col.is_some() && second_col.is_some() {
        return Ok((DEFAULT_COL1, Some(DEFAULT_COL2)));
    }

    // Fallback: A1:Z100
    let row_count = range.height().min(FALLBACK_MAX_ROWS);
    let mut found: [Option<usize>; 2] = [None, None];
    let mut found_count = 0usize;

    'outer: for row_idx in 0..row_count {
        for col_idx in 0..FALLBACK_MAX_COLS {
            if let Some(cell) = range.get((row_idx, col_idx)) {
                if is_exact_cost_term(cell) {
                    // Kein Duplikat
                    if found[0] != Some(col_idx) {
                        found[found_count] = Some(col_idx);
                        found_count += 1;
                        if found_count == 2 {
                            break 'outer;
                        }
                    }
                }
            }
        }
    }

    let resolved_first = first_col
        .or(found[0])
        .ok_or(ScanError::CostColumnNotFound)?;
    let resolved_second = second_col.or_else(|| found[1].filter(|&col| col > resolved_first));

    Ok((resolved_first, resolved_second))
}

/// Liest eine vollständige Finanzierungszeile (LC, Jahre, EUR) anhand der Begriffe in
/// Spalte D. Die Jahresspalten liegen unmittelbar rechts von der LC-Gesamtspalte
/// (col1 + 1..3), die EUR-Gesamtspalte ist col2 (sofern vorhanden).
fn find_financing_row(
    range: &Range<Data>,
    terms: &[&str],
    col1: usize,
    col2: Option<usize>,
) -> FinancingRow {
    for row in range.rows() {
        if let Some(Data::String(s)) = row.get(3) {
            if terms.contains(&s.trim()) {
                let amount = |c: usize| row.get(c).and_then(cell_amount);
                return FinancingRow {
                    lc: amount(col1),
                    y1: amount(col1 + 1),
                    y2: amount(col1 + 2),
                    y3: amount(col1 + 3),
                    eur: col2.and_then(amount),
                };
            }
        }
    }
    FinancingRow::default()
}

fn scan_file_inner(path: &Path) -> Result<BudgetData, ScanError> {
    let mut wb = open_workbook_auto(path).map_err(|e| ScanError::OpenFailed(e.to_string()))?;

    let sheet_names = wb.sheet_names();

    let sheet_name = shared_constants::BUDGET_SHEET_NAMES
        .iter()
        .find(|&&name| sheet_names.iter().any(|s| s == name))
        .copied()
        .ok_or_else(|| ScanError::NoMatchingSheet {
            available: sheet_names.join(", "),
        })?;

    let range = wb
        .worksheet_range(sheet_name)
        .map_err(|e| ScanError::ReadError(e.to_string()))?;

    // Schnelle Zell-Lese-Helfer
    let get_str = |row: usize, col: usize| -> String {
        range
            .get((row, col))
            .and_then(cell_text_owned)
            .unwrap_or_default()
    };

    // Version-Check: A2 muss "V2" enthalten
    let version = get_str(1, 0);
    if !version.to_uppercase().contains("V2") {
        return Err(ScanError::InvalidVersion { found: version });
    }

    let (col1, col2) = find_cost_columns(&range)?;

    // Finanzierungsquellen in voller Aufschlüsselung (LC-Gesamt, Jahre, EUR).
    let financing = FinancingDetail {
        eigenmittel: find_financing_row(&range, EIGENLEISTUNG_TERMS, col1, col2),
        drittmittel: find_financing_row(&range, DRITTMITTEL_TERMS, col1, col2),
        kmw_mittel: find_financing_row(&range, KMW_TERMS, col1, col2),
    };

    // Positionen extrahieren
    let re = &*POSITION_RE;
    let mut positions = Vec::with_capacity(128);
    let mut first_match_found = false;
    let mut empty_streak = 0u32;

    for row in range.rows() {
        let cell = &row[0];

        // Fast path: String-Zellen direkt ohne Allokation prüfen
        let text = match cell {
            Data::String(s) => s.as_str(),
            Data::Empty => {
                if first_match_found {
                    empty_streak += 1;
                    if empty_streak >= MAX_EMPTY_ROWS as u32 {
                        break;
                    }
                }
                continue;
            }
            _ => {
                if first_match_found {
                    empty_streak += 1;
                    if empty_streak >= MAX_EMPTY_ROWS as u32 {
                        break;
                    }
                }
                continue;
            }
        };

        let trimmed = text.trim();

        if trimmed.is_empty() {
            if first_match_found {
                empty_streak += 1;
                if empty_streak >= MAX_EMPTY_ROWS as u32 {
                    break;
                }
            }
            continue;
        }

        // Stop-Words: schneller Check über Bytes statt .contains()
        if STOP_WORDS.iter().any(|&w| trimmed.contains(w)) {
            break;
        }

        if let Some(m) = re.find(trimmed) {
            first_match_found = true;
            empty_streak = 0;
            let matched: &str = m.as_str();

            let number = matched.to_string();
            positions.push(BudgetPosition {
                kategorie: category_for(&number),
                number,
                label: row.get(1).and_then(cell_text_owned).unwrap_or_default(),
                lc: row.get(col1).and_then(cell_amount),
                y1: row.get(col1 + 1).and_then(cell_amount),
                y2: row.get(col1 + 2).and_then(cell_amount),
                y3: row.get(col1 + 3).and_then(cell_amount),
                eur: col2.and_then(|c| row.get(c)).and_then(cell_amount),
            });
        }
    }

    Ok(BudgetData {
        file_path: path.to_path_buf(),
        sheet_name: sheet_name.to_string(),
        version,
        project_title: get_str(1, 2),
        project_number: {
            let num = get_str(1, 8);
            if !num.is_empty() {
                num
            } else {
                extract_project_number_from_filename(path).unwrap_or_default()
            }
        },
        language: get_str(2, 8),
        local_currency: get_str(3, 8),
        cost_col1: col1,
        cost_col2: col2,
        financing,
        positions,
    })
}

// ── Output-Ordner Logik ──────────────────────────────────────────────────────

/// Extrahiert eine gültige Projektnummer aus dem Dateinamen.
/// Unterstützt beide Formate: `nnnn_nnnn_nnn` und `b nn nnnn nnn`.
fn extract_project_number_from_filename(path: &Path) -> Option<String> {
    let file_stem = path.file_stem()?.to_str()?;
    PROJECT_NUM_RE
        .find(file_stem)
        .map(|m| m.as_str().to_string())
}

/// Findet einen freien Output-Ordner: `base/output`, `base/output_1`, `base/output_2`, etc.
pub fn resolve_output_dir(base: &Path) -> PathBuf {
    let candidate = base.join("output");
    if !candidate.exists() {
        return candidate;
    }
    let mut counter = 1u32;
    loop {
        let candidate = base.join(format!("output_{counter}"));
        if !candidate.exists() {
            return candidate;
        }
        counter += 1;
    }
}

// ── Tests ────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn extract_numeric_from_filename() {
        let path = Path::new("/data/Budget_2025_0004_003.xlsx");
        assert_eq!(
            extract_project_number_from_filename(path).as_deref(),
            Some("2025_0004_003")
        );
    }

    #[test]
    fn extract_numeric_at_start() {
        let path = Path::new("2025_0004_003_Budget.xlsx");
        assert_eq!(
            extract_project_number_from_filename(path).as_deref(),
            Some("2025_0004_003")
        );
    }

    #[test]
    fn extract_numeric_after_hyphen() {
        let path = Path::new("Report-2025_0004_003.xlsx");
        assert_eq!(
            extract_project_number_from_filename(path).as_deref(),
            Some("2025_0004_003")
        );
    }

    #[test]
    fn extract_alpha_from_filename() {
        let path = Path::new("Budget_a 12 3456 789.xlsx");
        assert_eq!(
            extract_project_number_from_filename(path).as_deref(),
            Some("a 12 3456 789")
        );
    }

    #[test]
    fn extract_alpha_uppercase() {
        let path = Path::new("Pruefung_Z 99 9999 999_Report.xlsx");
        assert_eq!(
            extract_project_number_from_filename(path).as_deref(),
            Some("Z 99 9999 999")
        );
    }

    #[test]
    fn extract_returns_none_when_no_match() {
        let path = Path::new("Kosten-_und_Finanzierungsplan_V2.xlsx");
        assert_eq!(extract_project_number_from_filename(path), None);
    }

    #[test]
    fn excel_number_has_priority() {
        // Simuliert: Excel-Feld ist nicht leer → Dateiname wird ignoriert
        let excel_num = "2025_0004_003";
        let fallback = extract_project_number_from_filename(Path::new("Budget_9999_9999_999.xlsx"));
        let result = if !excel_num.is_empty() {
            excel_num.to_string()
        } else {
            fallback.unwrap_or_default()
        };
        assert_eq!(result, "2025_0004_003");
    }

    #[test]
    fn fallback_when_excel_empty() {
        // Simuliert: Excel-Feld ist leer → Dateiname wird verwendet
        let excel_num = "";
        let fallback = extract_project_number_from_filename(Path::new("Budget_2025_0004_003.xlsx"));
        let result = if !excel_num.is_empty() {
            excel_num.to_string()
        } else {
            fallback.unwrap_or_default()
        };
        assert_eq!(result, "2025_0004_003");
    }
}
