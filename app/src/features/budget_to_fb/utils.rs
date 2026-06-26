use rust_xlsxwriter::{Workbook, XlsxError};

/// Konvertiert Tabellen-Spalten und -Zeilen in einen CSV-String (Semicolon-separiert).
/// Perfekt unit-testbar!
pub fn generate_csv_string(headers: &[String], rows: &[Vec<String>]) -> String {
    let mut out = String::new();
    out.push_str(&headers.join(";"));
    out.push('\n');
    for row in rows {
        out.push_str(&row.join(";"));
        out.push('\n');
    }
    out
}

/// Erstellt ein Excel-Workbook aus flachen Tabellendaten.
/// Komplett entkoppelt von der UI-Bibliothek.
pub fn create_excel_report(headers: &[String], rows: &[Vec<String>]) -> Result<Workbook, XlsxError> {
    let mut workbook = Workbook::new();
    let sheet = workbook.add_worksheet();
    
    // Header schreiben
    for (c, title) in headers.iter().enumerate() {
        sheet.write_string(0, c as u16, title)?;
    }
    
    // Datenzeilen schreiben
    for (r, row) in rows.iter().enumerate() {
        for (c, text) in row.iter().enumerate() {
            sheet.write_string((r + 1) as u32, c as u16, text)?;
        }
    }
    
    Ok(workbook)
}