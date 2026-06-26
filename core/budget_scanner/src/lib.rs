use calamine::{open_workbook_auto, Data, Range, Reader};
use rayon::prelude::*;
use serde::Serialize;
use std::io::Write;
use std::path::{Path, PathBuf};
use std::sync::LazyLock;
use thiserror::Error;
use walkdir::WalkDir;

// ── Constants ────────────────────────────────────────────────────────────────

const SHEET_NAMES: &[&str] = &["Budget", "Presupuesto", "Plano de custos e financiamento"];
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

const MAX_EMPTY_ROWS: usize = 100;
const DEFAULT_COL1: usize = 8; // Spalte I
const DEFAULT_COL2: usize = 13; // Spalte N
const FALLBACK_MAX_ROWS: usize = 100;
const FALLBACK_MAX_COLS: usize = 26;

// Regex wird einmalig kompiliert
static POSITION_RE: LazyLock<regex::Regex> =
    LazyLock::new(|| regex::Regex::new(r"^[1-8]\.\d*").unwrap());

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
    pub cost_col1: String,
    pub cost_col2: String,
    /// Kosten je Jahr/Phase (Spalten J/K/L, jeweils Lokalwährung).
    pub cost_year1: String,
    pub cost_year2: String,
    pub cost_year3: String,
}

/// FinancingRow bündelt eine Finanzierungszeile (Eigenleistung/Drittmittel/KMW) mit
/// Gesamt-LC, den drei Jahres-/Phasenwerten (LC) und dem EUR-Gesamtwert.
#[derive(Debug, Clone, Default, Serialize)]
pub struct FinancingRow {
    pub lc: String,
    pub year1: String,
    pub year2: String,
    pub year3: String,
    pub eur: String,
}

/// FinancingDetail enthält die drei Finanzierungsquellen in voller Aufschlüsselung
/// (LC-Gesamt, Jahre, EUR) und ist die kanonische Repräsentation der Finanzierung.
#[derive(Debug, Clone, Default, Serialize)]
pub struct FinancingDetail {
    pub eigenleistung: FinancingRow,
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
                let get = |c: usize| row.get(c).and_then(cell_text_owned).unwrap_or_default();
                return FinancingRow {
                    lc: get(col1),
                    year1: get(col1 + 1),
                    year2: get(col1 + 2),
                    year3: get(col1 + 3),
                    eur: col2.map(get).unwrap_or_default(),
                };
            }
        }
    }
    FinancingRow::default()
}

fn scan_file_inner(path: &Path) -> Result<BudgetData, ScanError> {
    let mut wb = open_workbook_auto(path).map_err(|e| ScanError::OpenFailed(e.to_string()))?;

    let sheet_names = wb.sheet_names();

    let sheet_name = SHEET_NAMES
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
        eigenleistung: find_financing_row(&range, EIGENLEISTUNG_TERMS, col1, col2),
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

            positions.push(BudgetPosition {
                number: matched.to_string(),
                label: row.get(1).and_then(cell_text_owned).unwrap_or_default(),
                cost_col1: row.get(col1).and_then(cell_text_owned).unwrap_or_default(),
                cost_col2: col2
                    .and_then(|c| row.get(c))
                    .and_then(cell_text_owned)
                    .unwrap_or_default(),
                cost_year1: row
                    .get(col1 + 1)
                    .and_then(cell_text_owned)
                    .unwrap_or_default(),
                cost_year2: row
                    .get(col1 + 2)
                    .and_then(cell_text_owned)
                    .unwrap_or_default(),
                cost_year3: row
                    .get(col1 + 3)
                    .and_then(cell_text_owned)
                    .unwrap_or_default(),
            });
        }
    }

    Ok(BudgetData {
        file_path: path.to_path_buf(),
        sheet_name: sheet_name.to_string(),
        version,
        project_title: get_str(1, 2),
        project_number: get_str(1, 8),
        language: get_str(2, 8),
        local_currency: get_str(3, 8),
        cost_col1: col1,
        cost_col2: col2,
        financing,
        positions,
    })
}

// ── Output-Ordner Logik ──────────────────────────────────────────────────────

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
