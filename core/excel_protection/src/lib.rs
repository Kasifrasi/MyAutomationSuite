//! ECMA-376 Workbook & Worksheet Protection: SHA-512 hash + ZIP/XML injection.
//!
//! # Architektur-Gedanke für Blattschutz & Mappenschutz
//! Anstatt die Zip-Datei mehrfach zu öffnen, nutzen wir eine `apply_protection` Methode.
//! Diese iteriert genau **einmal** über alle Dateien in der .xlsx (ZIP):
//! 1. Ist es `xl/workbook.xml`? -> Mappenschutz-Hash injizieren.
//! 2. Ist es `xl/worksheets/sheet*.xml`? -> Blattschutz-Hash injizieren.
//! 3. Sonst? -> Roh (unkomprimiert) kopieren.

use base64::{engine::general_purpose, Engine as _};
use byteorder::{WriteBytesExt, LE};
use quick_xml::events::Event;
use quick_xml::reader::Reader;
use quick_xml::writer::Writer;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha512};
use std::fs::File;
use std::io::{Cursor, Read, Write};
use zip::write::FileOptions;
use zip::{ZipArchive, ZipWriter};

const DEFAULT_SPIN_COUNT: u32 = 100_000;

#[derive(Debug)]
pub enum ProtectionError {
    Io(std::io::Error),
    Zip(zip::result::ZipError),
    Xml(quick_xml::Error),
    InvalidUtf8(std::str::Utf8Error),
    SheetNotFound(String),
}

impl std::fmt::Display for ProtectionError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Io(e) => write!(f, "I/O error: {e}"),
            Self::Zip(e) => write!(f, "ZIP error: {e}"),
            Self::Xml(e) => write!(f, "XML error: {e}"),
            Self::InvalidUtf8(e) => write!(f, "UTF-8 error: {e}"),
            Self::SheetNotFound(s) => write!(f, "Sheet not found in workbook: {s}"),
        }
    }
}
impl std::error::Error for ProtectionError {}

impl From<std::io::Error> for ProtectionError {
    fn from(e: std::io::Error) -> Self {
        Self::Io(e)
    }
}
impl From<zip::result::ZipError> for ProtectionError {
    fn from(e: zip::result::ZipError) -> Self {
        Self::Zip(e)
    }
}
impl From<quick_xml::Error> for ProtectionError {
    fn from(e: quick_xml::Error) -> Self {
        Self::Xml(e)
    }
}
impl From<std::str::Utf8Error> for ProtectionError {
    fn from(e: std::str::Utf8Error) -> Self {
        Self::InvalidUtf8(e)
    }
}

/// Fester Salt für den Batch-Vorgang, damit Hashes deterministisch vorgekaut werden können.
const FIXED_SALT: [u8; 16] = [
    0x6b, 0x6d, 0x77, 0x66, 0x62, 0x5f, 0x72, 0x70, 0x74, 0x5f, 0x76, 0x31, 0x5f, 0x21, 0x21, 0x00,
];

#[derive(Serialize, Deserialize, Debug, Clone, Default)]
pub struct SheetProtectionOptions {
    pub select_locked_cells: bool,
    pub select_unlocked_cells: bool,
    pub format_cells: bool,
    pub format_columns: bool,
    pub format_rows: bool,
    pub insert_columns: bool,
    pub insert_rows: bool,
    pub insert_hyperlinks: bool,
    pub delete_columns: bool,
    pub delete_rows: bool,
    pub sort: bool,
    pub auto_filter: bool,
    pub pivot_tables: bool,
    pub edit_objects: bool,
    pub edit_scenarios: bool,
}

