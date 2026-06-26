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

                    // Alle Daten in EINEM Array zusammenfassen (Batch-Verarbeitung)
                    let mut tmp_json_file = match tempfile::Builder::new()
                        .prefix("vp_budgets_")
                        .suffix(".json")
                        .tempfile()
                    {
                        Ok(f) => f,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("error".into());
                                vp.set_status_message(format!("Temp-JSON Fehler: {e}").into());
                            });
                            return;
                        }
                    };

                    let json = match serde_json::to_string(&result.successes) {
                        Ok(j) => j,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("error".into());
                                vp.set_status_message(format!("JSON-Serialize Fehler: {e}").into());
                            });
                            return;
                        }
                    };

                    if let Err(e) = std::io::Write::write_all(&mut tmp_json_file, json.as_bytes()) {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let vp = ui.global::<VorpruefungState>();
                            vp.set_status_type("error".into());
                            vp.set_status_message(format!("Temp-JSON Schreibfehler: {e}").into());
                        });
                        return;
                    }
                    let _ = std::io::Write::flush(&mut tmp_json_file);
                    let tmp_json_path = tmp_json_file.into_temp_path();

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
                        .arg("-filename")
                        .arg(&name);

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("error".into());
                                vp.set_status_message(
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
                            let vp = ui.global::<VorpruefungState>();
                            vp.set_status_type("error".into());
                            vp.set_status_message("Konnte Ausgabe nicht lesen.".into());
                        });
                        return;
                    };

                    use std::io::BufRead;
                    #[derive(serde::Deserialize)]
                    struct ProgressMessage {
                        status: String,
                        message: String,
                        current: Option<u32>,
                        total: Option<u32>,
                        file: Option<String>,
                    }

                    let reader = std::io::BufReader::new(stdout);
                    let mut ok_count = 0u32;

                    for line in reader.lines().map_while(Result::ok) {
                        if let Ok(msg) = serde_json::from_str::<ProgressMessage>(&line) {
                            if msg.status == "success" || msg.status == "progress" {
                                ok_count += 1;
                                if let Some(ref f) = msg.file {
                                    rows_info.push((
                                        f.clone(),
                                        "OK".into(),
                                        "Vorprüfung generiert".into(),
                                    ));
                                }
                            } else if msg.status == "error" {
                                if let Some(ref f) = msg.file {
                                    rows_info.push((
                                        f.clone(),
                                        "Fehler".into(),
                                        msg.message.clone(),
                                    ));
                                }
                            }

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
                        }
                    }

                    let _ = child.wait();

                    // --- RUST EXCEL PROTECTION ---
                    if wb_hash.is_some() {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let vp = ui.global::<VorpruefungState>();
                            vp.set_status_message("Wende Schutz an...".into());
                        });

                        use rayon::prelude::*;
                        if let Ok(entries) = std::fs::read_dir(&output_dir) {
                            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                            paths.into_par_iter().for_each(|p| {
                                if p.extension().is_some_and(|ext| ext == "xlsx") {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &p,
                                        wb_hash.as_ref(),
                                        None,
                                        None,
                                    );
                                }
                            });
                        }
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
