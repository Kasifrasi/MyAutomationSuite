use super::config::{apply_b2f_defaults, load_b2f_settings, save_b2f_settings};
use crate::shared::models::ExportOptions;
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

                let wb_config = if b2f.get_protect_workbook() {
                    Some(excel_protection::WorkbookConfig {
                        password: Some(b2f.get_workbook_password().to_string()),
                    })
                } else {
                    None
                };

                let sheet_configs = if b2f.get_protect_sheet() {
                    vec![excel_protection::SheetConfig {
                        name: String::new(),
                        index: None, // None = alle Sheets schützen
                        options: b2f.get_sheet_permissions().into(),
                        password: Some(b2f.get_sheet_password().to_string()),
                    }]
                } else {
                    vec![]
                };

                let options = ExportOptions {
                    hide_columns: b2f.get_hide_columns(),
                    hide_lang_sheet: b2f.get_hide_lang_sheet(),
                    empty_rows: b2f.get_empty_rows(),
                    is_template: false,
                    workbook: wb_config.clone().unwrap_or_default(),
                    sheet_configs: sheet_configs.clone(),
                };

                let mut sidecar_options = options.clone();
                sidecar_options.workbook = excel_protection::WorkbookConfig::default();
                sidecar_options.sheet_configs = vec![];
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

                    // 4. Sidecar Runner
                    let mut rows_info: Vec<(String, String, String)> = Vec::new();

                    let sidecar_exe = crate::shared::process::get_fb_path();

                    let ok_count = match crate::shared::process::run_sidecar_batch(
                        &sidecar_exe,
                        &result.successes,
                        &output_dir,
                        &filename,
                        Some(&options_json),
                        None,
                        wb_config.clone(),
                        sheet_configs.clone(),
                        &[],
                        |msg| {
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
                        },
                    ) {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let b2f = ui.global::<BudgetState>();
                                b2f.set_status_type("error".into());
                                b2f.set_status_message(e.into());
                            });
                            return;
                        }
                    };

                    for data in &result.successes {
                        let src_name = data
                            .file_path
                            .file_name()
                            .map(|n| n.to_string_lossy().to_string())
                            .unwrap_or_default();
                        rows_info.push((src_name, "OK".into(), "Generiert".into()));
                    }

                    // 6. Fehler-CSV schreiben
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle aktualisieren
                    let success_count = ok_count;
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
