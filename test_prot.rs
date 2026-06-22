use excel_protection::{apply_protection_in_place, precompute_hash, SheetProtectionOptions};

fn main() {
    let _ = std::fs::copy("test.xlsx", "test_prot.xlsx");
    let wb = precompute_hash("test");
    let sh = precompute_hash("test");
    let opts = SheetProtectionOptions::default();
    
    apply_protection_in_place(
        std::path::Path::new("test_prot.xlsx"),
        Some(&wb),
        Some(&sh),
        Some(&opts),
    ).unwrap();
}
