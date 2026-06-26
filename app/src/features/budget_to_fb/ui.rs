use super::config::{apply_b2f_defaults, load_b2f_settings, save_b2f_settings};
use super::models::FbExportModel;
use crate::shared::models::{ExportOptions, ProgressMessage};
use crate::shared::process::get_fb_path;
use crate::{BudgetState, MainWindow};
use slint::{ComponentHandle, Model};

pub fn setup(ui: &MainWindow) {
    apply_b2f_defaults(ui);
    load_b2f_settings(ui);

    // ==========================================
    // Budget-to-FB Callbacks
    // ==========================================

    ui.global::<BudgetState>().on_select_src({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<BudgetState>()
                        .set_src_folder(path.to_string_lossy().to_string().into());
                    save_b2f_settings(&ui);
                }
            }
        }
    });

    ui.global::<BudgetState>().on_select_out({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<BudgetState>()
                        .set_out_folder(path.to_string_lossy().to_string().into());
                    save_b2f_settings(&ui);
                }
            }
        }
    });

    ui.global::<BudgetState>().on_scan({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();

                let src = b2f.get_src_folder().to_string();
                let out_base = b2f.get_out_folder().to_string();

                if src.is_empty() {
                    b2f.set_status_type("error".into());
                    b2f.set_status_message("Bitte Quellordner wählen.".into());
                    return;
                }
                if out_base.is_empty() {
                    b2f.set_status_type("error".into());
                    b2f.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }

                b2f.set_status_type("pending".into());
                b2f.set_status_message("Scannt...".into());

                let filename = b2f.get_name().to_string();

                let sp = b2f.get_sheet_permissions();
                let options = ExportOptions {
                    protect_sheet: b2f.get_protect_sheet(),
                    protect_workbook: b2f.get_protect_workbook(),
                    sheet_password: b2f.get_sheet_password().to_string(),
                    workbook_password: b2f.get_workbook_password().to_string(),
                    hide_columns: b2f.get_hide_columns(),
                    hide_lang_sheet: b2f.get_hide_lang_sheet(),
                    empty_rows: b2f.get_empty_rows(),
                    protection: sp.into(),
                };

                let wb_hash = if options.protect_workbook {
                    Some(excel_protection::precompute_hash(
                        &options.workbook_password,
                    ))
                } else {
                    None
                };

                let sh_hash = if options.protect_sheet {
                    Some(excel_protection::precompute_hash(&options.sheet_password))
                } else {
                    None
                };

                let sh_opts = if options.protect_sheet {
                    Some(options.protection.clone()) // 2. Einfach das fertige Objekt klonen!
                } else {
                    None
                };

                // Dem Sidecar geben wir protect=false mit, damit es das XML nicht verschlüsselt
                let mut sidecar_options = options.clone();
                sidecar_options.protect_sheet = false;
                sidecar_options.protect_workbook = false;
                let options_json = serde_json::to_string(&sidecar_options).unwrap_or_default();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let start_time = std::time::Instant::now();
                    let src_path = std::path::PathBuf::from(&src);
                    let out_base_path = std::path::PathBuf::from(&out_base);

                    // 1. Budget-Dateien scannen
                    let result = budget_scanner::scan_directory(&src_path);

                    // 3. Output-Ordner bestimmen
                    let output_dir = budget_scanner::resolve_output_dir(&out_base_path);
                    let _ = std::fs::create_dir_all(&output_dir);

                    // 4. Temporäres JSON erstellen (wird automatisch gelöscht, wenn tmp_json_file out-of-scope geht)
                    let mut tmp_json_file = match tempfile::Builder::new()
                        .prefix("scan_")
                        .suffix(".json")
                        .tempfile()
                    {
                        Ok(f) => f,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let b2f = ui.global::<BudgetState>();
                                b2f.set_status_type("error".into());
                                b2f.set_status_message(
                                    format!("Fehler beim Erstellen der temporären Datei: {e}")
                                        .into(),
                                );
                            });
                            return;
                        }
                    };

                    // Generische BudgetData auf das flache FB-Sidecar-Schema abbilden.
                    let fb_models: Vec<FbExportModel> =
                        result.successes.iter().map(FbExportModel::from_budget).collect();

                    if let Err(e) = std::io::Write::write_all(
                        &mut tmp_json_file,
                        serde_json::to_string(&fb_models)
                            .unwrap_or_default()
                            .as_bytes(),
                    ) {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_type("error".into());
                            b2f.set_status_message(
                                format!("Fehler beim Speichern der JSON: {e}").into(),
                            );
                        });
                        return;
                    }
                    let _ = std::io::Write::flush(&mut tmp_json_file);

                    // WICHTIG FÜR WINDOWS:
                    // .into_temp_path() schließt den Dateihandle in Rust, aber die Datei bleibt auf
                    // der Festplatte erhalten, bis tmp_json_path am Ende des Threads gelöscht wird.
                    // Das verhindert "Access Denied" Fehler beim Go-Sidecar.
                    let tmp_json_path = tmp_json_file.into_temp_path();

                    // 5. Go Sidecar aufrufen
                    let sidecar_exe = get_fb_path();

                    let mut cmd = std::process::Command::new(&sidecar_exe);

                    #[cfg(target_os = "windows")]
                    {
                        use std::os::windows::process::CommandExt;
                        const CREATE_NO_WINDOW: u32 = 0x08000000;
                        cmd.creation_flags(CREATE_NO_WINDOW);
                    }

                    cmd.arg("-input")
                        .arg(&tmp_json_path)
                        .arg("-output")
                        .arg(&output_dir)
                        .arg("-options")
                        .arg(&options_json)
                        .arg("-filename")
                        .arg(&filename);

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let b2f = ui.global::<BudgetState>();
                                b2f.set_status_type("error".into());
                                b2f.set_status_message(
                                    format!(
                                        "Fehler beim Starten von {}: {e}",
                                        sidecar_exe.display()
                                    )
                                    .into(),
                                );
                            });
                            return;
                        }
                    };

                    let Some(stdout) = child.stdout.take() else {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_type("error".into());
                            b2f.set_status_message(
                                "Konnte die Ausgabe des Hintergrund-Prozesses nicht lesen.".into(),
                            );
                        });
                        return; // Thread sicher beenden
                    };
                    let reader = std::io::BufReader::new(stdout);

                    use std::io::BufRead;
                    for line in reader.lines().map_while(Result::ok) {
                        if let Ok(msg) = serde_json::from_str::<ProgressMessage>(&line) {
                            let _ = ui_handle_clone.upgrade_in_event_loop({
                                let msg_status = msg.status.clone();
                                let msg_text = msg.message.clone();
                                let current = msg.current.unwrap_or(0);
                                let total = msg.total.unwrap_or(0);

                                move |ui| {
                                    let b2f = ui.global::<BudgetState>();

                                    if msg_status == "error" {
                                        b2f.set_status_type("error".into());
                                    } else if msg_status == "done" {
                                        b2f.set_status_type("success".into());
                                    } else {
                                        b2f.set_status_type("pending".into());
                                    }

                                    if total > 0 {
                                        b2f.set_status_message(
                                            format!("{current}/{total} - {msg_text}").into(),
                                        );
                                    } else {
                                        b2f.set_status_message(msg_text.into());
                                    }
                                }
                            });
                        }
                    }

                    let _ = child.wait();
                    // Tempfile löscht sich automatisch am Ende des Scopes von tmp_json_file

                    // --- RUST EXCEL PROTECTION ---
                    // Wir durchlaufen alle generierten XLSX Dateien und wenden den schnellen XML-Schutz an
                    if wb_hash.is_some() || sh_hash.is_some() {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let b2f = ui.global::<BudgetState>();
                            b2f.set_status_message("Wende Schutz an...".into());
                        });

                        use rayon::prelude::*;
                        if let Ok(entries) = std::fs::read_dir(&output_dir) {
                            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                            paths.into_par_iter().for_each(|p| {
                                if p.extension().is_some_and(|ext| ext == "xlsx") {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &p,
                                        wb_hash.as_ref(),
                                        sh_hash.as_ref(),
                                        sh_opts.as_ref(),
                                    );
                                }
                            });
                        }
                    }

                    // 6. Fehler-CSV schreiben
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle aktualisieren
                    let success_count = result.successes.len();
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let b2f = ui.global::<BudgetState>();

                        let mk_col = |t: &str| {
                            let mut c = slint::TableColumn::default();
                            c.title = t.into();
                            c
                        };
                        let columns = slint::ModelRc::new(slint::VecModel::from(vec![
                            mk_col("Dateiname"),
                            mk_col("Status"),
                            mk_col("Details"),
                        ]));
                        b2f.set_table_columns(columns);

                        let mut rows: Vec<slint::ModelRc<slint::StandardListViewItem>> = Vec::new();

                        for data in &result.successes {
                            let fname = data
                                .file_path
                                .file_name()
                                .map(|n| n.to_string_lossy().to_string())
                                .unwrap_or_default();

                            rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    &fname,
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from("OK")),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    "Generiert",
                                )),
                            ])));
                        }

                        for f in &result.failures {
                            rows.push(slint::ModelRc::new(slint::VecModel::from(vec![
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    &f.file_name,
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    "Fehler",
                                )),
                                slint::StandardListViewItem::from(slint::SharedString::from(
                                    f.reason.to_string(),
                                )),
                            ])));
                        }

                        let table_data = slint::ModelRc::new(slint::VecModel::from(rows));
                        b2f.set_table_data(table_data);

                        let elapsed_sec = start_time.elapsed().as_secs_f64();
                        b2f.set_status_type("success".into());
                        b2f.set_status_message(
                            format!(
                                "Erfolgreich abgeschlossen! {} Datei(en) in {:.2}s erstellt.",
                                success_count, elapsed_sec
                            )
                            .into(),
                        );
                    });
                });
            }
        }
    });

    // ==========================================
    // CSV Export (Jetzt im Hintergrund-Thread!)
    // ==========================================
    ui.global::<BudgetState>().on_do_export_txt({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();

                // 1. Daten thread-sicher auf dem Haupt-Thread auslesen
                let (headers, rows) = extract_table_data(&b2f);

                // 2. Datei-Dialog (muss auf dem Haupt-Thread laufen)
                if let Some(path) = rfd::FileDialog::new()
                    .set_file_name("scan_ergebnis.csv")
                    .add_filter("CSV", &["csv"])
                    .save_file()
                {
                    b2f.set_status_type("pending".into());
                    b2f.set_status_message("Exportiere CSV...".into());

                    // 3. Thread spawnen für die Formatierung und das direkte Streamen auf die Festplatte
                    let ui_handle_clone = ui_handle.clone();
                    std::thread::spawn(move || {
                        match super::utils::export_csv_to_file(&headers, &rows, &path) {
                            Ok(()) => {
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let b2f = ui.global::<BudgetState>();
                                    b2f.set_status_type("success".into());
                                    b2f.set_status_message(
                                        format!("CSV exportiert: {}", path.display()).into(),
                                    );
                                });
                            }
                            Err(e) => {
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let b2f = ui.global::<BudgetState>();
                                    b2f.set_status_type("error".into());
                                    b2f.set_status_message(
                                        format!("CSV-Export Fehler: {e}").into(),
                                    );
                                });
                            }
                        }
                    });
                }
            }
        }
    });

    // ==========================================
    // Excel Export (Jetzt im Hintergrund-Thread!)
    // ==========================================
    ui.global::<BudgetState>().on_do_export_excel({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();

                // 1. Daten thread-sicher auf dem Haupt-Thread auslesen
                let (headers, rows) = extract_table_data(&b2f);

                // 2. Datei-Dialog (muss auf dem Haupt-Thread laufen)
                if let Some(path) = rfd::FileDialog::new()
                    .set_file_name("scan_ergebnis.xlsx")
                    .add_filter("Excel", &["xlsx"])
                    .save_file()
                {
                    b2f.set_status_type("pending".into());
                    b2f.set_status_message("Exportiere Excel...".into());

                    // 3. Thread spawnen für das Erstellen und direkte Speichern des Workbooks
                    let ui_handle_clone = ui_handle.clone();
                    std::thread::spawn(move || {
                        match super::utils::export_excel_to_file(&headers, &rows, &path) {
                            Ok(()) => {
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let b2f = ui.global::<BudgetState>();
                                    b2f.set_status_type("success".into());
                                    b2f.set_status_message(
                                        format!("Excel exportiert: {}", path.display()).into(),
                                    );
                                });
                            }
                            Err(e) => {
                                let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                    let b2f = ui.global::<BudgetState>();
                                    b2f.set_status_type("error".into());
                                    b2f.set_status_message(
                                        format!("Excel-Generierungs/Speicher Fehler: {e}").into(),
                                    );
                                });
                            }
                        }
                    });
                }
            }
        }
    });

    ui.global::<BudgetState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                b2f.set_status_type("idle".into());
                b2f.set_status_message("".into());
            }
        }
    });

    ui.global::<BudgetState>().on_do_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_b2f_defaults(&ui);
                let b2f = ui.global::<BudgetState>();
                b2f.set_table_data(slint::ModelRc::default());
                b2f.set_table_columns(slint::ModelRc::default());
                save_b2f_settings(&ui);
            }
        }
    });

    ui.global::<BudgetState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let b2f = ui.global::<BudgetState>();
                b2f.set_show_settings(!b2f.get_show_settings());
            }
        }
    });

    ui.global::<BudgetState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_b2f_settings(&ui);
            }
        }
    });
}

// Hilfsfunktion zum thread-sicheren Auslesen der Slint-Tabelle auf dem Haupt-Thread
fn extract_table_data(b2f: &BudgetState) -> (Vec<String>, Vec<Vec<String>>) {
    let columns = b2f.get_table_columns();
    let table_data = b2f.get_table_data();

    let col_count = columns.row_count();
    let mut headers = Vec::with_capacity(col_count);
    for c in 0..col_count {
        let title = columns
            .row_data(c)
            .map(|col| col.title.to_string())
            .unwrap_or_default();
        headers.push(title);
    }

    let row_count = table_data.row_count();
    let mut rows = Vec::with_capacity(row_count);
    for r in 0..row_count {
        if let Some(row) = table_data.row_data(r) {
            let mut row_vec = Vec::with_capacity(col_count);
            for c in 0..col_count {
                let text = row
                    .row_data(c)
                    .map(|item| item.text.to_string())
                    .unwrap_or_default();
                row_vec.push(text);
            }
            rows.push(row_vec);
        }
    }

    (headers, rows)
}
