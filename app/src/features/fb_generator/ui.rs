use super::config::{apply_fb_defaults, load_fb_settings, save_fb_settings};
use crate::shared::models::{ExportOptions, ProgressMessage};
use crate::shared::process::get_fb_path;
use crate::{FBState, MainWindow};
use slint::ComponentHandle;

pub fn setup(ui: &MainWindow) {
    apply_fb_defaults(ui);
    load_fb_settings(ui);

    ui.global::<FBState>().on_select_folder({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                if let Some(path) = rfd::FileDialog::new().pick_folder() {
                    let fb = ui.global::<FBState>();
                    fb.set_folder(path.to_string_lossy().to_string().into());
                    fb.set_status_type("idle".into());
                    fb.set_status_message("".into());
                }
            }
        }
    });

    ui.global::<FBState>().on_reset({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                apply_fb_defaults(&ui);
                save_fb_settings(&ui);
            }
        }
    });

    ui.global::<FBState>().on_dismiss_status({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();
                fb.set_status_type("idle".into());
                fb.set_status_message("".into());
            }
        }
    });

    ui.global::<FBState>().on_save_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                save_fb_settings(&ui);
            }
        }
    });

    ui.global::<FBState>().on_toggle_settings({
        let ui_handle = ui.as_weak();
        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();
                fb.set_show_settings(!fb.get_show_settings());
            }
        }
    });

    ui.global::<FBState>().on_generate_report({
        let ui_handle = ui.as_weak();

        move || {
            if let Some(ui) = ui_handle.upgrade() {
                let fb = ui.global::<FBState>();

                let folder = fb.get_folder().to_string();
                if folder.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte Ausgabeordner wählen.".into());
                    return;
                }

                let name = fb.get_name().to_string();
                if name.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte Dateinamens-Muster angeben.".into());
                    return;
                }

                let langs = fb.get_langs();
                let mut lang_list = Vec::new();
                if langs.de {
                    lang_list.push("de");
                }
                if langs.en {
                    lang_list.push("en");
                }
                if langs.fr {
                    lang_list.push("fr");
                }
                if langs.es {
                    lang_list.push("es");
                }
                if langs.pt {
                    lang_list.push("pt");
                }

                if lang_list.is_empty() {
                    fb.set_status_type("error".into());
                    fb.set_status_message("Bitte mindestens eine Sprache wählen.".into());
                    return;
                }

                fb.set_status_type("pending".into());
                fb.set_status_message("Export läuft...".into());

                let cats = fb.get_categories();
                let counts = [
                    cats.cat1 as u16,
                    cats.cat2 as u16,
                    cats.cat3 as u16,
                    cats.cat4 as u16,
                    cats.cat5 as u16,
                    cats.cat6 as u16,
                    cats.cat7 as u16,
                    cats.cat8 as u16,
                ];

                let name_clone = name.clone();

                let wb_config = if fb.get_protect_workbook() {
                    Some(excel_protection::WorkbookConfig {
                        password: Some(fb.get_workbook_password().to_string()),
                    })
                } else {
                    None
                };

                let sheet_configs = if fb.get_protect_sheet() {
                    vec![excel_protection::SheetConfig {
                        name: String::new(),
                        index: None, // None = alle Sheets schützen
                        options: fb.get_sheet_permissions().into(),
                        password: Some(fb.get_sheet_password().to_string()),
                    }]
                } else {
                    vec![]
                };

                let options = ExportOptions {
                    hide_columns: fb.get_hide_columns(),
                    hide_lang_sheet: fb.get_hide_lang_sheet(),
                    empty_rows: fb.get_empty_rows(),
                    is_template: true,
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
                    let mut templates = Vec::new();

                    for lang in &lang_list {
                        let mut positions = Vec::new();
                        for (i, &pos_count) in counts.iter().enumerate() {
                            let category = i + 1;
                            for p in 0..pos_count {
                                positions.push(budget_scanner::BudgetPosition {
                                    number: format!("{category}.{}", p + 1),
                                    label: String::new(),
                                    kategorie: String::new(),
                                    lc: None,
                                    y1: None,
                                    y2: None,
                                    y3: None,
                                    eur: None,
                                });
                            }
                        }

                        templates.push(budget_scanner::BudgetData {
                            file_path: std::path::PathBuf::from(format!("Vorlage_{lang}.xlsx")),
                            sheet_name: "Budget".into(),
                            version: String::new(),
                            project_title: "".into(),
                            project_number: "Vorlage".into(),
                            language: lang.to_string(),
                            local_currency: "".into(),
                            cost_col1: 8,
                            cost_col2: Some(13),
                            financing: budget_scanner::FinancingDetail::default(),
                            positions,
                        });
                    }

                    let output_dir = std::path::PathBuf::from(&folder);
                    let _ = std::fs::create_dir_all(&output_dir);

                    let mut tmp_json_file = match tempfile::Builder::new()
                        .prefix("template_")
                        .suffix(".json")
                        .tempfile()
                    {
                        Ok(f) => f,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fb = ui.global::<FBState>();
                                fb.set_status_type("error".into());
                                fb.set_status_message(
                                    format!("Fehler beim Erstellen der temporären Datei: {e}")
                                        .into(),
                                );
                            });
                            return;
                        }
                    };

                    if let Err(e) = std::io::Write::write_all(
                        &mut tmp_json_file,
                        serde_json::to_string(&templates)
                            .unwrap_or_default()
                            .as_bytes(),
                    ) {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_type("error".into());
                            fb.set_status_message(
                                format!("Fehler beim Speichern der JSON: {e}").into(),
                            );
                        });
                        return;
                    }
                    let _ = std::io::Write::flush(&mut tmp_json_file);

                    // .into_temp_path() schließt den Dateihandle in Rust, aber die Datei bleibt auf
                    // der Festplatte erhalten, bis tmp_json_path am Ende des Threads gelöscht wird.
                    let tmp_json_path = tmp_json_file.into_temp_path();

                    // 4. Go Sidecar aufrufen
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
                        .arg(&name_clone);

                    cmd.stdout(std::process::Stdio::piped());

                    let mut child = match cmd.spawn() {
                        Ok(c) => c,
                        Err(e) => {
                            let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                                let fb = ui.global::<FBState>();
                                fb.set_status_type("error".into());
                                fb.set_status_message(
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
                            let fb = ui.global::<FBState>();
                            fb.set_status_type("error".into());
                            fb.set_status_message(
                                "Konnte die Ausgabe des Hintergrund-Prozesses nicht lesen.".into(),
                            );
                        });
                        return;
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
                                    let fb = ui.global::<FBState>();

                                    if msg_status == "error" {
                                        fb.set_status_type("error".into());
                                    } else if msg_status == "done" {
                                        fb.set_status_type("success".into());
                                    } else {
                                        fb.set_status_type("pending".into());
                                    }

                                    if total > 0 {
                                        fb.set_status_message(
                                            format!("{current}/{total} - {msg_text}").into(),
                                        );
                                    } else {
                                        fb.set_status_message(msg_text.into());
                                    }
                                }
                            });
                        }
                    }

                    let _ = child.wait();
                    // Tempfile wird am Ende des Scopes automatisch gelöscht

                    if wb_config.is_some() || !sheet_configs.is_empty() {
                        let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                            let fb = ui.global::<FBState>();
                            fb.set_status_message("Wende Schutz an...".into());
                        });

                        use rayon::prelude::*;
                        if let Ok(entries) = std::fs::read_dir(&output_dir) {
                            let paths: Vec<_> = entries.flatten().map(|e| e.path()).collect();
                            paths.into_par_iter().for_each(|p| {
                                if p.extension().is_some_and(|ext| ext == "xlsx") {
                                    let _ = excel_protection::apply_protection(
                                        &p,
                                        wb_config.as_ref(),
                                        &sheet_configs,
                                        None,
                                    );
                                }
                            });
                        }
                    }

                    let success_count = templates.len();
                    let _ = ui_handle_clone.upgrade_in_event_loop(move |ui| {
                        let fb = ui.global::<FBState>();
                        let elapsed_sec = start_time.elapsed().as_secs_f64();
                        fb.set_status_type("success".into());
                        fb.set_status_message(
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
}