impl SheetProtectionOptions {
    /// Wandelt die Optionen in die entsprechenden XML-Attribute um (z.B. `selectLockedCells="1" formatCells="0"`)
    pub fn to_xml_attributes(&self) -> String {
        let b2s = |b: bool| if b { "0" } else { "1" }; // In Excel XML bedeutet 1 meist "gesperrt/verboten" und 0 "erlaubt" (Vorsicht bei selectLockedCells etc.)

        // Die Logik für Excel-Schutzattribute ist invertiert.
        // Wenn man in Excel den Haken bei "Zellen formatieren" setzt (erlaubt), steht im XML: formatCells="0"
        // Wenn der Haken nicht gesetzt ist (verboten), steht im XML: formatCells="1"
        // Bei SelectLockedCells/SelectUnlockedCells ist "1" = erlaubt, "0" = verboten (das ist die Ausnahme!)

        let select_locked = if self.select_locked_cells { "0" } else { "1" };
        let select_unlocked = if self.select_unlocked_cells { "0" } else { "1" };

        format!(
            r#"sheet="1" selectLockedCells="{}" selectUnlockedCells="{}" formatCells="{}" formatColumns="{}" formatRows="{}" insertColumns="{}" insertRows="{}" insertHyperlinks="{}" deleteColumns="{}" deleteRows="{}" sort="{}" autoFilter="{}" pivotTables="{}" objects="{}" scenarios="{}""#,
            select_locked,
            select_unlocked,
            b2s(self.format_cells),
            b2s(self.format_columns),
            b2s(self.format_rows),
            b2s(self.insert_columns),
            b2s(self.insert_rows),
            b2s(self.insert_hyperlinks),
            b2s(self.delete_columns),
            b2s(self.delete_rows),
            b2s(self.sort),
            b2s(self.auto_filter),
            b2s(self.pivot_tables),
            b2s(self.edit_objects),
            b2s(self.edit_scenarios),
        )
    }
}
#[derive(Clone)]
pub struct PrecomputedHash {
    pub salt_b64: String,
    pub hash_b64: String,
    pub spin_count: u32,
    pub password_empty: bool,
}

/// Konfiguration für den Mappenschutz (workbook protection).
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct WorkbookConfig {
    /// Passwort für den Mappenschutz. `None` = kein Mappenschutz.
    pub password: Option<String>,
}

/// Konfiguration für den Blattschutz eines einzelnen Sheets.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SheetConfig {
    /// Name des Sheets (muss exakt dem Tab-Namen in der Excel-Datei entsprechen).
    /// Kann leer sein, wenn `index` verwendet wird.
    pub name: String,
    /// Index des Sheets (0-basiert, entspricht der Position im Workbook).
    /// Wird verwendet, wenn der Sheet-Name nicht bekannt ist.
    #[serde(default)]
    pub index: Option<u32>,
    /// Schutzoptionen (welche Aktionen erlaubt/verboten sind).
    pub options: SheetProtectionOptions,
    /// Passwort für dieses Sheet. `None` = Schutz ohne Passwort.
    #[serde(default)]
    pub password: Option<String>,
}

/// Löst Sheet-Namen und/oder -Indizes in die internen Dateipfade innerhalb der .xlsx (ZIP) auf.
///
/// Parst `xl/workbook.xml` (Sheet-Name → rId) und `xl/_rels/workbook.xml.rels`
/// (rId → Dateipfad). Nicht auflösbare Namen/Indizes werden stillschweigend übersprungen.
fn resolve_sheet_files(
    archive: &mut ZipArchive<File>,
    sheet_names: &[String],
    sheet_indices: &[u32],
) -> Vec<String> {
    if sheet_names.is_empty() && sheet_indices.is_empty() {
        return Vec::new();
    }

    let mut sheet_to_rid: std::collections::HashMap<String, String> =
        std::collections::HashMap::new();
    let mut ordered_sheets: Vec<(String, String)> = Vec::new(); // (name, rid) in Workbook-Reihenfolge

    if let Ok(mut wb_file) = archive.by_name("xl/workbook.xml") {
        let mut content = Vec::new();
        let _ = wb_file.read_to_end(&mut content);
        let mut reader = Reader::from_reader(&content[..]);
        let mut buf = Vec::new();
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Empty(ref e)) | Ok(Event::Start(ref e)) => {
                    if std::str::from_utf8(e.name().into_inner()).unwrap_or("") == "sheet" {
                        let mut name = String::new();
                        let mut rid = String::new();
                        for attr in e.attributes().flatten() {
                            match std::str::from_utf8(attr.key.into_inner()).unwrap_or("") {
                                "name" => {
                                    name =
                                        std::str::from_utf8(&attr.value).unwrap_or("").to_string();
                                }
                                "r:id" | "id" => {
                                    rid =
                                        std::str::from_utf8(&attr.value).unwrap_or("").to_string();
                                }
                                _ => {}
                            }
                        }
                        if !name.is_empty() && !rid.is_empty() {
                            ordered_sheets.push((name.clone(), rid.clone()));
                            sheet_to_rid.insert(name, rid);
                        }
                    }
                }
                Ok(Event::Eof) | Err(_) => break,
                _ => {}
            }
            buf.clear();
        }
    }

    let mut rid_to_file: std::collections::HashMap<String, String> =
        std::collections::HashMap::new();

    if let Ok(mut rels_file) = archive.by_name("xl/_rels/workbook.xml.rels") {
        let mut content = Vec::new();
        let _ = rels_file.read_to_end(&mut content);
        let mut reader = Reader::from_reader(&content[..]);
        let mut buf = Vec::new();
        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Empty(ref e)) | Ok(Event::Start(ref e)) => {
                    if std::str::from_utf8(e.name().into_inner()).unwrap_or("") == "Relationship" {
                        let mut id = String::new();
                        let mut target = String::new();
                        for attr in e.attributes().flatten() {
                            match std::str::from_utf8(attr.key.into_inner()).unwrap_or("") {
                                "Id" => {
                                    id = std::str::from_utf8(&attr.value).unwrap_or("").to_string();
                                }
                                "Target" => {
                                    target =
                                        std::str::from_utf8(&attr.value).unwrap_or("").to_string();
                                }
                                _ => {}
                            }
                        }
                        if !id.is_empty() && !target.is_empty() {
                            rid_to_file.insert(id, format!("xl/{}", target));
                        }
                    }
                }
                Ok(Event::Eof) | Err(_) => break,
                _ => {}
            }
            buf.clear();
        }
    }

    let mut result = Vec::new();

    // Nach Name auflösen
    for name in sheet_names {
        if let Some(rid) = sheet_to_rid.get(name) {
            if let Some(file) = rid_to_file.get(rid) {
                result.push(file.clone());
            }
        }
    }

    // Nach Index auflösen (0-basiert, entspricht der Reihenfolge im Workbook)
    for &idx in sheet_indices {
        if let Some((_, rid)) = ordered_sheets.get(idx as usize) {
            if let Some(file) = rid_to_file.get(rid) {
                result.push(file.clone());
            }
        }
    }

    result
}

