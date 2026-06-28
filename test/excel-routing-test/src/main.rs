use excel_protection::{apply_protection, SheetConfig, SheetProtectionOptions};
use rayon::prelude::*;
use std::fs;
use std::path::Path;
use std::time::Instant;

fn main() {
    let source = Path::new("../../testdata/Pruefvorlage/Pruefvorlage_2026_0004_001_deutsch_1.xlsx");
    let tmp_dir = Path::new("tmp_speedtest");

    // Clean and recreate temp directory
    if tmp_dir.exists() {
        fs::remove_dir_all(tmp_dir).unwrap();
    }
    fs::create_dir_all(tmp_dir).unwrap();

    let num_files = 1000;
    let mut files = Vec::new();

    println!("Bereite {} Dateien für den Speedtest vor...", num_files);
    for i in 0..num_files {
        let dest = tmp_dir.join(format!("file_{}.xlsx", i));
        fs::copy(source, &dest).unwrap();
        files.push(dest);
    }

    let target_color = "FFFAE5"; // Gelb
    let columns_to_unlock: Vec<u32> = (1..=10).collect();

    let default_config = SheetConfig {
        name: "".to_string(),
        index: None,
        options: SheetProtectionOptions {
            select_locked_cells: true,
            select_unlocked_cells: true,
            format_cells: false,
            format_columns: false,
            format_rows: false,
            insert_columns: false,
            insert_rows: false,
            insert_hyperlinks: false,
            delete_columns: false,
            delete_rows: false,
            sort: false,
            auto_filter: true,
            pivot_tables: false,
            edit_objects: false,
            edit_scenarios: false,
        },
        password: Some("testpasswort".to_string()),
    };

    println!(
        "Starte parallele Verarbeitung von {} Dateien mit Rayon...",
        num_files
    );
    let start_time = Instant::now();

    files.par_iter().for_each(|dest| {
        // General protection step (Sheet schützen + Color routing in one pass)
        apply_protection(
            dest,
            None,
            &[default_config.clone()],
            Some(target_color),
            &columns_to_unlock,
        )
        .expect("Fehler bei apply protection");
    });

    let duration = start_time.elapsed();
    let ms_per_file = duration.as_millis() as f64 / num_files as f64;

    println!("--------------------------------------------------");
    println!("🏁 Speedtest abgeschlossen!");
    println!("Gesamtdauer für {} Dateien: {:?}", num_files, duration);
    println!("Durchschnittliche Dauer pro Datei: {:.2} ms", ms_per_file);
    println!("--------------------------------------------------");
}
