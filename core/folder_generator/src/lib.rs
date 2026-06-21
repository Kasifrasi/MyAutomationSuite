use std::path::{Path, PathBuf};

// ==========================================
// Error types
// ==========================================

#[derive(Debug, thiserror::Error)]
pub enum FolderError {
    #[error("Ordner existiert bereits: {0}")]
    AlreadyExists(PathBuf),

    #[error("Fehler beim Erstellen von '{path}': {source}")]
    CreateDir {
        path: PathBuf,
        source: std::io::Error,
    },

    #[error("Fehler beim Kopieren der Vorlage nach '{dest}': {source}")]
    CopyTemplate {
        dest: PathBuf,
        source: std::io::Error,
    },

    #[error("CSV konnte nicht gelesen werden: {0}")]
    CsvRead(std::io::Error),
}

// ==========================================
// CSV import result
// ==========================================

#[derive(Debug, Default)]
pub struct CsvImportResult {
    pub created: u32,
    pub skipped: u32,
    pub errors: Vec<String>,
}

// ==========================================
// Constants
// ==========================================

pub const SUBFOLDERS: &[&str] = &[
    "00. Prüfungen",
    "01. Vorprojektsaldo",
    "02. Vertrag",
    "03. Budget",
    "04. Finanzberichte",
    "05. Bankbelege",
    "06. Mittelanforderungen",
    "07. Narrative Berichte",
    "Papierkorb",
];

// ==========================================
// Validation
// ==========================================

/// Checks whether `name` is a valid 13-character project number.
///
/// Accepted formats:
/// - Numeric: `nnnn_nnnn_nnn` (digits with underscores at positions 4 and 9)
/// - Alpha:   `b nn nnnn nnn` (letter, then digits with spaces at positions 1, 4, 9)
pub fn is_valid_project_number(name: &str) -> bool {
    if name.len() != 13 {
        return false;
    }
    let b = name.as_bytes();

    // Numerisch: nnnn_nnnn_nnn
    let numeric = b.iter().enumerate().all(|(i, &c)| match i {
        4 | 9 => c == b'_',
        _ => c.is_ascii_digit(),
    });

    // Alpha: b nn nnnn nnn
    let alpha = b[0].is_ascii_alphabetic()
        && b.iter().enumerate().skip(1).all(|(i, &c)| match i {
            1 | 4 | 9 => c == b' ',
            _ => c.is_ascii_digit(),
        });

    numeric || alpha
}

// ==========================================
// Formatting
// ==========================================

/// Auto-formats raw user input into a proper project number.
///
/// - Strips non-alphanumeric characters
/// - If the first character is a letter → alpha format: `b nn nnnn nnn`
/// - If the first character is a digit → numeric format: `nnnn_nnnn_nnn`
pub fn format_project_name(raw: &str) -> String {
    let chars: Vec<char> = raw.chars().filter(|c| c.is_ascii_alphanumeric()).collect();

    if chars.is_empty() {
        return String::new();
    }

    let trailing_sep = raw.ends_with(|c: char| c.is_ascii_punctuation() || c.is_ascii_whitespace());

    if chars[0].is_ascii_alphabetic() {
        // Alpha-Modus: b zz zzzz zzz (max 1 Buchstabe + 9 Ziffern)
        let letter = chars[0].to_ascii_uppercase();
        let digits: Vec<char> = chars[1..]
            .iter()
            .filter(|c| c.is_ascii_digit())
            .copied()
            .take(9)
            .collect();
        let mut result = String::with_capacity(13);
        result.push(letter);
        if !digits.is_empty() || trailing_sep {
            result.push(' ');
        }
        for (i, d) in digits.iter().enumerate() {
            if i == 2 || i == 6 {
                result.push(' ');
            }
            result.push(*d);
        }
        if (digits.len() == 2 || digits.len() == 6) && trailing_sep {
            result.push(' ');
        }
        result
    } else {
        // Numerisch-Modus: zzzz_zzzz_zzz (max 11 Ziffern)
        let digits: Vec<char> = chars
            .iter()
            .filter(|c| c.is_ascii_digit())
            .copied()
            .take(11)
            .collect();
        let mut result = String::with_capacity(13);
        for (i, d) in digits.iter().enumerate() {
            if i == 4 || i == 8 {
                result.push('_');
            }
            result.push(*d);
        }
        if (digits.len() == 4 || digits.len() == 8) && trailing_sep {
            result.push('_');
        }
        result
    }
}