/// 1. Hash-Berechnung (wird nur 1x pro Batch aufgerufen)
pub fn precompute_hash(password: &str) -> PrecomputedHash {
    if password.is_empty() {
        return PrecomputedHash {
            salt_b64: String::new(),
            hash_b64: String::new(),
            spin_count: DEFAULT_SPIN_COUNT,
            password_empty: true,
        };
    }

    let pw_utf16: Vec<u8> = password
        .encode_utf16()
        .flat_map(|c| c.to_le_bytes())
        .collect();
    let mut hasher = Sha512::new();
    hasher.update(FIXED_SALT);
    hasher.update(&pw_utf16);
    let mut hash = hasher.finalize();

    for i in 0..DEFAULT_SPIN_COUNT {
        let mut iterator = [0u8; 4];
        (&mut iterator[..]).write_u32::<LE>(i).unwrap();
        let mut next_hasher = Sha512::new();
        next_hasher.update(hash);
        next_hasher.update(iterator);
        hash = next_hasher.finalize();
    }

    PrecomputedHash {
        salt_b64: general_purpose::STANDARD.encode(FIXED_SALT),
        hash_b64: general_purpose::STANDARD.encode(hash),
        spin_count: DEFAULT_SPIN_COUNT,
        password_empty: false,
    }
}

// ============================================================================
// XML INJECTION LOGIK (Workbook)
// ============================================================================

