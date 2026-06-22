use rust_xlsxwriter::{Format, Workbook, XlsxError};

fn main() -> Result<(), XlsxError> {
    let mut workbook = Workbook::new();
    let worksheet = workbook.add_worksheet();
    let unlocked = Format::new().set_unlocked();
    worksheet.protect();
    worksheet.write_string(0, 0, "Cell B1 is locked. It cannot be edited.")?;
    worksheet.write_formula(0, 1, "=1+2")?; // Locked by default.
    workbook.save("worksheet_protection.xlsx")?;
    Ok(())
}