// ==========================================
// Folder creation
// ==========================================

/// Creates a project folder structure under `target/project_name`.
///
/// - Creates the project root directory
/// - Creates all given subfolders
/// - Copies `template` as `Pruefung_{project_name}.{ext}`
/// - Writes a `.root.txt` marker file
pub fn create_project_folder(
    project_name: &str,
    target: &Path,
    template: &Path,
    subfolders: &[&str],
) -> Result<PathBuf, FolderError> {
    let project_root = target.join(project_name);
    if project_root.exists() {
        return Err(FolderError::AlreadyExists(project_root));
    }

    std::fs::create_dir(&project_root).map_err(|e| FolderError::CreateDir {
        path: project_root.clone(),
        source: e,
    })?;

    for sub in subfolders {
        let sub_path = project_root.join(sub);
        std::fs::create_dir(&sub_path).map_err(|e| FolderError::CreateDir {
            path: sub_path,
            source: e,
        })?;
    }

    let ext = template
        .extension()
        .and_then(|e| e.to_str())
        .unwrap_or("xlsm");
    let dest = project_root.join(format!("Pruefung_{project_name}.{ext}"));
    std::fs::copy(template, &dest).map_err(|e| FolderError::CopyTemplate { dest, source: e })?;

    let _ = std::fs::write(
        project_root.join(".root.txt"),
        "Ordner automatisch generiert.",
    );

    Ok(project_root)
}

// ==========================================
// CSV batch import
// ==========================================

/// Reads a CSV file and creates a project folder for each non-empty cell.
///
/// Supports both comma and semicolon as delimiters.
/// Returns [`CsvImportResult`] with counts of created/skipped folders and any errors.
pub fn import_csv(
    csv_path: &Path,
    target: &Path,
    template: &Path,
    subfolders: &[&str],
) -> Result<CsvImportResult, FolderError> {
    let content = std::fs::read_to_string(csv_path).map_err(FolderError::CsvRead)?;

    let mut result = CsvImportResult::default();

    for line in content.lines() {
        for cell in line.split([',', ';']) {
            let name = cell.trim().trim_matches('"').trim();
            if name.is_empty() {
                continue;
            }
            match create_project_folder(name, target, template, subfolders) {
                Ok(_) => result.created += 1,
                Err(FolderError::AlreadyExists(_)) => result.skipped += 1,
                Err(e) => result.errors.push(e.to_string()),
            }
        }
    }

    Ok(result)
}

// ==========================================
// Tests
// ==========================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn valid_numeric_project_numbers() {
        assert!(is_valid_project_number("2025_0004_003"));
        assert!(is_valid_project_number("0000_0000_000"));
        assert!(is_valid_project_number("9999_9999_999"));
    }

    #[test]
    fn valid_alpha_project_numbers() {
        assert!(is_valid_project_number("a 12 3456 789"));
        assert!(is_valid_project_number("Z 99 9999 999"));
    }

    #[test]
    fn invalid_project_numbers() {
        assert!(!is_valid_project_number(""));
        assert!(!is_valid_project_number("too_short"));
        assert!(!is_valid_project_number("20250004003__"));
        assert!(!is_valid_project_number("2025-0004-003"));
        assert!(!is_valid_project_number("abcd_efgh_ijk"));
    }

    #[test]
    fn format_numeric_input() {
        assert_eq!(format_project_name("20250004003"), "2025_0004_003");
        assert_eq!(format_project_name("2025_0004"), "2025_0004");
        assert_eq!(format_project_name("12345678901"), "1234_5678_901");
    }

    #[test]
    fn format_alpha_input() {
        assert_eq!(format_project_name("a123456789"), "a 12 3456 789");
        assert_eq!(format_project_name("Z99"), "Z 99");
    }

    #[test]
    fn format_empty_and_partial() {
        assert_eq!(format_project_name(""), "");
        assert_eq!(format_project_name("5"), "5");
        assert_eq!(format_project_name("a"), "a");
    }

    #[test]
    fn format_strips_non_alphanumeric() {
        assert_eq!(format_project_name("20-25_00-04"), "2025_0004");
        assert_eq!(format_project_name("a..12..34"), "a 12 34");
    }
}
