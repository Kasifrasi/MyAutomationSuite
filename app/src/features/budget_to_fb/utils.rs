use rust_xlsxwriter::{Workbook, XlsxError};

pub fn generate_csv_string(headers: &[String], rows: &[Vec<String>]) -> String {
    let mut wtr = csv::WriterBuilder::new()
        .delimiter(b';')
        .from_writer(vec![]);

    let _ = wtr.write_record(headers);
    for row in rows {
        let _ = wtr.write_record(row);
    }

    let data = wtr.into_inner().unwrap_or_default();
    String::from_utf8(data).unwrap_or_default()
}

/// Erstellt ein Excel-Workbook aus flachen Tabellendaten.
/// Komplett entkoppelt von der UI-Bibliothek.
pub fn create_excel_report(
    headers: &[String],
    rows: &[Vec<String>],
) -> Result<Workbook, XlsxError> {
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
