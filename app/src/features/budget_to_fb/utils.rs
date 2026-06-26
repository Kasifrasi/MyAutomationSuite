use rust_xlsxwriter::{Workbook, XlsxError};
use std::path::Path;

pub fn export_csv_to_file(
    headers: &[String],
    rows: &[Vec<String>],
    path: &Path,
) -> Result<(), std::io::Error> {
    let mut wtr = csv::WriterBuilder::new().delimiter(b';').from_path(path)?;

    wtr.write_record(headers)?;
    for row in rows {
        wtr.write_record(row)?;
    }

    wtr.flush()?;
    Ok(())
}

pub fn export_excel_to_file(
    headers: &[String],
    rows: &[Vec<String>],
    path: &Path,
) -> Result<(), XlsxError> {
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

    workbook.save(path)
}