pub fn inject_workbook_protection(
    xml_content: &[u8],
    hash: Option<&PrecomputedHash>,
) -> Result<Vec<u8>, ProtectionError> {
    let protection_tag = match hash {
        Some(h) if h.password_empty => r#"<workbookProtection lockStructure="1"/>"#.to_string(),
        Some(h) => format!(
            r#"<workbookProtection lockStructure="1" workbookAlgorithmName="SHA-512" workbookHashValue="{}" workbookSaltValue="{}" workbookSpinCount="{}"/>"#,
            h.hash_b64, h.salt_b64, h.spin_count
        ),
        None => "".to_string(),
    };

    let mut reader = Reader::from_reader(xml_content);
    let mut writer = Writer::new(Cursor::new(Vec::new()));
    let mut buf = Vec::new();
    let mut inserted = false;

    let try_insert =
        |w: &mut Writer<_>, tag: &str, inserted: &mut bool| -> Result<(), ProtectionError> {
            if !*inserted {
                if !tag.is_empty() {
                    write_tag(w, tag)?;
                }
                *inserted = true;
            }
            Ok(())
        };

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;

                if !inserted
                    && (name_str == "sheets"
                        || name_str == "bookViews"
                        || name_str == "functionGroups")
                {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }

                if name_str == "workbookProtection" {
                    let target_name = name_str.as_bytes().to_vec();
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;

                    // Skip the existing tag and its content
                    let mut depth = 1;
                    let mut skip_buf = Vec::new();
                    loop {
                        match reader.read_event_into(&mut skip_buf) {
                            Ok(Event::Start(e)) if e.name().into_inner() == target_name => {
                                depth += 1
                            }
                            Ok(Event::End(e)) if e.name().into_inner() == target_name => {
                                depth -= 1;
                                if depth == 0 {
                                    break;
                                }
                            }
                            Ok(Event::Eof) => break,
                            Err(_) => break,
                            _ => {}
                        }
                        skip_buf.clear();
                    }
                    continue;
                }
                writer.write_event(Event::Start(e.clone()))?;
            }
            Ok(Event::Empty(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;

                if !inserted
                    && (name_str == "sheets"
                        || name_str == "bookViews"
                        || name_str == "functionGroups")
                {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }

                if name_str == "workbookProtection" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                    continue;
                }
                writer.write_event(Event::Empty(e.clone()))?;
            }
            Ok(Event::End(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;
                if !inserted && name_str == "workbook" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }
                writer.write_event(Event::End(e.clone()))?;
            }
            Ok(Event::Eof) => break,
            Ok(e) => {
                writer.write_event(e)?;
            }
            Err(e) => return Err(e.into()),
        }
        buf.clear();
    }

    Ok(writer.into_inner().into_inner())
}

// ============================================================================
// XML INJECTION LOGIK (Worksheet)
// ============================================================================

pub fn inject_sheet_protection(
    xml_content: &[u8],
    hash: Option<&PrecomputedHash>,
    options: Option<&SheetProtectionOptions>,
) -> Result<Vec<u8>, ProtectionError> {
    let protection_tag = match (hash, options) {
        (Some(h), Some(opts)) => {
            let opts_str = opts.to_xml_attributes();
            if h.password_empty {
                format!(r#"<sheetProtection {}/>"#, opts_str)
            } else {
                format!(
                    r#"<sheetProtection algorithmName="SHA-512" hashValue="{}" saltValue="{}" spinCount="{}" {}/>"#,
                    h.hash_b64, h.salt_b64, h.spin_count, opts_str
                )
            }
        }
        _ => "".to_string(),
    };

    let mut reader = Reader::from_reader(xml_content);
    let mut writer = Writer::new(Cursor::new(Vec::new()));
    let mut buf = Vec::new();
    let mut inserted = false;

    let try_insert =
        |w: &mut Writer<_>, tag: &str, inserted: &mut bool| -> Result<(), ProtectionError> {
            if !*inserted {
                if !tag.is_empty() {
                    write_tag(w, tag)?;
                }
                *inserted = true;
            }
            Ok(())
        };

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;

                if name_str == "sheetProtection" {
                    let target_name = name_str.as_bytes().to_vec();
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;

                    // Skip the existing tag and its content
                    let mut depth = 1;
                    let mut skip_buf = Vec::new();
                    loop {
                        match reader.read_event_into(&mut skip_buf) {
                            Ok(Event::Start(e)) if e.name().into_inner() == target_name => {
                                depth += 1
                            }
                            Ok(Event::End(e)) if e.name().into_inner() == target_name => {
                                depth -= 1;
                                if depth == 0 {
                                    break;
                                }
                            }
                            Ok(Event::Eof) | Err(_) => break,
                            _ => {}
                        }
                        skip_buf.clear();
                    }
                    continue;
                }

                writer.write_event(Event::Start(e.clone()))?;
            }
            Ok(Event::Empty(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;

                if name_str == "sheetProtection" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                    continue;
                }

                writer.write_event(Event::Empty(e.clone()))?;

                if !inserted && name_str == "sheetData" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }
            }
            Ok(Event::End(ref e)) => {
                let name_str = std::str::from_utf8(e.name().into_inner())?;

                if !inserted && name_str == "worksheet" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }

                writer.write_event(Event::End(e.clone()))?;

                if !inserted && name_str == "sheetData" {
                    try_insert(&mut writer, &protection_tag, &mut inserted)?;
                }
            }
            Ok(Event::Eof) => break,
            Ok(e) => {
                writer.write_event(e)?;
            }
            Err(e) => return Err(e.into()),
        }
        buf.clear();
    }

    Ok(writer.into_inner().into_inner())
}

