use super::config::{apply_vp_defaults, load_vp_settings, save_vp_settings};

use crate::shared::process::get_vorpruefung_path;
use crate::{MainWindow, VorpruefungState};
use slint::ComponentHandle;

pub fn setup(ui: &MainWindow) {
    apply_vp_defaults(ui);
    load_vp_settings(ui);

    // ==========================================
    // Budget-zu-Prüfvorlage (Vorpruefung) Callbacks
    // ==========================================

    ui.global::<VorpruefungState>().on_select_src({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<VorpruefungState>()
                        .set_src_folder(path.to_string_lossy().to_string().into());
                    save_vp_settings(&ui);
                }
            }
        }
    });

    ui.global::<VorpruefungState>().on_select_out({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    ui.global::<VorpruefungState>()
                        .set_out_folder(path.to_string_lossy().to_string().into());
                    save_vp_settings(&ui);
                }
            }
        }
    });

    ui.global::<VorpruefungState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();
                vp.set_status_type("idle".into());
                vp.set_status_message("".into());
            }
        }
    });

    ui.global::<VorpruefungState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();
                vp.set_show_settings(!vp.get_show_settings());
            }
        }
    });

    ui.global::<VorpruefungState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_vp_settings(&ui);
            }
        }
    });

    ui.global::<VorpruefungState>().on_do_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_vp_defaults(&ui);
                save_vp_settings(&ui);
            }
        }
    });

    ui.global::<VorpruefungState>().on_generate({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let vp = ui.global::<VorpruefungState>();

                let src = vp.get_src_folder().to_string();
                let out_base = vp.get_out_folder().to_string();
                let name = vp.get_name().to_string();

                if src.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Budget-Ordner wählen.".into());
                    return;
                }
                if out_base.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }
                if name.is_empty() {
                    vp.set_status_type("error".into());
                    vp.set_status_message("Bitte Dateinamens-Muster angeben.".into());
                    return;
                }

                vp.set_status_type("pending".into());
                vp.set_status_message("Scannt Budgets...".into());

                let protect_workbook = vp.get_protect_workbook();
                let wb_password = vp.get_workbook_password().to_string();

                let ui_handle_clone = ui_handle.clone();
                std::thread::spawn(move || {
                    let start_time = std::time::Instant::now();
                    let src_path = std::path::PathBuf::from(&src);
                    let out_base_path = std::path::PathBuf::from(&out_base);

                    // 1. Budgets scannen
                    let result = budget_scanner::scan_directory(&src_path);

                    // 2. Output-Ordner
                    let output_dir = budget_scanner::resolve_output_dir(&out_base_path);
                    let _ = std::fs::create_dir_all(&output_dir);

                    let wb_hash = if protect_workbook {
                        Some(excel_protection::precompute_hash(&wb_password))
                    } else {
                        None
                    };

                    let sidecar_exe = get_vorpruefung_path();
                    let total = result.successes.len() as u32;

                    // (Quelldateiname, Status, Detail)
                    let mut rows_info: Vec<(String, String, String)> = Vec::new();

                    let ok_count = match crate::shared::process::run_sidecar_batch(
                        &sidecar_exe,
                        &result.successes,
                        &output_dir,
                        &name,
                        None,
                        wb_hash,
                        None,
                        None,
                        |msg| {
                            let _ = ui_handle_clone.upgrade_in_event_loop({
                                let msg_status = msg.status.clone();
                                let msg_text = msg.message.clone();
                                let current = msg.current.unwrap_or(0);
                                let total = msg.total.unwrap_or(0);

                                move |ui| {
                                    let vp = ui.global::<VorpruefungState>();
                                    if msg_status == "error" {
                                        vp.set_status_type("error".into());
                                    } else if msg_status == "done" {
                                        vp.set_status_type("success".into());
                                    } else {
                                        vp.set_status_type("pending".into());
                                    }

                                    if total > 0 {
                                        vp.set_status_message(
                                            format!("{current}/{total} - {msg_text}").into(),
                                        );
                                    } else {
                                        vp.set_status_message(msg_text.into());
                                    }
                                }
                            });
                        },
                    ) {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("error".into());
                                vp.set_status_message(e.into());
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

                    // 6. Scan-Fehler ergänzen + CSV
                    for f in &result.failures {
                        rows_info.push((
                            f.file_name.clone(),
                            "Fehler".into(),
                            f.reason.to_string(),
                        ));
                    }
                    if !result.failures.is_empty() {
                        let csv_path = output_dir.join("scan_fehler.csv");
                        let _ = budget_scanner::write_failure_report(&result.failures, &csv_path);
                    }

                    // 7. Tabelle + Status aktualisieren
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let vp = ui.global::<VorpruefungState>();

                        let mk_col = |t: &str| {
                            let mut c = slint::TableColumn::default();
                            c.title = t.into();
                            c
                        };
                        vp.set_table_columns(slint::ModelRc::new(slint::VecModel::from(vec![
                            mk_col("Budget-Datei"),
                            mk_col("Status"),
                            mk_col("Details"),
                        ])));

                        let rows: Vec<slint::ModelRc<slint::StandardListViewItem>> = rows_info
                            .iter()
                            .map(|(file, status, detail)| {
                                slint::ModelRc::new(slint::VecModel::from(vec![
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        file.as_str(),
                                    )),
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        status.as_str(),
                                    )),
                                    slint::StandardListViewItem::from(slint::SharedString::from(
                                        detail.as_str(),
                                    )),
                                ]))
                            })
                            .collect();
                        vp.set_table_data(slint::ModelRc::new(slint::VecModel::from(rows)));

                        let elapsed = start_time.elapsed().as_secs_f64();
                        let fail_count = result.failures.len() as u32 + (total - ok_count);
                        if ok_count == 0 {
                            vp.set_status_type("error".into());
                            vp.set_status_message(
                                format!("Keine Prüfvorlage erstellt ({fail_count} Fehler).").into(),
                            );
                        } else {
                            vp.set_status_type("success".into());
                            vp.set_status_message(
                                format!(
                                    "{ok_count} Prüfvorlage(n) in {elapsed:.2}s erstellt{}.",
                                    if fail_count > 0 {
                                        format!(", {fail_count} Fehler")
                                    } else {
                                        String::new()
                                    }
                                )
                                .into(),
                            );
                        }
                    });
                });
            }
        }
    });
}
