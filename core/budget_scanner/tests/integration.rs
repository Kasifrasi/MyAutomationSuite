use budget_scanner::{scan_directory, scan_file, write_failure_report, ScanError};
use std::path::Path;

const ASSETS: &str = "assets";

fn asset(name: &str) -> std::path::PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join(ASSETS).join(name)
}

#[test]
fn scan_german_v2() {
    let result = scan_file(&asset("Kosten-_und_Finanzierungsplan_V2.xlsx"));
    let data = result.expect("Scan should succeed");

    assert_eq!(data.sheet_name, "Budget");
    assert!(data.version.to_uppercase().contains("V2"));
    assert!(!data.positions.is_empty(), "Should find positions");
    assert!(
        data.positions[0].number.starts_with("1."),
        "First position should start with 1."
    );
}

#[test]
fn scan_english_v2() {
    let result = scan_file(&asset("Cost-_and_financing_plan_V2.xlsx"));
    let data = result.expect("Scan should succeed");

    assert_eq!(data.sheet_name, "Budget");
    assert!(data.version.to_uppercase().contains("V2"));
    assert!(!data.positions.is_empty());
}

#[test]
fn scan_french_v2() {
    let result = scan_file(&asset("F_Budget_et_plan_de_financement_V2.xlsx"));
    let data = result.expect("Scan should succeed");

    assert_eq!(data.sheet_name, "Budget");
    assert!(!data.positions.is_empty());
}

#[test]
fn scan_spanish_v2() {
    let result = scan_file(&asset("Plan_de_costos_y_financiamento_V2.xlsx"));
    let data = result.expect("Scan should succeed");

    assert_eq!(data.sheet_name, "Presupuesto");
    assert!(!data.positions.is_empty());
}

#[test]
fn scan_portuguese_v2() {
    let result = scan_file(&asset("Plano_de_custos_e_financiamento_V2.xlsx"));
    let data = result.expect("Scan should succeed");

    assert_eq!(data.sheet_name, "Plano de custos e financiamento");
    assert!(!data.positions.is_empty());
}

#[test]
fn scan_no_matching_sheet() {
    let result = scan_file(&asset("Anlage_4_Finanzbericht_Vorlage.xlsx"));
    let failure = result.expect_err("Should fail — no matching sheet");

    assert!(
        matches!(failure.reason, ScanError::NoMatchingSheet { .. }),
        "Expected NoMatchingSheet, got: {:?}",
        failure.reason
    );
}

#[test]
fn scan_directory_finds_files() {
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join(ASSETS);
    let result = scan_directory(&dir);

    assert!(
        !result.successes.is_empty(),
        "Should have at least one success"
    );
    // Anlage_4 hat kein passendes Sheet → sollte in failures sein
    assert!(
        !result.failures.is_empty(),
        "Should have at least one failure"
    );
}

#[test]
fn write_failure_report_creates_file() {
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join(ASSETS);
    let result = scan_directory(&dir);

    let report_path = std::env::temp_dir().join("budget_scanner_test_report.csv");
    write_failure_report(&result.failures, &report_path).expect("Should write report");

    let content = std::fs::read_to_string(&report_path).expect("Should read report");
    assert!(content.contains("Dateiname;Grund;Pfad"));
    assert!(content.contains("Anlage_4"));

    let _ = std::fs::remove_file(&report_path);
}
