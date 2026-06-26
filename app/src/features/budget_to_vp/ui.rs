use super::config::{apply_vp_defaults, load_vp_settings, save_vp_settings};
use super::utils::vp_output_name;
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
                    let mut used_names: std::collections::HashSet<String> =
                        std::collections::HashSet::new();
                    let mut ok_count = 0u32;

                    for (i, data) in result.successes.iter().enumerate() {
                        let current = (i + 1) as u32;
                        let src_name = data
                            .file_path
                            .file_name()
                            .map(|n| n.to_string_lossy().to_string())
                            .unwrap_or_default();

                        let _ = ui_handle_clone.upgrade_in_event_loop({
                            let label = src_name.clone();
                            move |ui| {
                                let vp = ui.global::<VorpruefungState>();
                                vp.set_status_type("pending".into());
                                vp.set_status_message(
                                    format!("{current}/{total} – {label}").into(),
                                );
                            }
                        });

                        // 3. Volle kanonische BudgetData an das Sidecar geben (es pickt
                        //    sich die benötigten Felder selbst heraus).
                        let json = match serde_json::to_string(data) {
                            Ok(j) => j,
                            Err(e) => {
                                rows_info.push((
                                    src_name,
                                    "Fehler".into(),
                                    format!("JSON-Fehler: {e}"),
                                ));
                                continue;
                            }
                        };

                        let mut tmp_json_file = match tempfile::Builder::new()
                            .prefix("vp_budget_")
                            .suffix(".json")
                            .tempfile()
                        {
                            Ok(f) => f,
                            Err(e) => {
                                rows_info.push((
                                    src_name,
                                    "Fehler".into(),
                                    format!("Temp-JSON: {e}"),
                                ));
                                continue;
                            }
                        };

                        if let Err(e) =
                            std::io::Write::write_all(&mut tmp_json_file, json.as_bytes())
                        {
                            rows_info.push((src_name, "Fehler".into(), format!("Temp-JSON: {e}")));
                            continue;
                        }
                        let _ = std::io::Write::flush(&mut tmp_json_file);

                        // .into_temp_path() schließt den File-Handle für Windows,
                        // räumt aber die Datei trotzdem am Ende des Schleifendurchlaufs (Scope) ab.
                        let tmp_json_path = tmp_json_file.into_temp_path();

                        // 4. Zieldateiname + Sidecar-Aufruf
                        let out_name = vp_output_name(&name, data, &mut used_names);
                        let out_path = output_dir.join(&out_name);

                        let mut cmd = std::process::Command::new(&sidecar_exe);
                        #[cfg(target_os = "windows")]
                        {
                            use std::os::windows::process::CommandExt;
                            const CREATE_NO_WINDOW: u32 = 0x08000000;
                            cmd.creation_flags(CREATE_NO_WINDOW);
                        }
                        cmd.arg("-budget")
                            .arg(&tmp_json_path)
                            .arg("-o")
                            .arg(&out_path);

                        let run = cmd.output();
                        // Temp-JSON löscht sich am Ende des Blocks automatisch durch Drop von TempPath

                        match run {
                            Ok(o) if o.status.success() => {
                                // 5. Optionaler Mappenschutz (sperrt keine Eingabezellen)
                                if let Some(h) = wb_hash.as_ref() {
                                    let _ = excel_protection::apply_protection_in_place(
                                        &out_path,
                                        Some(h),
                                        None,
                                        None,
                                    );
                                }
                                ok_count += 1;
                                rows_info.push((src_name, "OK".into(), out_name));
                            }
                            Ok(o) => {
                                let err = String::from_utf8_lossy(&o.stderr);
                                let detail = err
                                    .lines()
                                    .last()
                                    .map(|s| s.to_string())
                                    .filter(|s| !s.is_empty())
                                    .unwrap_or_else(|| "Generierung fehlgeschlagen".into());
                                rows_info.push((src_name, "Fehler".into(), detail));
                            }
                            Err(e) => {
                                rows_info.push((
                                    src_name,
                                    "Fehler".into(),
                                    format!("Sidecar-Start: {e}"),
                                ));
                            }
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