fn write_tag<W: std::io::Write>(writer: &mut Writer<W>, tag: &str) -> Result<(), ProtectionError> {
    let mut temp_reader = Reader::from_str(tag);
    temp_reader.config_mut().trim_text_start = true;
    temp_reader.config_mut().trim_text_end = true;
    loop {
        match temp_reader.read_event() {
            Ok(Event::Eof) => break,
            Ok(e) => {
                writer.write_event(e)?;
            }
            Err(e) => return Err(e.into()),
        }
    }
    Ok(())
}

/// Wendet Workbook- und Sheet-Schutz auf eine .xlsx-Datei an.
///
/// - `wb_config`: Mappenschutz (None = kein Schutz, Some = Schutz mit/ohne Passwort)
/// - `sheet_configs`: Liste der zu schützenden Sheets mit individuellen Optionen und Passwort
/// - `protect_all_sheets`: `true` = alle Sheets schützen (sheet_configs wird ignoriert),
///   `false` = nur die in sheet_configs genannten Sheets
pub fn apply_protection(
    path: &std::path::Path,
    wb_config: Option<&WorkbookConfig>,
    sheet_configs: &[SheetConfig],
    protect_all_sheets: bool,
) -> Result<(), ProtectionError> {
    // 1. Workbook-Hash precomputen
    let wb_hash = wb_config
        .and_then(|c| c.password.as_deref())
        .map(precompute_hash);

    // 2. Per-Sheet-Hashes precomputen (gleiche Passwörter werden nur 1x gehasht)
    let mut hash_cache: std::collections::HashMap<String, PrecomputedHash> =
        std::collections::HashMap::new();
    let empty_hash = PrecomputedHash {
        salt_b64: String::new(),
        hash_b64: String::new(),
        spin_count: DEFAULT_SPIN_COUNT,
        password_empty: true,
    };
    for config in sheet_configs {
        if let Some(pw) = config.password.as_deref() {
            hash_cache
                .entry(pw.to_string())
                .or_insert_with(|| precompute_hash(pw));
        }
    }

    let mut sheet_hashes: std::collections::HashMap<
        String,
        (PrecomputedHash, &SheetProtectionOptions),
    > = std::collections::HashMap::new();
    for config in sheet_configs {
        let hash = config
            .password
            .as_deref()
            .and_then(|pw| hash_cache.get(pw))
            .cloned()
            .unwrap_or_else(|| empty_hash.clone());
        sheet_hashes.insert(config.name.clone(), (hash, &config.options));
    }

    // 3. Geschützte Sheet-Dateien auflösen
    let protected_files: std::collections::HashSet<String> = if protect_all_sheets {
        std::collections::HashSet::new() // leer = alle schützen (wird unten geprüft)
    } else {
        let names: Vec<String> = sheet_configs
            .iter()
            .filter(|c| !c.name.is_empty())
            .map(|c| c.name.clone())
            .collect();
        let indices: Vec<u32> = sheet_configs.iter().filter_map(|c| c.index).collect();
        let file = File::open(path)?;
        let mut archive = ZipArchive::new(file)?;
        resolve_sheet_files(&mut archive, &names, &indices)
            .into_iter()
            .collect()
    };

    // Index → Name aus workbook.xml auflösen
    let mut index_to_name: std::collections::HashMap<u32, String> =
        std::collections::HashMap::new();
    let mut resolved_sheet_configs: Vec<SheetConfig> = Vec::new();

    if !protect_all_sheets {
        let file = File::open(path)?;
        let mut archive = ZipArchive::new(file)?;
        if let Ok(mut wb_file) = archive.by_name("xl/workbook.xml") {
            let mut content = Vec::new();
            let _ = wb_file.read_to_end(&mut content);
            let mut reader = Reader::from_reader(&content[..]);
            let mut buf = Vec::new();
            let mut idx = 0u32;
            loop {
                match reader.read_event_into(&mut buf) {
                    Ok(Event::Empty(ref e)) | Ok(Event::Start(ref e)) => {
                        if std::str::from_utf8(e.name().into_inner()).unwrap_or("") == "sheet" {
                            let mut sname = String::new();
                            for attr in e.attributes().flatten() {
                                if std::str::from_utf8(attr.key.into_inner()).unwrap_or("")
                                    == "name"
                                {
                                    sname =
                                        std::str::from_utf8(&attr.value).unwrap_or("").to_string();
                                }
                            }
                            index_to_name.insert(idx, sname);
                            idx += 1;
                        }
                    }
                    Ok(Event::Eof) | Err(_) => break,
                    _ => {}
                }
                buf.clear();
            }
        }

        // Index-basierte Configs auf aufgelöste Namen mappen
        for config in sheet_configs {
            if config.name.is_empty() {
                if let Some(sheet_idx) = config.index {
                    if let Some(resolved_name) = index_to_name.get(&sheet_idx) {
                        resolved_sheet_configs.push(SheetConfig {
                            name: resolved_name.clone(),
                            index: config.index,
                            options: config.options.clone(),
                            password: config.password.clone(),
                        });
                    }
                }
            } else {
                resolved_sheet_configs.push(config.clone());
            }
        }
    }

    let sheet_configs = if !protect_all_sheets && !resolved_sheet_configs.is_empty() {
        resolved_sheet_configs.as_slice()
    } else {
        sheet_configs
    };

    // 4. ZIP einmal durchlaufen und Schutz injizieren
    let file = File::open(path)?;
    let mut archive = ZipArchive::new(file)?;

    let temp_path = path.with_extension("tmp");
    let out_file = File::create(&temp_path)?;
    let mut zip_writer = ZipWriter::new(out_file);

    for i in 0..archive.len() {
        let mut entry = archive.by_index(i)?;
        let name = entry.name().to_string();
        let compression = entry.compression();
        let unix_mode = entry.unix_mode();

        if name == "xl/workbook.xml" {
            let mut content = Vec::new();
            entry.read_to_end(&mut content)?;
            let new_xml = inject_workbook_protection(&content, wb_hash.as_ref())?;
            let options = FileOptions::<()>::default()
                .compression_method(compression)
                .unix_permissions(unix_mode.unwrap_or(0o644));
            zip_writer.start_file(&name, options)?;
            zip_writer.write_all(&new_xml)?;
        } else if name.starts_with("xl/worksheets/sheet") && name.ends_with(".xml") {
            // Sheet-Index aus Dateiname ableiten (sheet1.xml → Index 0)
            let file_index = name
                .trim_start_matches("xl/worksheets/sheet")
                .trim_end_matches(".xml")
                .parse::<u32>()
                .ok()
                .map(|i| i.saturating_sub(1));
            let resolved_name = file_index.and_then(|i| index_to_name.get(&i).cloned());

            let should_protect = protect_all_sheets
                || protected_files.contains(&name)
                || resolved_name
                    .as_ref()
                    .is_some_and(|rn| sheet_hashes.contains_key(rn));

            if should_protect {
                // Per-Sheet-Konfiguration suchen
                let sheet_config = if protect_all_sheets {
                    // "Alle schützen": Erster Config gilt als Template für jedes Sheet
                    sheet_configs.first()
                } else {
                    sheet_configs.iter().find(|c| {
                        protected_files.contains(&name)
                            || (!c.name.is_empty()
                                && resolved_name.as_ref().is_some_and(|rn| rn == &c.name))
                    })
                };

                match sheet_config {
                    Some(config) => {
                        let (hash, opts) = sheet_hashes.get(&config.name).unwrap();
                        let mut content = Vec::new();
                        entry.read_to_end(&mut content)?;
                        let new_xml = inject_sheet_protection(&content, Some(hash), Some(*opts))?;
                        let options = FileOptions::<()>::default()
                            .compression_method(compression)
                            .unix_permissions(unix_mode.unwrap_or(0o644));
                        zip_writer.start_file(&name, options)?;
                        zip_writer.write_all(&new_xml)?;
                    }
                    None => {
                        // Kein spezifisches Config: Schutz ohne Optionen (nur Passwort oder leer)
                        let mut content = Vec::new();
                        entry.read_to_end(&mut content)?;
                        let new_xml = inject_sheet_protection(&content, None, None)?;
                        let options = FileOptions::<()>::default()
                            .compression_method(compression)
                            .unix_permissions(unix_mode.unwrap_or(0o644));
                        zip_writer.start_file(&name, options)?;
                        zip_writer.write_all(&new_xml)?;
                    }
                }
            } else {
                zip_writer.raw_copy_file(entry)?;
            }
        } else {
            zip_writer.raw_copy_file(entry)?;
        }
    }

    zip_writer.finish()?;
    std::fs::rename(&temp_path, path)?;
    Ok(())
}
