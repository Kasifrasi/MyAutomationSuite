use excel_protection::color_routing::*;
use std::fs;

#[test]
fn test_daten_rewrite() {
    let xml = fs::read_to_string("../sidecars/Vorpruefung/test_bug.xlsx").unwrap_or_default();
    println!("File read");
}
