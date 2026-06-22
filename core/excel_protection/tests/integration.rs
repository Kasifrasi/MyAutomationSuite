use excel_protection::{inject_sheet_protection, precompute_hash, SheetProtectionOptions};

#[test]
fn test_sheet_protection_full() {
    let xml = br#"<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
    <sheetData>
        <row r="1" spans="1:1"/>
    </sheetData>
    <pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/>
</worksheet>"#;

    let hash = precompute_hash("password");
    let opts = SheetProtectionOptions::default();
    
    let injected = inject_sheet_protection(xml, Some(&hash), Some(&opts)).unwrap();
    let s = String::from_utf8_lossy(&injected);
    println!("Injected:\n{}", s);
}
